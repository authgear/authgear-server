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
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
				Query:      nil,
			})

			test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1?token_ids=0x1&token_ids=0x2", &web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
				Query: &url.Values{
					"token_ids": []string{
						"0x1", "0x2",
					},
				},
			})
		})

		Convey("convert contract_id to url", func() {
			test := func(contractID *web3.ContractID, expected *url.URL) {
				uri, err := contractID.URL()
				So(err, ShouldBeNil)
				So(uri, ShouldResemble, expected)
			}

			test(&web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
			})

			test(&web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
				Query: &url.Values{
					"token_ids": []string{
						"0x1", "0x2",
					},
				},
			}, &url.URL{
				Scheme:   "ethereum",
				Opaque:   "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
				RawQuery: "token_ids=0x1&token_ids=0x2",
			})
		})

		Convey("create new contract_id", func() {
			test := func(blockchain string, network string, address string, query *url.Values, expected *web3.ContractID) {
				cid, err := web3.NewContractID(blockchain, network, address, query)
				So(err, ShouldBeNil)
				So(cid, ShouldResemble, expected)
			}

			test("ethereum", "1", "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d", nil, &web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
				Query:      nil,
			})

			test("ethereum", "1", "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d", &url.Values{
				"token_ids": []string{
					"0x1", "0x2",
				},
			}, &web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
				Query: &url.Values{
					"token_ids": []string{
						"0x1", "0x2",
					},
				},
			})
		})
	})
}
