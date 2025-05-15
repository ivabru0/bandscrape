package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"libremc.net/bandscrape/server/database"
	"libremc.net/bandscrape/server/web/handlers"
)

func shutdownOnInterrupt(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("Graceful shutdown failed, forcing:", err)
		if err := server.Close(); err != nil {
			log.Println("Force shutdown failed:", err)
		}
	}
}

// StartWeb initializes and starts up the main BandScrape web server, waiting
// for an interrupt to gracefully shut it down.
func StartWeb(address string, db *database.Database) error {
	mux := http.NewServeMux()

	lookupHandler, err := handlers.NewLookupHandler(db)
	if err != nil {
		return fmt.Errorf("get lookup handler: %w", err)
	}

	submitHandler, err := handlers.NewSubmitHandler(db)
	if err != nil {
		return fmt.Errorf("get submit handler: %w", err)
	}

	mux.Handle("/lookup", lookupHandler)
	mux.Handle("/submit", submitHandler)
	mux.HandleFunc("/", handlers.HandleRoot)

	server := &http.Server{
		Addr: address,
		Handler: http.TimeoutHandler(
			http.MaxBytesHandler(
				mux,
				100000,
			),
			5*time.Second,
			"Timed out\n",
		),
	}

	go shutdownOnInterrupt(server)

	log.Println("Listening on", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http error: %w", err)
	}

	return nil
}
