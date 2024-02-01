package main

import (
	"api.go-rifqio.my.id/internal/data"
	"fmt"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "Create a new movie")
}

func (app *application) showMovieHandler(res http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil {
		http.NotFound(res, req)
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
		app.logger.Println(err)
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
	}
}
