package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/corradoisidoro/sentinel-rbac/internal/config"
	"github.com/corradoisidoro/sentinel-rbac/internal/handler"
	"github.com/corradoisidoro/sentinel-rbac/internal/middleware"
	"github.com/corradoisidoro/sentinel-rbac/internal/repository"
	"github.com/corradoisidoro/sentinel-rbac/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("database URL cannot be empty")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT secret cannot be empty")
	}

	// Connect to Database
	log.Println("[INFO] connecting to database...")

	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql db: %v", err)
	}
	defer sqlDB.Close()

	// Run Migrations
	log.Println("[INFO] running migrations...")

	if err := repository.Migrate(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// Initialize Layers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg)
	userHandler := handler.NewUserHandler(userService)

	jwtSecret := []byte(cfg.JWTSecret)
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret, userRepo)

	// Setup Router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Add rate limiter (production defaults – global safety limiter)
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		GlobalRPS:   500,
		GlobalBurst: 1000,

		IPRPS:   20,
		IPBurst: 40,

		RouteRPS:   100,
		RouteBurst: 200,
	})

	router.Use(rateLimiter)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
		auth.POST("/logout", authMiddleware.RequireAuth, userHandler.Logout)
	}

	// User routes
	users := api.Group("/users")
	users.Use(authMiddleware.RequireAuth)
	{
		users.GET("/profile", userHandler.Profile)
		users.GET("/admin", authMiddleware.AuthorizeRole("admin"), userHandler.Admin)
	}

	// Start Server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("[INFO] starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("[INFO] server exited properly")
}
