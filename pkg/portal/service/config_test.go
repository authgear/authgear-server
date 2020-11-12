package service_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/portal/service"
)

func TestGenerateIngresses(t *testing.T) {
	Convey("GenerateIngresses", t, func() {
		data := &service.IngressTemplateData{
			AppID:    "accounts",
			DomainID: "domainid",
			IsCustom: false,
			Host:     "accounts.example.com",
		}

		Convey("empty template results in no ingresses", func() {
			_, err := service.GenerateIngresses(data, []byte(``))
			So(err, ShouldBeNil)
		})

		Convey("single document results in 1 ingress", func() {
			ingresses, err := service.GenerateIngresses(data, []byte(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: authgear
  name: {{ .Host }}
  labels:
    authgear.com/app-id: {{ .AppID }}
{{ if .IsCustom }}
    authgear.com/domain-id: {{ .DomainID }}
{{ end }}
spec:
  rules:
  - host: '{{ .Host }}'
    http:
      paths:
      - backend:
          serviceName: authgear
          servicePort: http
        path: /
{{ if .TLSSecretName }}
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: '{{ .TLSSecretName }}'
{{ end }}
`))
			So(err, ShouldBeNil)
			So(len(ingresses), ShouldEqual, 1)
			So(ingresses[0].Spec.Rules[0].Host, ShouldEqual, "accounts.example.com")
		})

		Convey("ignore empty documents", func() {
			ingresses, err := service.GenerateIngresses(data, []byte(`
---
---
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: authgear
  name: {{ .Host }}
  labels:
    authgear.com/app-id: {{ .AppID }}
{{ if .IsCustom }}
    authgear.com/domain-id: {{ .DomainID }}
{{ end }}
spec:
  rules:
  - host: '{{ .Host }}'
    http:
      paths:
      - backend:
          serviceName: authgear
          servicePort: http
        path: /
{{ if .TLSSecretName }}
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: '{{ .TLSSecretName }}'
{{ end }}
---
---
---
`))
			So(err, ShouldBeNil)
			So(len(ingresses), ShouldEqual, 1)
			So(ingresses[0].Spec.Rules[0].Host, ShouldEqual, "accounts.example.com")
		})

		Convey("n documents result in n ingresses", func() {
			ingresses, err := service.GenerateIngresses(data, []byte(`
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: authgear
  name: {{ .Host }}
  labels:
    authgear.com/app-id: {{ .AppID }}
{{ if .IsCustom }}
    authgear.com/domain-id: {{ .DomainID }}
{{ end }}
spec:
  rules:
  - host: '{{ .Host }}'
    http:
      paths:
      - backend:
          serviceName: authgear
          servicePort: http
        path: /
{{ if .TLSSecretName }}
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: '{{ .TLSSecretName }}'
{{ end }}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: authgear
  name: {{ .Host }}
  labels:
    authgear.com/app-id: {{ .AppID }}
{{ if .IsCustom }}
    authgear.com/domain-id: {{ .DomainID }}
{{ end }}
spec:
  rules:
  - host: '{{ .Host }}'
    http:
      paths:
      - backend:
          serviceName: authgear
          servicePort: http
        path: /
{{ if .TLSSecretName }}
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: '{{ .TLSSecretName }}'
{{ end }}
---
`))
			So(err, ShouldBeNil)
			So(len(ingresses), ShouldEqual, 2)
			So(ingresses[0].Spec.Rules[0].Host, ShouldEqual, "accounts.example.com")
			So(ingresses[1].Spec.Rules[0].Host, ShouldEqual, "accounts.example.com")
		})
	})
}
