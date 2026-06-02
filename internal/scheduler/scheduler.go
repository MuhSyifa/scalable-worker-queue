package scheduler

import (
	"context"
	"fmt"
	"golang-worker-queue/internal/cache"
	"golang-worker-queue/internal/queue"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Scheduler struct {
	client     *redis.Client
	redisQueue *queue.RedisQueue
	lock       *cache.DistributedLock
}

func NewScheduler(client *redis.Client, redisQueue *queue.RedisQueue, lock *cache.DistributedLock) *Scheduler {
	return &Scheduler{
		client:     client,
		redisQueue: redisQueue,
		lock:       lock,
	}
}

// Start runs a background loop to check for scheduled jobs
func (s *Scheduler) Start(ctx context.Context) {
	log.Info().Msg("Starting job scheduler")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stopping job scheduler")
			return
		case <-ticker.C:
			s.processScheduledJobs(ctx)
		}
	}
}

func (s *Scheduler) processScheduledJobs(ctx context.Context) {
	// Try to acquire distributed lock to prevent multiple instances from moving the same jobs
	lockKey := "scheduler:move_jobs"
	acquired, err := s.lock.Acquire(ctx, lockKey, 5*time.Second)
	if err != nil || !acquired {
		return // Failed to acquire lock or already locked by another instance
	}
	defer s.lock.Release(ctx, lockKey)

	now := time.Now().Unix()

	// 1. Get all jobs scheduled for now or in the past
	// Note: In a real system, we'd use a Lua script for atomicity (ZRangeByScore + ZRem + LPush)
	// For simplicity, we fetch, enqueue, then remove here.
	opt := &redis.ZRangeBy{
		Min: "0",
		Max: float64ToString(float64(now)),
	}

	jobs, err := s.client.ZRangeByScore(ctx, queue.ScheduledQueue, opt).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch scheduled jobs")
		return
	}

	for _, jobData := range jobs {
		// Moving from ZSET to List
		// Assuming Priority 0 for simplicity. A full implementation would parse the JSON and use the priority
		qName := queue.QueuePrefix + "0"
		pipe := s.client.Pipeline()
		pipe.ZRem(ctx, queue.ScheduledQueue, jobData)
		pipe.LPush(ctx, qName, jobData)

		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to move scheduled job to active queue")
		} else {
			log.Debug().Msg("Moved scheduled job to active queue")
		}
	}
}

func float64ToString(f float64) string {
	return string(fmt.Appendf(nil, "%f", f))
}
