package types

// MatchOutcome represents the result of a match from the tournament perspective.
// The values are intentionally simple strings so they are easy to log, store,
// and serialize if callers choose to do so.
type MatchOutcome string

const (
	MatchOutcomeWin   MatchOutcome = "WIN"
	MatchOutcomeLoss  MatchOutcome = "LOSS"
	MatchOutcomeTie   MatchOutcome = "TIE"
	MatchOutcomeOther MatchOutcome = "OTHER"
)

// MatchResult describes the outcome of a single match between two players in a
// given round of a tournament. Player identifiers should prefer RK9 player IDs
// when available, and fall back to player names when necessary.
type MatchResult struct {
	Round int // Tournament round number (1-based)
	Table int // Table number if known, 0 if unknown

	Player1 string       // Player identifier (ID or name) in the first seat
	Player2 string       // Player identifier (ID or name) in the second seat
	Outcome MatchOutcome // Outcome from the perspective of the match as a whole

	// Winner is the identifier (ID or name) of the winning player when there is
	// a clear winner. It is empty for ties or unknown outcomes.
	Winner string
}



