package repository

import (
	"fmt"
	"sync"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

type InMemoryHoldingRepository struct {
	mu       sync.RWMutex
	holdings map[string]*domain.Holding // key: userID:symbol
}

func NewInMemoryHoldingRepository() *InMemoryHoldingRepository {
	return &InMemoryHoldingRepository{
		holdings: make(map[string]*domain.Holding),
	}
}

func (r *InMemoryHoldingRepository) Upsert(holding *domain.Holding) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", holding.UserID, holding.Symbol)
	r.holdings[key] = holding
	return nil
}

func (r *InMemoryHoldingRepository) GetByUserAndSymbol(userID, symbol string) (*domain.Holding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userID, symbol)
	holding, exists := r.holdings[key]
	if !exists {
		return nil, fmt.Errorf("holding not found")
	}
	return holding, nil
}

func (r *InMemoryHoldingRepository) ListByUser(userID string) ([]*domain.Holding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var holdings []*domain.Holding
	for _, holding := range r.holdings {
		if holding.UserID == userID {
			holdings = append(holdings, holding)
		}
	}
	return holdings, nil
}
