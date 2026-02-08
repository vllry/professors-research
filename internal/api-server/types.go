package apiserver

// JSON-friendly structures for CardSet API
type CardJSON struct {
	Name    string `json:"name"`
	SetCode string `json:"setCode"`
	Number  string `json:"number"`
}

type AnyOfJSON struct {
	Cards []struct {
		Card  CardJSON `json:"card"`
		Count int      `json:"count"`
	} `json:"cards"`
}

type AllOfJSON struct {
	Cards []struct {
		Card  CardJSON `json:"card"`
		Count int      `json:"count"`
	} `json:"cards"`
}

type CardSetJSON struct {
	AnyOfs []AnyOfJSON `json:"anyOfs,omitempty"`
	AllOfs []AllOfJSON `json:"allOfs,omitempty"`
}

type PrizeOddsRequest struct {
	Decklist string                   `json:"decklist"`
	CardSets map[string][]CardSetJSON `json:"cardSets,omitempty"`
	Prized   *bool                    `json:"prized,omitempty"` // true = in 6 prize cards (default), false = in 54 not-prized cards
}

type PrizeOddsResponse struct {
	Odds        map[string][]float64          `json:"odds"`
	CardSetOdds map[string]map[string]float64 `json:"cardSetOdds,omitempty"`
	Errors      []APIError                    `json:"errors,omitempty"`
}

type StartOddsRequest struct {
	Decklist string                   `json:"decklist"`
	CardSets map[string][]CardSetJSON `json:"cardSets,omitempty"`
}

type StartOddsResponse struct {
	Odds             map[string][]float64          `json:"odds"`
	PossibleStarters map[string]float64            `json:"possibleStarters"`
	ForcedStarters   map[string]float64            `json:"forcedStarters"`
	MulliganOdds     float64                       `json:"mulliganOdds"`
	AtLeastOneBasic  float64                       `json:"atLeastOneBasic"`
	AtLeastTwoBasic  float64                       `json:"atLeastTwoBasic"`
	CardSetOdds      map[string]map[string]float64 `json:"cardSetOdds,omitempty"`
	Errors           []APIError                    `json:"errors,omitempty"`
}

type DrawSupporterOddsRequest struct {
	DeckSize    int `json:"deckSize"`
	KnownBottom int `json:"knownBottom"`
	HandSize    int `json:"handSize"`
	PrizeCards  int `json:"prizeCards"`
}

type DrawSupporterOddsResponse struct {
	// Odds maps supporter name -> odds for 1,2,3,4 copies of the target in the pool.
	Odds map[string][]float64 `json:"odds"`
	// PairOdds maps supporter name -> 4x4 table for odds of drawing at least one of BOTH cards,
	// indexed by [countA-1][countB-1] where countA,countB are 1..4.
	PairOdds map[string][][]float64 `json:"pairOdds"`
	// BottomOdds maps supporter name -> odds for 1,2,3,4 copies of the target
	// among known bottom cards when the draw goes past the top of deck.
	BottomOdds map[string][]float64 `json:"bottomOdds,omitempty"`
	// DrawCounts maps supporter name -> draw count for the top/shuffled pool.
	DrawCounts map[string]int `json:"drawCounts"`
	// EffectiveDrawCounts maps supporter name -> actual draw count used in the model
	// after clamping to the available pool.
	EffectiveDrawCounts map[string]int `json:"effectiveDrawCounts"`
	// BottomDrawCounts maps supporter name -> number of cards drawn into known bottom.
	BottomDrawCounts map[string]int `json:"bottomDrawCounts,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// ErrorType represents the type of error or warning
type ErrorType string

const (
	ErrorTypeUnidentifiedCard ErrorType = "UNIDENTIFIED_CARD"
	// Add more error types as needed
)

// APIError represents an error or warning in API responses
type APIError struct {
	Type ErrorType `json:"type"`
	Info string    `json:"info"`
}
