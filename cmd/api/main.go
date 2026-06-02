package main

import (
	"log"
	"golang-worker-queue/internal/config"
	"golang-worker-queue/internal/database"
	"golang-worker-queue/internal/delivery/http"
	"golang-worker-queue/internal/logger"
	"golang-worker-queue/internal/queue"
	"golang-worker-queue/internal/repository"
	"golang-worker-queue/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.InitLogger(cfg.App.Environment)

	db, err := database.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	jobRepo := repository.NewPostgresJobRepository(db)
	redisQueue := queue.NewRedisQueue(redisClient)

	jobService := service.NewJobService(jobRepo, redisQueue)
	jobHandler := http.NewJobHandler(jobService)

	router := http.SetupRouter(jobHandler)

	log.Printf("Starting API server on port %s", cfg.App.Port)
	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
