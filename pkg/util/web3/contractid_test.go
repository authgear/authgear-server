package web3_test

import (
	"net/url"
	"testing"

	"github.com/authgear/authgear-server/pkg/util/web3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("contract_id", t, func() {
		Convey("parse ethereum", func() {
			test := func(uri string, expected *web3.ContractID) {
				contractID, err := web3.ParseContractID(uri)
				So(err, ShouldBeNil)
				So(contractID, ShouldResemble, expected)
			}

			test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1", &web3.ContractID{
				Blockchain:      "ethereum",
				Network:         "1",
				ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
			})
		})

		Convey("convert contract_id to url", func() {
			test := func(contractID *web3.ContractID, expected *url.URL) {
				uri, err := contractID.URL()
				So(err, ShouldBeNil)
				So(uri, ShouldResemble, expected)
			}

			test(&web3.ContractID{
				Blockchain:      "ethereum",
				Network:         "1",
				ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
			})
		})
	})
}
