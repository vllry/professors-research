package prizeodds

import (
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

// TestUnionFromMultipleCardSets verifies that when we combine multiple CardSets
// (like in the API server), the union probability correctly accounts for intersections.
// This simulates what happens in the API server when processing limited_dragapult.
func TestUnionFromMultipleCardSets(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Dragapult ex", SetCode: "TWM", Number: "130"}: 3,
			{Name: "Night Stretcher", SetCode: "SFA", Number: "61"}: 2,
		},
	}

	dragapult := basictypes.Card{Name: "Dragapult ex", SetCode: "TWM", Number: "130"}
	nightStretcher := basictypes.Card{Name: "Night Stretcher", SetCode: "SFA", Number: "61"}

	// Simulate what the API server does:
	// CardSet 1: 2+ Dragapult ex (from anyOfs)
	cardSet1 := basictypes.NewCardSet([]basictypes.AnyOfPattern{
		{
			Cards: map[basictypes.Card]int{
				dragapult: 2,
			},
		},
	})
	expanded1 := cardSet1.Expand(decklist)

	// CardSet 2: 1+ Dragapult ex AND 2+ Night Stretcher (from allOfs)
	cardSet2 := basictypes.AllOf([]basictypes.Card{
		dragapult,
		nightStretcher,
		nightStretcher,
	})
	expanded2 := cardSet2.Expand(decklist)

	// Combine all combinations (as API server does)
	allCombinations := append(expanded1.Combinations, expanded2.Combinations...)
	t.Logf("Total combinations from both CardSets: %d", len(allCombinations))
	for i, combo := range allCombinations {
		t.Logf("  [%d] %v", i, combo)
	}

	// Calculate union probability (as API server does)
	unionProb := CalculateUnionProbability(allCombinations, decklist, true)
	t.Logf("Union probability: %v", unionProb)

	// Verify this matches the expected inclusion-exclusion result
	prob1 := CalculateUnionProbability(expanded1.Combinations, decklist, true)
	prob2 := CalculateUnionProbability(expanded2.Combinations, decklist, true)

	// Intersection: 2+ Dragapult AND 2+ Night Stretcher
	intersectionCombo := []basictypes.Card{
		dragapult,
		dragapult,
		nightStretcher,
		nightStretcher,
	}
	intersectionProb := calculateCombinationProbability(intersectionCombo, decklist, 6)

	expectedUnion := prob1 + prob2 - intersectionProb
	t.Logf("P(CardSet 1) = %v", prob1)
	t.Logf("P(CardSet 2) = %v", prob2)
	t.Logf("P(CardSet 1 ∩ CardSet 2) = %v", intersectionProb)
	t.Logf("Expected union = %v", expectedUnion)

	tolerance := 0.0001
	diff := abs(unionProb - expectedUnion)
	if diff > tolerance {
		t.Errorf("Union probability %v does not match expected %v (diff: %v)",
			unionProb, expectedUnion, diff)
	}

	// Critical check: Verify that configurations satisfying both conditions
	// (like 2 Dragapult + 2 Night Stretcher) are NOT double-counted
	// The union should be LESS than the sum (unless intersection is 0)
	if unionProb > prob1+prob2+tolerance {
		t.Errorf("Union %v exceeds sum %v - intersection not being subtracted!",
			unionProb, prob1+prob2)
	}

	// Verify intersection is being subtracted correctly
	// Union should equal sum - intersection (inclusion-exclusion)
	expectedFromInclusionExclusion := prob1 + prob2 - intersectionProb
	if abs(unionProb-expectedFromInclusionExclusion) > tolerance {
		t.Errorf("Union %v does not match inclusion-exclusion formula: P(A) + P(B) - P(A∩B) = %v",
			unionProb, expectedFromInclusionExclusion)
	}

	// If intersection > 0, union must be < sum
	if intersectionProb > tolerance && unionProb >= prob1+prob2-tolerance {
		t.Errorf("Union %v ≈ sum %v but intersection %v > 0 - intersection not properly subtracted!",
			unionProb, prob1+prob2, intersectionProb)
	}

	t.Logf("✓ Union correctly accounts for intersection: union = sum - intersection")
	t.Logf("  %v = %v - %v", unionProb, prob1+prob2, intersectionProb)
}

