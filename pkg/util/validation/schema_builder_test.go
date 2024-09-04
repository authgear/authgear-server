package validation

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaBuilder(t *testing.T) {
	Convey("SchemaBuilder", t, func() {
		test := func(b SchemaBuilder, expected string) {
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		Convey("type required properties enum", func() {
			b := SchemaBuilder{}
			b.Type(TypeObject).Required("channel")

			b.Properties().
				Property("channel", SchemaBuilder{}.Type(TypeString).Enum("sms", "email", "whatsapp"))

			test(b, `
{
    "type": "object",
    "required": [
        "channel"
    ],
    "properties": {
        "channel": {
            "type": "string",
            "enum": [
                "sms",
                "email",
                "whatsapp"
            ]
        }
    }
}
`)
		})

		Convey("type required properties const oneOf", func() {
			b := SchemaBuilder{}.Type(TypeObject)

			b1 := SchemaBuilder{}.
				Required("resend")
			b1.Properties().Property("resend", SchemaBuilder{}.Type(TypeBoolean).Const(true))

			b2 := SchemaBuilder{}.
				Required("check")
			b2.Properties().Property("check", SchemaBuilder{}.Type(TypeBoolean).Const(true))

			b3 := SchemaBuilder{}.
				Required("code")
			b3.Properties().Property("code", SchemaBuilder{}.Type(TypeString))

			b.OneOf(b1, b2, b3)

			test(b, `
{
    "type": "object",
    "oneOf": [
        {
            "required": [
                "resend"
            ],
            "properties": {
                "resend": {
                    "type": "boolean",
                    "const": true
                }
            }
        },
        {
            "required": [
                "check"
            ],
            "properties": {
                "check": {
                    "type": "boolean",
                    "const": true
                }
            }
        },
        {
            "required": [
                "code"
            ],
            "properties": {
                "code": {
                    "type": "string"
                }
            }
        }
    ]
}
`)
		})

		Convey("items contains format", func() {
			b := SchemaBuilder{}
			b.Type(TypeArray)
			b.Items(SchemaBuilder{}.Type(TypeString))
			oneOf := SchemaBuilder{}.OneOf(SchemaBuilder{}.Format("uri"))
			b.Contains(oneOf)

			test(b, `
{
    "type": "array",
    "items": {
        "type": "string"
    },
    "contains": {
        "oneOf": [
            {
                "format": "uri"
            }
        ]
    }
}
`)
		})

		Convey("if then else", func() {
			b := SchemaBuilder{}
			b.If(SchemaBuilder{}.Type(TypeInteger))
			b.Then(SchemaBuilder{}.MinimumInt64(1))
			b.Else(SchemaBuilder{}.MaximumFloat64(2.0))

			test(b, `
{
    "if": {
        "type": "integer"
    },
    "then": {
        "minimum": 1
    },
    "else": {
        "maximum": 2
    }
}
`)
		})

		Convey("mutation on copied builder should not affect original builder", func() {
			b := SchemaBuilder{}
			b.Type(TypeObject).Required("channel")

			b.Properties().
				Property("channel", SchemaBuilder{}.Type(TypeString).Enum("sms", "email", "whatsapp"))

			test(b, `
{
    "type": "object",
    "required": [
        "channel"
    ],
    "properties": {
        "channel": {
            "type": "string",
            "enum": [
                "sms",
                "email",
                "whatsapp"
            ]
        }
    }
}
`)
			newB := b.Copy()
			So(newB, ShouldResemble, b)

			newB.Properties().Property("myNewProperty", SchemaBuilder{}.Type(TypeString))
			test(newB, `
{
    "type": "object",
    "required": [
        "channel"
    ],
    "properties": {
        "channel": {
            "type": "string",
            "enum": [
                "sms",
                "email",
                "whatsapp"
            ]
        },
        "myNewProperty": {
            "type": "string"
        }
    }
}
`)
			test(b, `
{
    "type": "object",
    "required": [
        "channel"
    ],
    "properties": {
        "channel": {
            "type": "string",
            "enum": [
                "sms",
                "email",
                "whatsapp"
            ]
        }
    }
}
`)
		})
		Convey("nullable should append type", func() {
			b := SchemaBuilder{}
			b.Type(TypeObject).Required("channel")

			b.Properties().
				Property("channel", SchemaBuilder{}.Type(TypeString).Enum("sms", "email", "whatsapp"))

			test(b, `
{
    "type": "object",
    "required": [
        "channel"
    ],
    "properties": {
        "channel": {
            "type": "string",
            "enum": [
                "sms",
                "email",
                "whatsapp"
            ]
        }
    }
}
`)
			b.Nullable()
			test(b, `
{
    "type": ["object", "null"],
    "required": [
        "channel"
    ],
    "properties": {
        "channel": {
            "type": "string",
            "enum": [
                "sms",
                "email",
                "whatsapp"
            ]
        }
    }
}
`)
		})
	})
}
