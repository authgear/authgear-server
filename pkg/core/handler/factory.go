package handler

import (
	"net/http"
)

type Factory interface {
	NewHandler(request *http.Request) Handler
}
