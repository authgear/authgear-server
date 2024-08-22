package samlprotocol

import (
	"encoding/xml"
	"time"

	"github.com/beevik/etree"
	crewjamsaml "github.com/crewjam/saml"
)

// Copied from https://github.com/crewjam/saml/blob/main/metadata.go#L53
// The type of ValidUntil is time.Time causing it cannot be omitted
// So we make our own EntityDescriptor
type EntityDescriptor struct {
	XMLName                       xml.Name      `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntityDescriptor"`
	EntityID                      string        `xml:"entityID,attr"`
	ID                            string        `xml:",attr,omitempty"`
	ValidUntil                    *time.Time    `xml:"validUntil,attr,omitempty"`
	CacheDuration                 time.Duration `xml:"cacheDuration,attr,omitempty"`
	Signature                     *etree.Element
	RoleDescriptors               []crewjamsaml.RoleDescriptor               `xml:"RoleDescriptor"`
	IDPSSODescriptors             []crewjamsaml.IDPSSODescriptor             `xml:"IDPSSODescriptor"`
	SPSSODescriptors              []crewjamsaml.SPSSODescriptor              `xml:"SPSSODescriptor"`
	AuthnAuthorityDescriptors     []crewjamsaml.AuthnAuthorityDescriptor     `xml:"AuthnAuthorityDescriptor"`
	AttributeAuthorityDescriptors []crewjamsaml.AttributeAuthorityDescriptor `xml:"AttributeAuthorityDescriptor"`
	PDPDescriptors                []crewjamsaml.PDPDescriptor                `xml:"PDPDescriptor"`
	AffiliationDescriptor         *crewjamsaml.AffiliationDescriptor
	Organization                  *crewjamsaml.Organization
	ContactPerson                 *crewjamsaml.ContactPerson
	AdditionalMetadataLocations   []string `xml:"AdditionalMetadataLocation"`
}

// MarshalXML implements xml.Marshaler
func (m EntityDescriptor) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	type Alias EntityDescriptor
	aux := &struct {
		ValidUntil    *crewjamsaml.RelaxedTime `xml:"validUntil,attr,omitempty"`
		CacheDuration crewjamsaml.Duration     `xml:"cacheDuration,attr,omitempty"`
		*Alias
	}{
		ValidUntil:    nil,
		CacheDuration: crewjamsaml.Duration(m.CacheDuration),
		Alias:         (*Alias)(&m),
	}
	if m.ValidUntil != nil {
		validUntil := crewjamsaml.RelaxedTime(*m.ValidUntil)
		aux.ValidUntil = &validUntil
	}
	return e.Encode(aux)
}

// UnmarshalXML implements xml.Unmarshaler
func (m *EntityDescriptor) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias EntityDescriptor
	aux := &struct {
		ValidUntil    *crewjamsaml.RelaxedTime `xml:"validUntil,attr,omitempty"`
		CacheDuration crewjamsaml.Duration     `xml:"cacheDuration,attr,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := d.DecodeElement(aux, &start); err != nil {
		return err
	}
	if aux.ValidUntil != nil {
		validUntil := time.Time(*aux.ValidUntil)
		m.ValidUntil = &validUntil
	}
	m.CacheDuration = time.Duration(aux.CacheDuration)
	return nil
}

type Metadata struct {
	EntityDescriptor
}

func (m *Metadata) ToXMLBytes() []byte {
	buf, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		panic(err)
	}
	return buf
}
