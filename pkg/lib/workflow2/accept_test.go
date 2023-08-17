package workflow2

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

		w := NewWorkflow(newWorkflowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := Accept(ctx, deps, NewWorkflows(w), nil)
		So(errors.Is(err, ErrNoChange), ShouldBeTrue)
	})

	Convey("Bare intent can derive edges that reacts to input", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(newWorkflowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"login_id": "user@example.com"
		}`))

		So(err, ShouldBeNil)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "instance_id": "Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
    "intent": {
        "data": {
            "PretendLoginIDExists": false
        },
        "kind": "intentAuthenticate"
    },
    "nodes": [
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "LoginID": "user@example.com"
                    },
                    "kind": "intentSignup"
                },
                "nodes": [
                    {
                        "type": "SUB_WORKFLOW",
                        "workflow": {
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
    "workflow_id": "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4"
}
		`)
	})

	Convey("Input that cause error will not change the workflow", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(newWorkflowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		jsonStr := `
{
    "instance_id": "Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
    "intent": {
        "data": {
            "PretendLoginIDExists": false
        },
        "kind": "intentAuthenticate"
    },
    "nodes": [
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "LoginID": "user@example.com"
                    },
                    "kind": "intentSignup"
                },
                "nodes": [
                    {
                        "type": "SUB_WORKFLOW",
                        "workflow": {
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
    "workflow_id": "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4"
}
		`

		err := Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"login_id": "user@example.com"
		}`))
		So(err, ShouldBeNil)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, jsonStr)

		err = Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"otp": "nonsense"
		}`))
		So(errors.Is(err, ErrInvalidOTP), ShouldBeTrue)
		bytes, err = json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, jsonStr)
	})

	Convey("Support ErrUpdateNode", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(newWorkflowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"login_id": "user@example.com"
		}`))
		So(err, ShouldBeNil)

		err = Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"resend": true
		}`))
		So(err, ShouldBeNil)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "instance_id": "1MJNB5XDPQQ6TPW2WF73FG01K46CQ9ZD",
    "intent": {
        "data": {
            "PretendLoginIDExists": false
        },
        "kind": "intentAuthenticate"
    },
    "nodes": [
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "LoginID": "user@example.com"
                    },
                    "kind": "intentSignup"
                },
                "nodes": [
                    {
                        "type": "SUB_WORKFLOW",
                        "workflow": {
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
    "workflow_id": "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4"
}
		`)

	})

	Convey("Sub-workflow can end, and the main workflow can start another sub-workflow", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(newWorkflowID(), &intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"login_id": "user@example.com"
		}`))
		So(err, ShouldBeNil)

		err = Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"otp": "123456"
		}`))
		So(err, ShouldBeNil)

		err = Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"create_password": true
		}`))
		So(err, ShouldBeNil)

		err = Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{
			"new_password": "password"
		}`))
		So(errors.Is(err, ErrEOF), ShouldBeTrue)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "instance_id": "M4PEY3W99C69WACPPS9KWK5N09Y46XBC",
    "intent": {
        "data": {
            "PretendLoginIDExists": false
        },
        "kind": "intentAuthenticate"
    },
    "nodes": [
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "LoginID": "user@example.com"
                    },
                    "kind": "intentSignup"
                },
                "nodes": [
                    {
                        "type": "SUB_WORKFLOW",
                        "workflow": {
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
                        "type": "SUB_WORKFLOW",
                        "workflow": {
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
    "workflow_id": "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4"
}
		`)
	})

	Convey("boundary is respected", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(newWorkflowID(), &intentTestBoundarySteps{})

		err := Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{}`))
		So(err, ShouldBeNil)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "workflow_id": "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
    "instance_id": "Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
    "intent": {
        "data": {},
        "kind": "intentTestBoundarySteps"
    },
    "nodes": [
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "Name": "0"
                    },
                    "kind": "intentTestBoundaryStep"
                },
                "nodes": [
                    {
                        "simple": {
                            "data": {},
                            "kind": "nodeTestBoundary"
                        },
                        "type": "SIMPLE"
                    }
                ]
            }
        },
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "Name": "1"
                    },
                    "kind": "intentTestBoundaryStep"
                }
            }
        }
    ]
}
		`)

		err = Accept(ctx, deps, NewWorkflows(w), json.RawMessage(`{}`))
		So(err, ShouldBeNil)

		bytes, err = json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "workflow_id": "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
    "instance_id": "1MJNB5XDPQQ6TPW2WF73FG01K46CQ9ZD",
    "intent": {
        "data": {},
        "kind": "intentTestBoundarySteps"
    },
    "nodes": [
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "Name": "0"
                    },
                    "kind": "intentTestBoundaryStep"
                },
                "nodes": [
                    {
                        "simple": {
                            "data": {},
                            "kind": "nodeTestBoundary"
                        },
                        "type": "SIMPLE"
                    }
                ]
            }
        },
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "Name": "1"
                    },
                    "kind": "intentTestBoundaryStep"
                },
                "nodes": [
                    {
                        "simple": {
                            "data": {},
                            "kind": "nodeTestBoundary"
                        },
                        "type": "SIMPLE"
                    }
                ]
            }
        },
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "intent": {
                    "data": {
                        "Name": "2"
                    },
                    "kind": "intentTestBoundaryStep"
                }
            }
        }
    ]
}
		`)
	})
}
