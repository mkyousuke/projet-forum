package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const TMDBAPIKey = "712f38fb8872be0ae97293108b1d160a"

type Movie struct {
	Title      string `json:"title"`
	Overview   string `json:"overview"`
	PosterPath string `json:"poster_path"`
}

type TmdbResponse struct {
	Results []Movie `json:"results"`
}

func TmdbHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handler TmdbHandler appelé")

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/popular?api_key=%s&language=fr-FR", TMDBAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Erreur lors de la requête à TMDb", http.StatusInternalServerError)
		fmt.Println("Erreur requête TMDb:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur lecture du body TMDb", http.StatusInternalServerError)
		fmt.Println("Erreur lecture body:", err)
		return
	}

	var tmdbResp TmdbResponse
	if err := json.Unmarshal(body, &tmdbResp); err != nil {
		http.Error(w, "Erreur lors du décodage JSON TMDb", http.StatusInternalServerError)
		fmt.Println("Erreur JSON:", err)
		return
	}


	renderTemplate(w, r, "API.html", tmdbResp)
}
