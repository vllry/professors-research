package matchups

import (
	"encoding/json"
	"os"

	"github.com/vllry/professors-research/pkg/types"
)

type Archetype struct {
	Name     string         `json:"name"`
	Requires map[string]int `json:"requires"`
}

func LoadArchetypes(path string) ([]Archetype, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var archetypes []Archetype
	if err := json.Unmarshal(data, &archetypes); err != nil {
		return nil, err
	}
	return archetypes, nil
}

// ClassifyDeck returns the name of the first matching archetype, or "Unknown".
func ClassifyDeck(deck types.Decklist, archetypes []Archetype) string {
	counts := DeckCardNameCounts(deck)
	for _, arch := range archetypes {
		if deckMatchesRequirements(counts, arch.Requires) {
			return arch.Name
		}
	}
	return "Unknown"
}

// ClassifyVariants returns the stringified indices of all matching variants.
// If no variant matches, the result contains only "other".
// Decks matching multiple variants are included in all of them.
func ClassifyVariants(deck types.Decklist, variants []map[string]int) []string {
	counts := DeckCardNameCounts(deck)
	var matched []string
	for i, req := range variants {
		if deckMatchesRequirements(counts, req) {
			matched = append(matched, variantKey(i))
		}
	}
	if len(matched) == 0 {
		return []string{"other"}
	}
	return matched
}

func deckMatchesRequirements(counts map[string]int, requires map[string]int) bool {
	for cardName, minCount := range requires {
		if counts[cardName] < minCount {
			return false
		}
	}
	return true
}

// DeckCardNameCounts aggregates card counts by name across all printings.
func DeckCardNameCounts(deck types.Decklist) map[string]int {
	counts := make(map[string]int, len(deck.Cards))
	for card, count := range deck.Cards {
		counts[card.Name] += count
	}
	return counts
}
