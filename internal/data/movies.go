package data

import (
	"api.go-rifqio.my.id/internal/validator"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Year      int32     `json:"year"`
	Runtime   int32     `json:"runtime"`
	Genres    []string  `json:"genres"`
	Director  string    `json:"director"`
	Actors    []string  `json:"actors"`
	Plot      string    `json:"plot"`
	PosterURL string    `json:"poster_url"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"-"`
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// If the receiver is a struct or array, any of whose elements is a pointer to something that may be mutated,
// prefer a pointer receiver to make the intention of mutability clear to the reader.
func (m *MovieModel) Insert(movie *Movie) error {
	query := `insert into movies (title, year, runtime, genres, director, actors, plot, poster_url)
			  values($1, $2, $3, $4, $5, $6, $7, $8)
			  returning id, created_at, version`

	// Use pq.Array to type cast []string to type array in postgres before executing
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.Director,
		pq.Array(movie.Actors),
		movie.Plot,
		movie.PosterURL,
	}

	// ! Important normally we would use DB.Exec() to insert to database but
	// since we are using returning statement above, we have to use DB.QueryRow()

	// Use the QueryRow() method to execute the SQL query on the connection pool,
	// passing in the args slice as a variadic parameter and scanning the system
	// generated id, created_at and version values into the movie struct.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrNoRecordsFound
	}

	query := `select * from movies where id = $1`

	// Declare a Movie struct to hold the data returned by the query.
	var movie Movie
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Director,
		pq.Array(&movie.Actors),
		&movie.Plot,
		&movie.PosterURL,
		&movie.CreatedAt,
		&movie.Version,
	)

	if err != nil {
		if errors.Is(err, ErrNoRecordsFound) {
			return nil, ErrNoRecordsFound
		}
		return nil, err
	}

	return &movie, nil
}

func (m *MovieModel) Update(movie *Movie) error {
	return nil
}

func (m *MovieModel) Delete(id int64) error {
	return nil
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "title must be provided")
	v.Check(len(movie.Title) <= 100, "title", "title max length is 100 characters")

	v.Check(movie.Year != 0, "year", "year must be provided")
	v.Check(movie.Year >= 0, "year", "year is invalid")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "year cannot be in the future")
}
