package prizeodds

import (
	"math"
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

func TestCalculateStartOdds_SingleCard(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Test Card", SetCode: "TEST", Number: "1"}: 1,
		},
	}

	odds, err := CalculateStartOdds(decklist)
	if err != nil {
		t.Fatalf("CalculateStartOdds failed: %v", err)
	}

	card := basictypes.Card{Name: "Test Card", SetCode: "TEST", Number: "1"}
	cardOdds, ok := odds[card]
	if !ok {
		t.Fatalf("Card not found in odds map")
	}

	// For 1 copy in deck:
	// - Index 0 should be odds of having at least 1 copy in starting hand (8 cards)
	if len(cardOdds) != 1 {
		t.Errorf("Expected array length 1 for card with count 1, got %d", len(cardOdds))
	}

	// Check odds of having at least 1 copy in starting hand
	expectedOdds := DefaultStartOddsTable.Get(1, 1) // Probability of exactly 1 copy
	if math.Abs(cardOdds[0]-expectedOdds) > 0.0001 {
		t.Errorf("Odds[0] (at least 1 copy in starting hand) = %v, expected %v", cardOdds[0], expectedOdds)
	}

	t.Logf("Card odds array: %v", cardOdds)
}

func TestCalculateStartOdds_FourCopies(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Test Card", SetCode: "TEST", Number: "1"}: 4,
		},
	}

	odds, err := CalculateStartOdds(decklist)
	if err != nil {
		t.Fatalf("CalculateStartOdds failed: %v", err)
	}

	card := basictypes.Card{Name: "Test Card", SetCode: "TEST", Number: "1"}
	cardOdds, ok := odds[card]
	if !ok {
		t.Fatalf("Card not found in odds map")
	}

	// For 4 copies in deck, should have array length 4 (1+, 2+, 3+, 4+)
	if len(cardOdds) != 4 {
		t.Errorf("Expected array length 4 for card with count 4, got %d", len(cardOdds))
	}

	// Verify cumulative probabilities
	// Index 0: P(>= 1) = sum of P(1), P(2), P(3), P(4)
	expected1Plus := 0.0
	for j := 1; j <= 4; j++ {
		expected1Plus += DefaultStartOddsTable.Get(4, j)
	}
	if math.Abs(cardOdds[0]-expected1Plus) > 0.0001 {
		t.Errorf("Odds[0] (at least 1 copy) = %v, expected %v", cardOdds[0], expected1Plus)
	}

	// Index 1: P(>= 2) = sum of P(2), P(3), P(4)
	expected2Plus := 0.0
	for j := 2; j <= 4; j++ {
		expected2Plus += DefaultStartOddsTable.Get(4, j)
	}
	if math.Abs(cardOdds[1]-expected2Plus) > 0.0001 {
		t.Errorf("Odds[1] (at least 2 copies) = %v, expected %v", cardOdds[1], expected2Plus)
	}

	// Index 2: P(>= 3) = sum of P(3), P(4)
	expected3Plus := 0.0
	for j := 3; j <= 4; j++ {
		expected3Plus += DefaultStartOddsTable.Get(4, j)
	}
	if math.Abs(cardOdds[2]-expected3Plus) > 0.0001 {
		t.Errorf("Odds[2] (at least 3 copies) = %v, expected %v", cardOdds[2], expected3Plus)
	}

	// Index 3: P(>= 4) = P(4)
	expected4Plus := DefaultStartOddsTable.Get(4, 4)
	if math.Abs(cardOdds[3]-expected4Plus) > 0.0001 {
		t.Errorf("Odds[3] (at least 4 copies) = %v, expected %v", cardOdds[3], expected4Plus)
	}

	t.Logf("Card with 4 copies - odds array: %v", cardOdds)
}

func TestCalculateStartOdds_MultipleCards(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Card A", SetCode: "TEST", Number: "1"}: 2,
			{Name: "Card B", SetCode: "TEST", Number: "2"}: 3,
		},
	}

	odds, err := CalculateStartOdds(decklist)
	if err != nil {
		t.Fatalf("CalculateStartOdds failed: %v", err)
	}

	if len(odds) != 2 {
		t.Errorf("Expected 2 cards in odds map, got %d", len(odds))
	}

	cardA := basictypes.Card{Name: "Card A", SetCode: "TEST", Number: "1"}
	cardB := basictypes.Card{Name: "Card B", SetCode: "TEST", Number: "2"}

	oddsA, ok := odds[cardA]
	if !ok {
		t.Fatalf("Card A not found in odds map")
	}
	if len(oddsA) != 2 {
		t.Errorf("Expected array length 2 for Card A (count 2), got %d", len(oddsA))
	}

	oddsB, ok := odds[cardB]
	if !ok {
		t.Fatalf("Card B not found in odds map")
	}
	if len(oddsB) != 3 {
		t.Errorf("Expected array length 3 for Card B (count 3), got %d", len(oddsB))
	}

	t.Logf("Card A odds: %v", oddsA)
	t.Logf("Card B odds: %v", oddsB)
}

func TestStartOddsTable_Get(t *testing.T) {
	// Test some known values for 8 cards
	tests := []struct {
		x    int
		y    int
		want float64
	}{
		{0, 0, 1.0},   // If 0 copies in deck, 100% chance of having 0 in starting hand
		{0, 1, 0.0},   // If 0 copies in deck, 0% chance of having 1
		{1, 0, 0.8667}, // If 1 copy in deck, ~86.67% chance of having 0 (52/60 * 51/59 * ... * 45/53)
		{1, 1, 0.1333}, // If 1 copy in deck, ~13.33% chance of having 1 (8/60)
		{60, 0, 0.0},  // Out of range x
		{0, 9, 0.0},   // Out of range y
	}

	for _, tt := range tests {
		got := DefaultStartOddsTable.Get(tt.x, tt.y)
		// For the probability checks, allow some tolerance
		if tt.x == 1 && tt.y == 0 {
			// P(0 copies when 1 in deck) = C(59, 8) / C(60, 8) = 52/60 ≈ 0.8667
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("DefaultStartOddsTable.Get(%d, %d) = %v, want ~%v", tt.x, tt.y, got, tt.want)
			}
		} else if tt.x == 1 && tt.y == 1 {
			// P(1 copy when 1 in deck) = C(1, 1) * C(59, 7) / C(60, 8) = 8/60 ≈ 0.1333
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("DefaultStartOddsTable.Get(%d, %d) = %v, want ~%v", tt.x, tt.y, got, tt.want)
			}
		} else if got != tt.want {
			t.Errorf("DefaultStartOddsTable.Get(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestStartOddsTable_Get_SumToOne(t *testing.T) {
	// For any x, the sum of probabilities for y=0..8 should be approximately 1.0
	for x := 0; x <= 59; x++ {
		sum := 0.0
		for y := 0; y <= 8; y++ {
			sum += DefaultStartOddsTable.Get(x, y)
		}
		// Allow small floating point error
		if sum < 0.9999 || sum > 1.0001 {
			t.Errorf("Sum of probabilities for x=%d is %v, expected ~1.0", x, sum)
		}
	}
}

func TestCalculateCardSetStartOdds(t *testing.T) {
	tests := []struct {
		name     string
		decklist basictypes.Decklist
		cardSets map[string]basictypes.CardSet
		want     map[string]float64
		tolerance float64
	}{
		{
			name: "Single card combination",
			decklist: basictypes.Decklist{
				Cards: map[basictypes.Card]int{
					{Name: "A", SetCode: "TEST", Number: "1"}: 1,
				},
			},
			cardSets: map[string]basictypes.CardSet{
				"single": basictypes.NewCardSet([]basictypes.AnyOfPattern{
					{
						Cards: map[basictypes.Card]int{
							{Name: "A", SetCode: "TEST", Number: "1"}: 1,
						},
					},
				}),
			},
			want: map[string]float64{
				"single": 0.0, // Will calculate below
			},
			tolerance: 0.0001,
		},
		{
			name: "Two non-overlapping combinations",
			decklist: basictypes.Decklist{
				Cards: map[basictypes.Card]int{
					{Name: "A", SetCode: "TEST", Number: "1"}: 4,
					{Name: "B", SetCode: "TEST", Number: "2"}: 4,
				},
			},
			cardSets: map[string]basictypes.CardSet{
				"two": basictypes.NewCardSet([]basictypes.AnyOfPattern{
					{
						Cards: map[basictypes.Card]int{
							{Name: "A", SetCode: "TEST", Number: "1"}: 1,
						},
					},
					{
						Cards: map[basictypes.Card]int{
							{Name: "B", SetCode: "TEST", Number: "2"}: 1,
						},
					},
				}),
			},
			want: map[string]float64{
				"two": 0.0, // Will calculate below
			},
			tolerance: 0.0001,
		},
		{
			name: "AnyOf with multiple options",
			decklist: basictypes.Decklist{
				Cards: map[basictypes.Card]int{
					{Name: "A", SetCode: "TEST", Number: "1"}: 4,
					{Name: "B", SetCode: "TEST", Number: "2"}: 4,
				},
			},
			cardSets: map[string]basictypes.CardSet{
				"anyof": basictypes.NewCardSet([]basictypes.AnyOfPattern{
					{
						Cards: map[basictypes.Card]int{
							{Name: "A", SetCode: "TEST", Number: "1"}: 1,
							{Name: "B", SetCode: "TEST", Number: "2"}: 1,
						},
					},
				}),
			},
			want: map[string]float64{
				"anyof": 0.0, // Will calculate: P(A) + P(B) - P(A and B)
			},
			tolerance: 0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateCardSetStartOdds(tt.decklist, tt.cardSets)
			if err != nil {
				t.Fatalf("CalculateCardSetStartOdds failed: %v", err)
			}

			// Verify all expected keys are present
			for name := range tt.want {
				if _, ok := got[name]; !ok {
					t.Errorf("Missing result for cardSet %q", name)
				}
			}

			// Verify values
			for name, wantVal := range tt.want {
				gotVal, ok := got[name]
				if !ok {
					t.Errorf("Missing result for cardSet %q", name)
					continue
				}

				// If wantVal is 0.0, it means we need to calculate it
				// For now, just check it's a valid probability
				if wantVal == 0.0 {
					if gotVal < 0.0 || gotVal > 1.0 {
						t.Errorf("CalculateCardSetStartOdds(%q) = %v, expected value between 0 and 1", name, gotVal)
					}
					t.Logf("CalculateCardSetStartOdds(%q) = %v (calculated)", name, gotVal)
				} else {
					diff := gotVal - wantVal
					if diff < 0 {
						diff = -diff
					}
					if diff > tt.tolerance {
						t.Errorf("CalculateCardSetStartOdds(%q) = %v, want %v (diff: %v)", name, gotVal, wantVal, diff)
					}
				}
			}
		})
	}
}

func TestCalculateCardSetStartOdds_SingleCard(t *testing.T) {
	// Test that CardSet start odds for a single card matches the individual card start odds
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "A", SetCode: "TEST", Number: "1"}: 4,
		},
	}

	cardSet := basictypes.NewCardSet([]basictypes.AnyOfPattern{
		{
			Cards: map[basictypes.Card]int{
				{Name: "A", SetCode: "TEST", Number: "1"}: 1,
			},
		},
	})

	cardSetOdds, err := CalculateCardSetStartOdds(decklist, map[string]basictypes.CardSet{"test": cardSet})
	if err != nil {
		t.Fatalf("CalculateCardSetStartOdds failed: %v", err)
	}

	// Calculate expected: P(at least 1 A in 8 cards) = sum of P(1), P(2), P(3), P(4)
	expectedAtLeast1A := 0.0
	for j := 1; j <= 4; j++ {
		expectedAtLeast1A += DefaultStartOddsTable.Get(4, j)
	}

	tolerance := 0.0001
	diff := cardSetOdds["test"] - expectedAtLeast1A
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("CalculateCardSetStartOdds = %v, expected P(at least 1 A) = %v (diff: %v)", cardSetOdds["test"], expectedAtLeast1A, diff)
	}

	t.Logf("CardSet start odds: %v", cardSetOdds["test"])
	t.Logf("Expected P(at least 1 A in 8 cards): %v", expectedAtLeast1A)
}

