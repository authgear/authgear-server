package web3_test

import (
	"encoding/json"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/web3"
)

func TestEIP681(t *testing.T) {
	Convey("eip681", t, func() {
		Convey("parse eip681", func() {
			test := func(uri string, expected *web3.EIP681) {
				eip681, err := web3.ParseEIP681(uri)
				So(err, ShouldBeNil)
				So(eip681, ShouldResemble, expected)
			}

			test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1", &web3.EIP681{
				ChainID: 1,
				Address: "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
				Query:   url.Values{},
			})

			test("ethereum:0xdc0479cc5bba033b3e7de9f178607150b3abce1f@1231", &web3.EIP681{
				ChainID: 1231,
				Address: "0xdC0479CC5BbA033B3e7De9F178607150B3AbCe1f",
				Query:   url.Values{},
			})

			test("ethereum:0x71c7656ec7ab88b098defb751b7401b5f6d8976f@23821", &web3.EIP681{
				ChainID: 23821,
				Address: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
				Query:   url.Values{},
			})

			test("ethereum:0x71C7656EC7AB88B098DEFB751B7401B5F6D8976F@23821", &web3.EIP681{
				ChainID: 23821,
				Address: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
				Query:   url.Values{},
			})

			test("ethereum:0x71c7656ec7ab88b098defb751b7401b5f6d8976f@23821?token_ids=0x1&token_ids=0x2", &web3.EIP681{
				ChainID: 23821,
				Address: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
				Query: url.Values{
					"token_ids": []string{
						"0x1", "0x2",
					},
				},
			})
		})

		Convey("convert eip681 to url", func() {
			test := func(eip681 *web3.EIP681, expected *url.URL) {
				uri := eip681.URL()
				So(uri, ShouldResemble, expected)
			}

			test(&web3.EIP681{
				ChainID: 1,
				Address: "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D@1",
			})

			test(&web3.EIP681{
				ChainID: 1231,
				Address: "0xdC0479CC5BbA033B3e7De9F178607150B3AbCe1f",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xdC0479CC5BbA033B3e7De9F178607150B3AbCe1f@1231",
			})

			test(&web3.EIP681{
				ChainID: 23821,
				Address: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F@23821",
			})

			test(&web3.EIP681{
				ChainID: 23821,
				Address: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
				Query: url.Values{
					"token_ids": []string{
						"0x1", "0x2",
					},
				},
			}, &url.URL{
				Scheme:   "ethereum",
				Opaque:   "0x71C7656EC7ab88b098defB751B7401B5f6d8976F@23821",
				RawQuery: "token_ids=0x1&token_ids=0x2",
			})
		})

		Convey("non-pointer EIP681 and JSON", func() {
			type A struct {
				EIP681 web3.EIP681
			}
			a := A{
				EIP681: web3.EIP681{
					ChainID: 1,
					Address: web3.EIP55("0xEC7F0e0C2B7a356b5271D13e75004705977Fd010"),
					Query: url.Values{
						"a": []string{"b"},
					},
				},
			}

			jsonBytes, err := json.Marshal(a)
			So(err, ShouldBeNil)
			So(string(jsonBytes), ShouldEqual, `{"EIP681":"ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b"}`)

			var b A
			err = json.Unmarshal(jsonBytes, &b)
			So(err, ShouldBeNil)
			So(a.EIP681.String(), ShouldEqual, "ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b")
		})

		Convey("pointer to EIP681 and JSON", func() {
			type A struct {
				EIP681 *web3.EIP681
			}
			a := A{
				EIP681: &web3.EIP681{
					ChainID: 1,
					Address: web3.EIP55("0xEC7F0e0C2B7a356b5271D13e75004705977Fd010"),
					Query: url.Values{
						"a": []string{"b"},
					},
				},
			}

			jsonBytes, err := json.Marshal(a)
			So(err, ShouldBeNil)
			So(string(jsonBytes), ShouldEqual, `{"EIP681":"ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b"}`)

			var b A
			err = json.Unmarshal(jsonBytes, &b)
			So(err, ShouldBeNil)
			So(a.EIP681.String(), ShouldEqual, "ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b")

			var c A
			err = json.Unmarshal([]byte(`{"EIP681":null}`), &c)
			So(err, ShouldBeNil)
			So(c.EIP681, ShouldBeNil)
		})
	})
}
