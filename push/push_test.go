package push

import (
	"encoding/json"
	"errors"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"

	"testing"
)

type mockSender struct {
	note   map[string]interface{}
	device skydb.Device
	err    error
}

func (s *mockSender) Send(m Mapper, device skydb.Device) error {
	if s.err != nil {
		return s.err
	}

	s.note = m.Map()
	s.device = device

	return nil
}

func TestRouteSender(t *testing.T) {
	Convey("RouteSender", t, func() {
		routeSender := NewRouteSender()

		Convey("with no senders", func() {
			So(routeSender.Len(), ShouldEqual, 0)
		})

		apnsSender := mockSender{}
		gcmSender := mockSender{}

		routeSender.Route("aps", &apnsSender)
		routeSender.Route("gcm", &gcmSender)

		Convey("routes notification", func() {
			device := skydb.Device{
				Type: "aps",
			}
			message := map[string]interface{}{
				"aps": map[string]interface{}{
					"category": "NEW_MESSAGE_CATEGORY",
				},
			}

			err := routeSender.Send(MapMapper(message), device)
			So(err, ShouldBeNil)
			So(apnsSender.note, ShouldResemble, message)
			So(apnsSender.device, ShouldResemble, device)
		})

		Convey("errors if cannot found a sender", func() {
			device := skydb.Device{
				Type: "sns",
			}

			err := routeSender.Send(EmptyMapper, device)
			So(err.Error(), ShouldEqual, "cannot find sender with type = sns")
		})

		Convey("propagates inner error from sender", func() {
			device := skydb.Device{
				Type: "gcm",
			}

			gcmSender.err = errors.New("mysterious error")
			err := routeSender.Send(EmptyMapper, device)
			So(err, ShouldEqual, gcmSender.err)
		})
	})
}

func jsonToMap(j string) map[string]interface{} {
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(j), &m); err != nil {
		panic(err)
	}
	return m
}
