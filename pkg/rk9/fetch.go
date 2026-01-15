package rk9

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	rk9BaseURL      = "https://rk9.gg"
	defaultTimeout  = 20 * time.Second
	maxRetries      = 3
	initialBackoff  = 400 * time.Millisecond
	maxBackoffDelay = 5 * time.Second
)

// fetchRosterPage downloads the HTML for the RK9 roster page for a given
// tournament ID.
func fetchRosterPage(tournamentID string) (string, error) {
	url := fmt.Sprintf("%s/roster/%s", rk9BaseURL, tournamentID)
	return fetchHTMLWithRetry(url)
}

// fetchPairingsPage downloads the HTML for the RK9 pairings page for a given
// tournament ID.
func fetchPairingsPage(tournamentID string) (string, error) {
	url := fmt.Sprintf("%s/pairings/%s", rk9BaseURL, tournamentID)
	return fetchHTMLWithRetry(url)
}

func fetchHTMLWithRetry(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	client := &http.Client{Timeout: defaultTimeout}
	return fetchHTMLWithRetryCtx(ctx, client, url, maxRetries)
}

func fetchHTMLWithRetryCtx(ctx context.Context, client *http.Client, url string, retries int) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		body, err := fetchHTMLOnce(ctx, client, url)
		if err == nil {
			return body, nil
		}

		lastErr = err
		// Only retry on transient errors.
		if !isTransientError(err) || attempt == retries {
			break
		}
		time.Sleep(backoffDuration(attempt))
	}

	return "", lastErr
}

func fetchHTMLOnce(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "professors-research/1.0 (tournament scraper)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("GET %s: status %d: %s", url, resp.StatusCode, string(b))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(b), nil
}

func isTransientError(err error) bool {
	// For now, be conservative and only retry on HTTP 5xx and network-level
	// failures where we can not determine a status code.
	// Callers can pass in richer error types in the future if needed.
	return true
}

func backoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		return initialBackoff
	}
	d := initialBackoff * time.Duration(1<<attempt)
	if d > maxBackoffDelay {
		d = maxBackoffDelay
	}
	return d
}

func timeoutFromOptionsSeconds(sec int) time.Duration {
	if sec <= 0 {
		return defaultTimeout
	}
	return time.Duration(sec) * time.Second
}



