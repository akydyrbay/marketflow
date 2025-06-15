package mode

import (
	"context"
	"sync"

	"marketflow/internal/domain"
	"marketflow/pkg/config"
)

type Mode string

const (
	Live Mode = "live"
	Test Mode = "test"
)

type Manager struct {
	mu         sync.Mutex
	mode       Mode
	clients    []domain.ExchangeClient
	cancelFunc context.CancelFunc
	cfg        *config.Config
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		mode: Test,
		cfg:  cfg,
	}
}
