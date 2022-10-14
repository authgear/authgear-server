package web3_test

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/util/hexstring"
	"github.com/authgear/authgear-server/pkg/util/web3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEIP55(t *testing.T) {
	Convey("eip55", t, func() {
		Convey("parse eip55", func() {
			test := func(address string, expected string) {
				eip55, err := web3.NewEIP55(address)
				So(err, ShouldBeNil)
				So(eip55, ShouldEqual, expected)
			}

			test("0xec7f0e0c2b7a356b5271d13e75004705977fd010", "0xEC7F0e0C2B7a356b5271D13e75004705977Fd010")

			test("0xEC7F0E0C2B7A356B5271D13E75004705977FD010", "0xEC7F0e0C2B7a356b5271D13e75004705977Fd010")

			test("0x6e992c5b27db78915ffecf227796b8983880cc9a", "0x6E992c5b27DB78915ffECF227796b8983880cc9A")

			test("0x6E992C5B27DB78915FFECF227796B8983880CC9A", "0x6E992c5b27DB78915ffECF227796b8983880cc9A")
		})

		Convey("convert eip55 to hexstring", func() {
			test := func(eip55 web3.EIP55, expected hexstring.T) {
				str, err := eip55.ToHexstring()
				So(err, ShouldBeNil)
				So(str, ShouldEqual, expected)
			}

			test("0xEC7F0e0C2B7a356b5271D13e75004705977Fd010", "0xec7f0e0c2b7a356b5271d13e75004705977fd010")

			test("0x6E992c5b27DB78915ffECF227796b8983880cc9A", "0x6e992c5b27db78915ffecf227796b8983880cc9a")
		})

		Convey("convert to eip55 from hexstring", func() {
			test := func(hex hexstring.T, expected web3.EIP55) {
				eip55, err := web3.NewEIP55FromHexstring(hex)
				So(err, ShouldBeNil)
				So(eip55, ShouldEqual, expected)
			}

			test("0xec7f0e0c2b7a356b5271d13e75004705977fd010", "0xEC7F0e0C2B7a356b5271D13e75004705977Fd010")

			test("0x6e992c5b27db78915ffecf227796b8983880cc9a", "0x6E992c5b27DB78915ffECF227796b8983880cc9A")
		})
	})
}
