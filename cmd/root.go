// Copyright © 2018 Alexander Pinnecke <alexander.pinnecke@googlemail.com>

package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var (
	baseLogger *logrus.Logger
	client     *api.Client
	cfg        = &config{}
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kube-vault",
	Short: "A slim sidecar / init container to fetch and renew vault secret leases.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error

		baseLogger = logrus.New()
		baseLogger.SetFormatter(&logrus.JSONFormatter{})

		err = envconfig.Process("", cfg)
		if err != nil {
			baseLogger.Fatalf("Failed to parse env config: %v", err)
		}

		if cfg.Verbose {
			baseLogger.SetLevel(logrus.DebugLevel)
		} else {
			baseLogger.SetLevel(logrus.InfoLevel)
		}

		client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			baseLogger.Fatalf("Failed to create vault client: %v", err)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
