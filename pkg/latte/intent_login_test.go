package latte

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIntentLoginInstantiate(t *testing.T) {
	Convey("IntentLogin.Instantiate", t, func() {
		InstantiateIntent(json.RawMessage(`
		{
			"kind": "latte.IntentLogin",
			"data": {
				"some_string": "a"
			}
		}
		`), "")

		InstantiateIntent(json.RawMessage(`
		{
			"kind": "latte.IntentLogin",
			"data": {}
		}
		`), `invalid value:
<root>: required
  map[actual:<nil> expected:[some_string] missing:[some_string]]`)
	})
}
