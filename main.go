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
	r.Map("record:fetch", handlers.RecordFetchHandler)
	r.Map("record:query", handlers.RecordQueryHandler)
	r.Map("record:save", handlers.RecordSaveHandler)
	r.Map("record:delete", handlers.RecordDeleteHandler)
	r.Preprocess(CheckAuth)
	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}
