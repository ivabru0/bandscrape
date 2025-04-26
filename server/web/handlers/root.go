package handlers

import (
	_ "embed"
	"net/http"
)

//go:embed static/nag.html
var nagHTML []byte

// HandleRoot handles the mux pattern /.
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Write(nagHTML)
}
