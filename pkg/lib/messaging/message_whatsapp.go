package messaging

import (
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type WhatsappMessage struct {
	message
	taskQueue task.Queue
	events    EventService

	Type         translation.MessageType
	Options      whatsapp.SendTemplateOptions
	IsNotCounted bool
}

type WhatsappSender interface {
	SendTemplate(opts *whatsapp.SendTemplateOptions) error
}

func (m *WhatsappMessage) Send(sender WhatsappSender) error {
	err := m.events.DispatchEventImmediately(&nonblocking.WhatsappSentEventPayload{
		Recipient:           m.Options.To,
		Type:                string(m.Type),
		IsNotCountedInUsage: m.IsNotCounted,
	})
	if err != nil {
		return err
	}

	// We call whatsapp api immediately to know if there is any error
	err = sender.SendTemplate(&m.Options)
	if err != nil {
		return err
	}

	m.isSent = true

	return nil
}
