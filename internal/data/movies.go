package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Year      int32     `json:"year"`
	Runtime   Runtime   `json:"runtime"`
	Genres    []string  `json:"genres"`
	Director  string    `json:"director"`
	Actor     []string  `json:"actor"`
	Plot      string    `json:"plot"`
	PosterURL string    `json:"poster_url"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"-"`
}
