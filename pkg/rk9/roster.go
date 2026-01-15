package rk9

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/vllry/professors-research/pkg/types"
)

// parseRoster extracts player decklists and tournament name from the RK9 roster
// page HTML. Returns a map of player identifier (ID or name) to decklist, the
// tournament name (if found), and any error encountered.
//
// If decklists are not available on the roster page, the returned map will be
// empty but no error will be returned (this is expected for some tournaments).
func parseRoster(html string) (map[string]types.Decklist, string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, "", errors.Wrap(err, "parse HTML")
	}

	tournamentName := extractTournamentName(doc)
	decklists := make(map[string]types.Decklist)

	// Try to extract player information and decklists from the roster page.
	// The structure may vary, so we'll try multiple approaches.
	doc.Find("table, .roster, [data-player], tr").Each(func(i int, s *goquery.Selection) {
		// Look for player rows or elements containing player data
		playerID, playerName := extractPlayerIdentifier(s)
		if playerID == "" && playerName == "" {
			return
		}

		// Use ID if available, otherwise use name
		identifier := playerID
		if identifier == "" {
			identifier = playerName
		}

		// Try to find decklist for this player
		decklistText := extractDecklist(s, doc)
		if decklistText != "" {
			decklist, err := types.NewDecklistFromLive(decklistText)
			if err == nil {
				decklists[identifier] = decklist
			}
			// Silently skip decklists that fail to parse - they may be incomplete
		}
	})

	// If we didn't find decklists in the main structure, try looking for links
	// to decklist pages or embedded decklist data
	if len(decklists) == 0 {
		doc.Find("a[href*='decklist'], a[href*='deck']").Each(func(i int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			playerID, playerName := extractPlayerIdentifier(s)
			identifier := playerID
			if identifier == "" {
				identifier = playerName
			}
			if identifier != "" && href != "" {
				// For now, we'll note that decklists may be on separate pages
				// This could be extended to fetch those pages if needed
				_ = href
			}
		})
	}

	return decklists, tournamentName, nil
}

// extractTournamentName attempts to find the tournament name from the page.
func extractTournamentName(doc *goquery.Document) string {
	// Try common heading selectors
	selectors := []string{
		"h1", "h2", "h3", "h4", "h5",
		".tournament-name", ".event-name", "[data-tournament-name]",
		"title",
	}

	var name string
	for _, sel := range selectors {
		doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			if name != "" {
				return // Already found a name
			}
			text := strings.TrimSpace(s.Text())
			if text != "" {
				// Check for tournament-like keywords or just take the first h4/h5
				lowerText := strings.ToLower(text)
				if strings.Contains(lowerText, "championship") ||
					strings.Contains(lowerText, "tournament") ||
					strings.Contains(lowerText, "regional") ||
					strings.Contains(lowerText, "pokémon") ||
					(sel == "h4" || sel == "h5") {
					name = text
				}
			}
		})
		if name != "" {
			return name
		}
	}

	// Fallback: look for h4 or h5 with any text
	if name == "" {
		doc.Find("h4, h5").Each(func(i int, s *goquery.Selection) {
			if name == "" {
				text := strings.TrimSpace(s.Text())
				if text != "" {
					name = text
				}
			}
		})
	}

	return name
}

// extractPlayerIdentifier attempts to extract a player ID or name from a selection.
// Returns (playerID, playerName). Either may be empty.
func extractPlayerIdentifier(s *goquery.Selection) (string, string) {
	// Try data attributes first (most reliable)
	if id, ok := s.Attr("data-player-id"); ok && id != "" {
		return id, ""
	}
	if id, ok := s.Attr("data-id"); ok && id != "" {
		return id, ""
	}

	// Try to find player name in text or child elements
	playerName := strings.TrimSpace(s.Text())
	if playerName == "" {
		// Look in child elements
		s.Find("td:first-child, .player-name, [data-name]").Each(func(i int, child *goquery.Selection) {
			if i == 0 {
				playerName = strings.TrimSpace(child.Text())
			}
		})
	}

	// Clean up player name (remove extra whitespace, records like "(16-0-2)", etc.)
	playerName = cleanPlayerName(playerName)

	return "", playerName
}

// cleanPlayerName removes common artifacts from player names like records,
// country codes, etc.
func cleanPlayerName(name string) string {
	// Remove records in parentheses like "(16-0-2)"
	re := regexp.MustCompile(`\s*\(\d+-\d+-\d+\)\s*`)
	name = re.ReplaceAllString(name, " ")

	// Remove country codes in brackets like "[NO]", "[US]"
	re = regexp.MustCompile(`\s*\[[A-Z]{2,3}\]\s*`)
	name = re.ReplaceAllString(name, " ")

	// Clean up whitespace
	name = strings.TrimSpace(name)
	re = regexp.MustCompile(`\s+`)
	name = re.ReplaceAllString(name, " ")

	return name
}

// extractDecklist attempts to extract a decklist from a selection or the document.
// Returns the raw decklist text if found, empty string otherwise.
func extractDecklist(s *goquery.Selection, doc *goquery.Document) string {
	// Look for decklist in data attributes
	if decklist, ok := s.Attr("data-decklist"); ok && decklist != "" {
		return decklist
	}

	// Look for decklist in nearby elements
	s.Find(".decklist, [data-decklist], pre, textarea").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		// Check if it looks like a decklist (has card count patterns)
		if matched, _ := regexp.MatchString(`^\d+\s+\w+`, text); matched {
			// This might be a decklist, but we'll let NewDecklistFromLive validate it
		}
	})

	// Look for links to decklist pages
	decklistLink := s.Find("a[href*='decklist'], a[href*='deck']").First()
	if decklistLink.Length() > 0 {
		// For now, we don't follow links automatically
		// This could be extended if needed
	}

	return ""
}

