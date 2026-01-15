package prizeodds

import (
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

// TestAtLeastSemantics verifies that the function correctly calculates "at least" probabilities
// rather than "exactly" probabilities.
func TestAtLeastSemantics(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "A", SetCode: "TEST", Number: "1"}: 4,
		},
	}

	// CardSet: [AnyOf(1 A)] - meaning we need at least 1 A
	cardSet := basictypes.NewCardSet([]basictypes.AnyOfPattern{
		{
			Cards: map[basictypes.Card]int{
				{Name: "A", SetCode: "TEST", Number: "1"}: 1,
			},
		},
	})

	result, err := CalculateCardSetPrizeOdds(decklist, map[string]basictypes.CardSet{"test": cardSet}, true)
	if err != nil {
		t.Fatalf("CalculateCardSetPrizeOdds failed: %v", err)
	}

	// P(at least 1 A) = 1 - P(0 A) = 1 - C(56, 6) / C(60, 6)
	expectedAtLeast1A := 1.0 - float64(comb(56, 6)) / float64(comb(60, 6))

	t.Logf("Result: %v", result["test"])
	t.Logf("Expected P(at least 1 A): %v", expectedAtLeast1A)

	tolerance := 0.0001
	diff := result["test"] - expectedAtLeast1A
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("Result should equal P(at least 1 A), got %v, expected %v (diff: %v)", result["test"], expectedAtLeast1A, diff)
	}

	// Verify it's NOT equal to P(exactly 1 A)
	// P(exactly 1 A) = C(4, 1) * C(56, 5) / C(60, 6)
	expectedExactly1A := float64(comb(4, 1) * comb(56, 5)) / float64(comb(60, 6))
	
	t.Logf("P(exactly 1 A): %v", expectedExactly1A)
	t.Logf("P(at least 1 A): %v", expectedAtLeast1A)
	t.Logf("Difference: %v", expectedAtLeast1A - expectedExactly1A)

	// The result should be greater than P(exactly 1 A) because it includes cases with 2, 3, 4 A's
	if result["test"] <= expectedExactly1A {
		t.Errorf("P(at least 1 A) should be greater than P(exactly 1 A), but got %v <= %v", result["test"], expectedExactly1A)
	}
}

// TestAtLeastSemantics_MultipleCards verifies "at least" semantics for combinations with multiple cards.
func TestAtLeastSemantics_MultipleCards(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "A", SetCode: "TEST", Number: "1"}: 4,
			{Name: "B", SetCode: "TEST", Number: "2"}: 4,
		},
	}

	// Combination [A, B] should mean "at least 1 A and at least 1 B"
	// This should include cases like [A, A, B], [A, B, B], [A, A, B, B], etc.
	comboAB := []basictypes.Card{
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "B", SetCode: "TEST", Number: "2"},
	}

	probAtLeastAB := calculateCombinationProbability(comboAB, decklist, 6)

	// Calculate P(exactly 1 A and exactly 1 B) for comparison
	probExactly1A1B := float64(comb(4, 1) * comb(4, 1) * comb(52, 4)) / float64(comb(60, 6))

	t.Logf("P(at least 1 A and at least 1 B): %v", probAtLeastAB)
	t.Logf("P(exactly 1 A and exactly 1 B): %v", probExactly1A1B)
	t.Logf("Difference: %v", probAtLeastAB - probExactly1A1B)

	// P(at least) should be greater than P(exactly) because it includes cases with more cards
	if probAtLeastAB <= probExactly1A1B {
		t.Errorf("P(at least 1 A and 1 B) should be greater than P(exactly 1 A and 1 B), but got %v <= %v", probAtLeastAB, probExactly1A1B)
	}

	// Verify it's reasonable: P(at least) should be less than 1.0
	if probAtLeastAB >= 1.0 {
		t.Errorf("P(at least 1 A and 1 B) should be less than 1.0, got %v", probAtLeastAB)
	}

	// Verify it's positive
	if probAtLeastAB <= 0.0 {
		t.Errorf("P(at least 1 A and 1 B) should be positive, got %v", probAtLeastAB)
	}
}

// TestAtLeastSemantics_Superset verifies that supersets satisfy the requirement.
// If we need [A, B], then [A, A, B] should satisfy it.
func TestAtLeastSemantics_Superset(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "A", SetCode: "TEST", Number: "1"}: 4,
			{Name: "B", SetCode: "TEST", Number: "2"}: 4,
		},
	}

	// Combination [A, B] - need at least 1 A and at least 1 B
	comboAB := []basictypes.Card{
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "B", SetCode: "TEST", Number: "2"},
	}

	probAB := calculateCombinationProbability(comboAB, decklist, 6)

	// Combination [A, A, B] - need at least 2 A and at least 1 B
	comboAAB := []basictypes.Card{
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "B", SetCode: "TEST", Number: "2"},
	}

	probAAB := calculateCombinationProbability(comboAAB, decklist, 6)

	t.Logf("P(at least 1 A and 1 B): %v", probAB)
	t.Logf("P(at least 2 A and 1 B): %v", probAAB)

	// P(at least 2 A and 1 B) should be less than P(at least 1 A and 1 B)
	// because it's a more restrictive requirement
	if probAAB >= probAB {
		t.Errorf("P(at least 2 A and 1 B) should be less than P(at least 1 A and 1 B), but got %v >= %v", probAAB, probAB)
	}

	// But both should be positive
	if probAB <= 0.0 || probAAB <= 0.0 {
		t.Errorf("Both probabilities should be positive, got P(AB)=%v, P(AAB)=%v", probAB, probAAB)
	}
}

