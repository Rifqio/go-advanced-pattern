package data

import (
	"database/sql"
	"errors"
)

var ErrNoRecordsFound = errors.New("record not found")

// Create a Models struct which wraps the MovieModel. I'll add other models to this,
// like a UserModel and PermissionModel, as the build progresses
type Models struct {
	Movie *MovieModel
}

// For ease of use, I also add a New() method which returns a Models struct containing
// the initialized MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movie: &MovieModel{DB: db},
	}
}
