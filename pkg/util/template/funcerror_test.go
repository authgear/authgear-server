package template

import (
	"bytes"
	"encoding/json"
	htmltemplate "html/template"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func TestFuncsError(t *testing.T) {
	makeTemplate := func() *htmltemplate.Template {
		tmpl := htmltemplate.New("")
		funcMap := MakeTemplateFuncMap(tmpl)
		// To workaround https://github.com/golang/go/issues/42506
		funcMap["unescapeHTML"] = func(s string) htmltemplate.HTML {
			// nolint: gosec
			return htmltemplate.HTML(s)
		}

		tmpl = tmpl.Funcs(funcMap)
		return tmpl
	}
	makeApiErr := func(jsonStr string) *apierrors.APIError {
		var apiErr *apierrors.APIError
		err := json.Unmarshal([]byte(jsonStr), &apiErr)
		if err != nil {
			panic(err)
		}
		return apiErr
	}

	execResolveErrTemplate := func(templateText string, execData interface{}) string {
		tmpl := makeTemplate()
		tmpl, err := tmpl.Parse(templateText)
		if err != nil {
			panic(err)
		}
		buf := &bytes.Buffer{}
		err = tmpl.Execute(buf, execData)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}

	Convey("resolveError", t, func() {

		makePasswordEmptyAPIErrFixture := func() *apierrors.APIError {
			return makeApiErr(`
{
	"code": 400,
	"info": {
		"causes": [
			{
				"details": {
					"actual": [
						"gorilla.csrf.Token",
						"x_step"
					],
					"expected": [
						"x_confirm_password",
						"x_password"
					],
					"missing": [
						"x_confirm_password",
						"x_password"
					]
				},
				"kind": "required",
				"location": ""
			}
		]
	},
	"message": "invalid value",
	"name": "Invalid",
	"reason": "ValidationFailed"
}
`)
		}
		makePasswordNewPasswordTypoAPIErrFixture := func() *apierrors.APIError {
			return makeApiErr(`
{
	"code": 400,
	"message": "new password typo",
	"name": "Invalid",
	"reason": "NewPasswordTypo"
}
			`)
		}
		Convey("it should allow by_reason optional", func() {
			apiErr := makePasswordEmptyAPIErrFixture()
			errMap := execResolveErrTemplate(`
{{- $err_map := (resolveError .RawError (dict 
	"newPasswordField" (dict
		"by_location"                  (list "x_password")
	)
	"confirmPasswordField" (dict
		"by_location"                  (list "x_confirm_password")
	)
)) -}}
{{- toPrettyJson $err_map | unescapeHTML -}}
			`, map[string]interface{}{"RawError": apiErr})

			So(errMap, ShouldEqual, `{
  "confirmPasswordField": {
    "code": 400,
    "info": {
      "causes": [
        {
          "details": {
            "actual": [
              "gorilla.csrf.Token",
              "x_step"
            ],
            "expected": [
              "x_confirm_password"
            ],
            "missing": [
              "x_confirm_password"
            ]
          },
          "kind": "required",
          "location": ""
        }
      ]
    },
    "message": "invalid value",
    "name": "Invalid",
    "reason": "ValidationFailed"
  },
  "newPasswordField": {
    "code": 400,
    "info": {
      "causes": [
        {
          "details": {
            "actual": [
              "gorilla.csrf.Token",
              "x_step"
            ],
            "expected": [
              "x_password"
            ],
            "missing": [
              "x_password"
            ]
          },
          "kind": "required",
          "location": ""
        }
      ]
    },
    "message": "invalid value",
    "name": "Invalid",
    "reason": "ValidationFailed"
  },
  "unknown": null
}`)

		})

		Convey("it should allow by_location optional", func() {
			apiErr := makePasswordEmptyAPIErrFixture()
			// Note `by_location` is not provided
			errMap := execResolveErrTemplate(`
{{- $err_map := (resolveError .RawError (dict 
	"newPasswordField" (dict
		"by_reason"                  (list "PasswordPolicyViolated" "InvalidCredentials")
	)
	"confirmPasswordField" (dict
		"by_reason"                  (list "NewPasswordTypo")
	)
)) -}}
{{- toPrettyJson $err_map | unescapeHTML -}}
			`, map[string]interface{}{"RawError": apiErr})
			So(errMap, ShouldEqual, `{
  "confirmPasswordField": null,
  "newPasswordField": null,
  "unknown": {
    "code": 400,
    "info": {
      "causes": [
        {
          "details": {
            "actual": [
              "gorilla.csrf.Token",
              "x_step"
            ],
            "expected": [
              "x_confirm_password",
              "x_password"
            ],
            "missing": [
              "x_confirm_password",
              "x_password"
            ]
          },
          "kind": "required",
          "location": ""
        }
      ]
    },
    "message": "invalid value",
    "name": "Invalid",
    "reason": "ValidationFailed"
  }
}`)
		})

		Convey("it should allow empty dict", func() {
			apiErr := makePasswordEmptyAPIErrFixture()
			errMap := execResolveErrTemplate(`
{{- $err_map := (resolveError .RawError (dict 
	"newPasswordField" (dict)
	"confirmPasswordField" (dict)
)) -}}
{{- toPrettyJson $err_map | unescapeHTML -}}
			`, map[string]interface{}{"RawError": apiErr})

			So(errMap, ShouldEqual, `{
  "confirmPasswordField": null,
  "newPasswordField": null,
  "unknown": {
    "code": 400,
    "info": {
      "causes": [
        {
          "details": {
            "actual": [
              "gorilla.csrf.Token",
              "x_step"
            ],
            "expected": [
              "x_confirm_password",
              "x_password"
            ],
            "missing": [
              "x_confirm_password",
              "x_password"
            ]
          },
          "kind": "required",
          "location": ""
        }
      ]
    },
    "message": "invalid value",
    "name": "Invalid",
    "reason": "ValidationFailed"
  }
}`)
		})

		Convey("it should return error with matching reasons", func() {
			apiErr := makePasswordNewPasswordTypoAPIErrFixture()
			errMap := execResolveErrTemplate(`
			{{- $err_map := (resolveError .RawError (dict
				"newPasswordField" (dict
					"by_reason"                    (list "InvalidCredentials" "PasswordPolicyViolated")
				)
				"confirmPasswordField" (dict
					"by_reason"                    (list "PasswordPolicyViolated" "NewPasswordTypo")
				)
			)) -}}
			{{- toPrettyJson $err_map | unescapeHTML -}}
						`, map[string]interface{}{"RawError": apiErr})

			So(errMap, ShouldEqual, `{
  "confirmPasswordField": {
    "code": 400,
    "message": "new password typo",
    "name": "Invalid",
    "reason": "NewPasswordTypo"
  },
  "newPasswordField": null,
  "unknown": null
}`)

		})

		Convey("it should return ValidationFailed error if ValidationFailed reason is provided", func() {
			apiErr := makePasswordEmptyAPIErrFixture()
			errMap := execResolveErrTemplate(`
{{- $err_map := (resolveError .RawError (dict 
	"myField" (dict
		"by_reason"                    (list "ValidationFailed")
		"by_location"                  (list "x_password")
	)
)) -}}
{{- toPrettyJson $err_map | unescapeHTML -}}
`, map[string]interface{}{"RawError": apiErr})
			So(errMap, ShouldEqual, `{
  "myField": {
    "code": 400,
    "info": {
      "causes": [
        {
          "details": {
            "actual": [
              "gorilla.csrf.Token",
              "x_step"
            ],
            "expected": [
              "x_confirm_password",
              "x_password"
            ],
            "missing": [
              "x_confirm_password",
              "x_password"
            ]
          },
          "kind": "required",
          "location": ""
        }
      ]
    },
    "message": "invalid value",
    "name": "Invalid",
    "reason": "ValidationFailed"
  },
  "unknown": null
}`)
		})

		Convey("it should slice validation error by the locations (required fields) provided", func() {
			apiErr := makePasswordEmptyAPIErrFixture()
			errMap := execResolveErrTemplate(`
{{- $err_map := (resolveError .RawError (dict 
	"newPasswordField" (dict
		"by_reason"                    (list "InvalidCredentials" "PasswordPolicyViolated")
		"by_location"                  (list "x_password")
	)
	"confirmPasswordField" (dict
		"by_reason"                    (list "PasswordPolicyViolated" "NewPasswordTypo")
		"by_location"                  (list "x_confirm_password")
	)
)) -}}
{{- toPrettyJson $err_map | unescapeHTML -}}
			`, map[string]interface{}{"RawError": apiErr})

			So(errMap, ShouldEqual, `{
  "confirmPasswordField": {
    "code": 400,
    "info": {
      "causes": [
        {
          "details": {
            "actual": [
              "gorilla.csrf.Token",
              "x_step"
            ],
            "expected": [
              "x_confirm_password"
            ],
            "missing": [
              "x_confirm_password"
            ]
          },
          "kind": "required",
          "location": ""
        }
      ]
    },
    "message": "invalid value",
    "name": "Invalid",
    "reason": "ValidationFailed"
  },
  "newPasswordField": {
    "code": 400,
    "info": {
      "causes": [
        {
          "details": {
            "actual": [
              "gorilla.csrf.Token",
              "x_step"
            ],
            "expected": [
              "x_password"
            ],
            "missing": [
              "x_password"
            ]
          },
          "kind": "required",
          "location": ""
        }
      ]
    },
    "message": "invalid value",
    "name": "Invalid",
    "reason": "ValidationFailed"
  },
  "unknown": null
}`)
		})
	})

	Convey("resolveError unexpected error", t, func() {
		Convey("it should return nil if apiErr.Info.Cause[].Kind is a map", func() {
			apiErr := makeApiErr(`
{
	"code": 400,
	"info": {
		"causes": [
			{
				"details": {
					"actual": [
						"gorilla.csrf.Token",
						"x_step"
					],
					"expected": [
						"x_confirm_password",
						"x_password"
					],
					"missing": [
						"x_confirm_password",
						"x_password"
					]
				},
				"kind": {
					"foo": "bar"
				},
				"location": ""
			}
		]
	},
	"message": "invalid value",
	"name": "Invalid",
	"reason": "ValidationFailed"
}
`)
			errMap := execResolveErrTemplate(`
{{- $err_map := (resolveError .RawError (dict 
	"newPasswordField" (dict
		"by_location"                  (list "x_password")
	)
	"confirmPasswordField" (dict
		"by_location"                  (list "x_confirm_password")
	)
)) -}}
{{- toPrettyJson $err_map | unescapeHTML -}}
			`, map[string]interface{}{"RawError": apiErr})
			So(errMap, ShouldEqual, `{
  "confirmPasswordField": null,
  "newPasswordField": null,
  "unknown": {
    "code": 400,
    "info": {
      "causes": [
        {
          "details": {
            "actual": [
              "gorilla.csrf.Token",
              "x_step"
            ],
            "expected": [
              "x_confirm_password",
              "x_password"
            ],
            "missing": [
              "x_confirm_password",
              "x_password"
            ]
          },
          "kind": {
            "foo": "bar"
          },
          "location": ""
        }
      ]
    },
    "message": "invalid value",
    "name": "Invalid",
    "reason": "ValidationFailed"
  }
}`)
		})
	})
}
