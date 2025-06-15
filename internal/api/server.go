package api

import (
	"net/http"

	"marketflow/internal/adapters/cache"
	"marketflow/internal/app/mode"
	"marketflow/internal/domain"
)

type Server struct {
	repo    domain.PriceRepository
	cache   *cache.RedisCache
	manager *mode.Manager
}

func NewServer(repo domain.PriceRepository, cache *cache.RedisCache, manager *mode.Manager) *Server {
	return &Server{
		repo:    repo,
		cache:   cache,
		manager: manager,
	}
}

func (s *Server) Router(input chan<- domain.PriceUpdate) http.Handler {
	mux := http.NewServeMux()

	return mux
}
