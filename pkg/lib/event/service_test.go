package event

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServiceDispatchEvent(t *testing.T) {
	Convey("Service", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")
		users := NewMockUserService(ctrl)
		database := NewMockDatabase(ctrl)
		sink := NewMockSink(ctrl)
		store := NewMockStore(ctrl)
		fallbackLanguage := "en"
		supportedLanguages := []string{"en"}
		logger := Logger{log.Null}
		localization := &config.LocalizationConfig{
			FallbackLanguage:   &fallbackLanguage,
			SupportedLanguages: supportedLanguages,
		}
		request, _ := http.NewRequest("GET", "/", nil)

		service := &Service{
			Context:      ctx,
			Request:      request,
			Logger:       logger,
			Database:     database,
			Clock:        clock,
			Users:        users,
			Localization: localization,
			Store:        store,
			Sinks:        []Sink{sink},
		}

		var seq0 int64
		store.EXPECT().NextSequenceNumber().AnyTimes().Return(seq0, nil)

		Convey("only use database hook once", func() {
			user := model.User{
				Meta: model.Meta{ID: "user-id"},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}

			database.EXPECT().UseHook(service).Times(1)
			sink.EXPECT().ReceiveBlockingEvent(gomock.Any()).Times(2).Return(nil)

			err := service.DispatchEvent(payload)
			So(err, ShouldBeNil)

			err = service.DispatchEvent(payload)
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

			database.EXPECT().UseHook(service).AnyTimes()
			sink.EXPECT().ReceiveBlockingEvent(&event.Event{
				ID:      "0000000000000000",
				Type:    MockBlockingEventType1,
				Seq:     0,
				Payload: payload,
				Context: event.Context{
					Timestamp:   1136214245,
					UserID:      &userID,
					Language:    fallbackLanguage,
					TriggeredBy: event.TriggeredByTypeUser,
				},
			}).Times(1).Return(nil)

			err := service.DispatchEvent(payload)
			So(err, ShouldBeNil)
			So(service.NonBlockingEvents, ShouldBeEmpty)
		})

		Convey("include user", func() {
			userID := "user-id"
			user := model.User{
				Meta: model.Meta{ID: userID},
			}
			payload := &MockBlockingEvent1{
				MockUserEventBase: MockUserEventBase{user},
			}
			service.Context = session.WithSession(
				service.Context,
				&idpsession.IDPSession{
					ID: "user-id-principal-id",
					Attrs: session.Attrs{
						UserID: userID,
					},
				},
			)

			database.EXPECT().UseHook(service).AnyTimes()
			sink.EXPECT().ReceiveBlockingEvent(
				&event.Event{
					ID:      "0000000000000000",
					Type:    MockBlockingEventType1,
					Seq:     0,
					Payload: payload,
					Context: event.Context{
						Timestamp:   1136214245,
						UserID:      &userID,
						Language:    fallbackLanguage,
						TriggeredBy: event.TriggeredByTypeUser,
					},
				},
			).Return(nil)

			err := service.DispatchEvent(payload)
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

			database.EXPECT().UseHook(service).AnyTimes()
			sink.EXPECT().ReceiveBlockingEvent(gomock.Any()).Return(fmt.Errorf("e"))

			err := service.DispatchEvent(payload)
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

			database.EXPECT().UseHook(service).AnyTimes()
			err := service.DispatchEvent(payload)
			So(err, ShouldBeNil)
			So(service.NonBlockingEvents, ShouldResemble, []*event.Event{
				&event.Event{
					ID:      "0000000000000000",
					Type:    payload.NonBlockingEventType(),
					Seq:     0,
					Payload: payload,
					Context: event.Context{
						Timestamp:   1136214245,
						UserID:      &userID,
						Language:    fallbackLanguage,
						TriggeredBy: event.TriggeredByTypeUser,
					},
					IsNonBlocking: true,
				},
			})
		})

		Convey("send events to sink when transaction was committed", func() {
			payload := &MockNonBlockingEvent1{
				MockUserEventBase: MockUserEventBase{model.User{
					Meta: model.Meta{ID: "user-id"},
				}},
			}
			service.NonBlockingEvents = []*event.Event{
				&event.Event{
					ID:      "0000000000000000",
					Type:    payload.NonBlockingEventType(),
					Seq:     0,
					Payload: payload,
					Context: event.Context{
						Timestamp:   1136214245,
						UserID:      nil,
						Language:    fallbackLanguage,
						TriggeredBy: event.TriggeredByTypeUser,
					},
					IsNonBlocking: true,
				},
			}

			sink.EXPECT().ReceiveNonBlockingEvent(&event.Event{
				ID:      "0000000000000000",
				Type:    payload.NonBlockingEventType(),
				Seq:     0,
				Payload: payload,
				Context: event.Context{
					Timestamp:   1136214245,
					UserID:      nil,
					Language:    fallbackLanguage,
					TriggeredBy: event.TriggeredByTypeUser,
				},
				IsNonBlocking: true,
			})

			service.DidCommitTx()
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

			database.EXPECT().UseHook(service).AnyTimes()
			sink.EXPECT().ReceiveBlockingEvent(&event.Event{
				ID:      "0000000000000000",
				Type:    MockBlockingEventType1,
				Seq:     0,
				Payload: blocking,
				Context: event.Context{
					Timestamp:   1136214245,
					UserID:      &userID,
					Language:    fallbackLanguage,
					TriggeredBy: event.TriggeredByTypeUser,
				},
			}).Return(fmt.Errorf("e"))
			sink.EXPECT().ReceiveNonBlockingEvent(gomock.Any()).Times(0)

			err := service.DispatchEvent(nonBlocking)
			So(err, ShouldBeNil)

			err = service.DispatchEvent(blocking)
			So(err, ShouldBeError, "e")

			err = service.WillCommitTx()
			So(err, ShouldBeNil)

			service.DidCommitTx()
		})
	})
}
