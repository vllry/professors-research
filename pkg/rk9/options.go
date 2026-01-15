package rk9

import "context"

type ProgressPhase string

const (
	PhaseFetchRoster        ProgressPhase = "fetch_roster"
	PhaseFetchPairings      ProgressPhase = "fetch_pairings"
	PhaseParseRoster        ProgressPhase = "parse_roster"
	PhaseFetchDecklists     ProgressPhase = "fetch_decklists"
	PhaseFetchPairingsFrags ProgressPhase = "fetch_pairings_fragments"
)

// ProgressEvent is emitted as FetchTournamentDataWithOptions makes progress.
type ProgressEvent struct {
	Phase   ProgressPhase
	Done    int
	Total   int
	Details string
}

type FetchOptions struct {
	// MaxDecklists limits how many decklists will be downloaded from the roster page.
	// 0 means no limit.
	MaxDecklists int

	// DecklistConcurrency controls how many decklist pages are fetched in parallel.
	DecklistConcurrency int

	// PairingsConcurrency controls how many pairings fragments are fetched in parallel.
	PairingsConcurrency int

	// MaxPairingsFragments limits how many pairings fragments (hx-get endpoints) are fetched.
	// 0 means no limit.
	MaxPairingsFragments int

	// OnProgress is called occasionally to report progress. It should be fast/non-blocking.
	OnProgress func(ProgressEvent)

	// HTTPTimeoutSeconds overrides the per-request timeout. 0 means default.
	HTTPTimeoutSeconds int
}

// FetchTournamentDataWithOptions is the main entry point for downloading and parsing RK9
// tournament data with progress reporting and concurrency.
func FetchTournamentDataWithOptions(ctx context.Context, tournamentID string, opts FetchOptions) (TournamentData, TournamentPages, error) {
	return fetchTournamentData(ctx, tournamentID, opts)
}




