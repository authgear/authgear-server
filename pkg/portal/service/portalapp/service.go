package portalapp

import "github.com/authgear/authgear-server/pkg/api/event"

type EventService interface {
	DispatchEvent(payload event.Payload) (err error)
}

type PortalAppService struct {
	Events EventService
}
