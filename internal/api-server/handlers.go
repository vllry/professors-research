package apiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/vllry/professors-research/internal/matchups"
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

	response := PrizeOddsResponse{
		Odds:   make(map[string][]float64),
		Errors: []APIError{},
	}

	// Identify basic Pokémon cards using the cache and check for unidentified cards
	basicPokemonCards := make(map[basictypes.Card]bool)
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
			continue
		}
		// Check if it's a PokemonCard with Stage Basic
		if pokemonCard, ok := detailedCard.(*basictypes.PokemonCard); ok {
			if pokemonCard.Stage == basictypes.StageBasic {
				basicPokemonCards[card] = true
			}
		}
	}

	odds, err := prizeodds.CalculatePrizeOddsWithOpeningHand(decklist, prized, basicPokemonCards)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to calculate prize odds: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert map[Card][]float64 to map[string][]float64 for JSON serialization
	for card, cardOdds := range odds {
		cardKey := fmt.Sprintf("%s %s %s", card.Name, card.SetCode, card.Number)
		response.Odds[cardKey] = cardOdds
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
		PossibleStarters: make(map[string]float64),
		ForcedStarters:   make(map[string]float64),
		Errors:           []APIError{},
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

func (s *Server) handleDrawSupporterOdds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DrawSupporterOddsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic input constraints
	if req.DeckSize < 1 || req.DeckSize > 60 {
		s.sendError(w, "deckSize must be between 1 and 60", http.StatusBadRequest)
		return
	}
	if req.KnownBottom < 0 || req.KnownBottom > req.DeckSize-1 {
		s.sendError(w, "knownBottom must be between 0 and deckSize-1", http.StatusBadRequest)
		return
	}
	if req.HandSize < 1 || req.HandSize > 60 {
		s.sendError(w, "handSize must be between 1 and 60", http.StatusBadRequest)
		return
	}
	if req.PrizeCards < 1 || req.PrizeCards > 6 {
		s.sendError(w, "prizeCards must be between 1 and 6", http.StatusBadRequest)
		return
	}
	if req.DeckSize+req.HandSize+req.PrizeCards > 59 {
		s.sendError(w, "deckSize + handSize + prizeCards must be <= 59 (at least 1 card must be in play)", http.StatusBadRequest)
		return
	}

	clampDrawCount := func(poolSize, drawCount int) int {
		if poolSize < 0 {
			poolSize = 0
		}
		if drawCount < 0 {
			drawCount = 0
		}
		if drawCount > poolSize {
			drawCount = poolSize
		}
		return drawCount
	}

	calcRow := func(poolSize, drawCount int) []float64 {
		drawCount = clampDrawCount(poolSize, drawCount)

		row := make([]float64, 4)
		for i := 1; i <= 4; i++ {
			target := i
			if target > poolSize {
				target = poolSize
			}
			row[i-1] = prizeodds.CalculateDrawOdds(poolSize, drawCount, target)
		}
		return row
	}

	calcPairTable := func(poolSize, drawCount int) [][]float64 {
		drawCount = clampDrawCount(poolSize, drawCount)

		table := make([][]float64, 4)
		for a := 1; a <= 4; a++ {
			row := make([]float64, 4)
			for b := 1; b <= 4; b++ {
				if a > poolSize || b > poolSize || a+b > poolSize {
					row[b-1] = 0.0
					continue
				}
				row[b-1] = prizeodds.CalculateDrawPairOdds(poolSize, drawCount, a, b)
			}
			table[a-1] = row
		}
		return table
	}

	poolTop := req.DeckSize - req.KnownBottom
	ionoDraw := req.PrizeCards
	researchDraw := 7

	// Lillie's shuffles deck+hand; known bottom is not preserved.
	lilliesPool := req.DeckSize + req.HandSize - 1
	lilliesDraw := 6
	if req.PrizeCards == 6 {
		lilliesDraw = 8
	}

	effectiveDrawCounts := map[string]int{
		"Iono":                   clampDrawCount(poolTop, ionoDraw),
		"Professor's Research":   clampDrawCount(poolTop, researchDraw),
		"Lillie's Determination": clampDrawCount(lilliesPool, lilliesDraw),
	}

	drawIntoBottom := func(drawCount int) int {
		overflow := drawCount - poolTop
		if overflow <= 0 {
			return 0
		}
		if overflow > req.KnownBottom {
			return req.KnownBottom
		}
		return overflow
	}

	bottomOdds := map[string][]float64{}
	bottomDrawCounts := map[string]int{}
	if req.KnownBottom > 0 {
		if bottomDraw := drawIntoBottom(ionoDraw); bottomDraw > 0 {
			bottomOdds["Iono"] = calcRow(req.KnownBottom, bottomDraw)
			bottomDrawCounts["Iono"] = bottomDraw
		}
		if bottomDraw := drawIntoBottom(researchDraw); bottomDraw > 0 {
			bottomOdds["Professor's Research"] = calcRow(req.KnownBottom, bottomDraw)
			bottomDrawCounts["Professor's Research"] = bottomDraw
		}
	}

	response := DrawSupporterOddsResponse{
		Odds: map[string][]float64{
			"Iono":                   calcRow(poolTop, effectiveDrawCounts["Iono"]),
			"Professor's Research":   calcRow(poolTop, effectiveDrawCounts["Professor's Research"]),
			"Lillie's Determination": calcRow(lilliesPool, effectiveDrawCounts["Lillie's Determination"]),
		},
		PairOdds: map[string][][]float64{
			"Iono":                   calcPairTable(poolTop, effectiveDrawCounts["Iono"]),
			"Professor's Research":   calcPairTable(poolTop, effectiveDrawCounts["Professor's Research"]),
			"Lillie's Determination": calcPairTable(lilliesPool, effectiveDrawCounts["Lillie's Determination"]),
		},
		DrawCounts: map[string]int{
			"Iono":                   ionoDraw,
			"Professor's Research":   researchDraw,
			"Lillie's Determination": lilliesDraw,
		},
		EffectiveDrawCounts: effectiveDrawCounts,
	}
	if len(bottomOdds) > 0 {
		response.BottomOdds = bottomOdds
		response.BottomDrawCounts = bottomDrawCounts
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleDeckVariants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeckVariantsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.TournamentIDs) == 0 {
		s.sendError(w, "tournamentIds must contain at least 1 tournament", http.StatusBadRequest)
		return
	}
	if req.Archetype == "" {
		s.sendError(w, "archetype is required", http.StatusBadRequest)
		return
	}
	if req.N <= 0 {
		s.sendError(w, "n must be a positive integer", http.StatusBadRequest)
		return
	}
	if s.archetypes == nil {
		s.sendError(w, "Archetype definitions not loaded", http.StatusServiceUnavailable)
		return
	}

	tournamentIDs, tournamentDirs, err := s.validateAndResolveTournamentDirs(req.TournamentIDs)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := matchups.ComputeDeckVariants(tournamentIDs, tournamentDirs, s.archetypes, req.Archetype, req.N)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to compute deck variants: %v", err), http.StatusInternalServerError)
		return
	}

	response := DeckVariantsResponse{
		TotalDecks: result.TotalDecks,
		CoreCards:  result.CoreCards,
		Packages:   result.Packages,
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
	_ http.ResponseWriter,
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

func (s *Server) handleMatchupStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MatchupStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tournamentIDs := req.TournamentIDs
	if len(tournamentIDs) == 0 && req.TournamentID != "" {
		tournamentIDs = []string{req.TournamentID}
	}
	if len(tournamentIDs) == 0 {
		s.sendError(w, "tournamentIds must contain at least 1 tournament", http.StatusBadRequest)
		return
	}
	if req.Archetype == "" {
		s.sendError(w, "archetype is required", http.StatusBadRequest)
		return
	}
	if s.archetypes == nil {
		s.sendError(w, "Archetype definitions not loaded", http.StatusServiceUnavailable)
		return
	}

	tournamentIDs, tournamentDirs, validateErr := s.validateAndResolveTournamentDirs(tournamentIDs)
	if validateErr != nil {
		s.sendError(w, validateErr.Error(), http.StatusBadRequest)
		return
	}

	filter := matchups.PlacementFilter{
		PlayerPercentile:   req.PlayerPlacement,
		OpponentPercentile: req.OpponentPlacement,
	}

	var (
		result *matchups.MatchupResult
		err    error
	)
	if len(tournamentIDs) == 1 {
		result, err = matchups.ComputeMatchups(tournamentDirs[0], s.archetypes, req.Archetype, req.Variants, filter)
	} else {
		result, err = matchups.ComputeMatchupsForTournaments(tournamentIDs, tournamentDirs, s.archetypes, req.Archetype, req.Variants, filter)
	}
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to compute matchup stats: %v", err), http.StatusInternalServerError)
		return
	}

	response := MatchupStatsResponse{
		Matchups:        result.Matchups,
		ArchetypeCounts: result.ArchetypeCounts,
		VariantCounts:   result.VariantCounts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
