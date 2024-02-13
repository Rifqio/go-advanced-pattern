package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(req *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": req.Method,
		"request_url":    req.URL.String(),
	})
}

// Base template for error response
func (app *application) errorResponse(res http.ResponseWriter, req *http.Request, status int, message interface{}) {
	baseResponse := envelope{"status": false, "statusCode": status, "error": message}

	err := app.writeJSON(res, status, baseResponse, nil)
	if err != nil {
		app.logError(req, err)
		res.WriteHeader(500)
	}
}

// Server error response will occur at unexpected error runtime
func (app *application) internalServerErrorResponse(res http.ResponseWriter, req *http.Request, err error) {
	app.logError(req, err)
	message := "The server encountered a problem and cannot process incoming request"
	app.errorResponse(res, req, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(res http.ResponseWriter, req *http.Request) {
	message := "The requested resource could not be found"
	app.errorResponse(res, req, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(res http.ResponseWriter, req *http.Request) {
	message := fmt.Sprintf("The %s method is not allowed for this resource", req.Method)
	app.errorResponse(res, req, http.StatusMethodNotAllowed, message)
}

func (app *application) failedValidationResponse(res http.ResponseWriter, req *http.Request, errors map[string]string) {
	app.errorResponse(res, req, http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictResponse(res http.ResponseWriter, req *http.Request) {
	message := "Unable to update the record due to an edit conflict, please try again"
	app.errorResponse(res, req, http.StatusConflict, message)
}
