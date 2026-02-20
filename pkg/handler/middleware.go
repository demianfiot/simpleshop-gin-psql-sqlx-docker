package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userId"
	userRole            = "userRole"
)

func (h *Handler) userIdentity(c *gin.Context) {
	fmt.Printf("DEBUG: Path: %s, Method: %s\n", c.Request.URL.Path, c.Request.Method) // ← Додайте
	header := c.GetHeader(authorizationHeader)

	if header == "" {
		NewErrorResponse(c, 401, "empty auth header")
		return
	}

	headerParts := strings.Split(header, " ")

	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		NewErrorResponse(c, 401, "invalid auth header")
		return
	}
	ctx := c.Request.Context()
	userId, userrole, err := h.services.Authorization.ParseToken(ctx, headerParts[1])
	if err != nil {
		NewErrorResponse(c, 401, err.Error())
		return
	}
	fmt.Printf("DEBUG: UserID: %d, Role: %s\n", userId, userrole)
	c.Set(userCtx, userId)
	c.Set(userRole, userrole)
	c.Next()
}

func (h *Handler) requireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: Checking role for path: %s\n", c.Request.URL.Path)
		roleValue, exists := c.Get(userRole)
		if !exists {
			NewErrorResponse(c, http.StatusForbidden, "user role not found")
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			NewErrorResponse(c, http.StatusForbidden, "invalid user role type")
			return
		}

		// check role
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		NewErrorResponse(c, http.StatusForbidden, "insufficient permissions")
		c.Abort()
	}
}
