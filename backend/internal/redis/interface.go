package redis

import (
	"context"
	"quantalpha/internal/models"
)

type JobProducer interface {
	PublishJob(ctx context.Context, job models.JobPayload) error
	Close() error
}
