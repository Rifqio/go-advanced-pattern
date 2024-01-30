package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthCheckHandler(res http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(res, "Status: OK")
	fmt.Fprintf(res, "Environment: %s\n", app.config.env)
	fmt.Fprintf(res, "Version: %s", version)
}
