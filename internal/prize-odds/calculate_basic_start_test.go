package prizeodds

import (
	"math"
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

func TestCalculateBasicPokemonStartOdds_SimpleCase(t *testing.T) {
	// Simple deck with 4 copies of one basic Pokemon
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Dreepy", SetCode: "TWM", Number: "128"}: 4,
			{Name: "Other Card", SetCode: "TWM", Number: "1"}: 56,
		},
	}

	basicPokemonCards := map[basictypes.Card]bool{
		{Name: "Dreepy", SetCode: "TWM", Number: "128"}: true,
	}

	possibleStarters, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

	// Check possible starters
	if len(possibleStarters) != 1 {
		t.Errorf("Expected 1 basic Pokemon in possibleStarters, got %d", len(possibleStarters))
	}

	dreepy := basictypes.Card{Name: "Dreepy", SetCode: "TWM", Number: "128"}
	prob, ok := possibleStarters[dreepy]
	if !ok {
		t.Fatalf("Dreepy not found in possibleStarters")
	}

	// Probability of at least 1 Dreepy in 7 draws from 4 copies
	expectedProb := 0.0
	for j := 1; j <= 4; j++ {
		expectedProb += DefaultSevenCardOddsTable.Get(4, j)
	}

	if math.Abs(prob-expectedProb) > 0.0001 {
		t.Errorf("Possible starter odds for Dreepy = %v, expected %v", prob, expectedProb)
	}

	// Check forced starters
	if len(forcedStarters) != 1 {
		t.Errorf("Expected 1 basic Pokemon in forcedStarters, got %d", len(forcedStarters))
	}

	forcedProb, ok := forcedStarters[dreepy]
	if !ok {
		t.Fatalf("Dreepy not found in forcedStarters")
	}

	// Forced starter: at least 1 Dreepy AND 0 other basic (no other basic in deck)
	// This should equal the probability of at least 1 Dreepy since there are no other basic
	if math.Abs(forcedProb-prob) > 0.0001 {
		t.Errorf("Forced starter odds for Dreepy = %v, expected %v (same as possible since no other basic)", forcedProb, prob)
	}

	// Check mulligan odds
	// P(no basic in 7 draws) = C(56, 7) / C(60, 7)
	expectedMulligan := calculateZeroInDraw(4, 7, 60)
	if math.Abs(mulliganOdds-expectedMulligan) > 0.0001 {
		t.Errorf("Mulligan odds = %v, expected %v", mulliganOdds, expectedMulligan)
	}

	t.Logf("Possible starters: %v", possibleStarters)
	t.Logf("Forced starters: %v", forcedStarters)
	t.Logf("Mulligan odds: %v", mulliganOdds)
}

func TestCalculateBasicPokemonStartOdds_TwoBasicPokemon(t *testing.T) {
	// Deck with 2 different basic Pokemon
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Dreepy", SetCode: "TWM", Number: "128"}: 4,
			{Name: "Duskull", SetCode: "SFA", Number: "18"}: 2,
			{Name: "Other Card", SetCode: "TWM", Number: "1"}: 54,
		},
	}

	basicPokemonCards := map[basictypes.Card]bool{
		{Name: "Dreepy", SetCode: "TWM", Number: "128"}: true,
		{Name: "Duskull", SetCode: "SFA", Number: "18"}: true,
	}

	possibleStarters, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

	// Should have 2 basic Pokemon
	if len(possibleStarters) != 2 {
		t.Errorf("Expected 2 basic Pokemon in possibleStarters, got %d", len(possibleStarters))
	}

	dreepy := basictypes.Card{Name: "Dreepy", SetCode: "TWM", Number: "128"}
	duskull := basictypes.Card{Name: "Duskull", SetCode: "SFA", Number: "18"}

	// Check Dreepy possible starter odds
	dreepyProb, ok := possibleStarters[dreepy]
	if !ok {
		t.Fatalf("Dreepy not found in possibleStarters")
	}
	expectedDreepyProb := 0.0
	for j := 1; j <= 4; j++ {
		expectedDreepyProb += DefaultSevenCardOddsTable.Get(4, j)
	}
	if math.Abs(dreepyProb-expectedDreepyProb) > 0.0001 {
		t.Errorf("Dreepy possible starter odds = %v, expected %v", dreepyProb, expectedDreepyProb)
	}

	// Check Duskull possible starter odds
	duskullProb, ok := possibleStarters[duskull]
	if !ok {
		t.Fatalf("Duskull not found in possibleStarters")
	}
	expectedDuskullProb := 0.0
	for j := 1; j <= 2; j++ {
		expectedDuskullProb += DefaultSevenCardOddsTable.Get(2, j)
	}
	if math.Abs(duskullProb-expectedDuskullProb) > 0.0001 {
		t.Errorf("Duskull possible starter odds = %v, expected %v", duskullProb, expectedDuskullProb)
	}

	// Check forced starters
	// Forced Dreepy: at least 1 Dreepy AND 0 Duskull
	// This uses multivariate hypergeometric: C(4, i) * C(54, 7-i) / C(60, 7) for i=1..4
	dreepyForced, ok := forcedStarters[dreepy]
	if !ok {
		t.Fatalf("Dreepy not found in forcedStarters")
	}

	expectedDreepyForced := 0.0
	for i := 1; i <= 4 && i <= 7; i++ {
		if 7-i <= 54 && 7-i >= 0 {
			numerator := float64(comb(4, i)) * float64(comb(54, 7-i))
			denominator := float64(comb(60, 7))
			if denominator > 0 {
				expectedDreepyForced += numerator / denominator
			}
		}
	}
	if math.Abs(dreepyForced-expectedDreepyForced) > 0.0001 {
		t.Errorf("Dreepy forced starter odds = %v, expected %v", dreepyForced, expectedDreepyForced)
	}

	// Check mulligan: P(no basic in 7 draws) = C(54, 7) / C(60, 7)
	expectedMulligan := calculateZeroInDraw(6, 7, 60) // 4 Dreepy + 2 Duskull = 6 total basic
	if math.Abs(mulliganOdds-expectedMulligan) > 0.0001 {
		t.Errorf("Mulligan odds = %v, expected %v", mulliganOdds, expectedMulligan)
	}

	t.Logf("Possible starters: %v", possibleStarters)
	t.Logf("Forced starters: %v", forcedStarters)
	t.Logf("Mulligan odds: %v", mulliganOdds)
}

func TestSevenCardOddsTable_Get(t *testing.T) {
	// Test some known values for 7 cards
	tests := []struct {
		x    int
		y    int
		want float64
	}{
		{0, 0, 1.0},   // If 0 copies in deck, 100% chance of having 0 in 7 draws
		{0, 1, 0.0},   // If 0 copies in deck, 0% chance of having 1
		{1, 0, 0.8833}, // If 1 copy in deck, ~88.33% chance of having 0 (53/60 * 52/59 * ... * 47/54)
		{1, 1, 0.1167}, // If 1 copy in deck, ~11.67% chance of having 1 (7/60)
		{60, 0, 0.0},  // Out of range x
		{0, 8, 0.0},   // Out of range y (max is 7)
	}

	for _, tt := range tests {
		got := DefaultSevenCardOddsTable.Get(tt.x, tt.y)
		// For the probability checks, allow some tolerance
		if tt.x == 1 && tt.y == 0 {
			// P(0 copies when 1 in deck) = C(59, 7) / C(60, 7) = 53/60 ≈ 0.8833
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("DefaultSevenCardOddsTable.Get(%d, %d) = %v, want ~%v", tt.x, tt.y, got, tt.want)
			}
		} else if tt.x == 1 && tt.y == 1 {
			// P(1 copy when 1 in deck) = C(1, 1) * C(59, 6) / C(60, 7) = 7/60 ≈ 0.1167
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("DefaultSevenCardOddsTable.Get(%d, %d) = %v, want ~%v", tt.x, tt.y, got, tt.want)
			}
		} else if got != tt.want {
			t.Errorf("DefaultSevenCardOddsTable.Get(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestSevenCardOddsTable_Get_SumToOne(t *testing.T) {
	// For any x, the sum of probabilities for y=0..7 should be approximately 1.0
	for x := 0; x <= 59; x++ {
		sum := 0.0
		for y := 0; y <= 7; y++ {
			sum += DefaultSevenCardOddsTable.Get(x, y)
		}
		// Allow small floating point error
		if sum < 0.9999 || sum > 1.0001 {
			t.Errorf("Sum of probabilities for x=%d is %v, expected ~1.0", x, sum)
		}
	}
}

func TestCalculateBasicPokemonStartOdds_EdgeCases(t *testing.T) {
	t.Run("NoBasicPokemon", func(t *testing.T) {
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "Trainer", SetCode: "TEST", Number: "1"}: 60,
			},
		}
		basicPokemonCards := map[basictypes.Card]bool{}

		possibleStarters, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

		if len(possibleStarters) != 0 {
			t.Errorf("Expected 0 basic Pokemon in possibleStarters, got %d", len(possibleStarters))
		}
		if len(forcedStarters) != 0 {
			t.Errorf("Expected 0 basic Pokemon in forcedStarters, got %d", len(forcedStarters))
		}
		// Mulligan should be 100% (no basic Pokemon means you always mulligan)
		if math.Abs(mulliganOdds-1.0) > 0.0001 {
			t.Errorf("Mulligan odds = %v, expected 1.0 (100%% mulligan)", mulliganOdds)
		}
	})

	t.Run("AllBasicPokemon", func(t *testing.T) {
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "Pikachu", SetCode: "TEST", Number: "1"}: 60,
			},
		}
		basicPokemonCards := map[basictypes.Card]bool{
			{Name: "Pikachu", SetCode: "TEST", Number: "1"}: true,
		}

		possibleStarters, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

		pikachu := basictypes.Card{Name: "Pikachu", SetCode: "TEST", Number: "1"}

		// Possible starter: P(>=1 Pikachu) = 1.0 (always have at least 1)
		if math.Abs(possibleStarters[pikachu]-1.0) > 0.0001 {
			t.Errorf("Possible starter odds for Pikachu = %v, expected 1.0", possibleStarters[pikachu])
		}

		// Forced starter: P(>=1 Pikachu AND 0 other basic) = 1.0 (no other basic)
		if math.Abs(forcedStarters[pikachu]-1.0) > 0.0001 {
			t.Errorf("Forced starter odds for Pikachu = %v, expected 1.0", forcedStarters[pikachu])
		}

		// Mulligan: P(0 basic) = 0.0 (impossible)
		if math.Abs(mulliganOdds-0.0) > 0.0001 {
			t.Errorf("Mulligan odds = %v, expected 0.0", mulliganOdds)
		}
	})

	t.Run("SingleCopyBasic", func(t *testing.T) {
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "Rare", SetCode: "TEST", Number: "1"}: 1,
				{Name: "Other", SetCode: "TEST", Number: "2"}: 59,
			},
		}
		basicPokemonCards := map[basictypes.Card]bool{
			{Name: "Rare", SetCode: "TEST", Number: "1"}: true,
		}

		possibleStarters, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

		rare := basictypes.Card{Name: "Rare", SetCode: "TEST", Number: "1"}

		// Possible starter: P(>=1 Rare) = P(exactly 1) = C(1,1) * C(59,6) / C(60,7) = 7/60
		expectedPossible := DefaultSevenCardOddsTable.Get(1, 1)
		if math.Abs(possibleStarters[rare]-expectedPossible) > 0.0001 {
			t.Errorf("Possible starter odds for Rare = %v, expected %v", possibleStarters[rare], expectedPossible)
		}

		// Forced starter: same as possible (no other basic)
		if math.Abs(forcedStarters[rare]-expectedPossible) > 0.0001 {
			t.Errorf("Forced starter odds for Rare = %v, expected %v", forcedStarters[rare], expectedPossible)
		}

		// Mulligan: P(0 Rare) = C(59, 7) / C(60, 7) = 53/60
		expectedMulligan := DefaultSevenCardOddsTable.Get(1, 0)
		if math.Abs(mulliganOdds-expectedMulligan) > 0.0001 {
			t.Errorf("Mulligan odds = %v, expected %v", mulliganOdds, expectedMulligan)
		}
	})
}

func TestCalculateBasicPokemonStartOdds_MathematicalProperties(t *testing.T) {
	t.Run("ForcedStartersSumPlusMulligan", func(t *testing.T) {
		// Property: Sum of all forced starters + mulligan should be <= 1
		// They represent mutually exclusive events that cover all possibilities
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "A", SetCode: "TEST", Number: "1"}: 4,
				{Name: "B", SetCode: "TEST", Number: "2"}: 3,
				{Name: "C", SetCode: "TEST", Number: "3"}: 2,
				{Name: "Other", SetCode: "TEST", Number: "4"}: 51,
			},
		}
		basicPokemonCards := map[basictypes.Card]bool{
			{Name: "A", SetCode: "TEST", Number: "1"}: true,
			{Name: "B", SetCode: "TEST", Number: "2"}: true,
			{Name: "C", SetCode: "TEST", Number: "3"}: true,
		}

		_, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

		sumForced := 0.0
		for _, prob := range forcedStarters {
			sumForced += prob
		}

		total := sumForced + mulliganOdds
		// The sum should be <= 1 (they're mutually exclusive events)
		// It might be slightly less than 1 if there are cases where you have multiple basic Pokemon
		// (which means none of the forced starter events occur)
		if total > 1.0001 {
			t.Errorf("Sum of forced starters (%v) + mulligan (%v) = %v, expected <= 1.0", sumForced, mulliganOdds, total)
		}

		t.Logf("Sum of forced starters: %v", sumForced)
		t.Logf("Mulligan odds: %v", mulliganOdds)
		t.Logf("Total: %v", total)
	})

	t.Run("PossibleStartersGreaterThanForcedStarters", func(t *testing.T) {
		// Property: For each basic Pokemon, possible starter odds >= forced starter odds
		// Because forced = possible AND (no other basic), which is a subset
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "A", SetCode: "TEST", Number: "1"}: 4,
				{Name: "B", SetCode: "TEST", Number: "2"}: 3,
				{Name: "Other", SetCode: "TEST", Number: "3"}: 53,
			},
		}
		basicPokemonCards := map[basictypes.Card]bool{
			{Name: "A", SetCode: "TEST", Number: "1"}: true,
			{Name: "B", SetCode: "TEST", Number: "2"}: true,
		}

		possibleStarters, forcedStarters, _ := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

		cardA := basictypes.Card{Name: "A", SetCode: "TEST", Number: "1"}
		cardB := basictypes.Card{Name: "B", SetCode: "TEST", Number: "2"}

		if possibleStarters[cardA] < forcedStarters[cardA] {
			t.Errorf("Possible starter odds for A (%v) should be >= forced starter odds (%v)", possibleStarters[cardA], forcedStarters[cardA])
		}

		if possibleStarters[cardB] < forcedStarters[cardB] {
			t.Errorf("Possible starter odds for B (%v) should be >= forced starter odds (%v)", possibleStarters[cardB], forcedStarters[cardB])
		}
	})
}

func TestCalculateBasicPokemonStartOdds_KnownValues(t *testing.T) {
	t.Run("FourCopies_ManualCalculation", func(t *testing.T) {
		// Deck: 4 copies of basic Pokemon, 56 other cards
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "Basic", SetCode: "TEST", Number: "1"}: 4,
				{Name: "Other", SetCode: "TEST", Number: "2"}: 56,
			},
		}
		basicPokemonCards := map[basictypes.Card]bool{
			{Name: "Basic", SetCode: "TEST", Number: "1"}: true,
		}

		possibleStarters, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

		basic := basictypes.Card{Name: "Basic", SetCode: "TEST", Number: "1"}

		// Manual calculation: P(>=1 Basic in 7 draws) = 1 - P(0 Basic)
		// P(0 Basic) = C(56, 7) / C(60, 7)
		expectedPossible := 1.0 - float64(comb(56, 7))/float64(comb(60, 7))
		if math.Abs(possibleStarters[basic]-expectedPossible) > 0.0001 {
			t.Errorf("Possible starter odds = %v, expected %v", possibleStarters[basic], expectedPossible)
		}

		// Forced starter: same as possible (no other basic)
		if math.Abs(forcedStarters[basic]-expectedPossible) > 0.0001 {
			t.Errorf("Forced starter odds = %v, expected %v", forcedStarters[basic], expectedPossible)
		}

		// Mulligan: P(0 Basic) = C(56, 7) / C(60, 7)
		expectedMulligan := float64(comb(56, 7)) / float64(comb(60, 7))
		if math.Abs(mulliganOdds-expectedMulligan) > 0.0001 {
			t.Errorf("Mulligan odds = %v, expected %v", mulliganOdds, expectedMulligan)
		}

		// Verify: possible + mulligan = 1
		if math.Abs(possibleStarters[basic]+mulliganOdds-1.0) > 0.0001 {
			t.Errorf("Possible starter (%v) + mulligan (%v) should equal 1.0, got %v", possibleStarters[basic], mulliganOdds, possibleStarters[basic]+mulliganOdds)
		}
	})

	t.Run("TwoBasicPokemon_ManualCalculation", func(t *testing.T) {
		// Deck: 4 copies of A, 2 copies of B, 54 other cards
		decklist := basictypes.Decklist{
			Cards: map[basictypes.Card]int{
				{Name: "A", SetCode: "TEST", Number: "1"}: 4,
				{Name: "B", SetCode: "TEST", Number: "2"}: 2,
				{Name: "Other", SetCode: "TEST", Number: "3"}: 54,
			},
		}
		basicPokemonCards := map[basictypes.Card]bool{
			{Name: "A", SetCode: "TEST", Number: "1"}: true,
			{Name: "B", SetCode: "TEST", Number: "2"}: true,
		}

		possibleStarters, forcedStarters, mulliganOdds := CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

		cardA := basictypes.Card{Name: "A", SetCode: "TEST", Number: "1"}
		cardB := basictypes.Card{Name: "B", SetCode: "TEST", Number: "2"}

		// Possible starter A: P(>=1 A) = 1 - P(0 A) = 1 - C(56, 7) / C(60, 7)
		expectedPossibleA := 1.0 - float64(comb(56, 7))/float64(comb(60, 7))
		if math.Abs(possibleStarters[cardA]-expectedPossibleA) > 0.0001 {
			t.Errorf("Possible starter odds for A = %v, expected %v", possibleStarters[cardA], expectedPossibleA)
		}

		// Possible starter B: P(>=1 B) = 1 - P(0 B) = 1 - C(58, 7) / C(60, 7)
		expectedPossibleB := 1.0 - float64(comb(58, 7))/float64(comb(60, 7))
		if math.Abs(possibleStarters[cardB]-expectedPossibleB) > 0.0001 {
			t.Errorf("Possible starter odds for B = %v, expected %v", possibleStarters[cardB], expectedPossibleB)
		}

		// Forced starter A: P(>=1 A AND 0 B) = sum over i=1..4: C(4, i) * C(54, 7-i) / C(60, 7)
		expectedForcedA := 0.0
		for i := 1; i <= 4 && i <= 7; i++ {
			if 7-i <= 54 && 7-i >= 0 {
				expectedForcedA += float64(comb(4, i)) * float64(comb(54, 7-i)) / float64(comb(60, 7))
			}
		}
		if math.Abs(forcedStarters[cardA]-expectedForcedA) > 0.0001 {
			t.Errorf("Forced starter odds for A = %v, expected %v", forcedStarters[cardA], expectedForcedA)
		}

		// Forced starter B: P(>=1 B AND 0 A) = sum over i=1..2: C(2, i) * C(54, 7-i) / C(60, 7)
		expectedForcedB := 0.0
		for i := 1; i <= 2 && i <= 7; i++ {
			if 7-i <= 54 && 7-i >= 0 {
				expectedForcedB += float64(comb(2, i)) * float64(comb(54, 7-i)) / float64(comb(60, 7))
			}
		}
		if math.Abs(forcedStarters[cardB]-expectedForcedB) > 0.0001 {
			t.Errorf("Forced starter odds for B = %v, expected %v", forcedStarters[cardB], expectedForcedB)
		}

		// Mulligan: P(0 A AND 0 B) = C(54, 7) / C(60, 7)
		expectedMulligan := float64(comb(54, 7)) / float64(comb(60, 7))
		if math.Abs(mulliganOdds-expectedMulligan) > 0.0001 {
			t.Errorf("Mulligan odds = %v, expected %v", mulliganOdds, expectedMulligan)
		}

		t.Logf("Possible A: %v, Forced A: %v", possibleStarters[cardA], forcedStarters[cardA])
		t.Logf("Possible B: %v, Forced B: %v", possibleStarters[cardB], forcedStarters[cardB])
		t.Logf("Mulligan: %v", mulliganOdds)
	})
}

