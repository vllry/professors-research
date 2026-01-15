package prizeodds

import (
	"math"
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

// TestMathematicalProperties verifies fundamental mathematical properties
// that must hold for any correct probability calculation.
func TestMathematicalProperties(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "A", SetCode: "TEST", Number: "1"}: 4,
			{Name: "B", SetCode: "TEST", Number: "2"}: 4,
			{Name: "C", SetCode: "TEST", Number: "3"}: 4,
		},
	}

	t.Run("ProbabilityBounds", func(t *testing.T) {
		// Property: All probabilities must be in [0, 1]
		combinations := [][]basictypes.Card{
			{{Name: "A", SetCode: "TEST", Number: "1"}},
			{{Name: "A", SetCode: "TEST", Number: "1"}, {Name: "B", SetCode: "TEST", Number: "2"}},
			{{Name: "A", SetCode: "TEST", Number: "1"}, {Name: "A", SetCode: "TEST", Number: "1"}},
		}

		for i, combo := range combinations {
			prob := calculateCombinationProbability(combo, decklist, 6)
			if prob < 0.0 || prob > 1.0 {
				t.Errorf("Combination %d: probability %v not in [0, 1]", i, prob)
			}
		}
	})

	t.Run("AtLeastGreaterThanExactly", func(t *testing.T) {
		// Property: P(at least k) >= P(exactly k)
		// For [A], P(at least 1 A) should be >= P(exactly 1 A)
		comboA := []basictypes.Card{{Name: "A", SetCode: "TEST", Number: "1"}}
		probAtLeast1A := calculateCombinationProbability(comboA, decklist, 6)

		// P(exactly 1 A) = C(4,1) * C(56, 5) / C(60, 6)
		probExactly1A := float64(comb(4, 1) * comb(56, 5)) / float64(comb(60, 6))

		if probAtLeast1A < probExactly1A {
			t.Errorf("P(at least 1 A) = %v should be >= P(exactly 1 A) = %v", probAtLeast1A, probExactly1A)
		}
	})

	t.Run("Monotonicity", func(t *testing.T) {
		// Property: More restrictive requirements have lower probabilities
		// P(at least 1 A) >= P(at least 2 A) >= P(at least 3 A)
		combo1A := []basictypes.Card{{Name: "A", SetCode: "TEST", Number: "1"}}
		combo2A := []basictypes.Card{
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "A", SetCode: "TEST", Number: "1"},
		}
		combo3A := []basictypes.Card{
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "A", SetCode: "TEST", Number: "1"},
		}

		prob1 := calculateCombinationProbability(combo1A, decklist, 6)
		prob2 := calculateCombinationProbability(combo2A, decklist, 6)
		prob3 := calculateCombinationProbability(combo3A, decklist, 6)

		if prob1 < prob2 {
			t.Errorf("Monotonicity violated: P(>=1 A) = %v < P(>=2 A) = %v", prob1, prob2)
		}
		if prob2 < prob3 {
			t.Errorf("Monotonicity violated: P(>=2 A) = %v < P(>=3 A) = %v", prob2, prob3)
		}
	})

	t.Run("InclusionExclusionProperty", func(t *testing.T) {
		// Property: For two combinations A and B:
		// P(A ∪ B) = P(A) + P(B) - P(A ∩ B)
		// This is verified by calculateUnionProbability, but we can test it directly
		comboA := []basictypes.Card{{Name: "A", SetCode: "TEST", Number: "1"}}
		comboB := []basictypes.Card{{Name: "B", SetCode: "TEST", Number: "2"}}

		probA := calculateCombinationProbability(comboA, decklist, 6)
		probB := calculateCombinationProbability(comboB, decklist, 6)

		// Intersection: at least 1 A and at least 1 B
		comboAB := []basictypes.Card{
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "B", SetCode: "TEST", Number: "2"},
		}
		probAB := calculateCombinationProbability(comboAB, decklist, 6)

		// Union: at least 1 A OR at least 1 B
		unionProb := CalculateUnionProbability([][]basictypes.Card{comboA, comboB}, decklist, true)

		// Inclusion-exclusion: P(A ∪ B) = P(A) + P(B) - P(A ∩ B)
		expectedUnion := probA + probB - probAB

		tolerance := 0.0001
		diff := math.Abs(unionProb - expectedUnion)
		if diff > tolerance {
			t.Errorf("Inclusion-exclusion violated: P(A ∪ B) = %v, expected %v (diff: %v)",
				unionProb, expectedUnion, diff)
		}
	})

	t.Run("UnionLessThanOrEqualSum", func(t *testing.T) {
		// Property: P(A ∪ B) <= P(A) + P(B) (union is at most the sum)
		comboA := []basictypes.Card{{Name: "A", SetCode: "TEST", Number: "1"}}
		comboB := []basictypes.Card{{Name: "B", SetCode: "TEST", Number: "2"}}

		probA := calculateCombinationProbability(comboA, decklist, 6)
		probB := calculateCombinationProbability(comboB, decklist, 6)
		unionProb := CalculateUnionProbability([][]basictypes.Card{comboA, comboB}, decklist, true)

		if unionProb > probA+probB+0.0001 { // small tolerance for floating point
			t.Errorf("Union property violated: P(A ∪ B) = %v > P(A) + P(B) = %v",
				unionProb, probA+probB)
		}
	})

	t.Run("OrderIndependence", func(t *testing.T) {
		// Property: Order of cards in combination doesn't matter
		combo1 := []basictypes.Card{
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "B", SetCode: "TEST", Number: "2"},
			{Name: "C", SetCode: "TEST", Number: "3"},
		}
		combo2 := []basictypes.Card{
			{Name: "C", SetCode: "TEST", Number: "3"},
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "B", SetCode: "TEST", Number: "2"},
		}

		prob1 := calculateCombinationProbability(combo1, decklist, 6)
		prob2 := calculateCombinationProbability(combo2, decklist, 6)

		if math.Abs(prob1-prob2) > 0.0001 {
			t.Errorf("Order independence violated: P([A,B,C]) = %v != P([C,A,B]) = %v", prob1, prob2)
		}
	})

	t.Run("EmptyCombination", func(t *testing.T) {
		// Property: Empty combination is always satisfied (probability = 1.0)
		prob := calculateCombinationProbability([]basictypes.Card{}, decklist, 6)
		if prob != 1.0 {
			t.Errorf("Empty combination should have probability 1.0, got %v", prob)
		}
	})

	t.Run("ImpossibleCombination", func(t *testing.T) {
		// Property: Impossible combinations have probability 0.0
		// Requiring 7 cards (more than 6 prizes)
		impossibleCombo := []basictypes.Card{
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "A", SetCode: "TEST", Number: "1"},
			{Name: "B", SetCode: "TEST", Number: "2"},
			{Name: "B", SetCode: "TEST", Number: "2"},
			{Name: "B", SetCode: "TEST", Number: "2"},
		}

		prob := calculateCombinationProbability(impossibleCombo, decklist, 6)
		if prob != 0.0 {
			t.Errorf("Impossible combination (7 cards) should have probability 0.0, got %v", prob)
		}
	})
}

// TestKnownValues compares against manually calculated probabilities
// for specific known cases.
func TestKnownValues(t *testing.T) {
	t.Run("SingleCard_OneCopy", func(t *testing.T) {
		// Known: P(at least 1 copy when 1 copy in deck) = 0.1
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "A", SetCode: "TEST", Number: "1"}: 1,
			},
		}

		combo := []basictypes.Card{{Name: "A", SetCode: "TEST", Number: "1"}}
		prob := calculateCombinationProbability(combo, decklist, 6)

		expected := 0.1 // C(1,1) * C(59, 5) / C(60, 6) = 6/60 = 0.1
		tolerance := 0.0001

		if math.Abs(prob-expected) > tolerance {
			t.Errorf("P(at least 1 A when 1 copy in deck) = %v, expected %v", prob, expected)
		}
	})

	t.Run("SingleCard_FourCopies", func(t *testing.T) {
		// Known: P(at least 1 A when 4 copies in deck) = 1 - C(56, 6) / C(60, 6)
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "A", SetCode: "TEST", Number: "1"}: 4,
			},
		}

		combo := []basictypes.Card{{Name: "A", SetCode: "TEST", Number: "1"}}
		prob := calculateCombinationProbability(combo, decklist, 6)

		expected := 1.0 - float64(comb(56, 6)) / float64(comb(60, 6))
		tolerance := 0.0001

		if math.Abs(prob-expected) > tolerance {
			t.Errorf("P(at least 1 A when 4 copies) = %v, expected %v", prob, expected)
		}
	})
}

