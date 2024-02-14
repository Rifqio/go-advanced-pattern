package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// If there was a panic, set "Connection
				res.Header().Set("Connection", "close")
				// use fmt.Errorf to normalize the error
				app.internalServerErrorResponse(res, req, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(res, req)
	})
}

func (app *application) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		app.logger.PrintHTTP(req)
		next.ServeHTTP(res, req)
	})
}
