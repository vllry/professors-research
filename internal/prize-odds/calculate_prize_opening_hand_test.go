package prizeodds

import (
	"math"
	"testing"

	basictypes "github.com/vllry/professors-research/pkg/types"
)

func TestCalculatePrizeOddsWithOpeningHand_AdjustmentDirection(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "BasicA", SetCode: "TEST", Number: "1"}:    4,
			{Name: "NonBasicB", SetCode: "TEST", Number: "2"}: 4,
		},
	}

	basicA := basictypes.Card{Name: "BasicA", SetCode: "TEST", Number: "1"}
	nonBasicB := basictypes.Card{Name: "NonBasicB", SetCode: "TEST", Number: "2"}
	basicPokemonCards := map[basictypes.Card]bool{
		basicA: true,
	}

	unadjusted, err := CalculatePrizeOdds(decklist, true)
	if err != nil {
		t.Fatalf("CalculatePrizeOdds failed: %v", err)
	}
	adjusted, err := CalculatePrizeOddsWithOpeningHand(decklist, true, basicPokemonCards)
	if err != nil {
		t.Fatalf("CalculatePrizeOddsWithOpeningHand failed: %v", err)
	}

	// Conditioning on "hand has at least 1 basic" should:
	// - decrease prize odds for Basic Pokémon
	// - increase prize odds for non-basic cards
	eps := 1e-9

	if !(adjusted[basicA][0] < unadjusted[basicA][0]-eps) {
		t.Fatalf("Expected basic card to be less likely prized after conditioning. adjusted=%v unadjusted=%v", adjusted[basicA][0], unadjusted[basicA][0])
	}
	if !(adjusted[nonBasicB][0] > unadjusted[nonBasicB][0]+eps) {
		t.Fatalf("Expected non-basic card to be more likely prized after conditioning. adjusted=%v unadjusted=%v", adjusted[nonBasicB][0], unadjusted[nonBasicB][0])
	}
}

func TestCalculatePrizeOddsWithOpeningHand_ExactDistributionSumsToOne(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "BasicA", SetCode: "TEST", Number: "1"}:    4,
			{Name: "NonBasicB", SetCode: "TEST", Number: "2"}: 4,
		},
	}

	basicA := basictypes.Card{Name: "BasicA", SetCode: "TEST", Number: "1"}
	basicPokemonCards := map[basictypes.Card]bool{
		basicA: true,
	}

	adjusted, err := CalculatePrizeOddsWithOpeningHand(decklist, true, basicPokemonCards)
	if err != nil {
		t.Fatalf("CalculatePrizeOddsWithOpeningHand failed: %v", err)
	}

	// Reconstruct exact P(X=j) from cumulative P(X>=k) output:
	// - p0 = 1 - P(X>=1)
	// - pj = P(X>=j) - P(X>=j+1) for 1<=j<m
	// - pm = P(X>=m)
	cum := adjusted[basicA]
	if len(cum) != 4 {
		t.Fatalf("Expected 4 cumulative entries, got %d", len(cum))
	}

	p0 := 1.0 - cum[0]
	p1 := cum[0] - cum[1]
	p2 := cum[1] - cum[2]
	p3 := cum[2] - cum[3]
	p4 := cum[3]

	sum := p0 + p1 + p2 + p3 + p4
	if math.Abs(sum-1.0) > 1e-9 {
		t.Fatalf("Expected reconstructed exact distribution to sum to 1. got=%v (p0..p4=%v,%v,%v,%v,%v)", sum, p0, p1, p2, p3, p4)
	}
}

func TestCalculatePrizeOddsWithOpeningHand_NotPrizedInversionHolds(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "BasicA", SetCode: "TEST", Number: "1"}:    4,
			{Name: "NonBasicB", SetCode: "TEST", Number: "2"}: 4,
		},
	}

	basicA := basictypes.Card{Name: "BasicA", SetCode: "TEST", Number: "1"}
	nonBasicB := basictypes.Card{Name: "NonBasicB", SetCode: "TEST", Number: "2"}
	basicPokemonCards := map[basictypes.Card]bool{
		basicA: true,
	}

	prized, err := CalculatePrizeOddsWithOpeningHand(decklist, true, basicPokemonCards)
	if err != nil {
		t.Fatalf("CalculatePrizeOddsWithOpeningHand(prized=true) failed: %v", err)
	}
	notPrized, err := CalculatePrizeOddsWithOpeningHand(decklist, false, basicPokemonCards)
	if err != nil {
		t.Fatalf("CalculatePrizeOddsWithOpeningHand(prized=false) failed: %v", err)
	}

	tolerance := 1e-9
	for _, card := range []basictypes.Card{basicA, nonBasicB} {
		p := prized[card]
		np := notPrized[card]
		if len(p) != len(np) {
			t.Fatalf("Expected same array lengths for prized and not-prized. card=%v prized=%d notPrized=%d", card, len(p), len(np))
		}
		for i := 0; i < len(p); i++ {
			expected := 1.0 - p[len(p)-1-i]
			if math.Abs(np[i]-expected) > tolerance {
				t.Fatalf("Inversion failed. card=%v index=%d notPrized=%v expected=%v", card, i, np[i], expected)
			}
		}
	}
}

func TestCalculatePrizeOddsWithOpeningHand_BasicCountZeroFallsBack(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "A", SetCode: "TEST", Number: "1"}: 4,
			{Name: "B", SetCode: "TEST", Number: "2"}: 4,
		},
	}

	unadjusted, err := CalculatePrizeOdds(decklist, true)
	if err != nil {
		t.Fatalf("CalculatePrizeOdds failed: %v", err)
	}
	adjusted, err := CalculatePrizeOddsWithOpeningHand(decklist, true, map[basictypes.Card]bool{})
	if err != nil {
		t.Fatalf("CalculatePrizeOddsWithOpeningHand failed: %v", err)
	}

	tolerance := 0.0
	for card, odds := range unadjusted {
		got := adjusted[card]
		if len(got) != len(odds) {
			t.Fatalf("Length mismatch for card %v. got=%d want=%d", card, len(got), len(odds))
		}
		for i := range odds {
			if math.Abs(got[i]-odds[i]) > tolerance {
				t.Fatalf("Odds mismatch for card %v at index %d. got=%v want=%v", card, i, got[i], odds[i])
			}
		}
	}
}

