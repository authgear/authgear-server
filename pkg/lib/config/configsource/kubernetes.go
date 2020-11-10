package configsource

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const (
	LabelAppID = "authgear.com/app-id"
)

type ingressSnapshot struct {
	Hosts []string
}

type KubernetesLogger struct{ *log.Logger }

func NewKubernetesLogger(lf *log.Factory) KubernetesLogger {
	return KubernetesLogger{lf.New("configsource-kubernetes")}
}

type Kubernetes struct {
	Logger        KubernetesLogger
	BaseResources *resource.Manager
	Clock         clock.Clock
	TrustProxy    config.TrustProxy
	Config        *Config

	Context    context.Context       `wire:"-"`
	Namespace  string                `wire:"-"`
	KubeConfig *rest.Config          `wire:"-"`
	Client     *kubernetes.Clientset `wire:"-"`
	done       chan<- struct{}       `wire:"-"`
	hostMap    *sync.Map             `wire:"-"`
	ingressMap *sync.Map             `wire:"-"`
	appMap     *sync.Map             `wire:"-"`
}

func (k *Kubernetes) Open() error {
	var kubeConfig *rest.Config
	var err error
	if k.Config.KubeConfig == "" {
		kubeConfig, err = rest.InClusterConfig()
		if errors.Is(err, rest.ErrNotInCluster) {
			kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
			kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		}
		if err != nil {
			return err
		}
	} else {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", k.Config.KubeConfig)
		if err != nil {
			return err
		}
	}

	k.KubeConfig = kubeConfig
	k.Client, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	k.Context = context.Background()
	k.Namespace = k.Config.KubeNamespace
	if k.Namespace == "" {
		k.Namespace = corev1.NamespaceDefault
	}

	k.hostMap = &sync.Map{}
	k.ingressMap = &sync.Map{}
	k.appMap = &sync.Map{}

	done := make(chan struct{})
	k.done = done

	ingressCtrl := k.newController(k.Client.NetworkingV1beta1().RESTClient(), "ingresses", &networkingv1beta1.Ingress{})
	secretCtrl := k.newController(k.Client.CoreV1().RESTClient(), "secrets", &corev1.Secret{})
	go ingressCtrl.Run(done)
	go secretCtrl.Run(done)
	go k.cleanupCache(done)

	return nil
}

func (k *Kubernetes) onUpdate(resource metav1.Object) {
	switch resource := resource.(type) {
	case *networkingv1beta1.Ingress:
		if appID, ok := resource.Labels[LabelAppID]; ok && appID != "" {
			k.updateHostMap(appID, resource)
		}
	case *corev1.Secret:
		if appID, ok := resource.Labels[LabelAppID]; ok && appID != "" {
			k.invalidateApp(appID)
		}
	default:
		panic(fmt.Sprintf("k8s_config: unexpected resource type: %T", resource))
	}
}

func (k *Kubernetes) onDelete(resource metav1.Object) {
	switch resource := resource.(type) {
	case *networkingv1beta1.Ingress:
		if appID, ok := resource.Labels[LabelAppID]; ok && appID != "" {
			k.invalidateHostMap(resource)
		}
	case *corev1.Secret:
		if appID, ok := resource.GetLabels()[LabelAppID]; ok && appID != "" {
			k.invalidateApp(appID)
		}
	default:
		panic(fmt.Sprintf("k8s_config: unexpected resource type: %T", resource))
	}
}

func (k *Kubernetes) invalidateHostMap(ingress *networkingv1beta1.Ingress) {
	snapshot, ok := k.ingressMap.Load(ingress.UID)
	if ok {
		for _, host := range snapshot.(*ingressSnapshot).Hosts {
			k.Logger.WithField("host", host).Info("host invalidated")
			k.hostMap.Delete(host)
		}
		k.ingressMap.Delete(ingress.UID)
	}

	for _, host := range extractIngressHosts(ingress) {
		k.Logger.WithField("host", host).Info("host invalidated")
		k.hostMap.Delete(host)
	}
}

func (k *Kubernetes) updateHostMap(appID string, ingress *networkingv1beta1.Ingress) {
	// Invalidate the hosts of the old ingress.
	snapshot, ok := k.ingressMap.Load(ingress.UID)
	if ok {
		for _, host := range snapshot.(*ingressSnapshot).Hosts {
			k.Logger.WithField("host", host).Info("host invalidated")
			k.hostMap.Delete(host)
		}
	}

	hosts := extractIngressHosts(ingress)

	k.ingressMap.Store(ingress.UID, &ingressSnapshot{
		Hosts: hosts,
	})

	for _, host := range hosts {
		k.hostMap.Store(host, appID)
		k.Logger.WithField("host", host).WithField("app_id", appID).Info("host accepted")
	}
}

func (k *Kubernetes) invalidateApp(appID string) {
	k.appMap.Delete(appID)
	k.Logger.WithField("app_id", appID).Info("invalidated cached config")
}

func (k *Kubernetes) cleanupCache(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return

		case <-time.After(time.Minute):
			now := k.Clock.NowMonotonic().Unix()
			numDel := 0
			k.appMap.Range(func(key, value interface{}) bool {
				app := value.(*k8sApp)
				if atomic.LoadInt64(&app.lastUsedAt) < now-60 {
					k.appMap.Delete(key)
					numDel++
				}
				return true
			})
			if numDel > 0 {
				k.Logger.WithField("deleted", numDel).Info("cleaned cached app configs")
			}
		}
	}
}

func (k *Kubernetes) Close() error {
	close(k.done)
	return nil
}

func (k *Kubernetes) ResolveAppID(r *http.Request) (string, error) {
	host := httputil.GetHost(r, bool(k.TrustProxy))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	appID, ok := k.hostMap.Load(host)
	if !ok {
		return "", ErrAppNotFound
	}
	return appID.(string), nil
}

func (k *Kubernetes) ResolveContext(appID string) (*config.AppContext, error) {
	value, _ := k.appMap.LoadOrStore(appID, &k8sApp{
		appID: appID,
		load:  &sync.Once{},
	})
	app := value.(*k8sApp)
	return app.Load(k)
}

func (k *Kubernetes) ReloadApp(appID string) {
	k.invalidateApp(appID)
}

func (k *Kubernetes) AppSelector(appID string) (string, error) {
	labelSelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{LabelAppID: appID},
	})
	if err != nil {
		return "", err
	}
	return labelSelector.String(), nil
}

func (k *Kubernetes) newController(
	client rest.Interface,
	resource string,
	objType runtime.Object,
) cache.Controller {
	listWatch := cache.NewListWatchFromClient(
		client,
		resource,
		k.Namespace,
		fields.Everything(),
	)
	if !k.Config.Watch {
		listWatch.WatchFunc = func(options metav1.ListOptions) (watch.Interface, error) {
			return emptyWatch(make(chan watch.Event)), nil
		}
	}

	// We use Informer because FIFODelta does not fire Deleted event.
	_, ctrl := cache.NewInformer(listWatch, objType, time.Hour, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			k.onUpdate(obj.(metav1.Object))
		},
		UpdateFunc: func(old, obj interface{}) {
			k.onUpdate(obj.(metav1.Object))
		},
		DeleteFunc: func(obj interface{}) {
			k.onDelete(obj.(metav1.Object))
		},
	})
	return ctrl
}

func MakeAppFS(secret *corev1.Secret) (resource.Fs, error) {
	// Construct a FS that treats `a` and `/a` the same.
	// The template is loaded by a file URI which is always an absoluted path.
	appFs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	create := func(name string, data []byte) {
		file, _ := appFs.Create(name)
		_, _ = file.Write(data)
	}

	for key, data := range secret.Data {
		path, err := UnescapePath(key)
		if err != nil {
			return nil, err
		}
		create(path, data)
	}

	return &resource.AferoFs{Fs: appFs, IsAppFs: true}, nil
}

type k8sApp struct {
	appID      string
	load       *sync.Once
	appCtx     *config.AppContext
	err        error
	lastUsedAt int64
}

func (a *k8sApp) Load(k *Kubernetes) (*config.AppContext, error) {
	if a.load != nil {
		a.load.Do(func() {
			a.appCtx, a.err = a.doLoad(k)
		})
	}
	atomic.StoreInt64(&a.lastUsedAt, k.Clock.NowMonotonic().Unix())
	return a.appCtx, a.err
}

func (a *k8sApp) doLoad(k *Kubernetes) (*config.AppContext, error) {
	labelSelector, err := k.AppSelector(a.appID)
	if err != nil {
		return nil, err
	}

	secrets, err := k.Client.CoreV1().Secrets(k.Namespace).
		List(k.Context, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	if len(secrets.Items) != 1 {
		return nil, fmt.Errorf(
			"%w: failed to query config resources (Secrets: %d)",
			ErrAppNotFound,
			len(secrets.Items),
		)
	}

	appFs, err := MakeAppFS(&secrets.Items[0])
	if err != nil {
		return nil, err
	}
	resources := k.BaseResources.Overlay(appFs)

	appConfig, err := LoadConfig(resources)
	if err != nil {
		return nil, err
	}
	return &config.AppContext{
		AppFs:     appFs,
		Resources: resources,
		Config:    appConfig,
	}, nil
}

type emptyWatch chan watch.Event

func (w emptyWatch) Stop() {
	close(w)
}
func (w emptyWatch) ResultChan() <-chan watch.Event {
	return w
}

func extractIngressHosts(ingress *networkingv1beta1.Ingress) []string {
	var hosts []string
	for _, rule := range ingress.Spec.Rules {
		hosts = append(hosts, rule.Host)
	}
	return hosts
}
