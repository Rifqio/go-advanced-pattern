package main

import (
	"api.go-rifqio.my.id/internal/data"
	"api.go-rifqio.my.id/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createMovieHandler(res http.ResponseWriter, req *http.Request) {
	type CreateMovieDTO struct {
		Title     string   `json:"title"`
		Year      int32    `json:"year"`
		Runtime   int32    `json:"runtime"`
		Genres    []string `json:"genres"`
		Director  string   `json:"director"`
		Actors    []string `json:"actors"`
		Plot      string   `json:"plot"`
		PosterURL string   `json:"poster_url"`
	}

	body := new(CreateMovieDTO)

	err := app.readJSON(res, req, &body)
	if err != nil {
		app.errorResponse(res, req, http.StatusBadRequest, err.Error())
		return
	}

	movie := &data.Movie{
		Title:     body.Title,
		Year:      body.Year,
		Runtime:   body.Runtime,
		Genres:    body.Genres,
		Director:  body.Director,
		Actors:    body.Actors,
		Plot:      body.Plot,
		PosterURL: body.PosterURL,
	}

	validate := validator.New()

	data.ValidateMovie(validate, movie)

	if !validate.Valid() {
		app.failedValidationResponse(res, req, validate.Errors)
		return
	}

	err = app.models.Movie.Insert(movie)
	if err != nil {
		app.serverErrorResponse(res, req, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(res, 201, data.Response{
		Status:  true,
		Result:  movie,
		Message: "Movie Created Successfully",
	}, headers)

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

	movie, err := app.models.Movie.Get(id)

	if err != nil {
		if errors.Is(err, data.ErrNoRecordsFound) {
			app.notFoundResponse(res, req)
		}
		app.serverErrorResponse(res, req, err)
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
