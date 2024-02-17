package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
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

func (app *application) rateLimiter(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a goroutine to delete old entries of cached ip once in every minute

	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening
			// while cleanup is taking place
			mu.Lock()

			// Loop through all the clients. If they haven't been seen within the
			// last three minutes, delete the entries
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if app.config.limiter.enabled {
			// Extract the client ip
			ip, _, err := net.SplitHostPort(req.RemoteAddr)

			if err != nil {
				app.internalServerErrorResponse(res, req, err)
				return
			}

			// Lock to prevent race condition or executed concurrently
			mu.Lock()

			// Check to see if the ip address already exist in the map.
			// If it doesn't initialize a new rate limiter and assign that ip
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}

			// Update the last seen of the client
			clients[ip].lastSeen = time.Now()

			// Call the Allow() method on the rate limiter for the current IP.
			// If the request isn't allowed, unlock the mutex and send 429 response
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.limitExceededResponse(res, req)
				return
			}

			// Unlock the mutex to avoid deadlock
			mu.Unlock()
		}
		next.ServeHTTP(res, req)
	})
}

func (app *application) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		app.logger.PrintHTTP(req)
		next.ServeHTTP(res, req)
	})
}
