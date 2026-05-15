package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"quantalpha/internal/config"
	"quantalpha/internal/database"
	"quantalpha/internal/handlers"
	"quantalpha/internal/middleware"
	"quantalpha/internal/models"
	"quantalpha/internal/redis"
	"quantalpha/internal/validator"
)

func main() {
	cfg := config.Load()

	gin.SetMode(cfg.Server.Mode)

	db, err := database.NewDB(nil, &cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	queries := database.NewQueries(db.Pool)

	prod, err := redis.NewProducer(&cfg.Redis)
	if err != nil {
		log.Printf("warning: failed to connect to redis: %v", err)
	}
	if prod != nil {
		defer prod.Close()
	}

	validator.Init()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	authHandler := handlers.NewAuthHandler(queries, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userHandler := handlers.NewUserHandler(queries)
	alphaHandler := handlers.NewAlphaHandler(queries)
	factorHandler := handlers.NewFactorHandler(queries)
	backtestHandler := handlers.NewBacktestHandler(queries, prod)
	modelHandler := handlers.NewModelHandler(queries, prod)
	analyticsHandler := handlers.NewAnalyticsHandler(queries)
	adminHandler := handlers.NewAdminHandler(queries)

	systemHandler := handlers.NewSystemHandler(queries, prod)

	r.GET("/health", systemHandler.Health)
	r.GET("/ready", systemHandler.Ready)
	r.POST("/auth/login", authHandler.Login)

	private := r.Group("")
	private.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		private.GET("/models", middleware.RBACMiddleware(models.RoleAdmin, models.RoleQR, models.RolePM, models.RoleDS), modelHandler.ListModels)
		private.GET("/factors", middleware.RBACMiddleware(models.RoleAdmin, models.RoleQR, models.RolePM, models.RoleDS), factorHandler.ListFactors)
		private.GET("/me/profile", userHandler.GetProfile)

		qr := private.Group("")
		qr.Use(middleware.RBACMiddleware(models.RoleAdmin, models.RoleQR))
		{
			qr.GET("/alphas/me", alphaHandler.ListMyAlphas)
			qr.GET("/alphas/me/submitted", alphaHandler.ListMySubmittedAlphas)
			qr.POST("/alphas", alphaHandler.CreateAlpha)
			qr.PUT("/alphas/:id", alphaHandler.UpdateAlpha)
			qr.POST("/alphas/:id/submit", alphaHandler.SubmitAlpha)
			qr.DELETE("/alphas/:id", alphaHandler.DeleteAlpha)
			qr.POST("/backtest/run", backtestHandler.RunBacktest)
			qr.GET("/backtest/me/status", backtestHandler.GetMyBacktestStatus)
			qr.GET("/backtest/:job_id", backtestHandler.GetBacktestStatus)
			qr.GET("/factors/:id/preview", factorHandler.GetFactorPreview)
		}

		ds := private.Group("")
		ds.Use(middleware.RBACMiddleware(models.RoleAdmin, models.RoleDS))
		{
			ds.POST("/factors/publish", factorHandler.CreateFactor)
			ds.PUT("/factors/:id", factorHandler.UpdateFactor)
			ds.DELETE("/factors/:id", factorHandler.DeleteFactor)
			ds.POST("/models/train", modelHandler.TrainModel)
			ds.DELETE("/models/:id", modelHandler.DeleteModel)
		}

		pm := private.Group("")
		pm.Use(middleware.RBACMiddleware(models.RoleAdmin, models.RolePM))
		{
			pm.GET("/alphas/submitted", alphaHandler.ListSubmittedAlphas)
			pm.GET("/backtest/status", backtestHandler.ListAllBacktestStatus)
			pm.GET("/analytics/correlation", analyticsHandler.GetCorrelation)
			pm.GET("/analytics/performance", analyticsHandler.GetPerformance)
			pm.GET("/audit-logs", analyticsHandler.GetAuditLogs)
		}

		admin := private.Group("/admin")
		admin.Use(middleware.AdminOnlyMiddleware())
		{
			admin.GET("/users", userHandler.ListUsers)
			admin.POST("/users", userHandler.CreateUser)
			admin.PATCH("/users/:id", userHandler.UpdateUser)
			admin.DELETE("/users/:id", userHandler.DeleteUser)
			admin.DELETE("/jobs/:job_id", adminHandler.DeleteJob)
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

