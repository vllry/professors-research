package prizeodds

import (
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

// TestIntersectionExplicitDoubleCounting verifies that prize configurations
// that satisfy BOTH CardSets (like 2 Dragapult + 2 Night Stretcher) are
// counted exactly ONCE, not twice, in the union probability.
func TestIntersectionExplicitDoubleCounting(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Dragapult ex", SetCode: "TWM", Number: "130"}: 3,
			{Name: "Night Stretcher", SetCode: "SFA", Number: "61"}: 2,
		},
	}

	dragapult := basictypes.Card{Name: "Dragapult ex", SetCode: "TWM", Number: "130"}
	nightStretcher := basictypes.Card{Name: "Night Stretcher", SetCode: "SFA", Number: "61"}

	// CardSet 1: 2+ Dragapult ex
	cardSet1 := basictypes.NewCardSet([]basictypes.AnyOfPattern{
		{Cards: map[basictypes.Card]int{dragapult: 2}},
	})
	expanded1 := cardSet1.Expand(decklist)

	// CardSet 2: 1+ Dragapult ex AND 2+ Night Stretcher
	cardSet2 := basictypes.AllOf([]basictypes.Card{
		dragapult,
		nightStretcher,
		nightStretcher,
	})
	expanded2 := cardSet2.Expand(decklist)

	// The configuration that satisfies BOTH: [2 Dragapult, 2 Night Stretcher]
	intersectionCombo := []basictypes.Card{
		dragapult, dragapult,
		nightStretcher, nightStretcher,
	}
	intersectionProb := calculateCombinationProbability(intersectionCombo, decklist, 6)

	t.Logf("Intersection combo (satisfies BOTH): %v", intersectionCombo)
	t.Logf("P(intersection) = %v", intersectionProb)

	// Calculate individual probabilities
	prob1 := CalculateUnionProbability(expanded1.Combinations, decklist, true)
	prob2 := CalculateUnionProbability(expanded2.Combinations, decklist, true)

	t.Logf("P(CardSet 1: 2+ Dragapult) = %v", prob1)
	t.Logf("P(CardSet 2: 1+ Dragapult AND 2+ Night Stretcher) = %v", prob2)

	// If we naively added them, we'd get:
	naiveSum := prob1 + prob2
	t.Logf("Naive sum (WRONG - double counts intersection): %v", naiveSum)

	// Correct union using inclusion-exclusion:
	allCombinations := append(expanded1.Combinations, expanded2.Combinations...)
	unionProb := CalculateUnionProbability(allCombinations, decklist, true)
	correctUnion := prob1 + prob2 - intersectionProb

	t.Logf("Correct union (P(A) + P(B) - P(A∩B)): %v", correctUnion)
	t.Logf("Calculated union: %v", unionProb)

	// Verify they match
	tolerance := 0.0001
	if abs(unionProb-correctUnion) > tolerance {
		t.Errorf("Union probability %v does not match inclusion-exclusion formula %v",
			unionProb, correctUnion)
	}

	// Critical assertion: Union must be LESS than naive sum (unless intersection is 0)
	if intersectionProb > tolerance && unionProb >= naiveSum-tolerance {
		t.Errorf("Union %v should be less than naive sum %v by intersection %v, but it's not!",
			unionProb, naiveSum, intersectionProb)
	}

	// Verify the difference equals the intersection
	expectedDifference := naiveSum - unionProb
	if abs(expectedDifference-intersectionProb) > tolerance {
		t.Errorf("Difference between naive sum and union %v does not equal intersection %v",
			expectedDifference, intersectionProb)
	}

	t.Logf("✓ Verified: Union = Sum - Intersection")
	t.Logf("  %v = %v - %v", unionProb, naiveSum, intersectionProb)
	t.Logf("✓ Configurations satisfying BOTH conditions are counted exactly ONCE")
}




