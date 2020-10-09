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

var ErrDuplicatedAppID = apierrors.AlreadyExists.WithReason("DuplicatedAppID").
	New("duplicated app ID")

type ingressDef struct {
	AppID         string
	Host          string
	TLSSecretName string
}

type ConfigServiceLogger struct{ *log.Logger }

func NewConfigServiceLogger(lf *log.Factory) ConfigServiceLogger {
	return ConfigServiceLogger{lf.New("config-service")}
}

type CreateAppOptions struct {
	AppID               string
	Hosts               []string
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

func (s *ConfigService) ListAllAppIDs() ([]string, error) {
	return s.ConfigSource.AppIDResolver.AllAppIDs()
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

	// Setup config resource
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.AppConfig.Kubernetes.NewResourcePrefix + opts.AppID,
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

	tlsCertConfig := s.AppConfig.Kubernetes.DefaultDomainTLSCert
	var ingresses []*networkingv1beta1.Ingress
	for _, host := range opts.Hosts {
		def := &ingressDef{
			AppID: opts.AppID,
			Host:  host,
		}
		if err = s.setupTLSCert(k, def, tlsCertConfig); err != nil {
			return fmt.Errorf("cannot setup TLS certificate: %w", err)
		}

		ingress := &networkingv1beta1.Ingress{}
		if err = s.generateIngress(def, ingress); err != nil {
			return fmt.Errorf("cannot generate ingress resource: %w", err)
		}
		ingresses = append(ingresses, ingress)
	}

	// Commit changes to Kubernetes
	_, err = k.Client.CoreV1().Secrets(k.Namespace).Create(s.Context, secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	for _, ingress := range ingresses {
		_, err = k.Client.NetworkingV1beta1().Ingresses(k.Namespace).Create(s.Context, ingress, metav1.CreateOptions{})
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

func (s *ConfigService) setupTLSCert(k *configsource.Kubernetes, def *ingressDef, source portalconfig.TLSCertConfig) error {
	switch source.Type {
	case portalconfig.TLSCertNone:
		return nil

	case portalconfig.TLSCertStatic:
		def.TLSSecretName = source.SecretName
		return nil

	case portalconfig.TLSCertCertManager:
		def.TLSSecretName = "tls-" + def.Host
		cert := &certmanagerv1.Certificate{
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

	default:
		panic("config_service: unknown certificate type")
	}
}
