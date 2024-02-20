package tests

import (
	"context"
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/authflowclient"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignup(t *testing.T) {
	Convey("Signup", t, func() {
		client := authflowclient.NewClient(context.Background(), "localhost:4000", "signup.portal.localhost:4000")

		var create_resp *authflowclient.FlowResponse
		var err error

		create_resp, err = client.Create(authflowclient.FlowReference{
			Type: "signup",
			Name: "default",
		}, "")

		if err != nil {
			t.Fatalf("failed to create flow: %v", err)
		}

		So(create_resp.Type, ShouldEqual, authflowclient.FlowTypeSignup)
		So(create_resp.Action.Type, ShouldEqual, authflowclient.FlowActionTypeIdentify)
		So(create_resp.StateToken, ShouldNotBeEmpty)

		identify_resp, err := client.Input(nil, nil, create_resp.StateToken, map[string]interface{}{
			"identification": "email",
			"login_id":       "signup@authgear.com",
		})

		if err != nil {
			t.Fatalf("failed to input: %v", err)
		}

		So(identify_resp.Type, ShouldEqual, authflowclient.FlowTypeSignup)
		So(identify_resp.Action.Type, ShouldEqual, authflowclient.FlowActionTypeVerify)
		So(identify_resp.StateToken, ShouldNotBeEmpty)

		verify_resp, err := client.Input(nil, nil, identify_resp.StateToken, map[string]interface{}{
			"code": "111111",
		})

		if err != nil {
			t.Fatalf("failed to input: %v", err)
		}

		So(verify_resp.Type, ShouldEqual, authflowclient.FlowTypeSignup)
		So(verify_resp.Action.Type, ShouldEqual, authflowclient.FlowActionTypeCreateAuthenticator)
		So(verify_resp.StateToken, ShouldNotBeEmpty)

		create_authenticator_resp, err := client.Input(nil, nil, verify_resp.StateToken, map[string]interface{}{
			"authentication": "primary_password",
			"new_password":   "password",
		})

		if err != nil {
			t.Fatalf("failed to input: %v", err)
		}

		So(create_authenticator_resp.Type, ShouldEqual, authflowclient.FlowTypeSignup)
		So(create_authenticator_resp.Action.Type, ShouldEqual, authflowclient.FlowActionTypeFinished)
	})
}
