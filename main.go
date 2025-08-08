package main

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

type PosterPaths []string

func getPosterPaths(ctx context.Context) PosterPaths {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	history := GetMovieHistory()

	var posterPaths PosterPaths

	for i, movie := range history {
		fmt.Printf("Processing movie %d: %s\n", i+1, movie.Movie.Title)
		if movie.Movie.IDs.Tmdb == 0 {
			fmt.Printf("  ⚠️  No TMDB ID available for %s\n", movie.Movie.Title)
			continue
		}
		details := GetMovieDetails(int(movie.Movie.IDs.Tmdb))
		fmt.Println(details.PosterPath)
		posterPaths = append(posterPaths, ImageBaseUrl+FileSize+details.PosterPath)
	}

	slices.Reverse(posterPaths)
	return posterPaths
}

func main() {
	lambda.Start(getPosterPaths)
}
