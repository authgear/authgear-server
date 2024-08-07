package saml

import (
	"bytes"
	"text/template"
	"time"

	crewjamsaml "github.com/crewjam/saml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

const MetadataValidDuration = time.Hour * 24

type Service struct {
	Clock                 clock.Clock
	AppID                 config.AppID
	SAMLEnvironmentConfig config.SAMLEnvironmentConfig
}

func (s *Service) idpEntityID() string {
	idpEntityIdTemplate, err := template.New("").Parse(s.SAMLEnvironmentConfig.IdPEntityIDTemplate)
	if err != nil {
		panic(err)
	}
	var idpEntityIDBytes bytes.Buffer
	err = idpEntityIdTemplate.Execute(&idpEntityIDBytes, map[string]interface{}{
		"app_id": s.AppID,
	})
	if err != nil {
		panic(err)
	}

	return idpEntityIDBytes.String()
}

func (s *Service) IdPMetadata() *Metadata {

	descriptor := crewjamsaml.EntityDescriptor{
		EntityID: s.idpEntityID(),
		// We need this because `omitempty` does not work with time.Time, so it always output '0001-01-01T00:00:00Z'
		// So give it a value
		ValidUntil: s.Clock.NowUTC().Add(MetadataValidDuration),
		IDPSSODescriptors: []crewjamsaml.IDPSSODescriptor{
			{
				SSODescriptor: crewjamsaml.SSODescriptor{
					RoleDescriptor: crewjamsaml.RoleDescriptor{
						ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
						KeyDescriptors:             []crewjamsaml.KeyDescriptor{
							// TODO
						},
					},
					NameIDFormats: []crewjamsaml.NameIDFormat{
						// TODO
					},
				},
				SingleSignOnServices: []crewjamsaml.Endpoint{
					// TODO
				},
			},
		},
	}

	return &Metadata{
		descriptor,
	}
}
