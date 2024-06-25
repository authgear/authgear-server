package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccept(t *testing.T) {
	ctx := context.TODO()
	deps := &Dependencies{}

	Convey("Ignore incompatible input", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewFlow(newFlowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		_, err := Accept(ctx, deps, NewFlows(w), nil)
		So(errors.Is(err, ErrNoChange), ShouldBeTrue)
	})

	Convey("Bare intent can derive edges that reacts to input", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewFlow(newFlowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		_, err := Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"login_id": "user@example.com"
		}`))

		So(err, ShouldBeNil)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "state_token": "authflowstate_Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
    "intent": {
        "data": {
            "PretendLoginIDExists": false
        },
        "kind": "intentAuthenticate"
    },
    "nodes": [
        {
            "type": "SUB_FLOW",
            "flow": {
                "intent": {
                    "data": {
                        "LoginID": "user@example.com"
                    },
                    "kind": "intentSignup"
                },
                "nodes": [
                    {
                        "type": "SUB_FLOW",
                        "flow": {
                            "intent": {
                                "data": {
                                    "LoginID": "user@example.com"
                                },
                                "kind": "intentAddLoginID"
                            },
                            "nodes": [
                                {
                                    "simple": {
                                        "data": {
                                            "LoginID": "user@example.com",
                                            "OTP": "123456"
                                        },
                                        "kind": "nodeVerifyLoginID"
                                    },
                                    "type": "SIMPLE"
                                }
                            ]
                        }
                    }
                ]
            }
        }
    ],
    "flow_id": "authflow_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4"
}
		`)
	})

	Convey("Input that cause error will not change the flow", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewFlow(newFlowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		jsonStr := `
{
    "state_token": "authflowstate_Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
    "intent": {
        "data": {
            "PretendLoginIDExists": false
        },
        "kind": "intentAuthenticate"
    },
    "nodes": [
        {
            "type": "SUB_FLOW",
            "flow": {
                "intent": {
                    "data": {
                        "LoginID": "user@example.com"
                    },
                    "kind": "intentSignup"
                },
                "nodes": [
                    {
                        "type": "SUB_FLOW",
                        "flow": {
                            "intent": {
                                "data": {
                                    "LoginID": "user@example.com"
                                },
                                "kind": "intentAddLoginID"
                            },
                            "nodes": [
                                {
                                    "simple": {
                                        "data": {
                                            "LoginID": "user@example.com",
                                            "OTP": "123456"
                                        },
                                        "kind": "nodeVerifyLoginID"
                                    },
                                    "type": "SIMPLE"
                                }
                            ]
                        }
                    }
                ]
            }
        }
    ],
    "flow_id": "authflow_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4"
}
		`

		_, err := Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"login_id": "user@example.com"
		}`))
		So(err, ShouldBeNil)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, jsonStr)

		_, err = Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"otp": "nonsense"
		}`))
		So(errors.Is(err, ErrInvalidOTP), ShouldBeTrue)
		bytes, err = json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, jsonStr)
	})

	Convey("Support ErrUpdateNode", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewFlow(newFlowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		_, err := Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"login_id": "user@example.com"
		}`))
		So(err, ShouldBeNil)

		_, err = Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"resend": true
		}`))
		So(err, ShouldBeNil)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "state_token": "authflowstate_1MJNB5XDPQQ6TPW2WF73FG01K46CQ9ZD",
    "intent": {
        "data": {
            "PretendLoginIDExists": false
        },
        "kind": "intentAuthenticate"
    },
    "nodes": [
        {
            "type": "SUB_FLOW",
            "flow": {
                "intent": {
                    "data": {
                        "LoginID": "user@example.com"
                    },
                    "kind": "intentSignup"
                },
                "nodes": [
                    {
                        "type": "SUB_FLOW",
                        "flow": {
                            "intent": {
                                "data": {
                                    "LoginID": "user@example.com"
                                },
                                "kind": "intentAddLoginID"
                            },
                            "nodes": [
                                {
                                    "simple": {
                                        "data": {
                                            "LoginID": "user@example.com",
                                            "OTP": "654321"
                                        },
                                        "kind": "nodeVerifyLoginID"
                                    },
                                    "type": "SIMPLE"
                                }
                            ]
                        }
                    }
                ]
            }
        }
    ],
    "flow_id": "authflow_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4"
}
		`)

	})

	Convey("Sub-flow can end, and the main flow can start another sub-flow", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewFlow(newFlowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		_, err := Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"login_id": "user@example.com"
		}`))
		So(err, ShouldBeNil)

		_, err = Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"otp": "123456"
		}`))
		So(err, ShouldBeNil)

		_, err = Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"create_password": true
		}`))
		So(err, ShouldBeNil)

		_, err = Accept(ctx, deps, NewFlows(w), json.RawMessage(`{
			"new_password": "password"
		}`))
		So(errors.Is(err, ErrEOF), ShouldBeTrue)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "state_token": "authflowstate_M4PEY3W99C69WACPPS9KWK5N09Y46XBC",
    "intent": {
        "data": {
            "PretendLoginIDExists": false
        },
        "kind": "intentAuthenticate"
    },
    "nodes": [
        {
            "type": "SUB_FLOW",
            "flow": {
                "intent": {
                    "data": {
                        "LoginID": "user@example.com"
                    },
                    "kind": "intentSignup"
                },
                "nodes": [
                    {
                        "type": "SUB_FLOW",
                        "flow": {
                            "intent": {
                                "data": {
                                    "LoginID": "user@example.com"
                                },
                                "kind": "intentAddLoginID"
                            },
                            "nodes": [
                                {
                                    "simple": {
                                        "data": {
                                            "LoginID": "user@example.com",
                                            "OTP": "123456"
                                        },
                                        "kind": "nodeVerifyLoginID"
                                    },
                                    "type": "SIMPLE"
                                },
                                {
                                    "simple": {
                                        "data": {
                                            "LoginID": "user@example.com"
                                        },
                                        "kind": "nodeLoginIDVerified"
                                    },
                                    "type": "SIMPLE"
                                }
                            ]
                        }
                    },
                    {
                        "type": "SUB_FLOW",
                        "flow": {
                            "intent": {
                                "data": {},
                                "kind": "intentCreatePassword"
                            },
                            "nodes": [
                                {
                                    "simple": {
                                        "data": {
                                            "HashedNewPassword": "password"
                                        },
                                        "kind": "nodeCreatePassword"
                                    },
                                    "type": "SIMPLE"
                                }
                            ]
                        }
                    }
                ]
            }
        }
    ],
    "flow_id": "authflow_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4"
}
		`)
	})
}
