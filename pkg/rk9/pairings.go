package rk9

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/vllry/professors-research/pkg/types"
)

// extractRoundNumber attempts to extract the round number from a selection.
// Returns 0 if not found.
func extractRoundNumber(s *goquery.Selection) int {
	// Try data-round attribute
	if roundStr, ok := s.Attr("data-round"); ok {
		if round, err := strconv.Atoi(roundStr); err == nil && round > 0 {
			return round
		}
	}

	// Try ID attribute (e.g., "round1", "R1")
	if id, ok := s.Attr("id"); ok {
		re := regexp.MustCompile(`(?i)(?:round|r)(\d+)`)
		matches := re.FindStringSubmatch(id)
		if len(matches) > 1 {
			if round, err := strconv.Atoi(matches[1]); err == nil && round > 0 {
				return round
			}
		}
	}

	// Try text content (e.g., "Round 1", "R1")
	text := strings.TrimSpace(s.Text())
	re := regexp.MustCompile(`(?i)(?:round\s*|r\s*)(\d+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if round, err := strconv.Atoi(matches[1]); err == nil && round > 0 {
			return round
		}
	}

	return 0
}

// parseRoundMatches extracts all matches from a round section/element.
func parseRoundMatches(roundElem *goquery.Selection, roundNum int) []types.MatchResult {
	var matches []types.MatchResult

	// Look for table rows (tr) that represent matches
	// Use Find() which searches descendants, not just direct children
	roundElem.Find("tr").Each(func(i int, row *goquery.Selection) {
		// Skip header rows (rows with th elements)
		if row.Find("th").Length() > 0 {
			return
		}
		// Skip rows that are part of thead
		parent := row.Parent()
		if parent.Length() > 0 && parent.Is("thead") {
			return
		}
		match := parseMatchRow(row, roundNum)
		if match != nil {
			matches = append(matches, *match)
		}
	})

	// Also look for div-based match structures
	roundElem.Find(".match, [data-match], div.match-row").Each(func(i int, matchElem *goquery.Selection) {
		match := parseMatchElement(matchElem, roundNum)
		if match != nil {
			matches = append(matches, *match)
		}
	})

	return matches
}

// parseMatchRow parses a table row that represents a match.
func parseMatchRow(row *goquery.Selection, roundNum int) *types.MatchResult {
	cells := row.Find("td")
	if cells.Length() < 2 {
		return nil // Not enough columns for a match
	}

	// Some RK9 tables use <td> for headers. Skip obvious header rows like:
	// "Player 1 | Table # | Player 2".
	rowText := strings.ToLower(strings.Join(strings.Fields(row.Text()), " "))
	if strings.Contains(rowText, "player 1") && strings.Contains(rowText, "player 2") && strings.Contains(rowText, "table") {
		return nil
	}

	var player1, player2 string
	var tableNum int
	var winner string
	var outcome types.MatchOutcome

	// Common table structure: Player1 | Table | Player2 | Result
	// Or: Table | Player1 | Player2 | Result
	cells.Each(func(i int, cell *goquery.Selection) {
		text := strings.TrimSpace(cell.Text())
		if text == "" {
			return
		}

		// Check if this looks like a table number (small positive integer)
		if tableNum == 0 {
			if num, err := strconv.Atoi(text); err == nil && num > 0 && num < 10000 {
				tableNum = num
				// Continue processing - don't treat table number as a player name, but keep going
				return
			}
		}

		// Check if this looks like a result (do this before player name to avoid confusion)
		if isResultText(text) {
			outcome, winner = parseResult(text, player1, player2)
			return
		}

		// Skip table headers
		if isTableHeader(text) {
			return
		}

		// Check if this looks like a player name
		// Only assign if it's not already a table number or result
		if player1 == "" {
			player1 = cleanPlayerName(text)
		} else if player2 == "" {
			player2 = cleanPlayerName(text)
		}
	})

	// If we have at least one player, create a match result
	if player1 != "" || player2 != "" {
		if outcome == "" {
			outcome = types.MatchOutcomeOther
		}
		return &types.MatchResult{
			Round:   roundNum,
			Table:   tableNum,
			Player1: player1,
			Player2: player2,
			Outcome: outcome,
			Winner:  winner,
		}
	}

	return nil
}

// parseMatchElement parses a div or other element that represents a match.
func parseMatchElement(elem *goquery.Selection, roundNum int) *types.MatchResult {
	var player1, player2 string
	var tableNum int
	var winner string
	var outcome types.MatchOutcome

	// Look for player names in common class names
	elem.Find(".player1, .player-1, [data-player1], .player:first-child").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			player1 = cleanPlayerName(s.Text())
		}
	})

	elem.Find(".player2, .player-2, [data-player2], .player:last-child").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			player2 = cleanPlayerName(s.Text())
		}
	})

	// Look for table number
	if tableStr, ok := elem.Attr("data-table"); ok {
		if num, err := strconv.Atoi(tableStr); err == nil {
			tableNum = num
		}
	}
	elem.Find(".table, .table-number, [data-table]").Each(func(i int, s *goquery.Selection) {
		if tableNum == 0 {
			text := strings.TrimSpace(s.Text())
			if num, err := strconv.Atoi(text); err == nil && num > 0 {
				tableNum = num
			}
		}
	})

	// Look for result
	elem.Find(".result, .match-result, [data-result]").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if isResultText(text) {
			outcome, winner = parseResult(text, player1, player2)
		}
	})

	if player1 != "" || player2 != "" {
		if outcome == "" {
			outcome = types.MatchOutcomeOther
		}
		return &types.MatchResult{
			Round:   roundNum,
			Table:   tableNum,
			Player1: player1,
			Player2: player2,
			Outcome: outcome,
			Winner:  winner,
		}
	}

	return nil
}

// isTableHeader checks if text looks like a table/table-number header cell.
func isTableHeader(text string) bool {
	text = strings.ToLower(text)
	return text == "table" || text == "table #" || text == "table#" || text == "#"
}

// isResultText checks if text looks like a match result indicator.
func isResultText(text string) bool {
	text = strings.ToUpper(strings.TrimSpace(text))
	return text == "W" || text == "L" || text == "T" || text == "WIN" || text == "LOSS" ||
		text == "TIE" || text == "DRAW" || strings.Contains(text, "WINS") ||
		strings.Contains(text, "WON") || regexp.MustCompile(`^\d+-\d+$`).MatchString(text)
}

// parseResult attempts to determine the match outcome and winner from result text.
// Returns (outcome, winner).
func parseResult(resultText, player1, player2 string) (types.MatchOutcome, string) {
	resultText = strings.ToUpper(strings.TrimSpace(resultText))

	// Check for explicit win/loss indicators
	if resultText == "W" || resultText == "WIN" || strings.Contains(resultText, "WINS") ||
		strings.Contains(resultText, "WON") {
		// Need to determine which player won - this may require looking at records
		// For now, return WIN with empty winner (caller may need to infer)
		return types.MatchOutcomeWin, ""
	}
	if resultText == "L" || resultText == "LOSS" {
		return types.MatchOutcomeLoss, ""
	}
	if resultText == "T" || resultText == "TIE" || resultText == "DRAW" {
		return types.MatchOutcomeTie, ""
	}

	// Check for score format (e.g., "2-0", "2-1")
	scoreRe := regexp.MustCompile(`^(\d+)-(\d+)$`)
	matches := scoreRe.FindStringSubmatch(resultText)
	if len(matches) == 3 {
		score1, _ := strconv.Atoi(matches[1])
		score2, _ := strconv.Atoi(matches[2])
		if score1 > score2 {
			return types.MatchOutcomeWin, player1
		} else if score2 > score1 {
			return types.MatchOutcomeWin, player2
		} else {
			return types.MatchOutcomeTie, ""
		}
	}

	return types.MatchOutcomeOther, ""
}

