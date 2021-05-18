package geoip

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

var DefaultDatabase *Database

func init() {
	DefaultDatabase, _ = Open("GeoLite2-Country.mmdb")
}

type Info struct {
	CountryCode        string
	EnglishCountryName string
}

type Database struct {
	reader *geoip2.Reader
}

func Open(path string) (*Database, error) {
	reader, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &Database{
		reader: reader,
	}, nil
}

func (d *Database) Close() error {
	return d.reader.Close()
}

func (d *Database) IPString(ipString string) (info *Info, ok bool) {
	if d == nil {
		return
	}
	ip := net.ParseIP(ipString)
	if ip == nil {
		return
	}
	country, err := d.reader.Country(ip)
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
