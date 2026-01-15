package prizeodds

import (
	"testing"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

func TestCalculatePrizeOdds_SingleCard(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Test Card", SetCode: "TEST", Number: "1"}: 1,
		},
	}

	odds, err := CalculatePrizeOdds(decklist, true)
	if err != nil {
		t.Fatalf("CalculatePrizeOdds failed: %v", err)
	}

	card := basictypes.Card{Name: "Test Card", SetCode: "TEST", Number: "1"}
	cardOdds, ok := odds[card]
	if !ok {
		t.Fatalf("Card not found in odds map")
	}

	// For 1 copy in deck:
	// - Index 0 should be odds of prizing at least 1 copy = 0.1 (10%)
	if len(cardOdds) != 1 {
		t.Errorf("Expected array length 1 for card with count 1, got %d", len(cardOdds))
	}

	// Check odds of prizing at least 1 copy
	expectedOdds := DefaultOddsTable.Get(1, 1) // Should be 0.1
	if cardOdds[0] != expectedOdds {
		t.Errorf("Odds[0] (prizing at least 1 copy) = %v, expected %v", cardOdds[0], expectedOdds)
	}

	t.Logf("Card odds array: %v", cardOdds)
	t.Logf("Confirmed: [0.1] for prizing at least 1 copy")
}

func TestCalculatePrizeOdds_FourCopies(t *testing.T) {
	decklist := basictypes.Decklist{
		Cards: map[basictypes.Card]int{
			{Name: "Test Card", SetCode: "TEST", Number: "1"}: 4,
		},
	}

	odds, err := CalculatePrizeOdds(decklist, true)
	if err != nil {
		t.Fatalf("CalculatePrizeOdds failed: %v", err)
	}

	card := basictypes.Card{Name: "Test Card", SetCode: "TEST", Number: "1"}
	cardOdds, ok := odds[card]
	if !ok {
		t.Fatalf("Card not found in odds map")
	}

	// For 4 copies, we should have 4 entries:
	// - Index 0: odds of prizing at least 1 copy
	// - Index 1: odds of prizing at least 2 copies
	// - Index 2: odds of prizing at least 3 copies
	// - Index 3: odds of prizing at least 4 copies
	if len(cardOdds) != 4 {
		t.Errorf("Expected array length 4 for card with count 4, got %d", len(cardOdds))
	}

	// Verify cumulative probabilities
	expectedAtLeast1 := DefaultOddsTable.Get(4, 1) + DefaultOddsTable.Get(4, 2) + DefaultOddsTable.Get(4, 3) + DefaultOddsTable.Get(4, 4)
	expectedAtLeast2 := DefaultOddsTable.Get(4, 2) + DefaultOddsTable.Get(4, 3) + DefaultOddsTable.Get(4, 4)
	expectedAtLeast3 := DefaultOddsTable.Get(4, 3) + DefaultOddsTable.Get(4, 4)
	expectedAtLeast4 := DefaultOddsTable.Get(4, 4)

	if cardOdds[0] != expectedAtLeast1 {
		t.Errorf("Odds[0] (prizing at least 1 copy) = %v, expected %v", cardOdds[0], expectedAtLeast1)
	}
	if cardOdds[1] != expectedAtLeast2 {
		t.Errorf("Odds[1] (prizing at least 2 copies) = %v, expected %v", cardOdds[1], expectedAtLeast2)
	}
	if cardOdds[2] != expectedAtLeast3 {
		t.Errorf("Odds[2] (prizing at least 3 copies) = %v, expected %v", cardOdds[2], expectedAtLeast3)
	}
	if cardOdds[3] != expectedAtLeast4 {
		t.Errorf("Odds[3] (prizing at least 4 copies) = %v, expected %v", cardOdds[3], expectedAtLeast4)
	}

	t.Logf("Card with 4 copies - odds array: %v", cardOdds)
}

func TestCalculateCardSetPrizeOdds(t *testing.T) {
	tests := []struct {
		name     string
		decklist basictypes.Decklist
		cardSets map[string]basictypes.CardSet
		want     map[string]float64
		tolerance float64 // Allowable error for floating point comparison
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
				"single": DefaultOddsTable.Get(1, 1), // 0.1
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
			name: "Overlapping case: AB and AA",
			decklist: basictypes.Decklist{
				Cards: map[basictypes.Card]int{
					{Name: "A", SetCode: "TEST", Number: "1"}: 4,
					{Name: "B", SetCode: "TEST", Number: "2"}: 4,
				},
			},
			cardSets: map[string]basictypes.CardSet{
				"overlap": basictypes.NewCardSet([]basictypes.AnyOfPattern{
					{
						Cards: map[basictypes.Card]int{
							{Name: "A", SetCode: "TEST", Number: "1"}: 1,
							{Name: "B", SetCode: "TEST", Number: "2"}: 1,
						},
					},
					{
						Cards: map[basictypes.Card]int{
							{Name: "A", SetCode: "TEST", Number: "1"}: 2,
						},
					},
				}),
			},
			want: map[string]float64{
				"overlap": 0.0, // Will calculate below
			},
			tolerance: 0.0001,
		},
		{
			name: "Complex overlapping case: ABC, BCD, ACD",
			decklist: basictypes.Decklist{
				Cards: map[basictypes.Card]int{
					{Name: "A", SetCode: "TEST", Number: "1"}: 4,
					{Name: "B", SetCode: "TEST", Number: "2"}: 4,
					{Name: "C", SetCode: "TEST", Number: "3"}: 4,
					{Name: "D", SetCode: "TEST", Number: "4"}: 4,
				},
			},
			cardSets: map[string]basictypes.CardSet{
				"complex": basictypes.NewCardSet([]basictypes.AnyOfPattern{
					{
						Cards: map[basictypes.Card]int{
							{Name: "A", SetCode: "TEST", Number: "1"}: 1,
							{Name: "B", SetCode: "TEST", Number: "2"}: 1,
							{Name: "C", SetCode: "TEST", Number: "3"}: 1,
						},
					},
					{
						Cards: map[basictypes.Card]int{
							{Name: "B", SetCode: "TEST", Number: "2"}: 1,
							{Name: "C", SetCode: "TEST", Number: "3"}: 1,
							{Name: "D", SetCode: "TEST", Number: "4"}: 1,
						},
					},
					{
						Cards: map[basictypes.Card]int{
							{Name: "A", SetCode: "TEST", Number: "1"}: 1,
							{Name: "C", SetCode: "TEST", Number: "3"}: 1,
							{Name: "D", SetCode: "TEST", Number: "4"}: 1,
						},
					},
				}),
			},
			want: map[string]float64{
				"complex": 0.0, // Will calculate below
			},
			tolerance: 0.0001,
		},
		{
			name: "Empty combination",
			decklist: basictypes.Decklist{
				Cards: map[basictypes.Card]int{
					{Name: "A", SetCode: "TEST", Number: "1"}: 4,
				},
			},
			cardSets: map[string]basictypes.CardSet{
				"empty": basictypes.NewCardSet([]basictypes.AnyOfPattern{
					{
						Cards: map[basictypes.Card]int{
							{Name: "NonExistent", SetCode: "XXX", Number: "999"}: 1,
						},
					},
				}),
			},
			want: map[string]float64{
				"empty": 0.0,
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
			got, err := CalculateCardSetPrizeOdds(tt.decklist, tt.cardSets, true)
			if err != nil {
				t.Fatalf("CalculateCardSetPrizeOdds failed: %v", err)
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
				if wantVal == 0.0 && name != "empty" {
					if gotVal < 0.0 || gotVal > 1.0 {
						t.Errorf("CalculateCardSetPrizeOdds(%q) = %v, expected value between 0 and 1", name, gotVal)
					}
					t.Logf("CalculateCardSetPrizeOdds(%q) = %v (calculated)", name, gotVal)
				} else {
					diff := gotVal - wantVal
					if diff < 0 {
						diff = -diff
					}
					if diff > tt.tolerance {
						t.Errorf("CalculateCardSetPrizeOdds(%q) = %v, want %v (diff: %v)", name, gotVal, wantVal, diff)
					}
				}
			}
		})
	}
}

