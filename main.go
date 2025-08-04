package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	history := GetMovieHistory()

	for i, movie := range history {
		fmt.Printf("Processing movie %d: %s\n", i+1, movie.Movie.Title)
		if movie.Movie.IDs.Tmdb == 0 {
			fmt.Printf("  ⚠️  No TMDB ID available for %s\n", movie.Movie.Title)
			continue
		}
		details := GetMovieDetails(int(movie.Movie.IDs.Tmdb))
		fmt.Println(details.PosterPath)
		GetImage(details.PosterPath, movie.WatchedAt)
	}
}
