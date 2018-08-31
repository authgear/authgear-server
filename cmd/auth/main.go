package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/handler"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func main() {
	authDependency := dependency.AuthDependency{
		DB: &db.DBProvider{},
	}

	server := server.NewServer("localhost:3000")

	handler.AttachLoginHandler(&server, authDependency)

	go func() {
		log.Printf("Auth gear boot")
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// wait interrupt signal
	select {
	case <-sig:
		log.Printf("Stoping http server ...\n")
	}

	// create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// shutdown the server
	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown request error: %v\n", err)
	}
}
