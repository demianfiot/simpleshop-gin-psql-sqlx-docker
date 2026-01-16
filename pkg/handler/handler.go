package handler

import (
	"prac/pkg/service"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services service.Service
}

func NewHandler(services service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	//cors -
	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
	// 	AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	// 	AllowHeaders: []string{
	// 		"Origin",
	// 		"Content-Type",
	// 		"Authorization",
	// 		"Accept",
	// 		"X-Requested-With",
	// 		"Access-Control-Allow-Headers",
	// 		"Access-Control-Allow-Origin",
	// 	},
	// 	ExposeHeaders:    []string{"Content-Length", "Authorization"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "OK",
			"service":   "backend",
			"timestamp": time.Now(),
		})
	})

	// wout auth
	auth := router.Group("/auth")
	{
		auth.POST("/signup", h.SignUp)
		auth.POST("/signin", h.SignIn)
	}

	// w/ auth
	api := router.Group("/api", h.userIdentity)
	{

		users := api.Group("/users", h.requireRole("admin"))
		{
			users.POST("", h.CreateUser)
			users.GET("", h.GetAllUsers)
			users.GET("/:id", h.GetUserByID)
			users.PATCH("/:id", h.UpdateUser)
			users.DELETE("/:id", h.DeleteUser)
		}

		profile := api.Group("/profile")
		{
			profile.GET("", h.GetProfile)
			profile.PATCH("", h.UpdateProfile)
		}

		products := api.Group("/products")
		{
			products.POST("", h.requireRole("seller", "admin"), h.CreateProduct)
			products.GET("", h.GetAllProducts)
			products.GET("/:id", h.GetProductByID)
			products.PATCH("/:id", h.requireRole("seller", "admin"), h.UpdateProduct)
			products.DELETE("/:id", h.requireRole("seller", "admin"), h.DeleteProduct)
		}

		orders := api.Group("/orders")
		{
			orders.POST("/", h.CreateOrder)
			orders.GET("/", h.GetUserOrders)
			orders.GET("/:id", h.GetOrderByID)
			orders.PATCH("/:id", h.requireRole("admin"), h.UpdateOrderStatus)
			orders.GET("/all", h.requireRole("admin"), h.GetAllOrders)
		}
	}

	return router
}
