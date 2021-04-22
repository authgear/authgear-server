package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	texttemplate "text/template"

	goyaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
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
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
)

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

	Context             context.Context                         `wire:"-"`
	Namespace           string                                  `wire:"-"`
	KubeConfig          *rest.Config                            `wire:"-"`
	Client              *kubernetes.Clientset                   `wire:"-"`
	DynamicClient       dynamic.Interface                       `wire:"-"`
	DiscoveryRESTMapper *restmapper.DeferredDiscoveryRESTMapper `wire:"-"`
}

func (k *Kubernetes) open() error {
	var kubeConfig *rest.Config
	var err error
	if k.KubernetesConfig.KubeConfig == "" {
		kubeConfig, err = rest.InClusterConfig()
		if errors.Is(err, rest.ErrNotInCluster) {
			kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
			kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		}
		if err != nil {
			return err
		}
	} else {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", k.KubernetesConfig.KubeConfig)
		if err != nil {
			return err
		}
	}

	k.KubeConfig = kubeConfig
	// setup k8s client for delete ingress / cert when deleting domains
	k.Client, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
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
	if k.Client == nil {
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

	err = k.Client.
		ExtensionsV1beta1().
		Ingresses(k.Namespace).
		DeleteCollection(context.Background(), metav1.DeleteOptions{}, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete extension v1beta1 ingress: %w", err)
	}

	err = k.Client.
		NetworkingV1beta1().
		Ingresses(k.Namespace).
		DeleteCollection(context.Background(), metav1.DeleteOptions{}, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete v1beta1 ingress: %w", err)
	}

	err = k.Client.
		NetworkingV1().
		Ingresses(k.Namespace).
		DeleteCollection(context.Background(), metav1.DeleteOptions{}, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete v1 ingress: %w", err)
	}

	client, err := certmanagerclientset.NewForConfig(k.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to new certmanager client: %w", err)
	}

	err = client.
		CertmanagerV1().
		Certificates(k.Namespace).
		DeleteCollection(context.Background(), metav1.DeleteOptions{}, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete cert manager cert: %w", err)
	}

	return nil
}

func (k *Kubernetes) generateResources(def *ResourceTemplateData) ([]*KubernetesResource, error) {
	b, err := ioutil.ReadFile(k.AppConfig.Kubernetes.IngressTemplateFile)
	if err != nil {
		return nil, err
	}
	return GenerateResources(def, b)
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
