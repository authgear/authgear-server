package subscription

import (
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/push"
)

// Notice encapsulates the information sent to subscribers when the content of
// a subscription has changed.
type Notice struct {
	SubscriptionID string
	Event          oddb.RecordHookEvent
	Record         *oddb.Record
}

// Notifier is the interface implemented by an object that knows how to deliver
// a Notice to a device.
type Notifier interface {
	Notify(device oddb.Device, notice Notice) error
}

type pushNotifier struct {
	sender push.Sender
}

// NewPushNotifier returns an Notifier which sends Notice
// using the given push.Sender.
func NewPushNotifier(sender push.Sender) Notifier {
	return &pushNotifier{sender}
}

func (notifier *pushNotifier) Notify(device oddb.Device, notice Notice) error {
	customMap := map[string]interface{}{
		"aps": map[string]interface{}{
			"content_available": 1,
		},
		"_ourd": map[string]interface{}{
			"subscription-id": notice.SubscriptionID,
		},
	}

	return notifier.sender.Send(push.MapMapper(customMap), device.Token)
}
