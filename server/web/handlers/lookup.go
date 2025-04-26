package handlers

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"libremc.net/bandscrape/server/database"
)

//go:embed static/lookup.tmpl
var lookupTmpl string

// LookupHandler is an HTTP handler which handles the mux pattern /lookup.
type LookupHandler struct {
	db   *database.Database
	tmpl *template.Template
}

// ServeHTTP implements the ServeHTTP method for [LookupHandler].
func (h *LookupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var tmplData []database.Track
	if r.Method == http.MethodPost {
		defer r.Body.Close()

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}

		tracks, err := h.db.GetTracks(
			r.Context(),
			map[string]string{
				"track_title": r.Form.Get("track_title"),
				"album_title": r.Form.Get("album_title"),
				"band_name":   r.Form.Get("band_name"),
				"track_url":   r.Form.Get("track_url"),
			},
		)
		if err != nil {
			http.Error(w, "Failed to get tracks: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tmplData = tracks
		log.Println("Lookup completed and returned", len(tmplData), "tracks!")
	}

	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, tmplData); err != nil {
		http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		http.Error(w, "Failed to rewrite buffer: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewLookupHandler initializes and returns a new [LookupHandler].
func NewLookupHandler(db *database.Database) (*LookupHandler, error) {
	tmpl, err := template.New("lookup").Parse(lookupTmpl)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	return &LookupHandler{
		db:   db,
		tmpl: tmpl,
	}, nil
}
