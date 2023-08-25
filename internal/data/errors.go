package data

import "errors"

// Global variable to use for each connexion to PSQL
var (
	ErrNoRows = errors.New("models: No matching record found")
)
