package main

import "net/http"

func (app *application) logError(_ *http.Request, err error) {
	app.logger.Println(err)
}

func (app *application) errorResponse(res http.ResponseWriter, req *http.Request, status int, message interface{}) {
	baseResponse := envelope{"status": false, "error": message, "message": "Internal Server Error"}

	err := app.writeJSON(res, status, baseResponse, nil)
	if err != nil {
		app.logError(req, err)
		res.WriteHeader(500)
	}
}
