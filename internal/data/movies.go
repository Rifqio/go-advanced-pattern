package data

import (
	"api.go-rifqio.my.id/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
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

/*
* Query Cheat Sheet
* DB.Query() -is used for SELECT queries which return multiple rows.
* DB.QueryRow() -is used for SELECT queries which return a single row.
* DB.Exec() -is used for INSERT, UPDATE and DELETE queries, and it does not return any rows.
 */

// MovieModel Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// Insert If the receiver is a struct or array, any of whose elements is a pointer to something that may be mutated,
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

func (m *MovieModel) GetAll(title string, genres []string, filter Filters) ([]*Movie, PaginationMetadata, error) {
	// @> means array operator for Postgres
	// implement full text search on title...
	// to_tsvector() takes a string and split to lexemes (one word or several word)
	// specifying 'simple' means transpose the title to lowercase
	// plainto_tsquery() function takes a search value and turns into formatted query term PostgreSQL
	// the @@ operator is the matches operator. In our statement we are using it to check whether
	// the generated query term matches the lexemes.

	// Also this workaround using fmt.Sprintf() since order by has no placeholder for arguments
	query := fmt.Sprintf(
		`select count(*) over(), * from movies 
         		where (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) or $1 = '')
			  	and (genres @> $2 or $2 = '{}')
			  	order by %s %s, id asc 
			  	limit $3 offset $4`, filter.sortColumn(), filter.sortDirection(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, pq.Array(genres), filter.limit(), filter.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, PaginationMetadata{}, err
	}

	defer rows.Close()

	var totalRecords int
	var movies []*Movie

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&totalRecords,
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
			return nil, PaginationMetadata{}, err
		}

		movies = append(movies, &movie)
	}

	if rows.Err() != nil {
		return nil, PaginationMetadata{}, err
	}

	paginationMetadata := calculatePaginationMetadata(totalRecords, filter.Page, filter.PageSize)
	return movies, paginationMetadata, nil
}

func (m *MovieModel) Count() (int, error) {
	query := `select count(*) from movies`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	err := m.DB.QueryRowContext(ctx, query).Scan(&count)

	if err != nil {
		return 0, nil
	}

	return count, nil
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
