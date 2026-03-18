package matchups

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vllry/professors-research/pkg/types"
)

type PackageConfig struct {
	Count int            `json:"count"`
	Cards map[string]int `json:"cards"`
}

type VariantPackage struct {
	Cards   []string        `json:"cards"`
	Configs []PackageConfig `json:"configs"`
}

type DeckVariantsResult struct {
	TotalDecks int              `json:"totalDecks"`
	CoreCards  map[string]int   `json:"coreCards"`
	Packages   []VariantPackage `json:"packages"`
}

// ComputeDeckVariants discovers correlated card packages among decks of the
// given archetype across the specified tournaments.
func ComputeDeckVariants(
	tournamentIDs, tournamentDirs []string,
	archetypes []Archetype,
	targetArchetype string,
	n int,
) (*DeckVariantsResult, error) {
	if len(tournamentIDs) != len(tournamentDirs) {
		return nil, fmt.Errorf("tournamentIDs and tournamentDirs must have the same length")
	}
	if len(tournamentIDs) == 0 {
		return nil, fmt.Errorf("at least one tournament is required")
	}

	combinedDecklists := make(map[string]types.Decklist)
	for i := range tournamentIDs {
		id := tournamentIDs[i]
		dir := tournamentDirs[i]
		ns := id + "|"

		decklists, err := LoadDecklists(filepath.Join(dir, "decklists.json"))
		if err != nil {
			return nil, fmt.Errorf("loading decklists for tournament %q: %w", id, err)
		}
		for player, dl := range decklists {
			combinedDecklists[ns+player] = dl
		}
	}

	return computeDeckVariantsFromData(combinedDecklists, archetypes, targetArchetype, n)
}

// computeDeckVariantsFromData is the pure-logic core, separated for testing.
func computeDeckVariantsFromData(
	decklists map[string]types.Decklist,
	archetypes []Archetype,
	targetArchetype string,
	n int,
) (*DeckVariantsResult, error) {
	var archetypeDecks []map[string]int
	for _, deck := range decklists {
		if ClassifyDeck(deck, archetypes) == targetArchetype {
			archetypeDecks = append(archetypeDecks, DeckCardNameCounts(deck))
		}
	}

	if len(archetypeDecks) == 0 {
		return &DeckVariantsResult{
			TotalDecks: 0,
			CoreCards:  map[string]int{},
			Packages:   []VariantPackage{},
		}, nil
	}

	coreCards := findCoreCards(archetypeDecks)
	packages := detectPackages(archetypeDecks, coreCards)

	sort.Slice(packages, func(i, j int) bool {
		if len(packages[i].Cards) != len(packages[j].Cards) {
			return len(packages[i].Cards) > len(packages[j].Cards)
		}
		return packages[i].Cards[0] < packages[j].Cards[0]
	})

	if n > 0 && n < len(packages) {
		packages = packages[:n]
	}

	return &DeckVariantsResult{
		TotalDecks: len(archetypeDecks),
		CoreCards:  coreCards,
		Packages:   packages,
	}, nil
}

// findCoreCards returns card names whose count is identical across every deck.
func findCoreCards(decks []map[string]int) map[string]int {
	if len(decks) == 0 {
		return map[string]int{}
	}

	allNames := make(map[string]bool)
	for _, d := range decks {
		for name := range d {
			allNames[name] = true
		}
	}

	core := make(map[string]int)
	for name := range allNames {
		ref := decks[0][name]
		same := true
		for _, d := range decks[1:] {
			if d[name] != ref {
				same = false
				break
			}
		}
		if same && ref > 0 {
			core[name] = ref
		}
	}
	return core
}

// detectPackages groups non-core cards into packages of correlated cards using
// partition fingerprinting. Cards that split the deck population into the same
// subsets (regardless of actual count values) are placed in the same package.
func detectPackages(decks []map[string]int, coreCards map[string]int) []VariantPackage {
	flexNames := make(map[string]bool)
	for _, deck := range decks {
		for name := range deck {
			if _, isCore := coreCards[name]; !isCore {
				flexNames[name] = true
			}
		}
	}
	// Include cards that appear in zero decks' core but are absent in some
	// decks (count 0 for missing entries is handled by map default).

	if len(flexNames) == 0 {
		return nil
	}

	// Group flex cards by their partition fingerprint.
	fpGroups := make(map[string][]string)
	for name := range flexNames {
		fp := partitionFingerprint(name, decks)
		fpGroups[fp] = append(fpGroups[fp], name)
	}

	var packages []VariantPackage
	for _, cardNames := range fpGroups {
		sort.Strings(cardNames)

		configCounts := make(map[string]int)
		configCards := make(map[string]map[string]int)

		for _, deck := range decks {
			cfg := make(map[string]int)
			for _, name := range cardNames {
				if c := deck[name]; c > 0 {
					cfg[name] = c
				}
			}
			key := configKey(cfg)
			configCounts[key]++
			if _, exists := configCards[key]; !exists {
				configCards[key] = cfg
			}
		}

		type entry struct {
			key   string
			count int
		}
		entries := make([]entry, 0, len(configCounts))
		for k, c := range configCounts {
			entries = append(entries, entry{k, c})
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].count != entries[j].count {
				return entries[i].count > entries[j].count
			}
			return entries[i].key < entries[j].key
		})

		configs := make([]PackageConfig, len(entries))
		for i, e := range entries {
			configs[i] = PackageConfig{
				Count: e.count,
				Cards: configCards[e.key],
			}
		}

		packages = append(packages, VariantPackage{
			Cards:   cardNames,
			Configs: configs,
		})
	}

	return packages
}

// partitionFingerprint computes a canonical string representing how a card's
// count values partition the set of decks. Two cards with the same fingerprint
// split the deck population into identical subsets, making them correlated.
func partitionFingerprint(cardName string, decks []map[string]int) string {
	// Group deck indices by the card's count value.
	groups := make(map[int][]int)
	for i, deck := range decks {
		groups[deck[cardName]] = append(groups[deck[cardName]], i)
	}

	// Sort groups by their first deck index for a canonical ordering.
	type group struct {
		indices []int
	}
	sorted := make([]group, 0, len(groups))
	for _, indices := range groups {
		sorted = append(sorted, group{indices})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].indices[0] < sorted[j].indices[0]
	})

	var b strings.Builder
	for i, g := range sorted {
		if i > 0 {
			b.WriteByte('|')
		}
		for j, idx := range g.indices {
			if j > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, "%d", idx)
		}
	}
	return b.String()
}

// configKey produces a canonical string for grouping identical card-count maps.
func configKey(cards map[string]int) string {
	if len(cards) == 0 {
		return ""
	}
	names := make([]string, 0, len(cards))
	for n := range cards {
		names = append(names, n)
	}
	sort.Strings(names)

	var b strings.Builder
	for i, n := range names {
		if i > 0 {
			b.WriteByte(';')
		}
		fmt.Fprintf(&b, "%s=%d", n, cards[n])
	}
	return b.String()
}
