package apiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type tournamentInfo struct {
	Year     int
	Location string
}

type tournamentFile struct {
	TournamentID   string `json:"tournamentId"`
	TournamentName string `json:"tournamentName"`
}

var yearRe = regexp.MustCompile(`\b(20\d{2})\b`)

func parseTournamentName(name string) tournamentInfo {
	name = strings.TrimSpace(name)
	if name == "" {
		return tournamentInfo{}
	}

	info := tournamentInfo{}

	m := yearRe.FindStringSubmatch(name)
	if len(m) >= 2 {
		if y, err := strconv.Atoi(m[1]); err == nil {
			info.Year = y
		}
	}

	if info.Year == 0 {
		// Can't reliably infer location without a year anchor.
		info.Location = name
		return info
	}

	parts := strings.SplitN(name, strconv.Itoa(info.Year), 2)
	if len(parts) != 2 {
		info.Location = name
		return info
	}
	afterYear := strings.TrimSpace(parts[1])
	if afterYear == "" {
		info.Location = name
		return info
	}

	// Heuristic: location is the text after the year and before "Pokémon"/"Pokemon"/"TCG".
	stopWords := []string{"Pokémon", "Pokemon", "TCG"}
	stop := len(afterYear)
	for _, w := range stopWords {
		if idx := strings.Index(afterYear, w); idx != -1 && idx < stop {
			stop = idx
		}
	}
	loc := strings.TrimSpace(afterYear[:stop])
	loc = strings.Trim(loc, " -–—,")
	if loc == "" {
		loc = name
	}

	info.Location = loc
	return info
}

func loadTournamentRegistry(dataDir string) map[string]tournamentInfo {
	tournamentsDir := filepath.Join(dataDir, "tournaments")
	entries, err := os.ReadDir(tournamentsDir)
	if err != nil {
		// Keep the server usable even without local tournament data.
		return map[string]tournamentInfo{}
	}

	registry := make(map[string]tournamentInfo, len(entries))
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}

		id := ent.Name()
		info := tournamentInfo{}

		tPath := filepath.Join(tournamentsDir, id, "tournament.json")
		if data, err := os.ReadFile(tPath); err == nil {
			var tf tournamentFile
			if err := json.Unmarshal(data, &tf); err == nil {
				// If the file claims a different ID, skip it to avoid exposing confusing/unsafe paths.
				if tf.TournamentID != "" && tf.TournamentID != id {
					log.Printf("Warning: %s tournamentId=%q does not match dir %q; skipping", tPath, tf.TournamentID, id)
					continue
				}
				info = parseTournamentName(tf.TournamentName)
			}
		}

		if info.Location == "" {
			info.Location = id
		}

		registry[id] = info
	}

	return registry
}

func (s *Server) validateAndResolveTournamentDirs(rawIDs []string) ([]string, []string, error) {
	if len(rawIDs) == 0 {
		return nil, nil, fmt.Errorf("tournamentIds must contain at least 1 tournament")
	}

	ids := make([]string, 0, len(rawIDs))
	dirs := make([]string, 0, len(rawIDs))

	for _, raw := range rawIDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			return nil, nil, fmt.Errorf("tournamentIds must not contain empty tournament IDs")
		}
		if _, ok := s.tournaments[id]; !ok {
			return nil, nil, fmt.Errorf("unknown tournamentId: %s", id)
		}

		ids = append(ids, id)
		dirs = append(dirs, filepath.Join(s.dataDir, "tournaments", id))
	}

	return ids, dirs, nil
}
