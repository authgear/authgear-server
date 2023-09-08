package authenticationflow

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMergeData(t *testing.T) {
	Convey("MergeData", t, func() {
		Convey("flatten", func() {
			d1 := mapData{
				"a": "d1",
				"b": "d1",
			}
			d2 := &DataFinishRedirectURI{
				FinishRedirectURI: "http://localhost/",
			}
			d3 := &DataFlowDetails{
				FlowReference: FlowReference{
					Type: FlowTypeLogin,
					ID:   "default",
				},
			}
			d4 := mapData{
				"a": "d4",
			}

			data := MergeData(d1, d2, d3, d4)

			dataBytes, err := json.Marshal(data)
			So(err, ShouldBeNil)
			So(string(dataBytes), ShouldEqualJSON, `
			{
				"a": "d4",
				"b": "d1",
				"finish_redirect_uri": "http://localhost/",
				"flow_reference": {
					"type": "login_flow",
					"id": "default"
				}
			}
			`)
		})
	})
}
