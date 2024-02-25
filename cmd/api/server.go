package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         app.config.port,
		Handler:      app.routes(),
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		// Create a quit channel which carries os.Signal value
		quit := make(chan os.Signal, 1)

		// To listen an incoming SIGINT and SIGTERM channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// Read the signal from the quit channel.
		// This code will block until signal received
		s := <-quit

		app.logger.PrintInfo("shutting down", map[string]string{"signal": s.String()})
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.PrintInfo("completing background task", map[string]string{
			"addr": srv.Addr,
		})

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.PrintInfo("Starting server on", map[string]string{"addr": app.config.env, "env": srv.Addr})
	err := srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
