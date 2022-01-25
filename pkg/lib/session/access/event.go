package access

import (
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

func NewEvent(timestamp time.Time, remoteIP httputil.RemoteIP, userAgentString httputil.UserAgentString) Event {
	return Event{
		Timestamp: timestamp,
		RemoteIP:  string(remoteIP),
		UserAgent: string(userAgentString),
	}
}
