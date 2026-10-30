package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"log/slog"
	texttemplate "text/template"

	goyaml "go.yaml.in/yaml/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	certmanagerclientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/kubeutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var LabelAppID = "authgear.com/app-id"
var LabelDomainID = "authgear.com/domain-id"

var KubernetesLogger = slogutil.NewLogger("kubernetes")

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
	ctx context.Context,
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

	templateBytes, err := os.ReadFile(k.AppConfig.Kubernetes.IngressTemplateFile)
	if err != nil {
		return fmt.Errorf("cannot generate domain related resources: %w", err)
	}

	resources, err := k.renderResourcesForDomain(def, templateBytes)
	if err != nil {
		return err
	}

	if k.DynamicClient == nil {
		if err := k.open(); err != nil {
			return fmt.Errorf("failed to init k8s dynamic client: %w", err)
		}
	}

	for _, r := range resources {
		dr, err := k.prepareResourceForDomain(r, def)
		if err != nil {
			return err
		}

		_, err = dr.Create(ctx, r.Object, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create resources: %v %w", r.Object, err)
		}
	}

	return nil
}

type DomainResourcesTarget struct {
	AppID    string
	DomainID string
	Domain   string
	IsCustom bool
}

type ApplyResourcesForDomainsOptions struct {
	// DryRun sends every write as a server-side dry run.
	DryRun bool
	// Prune deletes the resources labeled with the ID of a domain that the template does not render anymore.
	Prune  bool
	Output io.Writer
}

// ApplyResourcesForDomains renders the template for each domain, creates the missing resources,
// and replaces the labels, annotations and spec of the existing ones.
// A failed domain does not stop the others; all errors are returned.
func (k *Kubernetes) ApplyResourcesForDomains(ctx context.Context, domains []DomainResourcesTarget, opts ApplyResourcesForDomainsOptions) error {
	if k.DynamicClient == nil {
		if err := k.open(); err != nil {
			return fmt.Errorf("failed to init k8s dynamic client: %w", err)
		}
	}

	templateBytes, err := os.ReadFile(k.AppConfig.Kubernetes.IngressTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to read domain related k8s resources template: %w", err)
	}

	var existingByDomainID map[string][]existingDomainResource
	if opts.Prune {
		existingByDomainID, err = k.listResourcesByDomainID(ctx)
		if err != nil {
			return err
		}
	}

	suffix := ""
	if opts.DryRun {
		suffix = " (dry run)"
	}

	var errs []error
	for _, d := range domains {
		report := func(action string, kind string, name string) {
			fmt.Fprintf(opts.Output, "%s %s/%s (app: %s, domain: %s)%s\n", action, kind, name, d.AppID, d.Domain, suffix)
		}
		err := k.applyResourcesForDomain(ctx, d, templateBytes, existingByDomainID[d.DomainID], opts, report)
		if err != nil {
			err = fmt.Errorf("app %s, domain %s: %w", d.AppID, d.Domain, err)
			fmt.Fprintf(opts.Output, "error %v\n", err)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

type existingDomainResource struct {
	Object   unstructured.Unstructured
	Resource dynamic.ResourceInterface
}

func (k *Kubernetes) applyResourcesForDomain(
	ctx context.Context,
	d DomainResourcesTarget,
	templateBytes []byte,
	existing []existingDomainResource,
	opts ApplyResourcesForDomainsOptions,
	report func(action string, kind string, name string),
) error {
	def := &ResourceTemplateData{
		AppID:    d.AppID,
		DomainID: d.DomainID,
		IsCustom: d.IsCustom,
		Host:     d.Domain,
	}

	resources, err := k.renderResourcesForDomain(def, templateBytes)
	if err != nil {
		return err
	}

	var dryRun []string
	if opts.DryRun {
		dryRun = []string{metav1.DryRunAll}
	}

	rendered := make(map[schema.GroupKind]map[string]struct{})
	for _, r := range resources {
		dr, err := k.prepareResourceForDomain(r, def)
		if err != nil {
			return err
		}

		gk := r.GVK.GroupKind()
		if rendered[gk] == nil {
			rendered[gk] = make(map[string]struct{})
		}
		rendered[gk][r.Object.GetName()] = struct{}{}

		action, err := applyResource(ctx, dr, r.Object, dryRun)
		if err != nil {
			return fmt.Errorf("failed to apply %s/%s: %w", r.GVK.Kind, r.Object.GetName(), err)
		}
		report(action, r.GVK.Kind, r.Object.GetName())
	}

	if opts.Prune {
		for _, e := range existing {
			gk := e.Object.GroupVersionKind().GroupKind()
			if _, ok := rendered[gk][e.Object.GetName()]; ok {
				continue
			}
			err := e.Resource.Delete(ctx, e.Object.GetName(), metav1.DeleteOptions{DryRun: dryRun})
			if err != nil {
				return fmt.Errorf("failed to delete %s/%s: %w", gk.Kind, e.Object.GetName(), err)
			}
			report("deleted", gk.Kind, e.Object.GetName())
		}
	}

	return nil
}

// applyResource creates desired if it does not exist, otherwise replaces the labels, annotations and spec
// of the existing resource. It returns "created", "updated" or "unchanged".
func applyResource(ctx context.Context, dr dynamic.ResourceInterface, desired *unstructured.Unstructured, dryRun []string) (string, error) {
	existing, err := dr.Get(ctx, desired.GetName(), metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		_, err = dr.Create(ctx, desired, metav1.CreateOptions{DryRun: dryRun})
		if err != nil {
			return "", err
		}
		return "created", nil
	} else if err != nil {
		return "", err
	}

	updated := existing.DeepCopy()
	updated.SetLabels(desired.GetLabels())
	updated.SetAnnotations(desired.GetAnnotations())
	if spec, ok := desired.Object["spec"]; ok {
		updated.Object["spec"] = spec
	} else {
		delete(updated.Object, "spec")
	}

	// Compare with what the server would store, so that fields defaulted by the server do not count as changes.
	result, err := dr.Update(ctx, updated, metav1.UpdateOptions{DryRun: []string{metav1.DryRunAll}})
	if err != nil {
		return "", err
	}
	if sameLabelsAnnotationsAndSpec(existing, result) {
		return "unchanged", nil
	}

	if len(dryRun) == 0 {
		_, err = dr.Update(ctx, updated, metav1.UpdateOptions{})
		if err != nil {
			return "", err
		}
	}
	return "updated", nil
}

func sameLabelsAnnotationsAndSpec(a *unstructured.Unstructured, b *unstructured.Unstructured) bool {
	sameMap := func(x map[string]string, y map[string]string) bool {
		return (len(x) == 0 && len(y) == 0) || equality.Semantic.DeepEqual(x, y)
	}
	return sameMap(a.GetLabels(), b.GetLabels()) &&
		sameMap(a.GetAnnotations(), b.GetAnnotations()) &&
		equality.Semantic.DeepEqual(a.Object["spec"], b.Object["spec"])
}

// listResourcesByDomainID lists the resources of the supported kinds that have LabelDomainID.
func (k *Kubernetes) listResourcesByDomainID(ctx context.Context) (map[string][]existingDomainResource, error) {
	selector := labels.NewSelector()
	req, err := labels.NewRequirement(LabelDomainID, selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*req)

	result := make(map[string][]existingDomainResource)
	seen := make(map[schema.GroupKind]struct{})
	for _, gvk := range supportedGVKsForDomain {
		gk := gvk.GroupKind()
		if _, ok := seen[gk]; ok {
			continue
		}
		seen[gk] = struct{}{}

		mapping, err := k.DiscoveryRESTMapper.RESTMapping(gk)
		if meta.IsNoMatchError(err) {
			// The cluster does not serve this kind, e.g. extensions/v1beta1 Ingress.
			continue
		} else if err != nil {
			return nil, fmt.Errorf("failed to find the gvr: %w", err)
		}

		dr := k.DynamicClient.Resource(mapping.Resource).Namespace(k.Namespace)
		list, err := dr.List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, fmt.Errorf("failed to list %s: %w", gk.Kind, err)
		}
		for _, item := range list.Items {
			domainID := item.GetLabels()[LabelDomainID]
			result[domainID] = append(result[domainID], existingDomainResource{Object: item, Resource: dr})
		}
	}

	return result, nil
}

// renderResourcesForDomain renders the template, and checks that every resource is of a supported kind.
func (k *Kubernetes) renderResourcesForDomain(def *ResourceTemplateData, templateBytes []byte) ([]*KubernetesResource, error) {
	resources, err := GenerateResources(def, templateBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot generate domain related resources: %w", err)
	}

	for _, r := range resources {
		if !k.validateGVKForDomain(r.GVK) {
			return nil, fmt.Errorf("k8s gvk type is not supported: %v", r.GVK)
		}
	}

	return resources, nil
}

// prepareResourceForDomain sets the namespace and the labels of the resource, and returns its client.
func (k *Kubernetes) prepareResourceForDomain(r *KubernetesResource, def *ResourceTemplateData) (dynamic.ResourceInterface, error) {
	mapping, err := k.DiscoveryRESTMapper.RESTMapping(r.GVK.GroupKind(), r.GVK.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to find the gvr: %w", err)
	}

	if mapping.Scope.Name() != meta.RESTScopeNameNamespace {
		return nil, fmt.Errorf("create cluster-wide resources not supported")
	}
	dr := k.DynamicClient.Resource(mapping.Resource).Namespace(k.Namespace)
	r.Object.SetNamespace(k.Namespace)

	objLabels := r.Object.GetLabels()
	if objLabels == nil {
		objLabels = make(map[string]string)
	}
	objLabels[LabelAppID] = def.AppID
	objLabels[LabelDomainID] = def.DomainID
	r.Object.SetLabels(objLabels)

	return dr, nil
}

func (k *Kubernetes) DeleteResourcesForDomain(ctx context.Context, domainID string) error {
	logger := KubernetesLogger.GetLogger(ctx)
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

	count, err := deleteExtensionsV1beta1Ingresses(ctx, k.Client, k.Namespace, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete extension v1beta1 ingress: %w", err)
	}

	logger.Info(ctx, "deleted k8s extension v1beta1 ingresses", slog.Int("count", count))

	count, err = deleteNetworkingV1beta1Ingresses(ctx, k.Client, k.Namespace, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete v1beta1 ingress: %w", err)
	}

	logger.Info(ctx, "deleted k8s networking v1beta1 ingresses", slog.Int("count", count))

	count, err = deleteNetworkingV1Ingresses(ctx, k.Client, k.Namespace, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete v1 ingress: %w", err)
	}

	logger.Info(ctx, "deleted k8s networking v1 ingresses", slog.Int("count", count))

	count, err = deleteCertmanagerV1Certificate(ctx, k.CertManagerClient, k.Namespace, listOptions)
	if err != nil {
		return fmt.Errorf("failed to delete cert manager cert: %w", err)
	}

	logger.Info(ctx, "deleted k8s certs", slog.Int("count", count))

	return nil
}

// updating supportedGVKsForDomain also need to update DeleteResourcesForDomain
// to delete corresponding resources
var supportedGVKsForDomain = []schema.GroupVersionKind{
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

// validateGVKForDomain checked the supported gvk
func (k *Kubernetes) validateGVKForDomain(gvk *schema.GroupVersionKind) bool {
	for _, supported := range supportedGVKsForDomain {
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
		var document any
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
