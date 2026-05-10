package main

import (
	"cfk/cfk"
	"cfk/pkg/telemetry"
	"context"
	"log"
)

func main() {
	shutdown, err := telemetry.InitTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer shutdown(context.Background())

	app, err := cfk.NewApp(cfk.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}
