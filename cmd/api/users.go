package main

import (
	"api.go-rifqio.my.id/internal/data"
	"api.go-rifqio.my.id/internal/validator"
	"errors"
	"net/http"
	"time"
)

func (app *application) registerUserHandler(res http.ResponseWriter, req *http.Request) {
	type CreateUserDTO struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	body := new(CreateUserDTO)

	err := app.readJSON(res, req, &body)
	if err != nil {
		app.internalServerErrorResponse(res, req, err)
		return
	}

	user := &data.User{
		Name:      body.Name,
		Email:     body.Email,
		Activated: false,
	}

	err = user.Password.Set(body.Password)
	if err != nil {
		app.logger.PrintInfo(body.Password, nil)
		app.internalServerErrorResponse(res, req, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(res, req, v.Errors)
		return
	}

	err = app.models.User.Insert(user)
	if err != nil {
		if errors.Is(err, data.ErrDuplicateEmail) {
			v.AddError("email", "user with this email already exist")
			app.failedValidationResponse(res, req, v.Errors)
			return
		}
		app.internalServerErrorResponse(res, req, err)
		return
	}

	// After user is generated in the database, generate a new activation token
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivations)
	if err != nil {
		app.internalServerErrorResponse(res, req, err)
		return
	}

	// Launch a background go routine to send email
	app.background(func() {
		dataEmail := map[string]interface{}{
			"activationToken": token.PlainText,
			"userID":          user.ID,
		}
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", dataEmail)
		if err != nil {
			app.internalServerErrorResponse(res, req, err)
			return
		}
	})

	response := data.NewResponse()
	response.StatusCode = 201
	response.Result = user
	response.Message = "User Created Successfully"

	err = app.writeJSON(res, 201, response, nil)
	if err != nil {
		app.internalServerErrorResponse(res, req, err)
		return
	}
}
