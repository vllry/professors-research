package rk9

import (
	"context"

	"github.com/vllry/professors-research/pkg/types"
)

// TournamentData contains decklists and match results for a single RK9
// tournament. Player identifiers are the keys into the Decklists map, and are
// the same identifiers used in MatchResult.Player1 / Player2 / Winner.
type TournamentData struct {
	// TournamentID is the RK9 event identifier (e.g. "LV01YShqrqjMo62PxZPg").
	TournamentID string

	// TournamentName, when available, is a human-readable name parsed from the
	// roster or pairings page (e.g. "2026 Las Vegas Pokémon TCG Regional Championships").
	TournamentName string

	// Decklists maps a player identifier (preferably RK9 player ID, falling
	// back to name) to that player's decklist.
	Decklists map[string]types.Decklist

	// Matches contains the results of all matches in all rounds for the event.
	Matches []types.MatchResult
}

// FetchTournamentData downloads the roster and pairings pages for the given
// RK9 tournament ID, parses decklists and match results, and returns a
// populated TournamentData structure.
//
// It is the primary entry point for callers that want a single call to obtain
// all available tournament information from RK9.
func FetchTournamentData(tournamentID string) (TournamentData, error) {
	data, _, err := FetchTournamentDataWithOptions(context.Background(), tournamentID, FetchOptions{})
	return data, err
}



