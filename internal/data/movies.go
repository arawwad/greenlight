package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/arawwad/greenlight/internal/validator"
	"github.com/lib/pq"
)

type Movie struct {
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Genres    []string  `json:"genres,omitempty"`
	ID        int64     `json:"id"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, input *Movie) {
	v.Check(input.Title != "", "title", "must be provided")
	v.Check(len(input.Title) <= 500, "title", "title must not be more than 500 bytes long")

	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	v.Check(input.Year < int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(len(input.Genres) != 0, "genres", "must be provided")
	v.Check(len(input.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(input.Genres), "genres", "must not contain duplicate values")

	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime >= 0, "runtime", "must be positive number")
}

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
    SELECT id, created_at, title, year, runtime, genres, version
    FROM movies
    WHERE id = $1
  `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var movie Movie
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
  INSERT INTO movies (title, year, runtime, genres)
  VALUES ($1, $2, $3, $4)
  RETURNING id, created_at, version
  `

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
    UPDATE movies
    SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
    WHERE id = $5 AND version = $6
    RETURNING version;
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
    DELETE FROM movies
    WHERE id = $1;
  `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
