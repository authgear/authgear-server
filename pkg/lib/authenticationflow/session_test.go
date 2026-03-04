package authenticationflow

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWithSMSOTPSentCountAdded(t *testing.T) {
	Convey("WithSMSOTPSentCountAdded", t, func() {
		Convey("initialises the map and increments count for a new phone", func() {
			s := &Session{}
			updated := s.WithSMSOTPSentCountAdded("+6591234567")
			So(updated.SMSOTPSentCountByPhone["+6591234567"], ShouldEqual, 1)
		})

		Convey("increments existing count for a known phone", func() {
			s := &Session{
				SMSOTPSentCountByPhone: map[string]int{"+6591234567": 2},
			}
			updated := s.WithSMSOTPSentCountAdded("+6591234567")
			So(updated.SMSOTPSentCountByPhone["+6591234567"], ShouldEqual, 3)
		})

		Convey("tracks separate counts per phone", func() {
			s := &Session{
				SMSOTPSentCountByPhone: map[string]int{"+6591111111": 1},
			}
			updated := s.WithSMSOTPSentCountAdded("+6592222222")
			So(updated.SMSOTPSentCountByPhone["+6591111111"], ShouldEqual, 1)
			So(updated.SMSOTPSentCountByPhone["+6592222222"], ShouldEqual, 1)
		})

		Convey("does not mutate the original session", func() {
			s := &Session{
				SMSOTPSentCountByPhone: map[string]int{"+6591234567": 1},
			}
			_ = s.WithSMSOTPSentCountAdded("+6591234567")
			So(s.SMSOTPSentCountByPhone["+6591234567"], ShouldEqual, 1)
		})

		Convey("does not share the underlying map with the original", func() {
			s := &Session{
				SMSOTPSentCountByPhone: map[string]int{"+6591234567": 1},
			}
			updated := s.WithSMSOTPSentCountAdded("+6591234567")
			// Mutate the original map directly and verify the update is isolated.
			s.SMSOTPSentCountByPhone["+6591234567"] = 99
			So(updated.SMSOTPSentCountByPhone["+6591234567"], ShouldEqual, 2)
		})
	})
}

func TestWithSMSOTPVerifiedCountAdded(t *testing.T) {
	Convey("WithSMSOTPVerifiedCountAdded", t, func() {
		Convey("initialises the map and increments count for a new phone", func() {
			s := &Session{}
			updated := s.WithSMSOTPVerifiedCountAdded("+6591234567")
			So(updated.SMSOTPVerifiedCountByPhone["+6591234567"], ShouldEqual, 1)
		})

		Convey("increments existing count for a known phone", func() {
			s := &Session{
				SMSOTPVerifiedCountByPhone: map[string]int{"+6591234567": 1},
			}
			updated := s.WithSMSOTPVerifiedCountAdded("+6591234567")
			So(updated.SMSOTPVerifiedCountByPhone["+6591234567"], ShouldEqual, 2)
		})

		Convey("does not mutate the original session", func() {
			s := &Session{
				SMSOTPVerifiedCountByPhone: map[string]int{"+6591234567": 1},
			}
			_ = s.WithSMSOTPVerifiedCountAdded("+6591234567")
			So(s.SMSOTPVerifiedCountByPhone["+6591234567"], ShouldEqual, 1)
		})
	})
}

func TestPatchFrom(t *testing.T) {
	Convey("PatchFrom", t, func() {
		Convey("copies all fields from updated into receiver", func() {
			s := &Session{FlowID: "old-flow"}
			updated := &Session{
				FlowID:                     "new-flow",
				SMSOTPSentCountByPhone:     map[string]int{"+6591234567": 3},
				SMSOTPVerifiedCountByPhone: map[string]int{"+6591234567": 1},
			}
			s.PatchFrom(updated)
			So(s.FlowID, ShouldEqual, "new-flow")
			So(s.SMSOTPSentCountByPhone["+6591234567"], ShouldEqual, 3)
			So(s.SMSOTPVerifiedCountByPhone["+6591234567"], ShouldEqual, 1)
		})
	})
}

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
			Prompt:                   []string{"login"},
			SuppressIDPSessionCookie: true,
			State:                    "c",
			XState:                   "d",
			UILocales:                "e",
		}, nil, &SessionOptions{
			ClientID:                 "a",
			RedirectURI:              "b",
			Prompt:                   []string{"login"},
			SuppressIDPSessionCookie: true,
			State:                    "c",
			XState:                   "d",
			UILocales:                "e",
		})

		// Merge is partial.
		test(nil, &SessionOptions{
			ClientID:                 "a",
			RedirectURI:              "b",
			Prompt:                   []string{"login"},
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
			Prompt:                   []string{"login"},
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
			Prompt:                   []string{"login"},
			SuppressIDPSessionCookie: true,
			State:                    "cc",
			XState:                   "dd",
			UILocales:                "ee",
		})
	})
}
