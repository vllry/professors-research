package types

// CardSet is an interface representing a nonfixed combination of cards.
// It can be expanded into all possible combinations given a decklist.
type CardSet interface {
	Expand(decklist Decklist) CardSetExpanded
}

// CardSetExpanded represents all possible combinations of a CardSet.
// E.g. CardSet [AnyOf(1 Psychic Energy, 1 Fire Energy)] -> CardSetExpanded [1 Psychic Energy, 1 Fire Energy]
type CardSetExpanded struct {
	Combinations [][]Card
}

// AnyOfPattern represents an inclusive "any of these" pattern, such as "1 Psychic Energy or 1 Fire Energy".
type AnyOfPattern struct {
	Cards map[Card]int // Card -> count
}

// AnyOf is a CardSet implementation that represents a combination where
// one card must be chosen from each AnyOfPattern.
// This is the same behavior as the original CardSet.
// E.g. [AnyOfPattern(1 X, 1 Y), AnyOfPattern(1 Y, 1 Z)] means:
// - Choose one from {1 X, 1 Y} AND choose one from {1 Y, 1 Z}
// - Possible combinations: [1 X, 1 Y], [1 X, 1 Z], [1 Y, 1 Y], [1 Y, 1 Z]
type AnyOf struct {
	Patterns []AnyOfPattern
}

// Expand expands an AnyOf CardSet into all possible combinations.
func (a AnyOf) Expand(decklist Decklist) CardSetExpanded {
	var combinations [][]Card

	// Generate all possible combinations by taking one card from each AnyOfPattern
	a.generateCombinations(0, []Card{}, &combinations, decklist)

	return CardSetExpanded{
		Combinations: combinations,
	}
}

// generateCombinations recursively generates all valid combinations
func (a AnyOf) generateCombinations(patternIndex int, currentCombination []Card, combinations *[][]Card, decklist Decklist) {
	// Base case: we've processed all AnyOfPatterns
	if patternIndex >= len(a.Patterns) {
		// Check if this combination is valid (all card counts are within decklist limits)
		if a.isValidCombination(currentCombination, decklist) {
			// Make a copy of the combination to avoid aliasing issues
			combinationCopy := make([]Card, len(currentCombination))
			copy(combinationCopy, currentCombination)
			*combinations = append(*combinations, combinationCopy)
		}
		return
	}

	// For each card in the current AnyOfPattern, add it to the combination and recurse
	currentPattern := a.Patterns[patternIndex]
	for card, count := range currentPattern.Cards {
		// Add this card 'count' times to the current combination
		newCombination := make([]Card, len(currentCombination), len(currentCombination)+count)
		copy(newCombination, currentCombination)
		for i := 0; i < count; i++ {
			newCombination = append(newCombination, card)
		}
		a.generateCombinations(patternIndex+1, newCombination, combinations, decklist)
	}
}

// isValidCombination checks if a combination is valid given the decklist
func (a AnyOf) isValidCombination(combination []Card, decklist Decklist) bool {
	// Count how many of each card are in the combination
	combinationCounts := make(map[Card]int)
	for _, card := range combination {
		combinationCounts[card]++
	}

	// Check if all card counts in the combination are within decklist limits
	for card, count := range combinationCounts {
		if decklist.Cards[card] < count {
			return false
		}
	}

	return true
}

// AllOf is a CardSet implementation that represents a combination where
// all specified cards must be present. It is semantically equivalent to a list of cards.
// E.g. AllOf([1 X, 1 Y, 1 Z]) means all three cards must be present.
type AllOf []Card

// Expand expands an AllOf CardSet into all possible combinations.
// Since AllOf requires all cards to be present, it returns a single combination
// containing all the cards (if valid), or an empty set if not valid.
func (a AllOf) Expand(decklist Decklist) CardSetExpanded {
	// Count how many of each card are needed
	requiredCounts := make(map[Card]int)
	for _, card := range a {
		requiredCounts[card]++
	}

	// Check if all required cards are available in the decklist
	for card, required := range requiredCounts {
		if decklist.Cards[card] < required {
			// Not enough copies in deck, return empty
			return CardSetExpanded{
				Combinations: [][]Card{},
			}
		}
	}

	// All cards are available, return the single combination
	// Make a copy to avoid aliasing issues
	combination := make([]Card, len(a))
	copy(combination, a)
	return CardSetExpanded{
		Combinations: [][]Card{combination},
	}
}

// NewCardSet creates an AnyOf CardSet from a list of AnyOfPatterns.
// This is a convenience function for backward compatibility.
func NewCardSet(patterns []AnyOfPattern) CardSet {
	return AnyOf{
		Patterns: patterns,
	}
}

// NewAnyOf creates an AnyOf CardSet from a list of AnyOfPatterns.
func NewAnyOf(patterns []AnyOfPattern) AnyOf {
	return AnyOf{
		Patterns: patterns,
	}
}
