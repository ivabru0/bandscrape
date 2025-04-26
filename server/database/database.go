package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var initQuery string

// Database is a wrapper around [database/sql.DB] which describes a BandScrape
// database, defining some extra methods specific to it.
type Database struct {
	*sql.DB
}

// NewDatabase loads and returns a new [Database]. The database file is created
// in the specified data directory and initialized as necessary.
func NewDatabase(dataDir string) (*Database, error) {
	// Even though sql.Open will create our database file for us, it leads to
	// confusing errors if the file creation fails, so we'll do it ourselves.
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	dbFile := filepath.Join(dataDir, "bs.db")
	if _, err := os.Stat(dbFile); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(dbFile)
		if err != nil {
			return nil, fmt.Errorf("create file: %w", err)
		}

		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("close file: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec(initQuery); err != nil {
		return nil, fmt.Errorf("execute init query: %w", err)
	}

	return &Database{
		DB: db,
	}, nil
}
