// Package main is the entrypoint for the openmeteo-cli binary.
//
// This CLI fetches weather forecasts from Open-Meteo and outputs them in
// structured formats (TOON or JSON).
package main

import (
	"os"

	"openmeteo-cli/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
