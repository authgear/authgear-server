package declarative

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type recordingEventService struct {
	authflow.EventService
	dispatchedOnCommit []event.Payload
}

func (s *recordingEventService) DispatchEventOnCommit(ctx context.Context, payload event.Payload) error {
	s.dispatchedOnCommit = append(s.dispatchedOnCommit, payload)
	return nil
}

func TestIntentFillForm(t *testing.T) {
	Convey("IntentFillForm", t, func() {
		deps := &authflow.Dependencies{}
		input := &InputFillForm{Fields: []FormFieldInput{{Key: "/given_name", Value: "John"}}}

		newFlows := func(i *IntentFillForm) authflow.Flows {
			return authflow.NewFlows(authflow.NewFlow("flow-id", i))
		}

		Convey("one raw input feeds at most one form", func() {
			ctx := authflow.WithFormInputConsumption(context.Background())

			first := &IntentFillForm{UserID: "u"}
			result, err := first.ReactTo(ctx, deps, newFlows(first), input)
			So(err, ShouldBeNil)
			node := result.(*authflow.Node).Simple.(*NodeDoFillForm)
			So(node.Fields, ShouldResemble, input.Fields)
			So(node.UserID, ShouldEqual, "u")
			So(authflow.IsFormInputConsumed(ctx), ShouldBeTrue)

			second := &IntentFillForm{UserID: "u"}
			_, err = second.ReactTo(ctx, deps, newFlows(second), input)
			So(err, ShouldEqual, authflow.ErrIncompatibleInput)
		})

		Convey("a new request can feed a form again", func() {
			first := &IntentFillForm{}
			_, err := first.ReactTo(authflow.WithFormInputConsumption(context.Background()), deps, newFlows(first), input)
			So(err, ShouldBeNil)

			second := &IntentFillForm{}
			_, err = second.ReactTo(authflow.WithFormInputConsumption(context.Background()), deps, newFlows(second), input)
			So(err, ShouldBeNil)
		})

		Convey("an incompatible input", func() {
			i := &IntentFillForm{}
			_, err := i.ReactTo(authflow.WithFormInputConsumption(context.Background()), deps, newFlows(i), nil)
			So(err, ShouldEqual, authflow.ErrIncompatibleInput)
		})

		Convey("the values submitted with the identification", func() {
			ctx := authflow.WithFormInputConsumption(context.Background())
			i := &IntentFillForm{
				UserID:                        "u",
				IsSubmittedWithIdentification: true,
				SubmittedFields:               input.Fields,
			}

			schema, err := i.CanReactTo(ctx, deps, newFlows(i))
			So(err, ShouldBeNil)
			So(schema, ShouldBeNil)

			result, err := i.ReactTo(ctx, deps, newFlows(i), nil)
			So(err, ShouldBeNil)
			node := result.(*authflow.Node).Simple.(*NodeDoFillForm)
			So(node.Fields, ShouldResemble, input.Fields)
			// The node holds the only copy.
			So(i.SubmittedFields, ShouldBeNil)
			// A later fill_form does not consume the identify input.
			So(authflow.IsFormInputConsumed(ctx), ShouldBeTrue)
		})

		Convey("skipped when the flow continues as an existing user", func() {
			i := &IntentFillForm{}
			So(i.MilestoneSwitchToExistingUser(context.Background(), deps, newFlows(i), "existing"), ShouldBeNil)
			So(i.UserID, ShouldEqual, "existing")
			_, err := i.CanReactTo(context.Background(), deps, newFlows(i))
			So(err, ShouldEqual, authflow.ErrEOF)
		})
	})
}

func TestNodeDoFillForm(t *testing.T) {
	Convey("NodeDoFillForm", t, func() {
		Convey("MilestoneSwitchToExistingUser skips the update", func() {
			n := &NodeDoFillForm{UserID: "new"}
			So(n.MilestoneDidFillForm(), ShouldBeFalse)
			So(n.MilestoneSwitchToExistingUser(context.Background(), &authflow.Dependencies{}, authflow.Flows{}, "existing"), ShouldBeNil)
			So(n.MilestoneDidFillForm(), ShouldBeTrue)
		})
	})
}

func TestDispatchUserProfileUpdateIfFormFilled(t *testing.T) {
	Convey("dispatchUserProfileUpdateIfFormFilled", t, func() {
		events := &recordingEventService{}
		deps := &authflow.Dependencies{Events: events}
		ctx := context.Background()

		flowsWith := func(nodes ...*NodeDoFillForm) authflow.Flows {
			flow := authflow.NewFlow("flow-id", &IntentFillForm{})
			for _, n := range nodes {
				flow.Nodes = append(flow.Nodes, *authflow.NewNodeSimple(n))
			}
			return authflow.NewFlows(flow)
		}

		Convey("no form, no event", func() {
			So(dispatchUserProfileUpdateIfFormFilled(ctx, deps, flowsWith(), "u"), ShouldBeNil)
			So(events.dispatchedOnCommit, ShouldBeEmpty)
		})

		Convey("skipped forms are ignored", func() {
			So(dispatchUserProfileUpdateIfFormFilled(ctx, deps, flowsWith(&NodeDoFillForm{SkipUpdate: true}), "u"), ShouldBeNil)
			So(events.dispatchedOnCommit, ShouldBeEmpty)
		})

		Convey("one pair of events however many forms", func() {
			err := dispatchUserProfileUpdateIfFormFilled(ctx, deps, flowsWith(&NodeDoFillForm{}, &NodeDoFillForm{}, &NodeDoFillForm{SkipUpdate: true}), "u")
			So(err, ShouldBeNil)
			So(events.dispatchedOnCommit, ShouldHaveLength, 2)
			So(string(events.dispatchedOnCommit[0].(event.BlockingPayload).BlockingEventType()), ShouldEqual, "user.profile.pre_update")
			So(string(events.dispatchedOnCommit[1].(event.NonBlockingPayload).NonBlockingEventType()), ShouldEqual, "user.profile.updated")
			So(events.dispatchedOnCommit[0].UserID(), ShouldEqual, "u")
		})
	})
}
