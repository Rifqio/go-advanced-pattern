package data

import (
	"api.go-rifqio.my.id/internal/validator"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Year      int32     `json:"year"`
	Runtime   int32     `json:"runtime"`
	Genres    []string  `json:"genres"`
	Director  string    `json:"director"`
	Actor     []string  `json:"actor"`
	Plot      string    `json:"plot"`
	PosterURL string    `json:"poster_url"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"-"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "title must be provided")
	v.Check(len(movie.Title) <= 100, "title", "title max length is 100 characters")

	v.Check(movie.Year != 0, "year", "year must be provided")
	v.Check(movie.Year >= 0, "year", "year is invalid")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "year cannot be in the future")
}
