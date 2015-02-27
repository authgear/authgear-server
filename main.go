package main

import (
	"log"
	"net/http"

	"github.com/oursky/ourd/handlers"
	_ "github.com/oursky/ourd/oddb/fs"
	"github.com/oursky/ourd/router"
)

func main() {
	r := router.NewRouter()
	r.Map("", handlers.HomeHandler)
	r.Map("auth:signup", handlers.SignupHandler)
	r.Map("auth:login", handlers.LoginHandler)
	r.Map("record:fetch", handlers.RecordFetchHandler)
	r.Map("record:query", handlers.RecordQueryHandler)
	r.Map("record:save", handlers.RecordSaveHandler)
	r.Map("record:delete", handlers.RecordDeleteHandler)
	r.Preprocess(router.CheckAuth)
	r.Preprocess(router.AssignDBConn)
	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}
