package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type AnimeType uint

const (
	_ AnimeType = iota
	MOVIE
	ONA
	OVA
	SPECIAL
	TV
)

func parseAnimeType(s string) (AnimeType, error) {
	switch s {
	case "MOVIE":
		return MOVIE, nil
	case "ONA":
		return ONA, nil
	case "OVA":
		return OVA, nil
	case "SPECIAL":
		return SPECIAL, nil
	case "TV":
		return TV, nil
	default:
		return 0, fmt.Errorf("invalid AnimeType: %q", s)
	}
}

type AnimeStatus uint

const (
	_ AnimeStatus = iota
	PLANNED
	AIRING
	FINISHED
)

func parseAnimeStatus(s string) (AnimeStatus, error) {
	switch s {
	case "PLANNED":
		return PLANNED, nil
	case "AIRING":
		return AIRING, nil
	case "FINISHED":
		return FINISHED, nil
	default:
		return 0, fmt.Errorf("invalid AnimeStatus: %q", s)
	}
}

type AnimeSeason uint

const (
	_ AnimeSeason = iota
	SPRING
	SUMMER
	FALL
	WINTER
)

func parseAnimeSeason(s string) (AnimeSeason, error) {
	switch s {
	case "SPRING":
		return SPRING, nil
	case "SUMMER":
		return SUMMER, nil
	case "FALL":
		return FALL, nil
	case "WINTER":
		return WINTER, nil
	default:
		return 0, fmt.Errorf("invalid AnimeSeason: %q", s)
	}
}

type Anime struct {
	ID              uint        `json:"id"`
	MalID           uint        `json:"mal_id"`
	Type            AnimeType   `json:"type"`
	Status          AnimeStatus `json:"status"`
	Title           string      `json:"title"`
	Synonyms        []string    `json:"synonyms"`
	Source          string      `json:"source"`
	Rating          string      `json:"rating"`
	Episodes        uint        `json:"episodes"`
	EpisodeDuration uint        `json:"episode_duration"`
	Score           float32     `json:"score"`
	Synopsis        string      `json:"synopsis"`
	Genres          []string    `json:"genres"`
	Studios         []string    `json:"studios"`
	Season          AnimeSeason `json:"season"`
	Year            uint        `json:"year"`
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

	var typeIn, typeNotIn map[AnimeType]bool
	if l := uq["type_in"]; len(l) > 0 {
		typeIn = make(map[AnimeType]bool, len(l))
		for _, v := range l {
			e, err := parseAnimeType(v)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid value for type_in: %v", err), http.StatusBadRequest)
				return
			}
			typeIn[e] = true
		}
	} else if l = uq["type_not_in"]; len(l) > 0 {
		typeNotIn = make(map[AnimeType]bool, len(l))
		for _, v := range l {
			e, err := parseAnimeType(v)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid value for type_not_in: %v", err), http.StatusBadRequest)
				return
			}
			typeNotIn[e] = true
		}
	}

	var status, statusNot AnimeStatus
	if s := uq.Get("status"); s != "" {
		e, err := parseAnimeStatus(s)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid value for status: %v", err), http.StatusBadRequest)
			return
		}
		status = e
	} else if s = uq.Get("status_not"); s != "" {
		e, err := parseAnimeStatus(s)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid value for status_not: %v", err), http.StatusBadRequest)
			return
		}
		statusNot = e
	}

	minScore, _ := strconv.ParseFloat(uq.Get("min_score"), 32)
	maxScore, _ := strconv.ParseFloat(uq.Get("max_score"), 32)
	if minScore > maxScore {
		http.Error(w, "min_score cannot be greater than max_score", http.StatusBadRequest)
		return
	}

	var genreIn, genreNotIn map[string]bool
	if l := uq["genre_in"]; len(l) > 0 {
		genreIn = make(map[string]bool, len(l))
		for _, v := range l {
			genreIn[v] = true
		}
	}
	if l := uq["genre_not_in"]; len(l) > 0 {
		genreNotIn = make(map[string]bool, len(l))
		for _, v := range l {
			if genreIn[v] {
				http.Error(w, fmt.Sprintf("genre_in and genre_not_in cannot contain the same genre: %q", v), http.StatusBadRequest)
				return
			}
			genreNotIn[v] = true
		}
	}
	allGenres := uq.Get("all_genres") == "true"

	filteredAnimes := make([]*Anime, 0, limit)
	for i := range animes {
		anime := &animes[i]

		if typeIn != nil && !typeIn[anime.Type] ||
			typeNotIn != nil && typeNotIn[anime.Type] {
			continue
		}

		if status > 0 && status != anime.Status ||
			statusNot > 0 && statusNot == anime.Status {
			continue
		}

		if minScore > 0 && float32(minScore) > anime.Score ||
			maxScore > 0 && float32(maxScore) < anime.Score {
			continue
		}

		if genreIn != nil || genreNotIn != nil {
			ok := genreIn == nil
			count := 0
			for _, g := range anime.Genres {
				if genreIn[g] {
					ok = true
					count++
					if (!allGenres || count == len(genreIn)) && genreNotIn == nil {
						break
					}
				}
				if genreNotIn[g] {
					ok = false
					break
				}
			}
			if !ok || allGenres && count != len(genreIn) {
				continue
			}
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
