package main

import (
	"log"
	"net/http"

	"github.com/oursky/ourd/auth"
	"github.com/oursky/ourd/handler"
	_ "github.com/oursky/ourd/oddb/fs"
	"github.com/oursky/ourd/router"
)

func main() {
	authenticator := handler.Authentication{
		TokenStore: auth.FileStore("data/token"),
	}

	r := router.NewRouter()
	r.Map("", handler.HomeHandler)
	r.Map("auth:signup", authenticator.SignupHandler())
	r.Map("auth:login", authenticator.LoginHandler())
	r.Map("record:fetch", handler.RecordFetchHandler)
	r.Map("record:query", handler.RecordQueryHandler)
	r.Map("record:save", handler.RecordSaveHandler)
	r.Map("record:delete", handler.RecordDeleteHandler)
	r.Preprocess(router.CheckAuth)
	r.Preprocess(router.AssignDBConn)
	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}
