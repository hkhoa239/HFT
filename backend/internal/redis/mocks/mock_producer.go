package mocks

import (
	"context"
	"quantalpha/internal/models"
	"sync"
)

type MockProducer struct {
	mu            sync.Mutex
	PublishedJobs []models.JobPayload
	ShouldFail    bool
}

func NewMockProducer() *MockProducer {
	return &MockProducer{
		PublishedJobs: []models.JobPayload{},
	}
}

func (m *MockProducer) PublishJob(ctx context.Context, job models.JobPayload) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ShouldFail {
		return context.DeadlineExceeded
	}
	m.PublishedJobs = append(m.PublishedJobs, job)
	return nil
}

func (m *MockProducer) Close() error {
	return nil
}
