package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"api.go-rifqio.my.id/internal/validator"
)

const (
	ScopeActivations = "activation"
)

// Check that the plaintext token has been provided and is exactly 26 bytes	long.
func ValidateTokenPlainText(v *validator.Validator, tokenPlainText string) {
	v.Check(tokenPlainText != "", "token", "token must be provided")
	v.Check(len(tokenPlainText) == 26, "token", "token must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

// The New() method is a shortcut which creates a new Token struct and then inserts the
// data in the tokens table
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert() add the data for a specific token in the table
func (m TokenModel) Insert(token *Token) error {
	query := `insert into tokens (hash, user_id, expiry, scope)
			  values ($1, $2, $3, $4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAllForUser() deletes all the token for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `delete from tokens where scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

type Token struct {
	PlainText string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// Create new token containing userId, expiry, and scope information
	// the ttl parameter is to get the expiry time

	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte slice with length of 16 bytes
	randomBytes := make([]byte, 16)

	// Use the Read() function to fill the byte slice with random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the token
	// Plaintext field. This will be the token string that we send to the user in their welcome email.
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate SHA-256 hash of the plaintext token string. This will be the value
	// that we store in `hash` field in the database column.
	// Note that the sha256.Sum256() returns *array* of length 32, so to make it easier
	// we convert it to a slice using the [:] operator before storing it.
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}
