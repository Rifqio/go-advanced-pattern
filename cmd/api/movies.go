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
		app.internalServerErrorResponse(res, req, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	response := data.NewResponse()

	response = data.Response{
		StatusCode: http.StatusCreated,
		Result:     movie,
		Message:    "Movie Created Successfully",
	}

	err = app.writeJSON(res, response.StatusCode, response, headers)

	if err != nil {
		app.internalServerErrorResponse(res, req, err)
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
		app.internalServerErrorResponse(res, req, err)
		return
	}

	response := data.NewResponse()
	response = data.Response{
		Result:  movie,
		Message: "Movie Retrieved Successfully",
	}

	err = app.writeJSON(res, 200, response, nil)

	if err != nil {
		app.internalServerErrorResponse(res, req, err)
	}
}

func (app *application) showMoviesHandler(res http.ResponseWriter, req *http.Request) {
	var requestQuery struct {
		Title  string   `json:"title"`
		Genres []string `json:"genres"`
		data.Filters
	}

	validate := validator.New()

	// Call req.url.Query() to get the url.Values map containing query string data
	qs := req.URL.Query()

	requestQuery.Title = app.readString(qs, "title", "")
	requestQuery.Genres = app.readCSV(qs, "genres", []string{})

	requestQuery.Filters.Page = app.readInt(qs, "page", 1, validate)
	requestQuery.Filters.PageSize = app.readInt(qs, "page_size", 10, validate)

	requestQuery.Sort = app.readString(qs, "sort", "id")
	requestQuery.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-runtime"}

	if data.ValidateFilters(validate, &requestQuery.Filters); !validate.Valid() {
		app.failedValidationResponse(res, req, validate.Errors)
		return
	}

	movies, paginationMetadata, err := app.models.Movie.GetAll(requestQuery.Title, requestQuery.Genres, requestQuery.Filters)

	if err != nil {
		app.internalServerErrorResponse(res, req, err)
		return
	}

	response := data.NewResponse()
	response = data.Response{
		Result:     movies,
		Message:    "Movies Fetched Successfully",
		Pagination: &paginationMetadata,
	}

	err = app.writeJSON(res, 200, response, nil)

	if err != nil {
		app.internalServerErrorResponse(res, req, err)
		return
	}
}

func (app *application) updateMovieHandler(res http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil {
		app.notFoundResponse(res, req)
		return
	}

	// Changing the struct to pointer is to ignore the nil values
	// Struct don't need to change to pointer since the zero-values of struct is nil
	type UpdateMovieDTO struct {
		Title     *string  `json:"title"`
		Year      *int32   `json:"year"`
		Runtime   *int32   `json:"runtime"`
		Genres    []string `json:"genres"`
		Director  *string  `json:"director"`
		Actors    []string `json:"actors"`
		Plot      *string  `json:"plot"`
		PosterURL *string  `json:"poster_url"`
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
		app.internalServerErrorResponse(res, req, err)
		return
	}

	if body.Title != nil {
		movie.Title = *body.Title
	}

	if body.Year != nil {
		movie.Year = *body.Year
	}

	if body.Runtime != nil {
		movie.Runtime = *body.Runtime
	}

	if body.Genres != nil {
		movie.Genres = body.Genres
	}

	if body.PosterURL != nil {
		movie.PosterURL = *body.PosterURL
	}

	//movie = &data.Movie{
	//	Title:     *body.Title,
	//	Year:      *body.Year,
	//	Runtime:   *body.Runtime,
	//	Genres:    body.Genres,
	//	Director:  *body.Director,
	//	Actors:    body.Actors,
	//	Plot:      *body.Plot,
	//	PosterURL: *body.PosterURL,
	//	ID:        id,
	//}

	validate := validator.New()

	if data.ValidateMovie(validate, movie); !validate.Valid() {
		app.failedValidationResponse(res, req, validate.Errors)
	}

	err = app.models.Movie.Update(movie)
	if err != nil {
		if errors.Is(err, data.ErrEditConflict) {
			app.editConflictResponse(res, req)
			return
		}
		app.internalServerErrorResponse(res, req, err)
		return
	}

	response := data.NewResponse()
	response.Result = movie
	response.Message = "Movie Updated Successfully"

	err = app.writeJSON(res, 200, response, nil)

	if err != nil {
		app.internalServerErrorResponse(res, req, err)
		return
	}
}

func (app *application) deleteMovieHandler(res http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil {
		app.internalServerErrorResponse(res, req, err)
		return
	}

	err = app.models.Movie.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNoRecordsFound) {
			app.notFoundResponse(res, req)
			return
		}
		app.internalServerErrorResponse(res, req, err)
		return
	}

	response := data.NewResponse()
	response.Result = envelope{"id": id}
	response.Message = fmt.Sprintf("Movie With The Following ID %d has Been Deleted", id)

	err = app.writeJSON(res, 200, response, nil)
	if err != nil {
		app.internalServerErrorResponse(res, req, err)
		return
	}
}
