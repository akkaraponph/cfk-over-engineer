package main

import (
	"cfk/cfk"
	"cfk/pkg/database"
	"cfk/pkg/event"
	"context"
	"log"
	"time"
)

func main() {
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := database.TruncateProjections(db); err != nil {
		log.Fatalf("truncate projections: %v", err)
	}
	log.Println("truncated all projection tables")

	bus := event.NewBus(event.WithWorkerPool(1), event.WithBufferSize(1024))
	cfk.SubscribeProjectionHandlers(bus, db)

	ctx := context.Background()
	bus.Start(ctx)
	defer bus.Stop()

	count := 0
	err = database.ReplayEvents(db, func(evt event.Event) error {
		bus.Publish(evt)
		count++
		return nil
	})

	time.Sleep(500 * time.Millisecond)

	if err != nil {
		log.Fatalf("replay failed: %v", err)
	}

	log.Printf("replay complete: %d events replayed", count)
}
