package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestWorkflowV2RequestSchema(t *testing.T) {
	Convey("WorkflowV2RequestSchema", t, func() {
		test := func(jsonStr string, expected WorkflowV2Request) {
			body := strings.NewReader(jsonStr)
			r, _ := http.NewRequest("POST", "", body)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			var actual WorkflowV2Request
			err := httputil.BindJSONBody(r, w, WorkflowV2RequestSchema.Validator(), &actual)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		boolFalse := false
		boolTrue := true
		test(`{
			"action": "create",
			"url_query": "client_id=client_id",
			"bind_user_agent": false,
			"intent": {
				"kind": "kind",
				"data": {}
			}
		}`, WorkflowV2Request{
			Action:        "create",
			URLQuery:      "client_id=client_id",
			BindUserAgent: &boolFalse,
			Intent: &workflow.IntentJSON{
				Kind: "kind",
				Data: json.RawMessage([]byte("{}")),
			},
		})
		test(`{
			"action": "input",
			"workflow_id": "workflow_id",
			"instance_id": "instance_id",
			"input": {
				"kind": "kind",
				"data": {}
			}
		}`, WorkflowV2Request{
			Action:        "input",
			WorkflowID:    "workflow_id",
			InstanceID:    "instance_id",
			BindUserAgent: &boolTrue,
			Input: &workflow.InputJSON{
				Kind: "kind",
				Data: json.RawMessage([]byte("{}")),
			},
		})
		test(`{
			"action": "batch_input",
			"workflow_id": "workflow_id",
			"instance_id": "instance_id",
			"batch_input": [
				{
					"kind": "kind",
					"data": {}
				},
				{
					"kind": "kind",
					"data": {}
				}
			]
		}`, WorkflowV2Request{
			Action:        "batch_input",
			WorkflowID:    "workflow_id",
			InstanceID:    "instance_id",
			BindUserAgent: &boolTrue,
			BatchInput: []*workflow.InputJSON{
				{
					Kind: "kind",
					Data: json.RawMessage([]byte("{}")),
				},
				{
					Kind: "kind",
					Data: json.RawMessage([]byte("{}")),
				},
			},
		})
		test(`{
			"action": "get",
			"workflow_id": "workflow_id",
			"instance_id": "instance_id"
		}`, WorkflowV2Request{
			Action:        "get",
			WorkflowID:    "workflow_id",
			InstanceID:    "instance_id",
			BindUserAgent: &boolTrue,
		})
	})
}
