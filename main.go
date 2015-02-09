package main

import (
	"log"
	"net/http"

	"github.com/oursky/ourd/handlers"
)

func main() {
	r := NewRouter()
	r.HandleFunc("", handlers.HomeHandler)
	r.HandleFunc("auth:login", handlers.LoginHandler)
	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}
