package apiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/vllry/professors-research/internal/prize-odds"
	basictypes "github.com/vllry/professors-research/pkg/types"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dataLoaded := s.detailedCards.IsLoaded()
	loadErr := s.detailedCards.LoadError()
	
	status := "ok"
	statusCode := http.StatusOK
	if !dataLoaded {
		status = "loading"
		statusCode = http.StatusServiceUnavailable
	} else if loadErr != nil {
		status = "error"
		statusCode = http.StatusInternalServerError
	}

	response := map[string]interface{}{
		"status":     status,
		"dataLoaded": dataLoaded,
	}

	// Include error if loading failed
	if loadErr != nil {
		response["error"] = loadErr.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handlePrizeOdds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PrizeOddsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	decklist, err := s.validateAndParseDecklist(req.Decklist)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Default prized to true if not specified
	prized := true
	if req.Prized != nil {
		prized = *req.Prized
	}

	// Check if cache is loaded
	if !s.detailedCards.IsLoaded() {
		s.sendError(w, "Card cache not loaded", http.StatusServiceUnavailable)
		return
	}

	odds, err := prizeodds.CalculatePrizeOdds(decklist, prized)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to calculate prize odds: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert map[Card][]float64 to map[string][]float64 for JSON serialization
	response := PrizeOddsResponse{
		Odds:   make(map[string][]float64),
		Errors: []APIError{},
	}

	for card, cardOdds := range odds {
		cardKey := fmt.Sprintf("%s %s %s", card.Name, card.SetCode, card.Number)
		response.Odds[cardKey] = cardOdds
	}

	// Check for unidentified cards and add warnings
	// Pattern to match "Basic {X} Energy" cards (e.g., "Basic {R} Energy", "Basic Fire Energy", "Basic Water Energy")
	basicEnergyPattern := regexp.MustCompile(`^Basic\s+(\{[A-Z]\}|\w+)\s+Energy$`)
	unidentifiedCards := []basictypes.Card{}
	
	for card := range decklist.Cards {
		cardNumber, err := strconv.Atoi(card.Number)
		if err != nil {
			// Skip cards with non-numeric numbers
			continue
		}
		detailedCard := s.detailedCards.Get(card.SetCode, cardNumber)
		if detailedCard == nil {
			// Card not found in cache - check if it's a basic energy (no warning needed)
			if basicEnergyPattern.MatchString(card.Name) {
				// Basic energy cards are universal, skip warning
				continue
			}
			// Card not found in cache and not a basic energy
			unidentifiedCards = append(unidentifiedCards, card)
		}
	}
	
	// Add warnings for unidentified cards
	for _, card := range unidentifiedCards {
		cardKey := fmt.Sprintf("%s %s %s", card.Name, card.SetCode, card.Number)
		response.Errors = append(response.Errors, APIError{
			Type: ErrorTypeUnidentifiedCard,
			Info: fmt.Sprintf("Card '%s' could not be identified in the cache of known cards", cardKey),
		})
	}

	// Process CardSets if provided
	if len(req.CardSets) > 0 {
		// Create calculator functions that capture the prized parameter
		cardSetCalculator := func(decklist basictypes.Decklist, cardSets map[string]basictypes.CardSet) (map[string]float64, error) {
			return prizeodds.CalculateCardSetPrizeOdds(decklist, cardSets, prized)
		}
		unionCalculator := func(combinations [][]basictypes.Card, decklist basictypes.Decklist) float64 {
			return prizeodds.CalculateUnionProbability(combinations, decklist, prized)
		}

		cardSetOdds, err := s.processCardSets(w, req.CardSets, decklist, cardSetCalculator, unionCalculator)
		if err != nil {
			// Determine status code based on error message
			statusCode := http.StatusBadRequest
			if strings.Contains(err.Error(), "Failed to calculate") {
				statusCode = http.StatusInternalServerError
			}
			s.sendError(w, err.Error(), statusCode)
			return
		}
		response.CardSetOdds = cardSetOdds
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleStartOdds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartOddsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	decklist, err := s.validateAndParseDecklist(req.Decklist)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	odds, err := prizeodds.CalculateStartOdds(decklist)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to calculate start odds: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if cache is loaded
	if !s.detailedCards.IsLoaded() {
		s.sendError(w, "Card cache not loaded", http.StatusServiceUnavailable)
		return
	}

	// Initialize response with errors slice
	response := StartOddsResponse{
		Odds:             make(map[string][]float64),
		PossibleStarters:  make(map[string]float64),
		ForcedStarters:    make(map[string]float64),
		Errors:            []APIError{},
	}

	// Identify basic Pokemon cards using the cache and check for unidentified cards
	basicPokemonCards := make(map[basictypes.Card]bool)
	unidentifiedCards := []basictypes.Card{}
	
	// Pattern to match "Basic {X} Energy" cards (e.g., "Basic {R} Energy", "Basic Fire Energy", "Basic Water Energy")
	// Matches both formats: "Basic {R} Energy" (with curly braces) and "Basic Fire Energy" (with word names)
	basicEnergyPattern := regexp.MustCompile(`^Basic\s+(\{[A-Z]\}|\w+)\s+Energy$`)
	
	for card := range decklist.Cards {
		cardNumber, err := strconv.Atoi(card.Number)
		if err != nil {
			// Skip cards with non-numeric numbers
			continue
		}
		detailedCard := s.detailedCards.Get(card.SetCode, cardNumber)
		if detailedCard == nil {
			// Card not found in cache - check if it's a basic energy (no warning needed)
			if basicEnergyPattern.MatchString(card.Name) {
				// Basic energy cards are universal, skip warning
				continue
			}
			// Card not found in cache and not a basic energy
			unidentifiedCards = append(unidentifiedCards, card)
			continue
		}
		// Check if it's a PokemonCard with Stage Basic
		if pokemonCard, ok := detailedCard.(*basictypes.PokemonCard); ok {
			if pokemonCard.Stage == basictypes.StageBasic {
				basicPokemonCards[card] = true
			}
		}
	}
	
	// Add warnings for unidentified cards
	for _, card := range unidentifiedCards {
		cardKey := fmt.Sprintf("%s %s %s", card.Name, card.SetCode, card.Number)
		response.Errors = append(response.Errors, APIError{
			Type: ErrorTypeUnidentifiedCard,
			Info: fmt.Sprintf("Card '%s' could not be identified in the cache of known cards", cardKey),
		})
	}

	// Calculate basic Pokemon start odds
	possibleStarters, forcedStarters, mulliganOdds := prizeodds.CalculateBasicPokemonStartOdds(decklist, basicPokemonCards)

	// Calculate total basic count for aggregate odds
	totalBasicCount := 0
	for card, count := range decklist.Cards {
		if basicPokemonCards[card] {
			totalBasicCount += count
		}
	}

	// Calculate at least 1 basic: P(>=1 basic) = 1 - P(0 basic) = 1 - mulliganOdds
	atLeastOneBasic := 1.0 - mulliganOdds

	// Calculate at least 2 basic: P(>=2 basic) = 1 - P(0 basic) - P(1 basic)
	atLeastTwoBasic := prizeodds.CalculateAtLeastTwoBasic(totalBasicCount)

	// Update response with calculated values
	response.MulliganOdds = mulliganOdds
	response.AtLeastOneBasic = atLeastOneBasic
	response.AtLeastTwoBasic = atLeastTwoBasic

	for card, cardOdds := range odds {
		cardKey := fmt.Sprintf("%s %s %s", card.Name, card.SetCode, card.Number)
		response.Odds[cardKey] = cardOdds
	}

	// Convert basic Pokemon odds to string keys
	for card, prob := range possibleStarters {
		cardKey := fmt.Sprintf("%s %s %s", card.Name, card.SetCode, card.Number)
		response.PossibleStarters[cardKey] = prob
	}

	for card, prob := range forcedStarters {
		cardKey := fmt.Sprintf("%s %s %s", card.Name, card.SetCode, card.Number)
		response.ForcedStarters[cardKey] = prob
	}

	// Process CardSets if provided
	if len(req.CardSets) > 0 {
		cardSetCalculator := func(decklist basictypes.Decklist, cardSets map[string]basictypes.CardSet) (map[string]float64, error) {
			return prizeodds.CalculateCardSetStartOdds(decklist, cardSets)
		}
		unionCalculator := func(combinations [][]basictypes.Card, decklist basictypes.Decklist) float64 {
			return prizeodds.CalculateUnionProbabilityStart(combinations, decklist)
		}

		cardSetOdds, err := s.processCardSets(w, req.CardSets, decklist, cardSetCalculator, unionCalculator)
		if err != nil {
			// Determine status code based on error message
			statusCode := http.StatusBadRequest
			if strings.Contains(err.Error(), "Failed to calculate") {
				statusCode = http.StatusInternalServerError
			}
			s.sendError(w, err.Error(), statusCode)
			return
		}
		response.CardSetOdds = cardSetOdds
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// CardSetCalculator is a function type that calculates odds for a single CardSet.
// It takes a decklist and a map of CardSets, and returns the odds for each CardSet.
type CardSetCalculator func(decklist basictypes.Decklist, cardSets map[string]basictypes.CardSet) (map[string]float64, error)

// UnionProbabilityCalculator is a function type that calculates the union probability
// of multiple combinations.
type UnionProbabilityCalculator func(combinations [][]basictypes.Card, decklist basictypes.Decklist) float64

// validateAndParseDecklist validates that the decklist string is not empty and parses it.
// Returns the parsed decklist or an error with appropriate status code.
func (s *Server) validateAndParseDecklist(decklistStr string) (basictypes.Decklist, error) {
	if decklistStr == "" {
		return basictypes.Decklist{}, fmt.Errorf("decklist is required")
	}

	decklist, err := basictypes.NewDecklistFromLive(decklistStr)
	if err != nil {
		return basictypes.Decklist{}, fmt.Errorf("failed to parse decklist: %v", err)
	}

	return decklist, nil
}

// processCardSets processes CardSets from the request and calculates their odds.
// It handles both single CardSet and multiple CardSet cases (union probability).
func (s *Server) processCardSets(
	w http.ResponseWriter,
	reqCardSets map[string][]CardSetJSON,
	decklist basictypes.Decklist,
	cardSetCalculator CardSetCalculator,
	unionCalculator UnionProbabilityCalculator,
) (map[string]map[string]float64, error) {
	if len(reqCardSets) == 0 {
		return nil, nil
	}

	result := make(map[string]map[string]float64)

	for groupName, cardSetList := range reqCardSets {
		groupResults := make(map[string]float64)

		// If there's only one CardSet, calculate it normally
		// If there are multiple CardSets, calculate the union probability
		if len(cardSetList) == 1 {
			cardSet, err := convertCardSetJSONToCardSet(cardSetList[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse CardSet %s[0]: %v", groupName, err)
			}

			cardSetsMap := map[string]basictypes.CardSet{groupName: cardSet}
			cardSetOdds, err := cardSetCalculator(decklist, cardSetsMap)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate CardSet odds: %v", err)
			}

			groupResults[groupName] = cardSetOdds[groupName]
		} else {
			// Multiple CardSets: calculate union probability
			allCombinations := [][]basictypes.Card{}

			for i, cardSetJSON := range cardSetList {
				cardSet, err := convertCardSetJSONToCardSet(cardSetJSON)
				if err != nil {
					return nil, fmt.Errorf("failed to parse CardSet %s[%d]: %v", groupName, i, err)
				}

				expanded := cardSet.Expand(decklist)
				allCombinations = append(allCombinations, expanded.Combinations...)
			}

			// Calculate union probability of all combinations
			unionProb := unionCalculator(allCombinations, decklist)
			groupResults[groupName] = unionProb
		}

		result[groupName] = groupResults
	}

	return result, nil
}

// convertCardSetJSONToCardSet converts a CardSetJSON to a basictypes.CardSet
func convertCardSetJSONToCardSet(cardSetJSON CardSetJSON) (basictypes.CardSet, error) {
	// Handle AnyOf CardSet
	if len(cardSetJSON.AnyOfs) > 0 {
		patterns := make([]basictypes.AnyOfPattern, 0, len(cardSetJSON.AnyOfs))
		
		for _, anyOfJSON := range cardSetJSON.AnyOfs {
			cards := make(map[basictypes.Card]int)
			for _, cardEntry := range anyOfJSON.Cards {
				card := basictypes.Card{
					Name:    cardEntry.Card.Name,
					SetCode: cardEntry.Card.SetCode,
					Number:  cardEntry.Card.Number,
				}
				cards[card] = cardEntry.Count
			}
			patterns = append(patterns, basictypes.AnyOfPattern{Cards: cards})
		}
		
		return basictypes.NewCardSet(patterns), nil
	}
	
	// Handle AllOf CardSet
	if len(cardSetJSON.AllOfs) > 0 {
		// AllOf expects a single list of cards, so we take the first AllOf entry
		// (or combine all if there are multiple - but typically there should be one)
		var allCards []basictypes.Card
		
		for _, allOfJSON := range cardSetJSON.AllOfs {
			for _, cardEntry := range allOfJSON.Cards {
				card := basictypes.Card{
					Name:    cardEntry.Card.Name,
					SetCode: cardEntry.Card.SetCode,
					Number:  cardEntry.Card.Number,
				}
				// Add the card 'count' times
				for i := 0; i < cardEntry.Count; i++ {
					allCards = append(allCards, card)
				}
			}
		}
		
		return basictypes.AllOf(allCards), nil
	}
	
	return nil, fmt.Errorf("CardSet must have either 'anyOfs' or 'allOfs'")
}

