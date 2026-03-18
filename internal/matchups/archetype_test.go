package matchups

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/vllry/professors-research/pkg/types"
)

func makeDeck(cards map[string]int) types.Decklist {
	dl := types.Decklist{Cards: make(map[types.Card]int)}
	i := 0
	for name, count := range cards {
		dl.Cards[types.Card{Name: name, SetCode: "TST", Number: fmt.Sprintf("%d", i)}] = count
		i++
	}
	return dl
}

func TestDeckCardNameCounts(t *testing.T) {
	dl := types.Decklist{Cards: map[types.Card]int{
		{Name: "Charizard ex", SetCode: "MEW", Number: "6"}:  2,
		{Name: "Charizard ex", SetCode: "OBF", Number: "54"}: 1,
		{Name: "Arcanine ex", SetCode: "SVI", Number: "32"}:  3,
	}}

	counts := DeckCardNameCounts(dl)
	if counts["Charizard ex"] != 3 {
		t.Errorf("expected Charizard ex count 3, got %d", counts["Charizard ex"])
	}
	if counts["Arcanine ex"] != 3 {
		t.Errorf("expected Arcanine ex count 3, got %d", counts["Arcanine ex"])
	}
}

func TestClassifyDeck_FirstMatchWins(t *testing.T) {
	archetypes := []Archetype{
		{Name: "Charizard Arcanine", Requires: map[string]int{"Charizard ex": 1, "Arcanine ex": 1}},
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
		{Name: "Arcanine", Requires: map[string]int{"Arcanine ex": 1}},
	}

	tests := []struct {
		name string
		deck map[string]int
		want string
	}{
		{
			name: "matches first (more specific) archetype",
			deck: map[string]int{"Charizard ex": 2, "Arcanine ex": 1, "Filler": 57},
			want: "Charizard Arcanine",
		},
		{
			name: "matches second archetype when first fails",
			deck: map[string]int{"Charizard ex": 2, "Filler": 58},
			want: "Charizard",
		},
		{
			name: "matches third archetype",
			deck: map[string]int{"Arcanine ex": 3, "Filler": 57},
			want: "Arcanine",
		},
		{
			name: "no match returns Unknown",
			deck: map[string]int{"Filler": 60},
			want: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck := makeDeck(tt.deck)
			got := ClassifyDeck(deck, archetypes)
			if got != tt.want {
				t.Errorf("ClassifyDeck() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassifyDeck_MinCountEnforced(t *testing.T) {
	archetypes := []Archetype{
		{Name: "Double Charizard", Requires: map[string]int{"Charizard ex": 2}},
	}

	deck := makeDeck(map[string]int{"Charizard ex": 1, "Filler": 59})
	got := ClassifyDeck(deck, archetypes)
	if got != "Unknown" {
		t.Errorf("expected Unknown for 1 copy when 2 required, got %q", got)
	}

	deck2 := makeDeck(map[string]int{"Charizard ex": 2, "Filler": 58})
	got2 := ClassifyDeck(deck2, archetypes)
	if got2 != "Double Charizard" {
		t.Errorf("expected Double Charizard for 2 copies, got %q", got2)
	}
}

func TestClassifyVariants(t *testing.T) {
	variants := []map[string]int{
		{"Arcanine ex": 2, "Charizard ex": 2},
		{"Arcanine ex": 1},
	}

	tests := []struct {
		name string
		deck map[string]int
		want []string
	}{
		{
			name: "matches both variants (inclusive)",
			deck: map[string]int{"Charizard ex": 2, "Arcanine ex": 2, "Filler": 56},
			want: []string{"0", "1"},
		},
		{
			name: "matches only second variant",
			deck: map[string]int{"Charizard ex": 2, "Arcanine ex": 1, "Filler": 57},
			want: []string{"1"},
		},
		{
			name: "no variant match returns other",
			deck: map[string]int{"Charizard ex": 2, "Filler": 58},
			want: []string{"other"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck := makeDeck(tt.deck)
			got := ClassifyVariants(deck, variants)
			if len(got) != len(tt.want) {
				t.Errorf("ClassifyVariants() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ClassifyVariants() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestClassifyVariants_EmptyList(t *testing.T) {
	deck := makeDeck(map[string]int{"Charizard ex": 2, "Filler": 58})
	got := ClassifyVariants(deck, nil)
	if len(got) != 1 || got[0] != "other" {
		t.Errorf("expected [other] for nil variants, got %v", got)
	}
}

func TestClassifyVariants_MultipleOverlappingVariants(t *testing.T) {
	variants := []map[string]int{
		{"CardX": 1},
		{"CardY": 1},
		{"CardZ": 1},
	}

	tests := []struct {
		name string
		deck map[string]int
		want []string
	}{
		{
			name: "deck with X and Y matches both variants",
			deck: map[string]int{"CardX": 1, "CardY": 1, "Filler": 58},
			want: []string{"0", "1"},
		},
		{
			name: "deck with all three matches all variants",
			deck: map[string]int{"CardX": 1, "CardY": 1, "CardZ": 1, "Filler": 57},
			want: []string{"0", "1", "2"},
		},
		{
			name: "deck with only Y matches only second variant",
			deck: map[string]int{"CardY": 1, "Filler": 59},
			want: []string{"1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck := makeDeck(tt.deck)
			got := ClassifyVariants(deck, variants)
			if len(got) != len(tt.want) {
				t.Errorf("ClassifyVariants() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ClassifyVariants() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestLoadArchetypes(t *testing.T) {
	data := `[
		{"name": "Alpha", "requires": {"A": 1}},
		{"name": "Beta", "requires": {"B": 2, "C": 1}}
	]`

	dir := t.TempDir()
	path := filepath.Join(dir, "archetypes.json")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	archetypes, err := LoadArchetypes(path)
	if err != nil {
		t.Fatalf("LoadArchetypes() error: %v", err)
	}
	if len(archetypes) != 2 {
		t.Fatalf("expected 2 archetypes, got %d", len(archetypes))
	}
	if archetypes[0].Name != "Alpha" {
		t.Errorf("expected first archetype Alpha, got %q", archetypes[0].Name)
	}
	if archetypes[1].Requires["B"] != 2 {
		t.Errorf("expected Beta requires B=2, got %d", archetypes[1].Requires["B"])
	}

	// Verify round-trip via JSON
	out, err := json.Marshal(archetypes)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip []Archetype
	if err := json.Unmarshal(out, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if len(roundTrip) != 2 {
		t.Fatalf("round-trip produced %d archetypes", len(roundTrip))
	}
}
