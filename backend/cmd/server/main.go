package main

import (
	"log"

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

	authHandler := handlers.NewAuthHandler(queries, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userHandler := handlers.NewUserHandler(queries)
	alphaHandler := handlers.NewAlphaHandler(queries)
	factorHandler := handlers.NewFactorHandler(queries)
	backtestHandler := handlers.NewBacktestHandler(queries, prod)
	modelHandler := handlers.NewModelHandler(queries, prod)
	analyticsHandler := handlers.NewAnalyticsHandler(queries)
	adminHandler := handlers.NewAdminHandler(queries)

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

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
