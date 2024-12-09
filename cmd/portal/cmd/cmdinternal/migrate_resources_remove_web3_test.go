package cmdinternal

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMigrateRemoveWeb3(t *testing.T) {
	Convey("migrateRemoveWeb3", t, func() {
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
			err = migrateRemoveWeb3("app", src, false)
			So(err, ShouldResemble, expectedErr)
			So(src, ShouldResemble, expectedOutput) // src was modified in-place
		}

		toJSON := func(anything interface{}) string {
			b, err := json.Marshal(anything)
			if err != nil {
				panic(err)
			}
			return string(b)
		}

		toB64 := func(str string) string {
			return base64.StdEncoding.EncodeToString([]byte(str))
		}

		Convey("do nothing if the project is not using SIWE", func() {
			test(toJSON(map[string]interface{}{
				"authgear.yaml": toB64(`http:
  public_origin: http://localhost:3100
id: app
`),
			}), toJSON(map[string]interface{}{
				"authgear.yaml": toB64(`http:
  public_origin: http://localhost:3100
id: app
`),
			}), nil)
		})

		Convey("migrate authgear.yaml if authentication.identities contain siwe", func() {
			test(toJSON(map[string]interface{}{
				"authgear.yaml": toB64(`authentication:
  identities: ["siwe"]
http:
  public_origin: http://localhost:3100
id: app
`),
			}), toJSON(map[string]interface{}{
				"authgear.yaml": toB64(`authentication:
  identities:
  - login_id
  - oauth
  primary_authenticators:
  - password
http:
  public_origin: http://localhost:3100
id: app
identity:
  login_id:
    keys:
    - type: email
`),
			}), nil)
		})

		Convey("remove web3 in authgear.yaml if it is present", func() {
			test(toJSON(map[string]interface{}{
				"authgear.yaml": toB64(`authentication:
  identities: ["siwe"]
http:
  public_origin: http://localhost:3100
id: app
web3:
  siwe:
    networks:
    - "1"
`),
			}), toJSON(map[string]interface{}{
				"authgear.yaml": toB64(`authentication:
  identities:
  - login_id
  - oauth
  primary_authenticators:
  - password
http:
  public_origin: http://localhost:3100
id: app
identity:
  login_id:
    keys:
    - type: email
`),
			}), nil)
		})

		Convey("remove web3 in authgear.features.yaml if it is present", func() {
			test(toJSON(map[string]interface{}{
				"authgear.yaml": toB64(`authentication:
  identities: ["siwe"]
http:
  public_origin: http://localhost:3100
id: app
web3:
  siwe:
    networks:
    - ethereum:0x0@1
`),
				"authgear.features.yaml": toB64(`rate_limits:
  disabled: true
web3:
  nft:
    maximum: 2
`),
			}), toJSON(map[string]interface{}{
				"authgear.yaml": toB64(`authentication:
  identities:
  - login_id
  - oauth
  primary_authenticators:
  - password
http:
  public_origin: http://localhost:3100
id: app
identity:
  login_id:
    keys:
    - type: email
`),
				"authgear.features.yaml": toB64(`rate_limits:
  disabled: true
`),
			}), nil)
		})
	})
}
