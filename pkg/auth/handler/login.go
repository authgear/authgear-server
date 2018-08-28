package handler

import (
	"fmt"
	"io/ioutil"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type LoginHandlerFactory struct{}

func (f LoginHandlerFactory) NewHandler() handler.Handler {
	return &LoginHandler{}
}

// LoginHandler handles login request
type LoginHandler struct {
	db.GetDB `dependency:"DB"`
}

func (h LoginHandler) Handle(ctx handler.Context) {
	input, _ := ioutil.ReadAll(ctx.Request.Body)
	fmt.Fprintln(ctx.ResponseWriter, `{"user": "`+h.GetDB().GetRecord("user:"+string(input))+`"}`)
}
