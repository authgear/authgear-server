package configsource

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
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
	"github.com/authgear/authgear-server/pkg/util/fs"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

const (
	LabelHostMapping = "authgear.com/host-mapping"
	LabelConfigAppID = "authgear.com/config-app-id"
)

const HostMapJSON = "hosts.json"

type KubernetesLogger struct{ *log.Logger }

func NewKubernetesLogger(lf *log.Factory) KubernetesLogger {
	return KubernetesLogger{lf.New("configsource-kubernetes")}
}

type Kubernetes struct {
	Logger     KubernetesLogger
	Clock      clock.Clock
	TrustProxy config.TrustProxy
	Config     *Config

	Namespace string                `wire:"-"`
	Client    *kubernetes.Clientset `wire:"-"`
	done      chan<- struct{}       `wire:"-"`
	hostMap   *atomic.Value         `wire:"-"`
	appIDs    *atomic.Value         `wire:"-"`
	appMap    *sync.Map             `wire:"-"`
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

	k.Client, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	k.Namespace = k.Config.KubeNamespace
	if k.Namespace == "" {
		k.Namespace = corev1.NamespaceDefault
	}

	k.hostMap = &atomic.Value{}
	k.hostMap.Store(map[string]string{})
	k.appIDs = &atomic.Value{}
	k.appIDs.Store([]string{})
	k.appMap = &sync.Map{}

	done := make(chan struct{})
	k.done = done

	configMapCtrl := k.newController(corev1.ResourceConfigMaps, &corev1.ConfigMap{}, k.onUpdate, k.onDelete)
	secretCtrl := k.newController(corev1.ResourceSecrets, &corev1.Secret{}, k.onUpdate, k.onDelete)
	go configMapCtrl.Run(done)
	go secretCtrl.Run(done)
	go k.cleanupCache(done)

	return nil
}

func (k *Kubernetes) onUpdate(resource metav1.Object) {
	switch resource := resource.(type) {
	case *corev1.ConfigMap:
		if value, ok := resource.Labels[LabelHostMapping]; ok && value == "true" {
			data, ok := resource.Data[HostMapJSON]
			if !ok {
				k.Logger.
					WithField("namespace", resource.GetNamespace()).
					WithField("name", resource.GetName()).
					Error("host map JSON not found")
				return
			}
			k.updateHostMap([]byte(data))
		} else if appID, ok := resource.Labels[LabelConfigAppID]; ok && appID != "" {
			k.invalidateApp(appID)
		}
	case *corev1.Secret:
		if appID, ok := resource.Labels[LabelConfigAppID]; ok && appID != "" {
			k.invalidateApp(appID)
		}
	default:
		panic(fmt.Sprintf("k8s_config: unexpected resource type: %T", resource))
	}
}

func (k *Kubernetes) onDelete(resource metav1.Object) {
	if appID, ok := resource.GetLabels()[LabelConfigAppID]; ok && appID != "" {
		k.invalidateApp(appID)
	}
}

func (k *Kubernetes) updateHostMap(data []byte) {
	var hostMap map[string]string
	if err := json.Unmarshal(data, &hostMap); err != nil {
		k.Logger.WithError(err).Error("failed to parse host map")
		return
	}

	appIDMap := make(map[string]struct{})
	for _, appID := range hostMap {
		appIDMap[appID] = struct{}{}
	}
	appIDs := make([]string, 0, len(appIDMap))
	for appID := range appIDMap {
		appIDs = append(appIDs, appID)
	}
	sort.Strings(appIDs)

	k.hostMap.Store(hostMap)
	k.appIDs.Store(appIDs)
	k.Logger.Info("host map reloaded")
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

func (k *Kubernetes) AllAppIDs() ([]string, error) {
	appIDs := k.appIDs.Load().([]string)
	return appIDs, nil
}

func (k *Kubernetes) ResolveAppID(r *http.Request) (string, error) {
	host := httputil.GetHost(r, bool(k.TrustProxy))
	hostMap := k.hostMap.Load().(map[string]string)

	appID, ok := hostMap[host]
	if !ok {
		return "", ErrAppNotFound
	}
	return appID, nil
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
		MatchLabels: map[string]string{LabelConfigAppID: appID},
	})
	if err != nil {
		return "", err
	}
	return labelSelector.String(), nil
}

func (k *Kubernetes) newController(
	resource corev1.ResourceName,
	objType runtime.Object,
	onUpdate func(metav1.Object),
	onDelete func(metav1.Object),
) cache.Controller {
	listWatch := cache.NewListWatchFromClient(
		k.Client.CoreV1().RESTClient(),
		string(resource),
		k.Namespace,
		fields.Everything(),
	)
	if !k.Config.Watch {
		listWatch.WatchFunc = func(options metav1.ListOptions) (watch.Interface, error) {
			return emptyWatch(make(chan watch.Event)), nil
		}
	}

	fifo := cache.NewDeltaFIFO(cache.MetaNamespaceKeyFunc, nil)
	ctrl := cache.New(&cache.Config{
		Queue:         fifo,
		ListerWatcher: listWatch,
		ObjectType:    objType,

		Process: func(obj interface{}) error {
			for _, d := range obj.(cache.Deltas) {
				switch d.Type {
				case cache.Sync, cache.Added, cache.Updated:
					onUpdate(d.Object.(metav1.Object))
				case cache.Deleted:
					onDelete(d.Object.(metav1.Object))
				}
			}
			return nil
		},
	})
	return ctrl
}

func MakeAppFS(configMap *corev1.ConfigMap, secret *corev1.Secret) (fs.Fs, error) {
	// Construct a FS that treats `a` and `/a` the same.
	// The template is loaded by a file URI which is always an absoluted path.
	appFs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	create := func(name string, data []byte) {
		file, _ := appFs.Create(name)
		_, _ = file.Write(data)
	}

	for key, data := range configMap.Data {
		path, err := UnescapePath(key)
		if err != nil {
			return nil, err
		}
		create(path, []byte(data))
	}
	for key, data := range configMap.BinaryData {
		path, err := UnescapePath(key)
		if err != nil {
			return nil, err
		}
		create(path, data)
	}
	for path, data := range secret.Data {
		create(path, data)
	}

	return &fs.AferoFs{Fs: appFs}, nil
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

	configMaps, err := k.Client.CoreV1().ConfigMaps(k.Namespace).
		List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	secrets, err := k.Client.CoreV1().Secrets(k.Namespace).
		List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	if len(configMaps.Items) != 1 || len(secrets.Items) != 1 {
		return nil, fmt.Errorf(
			"%w: failed to query config resources (ConfigMaps: %d, Secrets: %d)",
			ErrAppNotFound,
			len(configMaps.Items),
			len(secrets.Items),
		)
	}

	appFs, err := MakeAppFS(&configMaps.Items[0], &secrets.Items[0])
	if err != nil {
		return nil, err
	}
	appConfig, err := loadConfig(appFs)
	if err != nil {
		return nil, err
	}
	return &config.AppContext{
		Fs:     appFs,
		Config: appConfig,
	}, nil
}

type emptyWatch chan watch.Event

func (w emptyWatch) Stop() {
	close(w)
}
func (w emptyWatch) ResultChan() <-chan watch.Event {
	return w
}
