package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"unicode/utf8"

	"libremc.net/bandscrape/server/database"
)

// SubmitHandler is an HTTP handler which handles the mux pattern /submit.
type SubmitHandler struct {
	db    *database.Database
	regex *regexp.Regexp
}

// ServeHTTP implements the ServeHTTP method for [SubmitHandler].
func (h *SubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	if r.Header.Get("Content-Encoding") != "gzip" {
		http.Error(w, "Content must be gzip-encoded", http.StatusBadRequest)
		return
	}

	gzipReader, err := gzip.NewReader(r.Body)
	if err != nil {
		http.Error(w, "Failed to create gzip reader: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer gzipReader.Close()

	var tracks []database.Track
	if err := json.NewDecoder(gzipReader).Decode(&tracks); err != nil {
		http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(tracks) < 1 {
		http.Error(w, "Too little tracks!", http.StatusBadRequest)
		return
	}

	for _, track := range tracks {
		if track.TrackID < 1 ||
			track.TrackID > 4294967296 ||
			len(track.TrackTitle) == 0 ||
			utf8.RuneCountInString(track.TrackTitle) > 300 ||
			utf8.RuneCountInString(track.AlbumTitle) > 300 ||
			len(track.BandName) == 0 ||
			utf8.RuneCountInString(track.BandName) > 100 ||
			!h.regex.MatchString(track.TrackURL) {
			log.Println("Received invalid track in submission:", track)
			http.Error(w, "Sanity check failed!", http.StatusBadRequest)
			return
		}
	}

	if err := h.db.AddTracks(r.Context(), tracks); err != nil {
		http.Error(w, "Failed to write to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Received submission with", len(tracks), "track(s)!")
	w.WriteHeader(http.StatusCreated)
}

// NewSubmitHandler initializes and returns a new [SubmitHandler].
func NewSubmitHandler(db *database.Database) (*SubmitHandler, error) {
	regex, err := regexp.Compile(`^https://.+/track/[a-z0-9_-]{1,300}$`)
	if err != nil {
		return nil, fmt.Errorf("compile regex: %w", err)
	}

	return &SubmitHandler{
		db:    db,
		regex: regex,
	}, nil
}
