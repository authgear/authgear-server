package main

import (
	"log"
	"net/http"

	"github.com/oursky/ourd/handler"
	_ "github.com/oursky/ourd/oddb/fs"
	"github.com/oursky/ourd/router"
)

func main() {
	r := router.NewRouter()
	r.Map("", handler.HomeHandler)
	r.Map("auth:signup", handler.SignupHandler)
	r.Map("auth:login", handler.LoginHandler)
	r.Map("record:fetch", handler.RecordFetchHandler)
	r.Map("record:query", handler.RecordQueryHandler)
	r.Map("record:save", handler.RecordSaveHandler)
	r.Map("record:delete", handler.RecordDeleteHandler)
	r.Preprocess(router.CheckAuth)
	r.Preprocess(router.AssignDBConn)
	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}
