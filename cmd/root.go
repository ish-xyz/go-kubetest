package cmd

import (
	"time"

	"github.com/ish-xyz/go-kubetest/pkg/controller"
	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// Used for flags
	testsdir   string // required flag
	kubeconfig string
	interval   int
	debug      bool
	once       bool
	rootCmd    = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVar(&testsdir, "testsdir", "", "The directory with tests definitions")
	rootCmd.MarkPersistentFlagRequired("testsdir")

	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Kubernetes config file path")
	rootCmd.PersistentFlags().IntVar(&interval, "interval", 1200, "The interval between one test execution and the next one")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Run the controller in debug mode")
	rootCmd.PersistentFlags().BoolVar(&once, "once", false, "Run the tests one time only")
}

func exec(cmd *cobra.Command, args []string) {

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// TODO: if kubeconfig not provided try in-cluster configuration
	_, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logrus.Fatal(err)
	}

	ldr := loader.NewLoader()
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logrus.Fatal(cmd.Help())
	}

	prv := provisioner.NewProvisioner(restConfig)
	ctrl := controller.NewController(prv)

	testSuites, _ := ldr.LoadTestSuites(testsdir)

	ctrl.Run(testSuites, 600*time.Second)
	/*
		TODO:
		- Create provisioner
		- Create controller
		- Create loader
		- Load test files
		- Run Controller with TestsList
	*/
}
