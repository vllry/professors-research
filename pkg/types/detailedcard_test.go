package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDetailedCardsFromJSON(t *testing.T) {
	// Test loading from a real JSON file
	jsonPath := filepath.Join("..", "..", "data", "cards", "sv01.json")
	file, err := os.Open(jsonPath)
	if err != nil {
		t.Fatalf("Failed to open JSON file: %v", err)
	}
	defer file.Close()

	cards, err := NewDetailedCardsFromJSON(file)
	if err != nil {
		t.Fatalf("Failed to load cards: %v", err)
	}

	if len(cards) == 0 {
		t.Fatal("Expected to load at least one card")
	}

	// Check first card
	firstCard := cards[0]
	if firstCard.SetCode() == "" {
		t.Error("Expected SetCode to be non-empty")
	}
	if firstCard.Name() == "" {
		t.Error("Expected Name to be non-empty")
	}
	if firstCard.Number() == "" {
		t.Error("Expected Number to be non-empty")
	}

	// Check if we have Pokemon cards with stages
	foundPokemon := false
	for _, card := range cards {
		if pokemonCard, ok := card.(*PokemonCard); ok {
			foundPokemon = true
			if pokemonCard.Stage == "" {
				t.Error("Expected Pokemon card to have a stage")
			}
			// Verify stage is one of the expected values
			stage := pokemonCard.Stage
			if stage != StageBasic && stage != StageStage1 && stage != StageStage2 {
				t.Errorf("Unexpected stage value: %s", stage)
			}
		}
	}

	if !foundPokemon {
		t.Error("Expected to find at least one Pokemon card")
	}

	t.Logf("Successfully loaded %d cards", len(cards))
	t.Logf("First card: %s %s %s", firstCard.Name(), firstCard.SetCode(), firstCard.Number())
	
	// Verify set code normalization: SV should be normalized to SVI
	if firstCard.SetCode() == "SVI" {
		t.Logf("Set code normalization working: SV -> SVI")
	} else if firstCard.SetCode() == "SV" {
		t.Errorf("Set code normalization failed: expected SVI, got SV")
	}
}

