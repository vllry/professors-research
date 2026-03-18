package matchups

import (
	"encoding/json"
	"os"

	"github.com/vllry/professors-research/pkg/types"
)

type decklistJSON struct {
	Player string `json:"player"`
	Cards  []struct {
		Count int      `json:"count"`
		Card  cardJSON `json:"card"`
	} `json:"cards"`
}

type cardJSON struct {
	Name    string `json:"name"`
	SetCode string `json:"setCode"`
	Number  string `json:"number"`
}

// LoadDecklists reads a decklists.json file and returns a map of player -> Decklist.
func LoadDecklists(path string) (map[string]types.Decklist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw []decklistJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]types.Decklist, len(raw))
	for _, entry := range raw {
		dl := types.Decklist{Cards: make(map[types.Card]int, len(entry.Cards))}
		for _, c := range entry.Cards {
			card := types.Card{
				Name:    c.Card.Name,
				SetCode: c.Card.SetCode,
				Number:  c.Card.Number,
			}
			dl.Cards[card] += c.Count
		}
		result[entry.Player] = dl
	}
	return result, nil
}

// LoadMatches reads a matches.json file and returns a slice of MatchResult.
func LoadMatches(path string) ([]types.MatchResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var matches []types.MatchResult
	if err := json.Unmarshal(data, &matches); err != nil {
		return nil, err
	}
	return matches, nil
}
