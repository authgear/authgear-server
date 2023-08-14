package workflow2

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPartiallyMergeFrom(t *testing.T) {
	Convey("PartiallyMergeFrom", t, func() {
		test := func(o1 *SessionOptions, o2 *SessionOptions, expected *SessionOptions) {
			actual := o1.PartiallyMergeFrom(o2)
			So(actual, ShouldResemble, expected)
		}

		test(nil, nil, &SessionOptions{})
		test(&SessionOptions{}, nil, &SessionOptions{})
		test(nil, &SessionOptions{}, &SessionOptions{})

		// The receiver is cloned.
		test(&SessionOptions{
			ClientID:                 "a",
			RedirectURI:              "b",
			SuppressIDPSessionCookie: true,
			State:                    "c",
			XState:                   "d",
			UILocales:                "e",
		}, nil, &SessionOptions{
			ClientID:                 "a",
			RedirectURI:              "b",
			SuppressIDPSessionCookie: true,
			State:                    "c",
			XState:                   "d",
			UILocales:                "e",
		})

		// Merge is partial.
		test(nil, &SessionOptions{
			ClientID:                 "a",
			RedirectURI:              "b",
			SuppressIDPSessionCookie: true,
			State:                    "c",
			XState:                   "d",
			UILocales:                "e",
		}, &SessionOptions{
			ClientID:  "a",
			State:     "c",
			XState:    "d",
			UILocales: "e",
		})

		// Only non-zero values are merged.
		test(&SessionOptions{
			ClientID:                 "a",
			RedirectURI:              "b",
			SuppressIDPSessionCookie: true,
			State:                    "c",
			XState:                   "d",
			UILocales:                "e",
		}, &SessionOptions{
			State:     "cc",
			XState:    "dd",
			UILocales: "ee",
		}, &SessionOptions{
			ClientID:                 "a",
			RedirectURI:              "b",
			SuppressIDPSessionCookie: true,
			State:                    "cc",
			XState:                   "dd",
			UILocales:                "ee",
		})
	})
}
