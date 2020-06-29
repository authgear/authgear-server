package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/skygeario/skygear-server/cmd/auth/server"
)

func main() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	ctrl := &server.Controller{}
	ctrl.Start()
}
