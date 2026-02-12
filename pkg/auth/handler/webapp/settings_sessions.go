package webapp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsSessionsHTML = template.RegisterHTML(
	"web/settings_sessions.html",
	Components...,
)

type Authorization struct {
	ID                    string
	ClientID              string
	ClientName            string
	Scope                 []string
	CreatedAt             time.Time
	HasFullUserInfoAccess bool
}

type SettingsSessionsViewModel struct {
	CurrentSessionID string
	Sessions         []*sessionlisting.Session
	Authorizations   []Authorization
}
