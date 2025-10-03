package hook

import (
	"context"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

//go:generate go tool mockgen -source=sink.go -destination=sink_mock_test.go -package hook

var SinkLogger = slogutil.NewLogger("hook-sink")

type StandardAttributesServiceNoEvent interface {
	UpdateStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type CustomAttributesServiceNoEvent interface {
	UpdateAllCustomAttributes(ctx context.Context, role accesscontrol.Role, userID string, reprForm map[string]interface{}) error
}

type RolesAndGroupsServiceNoEvent interface {
	ResetUserRole(ctx context.Context, options *rolesgroups.ResetUserRoleOptions) error
	ResetUserGroup(ctx context.Context, options *rolesgroups.ResetUserGroupOptions) error
}

type EventWebHook interface {
	SupportURL(u *url.URL) bool
	DeliverBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) (*event.HookResponse, error)
	DeliverNonBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) error
}

type EventDenoHook interface {
	SupportURL(u *url.URL) bool
	DeliverBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) (*event.HookResponse, error)
	DeliverNonBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) error
}

type Sink struct {
	Config             *config.HookConfig
	Clock              clock.Clock
	EventWebHook       EventWebHook
	EventDenoHook      EventDenoHook
	StandardAttributes StandardAttributesServiceNoEvent
	CustomAttributes   CustomAttributesServiceNoEvent
	RolesAndGroups     RolesAndGroupsServiceNoEvent
}

func (s *Sink) ReceiveBlockingEvent(ctx context.Context, e *event.Event) (err error) {
	if s.WillDeliverBlockingEvent(e.Type) {
		err = s.DeliverBlockingEvent(ctx, e)
		if err != nil {
			if !apierrors.IsKind(err, HookDisallowed) {
				err = fmt.Errorf("failed to dispatch event: %w", err)
			}
			return
		}
	}

	return
}

func (s *Sink) ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) (err error) {
	// Skip events that are not for webhook.
	payload := e.Payload.(event.NonBlockingPayload)
	if !payload.ForHook() {
		return
	}

	if s.WillDeliverNonBlockingEvent(e.Type) {
		err = s.DeliverNonBlockingEvent(ctx, e)
		if err != nil {
			return
		}
	}

	return
}

func (s *Sink) DeliverBlockingEvent(ctx context.Context, e *event.Event) error {
	startTime := s.Clock.NowMonotonic()
	totalTimeout := s.Config.SyncTotalTimeout.Duration()

	mutationsEverApplied := false
	for _, hook := range s.Config.BlockingHandlers {
		if hook.Event != string(e.Type) {
			continue
		}

		elapsed := s.Clock.NowMonotonic().Sub(startTime)
		if elapsed > totalTimeout {
			return HookDeliveryTimeout.NewWithInfo("webhook delivery timeout", apierrors.Details{
				"elapsed": elapsed,
				"limit":   totalTimeout,
			})
		}

		resp, err := s.deliverBlockingEvent(ctx, hook, e)
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

		result := e.ApplyHookResponse(ctx, *resp)
		if result.MutationsEverApplied {
			mutationsEverApplied = true
		}
	}

	if mutationsEverApplied {
		err := e.PerformEffects(ctx, event.MutationsEffectContext{
			StandardAttributes: s.StandardAttributes,
			CustomAttributes:   s.CustomAttributes,
			RolesAndGroups:     s.RolesAndGroups,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sink) DeliverNonBlockingEvent(ctx context.Context, e *event.Event) error {
	logger := SinkLogger.GetLogger(ctx)
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

		errToIgnore := s.deliverNonBlockingEvent(ctx, hook, e)
		if errToIgnore != nil {
			logger.WithError(errToIgnore).Error(ctx, "failed to dispatch non blocking event")
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

func (s *Sink) deliverBlockingEvent(ctx context.Context, cfg config.BlockingHandlersConfig, e *event.Event) (*event.HookResponse, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}
	switch {
	case s.EventWebHook.SupportURL(u):
		return s.EventWebHook.DeliverBlockingEvent(ctx, u, e)
	case s.EventDenoHook.SupportURL(u):
		return s.EventDenoHook.DeliverBlockingEvent(ctx, u, e)
	default:
		return nil, fmt.Errorf("unsupported hook URL: %v", u)
	}
}

func (s *Sink) deliverNonBlockingEvent(ctx context.Context, cfg config.NonBlockingHandlersConfig, e *event.Event) error {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return err
	}
	switch {
	case s.EventWebHook.SupportURL(u):
		return s.EventWebHook.DeliverNonBlockingEvent(ctx, u, e)
	case s.EventDenoHook.SupportURL(u):
		return s.EventDenoHook.DeliverNonBlockingEvent(ctx, u, e)
	default:
		return fmt.Errorf("unsupported hook URL: %v", u)
	}
}
