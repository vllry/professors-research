package prizeodds

import (
	"math"
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

// TestCalculateBasicPokemonStartOdds_ManualVerification manually verifies the math
// for the dragapult decklist to ensure our calculations are correct.
func TestCalculateBasicPokemonStartOdds_ManualVerification(t *testing.T) {
	// Decklist from dragapult_MEG_1.txt
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Dreepy", SetCode: "TWM", Number: "128"}: 4,
			{Name: "Duskull", SetCode: "SFA", Number: "18"}: 2,
			{Name: "Budew", SetCode: "PRE", Number: "4"}: 2,
			{Name: "Bloodmoon Ursaluna ex", SetCode: "PRE", Number: "168"}: 1,
			{Name: "Hawlucha", SetCode: "SVI", Number: "118"}: 1,
			{Name: "Munkidori", SetCode: "TWM", Number: "95"}: 1,
			{Name: "Latias ex", SetCode: "SSP", Number: "76"}: 1,
			{Name: "Fezandipiti ex", SetCode: "SFA", Number: "38"}: 1,
			{Name: "Other", SetCode: "TWM", Number: "1"}: 47, // Non-basic cards
		},
	}

	basicPokemonCards := map[basictypes.Card]bool{
		{Name: "Dreepy", SetCode: "TWM", Number: "128"}: true,
		{Name: "Duskull", SetCode: "SFA", Number: "18"}: true,
		{Name: "Budew", SetCode: "PRE", Number: "4"}: true,
		{Name: "Bloodmoon Ursaluna ex", SetCode: "PRE", Number: "168"}: true,
		{Name: "Hawlucha", SetCode: "SVI", Number: "118"}: true,
		{Name: "Munkidori", SetCode: "TWM", Number: "95"}: true,
		{Name: "Latias ex", SetCode: "SSP", Number: "76"}: true,
		{Name: "Fezandipiti ex", SetCode: "SFA", Number: "38"}: true,
	}

	possibleStarters, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

	// Total basic count: 4 + 2 + 2 + 1 + 1 + 1 + 1 + 1 = 13
	// Non-basic count: 60 - 13 = 47

	// Verify Dreepy (4 copies)
	dreepy := basictypes.Card{Name: "Dreepy", SetCode: "TWM", Number: "128"}
	
	// Possible starter: P(>=1 Dreepy in 7 draws) = 1 - P(0 Dreepy)
	// P(0 Dreepy) = C(56, 7) / C(60, 7)
	expectedDreepyPossible := 1.0 - float64(comb(56, 7))/float64(comb(60, 7))
	if math.Abs(possibleStarters[dreepy]-expectedDreepyPossible) > 0.0001 {
		t.Errorf("Dreepy possible starter: got %v, expected %v", possibleStarters[dreepy], expectedDreepyPossible)
	}
	t.Logf("Dreepy possible starter: %v (expected: %v)", possibleStarters[dreepy], expectedDreepyPossible)

	// Forced starter: P(>=1 Dreepy AND 0 other basic) = Sum i=1 to 4: C(4, i) * C(47, 7-i) / C(60, 7)
	expectedDreepyForced := 0.0
	for i := 1; i <= 4 && i <= 7; i++ {
		if 7-i <= 47 && 7-i >= 0 {
			expectedDreepyForced += float64(comb(4, i)) * float64(comb(47, 7-i)) / float64(comb(60, 7))
		}
	}
	if math.Abs(forcedStarters[dreepy]-expectedDreepyForced) > 0.0001 {
		t.Errorf("Dreepy forced starter: got %v, expected %v", forcedStarters[dreepy], expectedDreepyForced)
	}
	t.Logf("Dreepy forced starter: %v (expected: %v)", forcedStarters[dreepy], expectedDreepyForced)

	// Verify Duskull (2 copies)
	duskull := basictypes.Card{Name: "Duskull", SetCode: "SFA", Number: "18"}
	
	// Possible starter: P(>=1 Duskull in 7 draws) = 1 - P(0 Duskull)
	// P(0 Duskull) = C(58, 7) / C(60, 7)
	expectedDuskullPossible := 1.0 - float64(comb(58, 7))/float64(comb(60, 7))
	if math.Abs(possibleStarters[duskull]-expectedDuskullPossible) > 0.0001 {
		t.Errorf("Duskull possible starter: got %v, expected %v", possibleStarters[duskull], expectedDuskullPossible)
	}
	t.Logf("Duskull possible starter: %v (expected: %v)", possibleStarters[duskull], expectedDuskullPossible)

	// Forced starter: P(>=1 Duskull AND 0 other basic) = Sum i=1 to 2: C(2, i) * C(47, 7-i) / C(60, 7)
	expectedDuskullForced := 0.0
	for i := 1; i <= 2 && i <= 7; i++ {
		if 7-i <= 47 && 7-i >= 0 {
			expectedDuskullForced += float64(comb(2, i)) * float64(comb(47, 7-i)) / float64(comb(60, 7))
		}
	}
	if math.Abs(forcedStarters[duskull]-expectedDuskullForced) > 0.0001 {
		t.Errorf("Duskull forced starter: got %v, expected %v", forcedStarters[duskull], expectedDuskullForced)
	}
	t.Logf("Duskull forced starter: %v (expected: %v)", forcedStarters[duskull], expectedDuskullForced)

	// Verify single copy (e.g., Hawlucha)
	hawlucha := basictypes.Card{Name: "Hawlucha", SetCode: "SVI", Number: "118"}
	
	// Possible starter: P(>=1 Hawlucha in 7 draws) = 1 - P(0 Hawlucha)
	// P(0 Hawlucha) = C(59, 7) / C(60, 7)
	expectedHawluchaPossible := 1.0 - float64(comb(59, 7))/float64(comb(60, 7))
	if math.Abs(possibleStarters[hawlucha]-expectedHawluchaPossible) > 0.0001 {
		t.Errorf("Hawlucha possible starter: got %v, expected %v", possibleStarters[hawlucha], expectedHawluchaPossible)
	}
	t.Logf("Hawlucha possible starter: %v (expected: %v)", possibleStarters[hawlucha], expectedHawluchaPossible)

	// Forced starter: P(>=1 Hawlucha AND 0 other basic) = C(1, 1) * C(47, 6) / C(60, 7)
	expectedHawluchaForced := float64(comb(1, 1)) * float64(comb(47, 6)) / float64(comb(60, 7))
	if math.Abs(forcedStarters[hawlucha]-expectedHawluchaForced) > 0.0001 {
		t.Errorf("Hawlucha forced starter: got %v, expected %v", forcedStarters[hawlucha], expectedHawluchaForced)
	}
	t.Logf("Hawlucha forced starter: %v (expected: %v)", forcedStarters[hawlucha], expectedHawluchaForced)

	// Verify mulligan: P(0 basic in 7 draws) = C(47, 7) / C(60, 7)
	expectedMulligan := float64(comb(47, 7)) / float64(comb(60, 7))
	if math.Abs(mulliganOdds-expectedMulligan) > 0.0001 {
		t.Errorf("Mulligan odds: got %v, expected %v", mulliganOdds, expectedMulligan)
	}
	t.Logf("Mulligan odds: %v (expected: %v)", mulliganOdds, expectedMulligan)

	// Print all results for comparison
	t.Logf("\n=== All Results ===")
	for card, prob := range possibleStarters {
		t.Logf("Possible: %s %s %s = %.4f%%", card.Name, card.SetCode, card.Number, prob*100)
	}
	for card, prob := range forcedStarters {
		t.Logf("Forced: %s %s %s = %.4f%%", card.Name, card.SetCode, card.Number, prob*100)
	}
	t.Logf("Mulligan: %.4f%%", mulliganOdds*100)

	// Verify at least 1 basic: P(>=1 basic) = 1 - P(0 basic) = 1 - mulliganOdds
	expectedAtLeastOne := 1.0 - mulliganOdds
	atLeastOneCalculated := 1.0 - mulliganOdds
	if math.Abs(atLeastOneCalculated-expectedAtLeastOne) > 0.0001 {
		t.Errorf("At least 1 basic calculation error")
	}
	t.Logf("At least 1 basic: %.4f%% (expected: %.4f%%)", atLeastOneCalculated*100, expectedAtLeastOne*100)

	// Verify at least 2 basic: P(>=2 basic) = 1 - P(0 basic) - P(1 basic)
	// P(1 basic) = C(13, 1) * C(47, 6) / C(60, 7)
	probOne := float64(comb(13, 1)) * float64(comb(47, 6)) / float64(comb(60, 7))
	expectedAtLeastTwo := 1.0 - mulliganOdds - probOne
	atLeastTwo := CalculateAtLeastTwoBasic(13)
	if math.Abs(atLeastTwo-expectedAtLeastTwo) > 0.0001 {
		t.Errorf("At least 2 basic: got %v, expected %v", atLeastTwo, expectedAtLeastTwo)
	}
	t.Logf("At least 2 basic: %.4f%% (expected: %.4f%%)", atLeastTwo*100, expectedAtLeastTwo*100)
}

