package messaging

import (
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
)

type WhatsappMessage struct {
	message
	taskQueue task.Queue
	events    EventService

	Type        nonblocking.MessageType
	Options     whatsapp.SendTemplateOptions
	IsNotBilled bool
}

func (m *WhatsappMessage) Send() error {
	err := m.events.DispatchEvent(&nonblocking.WhatsappSentEventPayload{
		Recipient:   m.Options.To,
		Type:        m.Type,
		IsNotBilled: m.IsNotBilled,
	})
	if err != nil {
		return err
	}

	m.taskQueue.Enqueue(&tasks.SendMessagesParam{
		WhatsappMessages: []whatsapp.SendTemplateOptions{m.Options},
	})
	m.isSent = true

	return nil
}
