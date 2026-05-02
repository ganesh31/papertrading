package main

import (
	"os"

	"github.com/ganesh/papertrading/internal/ptcli"
)

func main() {
	if err := ptcli.Execute(); err != nil {
		os.Exit(1)
	}
}
