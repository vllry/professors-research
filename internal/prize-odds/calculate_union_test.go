package prizeodds

import (
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

// TestUnionIntersectionCorrectness verifies that the inclusion-exclusion principle
// correctly handles intersections to avoid double-counting.
// This is critical for cases like limited_dragapult where multiple CardSets
// can be satisfied simultaneously.
func TestUnionIntersectionCorrectness(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Dragapult ex", SetCode: "TWM", Number: "130"}: 3,
			{Name: "Night Stretcher", SetCode: "SFA", Number: "61"}: 2,
		},
	}

	dragapult := basictypes.Card{Name: "Dragapult ex", SetCode: "TWM", Number: "130"}
	nightStretcher := basictypes.Card{Name: "Night Stretcher", SetCode: "SFA", Number: "61"}

	// CardSet 1: 2+ Dragapult ex
	// This expands to: [Dragapult, Dragapult]
	cardSet1 := basictypes.NewCardSet([]basictypes.AnyOfPattern{
		{
			Cards: map[basictypes.Card]int{
				dragapult: 2,
			},
		},
	})
	expanded1 := cardSet1.Expand(decklist)
	t.Logf("CardSet 1 (2+ Dragapult) expands to %d combinations", len(expanded1.Combinations))
	for i, combo := range expanded1.Combinations {
		t.Logf("  [%d] %v", i, combo)
	}

	// CardSet 2: 1+ Dragapult ex AND 2+ Night Stretcher
	// This expands to: [Dragapult, Night Stretcher, Night Stretcher]
	cardSet2 := basictypes.AllOf([]basictypes.Card{
		dragapult,
		nightStretcher,
		nightStretcher,
	})
	expanded2 := cardSet2.Expand(decklist)
	t.Logf("CardSet 2 (1+ Dragapult AND 2+ Night Stretcher) expands to %d combinations", len(expanded2.Combinations))
	for i, combo := range expanded2.Combinations {
		t.Logf("  [%d] %v", i, combo)
	}

	// Calculate individual probabilities
	prob1 := CalculateUnionProbability(expanded1.Combinations, decklist, true)
	prob2 := CalculateUnionProbability(expanded2.Combinations, decklist, true)
	t.Logf("P(CardSet 1) = %v", prob1)
	t.Logf("P(CardSet 2) = %v", prob2)

	// Calculate union probability (what we should get)
	allCombinations := append(expanded1.Combinations, expanded2.Combinations...)
	unionProb := CalculateUnionProbability(allCombinations, decklist, true)
	t.Logf("P(CardSet 1 ∪ CardSet 2) = %v", unionProb)

	// Verify inclusion-exclusion: P(A ∪ B) = P(A) + P(B) - P(A ∩ B)
	// The intersection is: 2+ Dragapult AND 2+ Night Stretcher
	// This is [Dragapult, Dragapult, Night Stretcher, Night Stretcher]
	intersectionCombo := []basictypes.Card{
		dragapult,
		dragapult,
		nightStretcher,
		nightStretcher,
	}
	intersectionProb := calculateCombinationProbability(intersectionCombo, decklist, 6)
	t.Logf("P(CardSet 1 ∩ CardSet 2) = %v", intersectionProb)

	expectedUnion := prob1 + prob2 - intersectionProb
	t.Logf("Expected union (P(A) + P(B) - P(A∩B)) = %v", expectedUnion)

	// Allow small tolerance for floating point and truncation
	tolerance := 0.0001
	diff := abs(unionProb - expectedUnion)
	if diff > tolerance {
		t.Errorf("Inclusion-exclusion violated: P(A ∪ B) = %v, expected %v (diff: %v)",
			unionProb, expectedUnion, diff)
	}

	// Verify that union is less than or equal to sum (should be less due to intersection)
	if unionProb > prob1+prob2+tolerance {
		t.Errorf("Union probability %v exceeds sum %v (should be less due to intersection)",
			unionProb, prob1+prob2)
	}

	// Verify that union accounts for intersection (should be less than sum)
	// Note: If union is approximately equal to sum, the intersection is negligible,
	// which is mathematically valid but may indicate the test case has minimal overlap.
	if unionProb < prob1+prob2-tolerance {
		t.Logf("✓ Union correctly accounts for intersection (union < sum)")
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}




