package configsource

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
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

const HostMapJSON = "hosts.json"

type KubernetesLogger struct{ *log.Logger }

func NewKubernetesLogger(lf *log.Factory) KubernetesLogger {
	return KubernetesLogger{lf.New("kubernetes-config")}
}

type Kubernetes struct {
	Logger     LocalFSLogger
	Clock      clock.Clock
	TrustProxy config.TrustProxy
	Config     *Config

	namespace string                `wire:"-"`
	client    *kubernetes.Clientset `wire:"-"`
	done      chan<- struct{}       `wire:"-"`
	hostMap   *atomic.Value         `wire:"-"`
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

	k.client, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	k.namespace = k.Config.KubeNamespace
	if k.namespace == "" {
		k.namespace = corev1.NamespaceDefault
	}

	k.hostMap = &atomic.Value{}
	k.hostMap.Store(map[string]string{})
	k.appMap = &sync.Map{}

	done := make(chan struct{})
	k.done = done

	configMapCtrl := k.newController(corev1.ResourceConfigMaps, &corev1.ConfigMap{}, k.onUpdate, k.onDelete)
	secretCtrl := k.newController(corev1.ResourceSecrets, &corev1.Secret{}, k.onUpdate, k.onDelete)
	go configMapCtrl.Run(done)
	go secretCtrl.Run(done)

	return nil
}

func (k *Kubernetes) onUpdate(resource metav1.Object) {
	switch resource := resource.(type) {
	case *corev1.ConfigMap:
		if resource.Name == k.Config.KubeAppHostMapName {
			data, ok := resource.Data[HostMapJSON]
			if !ok {
				k.Logger.
					WithField("namespace", resource.GetNamespace()).
					WithField("name", resource.GetName()).
					Error("host map JSON not found")
				return
			}
			k.updateHostMap([]byte(data))
		}
	case *corev1.Secret:
	default:
		panic(fmt.Sprintf("k8s_config: unexpected resource type: %T", resource))
	}
}

func (k *Kubernetes) onDelete(resource metav1.Object) {
	fmt.Printf("delete %v\n", resource)
}

func (k *Kubernetes) updateHostMap(data []byte) {
	var hostMap map[string]string
	if err := json.Unmarshal(data, &hostMap); err != nil {
		k.Logger.WithError(err).Error("failed to parse host map")
		return
	}
	k.hostMap.Store(hostMap)
	k.Logger.Info("host map reloaded")
}

func (k *Kubernetes) Close() error {
	close(k.done)
	return nil
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

func (k *Kubernetes) newController(
	resource corev1.ResourceName,
	objType runtime.Object,
	onUpdate func(metav1.Object),
	onDelete func(metav1.Object),
) cache.Controller {
	listWatch := cache.NewListWatchFromClient(
		k.client.CoreV1().RESTClient(),
		string(resource),
		k.namespace,
		fields.Everything(),
	)
	if !k.Config.Watch {
		listWatch.WatchFunc = func(options metav1.ListOptions) (watch.Interface, error) {
			return emptyWatch(make(chan watch.Event)), nil
		}
	}

	fifo := cache.NewDeltaFIFO(cache.MetaNamespaceKeyFunc, nil)
	ctrl := cache.New(&cache.Config{
		Queue:            fifo,
		ListerWatcher:    listWatch,
		ObjectType:       objType,
		FullResyncPeriod: time.Hour,
		RetryOnError:     false,

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

func makeAppFs(configMap *corev1.ConfigMap, secret *corev1.Secret) fs.Fs {
	appFs := afero.NewMemMapFs()
	create := func(name string, data []byte) {
		file, _ := appFs.Create(name)
		_, _ = file.Write(data)
	}

	for path, data := range configMap.Data {
		create(path, []byte(data))
	}
	for path, data := range configMap.BinaryData {
		create(path, data)
	}
	for path, data := range secret.Data {
		create(path, data)
	}

	return &fs.AferoFs{Fs: appFs}
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
	atomic.StoreInt64(&a.lastUsedAt, k.Clock.NowMonotonic().UnixNano())
	return a.appCtx, a.err
}

func (a *k8sApp) doLoad(k *Kubernetes) (*config.AppContext, error) {
	resourceName := k.Config.KubeAppConfigPrefix + a.appID

	configMap, err := k.client.CoreV1().ConfigMaps(k.namespace).
		Get(resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secret, err := k.client.CoreV1().Secrets(k.namespace).
		Get(resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	appFs := makeAppFs(configMap, secret)
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
