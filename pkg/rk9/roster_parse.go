package rk9

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/vllry/professors-research/pkg/types"
)

type rosterDecklistRef struct {
	CanonicalPlayer string
	FirstName       string
	LastName        string
	Country         string
	MaskedID        string

	DecklistURL string // absolute URL to public decklist page
}

func extractTournamentNameFromHTML(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}
	return extractTournamentName(doc)
}

// parseRosterDecklistRefs reads the roster HTML and extracts public decklist links
// alongside the player identity info available on the roster table.
func parseRosterDecklistRefs(html string, tournamentID string) ([]rosterDecklistRef, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, errors.Wrap(err, "parse HTML")
	}

	var out []rosterDecklistRef
	seen := make(map[string]bool)

	// Find rows that contain a public decklist link for this tournament.
	sel := fmt.Sprintf("a[href^='/decklist/public/%s/']", tournamentID)
	doc.Find(sel).Each(func(i int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}
		deckURL := rk9BaseURL + href

		tr := a.Closest("tr")
		if tr.Length() == 0 {
			return
		}

		tds := tr.Find("td")
		if tds.Length() < 4 {
			return
		}

		// Observed roster columns:
		// 0: masked ID, 1: first name, 2: last name (often inside nested tags), 3: country, ...
		masked := strings.TrimSpace(tds.Eq(0).Text())
		first := strings.TrimSpace(tds.Eq(1).Text())
		last := strings.TrimSpace(tds.Eq(2).Text())
		country := strings.TrimSpace(tds.Eq(3).Text())

		first = collapseWS(first)
		last = collapseWS(last)
		country = strings.ToUpper(collapseWS(country))

		if first == "" && last == "" {
			return
		}

		canonical := collapseWS(strings.TrimSpace(strings.Join([]string{first, last}, " ")))
		if country != "" {
			canonical = fmt.Sprintf("%s [%s]", canonical, country)
		}
		if masked != "" {
			canonical = fmt.Sprintf("%s %s", masked, canonical)
			canonical = collapseWS(canonical)
		}

		key := canonical + "|" + deckURL
		if seen[key] {
			return
		}
		seen[key] = true

		out = append(out, rosterDecklistRef{
			CanonicalPlayer: canonical,
			FirstName:       first,
			LastName:        last,
			Country:         country,
			MaskedID:        masked,
			DecklistURL:     deckURL,
		})
	})

	return out, nil
}

// parsePublicDecklistHTML parses a public decklist page and builds a types.Decklist.
// It prefers the English translation block (lang-EN), since that is most compatible
// with our card name parsing conventions.
func parsePublicDecklistHTML(html string) (types.Decklist, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return types.Decklist{}, errors.Wrap(err, "parse HTML")
	}

	// Prefer EN list items with structured attributes.
	var cards = make(map[types.Card]int)
	doc.Find(".translation.lang-EN li[data-setnum][data-quantity]").Each(func(i int, li *goquery.Selection) {
		setnum, _ := li.Attr("data-setnum")     // e.g. "TWM-141"
		qtyStr, _ := li.Attr("data-quantity")   // e.g. "2"
		nameAttr, _ := li.Attr("data-cardname") // e.g. "Raging Bolt ex"

		setCode, number := splitSetNum(setnum)
		if setCode == "" || number == "" {
			return
		}
		qty, err := strconv.Atoi(strings.TrimSpace(qtyStr))
		if err != nil || qty <= 0 {
			return
		}

		name := strings.TrimSpace(nameAttr)
		if name == "" {
			// Fallback to visible text.
			name = strings.TrimSpace(li.Clone().Children().Remove().End().Text())
			name = collapseWS(name)
		}
		if name == "" {
			return
		}

		cards[types.Card{
			SetCode: setCode,
			Number:  number,
			Name:    name,
		}] += qty
	})

	// Fallback if EN translation isn't present: accept any language block.
	if len(cards) == 0 {
		doc.Find("li[data-setnum][data-quantity]").Each(func(i int, li *goquery.Selection) {
			setnum, _ := li.Attr("data-setnum")
			qtyStr, _ := li.Attr("data-quantity")
			nameAttr, _ := li.Attr("data-cardname")

			setCode, number := splitSetNum(setnum)
			if setCode == "" || number == "" {
				return
			}
			qty, err := strconv.Atoi(strings.TrimSpace(qtyStr))
			if err != nil || qty <= 0 {
				return
			}

			name := strings.TrimSpace(nameAttr)
			if name == "" {
				name = collapseWS(strings.TrimSpace(li.Text()))
			}
			if name == "" {
				return
			}
			cards[types.Card{SetCode: setCode, Number: number, Name: name}] += qty
		})
	}

	if len(cards) == 0 {
		return types.Decklist{}, fmt.Errorf("no cards found on decklist page")
	}

	dl := types.Decklist{Cards: cards}
	// Validate deck size (should be 60). Some decklists might be missing; still return error.
	total := 0
	for _, n := range dl.Cards {
		total += n
	}
	if total != 60 {
		return types.Decklist{}, fmt.Errorf("decklist contains %d cards (expected 60)", total)
	}
	return dl, nil
}

func splitSetNum(setnum string) (setCode, number string) {
	setnum = strings.TrimSpace(setnum)
	// Common format: "TWM-141" or "SVI-001"
	parts := strings.Split(setnum, "-")
	if len(parts) < 2 {
		return "", ""
	}
	setCode = strings.TrimSpace(parts[0])
	number = strings.TrimSpace(parts[1])
	if setCode == "" || number == "" {
		return "", ""
	}
	return setCode, number
}

func collapseWS(s string) string {
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

// buildAliasToCanonical builds a lookup from pairings-style short names to the roster
// canonical name so matches can be reconciled to decklists.
func buildAliasToCanonical(refs []rosterDecklistRef) map[string]string {
	alias := make(map[string]string)
	for _, r := range refs {
		canon := r.CanonicalPlayer
		if canon == "" || r.FirstName == "" {
			continue
		}
		// Canonical itself
		alias[canon] = canon

		lastInitial := ""
		if r.LastName != "" {
			lastInitial = strings.ToUpper(string([]rune(strings.TrimSpace(r.LastName))[0]))
		}
		if lastInitial != "" {
			// Pairings uses "First L."
			short := fmt.Sprintf("%s %s.", r.FirstName, lastInitial)
			short = collapseWS(short)
			alias[short] = canon
		}
		// Also map without masked id / country suffix, to catch different cleanups.
		noMasked := canon
		if r.MaskedID != "" {
			noMasked = strings.TrimSpace(strings.TrimPrefix(noMasked, r.MaskedID))
		}
		noMasked = strings.TrimSpace(noMasked)
		alias[noMasked] = canon

		// If country is present, map without it too.
		noCountry := strings.TrimSpace(regexp.MustCompile(`\s*\[[A-Z]{2,3}\]\s*$`).ReplaceAllString(noMasked, ""))
		if noCountry != "" {
			alias[noCountry] = canon
		}
	}
	return alias
}

// parsePairingsHxGets extracts all hx-get endpoints for round/pod fragments.
func parsePairingsHxGets(pairingsHTML string, tournamentID string) ([]pairingsFragmentRef, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(pairingsHTML))
	if err != nil {
		return nil, errors.Wrap(err, "parse HTML")
	}
	var out []pairingsFragmentRef
	seen := make(map[string]bool)

	doc.Find("[hx-get]").Each(func(i int, s *goquery.Selection) {
		hx, ok := s.Attr("hx-get")
		if !ok {
			return
		}
		hx = strings.TrimSpace(hx)
		if !strings.HasPrefix(hx, "/pairings/"+tournamentID) {
			return
		}
		u, err := url.Parse(hx)
		if err != nil {
			return
		}
		q := u.Query()
		pod, _ := strconv.Atoi(q.Get("pod"))
		rnd, _ := strconv.Atoi(q.Get("rnd"))
		if pod < 0 || rnd <= 0 {
			return
		}
		full := rk9BaseURL + hx
		if seen[full] {
			return
		}
		seen[full] = true
		out = append(out, pairingsFragmentRef{Pod: pod, Round: rnd, URL: full})
	})

	return out, nil
}




