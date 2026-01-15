package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/vllry/professors-research/pkg/rk9"
	"github.com/vllry/professors-research/pkg/types"
)

type decklistJSON struct {
	Player string `json:"player"`
	Cards  []struct {
		Count int      `json:"count"`
		Card  cardJSON `json:"card"`
	} `json:"cards"`
}

type cardJSON struct {
	Name    string `json:"name"`
	SetCode string `json:"setCode"`
	Number  string `json:"number"`
}

type tournamentJSON struct {
	TournamentID   string    `json:"tournamentId"`
	TournamentName string    `json:"tournamentName,omitempty"`
	DownloadedAt   time.Time `json:"downloadedAt"`
	DecklistsFile  string    `json:"decklistsFile"`
	MatchesFile    string    `json:"matchesFile"`
}

func main() {
	tournamentID := flag.String("tournament", "", "RK9 tournament ID (e.g. LV01YShqrqjMo62PxZPg)")
	outDir := flag.String("out", "", "Output directory (default: ./data/tournaments/<tournamentID>)")
	saveHTML := flag.Bool("save-html", true, "Save fetched roster.html and pairings.html for debugging")
	maxDecklists := flag.Int("max-decklists", 0, "Limit how many decklists to download (0 = all). Useful for debugging.")
	decklistConcurrency := flag.Int("decklist-concurrency", 6, "How many decklists to download in parallel")
	pairingsConcurrency := flag.Int("pairings-concurrency", 8, "How many pairings fragments to download in parallel")
	httpTimeout := flag.Int("http-timeout-seconds", 25, "Per-request HTTP timeout in seconds")
	flag.Parse()

	if strings.TrimSpace(*tournamentID) == "" {
		fatalf("missing required -tournament")
	}

	targetDir := strings.TrimSpace(*outDir)
	if targetDir == "" {
		targetDir = filepath.Join(".", "data", "tournaments", safeFilename(*tournamentID))
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fatalf("mkdir %s: %v", targetDir, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	start := time.Now()
	lastPrint := time.Now()
	data, pages, err := rk9.FetchTournamentDataWithOptions(ctx, *tournamentID, rk9.FetchOptions{
		MaxDecklists:        *maxDecklists,
		DecklistConcurrency: *decklistConcurrency,
		PairingsConcurrency: *pairingsConcurrency,
		HTTPTimeoutSeconds:  *httpTimeout,
		OnProgress: func(e rk9.ProgressEvent) {
			// Throttle progress printing a bit.
			if time.Since(lastPrint) < 750*time.Millisecond && e.Done != e.Total {
				return
			}
			lastPrint = time.Now()
			if e.Total > 0 {
				fmt.Fprintf(os.Stderr, "[%s] %s %d/%d\n", time.Since(start).Truncate(time.Second), e.Phase, e.Done, e.Total)
			} else if e.Details != "" {
				fmt.Fprintf(os.Stderr, "[%s] %s %s\n", time.Since(start).Truncate(time.Second), e.Phase, e.Details)
			} else {
				fmt.Fprintf(os.Stderr, "[%s] %s\n", time.Since(start).Truncate(time.Second), e.Phase)
			}
		},
	})
	if err != nil {
		fatalf("fetch tournament data: %v", err)
	}

	if *saveHTML {
		if err := os.WriteFile(filepath.Join(targetDir, "roster.html"), []byte(pages.RosterHTML), 0o644); err != nil {
			fatalf("write roster.html: %v", err)
		}
		if err := os.WriteFile(filepath.Join(targetDir, "pairings.html"), []byte(pages.PairingsHTML), 0o644); err != nil {
			fatalf("write pairings.html: %v", err)
		}
	}

	// Write decklists
	decklistsPath := filepath.Join(targetDir, "decklists.json")
	decklistsPayload := encodeDecklists(data.Decklists)
	if err := writeJSON(decklistsPath, decklistsPayload); err != nil {
		fatalf("write %s: %v", decklistsPath, err)
	}

	// Write per-player decklist text (stable ordering)
	decklistsDir := filepath.Join(targetDir, "decklists")
	if err := os.MkdirAll(decklistsDir, 0o755); err != nil {
		fatalf("mkdir %s: %v", decklistsDir, err)
	}
	for player, dl := range data.Decklists {
		txt := formatDecklistLive(dl)
		if strings.TrimSpace(txt) == "" {
			continue
		}
		fn := safeFilename(player)
		if fn == "" {
			fn = "unknown"
		}
		if err := os.WriteFile(filepath.Join(decklistsDir, fn+".txt"), []byte(txt), 0o644); err != nil {
			fatalf("write decklist for %s: %v", player, err)
		}
	}

	// Write matches
	matchesPath := filepath.Join(targetDir, "matches.json")
	matches := data.Matches
	if matches == nil {
		matches = []types.MatchResult{}
	}
	if err := writeJSON(matchesPath, matches); err != nil {
		fatalf("write %s: %v", matchesPath, err)
	}

	// Save fetched HTML snapshots if requested.
	if *saveHTML {
		// We fetch again via the package internals? We don't expose raw HTML.
		// For now, skip (package can be extended later if needed).
		_ = saveHTML
	}

	// Write tournament metadata
	meta := tournamentJSON{
		TournamentID:   data.TournamentID,
		TournamentName: data.TournamentName,
		DownloadedAt:   time.Now().UTC(),
		DecklistsFile:  "decklists.json",
		MatchesFile:    "matches.json",
	}
	if err := writeJSON(filepath.Join(targetDir, "tournament.json"), meta); err != nil {
		fatalf("write tournament.json: %v", err)
	}

	fmt.Printf("Wrote %d decklists and %d matches to %s\n", len(data.Decklists), len(data.Matches), targetDir)
}

func encodeDecklists(m map[string]types.Decklist) []decklistJSON {
	players := make([]string, 0, len(m))
	for p := range m {
		players = append(players, p)
	}
	sort.Strings(players)

	out := make([]decklistJSON, 0, len(players))
	for _, p := range players {
		dl := m[p]
		items := make([]struct {
			Count int      `json:"count"`
			Card  cardJSON `json:"card"`
		}, 0, len(dl.Cards))

		// Stable ordering for JSON output.
		type row struct {
			c types.Card
			n int
		}
		var rows []row
		for c, n := range dl.Cards {
			rows = append(rows, row{c: c, n: n})
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].c.SetCode != rows[j].c.SetCode {
				return rows[i].c.SetCode < rows[j].c.SetCode
			}
			if rows[i].c.Number != rows[j].c.Number {
				return rows[i].c.Number < rows[j].c.Number
			}
			return rows[i].c.Name < rows[j].c.Name
		})

		for _, r := range rows {
			items = append(items, struct {
				Count int      `json:"count"`
				Card  cardJSON `json:"card"`
			}{
				Count: r.n,
				Card: cardJSON{
					Name:    r.c.Name,
					SetCode: r.c.SetCode,
					Number:  r.c.Number,
				},
			})
		}

		out = append(out, decklistJSON{
			Player: p,
			Cards:  items,
		})
	}
	return out
}

func formatDecklistLive(dl types.Decklist) string {
	// Attempt to format as Live decklist-style lines:
	// "<count> <name> <setCode> <number>"
	type row struct {
		c types.Card
		n int
	}
	var rows []row
	for c, n := range dl.Cards {
		rows = append(rows, row{c: c, n: n})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].c.SetCode != rows[j].c.SetCode {
			return rows[i].c.SetCode < rows[j].c.SetCode
		}
		if rows[i].c.Number != rows[j].c.Number {
			return rows[i].c.Number < rows[j].c.Number
		}
		return rows[i].c.Name < rows[j].c.Name
	})

	var b strings.Builder
	total := 0
	for _, r := range rows {
		total += r.n
		fmt.Fprintf(&b, "%d %s %s %s\n", r.n, r.c.Name, r.c.SetCode, r.c.Number)
	}
	if total > 0 {
		fmt.Fprintf(&b, "\nTotal Cards: %d\n", total)
	}
	return b.String()
}

func writeJSON(path string, v any) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func safeFilename(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Keep it simple: replace common problematic chars.
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
		"\t", "_",
	)
	s = replacer.Replace(s)
	// Collapse consecutive underscores
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return strings.Trim(s, "_")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
