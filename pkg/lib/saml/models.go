package saml

import (
	"encoding/xml"

	crewjamsaml "github.com/crewjam/saml"
)

type Metadata struct {
	crewjamsaml.EntityDescriptor
}

func (m *Metadata) ToXMLBytes() []byte {
	buf, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		panic(err)
	}
	return buf
}
