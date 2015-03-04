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
	tokenStore := auth.FileStore("data/token")

	authenticator := handler.Authentication{
		TokenStore: tokenStore,
	}

	recordService := handler.RecordService{
		TokenStore: tokenStore,
	}

	r := router.NewRouter()
	r.Map("", handler.HomeHandler)
	r.Map("auth:signup", authenticator.SignupHandler())
	r.Map("auth:login", authenticator.LoginHandler())
	r.Map("record:fetch", recordService.RecordFetchHandler())
	r.Map("record:query", recordService.RecordQueryHandler())
	r.Map("record:save", recordService.RecordSaveHandler())
	r.Map("record:delete", recordService.RecordDeleteHandler())
	r.Preprocess(router.AssignDBConn)
	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}
