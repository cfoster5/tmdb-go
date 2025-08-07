package main

import (
	"fmt"
	"image/color"
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

	config := CollageConfig{
		Title:           "Movies Watched in 2025",
		Year:            2025,
		BackgroundStart: color.RGBA{102, 126, 234, 255}, // Purple-blue
		BackgroundEnd:   color.RGBA{118, 75, 162, 255},  // Darker purple
		TextColor:       color.RGBA{255, 255, 255, 255}, // White
	}

	err = GenerateCollageImage(history, config)
	if err != nil {
		log.Printf("Error generating collage: %v", err)
	}
}
