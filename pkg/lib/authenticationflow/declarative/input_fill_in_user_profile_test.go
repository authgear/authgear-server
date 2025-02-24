package declarative

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestInputSchemaFillInUserProfile(t *testing.T) {
	Convey("InputSchemaFillInUserProfile", t, func() {
		test := func(s *InputSchemaFillInUserProfile, rawMessage json.RawMessage, expected error) {
			_, err := s.MakeInput(context.Background(), rawMessage)
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
		test(&InputSchemaFillInUserProfile{
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
		test(&InputSchemaFillInUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/given_name", "value": "john" }
			]
		}`), fmt.Errorf(`invalid value:
/attributes/0/pointer: const
  map[actual:/given_name expected:/family_name]`))

		// Validate missing required pointers.
		test(&InputSchemaFillInUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/family_name", "value": "john" }
			]
		}`), fmt.Errorf(`invalid value:
/attributes/0/pointer: const
  map[actual:/family_name expected:/given_name]`))

		// Validate unknown pointers.
		test(&InputSchemaFillInUserProfile{
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
		test(&InputSchemaFillInUserProfile{
			Attributes: attributes,
		}, json.RawMessage(`{
			"attributes": [
				{ "pointer": "/given_name", "value": "john" },
				{ "pointer": "/family_name", "value": "john" }
			]
		}`), nil)

		// Valid with optional pointer.
		test(&InputSchemaFillInUserProfile{
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
