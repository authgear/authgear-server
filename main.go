package main

import (
	"log"
	"net/http"

	"github.com/oursky/ourd/handlers"
)

func main() {
	r := NewRouter()
	r.Map("", handlers.HomeHandler)
	r.Map("auth:signup", handlers.SignupHandler)
	r.Map("auth:login", handlers.LoginHandler)
	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}
