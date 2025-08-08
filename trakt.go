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
	centralTime, err := time.LoadLocation("America/Chicago")
	if err != nil {
		log.Printf("Error loading timezone, falling back to UTC: %v", err)
		centralTime = time.UTC
	}

	year := time.Now().Year()
	t := time.Date(year, month, 1, 0, 0, 0, 0, centralTime)
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func getCurrentTime() string {
	centralTime, err := time.LoadLocation("America/Chicago")
	if err != nil {
		log.Printf("Error loading timezone, falling back to UTC: %v", err)
		centralTime = time.UTC
	}

	// Get current time, then convert to UTC for API
	now := time.Now().In(centralTime)
	// Set to end of today in CST
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, centralTime)
	return endOfDay.UTC().Format("2006-01-02T15:04:05.000Z")
}

func GetMovieHistory() WatchedHistory {
	jan1 := getMonthStart(time.January)
	endTime := getCurrentTime()

	req, err := http.NewRequest("GET", traktURL+"/users/cfoster5/history/movies", nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}
	req.Header.Add("trakt-api-key", os.Getenv("TRAKT_KEY"))
	req.Header.Add("trakt-api-version", "2")

	params := url.Values{}
	params.Add("start_at", jan1)
	params.Add("end_at", endTime)
	params.Add("limit", "100")
	req.URL.RawQuery = params.Encode()

	log.Printf("Fetching Trakt history from %s to %s", jan1, endTime)

	res, err := HttpClient.Do(req)
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
