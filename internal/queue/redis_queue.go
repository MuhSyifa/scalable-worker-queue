package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-worker-queue/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	QueuePrefix    = "queue:"
	ScheduledQueue = "scheduled:jobs"
	ActiveQueue    = "active:jobs" // Used for reliable queue pattern (RPOPLPUSH)
)

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

// Enqueue adds a job to the appropriate queue
func (q *RedisQueue) Enqueue(ctx context.Context, job *domain.Job) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
		// Add to scheduled set
		return q.client.ZAdd(ctx, ScheduledQueue, redis.Z{
			Score:  float64(job.ScheduledAt.Unix()),
			Member: string(jobData),
		}).Err()
	}

	// Add to normal queue based on priority
	queueName := fmt.Sprintf("%s%d", QueuePrefix, job.Priority)
	return q.client.LPush(ctx, queueName, string(jobData)).Err()
}

// Dequeue blocks until a job is available in any of the queues
func (q *RedisQueue) Dequeue(ctx context.Context, queues []string, timeout time.Duration) (*domain.Job, error) {
	// We listen on multiple queues for priority support. 
	// E.g., BLPOP queue:10 queue:5 queue:0
	keys := make([]string, len(queues))
	for i, qName := range queues {
		keys[i] = fmt.Sprintf("%s%d", QueuePrefix, qName) // Assume qName is priority for now, actually better to just pass the raw queue names
	}
	
	// Better approach: the caller provides the exact queue names ordered by priority
	result, err := q.client.BLPop(ctx, timeout, queues...).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Timeout
		}
		return nil, err
	}

	// result[0] is queue name, result[1] is the popped value
	var job domain.Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}
