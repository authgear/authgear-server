package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/spf13/afero"
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
)

var LabelDomainID = "authgear.com/domain-id"

var ErrDuplicatedAppID = apierrors.AlreadyExists.WithReason("DuplicatedAppID").
	New("duplicated app ID")

var ErrGetStaticAppIDsNotSupported = errors.New("only local FS config source can get static app ID")

type ingressDef struct {
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
	AppID               string
	AppConfigYAML       []byte
	SecretConfigYAML    []byte
	TranslationJSONPath string
	TranslationJSON     []byte
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

func (s *ConfigService) UpdateConfig(appID string, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Kubernetes:
		err := s.updateKubernetes(src, appID, updateFiles, deleteFiles)
		if err != nil {
			return err
		}
		s.Controller.ReloadApp(appID)

	case *configsource.LocalFS:
		err := s.updateLocalFS(src, appID, updateFiles, deleteFiles)
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

func (s *ConfigService) updateKubernetes(k *configsource.Kubernetes, appID string, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
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
	for _, file := range updateFiles {
		path := strings.TrimPrefix(file.Path, "/")
		secret.Data[configsource.EscapePath(path)] = []byte(file.Content)
		updated = true
	}
	for _, path := range deleteFiles {
		path := strings.TrimPrefix(path, "/")
		if _, ok := secret.Data[configsource.EscapePath(path)]; ok {
			delete(secret.Data, configsource.EscapePath(path))
			updated = true
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

func (s *ConfigService) updateLocalFS(l *configsource.LocalFS, appID string, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
	fs := l.Fs
	for _, file := range updateFiles {
		err := fs.MkdirAll(filepath.Dir(file.Path), 0777)
		if err != nil {
			return err
		}
		err = afero.WriteFile(fs, file.Path, []byte(file.Content), 0666)
		if err != nil {
			return err
		}
	}
	for _, path := range deleteFiles {
		err := fs.Remove(path)
		// Ignore file not found errors
		if err != nil && !os.IsNotExist(err) {
			return err
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

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-" + opts.AppID,
			Labels: map[string]string{
				configsource.LabelAppID: opts.AppID,
			},
		},
		Data: map[string][]byte{
			configsource.EscapePath(configsource.AuthgearYAML):       opts.AppConfigYAML,
			configsource.EscapePath(configsource.AuthgearSecretYAML): opts.SecretConfigYAML,
		},
	}
	if opts.TranslationJSONPath != "" {
		secret.Data[configsource.EscapePath(opts.TranslationJSONPath)] = opts.TranslationJSON
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

	def := &ingressDef{
		AppID:    appID,
		DomainID: domainID,
		IsCustom: isCustom,
		Host:     domain,
	}
	setupCert, err := s.setupTLSCert(k, def, tlsCertConfig)
	if err != nil {
		return fmt.Errorf("cannot setup TLS certificate: %w", err)
	}

	ingress := &networkingv1beta1.Ingress{}
	if err := s.generateIngress(def, ingress); err != nil {
		return fmt.Errorf("cannot generate ingress resource: %w", err)
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

	ingress.OwnerReferences = append(ingress.OwnerReferences,
		*metav1.NewControllerRef(&appList.Items[0], corev1.SchemeGroupVersion.WithKind("Secret")),
	)

	ingress, err = k.Client.NetworkingV1beta1().Ingresses(k.Namespace).Create(s.Context, ingress, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	err = setupCert(ingress)
	if err != nil {
		return err
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

func (s *ConfigService) generateIngress(def *ingressDef, ingress *networkingv1beta1.Ingress) error {
	tpl, err := ioutil.ReadFile(s.AppConfig.Kubernetes.IngressTemplateFile)
	if err != nil {
		return err
	}

	template, err := texttemplate.New("ingress").Parse(string(tpl))
	if err != nil {
		return err
	}

	ingressYAML := bytes.Buffer{}
	err = template.Execute(&ingressYAML, def)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(ingressYAML.Bytes(), &ingress)
	if err != nil {
		return err
	}

	return nil
}

func (s *ConfigService) setupTLSCert(
	k *configsource.Kubernetes,
	def *ingressDef,
	source portalconfig.TLSCertConfig,
) (func(*networkingv1beta1.Ingress) error, error) {
	switch source.Type {
	case portalconfig.TLSCertNone:
		return func(*networkingv1beta1.Ingress) error { return nil }, nil

	case portalconfig.TLSCertStatic:
		def.TLSSecretName = source.SecretName
		return func(*networkingv1beta1.Ingress) error { return nil }, nil

	case portalconfig.TLSCertCertManager:
		def.TLSSecretName = "tls-" + def.Host
		return func(ingress *networkingv1beta1.Ingress) error {
			cert := &certmanagerv1.Certificate{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: k.Namespace,
					Name:      "cert-" + def.Host,
					Labels: map[string]string{
						configsource.LabelAppID: def.AppID,
					},
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(ingress, networkingv1beta1.SchemeGroupVersion.WithKind("Ingress")),
					},
				},
				Spec: certmanagerv1.CertificateSpec{
					SecretName: def.TLSSecretName,
					IssuerRef: certmanagermetav1.ObjectReference{
						Kind: source.IssuerKind,
						Name: source.IssuerName,
					},
					CommonName: def.Host,
				},
			}

			client, err := certmanagerclientset.NewForConfig(k.KubeConfig)
			if err != nil {
				return err
			}

			_, err = client.CertmanagerV1().Certificates(k.Namespace).
				Create(context.Background(), cert, metav1.CreateOptions{})
			if err != nil {
				return err
			}

			return nil
		}, nil

	default:
		panic("config_service: unknown certificate type")
	}
}
