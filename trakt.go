package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type WatchedHistory []WatchedHistoryElement

type WatchedHistoryElement struct {
	ID        int64      `json:"id"`
	WatchedAt time.Time  `json:"watched_at"`
	Action    Action     `json:"action"`
	Type      Type       `json:"type"`
	Movie     TraktMovie `json:"movie"`
}

type TraktMovie struct {
	Title string `json:"title"`
	Year  int64  `json:"year"`
	IDs   IDs    `json:"ids"`
}

type IDs struct {
	Trakt int64  `json:"trakt"`
	Slug  string `json:"slug"`
	Imdb  string `json:"imdb"`
	Tmdb  int64  `json:"tmdb"`
}

type Action string

const (
	Checkin Action = "checkin"
	Watch   Action = "watch"
)

type Type string

const (
	Movie Type = "movie"
)

const traktURL = "https://api.trakt.tv"

func getMonthStart(month time.Month) string {
	year := time.Now().Year()
	// Construct the first day of the given month at UTC midnight
	t := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	// Return in TMDB/Trakt API format
	return t.Format("2006-01-02T15:04:05.000Z")
}

func GetMovieHistory() WatchedHistory {
	jan1 := getMonthStart(time.January)

	req, err := http.NewRequest("GET", traktURL+"/users/cfoster5/history/movies", nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}
	req.Header.Add("trakt-api-key", os.Getenv("TRAKT_KEY"))
	req.Header.Add("trakt-api-version", "2")

	params := url.Values{}
	params.Add("start_at", jan1)
	params.Add("end_at", "2025-08-02T00:00:00.000Z")
	params.Add("limit", "100")
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
	}

	// Parse JSON into our struct
	var history WatchedHistory
	err = json.Unmarshal(body, &history)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
	}
	return history
}
