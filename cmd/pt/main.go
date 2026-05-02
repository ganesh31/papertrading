package main

import (
	"log"

	"github.com/ganesh/papertrading/internal/ptcli"
	"github.com/joho/godotenv"
)

func main() {
	// Load repo-root .env so DATABASE_URL and other vars apply without exporting manually.
	_ = godotenv.Load()

	if err := ptcli.Execute(); err != nil {
		log.Fatal(err)
	}
}
