package types

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// DetailedCard represents a card with more than baseline information.
// Cards are split into distinct types with their respective schemas.
// We load this data from the data/cards directory.
type DetailedCard interface {
	SetCode() string // Returns the standard set abreviation used on the card (e.g. TWM)
	Name() string    // Returns the card name
	Number() string  // Returns the card number
	Kind() Kind      // Returns the card kind

}

// Kind represents the category of card (trainer, energy, pokemon)
type Kind string

// Stage represents the evolution stage of a Pokemon card.
type Stage string

const (
	KindTrainer Kind = "trainer"
	KindEnergy  Kind = "energy"
	KindPokemon Kind = "pokemon"

	StageBasic  Stage = "Basic"
	StageStage1 Stage = "Stage1"
	StageStage2 Stage = "Stage2"
)

// String returns the string representation of the Stage.
func (s Stage) String() string {
	return string(s)
}

// BaseCardInfo contains the essential fields common to all card types.
// This is embedded in each specific card type to avoid repetition.
type BaseCardInfo struct {
	setCode string // Set code extracted from physicalSetCode (e.g. TWM, SFA)
	number  string // Card number from localId
	name    string // Card name
	kind    Kind   // Card kind
}

// cardData represents the JSON structure of a card for efficient single-pass parsing.
type cardData struct {
	Category string `json:"category"`
	LocalID  string `json:"localId"`
	Name     string `json:"name"`
	Set      struct {
		ID string `json:"id"`
	} `json:"set"`
	Stage string `json:"stage,omitempty"` // Only present for Pokemon cards
}

// SetCode returns the set code from the embedded BaseCardInfo.
// This method is automatically available to all card types that embed BaseCardInfo.
func (b BaseCardInfo) SetCode() string {
	return b.setCode
}

// Name returns the card name.
func (b BaseCardInfo) Name() string {
	return b.name
}

// Number returns the card number.
func (b BaseCardInfo) Number() string {
	return b.number
}

// Kind returns the card kind.
func (b BaseCardInfo) Kind() Kind {
	return b.kind
}

// PokemonCard represents a Pokemon card with Pokemon-specific fields.
type PokemonCard struct {
	BaseCardInfo
	Stage Stage `json:"stage"`
}

// setBaseCardInfo sets the base card information from parsed data.
func (b *BaseCardInfo) setBaseCardInfo(setCode, number, name string) {
	b.setCode = setCode
	b.number = number
	b.name = name
}

// TrainerCard represents a Trainer card.
type TrainerCard struct {
	BaseCardInfo
}

// EnergyCard represents an Energy card.
type EnergyCard struct {
	BaseCardInfo
}

// cardJSONWrapper represents the JSON structure with a "cards" array.
type cardJSONWrapper struct {
	PhysicalSetCode string            `json:"physicalSetCode"`
	Cards           []json.RawMessage `json:"cards"`
}

// normalizeSetCode translates set codes from the dataset format to the decklist format.
// This handles cases where the dataset uses different codes than what appears in decklists.
func normalizeSetCode(setCode string) string {
	// Map of dataset codes to decklist codes
	setCodeMap := map[string]string{
		"SV": "SVI", // Dataset uses "SV" but decklists use "SVI"
	}

	if normalized, ok := setCodeMap[setCode]; ok {
		return normalized
	}
	return setCode
}

// NewDetailedCardsFromJSON loads PokemonCard, TrainerCard, and EnergyCard objects from a JSON file.
// The JSON file should have a structure like: {"cards": [...]}
// Each card has a "category" field that determines its type.
// This function parses each card only once for efficiency.
func NewDetailedCardsFromJSON(reader io.Reader) ([]DetailedCard, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON data: %w", err)
	}

	var wrapper cardJSONWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON wrapper: %w", err)
	}

	// Use physicalSetCode (e.g. "TWM", "SFA") instead of set.id (e.g. "sv06")
	// This matches the set codes used in decklists
	physicalSetCode := wrapper.PhysicalSetCode
	if physicalSetCode == "" {
		return nil, fmt.Errorf("JSON file missing physicalSetCode field")
	}

	// Normalize set code to match decklist format (e.g. SV -> SVI)
	physicalSetCode = normalizeSetCode(physicalSetCode)

	var cards []DetailedCard
	for i, cardDataRaw := range wrapper.Cards {
		// Parse the card data once to extract all needed fields
		var cardData cardData
		if err := json.Unmarshal(cardDataRaw, &cardData); err != nil {
			return nil, fmt.Errorf("failed to parse card at index %d: %w", i, err)
		}

		// Create the appropriate card type based on category
		switch strings.ToLower(cardData.Category) {
		case "pokemon":
			pokemonCard := &PokemonCard{
				Stage: Stage(cardData.Stage),
			}
			pokemonCard.setBaseCardInfo(physicalSetCode, cardData.LocalID, cardData.Name)
			pokemonCard.kind = KindPokemon
			cards = append(cards, pokemonCard)
		case "trainer":
			trainerCard := &TrainerCard{}
			trainerCard.setBaseCardInfo(physicalSetCode, cardData.LocalID, cardData.Name)
			trainerCard.kind = KindTrainer
			cards = append(cards, trainerCard)
		case "energy":
			energyCard := &EnergyCard{}
			energyCard.setBaseCardInfo(physicalSetCode, cardData.LocalID, cardData.Name)
			energyCard.kind = KindEnergy
			cards = append(cards, energyCard)
		default:
			return nil, fmt.Errorf("unknown card category '%s' at index %d", cardData.Category, i)
		}
	}

	return cards, nil
}
