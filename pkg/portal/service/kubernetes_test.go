package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"testing"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	fakedcertmanager "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/fake"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
)

func newKubernetesWithDynamicClient(ingressTemplatePath string, objects ...runtime.Object) (*service.Kubernetes, error) {
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
	}, objects...)

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

			ctx := context.Background()
			err = kube.CreateResourcesForDomain(ctx, "app-id-1", "domain-id-1", "test.example.com", true)
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
			result, err := os.ReadFile("testdata/ingress_result.yaml")
			So(string(b), ShouldEqual, string(result))
			So(err, ShouldBeNil)
		})

		Convey("Only ingress and cert resources are supported", func() {
			kube, err := newKubernetesWithDynamicClient("testdata/invalid_template.tpl.yaml")
			So(err, ShouldBeNil)

			ctx := context.Background()
			err = kube.CreateResourcesForDomain(ctx, "app-id-1", "domain-id-1", "test.example.com", true)
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

		ctx := context.Background()
		err = kube.DeleteResourcesForDomain(ctx, "domain-id-1")
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

func TestKubernetesApplyResourcesForDomains(t *testing.T) {
	ingressGVR := schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
	certGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}

	newObject := func(apiVersion string, kind string, name string, domainID string, annotations map[string]any, spec map[string]any) *unstructured.Unstructured {
		metadata := map[string]any{
			"name":      name,
			"namespace": "test-namespace",
		}
		if domainID != "" {
			metadata["labels"] = map[string]any{
				service.LabelAppID:    "app-id-1",
				service.LabelDomainID: domainID,
			}
		}
		if annotations != nil {
			metadata["annotations"] = annotations
		}
		return &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata":   metadata,
			"spec":       spec,
		}}
	}

	target := service.DomainResourcesTarget{
		AppID:    "app-id-1",
		DomainID: "domain-id-1",
		Domain:   "test.example.com",
		IsCustom: true,
	}
	rules := []any{map[string]any{"host": "test.example.com"}}
	cert := newObject("cert-manager.io/v1", "Certificate", "cert-test.example.com", "domain-id-1", nil, map[string]any{
		"dnsNames":   []any{"test.example.com"},
		"secretName": "tls-test.example.com",
	})

	Convey("ApplyResourcesForDomains", t, func() {
		ctx := context.Background()

		Convey("creates missing resources, updates changed ones and leaves the others", func() {
			// Created by an older template: legacy class annotation, no ingressClassName.
			legacyIngress := newObject("networking.k8s.io/v1", "Ingress", "nginx-test.example.com", "domain-id-1",
				map[string]any{
					"kubernetes.io/ingress.class":                 "nginx",
					"nginx.ingress.kubernetes.io/proxy-body-size": "10m",
				},
				map[string]any{"rules": rules},
			)
			kube, err := newKubernetesWithDynamicClient("testdata/apply.tpl.yaml", legacyIngress, cert)
			So(err, ShouldBeNil)

			var out bytes.Buffer
			err = kube.ApplyResourcesForDomains(ctx, []service.DomainResourcesTarget{target}, service.ApplyResourcesForDomainsOptions{Output: &out})
			So(err, ShouldBeNil)
			So(out.String(), ShouldEqual, `updated Ingress/nginx-test.example.com (app: app-id-1, domain: test.example.com)
created Ingress/traefik-test.example.com (app: app-id-1, domain: test.example.com)
unchanged Certificate/cert-test.example.com (app: app-id-1, domain: test.example.com)
`)

			nginx, err := kube.DynamicClient.Resource(ingressGVR).Namespace(kube.Namespace).Get(ctx, "nginx-test.example.com", metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(nginx.GetAnnotations(), ShouldResemble, map[string]string{
				"nginx.ingress.kubernetes.io/proxy-body-size": "10m",
			})
			className, _, _ := unstructured.NestedString(nginx.Object, "spec", "ingressClassName")
			So(className, ShouldEqual, "nginx")

			traefik, err := kube.DynamicClient.Resource(ingressGVR).Namespace(kube.Namespace).Get(ctx, "traefik-test.example.com", metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(traefik.GetLabels(), ShouldResemble, map[string]string{
				service.LabelAppID:    "app-id-1",
				service.LabelDomainID: "domain-id-1",
			})

			out.Reset()
			err = kube.ApplyResourcesForDomains(ctx, []service.DomainResourcesTarget{target}, service.ApplyResourcesForDomainsOptions{Output: &out})
			So(err, ShouldBeNil)
			So(out.String(), ShouldEqual, `unchanged Ingress/nginx-test.example.com (app: app-id-1, domain: test.example.com)
unchanged Ingress/traefik-test.example.com (app: app-id-1, domain: test.example.com)
unchanged Certificate/cert-test.example.com (app: app-id-1, domain: test.example.com)
`)
		})

		Convey("prune deletes only the resources of the domain that are not rendered", func() {
			nginx := newObject("networking.k8s.io/v1", "Ingress", "nginx-test.example.com", "domain-id-1", nil, map[string]any{
				"ingressClassName": "nginx",
				"rules":            rules,
			})
			otherDomain := newObject("networking.k8s.io/v1", "Ingress", "nginx-other.example.com", "domain-id-2", nil, map[string]any{
				"ingressClassName": "nginx",
			})
			unlabeled := newObject("networking.k8s.io/v1", "Ingress", "infrastructure-ingress", "", nil, map[string]any{})
			kube, err := newKubernetesWithDynamicClient("testdata/apply_traefik_only.tpl.yaml", nginx, otherDomain, unlabeled, cert)
			So(err, ShouldBeNil)

			var out bytes.Buffer
			err = kube.ApplyResourcesForDomains(ctx, []service.DomainResourcesTarget{target}, service.ApplyResourcesForDomainsOptions{
				Prune:  true,
				Output: &out,
			})
			So(err, ShouldBeNil)
			So(out.String(), ShouldEqual, `created Ingress/traefik-test.example.com (app: app-id-1, domain: test.example.com)
unchanged Certificate/cert-test.example.com (app: app-id-1, domain: test.example.com)
deleted Ingress/nginx-test.example.com (app: app-id-1, domain: test.example.com)
`)

			ingresses, err := kube.DynamicClient.Resource(ingressGVR).Namespace(kube.Namespace).List(ctx, metav1.ListOptions{})
			So(err, ShouldBeNil)
			var names []string
			for _, item := range ingresses.Items {
				names = append(names, item.GetName())
			}
			So(names, ShouldResemble, []string{"infrastructure-ingress", "nginx-other.example.com", "traefik-test.example.com"})

			certs, err := kube.DynamicClient.Resource(certGVR).Namespace(kube.Namespace).List(ctx, metav1.ListOptions{})
			So(err, ShouldBeNil)
			So(len(certs.Items), ShouldEqual, 1)
		})

		Convey("a failed domain does not stop the others", func() {
			kube, err := newKubernetesWithDynamicClient("testdata/invalid_template.tpl.yaml")
			So(err, ShouldBeNil)

			var out bytes.Buffer
			other := target
			other.Domain = "other.example.com"
			err = kube.ApplyResourcesForDomains(ctx, []service.DomainResourcesTarget{target, other}, service.ApplyResourcesForDomainsOptions{Output: &out})
			So(err, ShouldNotBeNil)
			So(strings.Count(out.String(), "error "), ShouldEqual, 2)
		})
	})
}
