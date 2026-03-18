package matchups

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/vllry/professors-research/pkg/types"
)

type MatchupRecord struct {
	Wins    int     `json:"wins"`
	Losses  int     `json:"losses"`
	Ties    int     `json:"ties"`
	WinRate float64 `json:"winRate"`
}

func newMatchupRecord(wins, losses, ties int) MatchupRecord {
	total := wins + losses + ties
	var winRate float64
	if total > 0 {
		winRate = (float64(wins) + float64(ties)/3.0) / float64(total)
	}
	return MatchupRecord{
		Wins:    wins,
		Losses:  losses,
		Ties:    ties,
		WinRate: winRate,
	}
}

type VariantMatchupStats struct {
	// CardCounts is the "card count map" definition for the variant.
	// For numeric variant keys, this corresponds to the variants[] entry at that index.
	// For "other", it is an empty map.
	CardCounts map[string]int `json:"cardCounts"`
	// Matchups maps opponent archetype -> record.
	Matchups map[string]MatchupRecord `json:"matchups"`
}

type MatchupResult struct {
	// Matchups maps variant key -> opponent archetype -> record.
	Matchups map[string]VariantMatchupStats `json:"matchups"`
	// ArchetypeCounts maps archetype name -> number of decks classified as that archetype.
	ArchetypeCounts map[string]int `json:"archetypeCounts"`
	// VariantCounts maps variant key -> number of decks classified as that variant.
	VariantCounts map[string]int `json:"variantCounts"`
}

func variantKey(i int) string {
	return strconv.Itoa(i)
}

// PlacementFilter controls percentile-based filtering for the matchups computation.
// Each field is a top-X% value (e.g. 10 = top 10%, 100 = all players).
// Zero means no filter (equivalent to 100).
type PlacementFilter struct {
	PlayerPercentile   float64
	OpponentPercentile float64
}

func needsPlacementFilter(percentile float64) bool {
	return percentile > 0 && percentile < 100
}

// computeMatchPoints returns the total match points for every player appearing
// in the given matches (3 per win, 1 per tie, 0 per loss).
func computeMatchPoints(matches []types.MatchResult) map[string]int {
	pts := make(map[string]int)
	for _, m := range matches {
		if m.Player1 != "" {
			if _, ok := pts[m.Player1]; !ok {
				pts[m.Player1] = 0
			}
		}
		if m.Player2 != "" {
			if _, ok := pts[m.Player2]; !ok {
				pts[m.Player2] = 0
			}
		}
		switch m.Outcome {
		case types.MatchOutcomeTie:
			pts[m.Player1]++
			pts[m.Player2]++
		default:
			if m.Winner == m.Player1 {
				pts[m.Player1] += 3
			} else if m.Winner == m.Player2 {
				pts[m.Player2] += 3
			}
		}
	}
	return pts
}

// buildPercentileSet returns the set of players whose match points place them
// in the top percentile% of all players. Returns nil when no filtering is
// needed (percentile <= 0 or >= 100).
func buildPercentileSet(matchPoints map[string]int, percentile float64) map[string]bool {
	if percentile <= 0 || percentile >= 100 {
		return nil
	}
	if len(matchPoints) == 0 {
		return nil
	}

	scores := make([]int, 0, len(matchPoints))
	for _, pts := range matchPoints {
		scores = append(scores, pts)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(scores)))

	cutIndex := int(float64(len(scores)) * percentile / 100.0)
	if cutIndex < 1 {
		cutIndex = 1
	}
	if cutIndex > len(scores) {
		cutIndex = len(scores)
	}
	threshold := scores[cutIndex-1]

	allowed := make(map[string]bool, cutIndex)
	for player, pts := range matchPoints {
		if pts >= threshold {
			allowed[player] = true
		}
	}
	return allowed
}

// ComputeMatchups loads tournament data from tournamentDir, classifies decks,
// and computes win rates for each variant of targetArchetype vs every opponent archetype.
func ComputeMatchups(
	tournamentDir string,
	archetypes []Archetype,
	targetArchetype string,
	variants []map[string]int,
	filter PlacementFilter,
) (*MatchupResult, error) {
	decklists, err := LoadDecklists(filepath.Join(tournamentDir, "decklists.json"))
	if err != nil {
		return nil, fmt.Errorf("loading decklists: %w", err)
	}
	matches, err := LoadMatches(filepath.Join(tournamentDir, "matches.json"))
	if err != nil {
		return nil, fmt.Errorf("loading matches: %w", err)
	}

	var playerFilter map[string]bool
	var opponentFilter map[string]bool
	if needsPlacementFilter(filter.PlayerPercentile) || needsPlacementFilter(filter.OpponentPercentile) {
		standings, err := LoadStandingsFromRosterHTML(filepath.Join(tournamentDir, "roster.html"))
		if err != nil {
			return nil, fmt.Errorf("loading standings from roster: %w", err)
		}
		playerFilter = buildStandingsPercentileSet(standings, filter.PlayerPercentile)
		opponentFilter = buildStandingsPercentileSet(standings, filter.OpponentPercentile)
	}

	return computeMatchupsFromData(decklists, matches, archetypes, targetArchetype, variants, playerFilter, opponentFilter)
}

// ComputeMatchupsForTournaments loads multiple tournaments, namespaces player identifiers
// per tournament to avoid collisions, and computes aggregated matchup stats.
// Placement percentiles are computed per tournament so that tournaments with
// different numbers of rounds are not mixed.
func ComputeMatchupsForTournaments(
	tournamentIDs []string,
	tournamentDirs []string,
	archetypes []Archetype,
	targetArchetype string,
	variants []map[string]int,
	filter PlacementFilter,
) (*MatchupResult, error) {
	if len(tournamentIDs) != len(tournamentDirs) {
		return nil, fmt.Errorf("tournamentIDs and tournamentDirs must have the same length")
	}
	if len(tournamentIDs) == 0 {
		return nil, fmt.Errorf("at least one tournament is required")
	}

	combinedDecklists := make(map[string]types.Decklist)
	combinedMatches := make([]types.MatchResult, 0)

	var combinedPlayerFilter map[string]bool
	var combinedOpponentFilter map[string]bool

	for i := range tournamentIDs {
		id := tournamentIDs[i]
		dir := tournamentDirs[i]
		ns := id + "|"

		decklists, err := LoadDecklists(filepath.Join(dir, "decklists.json"))
		if err != nil {
			return nil, fmt.Errorf("loading decklists for tournament %q: %w", id, err)
		}
		matches, err := LoadMatches(filepath.Join(dir, "matches.json"))
		if err != nil {
			return nil, fmt.Errorf("loading matches for tournament %q: %w", id, err)
		}

		var pf map[string]bool
		var of map[string]bool
		if needsPlacementFilter(filter.PlayerPercentile) || needsPlacementFilter(filter.OpponentPercentile) {
			standings, err := LoadStandingsFromRosterHTML(filepath.Join(dir, "roster.html"))
			if err != nil {
				return nil, fmt.Errorf("loading standings from roster for tournament %q: %w", id, err)
			}
			pf = buildStandingsPercentileSet(standings, filter.PlayerPercentile)
			of = buildStandingsPercentileSet(standings, filter.OpponentPercentile)
		}

		for player, dl := range decklists {
			combinedDecklists[ns+player] = dl
		}
		for _, m := range matches {
			nm := m
			nm.Player1 = ns + m.Player1
			nm.Player2 = ns + m.Player2
			if m.Winner != "" {
				nm.Winner = ns + m.Winner
			}
			combinedMatches = append(combinedMatches, nm)
		}

		if pf != nil {
			if combinedPlayerFilter == nil {
				combinedPlayerFilter = make(map[string]bool)
			}
			for player := range pf {
				combinedPlayerFilter[ns+player] = true
			}
		}
		if of != nil {
			if combinedOpponentFilter == nil {
				combinedOpponentFilter = make(map[string]bool)
			}
			for player := range of {
				combinedOpponentFilter[ns+player] = true
			}
		}
	}

	return computeMatchupsFromData(combinedDecklists, combinedMatches, archetypes, targetArchetype, variants, combinedPlayerFilter, combinedOpponentFilter)
}

// playerFilter/opponentFilter: nil means no filtering; otherwise only players
// present in the set are eligible for the respective role.
func computeMatchupsFromData(
	decklists map[string]types.Decklist,
	matches []types.MatchResult,
	archetypes []Archetype,
	targetArchetype string,
	variants []map[string]int,
	playerFilter map[string]bool,
	opponentFilter map[string]bool,
) (*MatchupResult, error) {
	// Pre-filter out players who never played an actual match.
	//
	// Some tournaments include decklists/roster entries for players who dropped
	// before playing (or otherwise never appear in the match results). Those
	// players should not affect archetype counts or matchup win rates.
	played := make(map[string]bool)
	for _, m := range matches {
		// Require both seats to be present; ignore byes/unknown-opponent rows.
		if m.Player1 == "" || m.Player2 == "" {
			continue
		}
		played[m.Player1] = true
		played[m.Player2] = true
	}
	if len(played) > 0 {
		filtered := make(map[string]types.Decklist, len(decklists))
		for player, dl := range decklists {
			if played[player] {
				filtered[player] = dl
			}
		}
		decklists = filtered
	}

	// Classify every player's deck.
	playerArchetype := make(map[string]string, len(decklists))
	archetypeCounts := make(map[string]int)
	for player, deck := range decklists {
		arch := ClassifyDeck(deck, archetypes)
		playerArchetype[player] = arch
		archetypeCounts[arch]++
	}

	// Sub-classify target archetype decks into variants (inclusive: a deck can match multiple).
	playerVariants := make(map[string][]string)
	variantCounts := make(map[string]int)
	for player, deck := range decklists {
		if playerArchetype[player] != targetArchetype {
			continue
		}
		vks := ClassifyVariants(deck, variants)
		playerVariants[player] = vks
		for _, vk := range vks {
			variantCounts[vk]++
		}
	}

	// Tally wins/losses/ties: variant key -> opponent archetype -> (w, l, t)
	type tally struct{ w, l, t int }
	raw := make(map[string]map[string]*tally)

	ensureTally := func(vk, oppArch string) *tally {
		if raw[vk] == nil {
			raw[vk] = make(map[string]*tally)
		}
		if raw[vk][oppArch] == nil {
			raw[vk][oppArch] = &tally{}
		}
		return raw[vk][oppArch]
	}

	for _, m := range matches {
		arch1, ok1 := playerArchetype[m.Player1]
		arch2, ok2 := playerArchetype[m.Player2]
		if !ok1 || !ok2 {
			continue
		}

		// Record from p1's perspective if p1 is in the target archetype.
		if arch1 == targetArchetype {
			if (playerFilter == nil || playerFilter[m.Player1]) &&
				(opponentFilter == nil || opponentFilter[m.Player2]) {
				for _, vk := range playerVariants[m.Player1] {
					t := ensureTally(vk, arch2)
					switch m.Outcome {
					case types.MatchOutcomeTie:
						t.t++
					default:
						if m.Winner == m.Player1 {
							t.w++
						} else if m.Winner == m.Player2 {
							t.l++
						}
					}
				}
			}
		}

		// Record from p2's perspective if p2 is in the target archetype.
		if arch2 == targetArchetype {
			if (playerFilter == nil || playerFilter[m.Player2]) &&
				(opponentFilter == nil || opponentFilter[m.Player1]) {
				for _, vk := range playerVariants[m.Player2] {
					t := ensureTally(vk, arch1)
					switch m.Outcome {
					case types.MatchOutcomeTie:
						t.t++
					default:
						if m.Winner == m.Player2 {
							t.w++
						} else if m.Winner == m.Player1 {
							t.l++
						}
					}
				}
			}
		}
	}

	// Convert tallies to MatchupRecords.
	matchupsByVariant := make(map[string]map[string]MatchupRecord, len(raw))
	for vk, opponents := range raw {
		matchupsByVariant[vk] = make(map[string]MatchupRecord, len(opponents))
		for oppArch, t := range opponents {
			matchupsByVariant[vk][oppArch] = newMatchupRecord(t.w, t.l, t.t)
		}
	}

	attachCardCounts := func(vk string) map[string]int {
		if vk == "other" {
			return map[string]int{}
		}
		i, err := strconv.Atoi(vk)
		if err != nil || i < 0 || i >= len(variants) {
			return map[string]int{}
		}
		if variants[i] == nil {
			return map[string]int{}
		}
		out := make(map[string]int, len(variants[i]))
		for k, v := range variants[i] {
			out[k] = v
		}
		return out
	}

	matchups := make(map[string]VariantMatchupStats, len(matchupsByVariant))
	for vk, opponents := range matchupsByVariant {
		matchups[vk] = VariantMatchupStats{
			CardCounts: attachCardCounts(vk),
			Matchups:   opponents,
		}
	}

	return &MatchupResult{
		Matchups:        matchups,
		ArchetypeCounts: archetypeCounts,
		VariantCounts:   variantCounts,
	}, nil
}
