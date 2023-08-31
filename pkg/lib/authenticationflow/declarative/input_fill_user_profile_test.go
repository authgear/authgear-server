package declarative

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestInputSchemaFillUserProfile(t *testing.T) {
	Convey("InputSchemaFillUserProfile", t, func() {
		test := func(s *InputSchemaFillUserProfile, rawMessage json.RawMessage, expected error) {
			_, err := s.MakeInput(rawMessage)
			if expected == nil {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, expected)
			}
		}

		attributes := []*config.AuthenticationFlowSignupFlowUserProfile{
			{Pointer: "/given_name", Required: true},
			{Pointer: "/family_name", Required: true},
			{Pointer: "/nickname", Required: false},
		}

		// Validate individual value.
		test(&InputSchemaFillUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/given_name", "value": 12 },
				{ "pointer": "/family_name", "value": "john" }
			]
		}`), fmt.Errorf(`invalid value:
/attributes/0/value: type
  map[actual:[integer number] expected:[string]]`))

		// Validate missing required pointers.
		test(&InputSchemaFillUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/given_name", "value": "john" }
			]
		}`), fmt.Errorf(`invalid value:
/attributes/0/pointer: const
  map[actual:/given_name expected:/family_name]`))

		// Validate missing required pointers.
		test(&InputSchemaFillUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/family_name", "value": "john" }
			]
		}`), fmt.Errorf(`invalid value:
/attributes/0/pointer: const
  map[actual:/family_name expected:/given_name]`))

		// Validate unknown pointers.
		test(&InputSchemaFillUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/given_name", "value": "john" },
				{ "pointer": "/family_name", "value": "john" },
				{ "pointer": "/unknown", "value": 42 }
			]
		}`), fmt.Errorf(`invalid value:
/attributes/2/pointer: enum
  map[actual:/unknown expected:[/given_name /family_name /nickname]]`))

		// Valid with required pointer.
		test(&InputSchemaFillUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/given_name", "value": "john" },
				{ "pointer": "/family_name", "value": "john" }
			]
		}`), nil)

		// Valid with optional pointer.
		test(&InputSchemaFillUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/given_name", "value": "john" },
				{ "pointer": "/family_name", "value": "john" },
				{ "pointer": "/nickname", "value": "john" }
			]
		}`), nil)
	})
}
