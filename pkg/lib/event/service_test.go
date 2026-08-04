package event

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"go.opentelemetry.io/otel/trace"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServiceDispatchEvent(t *testing.T) {
	Convey("Service", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")
		database := NewMockDatabase(ctrl)
		sink := NewMockSink(ctrl)
		store := NewMockStore(ctrl)
		resolver := NewMockResolver(ctrl)
		fallbackLanguage := "en"
		supportedLanguages := []string{"en"}
		localization := &config.LocalizationConfig{
			FallbackLanguage:   &fallbackLanguage,
			SupportedLanguages: supportedLanguages,
		}

		service := &Service{
			Database:     database,
			Clock:        clock,
			Localization: localization,
			Store:        store,
			Resolver:     resolver,
			Sinks:        []Sink{sink},
		}

		var seq0 int64

		Convey("only use database hook once", func() {
			user := model.User{
				Meta: model.Meta{ID: "user-id"},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).Times(1)
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			sink.EXPECT().WillDeliverBlockingEvent(gomock.Any()).AnyTimes().Return(true)
			sink.EXPECT().ReceiveBlockingEvent(ctx, gomock.Any()).Times(2).Return(nil)

			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)

			err = service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)
		})

		Convey("dispatch blocking event", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			sink.EXPECT().WillDeliverBlockingEvent(gomock.Any()).AnyTimes().Return(true)
			sink.EXPECT().ReceiveBlockingEvent(ctx, &event.Event{
				ID:      "0000000000000000",
				Type:    MockBlockingEventType1,
				Seq:     0,
				Payload: payload,
				Context: event.Context{
					Timestamp:          1136214245,
					UserID:             &userID,
					Language:           fallbackLanguage,
					PreferredLanguages: []string{},
					TriggeredBy:        event.TriggeredByTypeUser,
					AuditContext:       event.NewAuditContext("", nil),
					TrackingID:         "",
				},
			}).Times(1).Return(nil)

			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)
			So(service.NonBlockingPayloads, ShouldBeEmpty)
		})

		Convey("dispatch event with tracking id", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			traceID := trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
			spanID := trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
			sc := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: traceID,
				SpanID:  spanID,
			})
			ctx := trace.ContextWithSpanContext(context.Background(), sc)

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			sink.EXPECT().WillDeliverBlockingEvent(gomock.Any()).AnyTimes().Return(true)
			sink.EXPECT().ReceiveBlockingEvent(ctx, &event.Event{
				ID:      "0000000000000000",
				Type:    MockBlockingEventType1,
				Seq:     0,
				Payload: payload,
				Context: event.Context{
					Timestamp:          1136214245,
					UserID:             &userID,
					Language:           fallbackLanguage,
					PreferredLanguages: []string{},
					TriggeredBy:        event.TriggeredByTypeUser,
					AuditContext:       event.NewAuditContext("", nil),
					TrackingID:         "0102030405060708090a0b0c0d0e0f10-0102030405060708",
				},
			}).Times(1).Return(nil)

			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)
		})

		Convey("include user", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := session.WithSession(
				context.Background(),
				&idpsession.IDPSession{
					ID: "user-id-principal-id",
					Attrs: session.Attrs{
						UserID: userID,
					},
				},
			)

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			sink.EXPECT().WillDeliverBlockingEvent(gomock.Any()).AnyTimes().Return(true)
			sink.EXPECT().ReceiveBlockingEvent(
				ctx,
				&event.Event{
					ID:      "0000000000000000",
					Type:    MockBlockingEventType1,
					Seq:     0,
					Payload: payload,
					Context: event.Context{
						Timestamp:          1136214245,
						UserID:             &userID,
						Language:           fallbackLanguage,
						PreferredLanguages: []string{},
						TriggeredBy:        event.TriggeredByTypeUser,
						AuditContext:       event.NewAuditContext("", nil),
						TrackingID:         "",
					},
				},
			).Return(nil)

			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)
		})

		Convey("return sink error", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			sink.EXPECT().WillDeliverBlockingEvent(gomock.Any()).AnyTimes().Return(true)
			sink.EXPECT().ReceiveBlockingEvent(ctx, gomock.Any()).Return(fmt.Errorf("e"))

			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeError, "e")
		})

		Convey("dispatch non-blocking event", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockNonBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)
			So(service.NonBlockingPayloads, ShouldResemble, []event.NonBlockingPayload{
				payload,
			})
		})

		Convey("send events to sink when transaction was committed", func() {
			userID := "user-id"
			payload := &MockNonBlockingEvent1{
				MockUserEventBase: MockUserEventBase{model.User{
					Meta: model.Meta{ID: userID},
				}},
			}
			service.NonBlockingPayloads = []event.NonBlockingPayload{
				payload,
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			sink.EXPECT().ReceiveNonBlockingEvent(ctx, &event.Event{
				ID:      "0000000000000000",
				Type:    payload.NonBlockingEventType(),
				Seq:     0,
				Payload: payload,
				Context: event.Context{
					Timestamp:          1136214245,
					UserID:             &userID,
					Language:           fallbackLanguage,
					PreferredLanguages: []string{},
					TriggeredBy:        event.TriggeredByTypeUser,
					AuditContext:       event.NewAuditContext("", nil),
					TrackingID:         "",
				},
				IsNonBlocking: true,
			})

			err := service.WillCommitTx(ctx)
			So(err, ShouldBeNil)

			service.DidCommitTx(ctx)
		})

		Convey("skip non-blocking events if blocking event has error", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			blocking := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}
			nonBlocking := &MockNonBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			sink.EXPECT().WillDeliverBlockingEvent(gomock.Any()).AnyTimes().Return(true)
			sink.EXPECT().ReceiveBlockingEvent(ctx, &event.Event{
				ID:      "0000000000000000",
				Type:    MockBlockingEventType1,
				Seq:     0,
				Payload: blocking,
				Context: event.Context{
					Timestamp:          1136214245,
					UserID:             &userID,
					Language:           fallbackLanguage,
					PreferredLanguages: []string{},
					TriggeredBy:        event.TriggeredByTypeUser,
					AuditContext:       event.NewAuditContext("", nil),
					TrackingID:         "",
				},
			}).Return(fmt.Errorf("e"))
			sink.EXPECT().ReceiveNonBlockingEvent(ctx, gomock.Any()).Times(0)

			err := service.DispatchEventOnCommit(ctx, nonBlocking)
			So(err, ShouldBeNil)

			err = service.DispatchEventOnCommit(ctx, blocking)
			So(err, ShouldBeError, "e")

			err = service.WillCommitTx(ctx)
			So(err, ShouldBeNil)

			service.DidCommitTx(ctx)
		})

		Convey("PrepareBlockingEventWithTx with no sink returning true for the type skips resolve", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			sink.EXPECT().WillDeliverBlockingEvent(MockBlockingEventType1).Return(false)
			store.EXPECT().NextSequenceNumber(ctx).Times(1).Return(seq0, nil)
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).Times(0)

			e, err := service.PrepareBlockingEventWithTx(ctx, payload, event.PrepareBlockingEventOptions{})
			So(err, ShouldBeNil)
			So(e, ShouldNotBeNil)
		})

		Convey("PrepareBlockingEventWithTx with the hook sink returning true calls Resolve exactly once", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			sink.EXPECT().WillDeliverBlockingEvent(MockBlockingEventType1).Return(true)
			store.EXPECT().NextSequenceNumber(ctx).Times(1).Return(seq0, nil)
			resolver.EXPECT().Resolve(ctx, payload).Times(1).Return(nil)

			e, err := service.PrepareBlockingEventWithTx(ctx, payload, event.PrepareBlockingEventOptions{})
			So(err, ShouldBeNil)
			So(e, ShouldNotBeNil)
		})

		Convey("PrepareBlockingEventWithTx with opts.ResolvedUser set and the sink returning true calls ResolveWithUser and not Resolve", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}
			resolvedUser := &model.User{Meta: model.Meta{ID: userID}}

			ctx := context.Background()

			sink.EXPECT().WillDeliverBlockingEvent(MockBlockingEventType1).Return(true)
			store.EXPECT().NextSequenceNumber(ctx).Times(1).Return(seq0, nil)
			resolver.EXPECT().ResolveWithUser(ctx, payload, resolvedUser).Times(1).Return(nil)
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).Times(0)

			e, err := service.PrepareBlockingEventWithTx(ctx, payload, event.PrepareBlockingEventOptions{
				ResolvedUser: resolvedUser,
			})
			So(err, ShouldBeNil)
			So(e, ShouldNotBeNil)
		})

		Convey("PrepareBlockingEventWithTx with opts.ResolvedUser set but no sink returning true skips resolve entirely", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}
			resolvedUser := &model.User{Meta: model.Meta{ID: userID}}

			ctx := context.Background()

			sink.EXPECT().WillDeliverBlockingEvent(MockBlockingEventType1).Return(false)
			store.EXPECT().NextSequenceNumber(ctx).Times(1).Return(seq0, nil)
			resolver.EXPECT().ResolveWithUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).Times(0)

			e, err := service.PrepareBlockingEventWithTx(ctx, payload, event.PrepareBlockingEventOptions{
				ResolvedUser: resolvedUser,
			})
			So(err, ShouldBeNil)
			So(e, ShouldNotBeNil)
		})

		Convey("DispatchEventOnCommit with a blocking payload and no registered hook does not resolve", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			sink.EXPECT().WillDeliverBlockingEvent(MockBlockingEventType1).Return(false)
			sink.EXPECT().ReceiveBlockingEvent(ctx, gomock.Any()).Return(nil)
			resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).Times(0)

			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)
		})

		Convey("DispatchEventOnCommit with a non-blocking payload whose DeletedUserIDs is empty resolves exactly once, in WillCommitTx", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockNonBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			resolver.EXPECT().Resolve(ctx, payload).Times(1).Return(nil)

			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)

			err = service.WillCommitTx(ctx)
			So(err, ShouldBeNil)
		})

		Convey("DispatchEventOnCommit with a non-blocking payload whose DeletedUserIDs is non-empty resolves twice", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockNonBlockingEvent1{
				MockUserEventBase:  MockUserEventBase{user},
				MockDeletedUserIDs: []string{userID},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			database.EXPECT().UseHook(gomock.Any(), service).AnyTimes()
			resolver.EXPECT().Resolve(ctx, payload).Times(2).Return(nil)

			err := service.DispatchEventOnCommit(ctx, payload)
			So(err, ShouldBeNil)

			err = service.WillCommitTx(ctx)
			So(err, ShouldBeNil)
		})

		Convey("DispatchEventImmediately calls Resolve exactly once", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockNonBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			ctx := context.Background()

			store.EXPECT().NextSequenceNumber(ctx).AnyTimes().Return(seq0, nil)
			resolver.EXPECT().Resolve(ctx, payload).Times(1).Return(nil)
			sink.EXPECT().ReceiveNonBlockingEvent(ctx, gomock.Any()).Return(nil)

			err := service.DispatchEventImmediately(ctx, payload)
			So(err, ShouldBeNil)
		})
	})
}
