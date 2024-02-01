package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(res http.ResponseWriter, req *http.Request) {
	data := map[string]string{
		"status":     "available",
		"enviroment": app.config.env,
		"version":    version,
	}
	err := app.writeJSON(res, 200, data, nil)
	if err != nil {
		app.logger.Println(err)
		app.serverErrorResponse(res, req, err)
	}
}
