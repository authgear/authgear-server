package samlprotocol

import (
	"encoding/xml"
)

func (a *LogoutResponse) ToXMLBytes() []byte {
	buf, err := xml.Marshal(a)
	if err != nil {
		panic(err)
	}
	return buf
}
