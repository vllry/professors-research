package apiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/vllry/professors-research/internal/detailedcardcache"
)

func TestHandleStartOdds_UnidentifiedCardWarning(t *testing.T) {
	// Create a server with an empty cache (simulating cards not in cache)
	server, err := NewServer(Config{
		Port:    "8080",
		DataDir: "nonexistent", // Use a directory that doesn't exist to ensure empty cache
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Wait a bit for cache to "load" (it will fail to load, but IsLoaded will be true)
	// Actually, let's create a mock cache that's loaded but empty
	server.detailedCards = detailedcardcache.NewEmptyLoadedCache()

	decklist := `Pokémon: 1
1 Test Pokemon SVI 1

Energy: 59
57 Basic {R} Energy EVO 92
1 Basic Fire Energy SVI 1
1 Basic Water Energy SVI 1

Total Cards: 60`

	reqBody := StartOddsRequest{
		Decklist: decklist,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/start-odds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleStartOdds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response StartOddsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Basic {R} Energy should NOT generate a warning (it's a basic energy)
	// Basic Fire Energy should NOT generate a warning (it's a basic energy)
	// Test Pokemon should generate a warning (not in cache and not a basic energy)

	unidentifiedCount := 0
	basicEnergyWarnings := 0
	for _, err := range response.Errors {
		if err.Type == ErrorTypeUnidentifiedCard {
			unidentifiedCount++
			// Check that no basic energy cards are in the warnings
			if err.Info == "Card 'Basic {R} Energy EVO 92' could not be identified in the cache of known cards" ||
				err.Info == "Card 'Basic Fire Energy SVI 1' could not be identified in the cache of known cards" {
				basicEnergyWarnings++
			}
		}
	}

	if basicEnergyWarnings > 0 {
		t.Errorf("Basic energy cards should not generate warnings, but got %d warnings for basic energies", basicEnergyWarnings)
	}

	// Should have exactly 1 warning for the Test Pokemon
	if unidentifiedCount != 1 {
		t.Errorf("Expected 1 unidentified card warning, got %d. Errors: %v", unidentifiedCount, response.Errors)
	}
}

func TestHandlePrizeOdds_UnidentifiedCardWarning(t *testing.T) {
	// Create a server with an empty cache (simulating cards not in cache)
	server, err := NewServer(Config{
		Port:    "8080",
		DataDir: "nonexistent", // Use a directory that doesn't exist to ensure empty cache
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a mock cache that's loaded but empty
	server.detailedCards = detailedcardcache.NewEmptyLoadedCache()

	decklist := `Pokémon: 1
1 Test Pokemon SVI 1

Energy: 59
57 Basic {R} Energy EVO 92
1 Basic Fire Energy SVI 1
1 Basic Water Energy SVI 1

Total Cards: 60`

	reqBody := PrizeOddsRequest{
		Decklist: decklist,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/prize-odds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handlePrizeOdds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response PrizeOddsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Basic {R} Energy should NOT generate a warning (it's a basic energy)
	// Basic Fire Energy should NOT generate a warning (it's a basic energy)
	// Test Pokemon should generate a warning (not in cache and not a basic energy)

	unidentifiedCount := 0
	basicEnergyWarnings := 0
	for _, err := range response.Errors {
		if err.Type == ErrorTypeUnidentifiedCard {
			unidentifiedCount++
			// Check that no basic energy cards are in the warnings
			if err.Info == "Card 'Basic {R} Energy EVO 92' could not be identified in the cache of known cards" ||
				err.Info == "Card 'Basic Fire Energy SVI 1' could not be identified in the cache of known cards" {
				basicEnergyWarnings++
			}
		}
	}

	if basicEnergyWarnings > 0 {
		t.Errorf("Basic energy cards should not generate warnings, but got %d warnings for basic energies", basicEnergyWarnings)
	}

	// Should have exactly 1 warning for the Test Pokemon
	if unidentifiedCount != 1 {
		t.Errorf("Expected 1 unidentified card warning, got %d. Errors: %v", unidentifiedCount, response.Errors)
	}
}

func TestBasicEnergyPattern(t *testing.T) {
	basicEnergyPattern := regexp.MustCompile(`^Basic\s+(\{[A-Z]\}|\w+)\s+Energy$`)

	testCases := []struct {
		name     string
		expected bool
	}{
		{"Basic {R} Energy", true},
		{"Basic {P} Energy", true},
		{"Basic {W} Energy", true},
		{"Basic Fire Energy", true},
		{"Basic Water Energy", true},
		{"Basic Psychic Energy", true},
		{"Basic Lightning Energy", true},
		{"Basic {R} Energy EVO 92", false}, // Has extra text
		{"Basic Energy", false},            // Missing type
		{"Basic {R}", false},               // Missing "Energy"
		{"Fire Energy", false},              // Missing "Basic"
		{"Basic {RR} Energy", false},       // Multiple letters in braces
		{"Basic {R} Energy Extra", false},  // Extra text
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := basicEnergyPattern.MatchString(tc.name)
			if result != tc.expected {
				t.Errorf("Pattern match for '%s': expected %v, got %v", tc.name, tc.expected, result)
			}
		})
	}
}

