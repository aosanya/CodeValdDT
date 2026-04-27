// Command dev is the local-development CodeValdDT gRPC binary.
// It loads .env when present (via the Makefile `dev` target) so secrets such
// as DT_ARANGO_PASSWORD stay out of the source tree. The binary itself does
// not parse .env — the Makefile sources it before exec — so configuration is
// otherwise identical to cmd/server.
package main

import (
	"log"

	"github.com/aosanya/CodeValdDT/internal/app"
	"github.com/aosanya/CodeValdDT/internal/config"
)

func main() {
	if err := app.Run(config.Load()); err != nil {
		log.Fatalf("codevalddt-dev: %v", err)
	}
}
