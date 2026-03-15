package main

import (
	"log"

	"wack-backend/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
