package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
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

func (e AnimeType) String() string {
	switch e {
	case MOVIE:
		return "MOVIE"
	case ONA:
		return "ONA"
	case OVA:
		return "OVA"
	case SPECIAL:
		return "SPECIAL"
	case TV:
		return "TV"
	default:
		return ""
	}
}

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

func (e AnimeStatus) String() string {
	switch e {
	case PLANNED:
		return "PLANNED"
	case AIRING:
		return "AIRING"
	case FINISHED:
		return "FINISHED"
	default:
		return ""
	}
}

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

type AnimeRating uint

const (
	_ AnimeRating = iota
	G
	PG12
	R15_PLUS
	R18_PLUS
)

func (e AnimeRating) String() string {
	switch e {
	case G:
		return "G"
	case PG12:
		return "PG12"
	case R15_PLUS:
		return "R15+"
	case R18_PLUS:
		return "R18+"
	default:
		return ""
	}
}

func parseAnimeRating(s string) (AnimeRating, error) {
	switch s {
	case "G":
		return G, nil
	case "PG12":
		return PG12, nil
	case "R15+":
		return R15_PLUS, nil
	case "R18+":
		return R18_PLUS, nil
	default:
		return 0, fmt.Errorf("invalid AnimeRating: %q", s)
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

func (e AnimeSeason) String() string {
	switch e {
	case SPRING:
		return "SPRING"
	case SUMMER:
		return "SUMMER"
	case FALL:
		return "FALL"
	case WINTER:
		return "WINTER"
	default:
		return ""
	}
}

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
	Rating          AnimeRating `json:"rating"`
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

	templates = template.Must(template.ParseGlob("*.html"))
)

func init() {
	file, err := os.ReadFile("animes.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(file, &animes); err != nil {
		panic(err)
	}
	sort.Slice(animes, func(i, j int) bool {
		return animes[i].ID < animes[j].ID
	})
	animeMap = make(map[uint]*Anime, len(animes))
	for i := range animes {
		anime := &animes[i]
		animeMap[anime.ID] = anime
	}
}

func main() {
	api := http.NewServeMux()
	api.HandleFunc("GET /animes", listAnimes)
	api.HandleFunc("GET /animes/{id}", getAnime)

	http.Handle("GET /api/", http.StripPrefix("/api", jsonMiddleware(api)))
	http.Handle("GET /thumbs/", http.StripPrefix("/thumbs", http.FileServer(http.Dir("./thumbs"))))
	http.HandleFunc("GET /{$}", home)

	fmt.Println("running at http://localhost:3000")
	panic(http.ListenAndServe(":3000", nil))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func home(w http.ResponseWriter, r *http.Request) {
	start := max(0, len(animes)-15)
	recentAnimes := animes[start:]
	if err := templates.ExecuteTemplate(w, "home.html", recentAnimes); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func listAnimes(w http.ResponseWriter, r *http.Request) {
	uq := r.URL.Query()

	limit := 100
	if s := uq.Get("limit"); s != "" {
		n, _ := strconv.Atoi(s)
		if n < 1 || n > 1000 {
			http.Error(w, "limit must be an integer between 1 and 1000", http.StatusBadRequest)
			return
		}
		limit = n
	}

	offset := 0
	if s := uq.Get("offset"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			http.Error(w, "offset must be a non-negative integer", http.StatusBadRequest)
			return
		}
		offset = n
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
