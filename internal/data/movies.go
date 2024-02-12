package data

import (
	"api.go-rifqio.my.id/internal/validator"
	"context"
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
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
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrNoRecordsFound
	}

	query := `select * from movies where id = $1`

	// Declare a Movie struct to hold the data returned by the query.
	var movie Movie

	// Set timeout query for 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// The defer means context will always be released before the Get() method returns,
	// thereby preventing a memory leak
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
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
	query := `update movies set title = $1, year = $2, runtime = $3, genres = $4, 
              director = $5, actors = $6, plot = $7, poster_url = $8, version = version + 1 
              where id = $9 and version = $10 
              returning version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Director,
		pq.Array(&movie.Actors),
		&movie.Plot,
		&movie.PosterURL,
		&movie.ID,
		&movie.Version,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}

	return nil
}

func (m *MovieModel) Delete(id int64) error {
	query := `delete from movies where id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoRecordsFound
	}

	return nil
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "title must be provided")
	v.Check(len(movie.Title) <= 100, "title", "title max length is 100 characters")

	v.Check(movie.Year != 0, "year", "year must be provided")
	v.Check(movie.Year >= 0, "year", "year is invalid")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "year cannot be in the future")
}
