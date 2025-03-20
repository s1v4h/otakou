package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Anime struct {
	ID                     uint      `json:"id"`
	MalID                  uint      `json:"mal_id"`
	Type                   string    `json:"type"`
	Status                 string    `json:"status"`
	TitleRomanized         string    `json:"title_romanized"`
	TitleEnglish           string    `json:"title_english"`
	Synonyms               []string  `json:"synonyms"`
	Source                 string    `json:"source"`
	Rating                 string    `json:"rating"`
	Episodes               uint      `json:"episodes"`
	EpisodeDurationMinutes uint      `json:"episode_duration_minutes"`
	Score                  float32   `json:"score"`
	Synopsis               string    `json:"synopsis"`
	Genres                 []string  `json:"genres"`
	Studios                []string  `json:"studios"`
	StartDate              time.Time `json:"start_date"`
	EndDate                time.Time `json:"end_date"`
}

var (
	animes   []Anime
	animeMap map[uint]*Anime
)

func main() {
	file, err := os.ReadFile("animes.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(file, &animes); err != nil {
		panic(err)
	}
	animeMap = make(map[uint]*Anime, len(animes))
	for i := range animes {
		anime := &animes[i]
		animeMap[anime.ID] = anime
	}

	http.HandleFunc("GET /animes", listAnimes)
	http.HandleFunc("GET /animes/{id}", getAnime)

	fmt.Println("running at http://localhost:3000")
	panic(http.ListenAndServe(":3000", jsonMiddleware(http.DefaultServeMux)))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func listAnimes(w http.ResponseWriter, r *http.Request) {
	uq := r.URL.Query()

	limit, _ := strconv.Atoi(uq.Get("limit"))
	if limit <= 0 {
		limit = 100
	} else if limit > 1000 {
		limit = 1000
	}

	offset, _ := strconv.Atoi(uq.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	var typeIn, typeNotIn map[string]bool
	if l := uq["type_in"]; len(l) > 0 {
		typeIn = make(map[string]bool, len(l))
		for _, v := range l {
			typeIn[v] = true
		}
	} else if l = uq["type_not_in"]; len(l) > 0 {
		typeNotIn = make(map[string]bool, len(l))
		for _, v := range l {
			typeNotIn[v] = true
		}
	}

	filteredAnimes := make([]*Anime, 0, limit)
	for i := range animes {
		anime := &animes[i]

		if typeIn != nil && !typeIn[anime.Type] ||
			typeNotIn != nil && typeNotIn[anime.Type] {
			continue
		}

		if offset > 0 {
			offset--
			continue
		}

		filteredAnimes = append(filteredAnimes, anime)
		if len(filteredAnimes) == limit {
			break
		}
	}
	json.NewEncoder(w).Encode(filteredAnimes)
}

func getAnime(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if id <= 0 {
		http.Error(w, "id must be a non-zero uint", http.StatusBadRequest)
		return
	}

	if anime, ok := animeMap[uint(id)]; ok {
		json.NewEncoder(w).Encode(anime)
		return
	}
	http.NotFound(w, r)
}
