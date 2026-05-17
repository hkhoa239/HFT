package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"quantalpha/internal/config"
	"quantalpha/internal/models"
)

type Producer struct {
	client *redis.Client
}

func NewProducer(cfg *config.RedisConfig) (*Producer, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + fmt.Sprintf("%d", cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Producer{client: client}, nil
}

func (p *Producer) PublishJob(ctx context.Context, job models.JobPayload) error {
	payload := map[string]interface{}{
		"job_id":     job.JobID,
		"task_type":  job.TaskType,
		"user_id":    job.UserID,
		"alpha_id":   job.AlphaID,
		"script":     job.Script,
		"params":     job.Params,
		"created_at": job.CreatedAt,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: "job_queue",
		Values: map[string]interface{}{
			"payload": string(data),
		},
	}).Err()
}

func NewJobPayload(taskType, userID, alphaID, script string, params map[string]interface{}) models.JobPayload {
	return models.JobPayload{
		JobID:     uuid.New().String(),
		TaskType:  taskType,
		UserID:    userID,
		AlphaID:   alphaID,
		Script:    script,
		Params:    params,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (p *Producer) Close() error {
	return p.client.Close()
}
