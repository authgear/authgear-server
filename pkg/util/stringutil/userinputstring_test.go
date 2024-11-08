package stringutil

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUserInputString(t *testing.T) {
	Convey("TrimSpace", t, func() {
		s := UserInputString{
			UnsafeString: "\t  hi \n",
		}
		So(s.TrimSpace(), ShouldResemble, "hi")
	})

	Convey("UnmarshalText", t, func() {
		jsonStr := "{ \"teststr\": \" userinputstr \\n \" }"

		type testStruct struct {
			TestStr UserInputString `json:"teststr"`
		}

		var s testStruct

		err := json.Unmarshal([]byte(jsonStr), &s)

		So(err, ShouldBeNil)
		So(s.TestStr.UnsafeString, ShouldEqual, " userinputstr \n ")
		So(s.TestStr.TrimSpace(), ShouldEqual, "userinputstr")
	})

	Convey("MarshalText", t, func() {
		type testStruct struct {
			TestStr UserInputString `json:"teststr"`
		}

		result, err := json.Marshal(testStruct{
			TestStr: UserInputString{
				UnsafeString: " userinputstr \n ",
			},
		})

		So(err, ShouldBeNil)
		So(string(result), ShouldEqual, "{\"teststr\":\" userinputstr \\n \"}")
	})
}
