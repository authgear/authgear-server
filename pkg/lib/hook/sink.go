package hook

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=sink.go -destination=sink_mock_test.go -package hook

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("hook-sink")} }

type StandardAttributesServiceNoEvent interface {
	UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type CustomAttributesServiceNoEvent interface {
	UpdateAllCustomAttributes(role accesscontrol.Role, userID string, reprForm map[string]interface{}) error
}

type WebHook interface {
	DeliverBlockingEvent(cfg config.BlockingHandlersConfig, e *event.Event) (*event.HookResponse, error)
	DeliverNonBlockingEvent(cfg config.NonBlockingHandlersConfig, e *event.Event) error
}

type Sink struct {
	Logger             Logger
	Config             *config.HookConfig
	Clock              clock.Clock
	WebHook            WebHook
	StandardAttributes StandardAttributesServiceNoEvent
	CustomAttributes   CustomAttributesServiceNoEvent
}

func (s *Sink) ReceiveBlockingEvent(e *event.Event) (err error) {
	if s.WillDeliverBlockingEvent(e.Type) {
		err = s.DeliverBlockingEvent(e)
		if err != nil {
			if !apierrors.IsKind(err, WebHookDisallowed) {
				err = fmt.Errorf("failed to dispatch event: %w", err)
			}
			return
		}
	}

	return
}

func (s *Sink) ReceiveNonBlockingEvent(e *event.Event) (err error) {
	// Skip events that are not for webhook.
	payload := e.Payload.(event.NonBlockingPayload)
	if !payload.ForHook() {
		return
	}

	if s.WillDeliverNonBlockingEvent(e.Type) {
		if err := s.DeliverNonBlockingEvent(e); err != nil {
			s.Logger.WithError(err).Error("failed to dispatch non blocking event")
		}
	}

	return
}

func (s *Sink) DeliverBlockingEvent(e *event.Event) error {
	startTime := s.Clock.NowMonotonic()
	totalTimeout := s.Config.SyncTotalTimeout.Duration()

	mutationsEverApplied := false
	for _, hook := range s.Config.BlockingHandlers {
		if hook.Event != string(e.Type) {
			continue
		}

		elapsed := s.Clock.NowMonotonic().Sub(startTime)
		if elapsed > totalTimeout {
			return WebHookDeliveryTimeout.NewWithInfo("webhook delivery timeout", apierrors.Details{
				"elapsed": elapsed,
				"limit":   totalTimeout,
			})
		}

		resp, err := s.WebHook.DeliverBlockingEvent(hook, e)
		if err != nil {
			return err
		}

		if !resp.IsAllowed {
			return newErrorOperationDisallowed(
				string(e.Type),
				[]OperationDisallowedItem{{
					Title:  resp.Title,
					Reason: resp.Reason,
				}},
			)
		}

		var applied bool
		e, applied = e.ApplyMutations(resp.Mutations)
		if applied {
			mutationsEverApplied = true
		}
	}

	if mutationsEverApplied {
		userID := e.Payload.UserID()
		if mutations, ok := e.GenerateFullMutations(); ok {
			if mutations.User.StandardAttributes != nil {
				err := s.StandardAttributes.UpdateStandardAttributes(
					accesscontrol.RoleGreatest,
					userID,
					mutations.User.StandardAttributes,
				)
				if err != nil {
					return err
				}
			}
			if mutations.User.CustomAttributes != nil {
				err := s.CustomAttributes.UpdateAllCustomAttributes(
					accesscontrol.RoleGreatest,
					userID,
					mutations.User.CustomAttributes,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Sink) DeliverNonBlockingEvent(e *event.Event) error {
	if !e.IsNonBlocking {
		return nil
	}

	checkDeliver := func(events []string, target string) bool {
		for _, event := range events {
			if event == "*" {
				return true
			}
			if event == target {
				return true
			}
		}
		return false
	}

	for _, hook := range s.Config.NonBlockingHandlers {
		shouldDeliver := checkDeliver(hook.Events, string(e.Type))
		if !shouldDeliver {
			continue
		}

		err := s.WebHook.DeliverNonBlockingEvent(hook, e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sink) WillDeliverBlockingEvent(eventType event.Type) bool {
	for _, hook := range s.Config.BlockingHandlers {
		if hook.Event == string(eventType) {
			return true
		}
	}
	return false
}

func (s *Sink) WillDeliverNonBlockingEvent(eventType event.Type) bool {
	for _, hook := range s.Config.NonBlockingHandlers {
		for _, e := range hook.Events {
			if e == "*" {
				return true
			}
			if e == string(eventType) {
				return true
			}
		}
	}
	return false
}
