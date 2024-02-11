package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
)

type envelope map[string]interface{}

func (app *application) readIDParam(req *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(req.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, err
	}
	return id, nil
}

func (app *application) writeJSON(res http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// This loop is for sending multiple res header
	for key, value := range headers {
		res.Header()[key] = value
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	res.Write(js)

	return nil
}

func (app *application) readJSON(res http.ResponseWriter, req *http.Request, dst interface{}) error {
	err := json.NewDecoder(req.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		if errors.As(err, &syntaxError) {
			return fmt.Errorf("Body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		}

		if errors.Is(err, io.ErrUnexpectedEOF) {
			return errors.New("Body contains badly-formed JSON")
		}

		if errors.As(err, &unmarshalTypeError) {
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("Body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		}

		if errors.Is(err, io.EOF) {
			return errors.New("Body must not be empty")
		}

		if errors.As(err, &invalidUnmarshalError) {
			panic(err)
		}
		return err
	}
	return nil
}
