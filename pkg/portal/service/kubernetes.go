package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	texttemplate "text/template"

	goyaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	certmanagerclientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/kubeutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var LabelAppID = "authgear.com/app-id"
var LabelDomainID = "authgear.com/domain-id"

type KubernetesLogger struct{ *log.Logger }

func NewKubernetesLogger(lf *log.Factory) KubernetesLogger {
	return KubernetesLogger{lf.New("kubernetes")}
}

type ResourceTemplateData struct {
	AppID    string
	DomainID string
	IsCustom bool
	Host     string
}

type KubernetesResource struct {
	Object *unstructured.Unstructured
	GVK    *schema.GroupVersionKind
}

type Kubernetes struct {
	KubernetesConfig *portalconfig.KubernetesConfig
	AppConfig        *portalconfig.AppConfig
	Logger           KubernetesLogger

	Context             context.Context                `wire:"-"`
	Namespace           string                         `wire:"-"`
	KubeConfig          *rest.Config                   `wire:"-"`
	Client              kubernetes.Interface           `wire:"-"`
	CertManagerClient   certmanagerclientset.Interface `wire:"-"`
	DynamicClient       dynamic.Interface              `wire:"-"`
	DiscoveryRESTMapper meta.RESTMapper                `wire:"-"`
}

func (k *Kubernetes) open() error {
	kubeConfig, err := kubeutil.MakeKubeConfig(k.KubernetesConfig.KubeConfig)
	if err != nil {
		return err
	}

	k.KubeConfig = kubeConfig
	// setup k8s client for deleting ingress when deleting domains
	k.Client, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	// setup cert manager k8s client for deleting cert when deleting domains
	k.CertManagerClient, err = certmanagerclientset.NewForConfig(k.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to new certmanager client: %w", err)
	}

	// setup dynamic clients to create domain k8s resources based on template
	dc, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
	if err != nil {
		return err
	}
	k.DiscoveryRESTMapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	dyn, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}
	k.DynamicClient = dyn

	k.Namespace = k.KubernetesConfig.AppNamespace
	if k.Namespace == "" {
		k.Namespace = corev1.NamespaceDefault
	}

	return nil
}

func (k *Kubernetes) CreateResourcesForDomain(
	appID string,
	domainID string,
	domain string,
	isCustom bool,
) error {
	def := &ResourceTemplateData{
		AppID:    appID,
		DomainID: domainID,
		IsCustom: isCustom,
		Host:     domain,
	}

	resources, err := k.generateResources(def)
	if err != nil {
		return fmt.Errorf("cannot generate domain related resources: %w", err)
	}

	for _, r := range resources {
		if !k.validateGVKForDomain(r.GVK) {
			return fmt.Errorf("k8s gvk type is not supported: %v", r.GVK)
		}
	}

	if k.DynamicClient == nil {
		if err := k.open(); err != nil {
			return fmt.Errorf("failed to init k8s dynamic client: %w", err)
		}
	}

	for _, r := range resources {
		mapping, err := k.DiscoveryRESTMapper.RESTMapping(r.GVK.GroupKind(), r.GVK.Version)
		if err != nil {
			return fmt.Errorf("failed to find the gvr: %w", err)
		}

		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			dr = k.DynamicClient.Resource(mapping.Resource).Namespace(k.Namespace)
			r.Object.SetNamespace(k.Namespace)
		} else {
			return fmt.Errorf("create cluster-wide resources not supported")
		}

		labels := r.Object.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[LabelAppID] = appID
		labels[LabelDomainID] = domainID
		r.Object.SetLabels(labels)

		_, err = dr.Create(context.Background(), r.Object, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create resources: %v %w", r.Object, err)
		}
	}

	return nil
}

func (k *Kubernetes) DeleteResourcesForDomain(domainID string) error {
	if k.Client == nil || k.CertManagerClient == nil {
		if err := k.open(); err != nil {
			return fmt.Errorf("failed to init k8s client: %w", err)
		}
	}

	labelSelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{LabelDomainID: domainID},
	})
	if err != nil {
		return err
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	}

	ctx := context.Background()
	count, err := deleteExtensionsV1beta1Ingresses(ctx, k.Client, k.Namespace, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete extension v1beta1 ingress: %w", err)
	}

	k.Logger.WithField("count", count).Info("deleted k8s extension v1beta1 ingresses")

	count, err = deleteNetworkingV1beta1Ingresses(ctx, k.Client, k.Namespace, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete v1beta1 ingress: %w", err)
	}

	k.Logger.WithField("count", count).Info("deleted k8s networking v1beta1 ingresses")

	count, err = deleteNetworkingV1Ingresses(ctx, k.Client, k.Namespace, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete v1 ingress: %w", err)
	}

	k.Logger.WithField("count", count).Info("deleted k8s networking v1 ingresses")

	count, err = deleteCertmanagerV1Certificate(ctx, k.CertManagerClient, k.Namespace, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete cert manager cert: %w", err)
	}

	k.Logger.WithField("count", count).Info("deleted k8s certs")

	return nil
}

func (k *Kubernetes) generateResources(def *ResourceTemplateData) ([]*KubernetesResource, error) {
	b, err := ioutil.ReadFile(k.AppConfig.Kubernetes.IngressTemplateFile)
	if err != nil {
		return nil, err
	}
	return GenerateResources(def, b)
}

// validateGVKForDomain checked the supported gvk
func (k *Kubernetes) validateGVKForDomain(gvk *schema.GroupVersionKind) bool {

	// updating supportedGVKs list also need to update DeleteResourcesForDomain
	// to delete corresponding resources
	var supportedGVKs = []schema.GroupVersionKind{
		{
			Group:   "extensions",
			Version: "v1beta1",
			Kind:    "Ingress",
		},
		{
			Group:   "networking.k8s.io",
			Version: "v1",
			Kind:    "Ingress",
		},
		{
			Group:   "networking.k8s.io",
			Version: "v1beta1",
			Kind:    "Ingress",
		},
		{
			Group:   "cert-manager.io",
			Version: "v1",
			Kind:    "Certificate",
		},
	}

	for _, supported := range supportedGVKs {
		if supported.Group == gvk.Group &&
			supported.Version == gvk.Version &&
			supported.Kind == gvk.Kind {
			return true
		}
	}

	return false

}

func deleteExtensionsV1beta1Ingresses(
	ctx context.Context,
	k8sClient kubernetes.Interface,
	namespace string,
	listOptions metav1.ListOptions,
) (int, error) {
	client := k8sClient.ExtensionsV1beta1().Ingresses(namespace)
	ingresses, err := client.List(ctx, listOptions)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// the server could not find the requested resource
			// means the k8s cluster doesn't support the requested resource
			return 0, nil
		}
		return 0, err
	}

	count := len(ingresses.Items)
	for _, ingress := range ingresses.Items {
		err = client.Delete(ctx, ingress.Name, metav1.DeleteOptions{})
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func deleteNetworkingV1beta1Ingresses(
	ctx context.Context,
	k8sClient kubernetes.Interface,
	namespace string,
	listOptions metav1.ListOptions,
) (int, error) {
	client := k8sClient.NetworkingV1beta1().Ingresses(namespace)
	ingresses, err := client.List(ctx, listOptions)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// the server could not find the requested resource
			// means the k8s cluster doesn't support the requested resource
			return 0, nil
		}
		return 0, err
	}

	count := len(ingresses.Items)
	for _, ingress := range ingresses.Items {
		err = client.Delete(ctx, ingress.Name, metav1.DeleteOptions{})
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func deleteNetworkingV1Ingresses(
	ctx context.Context,
	k8sClient kubernetes.Interface,
	namespace string,
	listOptions metav1.ListOptions,
) (int, error) {
	client := k8sClient.NetworkingV1().Ingresses(namespace)
	ingresses, err := client.List(ctx, listOptions)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// the server could not find the requested resource
			// means the k8s cluster doesn't support the requested resource
			return 0, nil
		}
		return 0, err
	}

	count := len(ingresses.Items)
	for _, ingress := range ingresses.Items {
		err = client.Delete(ctx, ingress.Name, metav1.DeleteOptions{})
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func deleteCertmanagerV1Certificate(
	ctx context.Context,
	k8sClient certmanagerclientset.Interface,
	namespace string,
	listOptions metav1.ListOptions,
) (int, error) {
	client := k8sClient.CertmanagerV1().Certificates(namespace)
	certs, err := client.List(ctx, listOptions)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// the server could not find the requested resource
			// means the k8s cluster doesn't support the requested resource
			return 0, nil
		}
		return 0, err
	}

	count := len(certs.Items)
	for _, ingress := range certs.Items {
		err = client.Delete(ctx, ingress.Name, metav1.DeleteOptions{})
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func GenerateResources(def *ResourceTemplateData, templateBytes []byte) ([]*KubernetesResource, error) {
	tpl, err := texttemplate.New("domain-rescourses").Parse(string(templateBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse domain related k8s resources template: %w", err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, def)
	if err != nil {
		return nil, fmt.Errorf("failed to execute domain related k8s resources template: %w", err)
	}

	var output []*KubernetesResource
	decoder := goyaml.NewDecoder(bytes.NewReader(buf.Bytes()))
	for {
		var document interface{}
		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to decode k8s resources template yaml: %w", err)
		}

		// Handle empty document.
		if document == nil {
			continue
		}

		documentBytes, err := goyaml.Marshal(document)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal k8s resources template yaml: %w", err)
		}

		obj := &unstructured.Unstructured{}
		dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, gvk, err := dec.Decode(documentBytes, nil, obj)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal into k8s resources: %w", err)
		}

		output = append(output, &KubernetesResource{
			Object: obj,
			GVK:    gvk,
		})
	}
	return output, nil
}
