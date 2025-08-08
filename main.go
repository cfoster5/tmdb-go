package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var HttpClient = &http.Client{
	Timeout: 30 * time.Second,
}

type PosterPaths []string

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func getPosterPaths(ctx context.Context) (PosterPaths, error) {
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
	return posterPaths, nil
}

func handleHTTP(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	posterPaths, err := getPosterPaths(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		}, nil
	}

	// Convert to proper JSON
	body, err := json.Marshal(map[string]any{
		"poster_paths": posterPaths,
		"count":        len(posterPaths),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error": "Failed to marshal response"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handleHTTP)
}
