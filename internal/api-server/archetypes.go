package apiserver

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
)

func (s *Server) handleArchetypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.archetypes == nil {
		s.sendError(w, "Archetype definitions not loaded", http.StatusServiceUnavailable)
		return
	}

	names := make([]string, 0, len(s.archetypes))
	for _, arch := range s.archetypes {
		if arch.Name == "" {
			continue
		}
		names = append(names, arch.Name)
	}
	sort.Strings(names)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(names); err != nil {
		log.Printf("Failed to encode archetypes response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

