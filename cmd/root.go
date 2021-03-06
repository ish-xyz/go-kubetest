package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/ish-xyz/go-kubetest/pkg/assert"
	"github.com/ish-xyz/go-kubetest/pkg/controller"
	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/metrics"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// Used for flags
	namespace      string // required flag
	kubeconfig     string
	metricsAddress string
	cpuProfile     string
	interval       int
	debug          bool
	once           bool
	selectors      map[string]string

	rootCmd = &cobra.Command{
		Use:   "kubetest",
		Short: "A tool to test your kubernetes cluster",
		Long: `Kubetest run as in-cluster solution and run
			integration tests on your Kubernetes cluster`,
		Run: exec,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "The location where the tests definitions are (namespace or directory)")
	rootCmd.PersistentFlags().StringVarP(&metricsAddress, "metrics-address", "m", "0.0.0.0:9000", "Run the controller in debug mode")
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Kubernetes config file path")
	rootCmd.PersistentFlags().StringVarP(&cpuProfile, "cpu-profile", "p", "", "Path to save the cpu-profile file")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 1200, "The interval between one test execution and the next one")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Run the controller in debug mode")
	rootCmd.PersistentFlags().BoolVarP(&once, "once", "o", false, "Run controller only once")
	rootCmd.PersistentFlags().StringToStringVarP(
		&selectors,
		"select",
		"l",
		map[string]string{},
		"Pass the labels for test definitions. Empty selectors means all test definitions.",
	)
	rootCmd.MarkPersistentFlagRequired("namespace")
}

func handleErr(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

func exec(cmd *cobra.Command, args []string) {

	var restConfig *rest.Config
	var err error
	var ldr loader.Loader

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		handleErr(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// preliminary checks
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if kubeconfig == "" {
		restConfig, err = rest.InClusterConfig()
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	handleErr(err)

	dynclient, err := dynamic.NewForConfig(restConfig)
	handleErr(err)

	client, err := kubernetes.NewForConfig(restConfig)
	handleErr(err)

	metricsAddressList := strings.Split(metricsAddress, ":")
	address := metricsAddressList[0]
	port, err := strconv.Atoi(metricsAddressList[1])
	handleErr(err)

	metricsCtrl := metrics.NewMetricsController(dynclient, address, port)

	// initiate objects
	prv := provisioner.NewProvisioner(restConfig, client, dynclient)
	asrt := assert.NewAssert(prv)
	ldr = loader.NewKubernetesLoader(prv)
	controllerInstance := controller.NewController(ldr, prv, metricsCtrl, asrt)

	// Prepare selectors
	sl := make(map[string]interface{}, len(selectors))
	for k, v := range selectors {
		sl[fmt.Sprintf("metadata.labels.%s", k)] = v
	}

	// Start controller
	controllerInstance.Run(context.TODO(), namespace, sl, time.Duration(interval)*time.Second, once)
}
