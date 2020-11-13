package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	texttemplate "text/template"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/spf13/afero"
	goyaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var LabelDomainID = "authgear.com/domain-id"

var ErrDuplicatedAppID = apierrors.AlreadyExists.WithReason("DuplicatedAppID").
	New("duplicated app ID")

var ErrGetStaticAppIDsNotSupported = errors.New("only local FS config source can get static app ID")

type IngressTemplateData struct {
	AppID         string
	DomainID      string
	IsCustom      bool
	Host          string
	TLSSecretName string
}

type ConfigServiceLogger struct{ *log.Logger }

func NewConfigServiceLogger(lf *log.Factory) ConfigServiceLogger {
	return ConfigServiceLogger{lf.New("config-service")}
}

type CreateAppOptions struct {
	AppID     string
	Resources map[string][]byte
}

type ConfigService struct {
	Context      context.Context
	Logger       ConfigServiceLogger
	AppConfig    *portalconfig.AppConfig
	Controller   *configsource.Controller
	ConfigSource *configsource.ConfigSource
}

func (s *ConfigService) ResolveContext(appID string) (*config.AppContext, error) {
	return s.ConfigSource.ContextResolver.ResolveContext(appID)
}

func (s *ConfigService) GetStaticAppIDs() ([]string, error) {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Kubernetes:
		return nil, ErrGetStaticAppIDsNotSupported
	case *configsource.LocalFS:
		return src.AllAppIDs()
	default:
		return nil, errors.New("unsupported configuration source")
	}
}

func (s *ConfigService) Create(opts *CreateAppOptions) error {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Kubernetes:
		err := s.createKubernetes(src, opts)
		if err != nil {
			return err
		}

	case *configsource.LocalFS:
		return apierrors.NewForbidden("cannot create app for local FS")

	default:
		return errors.New("unsupported configuration source")
	}
	return nil
}

func (s *ConfigService) UpdateResources(appID string, files []*resource.ResourceFile) error {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Kubernetes:
		err := s.updateKubernetes(src, appID, files)
		if err != nil {
			return err
		}
		s.Controller.ReloadApp(appID)

	case *configsource.LocalFS:
		err := s.updateLocalFS(src, appID, files)
		if err != nil {
			return err
		}
		s.Controller.ReloadApp(appID)

	default:
		return errors.New("unsupported configuration source")
	}
	return nil
}

func (s *ConfigService) CreateDomain(appID string, domainID string, domain string, isCustom bool) error {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Kubernetes:
		err := s.createKubernetesIngress(src, appID, domainID, domain, isCustom)
		if err != nil {
			return err
		}

	case *configsource.LocalFS:
		return apierrors.NewForbidden("cannot create domain for local FS")

	default:
		return errors.New("unsupported configuration source")
	}
	return nil
}

func (s *ConfigService) DeleteDomain(domain *model.Domain) error {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Kubernetes:
		err := s.deleteKubernetesIngress(src, domain)
		if err != nil {
			return err
		}

	case *configsource.LocalFS:
		return apierrors.NewForbidden("cannot delete domain for local FS")

	default:
		return errors.New("unsupported configuration source")
	}
	return nil
}

func (s *ConfigService) updateKubernetes(k *configsource.Kubernetes, appID string, updates []*resource.ResourceFile) error {
	labelSelector, err := k.AppSelector(appID)
	if err != nil {
		return err
	}
	secrets, err := k.Client.CoreV1().Secrets(k.Namespace).
		List(s.Context, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		s.Logger.WithError(err).Warn("Failed to load secrets")
		return errors.New("failed to query data store")
	}

	if len(secrets.Items) != 1 {
		err = fmt.Errorf(
			"failed to query config resources (Secrets: %d)",
			len(secrets.Items),
		)
		s.Logger.WithError(err).Warn("Failed to load secrets")
		return errors.New("failed to query data store")
	}
	secret := secrets.Items[0]

	updated := false
	for _, u := range updates {
		key := configsource.EscapePath(u.Location.Path)
		if u.Data == nil {
			if _, ok := secret.Data[key]; ok {
				delete(secret.Data, key)
				updated = true
			}
		} else {
			if !bytes.Equal(secret.Data[key], u.Data) {
				secret.Data[key] = u.Data
				updated = true
			}
		}
	}

	if updated {
		_, err = k.Client.CoreV1().Secrets(k.Namespace).Update(s.Context, &secret, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ConfigService) updateLocalFS(l *configsource.LocalFS, appID string, updates []*resource.ResourceFile) error {
	fs := l.Fs
	for _, file := range updates {
		if file.Data == nil {
			err := fs.Remove(file.Location.Path)
			// Ignore file not found errors
			if err != nil && !os.IsNotExist(err) {
				return err
			}
		} else {
			err := fs.MkdirAll(filepath.Dir(file.Location.Path), 0777)
			if err != nil {
				return err
			}
			err = afero.WriteFile(fs, file.Location.Path, file.Data, 0666)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ConfigService) createKubernetes(k *configsource.Kubernetes, opts *CreateAppOptions) (err error) {
	_, err = k.ResolveContext(opts.AppID)
	if err != nil && !errors.Is(err, configsource.ErrAppNotFound) {
		return err
	} else if err == nil {
		return ErrDuplicatedAppID
	}

	secretData := make(map[string][]byte)
	for path, data := range opts.Resources {
		secretData[configsource.EscapePath(path)] = data
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-" + opts.AppID,
			Labels: map[string]string{
				configsource.LabelAppID: opts.AppID,
			},
		},
		Data: secretData,
	}

	_, err = k.Client.CoreV1().Secrets(k.Namespace).Create(s.Context, secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *ConfigService) createKubernetesIngress(
	k *configsource.Kubernetes,
	appID string,
	domainID string,
	domain string,
	isCustom bool,
) error {
	var tlsCertConfig portalconfig.TLSCertConfig
	if isCustom {
		tlsCertConfig = s.AppConfig.Kubernetes.CustomDomainTLSCert
	} else {
		tlsCertConfig = s.AppConfig.Kubernetes.DefaultDomainTLSCert
	}

	var cert *certmanagerv1.Certificate
	// Prepare template data.
	def := &IngressTemplateData{
		AppID:    appID,
		DomainID: domainID,
		IsCustom: isCustom,
		Host:     domain,
	}
	switch tlsCertConfig.Type {
	case portalconfig.TLSCertNone:
		break
	case portalconfig.TLSCertStatic:
		def.TLSSecretName = tlsCertConfig.SecretName
	case portalconfig.TLSCertCertManager:
		def.TLSSecretName = "tls-" + def.Host
		cert = &certmanagerv1.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: k.Namespace,
				Name:      "cert-" + def.Host,
				Labels: map[string]string{
					configsource.LabelAppID: def.AppID,
				},
			},
			Spec: certmanagerv1.CertificateSpec{
				SecretName: def.TLSSecretName,
				IssuerRef: certmanagermetav1.ObjectReference{
					Kind: tlsCertConfig.IssuerKind,
					Name: tlsCertConfig.IssuerName,
				},
				CommonName: def.Host,
			},
		}
	default:
		panic("config_service: unknown certificate type")
	}

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{configsource.LabelAppID: appID},
	})
	if err != nil {
		return err
	}
	appList, err := k.Client.CoreV1().Secrets(k.Namespace).List(s.Context, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return err
	} else if len(appList.Items) != 1 {
		return fmt.Errorf("cannot get existing app (Secrets: %d)", len(appList.Items))
	}

	ingresses, err := s.generateIngresses(def)
	if err != nil {
		return fmt.Errorf("cannot generate ingress resource: %w", err)
	}

	for _, ingress := range ingresses {
		ingress.OwnerReferences = append(ingress.OwnerReferences,
			*metav1.NewControllerRef(&appList.Items[0], corev1.SchemeGroupVersion.WithKind("Secret")),
		)

		_, err = k.Client.NetworkingV1beta1().Ingresses(k.Namespace).Create(s.Context, ingress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	if cert != nil {
		cert.ObjectMeta.OwnerReferences = append(
			cert.ObjectMeta.OwnerReferences,
			*metav1.NewControllerRef(&appList.Items[0], corev1.SchemeGroupVersion.WithKind("Secret")),
		)
		client, err := certmanagerclientset.NewForConfig(k.KubeConfig)
		if err != nil {
			return err
		}

		_, err = client.CertmanagerV1().Certificates(k.Namespace).
			Create(context.Background(), cert, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ConfigService) deleteKubernetesIngress(k *configsource.Kubernetes, domain *model.Domain) error {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{LabelDomainID: domain.ID},
	})
	if err != nil {
		return err
	}

	client := k.Client.NetworkingV1beta1().Ingresses(k.Namespace)
	ingresses, err := client.List(s.Context, metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return err
	}

	for _, ingress := range ingresses.Items {
		err = client.Delete(s.Context, ingress.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ConfigService) generateIngresses(def *IngressTemplateData) ([]*networkingv1beta1.Ingress, error) {
	b, err := ioutil.ReadFile(s.AppConfig.Kubernetes.IngressTemplateFile)
	if err != nil {
		return nil, err
	}
	return GenerateIngresses(def, b)
}

func GenerateIngresses(def *IngressTemplateData, templateBytes []byte) ([]*networkingv1beta1.Ingress, error) {
	tpl, err := texttemplate.New("ingress").Parse(string(templateBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ingress template: %w", err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, def)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ingress template: %w", err)
	}

	var output []*networkingv1beta1.Ingress
	decoder := goyaml.NewDecoder(bytes.NewReader(buf.Bytes()))
	for {
		var document interface{}
		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to decode yaml: %w", err)
		}

		// Handle empty document.
		if document == nil {
			continue
		}

		documentBytes, err := goyaml.Marshal(document)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal yaml: %w", err)
		}

		var ingress *networkingv1beta1.Ingress
		err = yaml.Unmarshal(documentBytes, &ingress)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal into Ingress: %w", err)
		}

		output = append(output, ingress)
	}
	return output, nil
}
