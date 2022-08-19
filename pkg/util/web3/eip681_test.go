package web3_test

import (
	"net/url"
	"testing"

	"github.com/authgear/authgear-server/pkg/util/web3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("contract_id", t, func() {
		Convey("parse eip681", func() {
			test := func(uri string, expected *web3.EIP681) {
				eip681, err := web3.ParseEIP681(uri)
				So(err, ShouldBeNil)
				So(eip681, ShouldResemble, expected)
			}

			test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1", &web3.EIP681{
				ChainID:         1,
				ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
			})

			test("ethereum:0xdc0479cc5bba033b3e7de9f178607150b3abce1f@1231", &web3.EIP681{
				ChainID:         1231,
				ContractAddress: "0xdc0479cc5bba033b3e7de9f178607150b3abce1f",
			})
		})

		Convey("convert eip681 to url", func() {
			test := func(eip681 *web3.EIP681, expected *url.URL) {
				uri := eip681.URL()
				So(uri, ShouldResemble, expected)
			}

			test(&web3.EIP681{
				ChainID:         1,
				ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
			})

			test(&web3.EIP681{
				ChainID:         1231,
				ContractAddress: "0xdc0479cc5bba033b3e7de9f178607150b3abce1f",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xdc0479cc5bba033b3e7de9f178607150b3abce1f@1231",
			})
		})
	})
}
