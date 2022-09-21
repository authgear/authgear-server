package web3_test

import (
	"net/url"
	"testing"

	"github.com/authgear/authgear-server/pkg/util/web3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEIP681(t *testing.T) {
	Convey("ceip681", t, func() {
		Convey("parse eip681", func() {
			test := func(uri string, expected *web3.EIP681) {
				eip681, err := web3.ParseEIP681(uri)
				So(err, ShouldBeNil)
				So(eip681, ShouldResemble, expected)
			}

			test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1", &web3.EIP681{
				ChainID: 1,
				Address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
			})

			test("ethereum:0xdc0479cc5bba033b3e7de9f178607150b3abce1f@1231", &web3.EIP681{
				ChainID: 1231,
				Address: "0xdc0479cc5bba033b3e7de9f178607150b3abce1f",
			})

			test("ethereum:0x71c7656ec7ab88b098defb751b7401b5f6d8976f@23821", &web3.EIP681{
				ChainID: 23821,
				Address: "0x71c7656ec7ab88b098defb751b7401b5f6d8976f",
			})

			test("ethereum:0x71C7656EC7AB88B098DEFB751B7401B5F6D8976F@23821", &web3.EIP681{
				ChainID: 23821,
				Address: "0x71c7656ec7ab88b098defb751b7401b5f6d8976f",
			})
		})

		Convey("convert eip681 to url", func() {
			test := func(eip681 *web3.EIP681, expected *url.URL) {
				uri := eip681.URL()
				So(uri, ShouldResemble, expected)
			}

			test(&web3.EIP681{
				ChainID: 1,
				Address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
			})

			test(&web3.EIP681{
				ChainID: 1231,
				Address: "0xdc0479cc5bba033b3e7de9f178607150b3abce1f",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xdc0479cc5bba033b3e7de9f178607150b3abce1f@1231",
			})

			test(&web3.EIP681{
				ChainID: 23821,
				Address: "0x71c7656ec7ab88b098defb751b7401b5f6d8976f",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0x71c7656ec7ab88b098defb751b7401b5f6d8976f@23821",
			})
		})
	})
}
