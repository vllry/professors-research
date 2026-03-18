package matchups

import (
	"testing"

	"github.com/vllry/professors-research/pkg/types"
)

func buildVariantTestData() map[string]types.Decklist {
	return map[string]types.Decklist{
		"alice": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Noctowl", SetCode: "T", Number: "10"}:     1,
			{Name: "Hoothoot", SetCode: "T", Number: "11"}:    2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      51,
		}},
		"bob": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Noctowl", SetCode: "T", Number: "10"}:     1,
			{Name: "Hoothoot", SetCode: "T", Number: "11"}:    2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      51,
		}},
		"carol": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Pigeot ex", SetCode: "T", Number: "20"}:   1,
			{Name: "Pidgey", SetCode: "T", Number: "21"}:      1,
			{Name: "Filler", SetCode: "T", Number: "99"}:      52,
		}},
		"dave": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Pigeot ex", SetCode: "T", Number: "20"}:   1,
			{Name: "Pidgey", SetCode: "T", Number: "21"}:      1,
			{Name: "Filler", SetCode: "T", Number: "99"}:      52,
		}},
		"eve": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Pigeot ex", SetCode: "T", Number: "20"}:   1,
			{Name: "Pidgey", SetCode: "T", Number: "21"}:      1,
			{Name: "Filler", SetCode: "T", Number: "99"}:      52,
		}},
		// Non-archetype deck
		"frank": {Cards: map[types.Card]int{
			{Name: "Dragapult ex", SetCode: "T", Number: "30"}: 3,
			{Name: "Filler", SetCode: "T", Number: "99"}:       57,
		}},
	}
}

func variantTestArchetypes() []Archetype {
	return []Archetype{
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 2}},
		{Name: "Dragapult", Requires: map[string]int{"Dragapult ex": 2}},
	}
}

func TestComputeDeckVariants_PackageGrouping(t *testing.T) {
	decklists := buildVariantTestData()
	archetypes := variantTestArchetypes()

	result, err := computeDeckVariantsFromData(decklists, archetypes, "Charizard", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalDecks != 5 {
		t.Errorf("TotalDecks = %d, want 5", result.TotalDecks)
	}

	// Core cards: Charizard ex=3, Charmander=3 are shared by all 5 decks.
	if result.CoreCards["Charizard ex"] != 3 {
		t.Errorf("CoreCards[Charizard ex] = %d, want 3", result.CoreCards["Charizard ex"])
	}
	if result.CoreCards["Charmander"] != 3 {
		t.Errorf("CoreCards[Charmander] = %d, want 3", result.CoreCards["Charmander"])
	}
	if _, ok := result.CoreCards["Noctowl"]; ok {
		t.Error("Noctowl should not be a core card")
	}
	if _, ok := result.CoreCards["Pigeot ex"]; ok {
		t.Error("Pigeot ex should not be a core card")
	}

	// All flex cards partition decks the same way ({alice,bob} vs {carol,dave,eve}),
	// so they form a single package with 2 configs.
	if len(result.Packages) != 1 {
		t.Fatalf("len(Packages) = %d, want 1", len(result.Packages))
	}

	pkg := result.Packages[0]

	// Package should contain all flex card names.
	wantCards := map[string]bool{"Filler": true, "Hoothoot": true, "Noctowl": true, "Pidgey": true, "Pigeot ex": true}
	if len(pkg.Cards) != len(wantCards) {
		t.Errorf("package cards = %v, want %v", pkg.Cards, wantCards)
	}
	for _, c := range pkg.Cards {
		if !wantCards[c] {
			t.Errorf("unexpected card %q in package", c)
		}
	}

	if len(pkg.Configs) != 2 {
		t.Fatalf("len(Configs) = %d, want 2", len(pkg.Configs))
	}

	// Pigeot config (3 decks) should be first.
	if pkg.Configs[0].Count != 3 {
		t.Errorf("Configs[0].Count = %d, want 3", pkg.Configs[0].Count)
	}
	if pkg.Configs[0].Cards["Pigeot ex"] != 1 {
		t.Errorf("Configs[0] should contain Pigeot ex=1, got %d", pkg.Configs[0].Cards["Pigeot ex"])
	}

	// Noctowl config (2 decks) should be second.
	if pkg.Configs[1].Count != 2 {
		t.Errorf("Configs[1].Count = %d, want 2", pkg.Configs[1].Count)
	}
	if pkg.Configs[1].Cards["Noctowl"] != 1 {
		t.Errorf("Configs[1] should contain Noctowl=1, got %d", pkg.Configs[1].Cards["Noctowl"])
	}
}

func TestComputeDeckVariants_IndependentPackages(t *testing.T) {
	// Two independent axes of variation: draw engine (UB/Hilda) and tech line
	// (Pidgeot/Pidgey vs Noctowl/Hoothoot). Tech lines have the same total
	// card count so Filler stays constant and is core.
	decklists := map[string]types.Decklist{
		"d1": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Ultra Ball", SetCode: "T", Number: "30"}:  2,
			{Name: "Hilda", SetCode: "T", Number: "31"}:       2,
			{Name: "Pidgeot ex", SetCode: "T", Number: "20"}:  1,
			{Name: "Pidgey", SetCode: "T", Number: "21"}:      2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      47,
		}},
		"d2": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Ultra Ball", SetCode: "T", Number: "30"}:  2,
			{Name: "Hilda", SetCode: "T", Number: "31"}:       2,
			{Name: "Noctowl", SetCode: "T", Number: "10"}:     1,
			{Name: "Hoothoot", SetCode: "T", Number: "11"}:    2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      47,
		}},
		"d3": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Ultra Ball", SetCode: "T", Number: "30"}:  3,
			{Name: "Hilda", SetCode: "T", Number: "31"}:       1,
			{Name: "Pidgeot ex", SetCode: "T", Number: "20"}:  1,
			{Name: "Pidgey", SetCode: "T", Number: "21"}:      2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      47,
		}},
		"d4": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Charmander", SetCode: "T", Number: "2"}:   3,
			{Name: "Ultra Ball", SetCode: "T", Number: "30"}:  3,
			{Name: "Hilda", SetCode: "T", Number: "31"}:       1,
			{Name: "Noctowl", SetCode: "T", Number: "10"}:     1,
			{Name: "Hoothoot", SetCode: "T", Number: "11"}:    2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      47,
		}},
	}
	archetypes := variantTestArchetypes()

	result, err := computeDeckVariantsFromData(decklists, archetypes, "Charizard", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalDecks != 4 {
		t.Errorf("TotalDecks = %d, want 4", result.TotalDecks)
	}

	// Core: Charizard ex=3, Charmander=3, Filler=47
	if result.CoreCards["Filler"] != 47 {
		t.Errorf("CoreCards[Filler] = %d, want 47", result.CoreCards["Filler"])
	}

	// Two independent packages because the two axes cross-cut the decks.
	if len(result.Packages) != 2 {
		t.Fatalf("len(Packages) = %d, want 2", len(result.Packages))
	}

	// Packages are sorted by card count descending. The tech-line package has
	// 4 cards (Hoothoot, Noctowl, Pidgey, Pidgeot ex), the draw package has 2
	// (Hilda, Ultra Ball).
	techPkg := result.Packages[0]
	drawPkg := result.Packages[1]

	if len(techPkg.Cards) != 4 {
		t.Errorf("tech package has %d cards, want 4: %v", len(techPkg.Cards), techPkg.Cards)
	}
	if len(drawPkg.Cards) != 2 {
		t.Errorf("draw package has %d cards, want 2: %v", len(drawPkg.Cards), drawPkg.Cards)
	}

	// Each package should have exactly 2 configs (each appearing in 2 decks).
	if len(techPkg.Configs) != 2 {
		t.Fatalf("tech package configs = %d, want 2", len(techPkg.Configs))
	}
	if len(drawPkg.Configs) != 2 {
		t.Fatalf("draw package configs = %d, want 2", len(drawPkg.Configs))
	}

	for _, cfg := range techPkg.Configs {
		if cfg.Count != 2 {
			t.Errorf("tech config count = %d, want 2", cfg.Count)
		}
	}
	for _, cfg := range drawPkg.Configs {
		if cfg.Count != 2 {
			t.Errorf("draw config count = %d, want 2", cfg.Count)
		}
	}

	// Verify draw package content: one config has UB=2,Hilda=2 and the other UB=3,Hilda=1.
	foundUB2 := false
	foundUB3 := false
	for _, cfg := range drawPkg.Configs {
		if cfg.Cards["Ultra Ball"] == 2 && cfg.Cards["Hilda"] == 2 {
			foundUB2 = true
		}
		if cfg.Cards["Ultra Ball"] == 3 && cfg.Cards["Hilda"] == 1 {
			foundUB3 = true
		}
	}
	if !foundUB2 || !foundUB3 {
		t.Errorf("draw package configs should contain UB=2/Hilda=2 and UB=3/Hilda=1, got %+v", drawPkg.Configs)
	}
}

func TestComputeDeckVariants_NCapsPackages(t *testing.T) {
	// Use the independent-packages data set which produces 2 packages.
	decklists := map[string]types.Decklist{
		"d1": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Ultra Ball", SetCode: "T", Number: "30"}:  2,
			{Name: "Pidgeot ex", SetCode: "T", Number: "20"}:  1,
			{Name: "Pidgey", SetCode: "T", Number: "21"}:      2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      52,
		}},
		"d2": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Ultra Ball", SetCode: "T", Number: "30"}:  2,
			{Name: "Noctowl", SetCode: "T", Number: "10"}:     1,
			{Name: "Hoothoot", SetCode: "T", Number: "11"}:    2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      52,
		}},
		"d3": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Ultra Ball", SetCode: "T", Number: "30"}:  3,
			{Name: "Pidgeot ex", SetCode: "T", Number: "20"}:  1,
			{Name: "Pidgey", SetCode: "T", Number: "21"}:      2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      51,
		}},
		"d4": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Ultra Ball", SetCode: "T", Number: "30"}:  3,
			{Name: "Noctowl", SetCode: "T", Number: "10"}:     1,
			{Name: "Hoothoot", SetCode: "T", Number: "11"}:    2,
			{Name: "Filler", SetCode: "T", Number: "99"}:      51,
		}},
	}
	archetypes := variantTestArchetypes()

	result, err := computeDeckVariantsFromData(decklists, archetypes, "Charizard", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Packages) != 1 {
		t.Errorf("len(Packages) = %d, want 1 (capped by n)", len(result.Packages))
	}
}

func TestComputeDeckVariants_NoDecksOfArchetype(t *testing.T) {
	decklists := buildVariantTestData()
	archetypes := variantTestArchetypes()

	result, err := computeDeckVariantsFromData(decklists, archetypes, "Gardevoir", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalDecks != 0 {
		t.Errorf("TotalDecks = %d, want 0", result.TotalDecks)
	}
	if len(result.Packages) != 0 {
		t.Errorf("len(Packages) = %d, want 0", len(result.Packages))
	}
}

func TestComputeDeckVariants_SingleDeck(t *testing.T) {
	decklists := map[string]types.Decklist{
		"solo": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Filler", SetCode: "T", Number: "99"}:      57,
		}},
	}
	archetypes := variantTestArchetypes()

	result, err := computeDeckVariantsFromData(decklists, archetypes, "Charizard", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalDecks != 1 {
		t.Errorf("TotalDecks = %d, want 1", result.TotalDecks)
	}

	// All cards are core since there's only one deck.
	if result.CoreCards["Charizard ex"] != 3 {
		t.Errorf("CoreCards[Charizard ex] = %d, want 3", result.CoreCards["Charizard ex"])
	}
	if result.CoreCards["Filler"] != 57 {
		t.Errorf("CoreCards[Filler] = %d, want 57", result.CoreCards["Filler"])
	}

	// No flex cards means no packages.
	if len(result.Packages) != 0 {
		t.Errorf("len(Packages) = %d, want 0 (no flex cards)", len(result.Packages))
	}
}

func TestComputeDeckVariants_VaryingCounts(t *testing.T) {
	// Two decks that share a card name but at different counts.
	decklists := map[string]types.Decklist{
		"p1": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Rare Candy", SetCode: "T", Number: "5"}:   4,
			{Name: "Filler", SetCode: "T", Number: "99"}:      53,
		}},
		"p2": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 3,
			{Name: "Rare Candy", SetCode: "T", Number: "5"}:   3,
			{Name: "Filler", SetCode: "T", Number: "99"}:      54,
		}},
	}
	archetypes := variantTestArchetypes()

	result, err := computeDeckVariantsFromData(decklists, archetypes, "Charizard", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Charizard ex=3 is core; Rare Candy and Filler vary.
	if result.CoreCards["Charizard ex"] != 3 {
		t.Errorf("CoreCards[Charizard ex] = %d, want 3", result.CoreCards["Charizard ex"])
	}
	if _, ok := result.CoreCards["Rare Candy"]; ok {
		t.Error("Rare Candy should not be core (varies between 3 and 4)")
	}
	if _, ok := result.CoreCards["Filler"]; ok {
		t.Error("Filler should not be core (varies between 53 and 54)")
	}

	// Rare Candy and Filler have the same partition fingerprint (both split
	// decks {p1} vs {p2}), so they form a single package with 2 configs.
	if len(result.Packages) != 1 {
		t.Fatalf("len(Packages) = %d, want 1", len(result.Packages))
	}

	pkg := result.Packages[0]
	if len(pkg.Configs) != 2 {
		t.Fatalf("len(Configs) = %d, want 2", len(pkg.Configs))
	}
	// Each config appears in 1 deck.
	for _, cfg := range pkg.Configs {
		if cfg.Count != 1 {
			t.Errorf("config count = %d, want 1", cfg.Count)
		}
	}
}
