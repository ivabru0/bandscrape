package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Track describes a row in the database's tracks table.
type Track struct {
	TrackID    int    `json:"track_id"`
	TrackTitle string `json:"track_title"`
	AlbumTitle string `json:"album_title"`
	BandName   string `json:"band_name"`
	TrackURL   string `json:"track_url"`
	Created    int
}

// AddTracks adds a submission (slice of tracks) to the database's tracks table.
func (db *Database) AddTracks(ctx context.Context, tracks []Track) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO tracks(track_id, track_title, album_title, band_name, track_url)
		VALUES(?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, track := range tracks {
		if _, err := stmt.ExecContext(
			ctx,
			track.TrackID,
			track.TrackTitle,
			track.AlbumTitle,
			track.BandName,
			track.TrackURL,
		); err != nil {
			return fmt.Errorf("execute for %d: %w", track.TrackID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetTracks returns a slice of tracks looked up from the database's tracks
// table, selected by the provided fields' values.
func (db *Database) GetTracks(ctx context.Context, fields map[string]string) ([]Track, error) {
	var whereClauses []string
	var queryArgs []any
	allowedKeys := map[string]bool{
		"track_title": true,
		"album_title": true,
		"band_name":   true,
		"track_url":   true,
	}

	for k, v := range fields {
		if allowedKeys[k] && v != "" {
			whereClauses = append(whereClauses, k+" = ?")
			queryArgs = append(queryArgs, v)
		}
	}

	if len(whereClauses) == 0 {
		return nil, errors.New("not enough data")
	}

	rows, err := db.QueryContext(
		ctx,
		"SELECT * FROM tracks WHERE "+strings.Join(whereClauses, " AND ")+" ORDER BY created DESC",
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		if err := rows.Scan(
			&track.TrackID,
			&track.TrackTitle,
			&track.AlbumTitle,
			&track.BandName,
			&track.TrackURL,
			&track.Created,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}
