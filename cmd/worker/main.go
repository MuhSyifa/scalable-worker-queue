package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang-worker-queue/internal/cache"
	"golang-worker-queue/internal/config"
	"golang-worker-queue/internal/database"
	"golang-worker-queue/internal/domain"
	"golang-worker-queue/internal/logger"
	"golang-worker-queue/internal/queue"
	"golang-worker-queue/internal/repository"
	"golang-worker-queue/internal/scheduler"
	"golang-worker-queue/internal/worker"
	"golang-worker-queue/internal/worker/handlers"
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
	distLock := cache.NewDistributedLock(redisClient)

	// Setup Processor & Handlers
	processor := worker.NewProcessor()
	processor.RegisterHandler(domain.JobTypeEmail, handlers.NewEmailHandler())
	processor.RegisterHandler(domain.JobTypePayment, handlers.NewPaymentHandler())
	// Register other handlers as needed

	// Setup Scheduler
	jobScheduler := scheduler.NewScheduler(redisClient, redisQueue, distLock)

	// Setup Worker Pool
	workerPool := worker.NewWorkerPool(cfg.Worker.Concurrency, redisQueue, jobRepo, processor)

	ctx, cancel := context.WithCancel(context.Background())

	// Start components
	go jobScheduler.Start(ctx)
	workerPool.Start(ctx)

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down worker process...")

	cancel() // Cancels context for scheduler and workers
	workerPool.Stop() // Waits for active jobs to complete
	
	log.Println("Worker process stopped gracefully.")
}
