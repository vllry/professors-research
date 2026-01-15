// tcgdex_standard_dump.go
//
// Downloads (a) all Scarlet & Violet era sets (set IDs starting with "sv"),
// (b) all MEG era sets (set IDs starting with "meg" or "mep"), and
// (c) all promo sets (set IDs ending with "p"), then keeps only cards
// whose regulation mark is >= minMark (default: "G").
// Writes one JSON file per set into ./data/cards.
//
// Usage:
//   go run tcgdex_standard_dump.go
//   go run tcgdex_standard_dump.go -minMark=H -concurrency=16 -overwrite
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const baseURL = "https://api.tcgdex.net/v2/en"

type SetBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cardResult struct {
	id   string
	card map[string]any
	err  error
}

var safeFilenameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func main() {
	outDir := flag.String("out", "./data/cards", "Output directory for per-set JSON files")
	minMark := flag.String("minMark", "G", "Minimum regulation mark to include (e.g. G, H, I)")
	concurrency := flag.Int("concurrency", runtime.NumCPU()*4, "Max concurrent card downloads per set")
	retries := flag.Int("retries", 4, "HTTP retries for transient errors (5xx/429/network)")
	timeout := flag.Duration("timeout", 30*time.Second, "Per-request HTTP timeout")
	overwrite := flag.Bool("overwrite", false, "Overwrite existing set files")
	includeSV := flag.Bool("includeSV", true, "Include sets with IDs starting with 'sv' (Scarlet & Violet era)")
	includeMEG := flag.Bool("includeMEG", true, "Include sets with IDs starting with 'meg' or 'mep' (MEG era)")
	includePromos := flag.Bool("includePromos", true, "Include promo sets (set IDs ending with 'p')")
	flag.Parse()

	if *concurrency < 1 {
		*concurrency = 1
	}
	min := strings.ToUpper(strings.TrimSpace(*minMark))
	if min == "" {
		min = "G"
	}

	rand.Seed(time.Now().UnixNano())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := &http.Client{Timeout: *timeout}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatalf("mkdir %s: %v", *outDir, err)
	}

	sets, err := fetchSetList(ctx, client, *retries)
	if err != nil {
		fatalf("fetch set list: %v", err)
	}

	// Filter and sort set IDs for stable output.
	var targets []SetBrief
	for _, s := range sets {
		if shouldDownloadSet(s, *includeSV, *includeMEG, *includePromos) {
			targets = append(targets, s)
		}
	}
	sort.Slice(targets, func(i, j int) bool {
		return strings.ToLower(targets[i].ID) < strings.ToLower(targets[j].ID)
	})

	fmt.Printf("Found %d sets; targeting %d sets (includeSV=%v includeMEG=%v includePromos=%v)\n",
		len(sets), len(targets), *includeSV, *includeMEG, *includePromos)

	var totalSets, writtenSets, skippedSets, failedSets int
	var totalCards, keptCards, failedCards int

	for _, s := range targets {
		totalSets++
		setID := s.ID
		outPath := filepath.Join(*outDir, safeFilename(setID)+".json")
		if !*overwrite {
			if _, err := os.Stat(outPath); err == nil {
				fmt.Printf("SKIP  %s (file exists)\n", setID)
				skippedSets++
				continue
			}
		}

		fmt.Printf("SET   %s\n", setID)

		setObj, err := fetchSet(ctx, client, setID, *retries)
		if err != nil {
			fmt.Printf("  ERROR fetching set: %v\n", err)
			failedSets++
			continue
		}

		cardIDs, err := extractCardIDs(setObj)
		if err != nil {
			fmt.Printf("  ERROR reading set cards: %v\n", err)
			failedSets++
			continue
		}

		totalCards += len(cardIDs)

		physicalSetCode := deriveSetCode(setObj)
		setMeta := shallowCopyMap(setObj)
		delete(setMeta, "cards") // avoid duplicating a huge list

		cards, failed := downloadAndFilterCards(ctx, client, cardIDs, setID, min, *concurrency, *retries)
		keptCards += len(cards)
		failedCards += len(failed)

		if len(cards) == 0 {
			// Don't clutter output with empty sets (e.g., old promo sets after rotation).
			fmt.Printf("  NOTE no cards kept for %s (minMark=%s); not writing file\n", setID, min)
			if len(failed) > 0 {
				fmt.Printf("  NOTE %d cards failed to fetch\n", len(failed))
			}
			writtenSets++ // consider "processed" even if no file written? keep it simple: count as written.
			continue
		}

		sortCardsByLocalIDThenName(cards)

		payload := map[string]any{
			"downloadedAt":        time.Now().UTC().Format(time.RFC3339),
			"source":              baseURL,
			"minRegulationMark":   min,
			"set":                 setMeta,
			"physicalSetCode":     physicalSetCode,
			"cards":               cards,
			"failedCardIDs":       failed,
			"failedCardIDs_count": len(failed),
		}

		if err := writeJSONAtomic(outPath, payload, *overwrite); err != nil {
			fmt.Printf("  ERROR writing file: %v\n", err)
			failedSets++
			continue
		}

		fmt.Printf("  WROTE %s (%d cards kept, %d failed)\n", outPath, len(cards), len(failed))
		writtenSets++
	}

	fmt.Println()
	fmt.Printf("Done.\n")
	fmt.Printf("Sets:  processed=%d written=%d skipped=%d failed=%d\n", totalSets, writtenSets, skippedSets, failedSets)
	fmt.Printf("Cards: seen=%d kept=%d failed=%d\n", totalCards, keptCards, failedCards)
}

func shouldDownloadSet(s SetBrief, includeSV, includeMEG, includePromos bool) bool {
	lid := strings.ToLower(strings.TrimSpace(s.ID))
	lname := strings.ToLower(strings.TrimSpace(s.Name))
	if lid == "" {
		return false
	}
	if includeSV && strings.HasPrefix(lid, "sv") {
		return true
	}
	if includeMEG && (strings.HasPrefix(lid, "meg") || strings.HasPrefix(lid, "mep")) {
		return true
	}
	if includePromos && strings.HasSuffix(lid, "p") {
		return true
	}
	// Mega Evolution & Phantasmal Flames sets might not follow the simple ID
	// prefix conventions above. As a safety net, include any sets whose
	// human-readable names clearly indicate they are part of the Mega
	// Evolution block or the Phantasmal Flames expansion.
	if includeMEG {
		if strings.Contains(lname, "mega evolution") || strings.Contains(lname, "mega evolutions") {
			return true
		}
		if strings.Contains(lname, "phantasmal flames") {
			return true
		}
	}
	return false
}

func fetchSetList(ctx context.Context, client *http.Client, retries int) ([]SetBrief, error) {
	b, err := fetchBytes(ctx, client, baseURL+"/sets", retries)
	if err != nil {
		return nil, err
	}

	// Primary: []SetBrief
	var direct []SetBrief
	if err := json.Unmarshal(b, &direct); err == nil && len(direct) > 0 {
		return direct, nil
	}

	// Fallback: look for common wrapper keys (defensive).
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(b, &wrapper); err != nil {
		return nil, fmt.Errorf("parse /sets response: %w", err)
	}
	for _, key := range []string{"sets", "data", "results"} {
		if raw, ok := wrapper[key]; ok {
			var tmp []SetBrief
			if err := json.Unmarshal(raw, &tmp); err == nil && len(tmp) > 0 {
				return tmp, nil
			}
		}
	}

	// Final fallback: generic array of objects with "id"/"name".
	var anyArr []map[string]any
	if err := json.Unmarshal(b, &anyArr); err == nil && len(anyArr) > 0 {
		out := make([]SetBrief, 0, len(anyArr))
		for _, m := range anyArr {
			id, _ := m["id"].(string)
			name, _ := m["name"].(string)
			if strings.TrimSpace(id) == "" {
				continue
			}
			out = append(out, SetBrief{ID: id, Name: name})
		}
		if len(out) > 0 {
			return out, nil
		}
	}

	return nil, errors.New("could not parse set list from /sets")
}

func fetchSet(ctx context.Context, client *http.Client, setID string, retries int) (map[string]any, error) {
	b, err := fetchBytes(ctx, client, baseURL+"/sets/"+setID, retries)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("decode set %s: %w", setID, err)
	}
	return m, nil
}

func fetchCard(ctx context.Context, client *http.Client, cardID string, retries int) (map[string]any, error) {
	b, err := fetchBytes(ctx, client, baseURL+"/cards/"+cardID, retries)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("decode card %s: %w", cardID, err)
	}
	return m, nil
}

func extractCardIDs(setObj map[string]any) ([]string, error) {
	raw, ok := setObj["cards"]
	if !ok {
		return nil, errors.New("set has no 'cards' field")
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("set.cards is %T, expected array", raw)
	}

	ids := make([]string, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id, _ := m["id"].(string)
		if strings.TrimSpace(id) == "" {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.New("set contains zero card IDs")
	}
	return ids, nil
}

func downloadAndFilterCards(
	ctx context.Context,
	client *http.Client,
	cardIDs []string,
	setID string,
	minMark string,
	concurrency int,
	retries int,
) ([]map[string]any, []string) {
	jobs := make(chan string)
	results := make(chan cardResult)

	var wg sync.WaitGroup
	workerCount := concurrency
	if workerCount > len(cardIDs) {
		workerCount = len(cardIDs)
	}
	if workerCount < 1 {
		workerCount = 1
	}

	// Start workers.
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				card, err := fetchCard(ctx, client, id, retries)
				results <- cardResult{id: id, card: card, err: err}
			}
		}()
	}

	// Feed jobs.
	go func() {
		defer close(jobs)
		for _, id := range cardIDs {
			select {
			case <-ctx.Done():
				return
			case jobs <- id:
			}
		}
	}()

	// Close results when workers finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	var kept []map[string]any
	var failed []string

	for res := range results {
		if res.err != nil {
			failed = append(failed, res.id)
			continue
		}
		if shouldIncludeCard(res.card, setID, minMark) {
			kept = append(kept, res.card)
		}
	}

	return kept, failed
}

func shouldIncludeCard(card map[string]any, setID string, minMark string) bool {
	// First, respect tcgdex's notion of Standard legality if present.
	if rawLegal, ok := card["legal"]; ok {
		if legalMap, ok := rawLegal.(map[string]any); ok {
			if rawStd, ok := legalMap["standard"]; ok {
				if isStd, ok := rawStd.(bool); ok && !isStd {
					// Explicitly marked non-standard; never include.
					return false
				}
			}
		}
	}

	if mark, ok := getRegulationMark(card); ok {
		return markGE(mark, minMark)
	}

	// If there's no regulation mark but the card isn't explicitly marked
	// non-standard, keep it. This lets us pick up newly added Standard-legal
	// cards where tcgdex hasn't populated regulation marks yet, while still
	// excluding obviously non-standard promos like `basep` where
	// legal.standard=false.
	return true
}

func getRegulationMark(card map[string]any) (string, bool) {
	// Most likely key.
	if v, ok := card["regulationMark"]; ok {
		if s, ok := v.(string); ok {
			s = strings.ToUpper(strings.TrimSpace(s))
			if s != "" {
				return s, true
			}
		}
	}

	// Defensive: case-insensitive scan for a similar key.
	for k, v := range card {
		if strings.EqualFold(k, "regulationMark") || strings.EqualFold(k, "regulationmark") || strings.EqualFold(k, "regulation") {
			if s, ok := v.(string); ok {
				s = strings.ToUpper(strings.TrimSpace(s))
				if s != "" {
					return s, true
				}
			}
		}
	}

	return "", false
}

func markGE(mark, min string) bool {
	mark = strings.ToUpper(strings.TrimSpace(mark))
	min = strings.ToUpper(strings.TrimSpace(min))
	if mark == "" || min == "" {
		return false
	}
	mr := []rune(mark)
	nr := []rune(min)
	if len(mr) == 0 || len(nr) == 0 {
		return false
	}
	return mr[0] >= nr[0]
}

func deriveSetCode(setObj map[string]any) string {
	// Newer sets may expose an "abbreviation" field.
	if v, ok := setObj["abbreviation"]; ok {
		switch t := v.(type) {
		case string:
			return strings.ToUpper(strings.TrimSpace(t))
		case map[string]any:
			for _, key := range []string{"official", "en", "intl", "international", "tcg"} {
				if s, ok := t[key].(string); ok {
					s = strings.TrimSpace(s)
					if s != "" {
						return strings.ToUpper(s)
					}
				}
			}
		}
	}

	// Docs mention "tcgOnline" for sets; this often matches printed set codes.
	if v, ok := setObj["tcgOnline"]; ok {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				return strings.ToUpper(s)
			}
		}
	}

	// Defensive fallback keys.
	for _, k := range []string{"ptcgoCode", "ptcgo", "code"} {
		if v, ok := setObj[k]; ok {
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					return strings.ToUpper(s)
				}
			}
		}
	}

	return ""
}

func sortCardsByLocalIDThenName(cards []map[string]any) {
	sort.Slice(cards, func(i, j int) bool {
		li := strings.TrimSpace(toStringish(cards[i]["localId"]))
		lj := strings.TrimSpace(toStringish(cards[j]["localId"]))

		// If both are numeric, numeric sort.
		if ai, err := strconv.Atoi(li); err == nil {
			if aj, err := strconv.Atoi(lj); err == nil {
				if ai != aj {
					return ai < aj
				}
			}
		}

		if li != lj {
			return li < lj
		}

		ni, _ := cards[i]["name"].(string)
		nj, _ := cards[j]["name"].(string)
		return ni < nj
	})
}

func toStringish(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		// JSON numbers decode to float64.
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		return ""
	}
}

func fetchBytes(ctx context.Context, client *http.Client, url string, retries int) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("new request %s: %w", url, err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "tcgdex-standard-dump/0.1")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("GET %s: %w", url, err)
			if attempt == retries {
				return nil, lastErr
			}
			time.Sleep(backoffDuration(attempt))
			continue
		}

		// Be defensive: cap max bytes per response.
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 100<<20)) // 100 MiB
		resp.Body.Close()

		if readErr != nil {
			lastErr = fmt.Errorf("read %s: %w", url, readErr)
			if attempt == retries {
				return nil, lastErr
			}
			time.Sleep(backoffDuration(attempt))
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return body, nil
		}

		snippet := truncate(string(body), 240)
		lastErr = fmt.Errorf("GET %s: status %d: %s", url, resp.StatusCode, snippet)

		// Retry only transient statuses.
		if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode <= 599) {
			if attempt == retries {
				return nil, lastErr
			}
			if ra := retryAfter(resp.Header.Get("Retry-After")); ra > 0 {
				time.Sleep(ra)
			} else {
				time.Sleep(backoffDuration(attempt))
			}
			continue
		}

		// Non-retryable.
		return nil, lastErr
	}

	return nil, lastErr
}

func retryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs <= 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}

func backoffDuration(attempt int) time.Duration {
	base := 400 * time.Millisecond
	max := 10 * time.Second

	shift := attempt
	if shift > 6 {
		shift = 6
	}
	d := base * time.Duration(1<<shift)
	if d > max {
		d = max
	}

	// jitter
	d += time.Duration(rand.Intn(250)) * time.Millisecond
	return d
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func safeFilename(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	return safeFilenameRe.ReplaceAllString(s, "_")
}

func shallowCopyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func writeJSONAtomic(path string, v any, overwrite bool) error {
	tmp := path + ".tmp"

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

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
	if err := f.Sync(); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if overwrite {
		if _, err := os.Stat(path); err == nil {
			if rmErr := os.Remove(path); rmErr != nil {
				_ = os.Remove(tmp)
				return rmErr
			}
		}
	}

	return os.Rename(tmp, path)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
