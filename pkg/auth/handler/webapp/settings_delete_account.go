package webapp

import (
	"context"

	"time"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsDeleteAccountHTML = template.RegisterHTML(
	"web/settings_delete_account.html",
	Components...,
)

type SettingsDeleteAccountViewModel struct {
	ExpectedAccountDeletionTime time.Time
}

type SettingsDeleteAccountUserService interface {
	ScheduleDeletionByEndUser(ctx context.Context, userID string) error
}

type SettingsDeleteAccountOAuthSessionService interface {
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
	Save(ctx context.Context, entry *oauthsession.Entry) error
}

type SettingsDeleteAccountSessionStore interface {
	Update(ctx context.Context, session *webapp.Session) (err error)
}

type SettingsDeleteAccountAuthenticationInfoService interface {
	Save(ctx context.Context, entry *authenticationinfo.Entry) (err error)
}
