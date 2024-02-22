package data

import (
	"api.go-rifqio.my.id/internal/validator"
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(user *User) error {
	query := `insert into users (name, email, password_hash, activated)
              values ($1, $2, $3, $4)
              returning id, created_at, version`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return ErrDuplicateEmail
		}
		return err
	}

	return nil
}

func (m *UserModel) GetByEmail(email string) (user *User, err error) {
	query := `select * from users where email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecordsFound
		}
		return nil, err
	}

	return user, nil
}

func (p *password) Set(plaintext string) error {
	// the formula of the cost
	// $2b$[cost]$[22-character salt][31-character hash]
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintext
	p.hash = hash

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

func ValidateEmail(validate *validator.Validator, email string) {
	validate.Check(email != "", "email", "email must be provided")
	validate.Check(validator.Matches(email, validator.EmailRX), "email", "email is not valid")
}

func ValidatePasswordPlaintext(validate *validator.Validator, password string) {
	validate.Check(password != "", "password", "password must be provided")
	validate.Check(len(password) >= 8, "password", "password must be at least 8 bytes long")
	validate.Check(len(password) <= 72, "password", "password must not be more than 72 bytes long")
}

func ValidateUser(validate *validator.Validator, user *User) {
	validate.Check(user.Name != "", "name", "name must be provided")
	validate.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(validate, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(validate, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash")
	}
}
