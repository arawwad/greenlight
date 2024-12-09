package data

import (
	"time"

	"github.com/arawwad/greenlight/internal/validator"
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
