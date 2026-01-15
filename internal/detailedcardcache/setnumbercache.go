// This file contains the main cache of DetailedCards.

package detailedcardcache

import (
	"sync"

	"github.com/vllry/professors-research/pkg/types"
)

// setNumberCache is a cache that maps set codes to a map of card numbers to DetailedCards.
type setNumberCache struct {
	mu    sync.RWMutex
	cache map[string]map[int]types.DetailedCard
}

// newSetNumberCache creates a new setNumberCache.
func newSetNumberCache() *setNumberCache {
	return &setNumberCache{
		cache: make(map[string]map[int]types.DetailedCard),
	}
}

// get returns a DetailedCard for the given set code and card number.
// Returns nil if the card is not found.
func (c *setNumberCache) get(setCode string, cardNumber int) types.DetailedCard {
	c.mu.RLock()
	defer c.mu.RUnlock()
	setMap, exists := c.cache[setCode]
	if !exists {
		return nil
	}
	return setMap[cardNumber]
}

// setMap replaces the current cache with the given map.
func (c *setNumberCache) setMap(newMap map[string]map[int]types.DetailedCard) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = newMap
}
