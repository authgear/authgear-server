package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type SendWhatsappCodeResult struct {
	Target     string
	CodeLength int
}

type SendWhatsappCode struct {
	Context *interaction.Context
	Target  string
}

func (s *SendWhatsappCode) Do() (*SendWhatsappCodeResult, error) {
	code, err := s.Context.WhatsappCodeProvider.GenerateCode(s.Target, s.Context.WebSessionID)
	if err != nil {
		return nil, err
	}
	err = s.Context.WhatsappCodeProvider.SendCode(s.Target, code.Code)
	if err != nil {
		return nil, err
	}

	return &SendWhatsappCodeResult{
		Target:     s.Target,
		CodeLength: code.CodeLength,
	}, nil
}

func NewSendWhatsappCode(ctx *interaction.Context, target string) *SendWhatsappCode {
	return &SendWhatsappCode{
		Context: ctx,
		Target:  target,
	}
}
