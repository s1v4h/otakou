package main

import (
	"fmt"
	"net/http"
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

func main() {
	fmt.Println("running at http://localhost:3000")
	panic(http.ListenAndServe(":3000", nil))
}
