package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/arawwad/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}

	p.hash = hash
	p.plaintext = &plaintext

	return nil
}

func (p *password) Matches(plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintext))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func ValidateUser(v *validator.Validator, input *User) {
	v.Check(input.Name != "", "name", "must be provided")
	v.Check(len(input.Name) <= 500, "name", "must not be more than 500 bytes")

	v.Check(input.Email != "", "emal", "must be provided")
	v.Check(validator.Matches(input.Email, validator.EmailRX), "email", "must be a valid email format")

	if input.Password.plaintext != nil {
		v.Check(*input.Password.plaintext != "", "password", "must be provided")
		v.Check(len(*input.Password.plaintext) >= 8, "password", "must not be less than 8 bytes")
		v.Check(len(*input.Password.plaintext) <= 72, "password", "must not be more than 72 bytes")

	}

	if input.Password.hash == nil {
		panic("missing hash for user password")
	}
}

type UserModel struct {
	DB *sql.DB
}

func (u *UserModel) Insert(user *User) error {
	query := `
    INSERT INTO users(name, email, password_hash, activated)
    VALUES($1, $2, $3, $4)
    RETURNING id, created_at, version
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, user.Name, user.Email, user.Password.hash, user.Activated).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return ErrDuplicateEmail
		} else {
			return err
		}
	}

	return nil
}

func (u *UserModel) GetByEmail(email string) (*User, error) {
	query := `
    SELECT id, created_at, name, password_hash, activated, version
    FROM users
    WHERE email = $1
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := User{
		Email: email,
	}
	err := u.DB.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.CreatedAt, &user.Name, &user.Password.hash, &user.Activated, &user.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		} else {
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserModel) Delete(id int) error {
	query := `
    DELETE FROM users
    WHERE id = $1
  `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := u.DB.ExecContext(ctx, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (u *UserModel) Update(user *User) error {
	query := `
    UPDATE users
    SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
    WHERE id = $5 AND version = $6
    RETURNING version
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, user.Name, user.Email, user.Password.hash, user.Activated, user.ID, user.Version).Scan(&user.Version)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return ErrDuplicateEmail
		}
		if err == sql.ErrNoRows {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}
