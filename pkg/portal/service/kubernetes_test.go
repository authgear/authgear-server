package service_test

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/portal/service"
)

func TestKubernetesGenerateResources(t *testing.T) {
	Convey("GenerateResources", t, func() {
		data := &service.ResourceTemplateData{
			AppID:    "accounts",
			DomainID: "domainid",
			IsCustom: false,
			Host:     "accounts.example.com",
		}

		Convey("empty template results in no ingresses", func() {
			_, err := service.GenerateResources(data, []byte(``))
			So(err, ShouldBeNil)
		})

		Convey("single document results in 1 ingress", func() {
			resources, err := service.GenerateResources(data, []byte(`
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
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: 'tls-{{ .Host }}'
`))
			So(err, ShouldBeNil)
			So(len(resources), ShouldEqual, 1)

			resourcesJSON, _ := json.Marshal(resources[0].Object)
			So(string(resourcesJSON), ShouldEqual, `{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"labels":{"authgear.com/app-id":"accounts"},"name":"accounts.example.com","namespace":"authgear"},"spec":{"rules":[{"host":"accounts.example.com","http":{"paths":[{"backend":{"serviceName":"authgear","servicePort":"http"},"path":"/"}]}}],"tls":[{"hosts":["accounts.example.com"],"secretName":"tls-accounts.example.com"}]}}`)

		})

		Convey("ignore empty documents", func() {
			resources, err := service.GenerateResources(data, []byte(`
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
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: 'tls-{{ .Host }}'
---
---
---`))
			So(err, ShouldBeNil)
			So(len(resources), ShouldEqual, 1)

			resourcesJSON, _ := json.Marshal(resources[0].Object)
			So(string(resourcesJSON), ShouldEqual, `{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"labels":{"authgear.com/app-id":"accounts"},"name":"accounts.example.com","namespace":"authgear"},"spec":{"rules":[{"host":"accounts.example.com","http":{"paths":[{"backend":{"serviceName":"authgear","servicePort":"http"},"path":"/"}]}}],"tls":[{"hosts":["accounts.example.com"],"secretName":"tls-accounts.example.com"}]}}`)
		})

		Convey("n documents result in n ingresses", func() {
			resources, err := service.GenerateResources(data, []byte(`
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
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: 'tls-{{ .Host }}'
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: authgear
  name: '{{ .Host }}-ingress-2'
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
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: 'tls-{{ .Host }}'
---
`))
			So(err, ShouldBeNil)
			So(len(resources), ShouldEqual, 2)

			resourcesJSON, _ := json.Marshal(resources[0].Object)
			So(string(resourcesJSON), ShouldEqual, `{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"labels":{"authgear.com/app-id":"accounts"},"name":"accounts.example.com","namespace":"authgear"},"spec":{"rules":[{"host":"accounts.example.com","http":{"paths":[{"backend":{"serviceName":"authgear","servicePort":"http"},"path":"/"}]}}],"tls":[{"hosts":["accounts.example.com"],"secretName":"tls-accounts.example.com"}]}}`)

			resourcesJSON, _ = json.Marshal(resources[1].Object)
			So(string(resourcesJSON), ShouldEqual, `{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"labels":{"authgear.com/app-id":"accounts"},"name":"accounts.example.com-ingress-2","namespace":"authgear"},"spec":{"rules":[{"host":"accounts.example.com","http":{"paths":[{"backend":{"serviceName":"authgear","servicePort":"http"},"path":"/"}]}}],"tls":[{"hosts":["accounts.example.com"],"secretName":"tls-accounts.example.com"}]}}`)
		})

		Convey("n documents result in ingress and cert", func() {
			resources, err := service.GenerateResources(data, []byte(`
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: authgear
  name: https-{{ .Host }}
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
  tls:
  - hosts:
    - '{{ .Host }}'
    secretName: 'tls-{{ .Host }}'
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  labels:
    authgear.com/app-id: {{ .AppID }}
  name: cert-{{ .Host }}
spec:
  dnsNames:
  - {{ .Host }}
  issuerRef:
    kind: Issuer
    name: letsencrypt-http01
  secretName: tls-{{ .Host }}
`))
			So(err, ShouldBeNil)
			So(len(resources), ShouldEqual, 2)

			resourcesJSON, _ := json.Marshal(resources[0].Object)
			So(string(resourcesJSON), ShouldEqual, `{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"labels":{"authgear.com/app-id":"accounts"},"name":"https-accounts.example.com","namespace":"authgear"},"spec":{"rules":[{"host":"accounts.example.com","http":{"paths":[{"backend":{"serviceName":"authgear","servicePort":"http"},"path":"/"}]}}],"tls":[{"hosts":["accounts.example.com"],"secretName":"tls-accounts.example.com"}]}}`)

			resourcesJSON, _ = json.Marshal(resources[1].Object)
			So(string(resourcesJSON), ShouldEqual, `{"apiVersion":"cert-manager.io/v1","kind":"Certificate","metadata":{"labels":{"authgear.com/app-id":"accounts"},"name":"cert-accounts.example.com"},"spec":{"dnsNames":["accounts.example.com"],"issuerRef":{"kind":"Issuer","name":"letsencrypt-http01"},"secretName":"tls-accounts.example.com"}}`)
		})
	})

}
