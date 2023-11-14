package webapp

import (
	"net/http"
)

type FlashMessage interface {
	Flash(rw http.ResponseWriter, messageType string)
}
