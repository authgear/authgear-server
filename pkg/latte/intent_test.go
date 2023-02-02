package latte

import (
	"encoding/json"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func InstantiateIntent(msg json.RawMessage, errStr string) {
	var intentJSON workflow.IntentJSON
	err := json.Unmarshal(msg, &intentJSON)
	So(err, ShouldBeNil)
	_, err = workflow.InstantiateIntent(intentJSON)
	if errStr == "" {
		So(err, ShouldBeNil)
	} else {
		So(err, ShouldBeError, errStr)
	}
}
