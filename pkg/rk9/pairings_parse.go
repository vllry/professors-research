package rk9

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/vllry/professors-research/pkg/types"
)

type pairingsFragmentRef struct {
	Pod   int
	Round int
	URL   string
}

// parsePairingsFragmentHTML parses the HTML fragment returned by the RK9 pairings
// hx-get endpoint (/pairings/<event>?pod=X&rnd=Y).
//
// The fragment typically contains multiple match rows, each as a div with class
// "row ... match ...". We use winner/loser CSS classes to infer the outcome.
func parsePairingsFragmentHTML(html string, round int, pod int, resolve func(string) string) []types.MatchResult {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}

	var out []types.MatchResult
	doc.Find(".match.complete, .row.match.complete").Each(func(i int, m *goquery.Selection) {
		p1 := m.Find(".player1").First()
		p2 := m.Find(".player2").First()
		if p1.Length() == 0 && p2.Length() == 0 {
			return
		}

		player1 := collapseWS(strings.TrimSpace(p1.Find(".name").Text()))
		if player1 == "" {
			player1 = collapseWS(strings.TrimSpace(p1.Text()))
		}
		player2 := collapseWS(strings.TrimSpace(p2.Find(".name").Text()))
		if player2 == "" {
			player2 = collapseWS(strings.TrimSpace(p2.Text()))
		}

		player1 = cleanPlayerName(player1)
		player2 = cleanPlayerName(player2)

		if resolve != nil {
			if r := resolve(player1); r != "" {
				player1 = r
			}
			if r := resolve(player2); r != "" {
				player2 = r
			}
		}

		table := 0
		tableText := collapseWS(strings.TrimSpace(m.Find(".tablenumber").First().Text()))
		if tableText != "" {
			if n, err := strconv.Atoi(tableText); err == nil {
				table = n
			}
		}

		outcome := types.MatchOutcomeOther
		winner := ""

		if hasClass(p1, "winner") {
			outcome = types.MatchOutcomeWin
			winner = player1
		} else if hasClass(p2, "winner") {
			outcome = types.MatchOutcomeWin
			winner = player2
		} else if hasClass(p1, "tie") || hasClass(p2, "tie") || hasClass(p1, "draw") || hasClass(p2, "draw") {
			outcome = types.MatchOutcomeTie
		}

		// If the element has an id like cell-<pod>-<round>-<table>-<seat>, override round/table when present.
		if id, ok := m.Find("[id^='cell-']").First().Attr("id"); ok && id != "" {
			if rr, tt := parseCellID(id); rr > 0 {
				round = rr
				if tt > 0 {
					table = tt
				}
			}
		}

		if player1 == "" && player2 == "" {
			return
		}

		out = append(out, types.MatchResult{
			Round:   round,
			Table:   table,
			Player1: player1,
			Player2: player2,
			Outcome: outcome,
			Winner:  winner,
		})
	})

	// If no matches were found, the fragment might be a placeholder or empty round.
	_ = pod
	return out
}

func hasClass(s *goquery.Selection, cls string) bool {
	if s == nil || s.Length() == 0 {
		return false
	}
	c, _ := s.Attr("class")
	for _, part := range strings.Fields(c) {
		if part == cls {
			return true
		}
	}
	return false
}

// parseCellID parses ids like "cell-0-14-1-1" => round=14, table=1.
func parseCellID(id string) (round int, table int) {
	re := regexp.MustCompile(`^cell-(\d+)-(\d+)-(\d+)-(\d+)$`)
	m := re.FindStringSubmatch(strings.TrimSpace(id))
	if len(m) != 5 {
		return 0, 0
	}
	r, _ := strconv.Atoi(m[2])
	t, _ := strconv.Atoi(m[3])
	return r, t
}

// parseRoundFromTabID parses ids like "P1R15" => 15.
func parseRoundFromTabID(id string) int {
	re := regexp.MustCompile(`(?i)r(\d+)`)
	m := re.FindStringSubmatch(id)
	if len(m) != 2 {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

// parsePairings remains for unit tests / standalone parsing of already-expanded HTML.
// It is not sufficient for real RK9 pages where content is loaded via htmx.
func parsePairings(html string) ([]types.MatchResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var matches []types.MatchResult
	doc.Find(".tab-pane[id*='R']").Each(func(i int, pane *goquery.Selection) {
		id, _ := pane.Attr("id")
		round := parseRoundFromTabID(id)
		if round <= 0 {
			return
		}
		inner, _ := pane.Html()
		matches = append(matches, parsePairingsFragmentHTML(inner, round, 0, nil)...)
	})

	// Fallback 1: parse any complete matches without round context as round 0.
	if len(matches) == 0 {
		matches = append(matches, parsePairingsFragmentHTML(html, 0, 0, nil)...)
	}

	// Fallback 2 (unit-test/simple HTML): parse table rows in [data-round] sections.
	if len(matches) == 0 {
		doc.Find("[data-round]").Each(func(i int, s *goquery.Selection) {
			round := extractRoundNumber(s)
			if round <= 0 {
				return
			}
			s.Find("tr").Each(func(_ int, row *goquery.Selection) {
				if row.Find("th").Length() > 0 {
					return
				}
				if m := parseMatchRow(row, round); m != nil {
					matches = append(matches, *m)
				}
			})
		})
	}

	// Fallback 3: parse any table rows as round 1.
	if len(matches) == 0 {
		doc.Find("tr").Each(func(_ int, row *goquery.Selection) {
			if row.Find("th").Length() > 0 {
				return
			}
			if m := parseMatchRow(row, 1); m != nil {
				matches = append(matches, *m)
			}
		})
	}

	return matches, nil
}


