package main

import (
	"api.go-rifqio.my.id/internal/data"
	"encoding/json"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(res http.ResponseWriter, req *http.Request) {
	type CreateMovieDTO struct {
		Title     string       `json:"title"`
		Year      int32        `json:"year"`
		Runtime   data.Runtime `json:"runtime"`
		Genres    []string     `json:"genres"`
		Director  string       `json:"director"`
		Actor     []string     `json:"actor"`
		Plot      string       `json:"plot"`
		PosterURL string       `json:"poster_url"`
	}

	body := new(CreateMovieDTO)

	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		app.errorResponse(res, req, http.StatusBadRequest, err.Error())
		return
	}

	err = app.writeJSON(res, 201, data.Response{
		Status:  true,
		Result:  body.Title,
		Message: "Movie Created Successfully",
	}, nil)
	if err != nil {
		app.serverErrorResponse(res, req, err)
	}
}

func (app *application) showMovieHandler(res http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil {
		app.notFoundResponse(res, req)
		return
	}

	movie := data.Movie{
		ID:        id,
		Title:     "The Matrix",
		Year:      1999,
		Runtime:   136,
		Genres:    []string{"Action", "Sci-Fi"},
		Director:  "Lana Wachowski",
		Actor:     []string{"Keanu Reeves", "Laurence Fishburne", "Carrie-Anne Moss", "Hugo Weaving"},
		Plot:      "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
		PosterURL: "https://images-na.ssl-images-amazon.com/images/I/51EG732BV3L._AC_.jpg",
		CreatedAt: time.Now(),
		Version:   1,
	}

	err = app.writeJSON(res, 200, data.Response{
		Status:  true,
		Result:  movie,
		Message: "Movie Retrieved Successfully",
	}, nil)

	if err != nil {
		app.serverErrorResponse(res, req, err)
	}
}
