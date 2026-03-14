package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bimidu/ctse-order-service/grpc/clients"
	grpcserver "github.com/Bimidu/ctse-order-service/grpc/server"
	"github.com/Bimidu/ctse-order-service/internal/config"
	"github.com/Bimidu/ctse-order-service/internal/database"
	"github.com/Bimidu/ctse-order-service/internal/handlers"
	"github.com/Bimidu/ctse-order-service/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	_ "github.com/Bimidu/ctse-order-service/docs"
)

// @title           CTSE Order Service API
// @version         1.0
// @description     Order, cart, and admin endpoints for the CTSE microsevices suite.
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	config.Load()

	if config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	database.Connect()

	clients.InitAuthClient()
	clients.InitProductClient()

	go grpcserver.Start()

	router := setupRouter()

	srv := &http.Server{
		Addr:         ":" + config.App.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Order Service HTTP listening on port %s", config.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "order-service",
			"status":  "running",
		})
	})

	cartH := handlers.NewCartHandler()
	orderH := handlers.NewOrderHandler()

	v1 := r.Group("/api/v1")
	v1.Use(middleware.AuthRequired())
	{
		cart := v1.Group("/cart")
		{
			cart.POST("/items", cartH.AddItem)
			cart.GET("", cartH.GetCart)
			cart.PUT("/items/:id", cartH.UpdateItem)
			cart.DELETE("/items/:id", cartH.RemoveItem)
			cart.DELETE("", cartH.ClearCart)
		}

		orders := v1.Group("/orders")
		{
			orders.POST("/checkout", orderH.Checkout)
			orders.GET("", orderH.GetMyOrders)
			orders.GET("/:id", orderH.GetOrder)
		}

		admin := v1.Group("/admin")
		admin.Use(middleware.RoleRequired("admin"))
		{
			admin.GET("/orders", orderH.GetAllOrders)
			admin.PUT("/orders/:id/status", orderH.UpdateOrderStatus)
		}
	}

	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	return r
}
