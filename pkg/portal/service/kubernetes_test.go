package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	fakedcertmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/fake"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeddynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/restmapper"
	coretesting "k8s.io/client-go/testing"
	"sigs.k8s.io/yaml"

	. "github.com/smartystreets/goconvey/convey"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func newKubernetesWithDynamicClient(ingressTemplatePath string) (*service.Kubernetes, error) {
	fakeDiscoveryClient := &fakediscovery.FakeDiscovery{Fake: &coretesting.Fake{}}
	fakeDiscoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "networking.k8s.io/v1",
			APIResources: []metav1.APIResource{
				{Name: "ingresses", Namespaced: true, Kind: "Ingress"},
			},
		},
		{
			GroupVersion: "cert-manager.io/v1",
			APIResources: []metav1.APIResource{
				{Name: "certificates", Namespaced: true, Kind: "Certificate"},
			},
		},
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
			},
		},
	}

	restMapperRes, err := restmapper.GetAPIGroupResources(fakeDiscoveryClient)
	if err != nil {
		return nil, fmt.Errorf("unexpected error while constructing resource list from fake discovery client: %v", err)
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(restMapperRes)
	scheme := runtime.NewScheme()
	fakeDynamicClient := fakeddynamic.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		schema.GroupVersionResource{
			Group:    "networking.k8s.io",
			Version:  "v1",
			Resource: "ingresses",
		}: "IngressList",
		schema.GroupVersionResource{
			Group:    "cert-manager.io",
			Version:  "v1",
			Resource: "certificates",
		}: "CertificateList",
	})

	return &service.Kubernetes{
		AppConfig: &portalconfig.AppConfig{
			Kubernetes: portalconfig.AppKubernetesConfig{
				IngressTemplateFile: ingressTemplatePath,
			},
		},
		DiscoveryRESTMapper: restMapper,
		DynamicClient:       fakeDynamicClient,
		Namespace:           "test-namespace",
	}, nil
}

func newKubernetesWithResources(namespace string, objects []runtime.Object, certs []runtime.Object) (*service.Kubernetes, error) {
	return &service.Kubernetes{
		Client:            fake.NewSimpleClientset(objects...),
		CertManagerClient: fakedcertmanager.NewSimpleClientset(certs...),
		Namespace:         namespace,
		Logger:            service.KubernetesLogger{log.Null},
	}, nil
}

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

	Convey("CreateResourcesForDomain", t, func() {

		Convey("create 2 ingresses and 1 cert from template", func() {
			kube, err := newKubernetesWithDynamicClient("testdata/ingress.tpl.yaml")
			So(err, ShouldBeNil)

			err = kube.CreateResourcesForDomain("app-id-1", "domain-id-1", "test.example.com", true)
			So(err, ShouldBeNil)

			ingresses, err := kube.DynamicClient.
				Resource(schema.GroupVersionResource{
					Group:    "networking.k8s.io",
					Version:  "v1",
					Resource: "ingresses",
				}).
				Namespace(kube.Namespace).
				List(context.TODO(), metav1.ListOptions{})
			So(err, ShouldBeNil)
			So(len(ingresses.Items), ShouldEqual, 2)

			certs, err := kube.DynamicClient.
				Resource(schema.GroupVersionResource{
					Group:    "cert-manager.io",
					Version:  "v1",
					Resource: "certificates",
				}).
				Namespace(kube.Namespace).
				List(context.TODO(), metav1.ListOptions{})
			So(err, ShouldBeNil)
			So(len(certs.Items), ShouldEqual, 1)

			objects := append(ingresses.Items, certs.Items...)

			b, _ := yaml.Marshal(objects)
			result, err := ioutil.ReadFile("testdata/ingress_result.yaml")
			So(string(b), ShouldEqual, string(result))
			So(err, ShouldBeNil)
		})

		Convey("Only ingress and cert resources are supported", func() {
			kube, err := newKubernetesWithDynamicClient("testdata/invalid_template.tpl.yaml")
			So(err, ShouldBeNil)

			err = kube.CreateResourcesForDomain("app-id-1", "domain-id-1", "test.example.com", true)
			So(err, ShouldBeError, "k8s gvk type is not supported: /v1, Kind=Pod")
		})

	})

	Convey("DeleteResourcesForDomain", t, func() {
		kube, err := newKubernetesWithResources(
			"test-namespace",
			[]runtime.Object{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test-namespace", Name: "pod1"},
				},
				&networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "app-domain-1",
						Labels: map[string]string{
							service.LabelDomainID: "domain-id-1",
						},
					},
				},
				&networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "app-domain-2",
						Labels: map[string]string{
							service.LabelDomainID: "domain-id-2",
						},
					},
				},
				&networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "infrastructure-ingress",
					},
				},
			},
			[]runtime.Object{
				&certmanagerv1.Certificate{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "cert-domain-1",
						Labels: map[string]string{
							service.LabelDomainID: "domain-id-1",
						},
					},
				},
				&certmanagerv1.Certificate{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "cert-domain-2",
						Labels: map[string]string{
							service.LabelDomainID: "domain-id-2",
						},
					},
				},
				&certmanagerv1.Certificate{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "infrastructure-domain",
					},
				},
			},
		)
		So(err, ShouldBeNil)

		err = kube.DeleteResourcesForDomain("domain-id-1")
		So(err, ShouldBeNil)

		podList, err := kube.Client.CoreV1().Pods("test-namespace").List(context.TODO(), metav1.ListOptions{})
		So(err, ShouldBeNil)
		So(len(podList.Items), ShouldEqual, 1)

		ingressesList, err := kube.Client.NetworkingV1().Ingresses("test-namespace").List(context.Background(), metav1.ListOptions{})
		So(err, ShouldBeNil)
		So(len(ingressesList.Items), ShouldEqual, 2)

		certList, err := kube.CertManagerClient.CertmanagerV1().Certificates("test-namespace").List(context.TODO(), metav1.ListOptions{})
		So(err, ShouldBeNil)
		So(len(certList.Items), ShouldEqual, 2)

	})

}
