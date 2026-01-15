package detailedcardcache

import (
	"sync"

	"github.com/vllry/professors-research/pkg/types"
)

// nameToDetailedCards is an internal cache of card names to DetailedCard pointers.
type nameToDetailedCards struct {
	mu    sync.RWMutex
	cache map[string][]*types.DetailedCard
}

// setMap replaces the current map with the given map.
func (c *nameToDetailedCards) setMap(newMap map[string][]*types.DetailedCard) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = newMap
}

// newNameToDetailedCards creates a new empty nameToDetailedCards cache.
func newNameToDetailedCards() *nameToDetailedCards {
	return &nameToDetailedCards{
		cache: make(map[string][]*types.DetailedCard),
	}
}

// get returns all DetailedCards with the given name.
func (c *nameToDetailedCards) get(name string) []types.DetailedCard {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cards, exists := c.cache[name]
	if !exists {
		return nil
	}
	// Convert []*types.DetailedCard to []types.DetailedCard
	result := make([]types.DetailedCard, len(cards))
	for i, card := range cards {
		result[i] = *card
	}
	return result
}
