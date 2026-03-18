package matchups

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/vllry/professors-research/pkg/types"
)

func TestNewMatchupRecord_WinRate(t *testing.T) {
	tests := []struct {
		name           string
		wins, losses   int
		ties           int
		wantRate       float64
	}{
		{"all wins", 10, 0, 0, 1.0},
		{"all losses", 0, 10, 0, 0.0},
		{"even", 5, 5, 0, 0.5},
		{"with ties", 10, 5, 3, (10.0 + 1.0) / 18.0},
		{"only ties", 0, 0, 3, 1.0 / 3.0},
		{"zero games", 0, 0, 0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newMatchupRecord(tt.wins, tt.losses, tt.ties)
			if rec.Wins != tt.wins || rec.Losses != tt.losses || rec.Ties != tt.ties {
				t.Errorf("record fields: got %d/%d/%d, want %d/%d/%d",
					rec.Wins, rec.Losses, rec.Ties, tt.wins, tt.losses, tt.ties)
			}
			if math.Abs(rec.WinRate-tt.wantRate) > 1e-9 {
				t.Errorf("WinRate = %f, want %f", rec.WinRate, tt.wantRate)
			}
		})
	}
}

func buildTestData() (map[string]types.Decklist, []types.MatchResult) {
	decklists := map[string]types.Decklist{
		"alice": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Arcanine ex", SetCode: "T", Number: "2"}:  2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       56,
		}},
		"bob": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       58,
		}},
		"carol": {Cards: map[types.Card]int{
			{Name: "Dragapult ex", SetCode: "T", Number: "4"}: 3,
			{Name: "Filler", SetCode: "T", Number: "3"}:       57,
		}},
		"dave": {Cards: map[types.Card]int{
			{Name: "Gardevoir ex", SetCode: "T", Number: "5"}: 2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       58,
		}},
	}

	matches := []types.MatchResult{
		// Alice (Charizard Arcanine variant 0) vs Carol (Dragapult) -- alice wins
		{Round: 1, Player1: "alice", Player2: "carol", Outcome: types.MatchOutcomeWin, Winner: "alice"},
		// Alice vs Dave (Gardevoir) -- alice loses
		{Round: 2, Player1: "alice", Player2: "dave", Outcome: types.MatchOutcomeWin, Winner: "dave"},
		// Alice vs Carol -- tie
		{Round: 3, Player1: "alice", Player2: "carol", Outcome: types.MatchOutcomeTie, Winner: ""},
		// Bob (Charizard variant other) vs Carol (Dragapult) -- bob wins
		{Round: 1, Player1: "bob", Player2: "carol", Outcome: types.MatchOutcomeWin, Winner: "bob"},
		// Bob vs Dave -- bob wins
		{Round: 2, Player1: "bob", Player2: "dave", Outcome: types.MatchOutcomeWin, Winner: "bob"},
		// Carol vs Dave -- carol wins (not involving target archetype)
		{Round: 3, Player1: "carol", Player2: "dave", Outcome: types.MatchOutcomeWin, Winner: "carol"},
	}

	return decklists, matches
}

func TestComputeMatchupsFromData(t *testing.T) {
	archetypes := []Archetype{
		{Name: "Charizard Arcanine", Requires: map[string]int{"Charizard ex": 1, "Arcanine ex": 1}},
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
		{Name: "Dragapult", Requires: map[string]int{"Dragapult ex": 1}},
		{Name: "Gardevoir", Requires: map[string]int{"Gardevoir ex": 1}},
	}

	variants := []map[string]int{
		{"Arcanine ex": 2},
	}

	decklists, matches := buildTestData()
	result, err := computeMatchupsFromData(decklists, matches, archetypes, "Charizard Arcanine", variants, nil, nil)
	if err != nil {
		t.Fatalf("computeMatchupsFromData() error: %v", err)
	}

	// Archetype counts
	if result.ArchetypeCounts["Charizard Arcanine"] != 1 {
		t.Errorf("expected 1 Charizard Arcanine deck, got %d", result.ArchetypeCounts["Charizard Arcanine"])
	}
	if result.ArchetypeCounts["Charizard"] != 1 {
		t.Errorf("expected 1 Charizard deck, got %d", result.ArchetypeCounts["Charizard"])
	}

	// Variant counts: alice matches variant 0, bob is Charizard (not target archetype)
	if result.VariantCounts["0"] != 1 {
		t.Errorf("expected 1 deck in variant 0, got %d", result.VariantCounts["0"])
	}

	// Variant card count map should be returned with each result.
	if got := result.Matchups["0"].CardCounts["Arcanine ex"]; got != 2 {
		t.Errorf("variant 0 cardCounts[Arcanine ex] = %d, want 2", got)
	}

	// Alice (variant 0) vs Dragapult: 1 win, 0 losses, 1 tie
	rec := result.Matchups["0"].Matchups["Dragapult"]
	if rec.Wins != 1 || rec.Losses != 0 || rec.Ties != 1 {
		t.Errorf("variant 0 vs Dragapult: got %d/%d/%d, want 1/0/1", rec.Wins, rec.Losses, rec.Ties)
	}
	// (1 + 1/3) / 2 = 2/3
	expectedRate := (1.0 + 1.0/3.0) / 2.0
	if math.Abs(rec.WinRate-expectedRate) > 1e-9 {
		t.Errorf("variant 0 vs Dragapult win rate = %f, want %f", rec.WinRate, expectedRate)
	}

	// Alice (variant 0) vs Gardevoir: 0 wins, 1 loss, 0 ties
	rec2 := result.Matchups["0"].Matchups["Gardevoir"]
	if rec2.Wins != 0 || rec2.Losses != 1 {
		t.Errorf("variant 0 vs Gardevoir: got %d/%d/%d, want 0/1/0", rec2.Wins, rec2.Losses, rec2.Ties)
	}
}

func TestComputeMatchupsFromData_BothPlayersTargetArchetype(t *testing.T) {
	archetypes := []Archetype{
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
	}

	decklists := map[string]types.Decklist{
		"alice": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Arcanine ex", SetCode: "T", Number: "2"}:  2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       56,
		}},
		"bob": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       58,
		}},
	}

	matches := []types.MatchResult{
		{Round: 1, Player1: "alice", Player2: "bob", Outcome: types.MatchOutcomeWin, Winner: "alice"},
	}

	variants := []map[string]int{
		{"Arcanine ex": 1},
	}

	result, err := computeMatchupsFromData(decklists, matches, archetypes, "Charizard", variants, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Alice is variant 0, Bob is variant "other".
	// Alice beat Bob: variant 0 vs Charizard => 1 win; other vs Charizard => 1 loss.
	rec := result.Matchups["0"].Matchups["Charizard"]
	if rec.Wins != 1 || rec.Losses != 0 {
		t.Errorf("variant 0 vs Charizard (mirror): got %d/%d/%d, want 1/0/0", rec.Wins, rec.Losses, rec.Ties)
	}

	rec2 := result.Matchups["other"].Matchups["Charizard"]
	if rec2.Wins != 0 || rec2.Losses != 1 {
		t.Errorf("other vs Charizard (mirror): got %d/%d/%d, want 0/1/0", rec2.Wins, rec2.Losses, rec2.Ties)
	}
}

func TestComputeMatchupsFromData_MissingDecklist(t *testing.T) {
	archetypes := []Archetype{
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
	}

	decklists := map[string]types.Decklist{
		"alice": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       58,
		}},
	}

	matches := []types.MatchResult{
		{Round: 1, Player1: "alice", Player2: "unknown_player", Outcome: types.MatchOutcomeWin, Winner: "alice"},
	}

	result, err := computeMatchupsFromData(decklists, matches, archetypes, "Charizard", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Match should be skipped because unknown_player has no decklist.
	if len(result.Matchups) != 0 {
		t.Errorf("expected no matchup data when opponent has no decklist, got %v", result.Matchups)
	}
}

func TestComputeMatchupsForTournaments_NamespacesPlayersAndAggregates(t *testing.T) {
	archetypes := []Archetype{
		{Name: "Charizard Arcanine", Requires: map[string]int{"Charizard ex": 1, "Arcanine ex": 1}},
		{Name: "Dragapult", Requires: map[string]int{"Dragapult ex": 1}},
	}
	variants := []map[string]int{
		{"Arcanine ex": 1},
	}

	writeTournament := func(root, id string) string {
		t.Helper()
		dir := filepath.Join(root, id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}

		decklists := []map[string]any{
			{
				"player": "same-name",
				"cards": []map[string]any{
					{"count": 2, "card": map[string]any{"name": "Charizard ex", "setCode": "T", "number": "1"}},
					{"count": 1, "card": map[string]any{"name": "Arcanine ex", "setCode": "T", "number": "2"}},
					{"count": 57, "card": map[string]any{"name": "Filler", "setCode": "T", "number": "3"}},
				},
			},
			{
				"player": "opponent",
				"cards": []map[string]any{
					{"count": 2, "card": map[string]any{"name": "Dragapult ex", "setCode": "T", "number": "4"}},
					{"count": 58, "card": map[string]any{"name": "Filler", "setCode": "T", "number": "3"}},
				},
			},
		}
		dlBytes, _ := json.Marshal(decklists)
		if err := os.WriteFile(filepath.Join(dir, "decklists.json"), dlBytes, 0o644); err != nil {
			t.Fatalf("write decklists.json: %v", err)
		}

		matches := []types.MatchResult{
			{Round: 1, Table: 1, Player1: "same-name", Player2: "opponent", Outcome: types.MatchOutcomeWin, Winner: "same-name"},
		}
		mBytes, _ := json.Marshal(matches)
		if err := os.WriteFile(filepath.Join(dir, "matches.json"), mBytes, 0o644); err != nil {
			t.Fatalf("write matches.json: %v", err)
		}

		return dir
	}

	root := t.TempDir()
	dir1 := writeTournament(root, "t1")
	dir2 := writeTournament(root, "t2")

	result, err := ComputeMatchupsForTournaments(
		[]string{"t1", "t2"},
		[]string{dir1, dir2},
		archetypes,
		"Charizard Arcanine",
		variants,
		PlacementFilter{},
	)
	if err != nil {
		t.Fatalf("ComputeMatchupsForTournaments error: %v", err)
	}

	// Two target decks across the two tournaments.
	if result.ArchetypeCounts["Charizard Arcanine"] != 2 {
		t.Fatalf("archetypeCounts[Charizard Arcanine] = %d, want 2", result.ArchetypeCounts["Charizard Arcanine"])
	}
	if result.VariantCounts["0"] != 2 {
		t.Fatalf("variantCounts[0] = %d, want 2", result.VariantCounts["0"])
	}

	// Aggregated matchup: two wins vs Dragapult.
	rec := result.Matchups["0"].Matchups["Dragapult"]
	if rec.Wins != 2 || rec.Losses != 0 || rec.Ties != 0 {
		t.Fatalf("variant 0 vs Dragapult got %d/%d/%d, want 2/0/0", rec.Wins, rec.Losses, rec.Ties)
	}

	// Card count map is returned for the variant.
	if result.Matchups["0"].CardCounts["Arcanine ex"] != 1 {
		t.Fatalf("variant 0 cardCounts[Arcanine ex] = %d, want 1", result.Matchups["0"].CardCounts["Arcanine ex"])
	}
}

func TestComputeMatchPoints(t *testing.T) {
	matches := []types.MatchResult{
		{Player1: "a", Player2: "b", Outcome: types.MatchOutcomeWin, Winner: "a"},
		{Player1: "a", Player2: "c", Outcome: types.MatchOutcomeWin, Winner: "c"},
		{Player1: "b", Player2: "c", Outcome: types.MatchOutcomeTie},
	}
	pts := computeMatchPoints(matches)

	// a: 1 win + 1 loss = 3
	if pts["a"] != 3 {
		t.Errorf("a = %d, want 3", pts["a"])
	}
	// b: 1 loss + 1 tie = 1
	if pts["b"] != 1 {
		t.Errorf("b = %d, want 1", pts["b"])
	}
	// c: 1 win + 1 tie = 4
	if pts["c"] != 4 {
		t.Errorf("c = %d, want 4", pts["c"])
	}
}

func TestBuildPercentileSet(t *testing.T) {
	pts := map[string]int{
		"p1": 12, // best
		"p2": 9,
		"p3": 6,
		"p4": 3, // worst
	}

	// No filter for 0 or 100.
	if s := buildPercentileSet(pts, 0); s != nil {
		t.Error("percentile 0 should return nil")
	}
	if s := buildPercentileSet(pts, 100); s != nil {
		t.Error("percentile 100 should return nil")
	}

	// Top 25% of 4 players => top 1 player.
	top25 := buildPercentileSet(pts, 25)
	if top25 == nil {
		t.Fatal("top 25% should not be nil")
	}
	if !top25["p1"] {
		t.Error("p1 should be in top 25%")
	}
	if top25["p2"] || top25["p3"] || top25["p4"] {
		t.Error("only p1 should be in top 25%")
	}

	// Top 50% of 4 players => top 2 players.
	top50 := buildPercentileSet(pts, 50)
	if !top50["p1"] || !top50["p2"] {
		t.Error("p1, p2 should be in top 50%")
	}
	if top50["p3"] || top50["p4"] {
		t.Error("p3, p4 should not be in top 50%")
	}
}

func TestBuildPercentileSet_Ties(t *testing.T) {
	// All players tied -- everyone should be included for any percentile.
	pts := map[string]int{"a": 6, "b": 6, "c": 6, "d": 6}
	top25 := buildPercentileSet(pts, 25)
	if top25 == nil {
		t.Fatal("should not be nil")
	}
	for _, p := range []string{"a", "b", "c", "d"} {
		if !top25[p] {
			t.Errorf("%s should be included when all tied", p)
		}
	}
}

func TestPlacementFilter_PlayerOnly(t *testing.T) {
	// 4 players, 6 matches. Target archetype is Charizard Arcanine (alice only).
	// Official standings: bob=1, alice=2, carol=3, dave=4.
	// Top 25% = top 1 player = bob. alice is not in top 25%.
	// With playerPlacement=25 on Charizard Arcanine, alice is filtered out => no matchup data.
	archetypes := []Archetype{
		{Name: "Charizard Arcanine", Requires: map[string]int{"Charizard ex": 1, "Arcanine ex": 1}},
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
		{Name: "Dragapult", Requires: map[string]int{"Dragapult ex": 1}},
		{Name: "Gardevoir", Requires: map[string]int{"Gardevoir ex": 1}},
	}

	decklists, matches := buildTestData()
	standings := map[string]int{"bob": 1, "alice": 2, "carol": 3, "dave": 4}
	pf := buildStandingsPercentileSet(standings, 25)

	result, err := computeMatchupsFromData(decklists, matches, archetypes, "Charizard Arcanine", nil, pf, nil)
	if err != nil {
		t.Fatal(err)
	}

	// alice is standing 2, top 25% includes standing 1 only (bob), so alice is excluded.
	if len(result.Matchups) != 0 {
		t.Errorf("expected no matchups with strict player filter, got %v", result.Matchups)
	}
}

func TestPlacementFilter_OpponentOnly(t *testing.T) {
	// Target archetype is Charizard Arcanine (alice). opponentPlacement=25 => top 25%.
	// Official standings: bob=1, alice=2, carol=3, dave=4.
	// Top 25% = bob. Alice's opponents are carol and dave; neither is top 25%.
	archetypes := []Archetype{
		{Name: "Charizard Arcanine", Requires: map[string]int{"Charizard ex": 1, "Arcanine ex": 1}},
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
		{Name: "Dragapult", Requires: map[string]int{"Dragapult ex": 1}},
		{Name: "Gardevoir", Requires: map[string]int{"Gardevoir ex": 1}},
	}

	decklists, matches := buildTestData()
	standings := map[string]int{"bob": 1, "alice": 2, "carol": 3, "dave": 4}
	of := buildStandingsPercentileSet(standings, 25)

	result, err := computeMatchupsFromData(decklists, matches, archetypes, "Charizard Arcanine", nil, nil, of)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Matchups) != 0 {
		t.Errorf("expected no matchups when opponents don't pass filter, got %v", result.Matchups)
	}
}

func TestPlacementFilter_MirrorMatch(t *testing.T) {
	// Two Charizard players: alice (strong) and bob (weak).
	// alice beats bob. Match points: alice=3, bob=0.
	// Official standings: alice=1, bob=2. playerPlacement=50 => top 1 = alice.
	// opponentPlacement=0 => no filter.
	// Only alice's perspective should be recorded.
	archetypes := []Archetype{
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
	}

	decklists := map[string]types.Decklist{
		"alice": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Arcanine ex", SetCode: "T", Number: "2"}:  2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       56,
		}},
		"bob": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       58,
		}},
	}

	matches := []types.MatchResult{
		{Round: 1, Player1: "alice", Player2: "bob", Outcome: types.MatchOutcomeWin, Winner: "alice"},
	}

	variants := []map[string]int{
		{"Arcanine ex": 1},
	}

	standings := map[string]int{"alice": 1, "bob": 2}
	pf := buildStandingsPercentileSet(standings, 50)

	result, err := computeMatchupsFromData(decklists, matches, archetypes, "Charizard", variants, pf, nil)
	if err != nil {
		t.Fatal(err)
	}

	// alice is variant 0, in top 50% => her perspective is recorded: 1 win vs Charizard.
	rec := result.Matchups["0"].Matchups["Charizard"]
	if rec.Wins != 1 || rec.Losses != 0 {
		t.Errorf("variant 0 vs Charizard: got %d/%d/%d, want 1/0/0", rec.Wins, rec.Losses, rec.Ties)
	}

	// bob is variant "other", NOT in top 50% => his perspective is filtered out.
	if _, ok := result.Matchups["other"]; ok {
		recOther := result.Matchups["other"].Matchups["Charizard"]
		if recOther.Wins != 0 || recOther.Losses != 0 || recOther.Ties != 0 {
			t.Errorf("other vs Charizard should be empty, got %d/%d/%d", recOther.Wins, recOther.Losses, recOther.Ties)
		}
	}
}

func TestPlacementFilter_MirrorBothFiltered(t *testing.T) {
	// Both players are Charizard. alice beats bob.
	// playerPlacement=50 (alice only), opponentPlacement=50 (alice only).
	// alice's perspective: alice passes player filter, bob fails opponent filter => excluded.
	// bob's perspective: bob fails player filter => excluded.
	// Result: no matchups recorded.
	archetypes := []Archetype{
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
	}

	decklists := map[string]types.Decklist{
		"alice": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       58,
		}},
		"bob": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "Filler", SetCode: "T", Number: "3"}:       58,
		}},
	}

	matches := []types.MatchResult{
		{Round: 1, Player1: "alice", Player2: "bob", Outcome: types.MatchOutcomeWin, Winner: "alice"},
	}

	standings := map[string]int{"alice": 1, "bob": 2}
	pf := buildStandingsPercentileSet(standings, 50)
	of := buildStandingsPercentileSet(standings, 50)

	result, err := computeMatchupsFromData(decklists, matches, archetypes, "Charizard", nil, pf, of)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Matchups) != 0 {
		t.Errorf("expected no matchups when opponent fails filter in mirror, got %v", result.Matchups)
	}
}

func TestComputeMatchupsForTournaments_WithPlacementFilter(t *testing.T) {
	archetypes := []Archetype{
		{Name: "Charizard Arcanine", Requires: map[string]int{"Charizard ex": 1, "Arcanine ex": 1}},
		{Name: "Dragapult", Requires: map[string]int{"Dragapult ex": 1}},
	}
	variants := []map[string]int{
		{"Arcanine ex": 1},
	}

	writeTournament := func(root, id string, winnerIsTarget bool) string {
		t.Helper()
		dir := filepath.Join(root, id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}

		decklists := []map[string]any{
			{
				"player": "target-player",
				"cards": []map[string]any{
					{"count": 2, "card": map[string]any{"name": "Charizard ex", "setCode": "T", "number": "1"}},
					{"count": 1, "card": map[string]any{"name": "Arcanine ex", "setCode": "T", "number": "2"}},
					{"count": 57, "card": map[string]any{"name": "Filler", "setCode": "T", "number": "3"}},
				},
			},
			{
				"player": "opp",
				"cards": []map[string]any{
					{"count": 2, "card": map[string]any{"name": "Dragapult ex", "setCode": "T", "number": "4"}},
					{"count": 58, "card": map[string]any{"name": "Filler", "setCode": "T", "number": "3"}},
				},
			},
		}
		dlBytes, _ := json.Marshal(decklists)
		if err := os.WriteFile(filepath.Join(dir, "decklists.json"), dlBytes, 0o644); err != nil {
			t.Fatalf("write decklists.json: %v", err)
		}

		winner := "target-player"
		if !winnerIsTarget {
			winner = "opp"
		}
		matches := []types.MatchResult{
			{Round: 1, Table: 1, Player1: "target-player", Player2: "opp", Outcome: types.MatchOutcomeWin, Winner: winner},
		}
		mBytes, _ := json.Marshal(matches)
		if err := os.WriteFile(filepath.Join(dir, "matches.json"), mBytes, 0o644); err != nil {
			t.Fatalf("write matches.json: %v", err)
		}

		// Roster standings: the winner should be standing 1 (top 50% of 2 players),
		// the loser standing 2.
		targetStanding := 1
		oppStanding := 2
		if !winnerIsTarget {
			targetStanding = 2
			oppStanding = 1
		}
		rosterHTML := fmt.Sprintf(`<!doctype html>
<html>
  <body>
    <table id="dtLiveRoster">
      <thead>
        <tr>
          <th>Player ID</th>
          <th>First name</th>
          <th>Last name</th>
          <th>Country</th>
          <th class="text-center">Deck List</th>
          <th class="text-center">Standing</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>p1</td><td>target</td><td>player</td><td>XX</td><td></td><td class="text-center">%d</td>
        </tr>
        <tr>
          <td>p2</td><td>opp</td><td>player</td><td>XX</td><td></td><td class="text-center">%d</td>
        </tr>
      </tbody>
    </table>
  </body>
</html>`, targetStanding, oppStanding)
		if err := os.WriteFile(filepath.Join(dir, "roster.html"), []byte(rosterHTML), 0o644); err != nil {
			t.Fatalf("write roster.html: %v", err)
		}

		// The standings loader keys players as "<Player ID> <First> <Last> [<Country>]".
		// Ensure our decklist/match identifiers match that.
		//
		// We wrote Player ID/First/Last/Country above such that the produced keys are:
		// "p1 target player [XX]" and "p2 opp player [XX]".
		// So we must also rewrite the decklists/matches identifiers to those same strings.
		//
		// (This test is specifically for the tournament-loading path.)
		dlBytes2, _ := json.Marshal([]map[string]any{
			{
				"player": "p1 target player [XX]",
				"cards": []map[string]any{
					{"count": 2, "card": map[string]any{"name": "Charizard ex", "setCode": "T", "number": "1"}},
					{"count": 1, "card": map[string]any{"name": "Arcanine ex", "setCode": "T", "number": "2"}},
					{"count": 57, "card": map[string]any{"name": "Filler", "setCode": "T", "number": "3"}},
				},
			},
			{
				"player": "p2 opp player [XX]",
				"cards": []map[string]any{
					{"count": 2, "card": map[string]any{"name": "Dragapult ex", "setCode": "T", "number": "4"}},
					{"count": 58, "card": map[string]any{"name": "Filler", "setCode": "T", "number": "3"}},
				},
			},
		})
		if err := os.WriteFile(filepath.Join(dir, "decklists.json"), dlBytes2, 0o644); err != nil {
			t.Fatalf("rewrite decklists.json: %v", err)
		}

		winnerKey := "p1 target player [XX]"
		if !winnerIsTarget {
			winnerKey = "p2 opp player [XX]"
		}
		matches2 := []types.MatchResult{
			{Round: 1, Table: 1, Player1: "p1 target player [XX]", Player2: "p2 opp player [XX]", Outcome: types.MatchOutcomeWin, Winner: winnerKey},
		}
		mBytes2, _ := json.Marshal(matches2)
		if err := os.WriteFile(filepath.Join(dir, "matches.json"), mBytes2, 0o644); err != nil {
			t.Fatalf("rewrite matches.json: %v", err)
		}
		return dir
	}

	root := t.TempDir()
	// t1: target-player wins (3 pts) => in top 50%.
	dir1 := writeTournament(root, "t1", true)
	// t2: target-player loses (0 pts) => NOT in top 50%.
	dir2 := writeTournament(root, "t2", false)

	result, err := ComputeMatchupsForTournaments(
		[]string{"t1", "t2"},
		[]string{dir1, dir2},
		archetypes,
		"Charizard Arcanine",
		variants,
		PlacementFilter{PlayerPercentile: 50},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Only the t1 match should be counted (target-player wins with 3 pts, top 50%).
	// t2's target-player has 0 pts (bottom 50%), filtered out.
	rec := result.Matchups["0"].Matchups["Dragapult"]
	if rec.Wins != 1 || rec.Losses != 0 {
		t.Errorf("expected 1/0/0 (only t1 counted), got %d/%d/%d", rec.Wins, rec.Losses, rec.Ties)
	}
}

func TestComputeMatchupsFromData_InclusiveVariants(t *testing.T) {
	// Test that a deck matching multiple variants is counted in all of them.
	archetypes := []Archetype{
		{Name: "Charizard", Requires: map[string]int{"Charizard ex": 1}},
		{Name: "Dragapult", Requires: map[string]int{"Dragapult ex": 1}},
	}

	// Variants: CardX and CardY. Alice's deck has both.
	variants := []map[string]int{
		{"CardX": 1},
		{"CardY": 1},
	}

	decklists := map[string]types.Decklist{
		"alice": {Cards: map[types.Card]int{
			{Name: "Charizard ex", SetCode: "T", Number: "1"}: 2,
			{Name: "CardX", SetCode: "T", Number: "2"}:        1,
			{Name: "CardY", SetCode: "T", Number: "3"}:        1,
			{Name: "Filler", SetCode: "T", Number: "99"}:      56,
		}},
		"bob": {Cards: map[types.Card]int{
			{Name: "Dragapult ex", SetCode: "T", Number: "4"}: 3,
			{Name: "Filler", SetCode: "T", Number: "99"}:      57,
		}},
	}

	matches := []types.MatchResult{
		{Round: 1, Player1: "alice", Player2: "bob", Outcome: types.MatchOutcomeWin, Winner: "alice"},
	}

	result, err := computeMatchupsFromData(decklists, matches, archetypes, "Charizard", variants, nil, nil)
	if err != nil {
		t.Fatalf("computeMatchupsFromData() error: %v", err)
	}

	// Alice matches both variant 0 (CardX) and variant 1 (CardY).
	// Variant counts should reflect this: alice is counted in both.
	if result.VariantCounts["0"] != 1 {
		t.Errorf("expected 1 deck in variant 0, got %d", result.VariantCounts["0"])
	}
	if result.VariantCounts["1"] != 1 {
		t.Errorf("expected 1 deck in variant 1, got %d", result.VariantCounts["1"])
	}

	// The match result (alice wins vs Dragapult) should be recorded in both variants.
	rec0 := result.Matchups["0"].Matchups["Dragapult"]
	if rec0.Wins != 1 || rec0.Losses != 0 {
		t.Errorf("variant 0 vs Dragapult: got %d/%d/%d, want 1/0/0", rec0.Wins, rec0.Losses, rec0.Ties)
	}

	rec1 := result.Matchups["1"].Matchups["Dragapult"]
	if rec1.Wins != 1 || rec1.Losses != 0 {
		t.Errorf("variant 1 vs Dragapult: got %d/%d/%d, want 1/0/0", rec1.Wins, rec1.Losses, rec1.Ties)
	}
}
