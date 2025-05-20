package internal

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDockerPublishPortOnAllInterfaces(t *testing.T) {
	Convey("dockerPublishPortOnAllInterfaces", t, func() {
		Convey("should format port string correctly", func() {
			testCases := []struct {
				input    int
				expected string
			}{
				{80, "80:80"},
				{443, "443:443"},
				{8080, "8080:8080"},
				{1234, "1234:1234"},
				{0, "0:0"},             // Edge case: zero port
				{65535, "65535:65535"}, // Edge case: max port number
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("when input is '%v'", tc.input), func() {
					result := dockerPublishPortOnAllInterfaces(tc.input)
					So(result, ShouldEqual, tc.expected)
				})
			}
		})
	})
}
