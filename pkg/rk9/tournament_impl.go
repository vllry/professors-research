package rk9

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/vllry/professors-research/pkg/types"
)

func fetchTournamentData(ctx context.Context, tournamentID string, opts FetchOptions) (TournamentData, TournamentPages, error) {
	if tournamentID == "" {
		return TournamentData{}, TournamentPages{}, fmt.Errorf("tournament ID must not be empty")
	}
	if opts.DecklistConcurrency <= 0 {
		opts.DecklistConcurrency = min(6, runtime.NumCPU()*2)
	}
	if opts.PairingsConcurrency <= 0 {
		opts.PairingsConcurrency = min(8, runtime.NumCPU()*2)
	}

	timeout := timeoutFromOptionsSeconds(opts.HTTPTimeoutSeconds)
	client := &http.Client{Timeout: timeout}

	emit := func(e ProgressEvent) {
		if opts.OnProgress != nil {
			opts.OnProgress(e)
		}
	}

	emit(ProgressEvent{Phase: PhaseFetchRoster, Details: "GET roster"})
	rosterHTML, err := fetchHTMLWithRetryCtx(ctx, client, rk9BaseURL+"/roster/"+tournamentID, maxRetries)
	if err != nil {
		return TournamentData{}, TournamentPages{}, fmt.Errorf("fetch roster page: %w", err)
	}
	emit(ProgressEvent{Phase: PhaseFetchPairings, Details: "GET pairings"})
	pairingsHTML, err := fetchHTMLWithRetryCtx(ctx, client, rk9BaseURL+"/pairings/"+tournamentID, maxRetries)
	if err != nil {
		return TournamentData{}, TournamentPages{}, fmt.Errorf("fetch pairings page: %w", err)
	}

	pages := TournamentPages{RosterHTML: rosterHTML, PairingsHTML: pairingsHTML}

	tournamentName := extractTournamentNameFromHTML(rosterHTML)

	emit(ProgressEvent{Phase: PhaseParseRoster, Details: "parse roster decklist links"})
	rosterRefs, err := parseRosterDecklistRefs(rosterHTML, tournamentID)
	if err != nil {
		return TournamentData{}, pages, fmt.Errorf("parse roster decklist refs: %w", err)
	}

	// Apply max decklists limit (useful for debugging).
	if opts.MaxDecklists > 0 && opts.MaxDecklists < len(rosterRefs) {
		rosterRefs = rosterRefs[:opts.MaxDecklists]
	}

	aliasToCanonical := buildAliasToCanonical(rosterRefs)

	// Fetch decklists concurrently.
	decklists := make(map[string]types.Decklist)
	var dlMu sync.Mutex
	var dlDone int64
	totalDL := len(rosterRefs)

	emit(ProgressEvent{Phase: PhaseFetchDecklists, Done: 0, Total: totalDL, Details: "download decklists"})

	dlJobs := make(chan rosterDecklistRef)
	var dlWg sync.WaitGroup
	for i := 0; i < opts.DecklistConcurrency; i++ {
		dlWg.Add(1)
		go func() {
			defer dlWg.Done()
			for ref := range dlJobs {
				if ref.DecklistURL == "" || ref.CanonicalPlayer == "" {
					atomic.AddInt64(&dlDone, 1)
					continue
				}
				html, err := fetchHTMLWithRetryCtx(ctx, client, ref.DecklistURL, maxRetries)
				if err == nil {
					if dl, err := parsePublicDecklistHTML(html); err == nil {
						dlMu.Lock()
						decklists[ref.CanonicalPlayer] = dl
						dlMu.Unlock()
					}
				}
				done := int(atomic.AddInt64(&dlDone, 1))
				// Emit progress occasionally
				if done == totalDL || done%25 == 0 {
					emit(ProgressEvent{Phase: PhaseFetchDecklists, Done: done, Total: totalDL})
				}
			}
		}()
	}
	go func() {
		defer close(dlJobs)
		for _, ref := range rosterRefs {
			select {
			case <-ctx.Done():
				return
			case dlJobs <- ref:
			}
		}
	}()
	dlWg.Wait()

	// Pairings are loaded via htmx (hx-get). Fetch all round fragments and parse those.
	frags, err := parsePairingsHxGets(pairingsHTML, tournamentID)
	if err != nil {
		return TournamentData{}, pages, fmt.Errorf("parse pairings hx-get: %w", err)
	}
	if opts.MaxPairingsFragments > 0 && opts.MaxPairingsFragments < len(frags) {
		frags = frags[:opts.MaxPairingsFragments]
	}

	var matches []types.MatchResult
	var mMu sync.Mutex
	var fDone int64
	totalF := len(frags)
	emit(ProgressEvent{Phase: PhaseFetchPairingsFrags, Done: 0, Total: totalF, Details: "download pairings fragments"})

	fJobs := make(chan pairingsFragmentRef)
	var fWg sync.WaitGroup
	for i := 0; i < opts.PairingsConcurrency; i++ {
		fWg.Add(1)
		go func() {
			defer fWg.Done()
			for f := range fJobs {
				html, err := fetchHTMLWithRetryCtx(ctx, client, f.URL, maxRetries)
				if err == nil {
					ms := parsePairingsFragmentHTML(html, f.Round, f.Pod, func(name string) string {
						if name == "" {
							return ""
						}
						if canon, ok := aliasToCanonical[name]; ok {
							return canon
						}
						return name
					})
					if len(ms) > 0 {
						mMu.Lock()
						matches = append(matches, ms...)
						mMu.Unlock()
					}
				}
				done := int(atomic.AddInt64(&fDone, 1))
				if done == totalF || done%10 == 0 {
					emit(ProgressEvent{Phase: PhaseFetchPairingsFrags, Done: done, Total: totalF})
				}
			}
		}()
	}
	go func() {
		defer close(fJobs)
		for _, f := range frags {
			select {
			case <-ctx.Done():
				return
			case fJobs <- f:
			}
		}
	}()
	fWg.Wait()

	return TournamentData{
		TournamentID:   tournamentID,
		TournamentName: tournamentName,
		Decklists:      decklists,
		Matches:        matches,
	}, pages, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


