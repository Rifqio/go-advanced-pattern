package main

import (
	"fmt"
	"net/http"
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

	fmt.Fprintf(res, "Show the details of movie %d\n", id)
}
