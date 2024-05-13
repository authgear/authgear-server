package web3_test

import (
	"encoding/json"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/web3"
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
				Address:    "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
				Query:      url.Values{},
			})

			test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1?token_ids=0x1&token_ids=0x2", &web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
				Query: url.Values{
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
				Address:    "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
			}, &url.URL{
				Scheme: "ethereum",
				Opaque: "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D@1",
			})

			test(&web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
				Query: url.Values{
					"token_ids": []string{
						"0x1", "0x2",
					},
				},
			}, &url.URL{
				Scheme:   "ethereum",
				Opaque:   "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D@1",
				RawQuery: "token_ids=0x1&token_ids=0x2",
			})
		})

		Convey("create new contract_id", func() {
			test := func(blockchain string, network string, address string, query url.Values, expected *web3.ContractID) {
				cid, err := web3.NewContractID(blockchain, network, address, query)
				So(err, ShouldBeNil)
				So(cid, ShouldResemble, expected)
			}

			test("ethereum", "1", "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d", nil, &web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
			})

			test("ethereum", "1", "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d", url.Values{
				"token_ids": []string{
					"0x1", "0x2",
				},
			}, &web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
				Query: url.Values{
					"token_ids": []string{
						"0x1", "0x2",
					},
				},
			})
		})

		Convey("non-pointer ContractID and JSON", func() {
			type A struct {
				ContractID web3.ContractID
			}
			a := A{
				ContractID: web3.ContractID{
					Blockchain: "ethereum",
					Network:    "1",
					Address:    web3.EIP55("0xEC7F0e0C2B7a356b5271D13e75004705977Fd010"),
					Query: url.Values{
						"a": []string{"b"},
					},
				},
			}

			jsonBytes, err := json.Marshal(a)
			So(err, ShouldBeNil)
			So(string(jsonBytes), ShouldEqual, `{"ContractID":"ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b"}`)

			var b A
			err = json.Unmarshal(jsonBytes, &b)
			So(err, ShouldBeNil)
			So(a.ContractID.String(), ShouldEqual, "ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b")
		})

		Convey("pointer to ContractID and JSON", func() {
			type A struct {
				ContractID *web3.ContractID
			}
			a := A{
				ContractID: &web3.ContractID{
					Blockchain: "ethereum",
					Network:    "1",
					Address:    web3.EIP55("0xEC7F0e0C2B7a356b5271D13e75004705977Fd010"),
					Query: url.Values{
						"a": []string{"b"},
					},
				},
			}

			jsonBytes, err := json.Marshal(a)
			So(err, ShouldBeNil)
			So(string(jsonBytes), ShouldEqual, `{"ContractID":"ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b"}`)

			var b A
			err = json.Unmarshal(jsonBytes, &b)
			So(err, ShouldBeNil)
			So(a.ContractID.String(), ShouldEqual, "ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b")

			var c A
			err = json.Unmarshal([]byte(`{"ContractID":null}`), &c)
			So(err, ShouldBeNil)
			So(c.ContractID, ShouldBeNil)
		})

		Convey("Clone and StringQuery", func() {
			a := web3.ContractID{
				Blockchain: "ethereum",
				Network:    "1",
				Address:    web3.EIP55("0xEC7F0e0C2B7a356b5271D13e75004705977Fd010"),
				Query: url.Values{
					"a": []string{"b"},
				},
			}

			b := a.Clone()
			b.Query.Set("a", "c")

			So(a.String(), ShouldEqual, "ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b")
			So(b.String(), ShouldEqual, "ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=c")

			c := a.StripQuery()
			So(a.String(), ShouldEqual, "ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1?a=b")
			So(c.String(), ShouldEqual, "ethereum:0xEC7F0e0C2B7a356b5271D13e75004705977Fd010@1")
		})
	})
}
