package main

import (
	"cfk/cfk"
	"log"
)

func main() {
	app, err := cfk.NewApp(cfk.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}
