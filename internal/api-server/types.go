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
