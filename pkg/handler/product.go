package handler

import (
	"net/http"
	"prac/todo"
	"strconv"

	"prac/pkg/repository"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateProduct(c *gin.Context) {
	var input todo.CreateProductInput

	if err := c.BindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// check ID potochngo korustuvacha
	sellerID, err := h.getUserID(c)
	if err != nil {
		NewErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	ctx := c.Request.Context()
	product, err := h.services.Product.CreateProduct(ctx, input, sellerID)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          product.ID,
		"description": product.Description,
		"name":        product.Name,
		"price":       product.Price,
		"stock":       product.Stock,
	})
}

func (h *Handler) GetAllProducts(c *gin.Context) {
	ctx := c.Request.Context()
	products, err := h.services.Product.GetAllProducts(ctx)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}

func (h *Handler) GetProductByID(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "invalid product id param")
		return
	}
	ctx := c.Request.Context()
	product, err := h.services.Product.GetProductByID(ctx, uint(productID))
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"product": product,
	})
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "invalid product id param")
		return
	}

	var input todo.UpdateProductInput
	if err := c.BindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// check ID dl9 prav
	currentUserID, err := h.getUserID(c)
	if err != nil {
		NewErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	ctx := c.Request.Context()
	product, err := h.services.Product.GetProductByID(ctx, uint(productID))
	if err != nil {
		NewErrorResponse(c, http.StatusNotFound, "product not found")
		return
	}
	if product.SellerID != currentUserID {
		NewErrorResponse(c, http.StatusForbidden, "you are not owner of this product")
		return
	}
	ctx = c.Request.Context()
	updatedProduct, err := h.services.Product.UpdateProduct(ctx, uint(productID), input, currentUserID)
	if err != nil {
		if err == repository.ErrProductNotFound {
			NewErrorResponse(c, http.StatusNotFound, "product not found")
			return
		}
		if err == repository.ErrAccessDenied {
			NewErrorResponse(c, http.StatusForbidden, "access denied")
			return
		}
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, updatedProduct)
}
func (h *Handler) DeleteProduct(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "invalid user id param")
		return
	}
	currentUserID, err := h.getUserID(c)
	if err != nil {
		NewErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	ctx := c.Request.Context()
	product, err := h.services.Product.GetProductByID(ctx, uint(productID))
	if err != nil {
		NewErrorResponse(c, http.StatusNotFound, "product not found")
		return
	}
	if product.SellerID != currentUserID {
		NewErrorResponse(c, http.StatusForbidden, "you are not owner of this product")
		return
	}
	ctx = c.Request.Context()
	err = h.services.Product.DeleteProduct(ctx, productID)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}
