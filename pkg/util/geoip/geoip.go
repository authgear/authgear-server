package geoip

import (
	_ "embed"
	"net"

	"github.com/oschwald/geoip2-golang"
)

//go:embed GeoLite2-Country.mmdb
var GeoLite2_Country_mmdb []byte

// reader has a Close() method, but it is intentionally that it is never closed.
var reader *geoip2.Reader

func init() {
	reader = open()
}

type Info struct {
	CountryCode        string
	EnglishCountryName string
}

func open() *geoip2.Reader {
	reader, err := geoip2.FromBytes(GeoLite2_Country_mmdb)
	if err != nil {
		panic(err)
	}
	return reader
}

func IPString(ipString string) (info *Info, ok bool) {
	ip := net.ParseIP(ipString)
	if ip == nil {
		return
	}
	country, err := reader.Country(ip)
	if err != nil {
		return
	}

	info = &Info{
		CountryCode:        country.Country.IsoCode,
		EnglishCountryName: country.Country.Names["en"],
	}
	ok = true
	return
}
