package cmd

import (
	"context"
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
	location       string // required flag
	kubeconfig     string
	metricsAddress string
	interval       int
	debug          bool
	filesystem     bool

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
	rootCmd.PersistentFlags().StringVar(&location, "location", "", "The location where the tests definitions are (namespace or directory)")
	rootCmd.PersistentFlags().StringVar(&metricsAddress, "metrics-address", "0.0.0.0:9000", "Run the controller in debug mode")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Kubernetes config file path")
	rootCmd.PersistentFlags().IntVar(&interval, "interval", 1200, "The interval between one test execution and the next one")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Run the controller in debug mode")
	rootCmd.PersistentFlags().BoolVar(&filesystem, "from-filesystem", false, "Wheter to load the tests definitions from filesystem or from the Kubernetes API")
	rootCmd.MarkPersistentFlagRequired("location")
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

	// initiate objects
	prv := provisioner.NewProvisioner(restConfig, client, dynclient)
	asrt := assert.NewAssert(prv)

	ldr = loader.NewKubernetesLoader(prv)
	if filesystem {
		ldr = loader.NewFileSystemLoader()
	}

	ms := metrics.NewServer(address, port)
	controllerInstance := controller.NewController(ldr, prv, ms, asrt)
	controllerInstance.Run(context.TODO(), location, time.Duration(interval)*time.Second)

}
