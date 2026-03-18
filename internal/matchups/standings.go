package matchups

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func normalizeSpaces(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// LoadStandingsFromRosterHTML parses the official standings from a tournament
// roster HTML page.
//
// It returns a map keyed by the same player identifier strings used in
// decklists.json and matches.json, e.g.:
// "1....0 Alex Oliveira [BR]" -> 578
func LoadStandingsFromRosterHTML(path string) (map[string]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return nil, err
	}

	table := doc.Find("table#dtLiveRoster").First()
	if table.Length() == 0 {
		return nil, fmt.Errorf("dtLiveRoster table not found")
	}

	colIndex := map[string]int{}
	table.Find("thead tr th").Each(func(i int, s *goquery.Selection) {
		label := strings.ToLower(normalizeSpaces(s.Text()))
		switch label {
		case "player id":
			colIndex["player_id"] = i
		case "first name":
			colIndex["first_name"] = i
		case "last name":
			colIndex["last_name"] = i
		case "country":
			colIndex["country"] = i
		case "standing":
			colIndex["standing"] = i
		}
	})

	required := []string{"player_id", "first_name", "last_name", "country", "standing"}
	for _, k := range required {
		if _, ok := colIndex[k]; !ok {
			return nil, fmt.Errorf("missing required column %q in dtLiveRoster header", k)
		}
	}

	standings := make(map[string]int)
	table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
		cells := tr.Find("td")
		if cells.Length() == 0 {
			return
		}
		at := func(k string) string {
			i := colIndex[k]
			if i < 0 || i >= cells.Length() {
				return ""
			}
			return normalizeSpaces(cells.Eq(i).Text())
		}

		playerID := at("player_id")
		first := at("first_name")
		last := at("last_name")
		country := at("country")
		standingStr := at("standing")
		if playerID == "" || first == "" || last == "" || country == "" || standingStr == "" {
			return
		}
		standing, err := strconv.Atoi(standingStr)
		if err != nil || standing <= 0 {
			return
		}

		playerKey := fmt.Sprintf("%s %s %s [%s]", playerID, first, last, country)
		standings[playerKey] = standing
	})

	if len(standings) == 0 {
		return nil, fmt.Errorf("no standings parsed from dtLiveRoster")
	}

	return standings, nil
}

// buildStandingsPercentileSet returns the set of players whose official standing
// places them in the top percentile% of the tournament.
//
// "Top" is defined by the roster's Standing column: lower is better (1 is best).
// Returns nil when no filtering is needed (percentile <= 0 or >= 100).
func buildStandingsPercentileSet(standings map[string]int, percentile float64) map[string]bool {
	if percentile <= 0 || percentile >= 100 {
		return nil
	}
	if len(standings) == 0 {
		return nil
	}

	type entry struct {
		player   string
		standing int
	}
	entries := make([]entry, 0, len(standings))
	for p, s := range standings {
		entries = append(entries, entry{player: p, standing: s})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].standing != entries[j].standing {
			return entries[i].standing < entries[j].standing
		}
		return entries[i].player < entries[j].player
	})

	cutIndex := int(float64(len(entries)) * percentile / 100.0)
	if cutIndex < 1 {
		cutIndex = 1
	}
	if cutIndex > len(entries) {
		cutIndex = len(entries)
	}

	allowed := make(map[string]bool, cutIndex)
	for i := 0; i < cutIndex; i++ {
		allowed[entries[i].player] = true
	}
	return allowed
}

