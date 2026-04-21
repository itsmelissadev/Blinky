package worker

import (
	"context"
	"sync"

	"blinky/internal/pkg/logger"
)

type Task func(ctx context.Context) error

type Manager struct {
	tasks []Task
	wg    sync.WaitGroup
	mu    sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		tasks: make([]Task, 0),
	}
}

func (m *Manager) Register(task Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = append(m.tasks, task)
}

func (m *Manager) Start(ctx context.Context) {
	for _, t := range m.tasks {
		m.wg.Add(1)
		go func(task Task) {
			defer m.wg.Done()
			if err := task(ctx); err != nil {
				logger.Error("[WORKER] Task failed: %v", err)
			}
		}(t)
	}
}

func (m *Manager) Wait() {
	m.wg.Wait()
}
