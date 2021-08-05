package smtp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

type SendTestEmailOptions struct {
	To           string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

type Service struct {
	Context context.Context
}

func (s *Service) SendTestEmail(app *model.App, options SendTestEmailOptions) (err error) {
	_ = NewTranslationService(s.Context, app)
	return
}
