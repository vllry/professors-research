package detailedcardcache

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/vllry/professors-research/pkg/types"
)

// DetailedCardCache provides read-only access to a cache of DetailedCards,
// with multiple views.
type DetailedCardCache struct {
	setNumberCache *setNumberCache
	nameLookup     *nameToDetailedCards
	mu             sync.RWMutex
	loaded         bool
	loadErr        error
}

// IsLoaded returns true if the cache has finished loading.
func (c *DetailedCardCache) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}

// LoadError returns any error that occurred during loading, or nil if loading succeeded or hasn't completed.
func (c *DetailedCardCache) LoadError() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loadErr
}

// GetBySetAndNumber returns a DetailedCard for the given set code and card number.
// Returns nil if the card is not found or if the cache hasn't finished loading.
func (c *DetailedCardCache) GetBySetAndNumber(setCode string, cardNumber int) types.DetailedCard {
	c.mu.RLock()
	loaded := c.loaded
	c.mu.RUnlock()
	if !loaded {
		return nil
	}
	return c.setNumberCache.get(setCode, cardNumber)
}

// Get returns a DetailedCard for the given set code and card number.
// This is a convenience method that calls GetBySetAndNumber.
func (c *DetailedCardCache) Get(setCode string, cardNumber int) types.DetailedCard {
	return c.GetBySetAndNumber(setCode, cardNumber)
}

// GetByName returns all DetailedCards with the given name.
func (c *DetailedCardCache) GetByName(name string) []types.DetailedCard {
	c.mu.RLock()
	loaded := c.loaded
	c.mu.RUnlock()
	if !loaded {
		return nil
	}
	return c.nameLookup.get(name)
}

// CreateDeduplicatedDecklist creates a new Decklist with duplicate cards (e.g. different prints) combined.
func (c *DetailedCardCache) CreateDeduplicatedDecklist(decklist types.Decklist) types.Decklist {
	deduplicatedDecklist := types.Decklist{
		Cards: make(map[types.Card]int),
	}

	// For each card in the decklist, check if it's a reprint of another card in the list.
	// We should decide which card to use out of the multiple prints in a deterministic way.
	cardsByName := make(map[string][]types.Card) // Name -> []Card
	cardCountByName := make(map[string]int)      // Name -> Count
	for card, count := range decklist.Cards {
		cardsByName[card.Name] = append(cardsByName[card.Name], card)
		cardCountByName[card.Name] += count
	}

	// Consolidate each set of name-identical cards into the distinct subset of cards.
	for name, sameNameCards := range cardsByName {
		// If there's only one print of this card, it's already deduplicated.
		if len(sameNameCards) == 1 {
			deduplicatedDecklist.Cards[sameNameCards[0]] = cardCountByName[name]
		} else {
			// We can use the DetailedCardCache to check if the card is a trainer or energy. This impacts how we deduplicate.
			// We can assume that all cards with the same name are of the same kind.
			detailedCards := c.GetByName(name)
			for _, detailedCard := range detailedCards {
				if detailedCard.Kind() == types.KindTrainer || detailedCard.Kind() == types.KindEnergy {
					// For Trainer and Energy cards, we can assume:
					// 1. Names will never be reused between different kinds.
					// 2. If there's a standard legal print of the card, all prints accross time are valid.
					// I'm not actually sure about the last one, e.g. versions of Great Ball with fundamentally different effects - is that more than an errata?
					deduplicatedDecklist.Cards[sameNameCards[0]] = cardCountByName[name] // Aggregate all cards under the 1st print's card. TODO: pick a deterministic way to do this.
					break
				} else if detailedCard.Kind() == types.KindPokemon {
					// For Pokemon cards, we need to check if all gameplay-affecting fields are the same between a pair of same-name cards.
					// We can do this by loading the DetailedCards and comparing.

					// TODO: need comparison method on DetailedCard
					// Need to do the n^2 comparison (or better?) for all pairs of same-name cards.
				} else {
					// For unknown card kinds, we can't deduplicate.
					// TODO log
					deduplicatedDecklist.Cards[sameNameCards[0]] = cardCountByName[name]
				}
			}

		}

	}

	return deduplicatedDecklist
}

// NewDetailedCardCache creates a new empty cache and starts loading data asynchronously.
func NewDetailedCardCache(dataDir string) *DetailedCardCache {
	cache := &DetailedCardCache{
		setNumberCache: newSetNumberCache(),
		nameLookup:     newNameToDetailedCards(),
		loaded:         false,
	}

	// Start loading in background
	go cache.load(dataDir)

	return cache
}

// load performs the actual loading of cards from the data directory.
// This method is called asynchronously and updates the cache state when complete.
func (c *DetailedCardCache) load(dataDir string) {
	cardsDir := filepath.Join(dataDir, "cards")
	entries, err := os.ReadDir(cardsDir)
	if err != nil {
		c.mu.Lock()
		c.loadErr = fmt.Errorf("failed to read cards directory: %w", err)
		c.loaded = true
		c.mu.Unlock()
		return
	}

	tempCache := make(map[string]map[int]types.DetailedCard)
	tempNameMap := make(map[string][]*types.DetailedCard)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(cardsDir, entry.Name())
		file, err := os.Open(filePath)
		if err != nil {
			c.mu.Lock()
			c.loadErr = fmt.Errorf("failed to open file %s: %w", filePath, err)
			c.loaded = true
			c.mu.Unlock()
			return
		}

		cards, err := types.NewDetailedCardsFromJSON(file)
		file.Close()
		if err != nil {
			c.mu.Lock()
			c.loadErr = fmt.Errorf("failed to load cards from %s: %w", filePath, err)
			c.loaded = true
			c.mu.Unlock()
			return
		}

		for _, card := range cards {
			setCode := card.SetCode()
			numberStr := card.Number()

			// Parse card number as int
			cardNumber, err := strconv.Atoi(numberStr)
			if err != nil {
				// Skip cards with non-numeric numbers
				continue
			}

			// Initialize set map if it doesn't exist
			if tempCache[setCode] == nil {
				tempCache[setCode] = make(map[int]types.DetailedCard)
			}

			tempCache[setCode][cardNumber] = card
			tempNameMap[card.Name()] = append(tempNameMap[card.Name()], &card) // Insert a reference to the card
		}
	}

	// Bulk-add both caches atomically
	c.setNumberCache.setMap(tempCache)
	c.nameLookup.setMap(tempNameMap)

	// Update loading state atomically once loading is complete
	c.mu.Lock()
	c.loaded = true
	c.mu.Unlock()
}

// NewEmptyLoadedCache creates a new empty cache that is marked as loaded.
// This is useful for testing purposes.
func NewEmptyLoadedCache() *DetailedCardCache {
	cache := &DetailedCardCache{
		setNumberCache: newSetNumberCache(),
		nameLookup:     newNameToDetailedCards(),
		loaded:         true,
	}
	return cache
}
