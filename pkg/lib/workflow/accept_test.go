package workflow

import (
	"encoding/json"
	"errors"
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccept(t *testing.T) {
	ctx := &Context{}

	Convey("Ignore incompatible input", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(&intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := w.Accept(ctx, nil)
		So(errors.Is(err, ErrNoChange), ShouldBeTrue)
	})

	Convey("Bare intent can derive edges that reacts to input", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(&intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := w.Accept(ctx, &inputLoginID{
			LoginID: "user@example.com",
		})

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

		w := NewWorkflow(&intentAuthenticate{
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

		err := w.Accept(ctx, &inputLoginID{
			LoginID: "user@example.com",
		})
		So(err, ShouldBeNil)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, jsonStr)

		err = w.Accept(ctx, &inputOTP{
			OTP: "nonsense",
		})
		So(errors.Is(err, ErrInvalidOTP), ShouldBeTrue)
		bytes, err = json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, jsonStr)
	})

	Convey("Support ErrUpdateNode", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(&intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := w.Accept(ctx, &inputLoginID{
			LoginID: "user@example.com",
		})
		So(err, ShouldBeNil)

		err = w.Accept(ctx, &inputResendOTP{})
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

		w := NewWorkflow(&intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := w.Accept(ctx, &inputLoginID{
			LoginID: "user@example.com",
		})
		So(err, ShouldBeNil)

		err = w.Accept(ctx, &inputOTP{
			OTP: "123456",
		})
		So(err, ShouldBeNil)

		err = w.Accept(ctx, &inputCreatePasswordFlow{})
		So(err, ShouldBeNil)

		err = w.Accept(ctx, &inputNewPassword{
			NewPassword: "password",
		})
		So(err, ShouldBeNil)

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

	Convey("A workflow can be ended at wish", t, func() {
		rng = rand.New(rand.NewSource(0))

		w := NewWorkflow(&intentAuthenticate{
			PretendLoginIDExists: false,
		})

		err := w.Accept(ctx, &inputLoginID{
			LoginID: "user@example.com",
		})
		So(err, ShouldBeNil)

		err = w.Accept(ctx, &inputOTP{
			OTP: "123456",
		})
		So(err, ShouldBeNil)

		err = w.Accept(ctx, &inputCreatePasswordFlow{})
		So(err, ShouldBeNil)

		err = w.Accept(ctx, &inputNewPassword{
			NewPassword: "password",
		})
		So(err, ShouldBeNil)

		err = w.Accept(ctx, &inputFinishSignup{})
		So(errors.Is(err, ErrEOF), ShouldBeTrue)

		bytes, err := json.Marshal(w)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "instance_id": "P44Q4ZP6CA6VCAGEHTM9PH5Y845Y0ZNE",
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
                    },
                    {
                        "type": "SUB_WORKFLOW",
                        "workflow": {
                            "intent": {
                                "data": {},
                                "kind": "intentFinishSignup"
                            }
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
}
