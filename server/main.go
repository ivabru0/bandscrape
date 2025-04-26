package main

import (
	"flag"
	"log"

	"libremc.net/bandscrape/server/database"
	"libremc.net/bandscrape/server/web"
)

func main() {
	address := flag.String("address", ":8585", "network address for BandScrape to listen on")
	dataDir := flag.String("data-dir", "bs_data", "path to store program data in")

	flag.Parse()

	db, err := database.NewDatabase(*dataDir)
	if err != nil {
		log.Fatalln("Failed to load database:", err)
	}
	defer db.Close()

	if err := web.StartWeb(*address, db); err != nil {
		db.Close()
		log.Fatalln("Web server error:", err)
	}
}
