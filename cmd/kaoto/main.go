package main

import (
	"flag"
	"os"

	"github.com/kaotoIO/kaoto-operator/cmd/kaoto/run"
	"github.com/kaotoIO/kaoto-operator/pkg/logger"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kaoto",
		Short: "kaoto",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	rootCmd.AddCommand(run.NewRunCmd())

	fs := flag.NewFlagSet("", flag.PanicOnError)

	klog.InitFlags(fs)
	logger.Options.BindFlags(fs)

	rootCmd.PersistentFlags().AddGoFlagSet(fs)

	if err := rootCmd.Execute(); err != nil {
		klog.ErrorS(err, "problem running command")
		os.Exit(1)
	}
}
