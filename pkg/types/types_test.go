package types

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestCardSet_Expand(t *testing.T) {
	// Load the dragapult decklist
	decklistPath := filepath.Join("..", "..", "testdata", "dragapult_MEG_1.txt")
	decklistContent, err := os.ReadFile(decklistPath)
	if err != nil {
		t.Fatalf("Failed to read decklist file: %v", err)
	}

	decklist, err := NewDecklistFromLive(string(decklistContent))
	if err != nil {
		t.Fatalf("Failed to parse decklist: %v", err)
	}

	// Define some cards from the decklist for testing
	dreepy := Card{SetCode: "TWM", Number: "128", Name: "Dreepy"}
	drakloak := Card{SetCode: "TWM", Number: "129", Name: "Drakloak"}
	dragapultEx := Card{SetCode: "TWM", Number: "130", Name: "Dragapult ex"}
	ultraBall := Card{SetCode: "SVI", Number: "196", Name: "Ultra Ball"}
	buddyPoffin := Card{SetCode: "TEF", Number: "144", Name: "Buddy-Buddy Poffin"}
	lilliesDetermination := Card{SetCode: "MEG", Number: "119", Name: "Lillie's Determination"}
	iono := Card{SetCode: "PAL", Number: "185", Name: "Iono"}
	psychicEnergy := Card{SetCode: "EVO", Number: "95", Name: "Basic {P} Energy"}
	luminousEnergy := Card{SetCode: "PAL", Number: "191", Name: "Luminous Energy"}

	tests := []struct {
		name     string
		cardSet  CardSet
		want     []map[Card]int // Expected combinations as maps of Card to count
	}{
		{
			name: "Single AnyOf with one card",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						dreepy: 1,
					},
				},
			}),
			want: []map[Card]int{
				{dreepy: 1},
			},
		},
		{
			name: "Single AnyOf with two options",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						psychicEnergy: 1,
						luminousEnergy: 1,
					},
				},
			}),
			want: []map[Card]int{
				{psychicEnergy: 1},
				{luminousEnergy: 1},
			},
		},
		{
			name: "Two AnyOfs, each with one option",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						dreepy: 1,
					},
				},
				{
					Cards: map[Card]int{
						ultraBall: 1,
					},
				},
			}),
			want: []map[Card]int{
				{dreepy: 1, ultraBall: 1},
			},
		},
		{
			name: "Two AnyOfs with multiple options",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						lilliesDetermination: 1,
						iono: 1,
					},
				},
				{
					Cards: map[Card]int{
						dreepy: 1,
						buddyPoffin: 1,
					},
				},
			}),
			want: []map[Card]int{
				{lilliesDetermination: 1, dreepy: 1},
				{lilliesDetermination: 1, buddyPoffin: 1},
				{iono: 1, dreepy: 1},
				{iono: 1, buddyPoffin: 1},
			},
		},
		{
			name: "AnyOf with count > 1",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						dreepy: 2,
					},
				},
			}),
			want: []map[Card]int{
				{dreepy: 2},
			},
		},
		{
			name: "Complex case: three AnyOfs",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						dreepy: 1,
					},
				},
				{
					Cards: map[Card]int{
						ultraBall: 1,
					},
				},
				{
					Cards: map[Card]int{
						psychicEnergy: 1,
						luminousEnergy: 1,
					},
				},
			}),
			want: []map[Card]int{
				{dreepy: 1, ultraBall: 1, psychicEnergy: 1},
				{dreepy: 1, ultraBall: 1, luminousEnergy: 1},
			},
		},
		{
			name: "Card not in decklist should be filtered out",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						Card{SetCode: "XXX", Number: "999", Name: "Non-existent Card"}: 1,
					},
				},
			}),
			want: []map[Card]int{},
		},
		{
			name: "Card count exceeds decklist should be filtered out",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						dreepy: 10, // Only 4 Dreepy in deck
					},
				},
			}),
			want: []map[Card]int{},
		},
		{
			name: "Multiple cards in one AnyOf with overlapping decklist counts",
			cardSet: NewCardSet([]AnyOfPattern{
				{
					Cards: map[Card]int{
						dreepy: 1,
						drakloak: 1,
					},
				},
				{
					Cards: map[Card]int{
						dreepy: 1,
						dragapultEx: 1,
					},
				},
			}),
			want: []map[Card]int{
				{dreepy: 2}, // 1 from first AnyOf + 1 from second AnyOf
				{dreepy: 1, dragapultEx: 1},
				{drakloak: 1, dreepy: 1},
				{drakloak: 1, dragapultEx: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cardSet.Expand(decklist)

			// Convert got.Combinations to []map[Card]int for comparison
			gotCombinations := make([]map[Card]int, len(got.Combinations))
			for i, combination := range got.Combinations {
				gotCombinations[i] = combinationToMap(combination)
			}

			// Normalize both slices for comparison (order-independent)
			gotNormalized := normalizeCombinations(gotCombinations)
			wantNormalized := normalizeCombinations(tt.want)

			if !reflect.DeepEqual(gotNormalized, wantNormalized) {
				t.Errorf("Expand() combinations don't match expected")
				t.Logf("Got %d combinations:", len(gotCombinations))
				for i, combo := range gotCombinations {
					t.Logf("  [%d] %v", i, combo)
				}
				t.Logf("Want %d combinations:", len(tt.want))
				for i, combo := range tt.want {
					t.Logf("  [%d] %v", i, combo)
				}
			}

			// Verify all combinations are valid (card counts don't exceed decklist)
			for i, combination := range got.Combinations {
				combinationCounts := combinationToMap(combination)
				for card, count := range combinationCounts {
					if decklist.Cards[card] < count {
						t.Errorf("Combination %d contains %d copies of %v, but decklist only has %d", i, count, card, decklist.Cards[card])
					}
				}
			}
		})
	}
}

// combinationToMap converts a slice of cards to a map of card counts
func combinationToMap(combination []Card) map[Card]int {
	counts := make(map[Card]int)
	for _, card := range combination {
		counts[card]++
	}
	return counts
}

// normalizeCombinations normalizes a slice of combinations for comparison
// by converting each to a canonical string representation and sorting
func normalizeCombinations(combinations []map[Card]int) []string {
	normalized := make([]string, len(combinations))
	for i, combo := range combinations {
		normalized[i] = combinationToString(combo)
	}
	// Sort for consistent comparison
	sort.Strings(normalized)
	return normalized
}

// combinationToString creates a string representation of a card count map for comparison
func combinationToString(counts map[Card]int) string {
	var parts []string
	for card, count := range counts {
		parts = append(parts, card.Name+":"+card.SetCode+":"+card.Number+"x"+strconv.Itoa(count))
	}
	// Sort parts for consistent comparison
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

