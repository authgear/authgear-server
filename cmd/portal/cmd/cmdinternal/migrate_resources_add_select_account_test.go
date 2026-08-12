package cmdinternal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMigrateAddSelectAccount(t *testing.T) {
	Convey("migrateAddSelectAccount", t, func() {
		toB64 := func(str string) string {
			return base64.StdEncoding.EncodeToString([]byte(str))
		}

		toJSON := func(anything any) string {
			b, err := json.Marshal(anything)
			if err != nil {
				panic(err)
			}
			return string(b)
		}

		test := func(srcJSON string, expectedOutputJSON string, expectedErr error) {
			src := make(map[string]string)
			err := json.Unmarshal([]byte(srcJSON), &src)
			if err != nil {
				panic(err)
			}
			expectedOutput := make(map[string]string)
			err = json.Unmarshal([]byte(expectedOutputJSON), &expectedOutput)
			if err != nil {
				panic(err)
			}
			err = migrateAddSelectAccount(context.Background(), "app", src, false)
			So(err, ShouldResemble, expectedErr)
			So(src, ShouldResemble, expectedOutput) // src was modified in-place
		}

		// unchanged expects the input to be left byte-for-byte alone, which also
		// proves the migration does not needlessly reformat untouched configs.
		unchanged := func(yaml string) {
			test(
				toJSON(map[string]any{"authgear.yaml": toB64(yaml)}),
				toJSON(map[string]any{"authgear.yaml": toB64(yaml)}),
				nil,
			)
		}

		Convey("do nothing if there is no authentication_flow", func() {
			unchanged(`http:
  public_origin: http://localhost:3100
id: app
`)
		})

		Convey("do nothing if there are no login_flows", func() {
			unchanged(`authentication_flow:
  signup_flows:
  - name: default
    steps:
    - one_of:
      - identification: email
      type: identify
id: app
`)
		})

		Convey("do nothing if login_flows is empty", func() {
			unchanged(`authentication_flow:
  login_flows: []
id: app
`)
		})

		Convey("do nothing if the first step is not identify", func() {
			unchanged(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - authentication: primary_password
      type: authenticate
    - one_of:
      - identification: email
      type: identify
id: app
`)
		})

		Convey("do nothing if select_account is already offered", func() {
			unchanged(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: select_account
      - identification: email
      type: identify
id: app
`)
		})

		Convey("add select_account to the default login flow", func() {
			test(toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - name: login_identify
      one_of:
      - identification: email
      type: identify
id: app
`),
			}), toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - name: login_identify
      one_of:
      - identification: select_account
      - identification: email
      type: identify
id: app
`),
			}), nil)
		})

		Convey("add select_account to a flow that is not named default", func() {
			test(toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: custom_only
    steps:
    - one_of:
      - identification: ldap
      type: identify
id: app
`),
			}), toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: custom_only
    steps:
    - one_of:
      - identification: select_account
      - identification: ldap
      type: identify
id: app
`),
			}), nil)
		})

		Convey("add select_account to every qualifying flow, leaving the others alone", func() {
			test(toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: email
      type: identify
  - name: phone_only
    steps:
    - one_of:
      - identification: phone
      type: identify
  - name: pw_first
    steps:
    - one_of:
      - authentication: primary_password
      type: authenticate
    - one_of:
      - identification: email
      type: identify
id: app
`),
			}), toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: select_account
      - identification: email
      type: identify
  - name: phone_only
    steps:
    - one_of:
      - identification: select_account
      - identification: phone
      type: identify
  - name: pw_first
    steps:
    - one_of:
      - authentication: primary_password
      type: authenticate
    - one_of:
      - identification: email
      type: identify
id: app
`),
			}), nil)
		})

		Convey("skip only the flow that already offers select_account", func() {
			test(toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: select_account
      - identification: email
      type: identify
  - name: alt
    steps:
    - one_of:
      - identification: username
      type: identify
id: app
`),
			}), toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: select_account
      - identification: email
      type: identify
  - name: alt
    steps:
    - one_of:
      - identification: select_account
      - identification: username
      type: identify
id: app
`),
			}), nil)
		})

		Convey("leave signup_flows and reauth_flows alone", func() {
			test(toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: ldap
      type: identify
  reauth_flows:
  - name: default
    steps:
    - one_of:
      - identification: id_token
      type: identify
  signup_flows:
  - name: default
    steps:
    - one_of:
      - identification: email
      type: identify
id: app
`),
			}), toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: select_account
      - identification: ldap
      type: identify
  reauth_flows:
  - name: default
    steps:
    - one_of:
      - identification: id_token
      type: identify
  signup_flows:
  - name: default
    steps:
    - one_of:
      - identification: email
      type: identify
id: app
`),
			}), nil)
		})

		Convey("preserve nested steps under a one_of branch", func() {
			test(toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: email
        steps:
        - one_of:
          - authentication: primary_password
          type: authenticate
      type: identify
id: app
`),
			}), toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: select_account
      - identification: email
        steps:
        - one_of:
          - authentication: primary_password
          type: authenticate
      type: identify
id: app
`),
			}), nil)
		})

		// Documents a consequence shared by every migration in this package:
		// a touched document is re-marshalled, which sorts keys alphabetically
		// and drops comments. Untouched documents are left byte-for-byte alone
		// (see the `unchanged` cases above).
		Convey("re-marshalling a touched config sorts keys and drops comments", func() {
			test(toJSON(map[string]any{
				"authgear.yaml": toB64(`id: app
authentication_flow:
  login_flows:
  # the default flow
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
`),
			}), toJSON(map[string]any{
				"authgear.yaml": toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: select_account
      - identification: email
      type: identify
id: app
`),
			}), nil)
		})

		Convey("is idempotent", func() {
			src := make(map[string]string)
			src["authgear.yaml"] = toB64(`authentication_flow:
  login_flows:
  - name: default
    steps:
    - one_of:
      - identification: email
      type: identify
id: app
`)
			So(migrateAddSelectAccount(context.Background(), "app", src, false), ShouldBeNil)
			afterFirst := src["authgear.yaml"]

			So(migrateAddSelectAccount(context.Background(), "app", src, false), ShouldBeNil)
			So(src["authgear.yaml"], ShouldEqual, afterFirst)
		})
	})
}
