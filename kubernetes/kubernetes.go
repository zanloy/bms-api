package kubernetes // import github.com/zanloy/bms-api/kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	veleroclient "github.com/vmware-tanzu/velero/pkg/client"
	"gopkg.in/olahol/melody.v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	ogkubernetes "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

/* Errors */
type KubernetesError error

var (
	ConfigNotFoundError    KubernetesError = fmt.Errorf("kubernetes config file missing.")
	NamespaceNotFoundError KubernetesError = fmt.Errorf("failed to find namespace")
	TypeCastError          KubernetesError = fmt.Errorf("failed to typecast object")
)

/* Package scoped variables */
var (
	logger        zerolog.Logger
	Clientset     ogkubernetes.Interface
	Config        *rest.Config
	Factory       informers.SharedInformerFactory
	HealthUpdates = melody.New()
	stopCh        <-chan struct{}
)

func FileExists(filename string) bool {
	filename = os.ExpandEnv(filename)
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Init(kubeconfig string) (err error) {
	// Setup logger
	logger = log.With().
		Timestamp().
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

	if FileExists(kubeconfig) {
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
	Factory = informers.NewSharedInformerFactory(Clientset, time.Minute*5)
	setupInformers()
	go Factory.Start(stopCh)

	/* Wait for cache to sync */
	logger.Info().Msg("Waiting for cache to sync...")
	startTime := time.Now()
	// TODO: Add a timeout to this.
	Factory.WaitForCacheSync(stopCh)
	logger.Info().Msg(fmt.Sprintf("Cache sync completed [%.2fs].", time.Since(startTime).Seconds()))

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
	ns, err := Factory.Core().V1().Namespaces().Lister().Get(name)
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

func VeleroBackups() (backups *velerov1.BackupList, err error) {
	backups = &velerov1.BackupList{}

	// Load Velero config (default: ~/.config/velero/config.json)
	veleroConfig, err := veleroclient.LoadConfig()

	// Use the Velero factory to build our velero objs.
	// TODO: Analyze if this could be used for the entire package since it builds for k8 and velero
	f := veleroclient.NewFactory("get", veleroConfig)

	// Create veleroClient
	veleroClient, err := f.Client()
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get backups
	backups, err = veleroClient.VeleroV1().Backups(f.Namespace()).List(ctx, metav1.ListOptions{})

	// Return our results/error
	return
}

func VeleroSchedules() (schedules *velerov1.ScheduleList, err error) {
	schedules = &velerov1.ScheduleList{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load Velero config (default: ~/.config/velero/config.json)
	veleroConfig, err := veleroclient.LoadConfig()

	// Use the Velero factory to build our velero objs.
	// TODO: Analyze if this could be used for the entire package since it builds for k8 and velero
	f := veleroclient.NewFactory("get", veleroConfig)

	// Create veleroClient
	veleroClient, err := f.Client()
	if err != nil {
		return
	}

	// Get schedules
	schedules, err = veleroClient.VeleroV1().Schedules(f.Namespace()).List(ctx, metav1.ListOptions{})

	// Return our results/error
	return
}
