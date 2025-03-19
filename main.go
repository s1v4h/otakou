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

var animes []Anime

func main() {
	file, err := os.ReadFile("animes.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(file, &animes); err != nil {
		panic(err)
	}

	http.HandleFunc("GET /animes", listAnimes)

	fmt.Println("running at http://localhost:3000")
	panic(http.ListenAndServe(":3000", nil))
}

func listAnimes(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	} else if limit > 1000 {
		limit = 1000
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	if offset >= len(animes) {
		w.Write([]byte("[]"))
		return
	}
	end := min(offset+limit, len(animes))
	json.NewEncoder(w).Encode(animes[offset:end])
}
