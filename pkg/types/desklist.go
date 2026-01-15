package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Card represents a generic identifiable card, with a name,
// This is the base information that we get from a Live decklist.
// Name can be used to help identifiy unknown trainer/energy cards, and for faster dedupe lookups.
// TODO: rename to something like CardIdentifier to avoid confusion with richer card type(s)
type Card struct {
	SetCode string
	Number  string
	Name    string
}

type Decklist struct {
	Cards map[Card]int // Card -> count
}

func NewDecklistFromLive(liveDecklist string) (Decklist, error) {
	decklist := Decklist{
		Cards: make(map[Card]int),
	}
	lines := strings.Split(liveDecklist, "\n")
	for _, line := range lines {
		// Valid card lines have the form "<number> <card name> <set code> <card number>",
		// where card name can contain spaces, punctuation, etc.
		//
		// Known/valid non-card lines have the form "<category>: <number>",
		// "Total Cards: <number>", or an empty line.

		split := strings.SplitN(line, " ", -1)
		// Try to process any line that must be a card.
		if len(split) >= 4 {
			cardCount, err := strconv.Atoi(split[0])
			if err != nil {
				return decklist, errors.Wrapf(err, "failed to parse card count from %s", split[0])
			}

			cardName := strings.Join(split[1:len(split)-2], " ")

			cardNumber, err := strconv.Atoi(split[len(split)-1]) // Last element
			if err != nil {
				return decklist, errors.Wrapf(err, "failed to parse card number from %s", split[len(split)-1])
			}

			setCode := split[len(split)-2]

			// Insert a new card or increment the count of an existing card
			decklist.Cards[Card{
				SetCode: setCode,
				Number:  strconv.Itoa(cardNumber),
				Name:    cardName,
			}] += cardCount
		}
	}

	// Count total cards (not unique cards)
	totalCards := 0
	for _, count := range decklist.Cards {
		totalCards += count
	}
	if totalCards != 60 {
		return decklist, errors.Wrapf(fmt.Errorf("decklist must contain 60 cards"), "contains %d cards", totalCards)
	}

	return decklist, nil
}
