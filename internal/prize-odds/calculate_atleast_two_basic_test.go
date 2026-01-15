package prizeodds

import (
	"math"
	"testing"
)

// TestCalculateAtLeastTwoBasic verifies that the function correctly calculates
// the probability of drawing at least 2 basic Pokemon (any combination, same or different).
func TestCalculateAtLeastTwoBasic(t *testing.T) {
	t.Run("ManualCalculation_13Basic", func(t *testing.T) {
		// Deck with 13 total basic Pokemon (from dragapult decklist)
		totalBasicCount := 13

		// P(0 basic) = C(47, 7) / C(60, 7)
		probZero := float64(comb(47, 7)) / float64(comb(60, 7))

		// P(1 basic) = C(13, 1) * C(47, 6) / C(60, 7)
		probOne := float64(comb(13, 1)) * float64(comb(47, 6)) / float64(comb(60, 7))

		// P(>=2 basic) = 1 - P(0 basic) - P(1 basic)
		expected := 1.0 - probZero - probOne

		result := CalculateAtLeastTwoBasic(totalBasicCount)
		if math.Abs(result-expected) > 0.0001 {
			t.Errorf("At least 2 basic: got %v, expected %v", result, expected)
		}
		t.Logf("At least 2 basic (13 total): %.4f%%", result*100)
	})

	t.Run("IncludesSameAndDifferentBasics", func(t *testing.T) {
		// Test with a deck that has multiple copies of the same basic
		// This verifies that "at least 2 basics" includes:
		// - 2 of the same basic (e.g., 2 Dreepy)
		// - 2 different basics (e.g., 1 Dreepy + 1 Duskull)
		// - Any combination totaling 2+ basic cards

		// Deck: 4 Dreepy, 2 Duskull, 2 Budew, 5 other singles = 13 total basic
		totalBasicCount := 13

		result := CalculateAtLeastTwoBasic(totalBasicCount)

		// The result should be the probability of drawing 2+ basic cards total,
		// regardless of which specific basic Pokemon they are
		// This includes all combinations:
		// - 2 Dreepy (same)
		// - 1 Dreepy + 1 Duskull (different)
		// - 1 Dreepy + 1 Budew (different)
		// - 2 Duskull (same)
		// - etc.

		// Verify it's between 0 and 1
		if result < 0.0 || result > 1.0 {
			t.Errorf("Probability out of bounds: %v", result)
		}

		// Verify it's less than "at least 1 basic" (which would be 1 - mulligan)
		probZero := float64(comb(47, 7)) / float64(comb(60, 7))
		atLeastOne := 1.0 - probZero
		if result >= atLeastOne {
			t.Errorf("At least 2 basic (%v) should be < at least 1 basic (%v)", result, atLeastOne)
		}

		t.Logf("At least 2 basic (any combination): %.4f%%", result*100)
		t.Logf("At least 1 basic: %.4f%%", atLeastOne*100)
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Less than 2 basic in deck
		result := CalculateAtLeastTwoBasic(1)
		if result != 0.0 {
			t.Errorf("With only 1 basic in deck, at least 2 should be 0, got %v", result)
		}

		// Exactly 2 basic in deck
		result = CalculateAtLeastTwoBasic(2)
		// Should be: 1 - P(0) - P(1) = 1 - C(58, 7)/C(60, 7) - C(2, 1)*C(58, 6)/C(60, 7)
		probZero := float64(comb(58, 7)) / float64(comb(60, 7))
		probOne := float64(comb(2, 1)) * float64(comb(58, 6)) / float64(comb(60, 7))
		expected := 1.0 - probZero - probOne
		if math.Abs(result-expected) > 0.0001 {
			t.Errorf("With 2 basic in deck: got %v, expected %v", result, expected)
		}

		// Many basic in deck (should approach 1.0)
		result = CalculateAtLeastTwoBasic(30)
		if result < 0.9 {
			t.Errorf("With 30 basic in deck, at least 2 should be very high, got %v", result)
		}
	})
}

