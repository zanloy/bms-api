package kubernetes // import github.com/zanloy/bms-api/kubernetes

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	veleroclient "github.com/vmware-tanzu/velero/pkg/client"
	veleroclientset "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	veleroinformers "github.com/vmware-tanzu/velero/pkg/generated/informers/externalversions"
	"github.com/zanloy/bms-api/helpers"
	"gopkg.in/olahol/melody.v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	ogkubernetes "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

/* Errors */
type KubernetesError error

var (
	ErrConfigNotFound    KubernetesError = fmt.Errorf("kubernetes config file missing")
	ErrNamespaceNotFound KubernetesError = fmt.Errorf("failed to find namespace")
	ErrTypeCast          KubernetesError = fmt.Errorf("failed to typecast object")
)

/* Package scoped variables */
var (
	logger          zerolog.Logger
	Clientset       ogkubernetes.Interface
	Config          *rest.Config
	Factory         informers.SharedInformerFactory
	VeleroClientset veleroclientset.Interface
	VeleroConfig    veleroclient.VeleroConfig
	VeleroFactory   veleroinformers.SharedInformerFactory

	HealthUpdates = melody.New()
	stopCh        <-chan struct{}
)

func Init(kubeconfig string) (err error) {
	// Setup logger
	logger = log.With().
		Str("component", "kubernetes").
		Logger()

	logger.Debug().Msg("Kubernetes initializing...")

	// If no config was passed in, try the default location first.
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			kubeconfig = os.ExpandEnv("~/.kube/config")
		}
	}

	if helpers.FileExists(kubeconfig) {
		logger.Debug().Msg(fmt.Sprintf("Found kubeconfig at %s, attempting to load it.", kubeconfig))
		Config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return err
		}
	} else {
		logger.Debug().Msg(fmt.Sprintf("Failed to find kubeconfig at %s, attempting to use in-cluster configuration...", kubeconfig))
		Config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	}

	Clientset, err = ogkubernetes.NewForConfig(Config)
	if err != nil {
		return err
	}

	// Load Velero config (default: ~/.config/velero/config.json)
	VeleroConfig, _ = veleroclient.LoadConfig()
	VeleroClientset, _ = veleroclient.NewFactory("bms", VeleroConfig).Client()

	logger.Debug().Msg("Kubernetes initilization complete.")
	return nil
}

func InitWithClientset(clientset ogkubernetes.Interface) {
	Clientset = clientset
}

func Start(stopChannel <-chan struct{}) {
	logger.Info().Msg("Kubernetes controller startup initialized.")
	stopCh = stopChannel

	/* Setup cache and informers */
	Factory = informers.NewSharedInformerFactory(Clientset, 0)

	/* Setup velero informer */
	VeleroFactory = veleroinformers.NewSharedInformerFactory(VeleroClientset, 0)

	setupInformers()
	Factory.Start(stopCh)
	VeleroFactory.Start(stopCh)

	// TODO: Add a timeout to this.
	logger.Info().Msg("Waiting for informer cache to sync...")
	startTime := time.Now()
	Factory.WaitForCacheSync(stopCh)
	logger.Info().Msg(fmt.Sprintf("Informer cache sync completed. [%.2fs]", time.Since(startTime).Seconds()))

	logger.Info().Msg("Kubernetes controller startup complete.")
}

// This function will fatally fail if the kubernetes package hasn't been
// initialized yet.
func mustBeInitialized() {
	if Clientset == nil {
		logger.Fatal().Msg("Attempted to read from Kubernetes before initialized.")
	}
}

func NamespaceExists(name string) bool {
	ns, err := Namespaces().Get(name)
	fmt.Println(name)
	fmt.Printf("ns = %+v\n", ns)
	fmt.Printf("err = %+v\n", err)
	return err != nil && ns != nil
}

func NamespacesArray() (namespaces []string, err error) {
	cached, err := Namespaces().List(labels.Everything())
	fmt.Printf("cached = %+v\n", cached)
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		return
	}

	namespaces = make([]string, len(cached))
	for idx, ns := range cached {
		namespaces[idx] = ns.Name
	}
	return
}

// Check is cache is synced
func WaitForCacheSync() {
	cache.WaitForCacheSync(stopCh, Factory.Core().V1().Namespaces().Informer().HasSynced)
}
