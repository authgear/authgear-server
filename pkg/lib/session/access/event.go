package access

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type Info struct {
	InitialAccess Event `json:"initial_access"`
	LastAccess    Event `json:"last_access"`
}

type Event struct {
	Timestamp time.Time `json:"time"`
	RemoteIP  string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

func NewEvent(timestamp time.Time, req *http.Request, trustProxy bool) Event {
	return Event{
		Timestamp: timestamp,
		RemoteIP:  httputil.GetIP(req, trustProxy),
		UserAgent: req.UserAgent(),
	}
}
