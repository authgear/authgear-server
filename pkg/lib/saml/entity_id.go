package saml

import (
	"bytes"
	"text/template"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func RenderSAMLEntityID(cfg config.SAMLEnvironmentConfig, appID string) string {
	idpEntityIdTemplate, err := template.New("").Parse(cfg.IdPEntityIDTemplate)
	if err != nil {
		panic(err)
	}
	var idpEntityIDBytes bytes.Buffer
	err = idpEntityIdTemplate.Execute(&idpEntityIDBytes, map[string]interface{}{
		"app_id": appID,
	})
	if err != nil {
		panic(err)
	}

	return idpEntityIDBytes.String()
}
