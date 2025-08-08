package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type MovieDetails struct {
	Adult               bool                 `json:"adult"`
	BackdropPath        string               `json:"backdrop_path"`
	BelongsToCollection *BelongsToCollection `json:"belongs_to_collection"` // Optional, can be null
	Budget              int64                `json:"budget"`
	Genres              []Genre              `json:"genres"`
	Homepage            string               `json:"homepage"`
	ID                  int64                `json:"id"`
	ImdbID              string               `json:"imdb_id"`
	OriginCountry       []string             `json:"origin_country"`
	OriginalLanguage    string               `json:"original_language"`
	OriginalTitle       string               `json:"original_title"`
	Overview            string               `json:"overview"`
	Popularity          float64              `json:"popularity"`
	PosterPath          string               `json:"poster_path"`
	ProductionCompanies []ProductionCompany  `json:"production_companies"`
	ProductionCountries []ProductionCountry  `json:"production_countries"`
	ReleaseDate         string               `json:"release_date"`
	Revenue             int64                `json:"revenue"`
	Runtime             int64                `json:"runtime"`
	SpokenLanguages     []SpokenLanguage     `json:"spoken_languages"`
	Status              string               `json:"status"`
	Tagline             string               `json:"tagline"`
	Title               string               `json:"title"`
	Video               bool                 `json:"video"`
	VoteAverage         float64              `json:"vote_average"`
	VoteCount           int64                `json:"vote_count"`
}

type BelongsToCollection struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

type Genre struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProductionCompany struct {
	ID            int64  `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

type ProductionCountry struct {
	ISO3166_1 string `json:"iso_3166_1"`
	Name      string `json:"name"`
}

type SpokenLanguage struct {
	EnglishName string `json:"english_name"`
	ISO639_1    string `json:"iso_639_1"`
	Name        string `json:"name"`
}

const apiUrl = "https://api.themoviedb.org/3"
const ImageBaseUrl = "https://image.tmdb.org/t/p/"
const FileSize = "w780"

func GetMovieDetails(movieId int) MovieDetails {

	url := apiUrl + "/movie/" + strconv.Itoa(movieId)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("TMDB_KEY"))

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
	var response MovieDetails
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
	}
	return response
}
