package apiserver

import (
	"bytes"
	"encoding/json"
	"math"
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

func TestHandleDrawSupporterOdds_Success(t *testing.T) {
	server, err := NewServer(Config{
		Port:    "8080",
		DataDir: "nonexistent",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	reqBody := DrawSupporterOddsRequest{
		DeckSize:    46,
		KnownBottom: 0,
		HandSize:    7,
		PrizeCards:  6,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/draw-supporter-odds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDrawSupporterOdds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response DrawSupporterOddsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify supporters exist and arrays are correct length.
	supporters := []string{"Iono", "Professor's Research", "Lillie's Determination"}
	for _, name := range supporters {
		row, ok := response.Odds[name]
		if !ok {
			t.Fatalf("missing supporter %q in response", name)
		}
		if len(row) != 4 {
			t.Fatalf("expected 4 odds values for %q, got %d", name, len(row))
		}

		if response.DrawCounts[name] == 0 {
			t.Fatalf("missing drawCounts for %q in response", name)
		}
		if response.EffectiveDrawCounts[name] == 0 {
			t.Fatalf("missing effectiveDrawCounts for %q in response", name)
		}
	}

	// Pair odds are defined for Iono/Research (top-of-deck pool) and Lillie's (shuffled pool).
	pairSupporters := []string{"Iono", "Professor's Research", "Lillie's Determination"}
	for _, name := range pairSupporters {
		pair, ok := response.PairOdds[name]
		if !ok {
			t.Fatalf("missing supporter %q in pairOdds response", name)
		}
		if len(pair) != 4 {
			t.Fatalf("expected 4 rows in pairOdds for %q, got %d", name, len(pair))
		}
		for i := range pair {
			if len(pair[i]) != 4 {
				t.Fatalf("expected 4 cols in pairOdds for %q row %d, got %d", name, i, len(pair[i]))
			}
		}
	}

	if response.DrawCounts["Iono"] != 6 {
		t.Fatalf("expected Iono draw count 6, got %d", response.DrawCounts["Iono"])
	}
	if response.EffectiveDrawCounts["Iono"] != 6 {
		t.Fatalf("expected Iono effective draw count 6, got %d", response.EffectiveDrawCounts["Iono"])
	}

	// Spot-check a few known values.
	// With 1 copy in pool, P(draw it) = draw/pool.
	const tol = 1e-12

	// Iono: pool=46, draw=6 => 6/46
	if math.Abs(response.Odds["Iono"][0]-(6.0/46.0)) > tol {
		t.Fatalf("unexpected Iono 1-copy odds: got %.15f", response.Odds["Iono"][0])
	}

	// Research: pool=46, draw=7 => 7/46
	if math.Abs(response.Odds["Professor's Research"][0]-(7.0/46.0)) > tol {
		t.Fatalf("unexpected Research 1-copy odds: got %.15f", response.Odds["Professor's Research"][0])
	}

	// Lillie's: pool=46+7-1=52, draw=8 (prizes==6) => 8/52
	if math.Abs(response.Odds["Lillie's Determination"][0]-(8.0/52.0)) > tol {
		t.Fatalf("unexpected Lillie's 1-copy odds: got %.15f", response.Odds["Lillie's Determination"][0])
	}

	// Pair spot-check: with 1 copy of each card, probability of drawing both specific cards is draw*(draw-1)/(pool*(pool-1)).
	ionoPool := 46
	ionoDraw := 6
	wantPair := float64(ionoDraw*(ionoDraw-1)) / float64(ionoPool*(ionoPool-1))
	if math.Abs(response.PairOdds["Iono"][0][0]-wantPair) > tol {
		t.Fatalf("unexpected Iono pair(1,1) odds: got %.15f", response.PairOdds["Iono"][0][0])
	}
}

func TestHandleDrawSupporterOdds_ValidationError(t *testing.T) {
	server, err := NewServer(Config{
		Port:    "8080",
		DataDir: "nonexistent",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	reqBody := DrawSupporterOddsRequest{
		DeckSize:    0,
		KnownBottom: 0,
		HandSize:    7,
		PrizeCards:  6,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/draw-supporter-odds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDrawSupporterOdds(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDrawSupporterOdds_BottomOdds(t *testing.T) {
	server, err := NewServer(Config{
		Port:    "8080",
		DataDir: "nonexistent",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	reqBody := DrawSupporterOddsRequest{
		DeckSize:    10,
		KnownBottom: 5,
		HandSize:    7,
		PrizeCards:  6,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/draw-supporter-odds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDrawSupporterOdds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response DrawSupporterOddsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.BottomOdds == nil {
		t.Fatalf("Expected bottomOdds to be present")
	}
	ionoBottom := response.BottomOdds["Iono"]
	if len(ionoBottom) != 4 {
		t.Fatalf("Expected 4 bottom odds for Iono, got %d", len(ionoBottom))
	}
	if response.BottomDrawCounts["Iono"] != 1 {
		t.Fatalf("Expected bottom draw count 1 for Iono, got %d", response.BottomDrawCounts["Iono"])
	}
	if _, ok := response.BottomOdds["Lillie's Determination"]; ok {
		t.Fatalf("Did not expect bottomOdds for Lillie's Determination")
	}
}

func TestHandleDrawSupporterOdds_NoBottomOddsWhenNotDrawingIntoBottom(t *testing.T) {
	server, err := NewServer(Config{
		Port:    "8080",
		DataDir: "nonexistent",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	reqBody := DrawSupporterOddsRequest{
		DeckSize:    46,
		KnownBottom: 5,
		HandSize:    7,
		PrizeCards:  1, // Iono draw=1, Research draw=7, top pool=41 => no bottom draw
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/draw-supporter-odds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDrawSupporterOdds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response DrawSupporterOddsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.BottomOdds != nil {
		t.Fatalf("Expected bottomOdds to be omitted when not drawing into bottom")
	}
	if response.BottomDrawCounts != nil {
		t.Fatalf("Expected bottomDrawCounts to be omitted when not drawing into bottom")
	}
}

func TestHandleDrawSupporterOdds_EffectiveDrawCountsClampToTopPool(t *testing.T) {
	server, err := NewServer(Config{
		Port:    "8080",
		DataDir: "nonexistent",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// deck=10, bottom=5 => top pool=5. Iono draws 6 (prizes=6) but effective should clamp to 5.
	reqBody := DrawSupporterOddsRequest{
		DeckSize:    10,
		KnownBottom: 5,
		HandSize:    7,
		PrizeCards:  6,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/draw-supporter-odds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDrawSupporterOdds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response DrawSupporterOddsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.DrawCounts["Iono"] != 6 {
		t.Fatalf("Expected Iono requested draw count 6, got %d", response.DrawCounts["Iono"])
	}
	if response.EffectiveDrawCounts["Iono"] != 5 {
		t.Fatalf("Expected Iono effective draw count 5, got %d", response.EffectiveDrawCounts["Iono"])
	}

	// With pool=5 and effective draw=5, odds of at least 1 copy is 1.0 for any targetCount>=1.
	if got := response.Odds["Iono"][0]; got != 1.0 {
		t.Fatalf("Expected Iono 1-copy odds 1.0 when drawing all top cards, got %.15f", got)
	}
}

func TestHandleDrawSupporterOdds_LilliesNotAffectedByKnownBottom(t *testing.T) {
	server, err := NewServer(Config{
		Port:    "8080",
		DataDir: "nonexistent",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	reqBody := DrawSupporterOddsRequest{
		DeckSize:    10,
		KnownBottom: 9, // extreme, but valid (<= deck-1)
		HandSize:    7,
		PrizeCards:  6, // Lillie's draw=8
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/draw-supporter-odds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDrawSupporterOdds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response DrawSupporterOddsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Lillie's uses pool = deck+hand-1 = 16, and draw=8 when prizes==6.
	if response.DrawCounts["Lillie's Determination"] != 8 {
		t.Fatalf("Expected Lillie's requested draw count 8, got %d", response.DrawCounts["Lillie's Determination"])
	}
	if response.EffectiveDrawCounts["Lillie's Determination"] != 8 {
		t.Fatalf("Expected Lillie's effective draw count 8, got %d", response.EffectiveDrawCounts["Lillie's Determination"])
	}
	// No bottom odds for Lillie's even though top pool is tiny.
	if response.BottomOdds != nil {
		if _, ok := response.BottomOdds["Lillie's Determination"]; ok {
			t.Fatalf("Did not expect bottomOdds for Lillie's Determination")
		}
	}
}

