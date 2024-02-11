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

	response := data.NewResponse()

	response.StatusCode = http.StatusCreated
	response.Result = movie
	response.Message = "Movie Created Successfully"

	err = app.writeJSON(res, response.StatusCode, response, headers)

	if err != nil {
		app.serverErrorResponse(res, req, err)
		return
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
			return
		}
		app.serverErrorResponse(res, req, err)
		return
	}

	response := data.NewResponse()
	response.Result = movie
	response.Message = "Movie Retrieved Successfully"

	err = app.writeJSON(res, 200, response, nil)

	if err != nil {
		app.serverErrorResponse(res, req, err)
	}
}

func (app *application) updateMovieHandler(res http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil {
		app.notFoundResponse(res, req)
		return
	}

	type UpdateMovieDTO struct {
		Title     string   `json:"title"`
		Year      int32    `json:"year"`
		Runtime   int32    `json:"runtime"`
		Genres    []string `json:"genres"`
		Director  string   `json:"director"`
		Actors    []string `json:"actors"`
		Plot      string   `json:"plot"`
		PosterURL string   `json:"poster_url"`
	}

	body := new(UpdateMovieDTO)

	err = app.readJSON(res, req, &body)
	if err != nil {
		app.errorResponse(res, req, http.StatusBadRequest, err.Error())
	}

	movie, err := app.models.Movie.Get(id)

	if err != nil {
		if errors.Is(err, data.ErrNoRecordsFound) {
			app.notFoundResponse(res, req)
			return
		}
		app.serverErrorResponse(res, req, err)
		return
	}

	movie = &data.Movie{
		Title:     body.Title,
		Year:      body.Year,
		Runtime:   body.Runtime,
		Genres:    body.Genres,
		Director:  body.Director,
		Actors:    body.Actors,
		Plot:      body.Plot,
		PosterURL: body.PosterURL,
		ID:        id,
	}

	validate := validator.New()

	if data.ValidateMovie(validate, movie); !validate.Valid() {
		app.failedValidationResponse(res, req, validate.Errors)
	}

	err = app.models.Movie.Update(movie)
	if err != nil {
		app.serverErrorResponse(res, req, err)
		return
	}

	response := data.NewResponse()
	response.Result = movie
	response.Message = "Movie Updated Successfully"

	err = app.writeJSON(res, 200, response, nil)

	if err != nil {
		app.serverErrorResponse(res, req, err)
		return
	}
}

func (app *application) deleteMovieHandler(res http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil {
		app.serverErrorResponse(res, req, err)
		return
	}

	err = app.models.Movie.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNoRecordsFound) {
			app.notFoundResponse(res, req)
			return
		}
		app.serverErrorResponse(res, req, err)
		return
	}

	response := data.NewResponse()
	response.Result = envelope{"id": id}
	response.Message = fmt.Sprintf("Movie With The Following ID %d has Been Deleted", id)

	err = app.writeJSON(res, 200, response, nil)
	if err != nil {
		app.serverErrorResponse(res, req, err)
		return
	}
}
