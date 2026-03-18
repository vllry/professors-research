package apiserver

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
)

// TournamentResponse is the JSON representation of a single tournament.
type TournamentResponse struct {
	ID       string `json:"id"`
	Year     int    `json:"year"`
	Location string `json:"location"`
}

func (s *Server) handleTournaments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tournaments := make([]TournamentResponse, 0, len(s.tournaments))
	for id, info := range s.tournaments {
		tournaments = append(tournaments, TournamentResponse{
			ID:       id,
			Year:     info.Year,
			Location: info.Location,
		})
	}
	sort.Slice(tournaments, func(i, j int) bool {
		return tournaments[i].ID < tournaments[j].ID
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tournaments); err != nil {
		log.Printf("Failed to encode tournaments response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
