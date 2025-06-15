package aggr

import (
	"time"

	"marketflow/internal/domain"
)

type Aggregator struct {
	Input  <-chan domain.PriceUpdate
	Repo   *domain.PriceRepository
	Cache  *domain.Cache
	Window time.Duration
}

func NewAggregator(input <-chan domain.PriceUpdate, repo domain.PriceRepository, cache domain.Cache, window time.Duration) *Aggregator {
	return &Aggregator{
		Input:  input,
		Repo:   &repo,
		Cache:  &cache,
		Window: window,
	}
}
