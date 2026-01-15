package prizeodds

import (
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

// TestTotalCardsNeeded_Bug tests whether totalCardsNeeded is calculated correctly
// when the same card appears in multiple combinations.
func TestTotalCardsNeeded_Bug(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "A", SetCode: "TEST", Number: "1"}: 4,
			{Name: "B", SetCode: "TEST", Number: "2"}: 4,
			{Name: "C", SetCode: "TEST", Number: "3"}: 4,
		},
	}

	// Three combinations that overlap in complex ways:
	// [A, B] - needs 1 A, 1 B
	// [A, A, C] - needs 2 A, 1 C
	// [B, C] - needs 1 B, 1 C
	combinations := [][]basictypes.Card{
		{{Name: "A", SetCode: "TEST", Number: "1"}, {Name: "B", SetCode: "TEST", Number: "2"}},
		{{Name: "A", SetCode: "TEST", Number: "1"}, {Name: "A", SetCode: "TEST", Number: "1"}, {Name: "C", SetCode: "TEST", Number: "3"}},
		{{Name: "B", SetCode: "TEST", Number: "2"}, {Name: "C", SetCode: "TEST", Number: "3"}},
	}

	// Intersection of all three: max(1,2) A, max(1,1) B, max(1,1) C = [A,A,B,C] = 4 cards
	// This should be valid (4 <= 6)

	result := CalculateUnionProbability(combinations, decklist, true)
	
	// The result should be a valid probability
	if result < 0 || result > 1 {
		t.Errorf("Result should be between 0 and 1, got %v", result)
	}

	// Manual calculation to verify:
	// P([A,B]) + P([A,A,C]) + P([B,C]) 
	// - P([A,B] ∩ [A,A,C]) - P([A,B] ∩ [B,C]) - P([A,A,C] ∩ [B,C])
	// + P([A,B] ∩ [A,A,C] ∩ [B,C])
	
	probAB := calculateCombinationProbability(combinations[0], decklist, 6)
	probAAC := calculateCombinationProbability(combinations[1], decklist, 6)
	probBC := calculateCombinationProbability(combinations[2], decklist, 6)
	
	// Intersections
	intersectionAB_AAC := []basictypes.Card{
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "B", SetCode: "TEST", Number: "2"},
		{Name: "C", SetCode: "TEST", Number: "3"},
	}
	probAB_AAC := calculateCombinationProbability(intersectionAB_AAC, decklist, 6)
	
	intersectionAB_BC := []basictypes.Card{
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "B", SetCode: "TEST", Number: "2"},
		{Name: "C", SetCode: "TEST", Number: "3"},
	}
	probAB_BC := calculateCombinationProbability(intersectionAB_BC, decklist, 6)
	
	intersectionAAC_BC := []basictypes.Card{
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "B", SetCode: "TEST", Number: "2"},
		{Name: "C", SetCode: "TEST", Number: "3"},
	}
	probAAC_BC := calculateCombinationProbability(intersectionAAC_BC, decklist, 6)
	
	// Triple intersection: [A,A,B,C]
	intersectionAll := []basictypes.Card{
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "A", SetCode: "TEST", Number: "1"},
		{Name: "B", SetCode: "TEST", Number: "2"},
		{Name: "C", SetCode: "TEST", Number: "3"},
	}
	probAll := calculateCombinationProbability(intersectionAll, decklist, 6)
	
	expected := probAB + probAAC + probBC - probAB_AAC - probAB_BC - probAAC_BC + probAll
	
	t.Logf("P([A,B]) = %v", probAB)
	t.Logf("P([A,A,C]) = %v", probAAC)
	t.Logf("P([B,C]) = %v", probBC)
	t.Logf("P([A,B] ∩ [A,A,C]) = %v", probAB_AAC)
	t.Logf("P([A,B] ∩ [B,C]) = %v", probAB_BC)
	t.Logf("P([A,A,C] ∩ [B,C]) = %v", probAAC_BC)
	t.Logf("P([A,B] ∩ [A,A,C] ∩ [B,C]) = %v", probAll)
	t.Logf("Expected union = %v", expected)
	t.Logf("Got union = %v", result)
	
	tolerance := 0.0001
	diff := result - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("Union probability mismatch: got %v, expected %v (diff: %v)", result, expected, diff)
	}
}

// TestSemanticIssue_ExactlyVsAtLeast tests whether the "exactly" semantic
// is correct for the use case. This is the key semantic question.
func TestSemanticIssue_ExactlyVsAtLeast(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "A", SetCode: "TEST", Number: "1"}: 4,
		},
	}

	// CardSet: [AnyOf(1 A)] - meaning we need 1 A
	// This expands to: [A]
	cardSet := basictypes.NewCardSet([]basictypes.AnyOfPattern{
		{
			Cards: map[basictypes.Card]int{
				{Name: "A", SetCode: "TEST", Number: "1"}: 1,
			},
		},
	})

	expanded := cardSet.Expand(decklist)
	if len(expanded.Combinations) != 1 {
		t.Fatalf("Expected 1 combination, got %d", len(expanded.Combinations))
	}

	// The combination is [A]
	// calculateCombinationProbability([A]) calculates P(exactly 1 A in prizes)
	// But semantically, if prizes contain [A, A], does that satisfy the requirement?
	// The answer depends on interpretation:
	// - If [A] means "exactly 1 A", then [A, A] does NOT satisfy it
	// - If [A] means "at least 1 A", then [A, A] DOES satisfy it

	// Current implementation calculates P(exactly 1 A)
	probExactly1A := calculateCombinationProbability(expanded.Combinations[0], decklist, 6)
	
	// What P(at least 1 A) would be:
	// P(at least 1 A) = 1 - P(0 A) = 1 - C(56, 6) / C(60, 6)
	probAtLeast1A := 1.0 - float64(comb(56, 6)) / float64(comb(60, 6))

	t.Logf("P(exactly 1 A) = %v", probExactly1A)
	t.Logf("P(at least 1 A) = %v", probAtLeast1A)
	t.Logf("Difference: %v", probAtLeast1A - probExactly1A)

	// The question: which one is correct for the use case?
	// I think "at least" makes more semantic sense, but the current code uses "exactly".
	// This might be a bug, or it might be intentional if the CardSet expansion is meant
	// to generate all possible ways to satisfy it (including cases with more cards).

	// For now, just document the behavior
	result, err := CalculateCardSetPrizeOdds(decklist, map[string]basictypes.CardSet{"test": cardSet}, true)
	if err != nil {
		t.Fatalf("CalculateCardSetPrizeOdds failed: %v", err)
	}

	t.Logf("CalculateCardSetPrizeOdds result = %v", result["test"])
	t.Logf("This should equal P(exactly 1 A) = %v", probExactly1A)
	
	if result["test"] != probExactly1A {
		t.Errorf("Result should equal P(exactly 1 A), but got %v vs %v", result["test"], probExactly1A)
	}
}

