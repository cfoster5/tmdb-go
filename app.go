package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// Define structs that match the TMDB API response structure
type Movie struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

type TMDBResponse struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

const baseURL = "https://api.themoviedb.org/3"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	myUrl := baseURL + "/discover/movie"
	req, err := http.NewRequest("GET", myUrl, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("TMDB_KEY"))

	params := url.Values{}
	params.Add("include_adult", "false")
	params.Add("include_video", "false")
	params.Add("language", "en-US")
	params.Add("page", "1")
	params.Add("sort_by", "popularity.desc")
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer res.Body.Close()

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Parse JSON into our struct
	var response TMDBResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Now you can loop through each movie
	fmt.Printf("Found %d movies:\n\n", len(response.Results))

	for i, movie := range response.Results {
		fmt.Printf("Movie %d:\n", i+1)
		fmt.Printf("  Title: %s\n", movie.Title)
		fmt.Printf("  Release Date: %s\n", movie.ReleaseDate)
		fmt.Printf("  Rating: %.1f/10\n", movie.VoteAverage)
		fmt.Printf("  Popularity: %.1f\n", movie.Popularity)
		fmt.Printf("  Overview: %.100s...\n", movie.Overview) // First 100 chars
		fmt.Println("  ---")
	}
}
