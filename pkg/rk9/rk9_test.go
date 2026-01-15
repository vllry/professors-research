package rk9

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/vllry/professors-research/pkg/types"
)

func TestParseRoster(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		wantDecklists  int
		wantTournament string
	}{
		{
			name: "basic roster with tournament name",
			html: `
			<html>
				<head><title>Tournament Pairings</title></head>
				<body>
					<h4>2026 Las Vegas Pokémon TCG Regional Championships</h4>
					<table>
						<tr>
							<td>1. Tord Reklev [NO]</td>
							<td>(16-0-2)</td>
						</tr>
						<tr>
							<td>2. Liam Halliburton [US]</td>
							<td>(14-2-2)</td>
						</tr>
					</table>
				</body>
			</html>`,
			wantDecklists:  0, // No decklists in this example
			wantTournament: "2026 Las Vegas Pokémon TCG Regional Championships",
		},
		{
			name: "roster with decklist data",
			html: `
			<html>
				<body>
					<h4>Test Tournament</h4>
					<div data-player-id="123" data-decklist="4 Pikachu SVI 1
4 Lightning Energy SVI 1
52 other cards...">
						Player Name
					</div>
				</body>
			</html>`,
			wantDecklists:  0, // Decklist parsing would need valid 60-card decklist
			wantTournament: "Test Tournament",
		},
		{
			name:           "empty HTML",
			html:           `<html><body></body></html>`,
			wantDecklists:  0,
			wantTournament: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decklists, tournamentName, err := parseRoster(tt.html)
			if err != nil {
				t.Fatalf("parseRoster() error = %v", err)
			}

			if len(decklists) != tt.wantDecklists {
				t.Errorf("parseRoster() decklists count = %d, want %d", len(decklists), tt.wantDecklists)
			}

			if tournamentName != tt.wantTournament {
				t.Errorf("parseRoster() tournamentName = %q, want %q", tournamentName, tt.wantTournament)
			}
		})
	}
}

func TestParsePairings(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantMin int // Minimum expected matches
	}{
		{
			name: "basic pairings table",
			html: `
			<html>
				<body>
					<table>
						<thead>
							<tr><th>Player 1</th><th>Table #</th><th>Player 2</th></tr>
						</thead>
						<tbody>
							<tr>
								<td>Tord Reklev [NO] (16-0-2)</td>
								<td>1</td>
								<td>Liam Halliburton [US] (14-2-2)</td>
							</tr>
							<tr>
								<td>Player A</td>
								<td>2</td>
								<td>Player B</td>
							</tr>
						</tbody>
					</table>
				</body>
			</html>`,
			wantMin: 2,
		},
		{
			name: "round-specific structure",
			html: `
			<html>
				<body>
					<div data-round="1">
						<table>
							<tr>
								<td>Player 1</td>
								<td>1</td>
								<td>Player 2</td>
							</tr>
						</table>
					</div>
					<div data-round="2">
						<table>
							<tr>
								<td>Player 3</td>
								<td>1</td>
								<td>Player 4</td>
							</tr>
						</table>
					</div>
				</body>
			</html>`,
			wantMin: 2,
		},
		{
			name:    "empty HTML",
			html:    `<html><body></body></html>`,
			wantMin: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := parsePairings(tt.html)
			if err != nil {
				t.Fatalf("parsePairings() error = %v", err)
			}

			if len(matches) < tt.wantMin {
				t.Errorf("parsePairings() matches count = %d, want at least %d. Matches: %+v", len(matches), tt.wantMin, matches)
			}

			// Verify match structure if we have matches
			for i, match := range matches {
				if match.Round <= 0 {
					t.Errorf("parsePairings() match[%d].Round = %d, want > 0", i, match.Round)
				}
				if match.Player1 == "" && match.Player2 == "" {
					t.Errorf("parsePairings() match[%d] has no players", i)
				}
			}
		})
	}
}

func TestCleanPlayerName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "name with record",
			input:    "Tord Reklev [NO] (16-0-2)",
			expected: "Tord Reklev",
		},
		{
			name:     "name with country code",
			input:    "Liam Halliburton [US]",
			expected: "Liam Halliburton",
		},
		{
			name:     "name with both",
			input:    "Player Name [CA] (10-5-1)",
			expected: "Player Name",
		},
		{
			name:     "simple name",
			input:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "extra whitespace",
			input:    "  Player   Name  ",
			expected: "Player Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanPlayerName(tt.input)
			if got != tt.expected {
				t.Errorf("cleanPlayerName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractRoundNumber(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		selector string
		expected int
	}{
		{
			name:     "data-round attribute",
			html:     `<div data-round="5">Round 5</div>`,
			selector: "div",
			expected: 5,
		},
		{
			name:     "id with round",
			html:     `<div id="round3">Content</div>`,
			selector: "div",
			expected: 3,
		},
		{
			name:     "id with R prefix",
			html:     `<div id="R7">Content</div>`,
			selector: "div",
			expected: 7,
		},
		{
			name:     "text content",
			html:     `<div>Round 12</div>`,
			selector: "div",
			expected: 12,
		},
		{
			name:     "no round found",
			html:     `<div>Some content</div>`,
			selector: "div",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			selection := doc.Find(tt.selector)
			if selection.Length() == 0 {
				t.Fatalf("Selector %q found no elements", tt.selector)
			}

			got := extractRoundNumber(selection.First())
			if got != tt.expected {
				t.Errorf("extractRoundNumber() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestParseResult(t *testing.T) {
	tests := []struct {
		name          string
		resultText    string
		player1       string
		player2       string
		wantOutcome   types.MatchOutcome
		wantWinner    string
	}{
		{
			name:        "win indicator",
			resultText:  "W",
			player1:     "Player 1",
			player2:     "Player 2",
			wantOutcome: types.MatchOutcomeWin,
			wantWinner:  "", // Can't determine from just "W"
		},
		{
			name:        "loss indicator",
			resultText:  "L",
			player1:     "Player 1",
			player2:     "Player 2",
			wantOutcome: types.MatchOutcomeLoss,
			wantWinner:  "",
		},
		{
			name:        "tie indicator",
			resultText:  "T",
			player1:     "Player 1",
			player2:     "Player 2",
			wantOutcome: types.MatchOutcomeTie,
			wantWinner:  "",
		},
		{
			name:        "score format player1 wins",
			resultText:  "2-0",
			player1:     "Player 1",
			player2:     "Player 2",
			wantOutcome: types.MatchOutcomeWin,
			wantWinner:  "Player 1",
		},
		{
			name:        "score format player2 wins",
			resultText:  "1-2",
			player1:     "Player 1",
			player2:     "Player 2",
			wantOutcome: types.MatchOutcomeWin,
			wantWinner:  "Player 2",
		},
		{
			name:        "score format tie",
			resultText:  "1-1",
			player1:     "Player 1",
			player2:     "Player 2",
			wantOutcome: types.MatchOutcomeTie,
			wantWinner:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutcome, gotWinner := parseResult(tt.resultText, tt.player1, tt.player2)
			if gotOutcome != tt.wantOutcome {
				t.Errorf("parseResult() outcome = %v, want %v", gotOutcome, tt.wantOutcome)
			}
			if gotWinner != tt.wantWinner {
				t.Errorf("parseResult() winner = %q, want %q", gotWinner, tt.wantWinner)
			}
		})
	}
}

