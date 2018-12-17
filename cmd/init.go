// Copyright © 2018 Alexander Pinnecke <alexander.pinnecke@googlemail.com>
//

package cmd

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"github.com/kelseyhightower/envconfig"
	"github.com/libri-gmbh/kube-vault-sidecar/pkg/siedecar"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run the sidecar as init container to fetch secrets and store credentials",
	Run: func(cmd *cobra.Command, args []string) {

		baseLogger := logrus.New()
		baseLogger.SetFormatter(&logrus.JSONFormatter{})
		logger := baseLogger.WithField("cmd", "init")

		var cfg config
		if err := envconfig.Process("", &cfg); err != nil {
			logger.Fatalf("Failed to parse env config: %v", err)
		}

		if cfg.Verbose {
			baseLogger.SetLevel(logrus.DebugLevel)
		} else {
			baseLogger.SetLevel(logrus.InfoLevel)
		}

		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			logger.Fatalf("Failed to create vault client: %v", err)
		}

		auth := siedecar.NewVaultAuthenticator(logger)
		token, err := auth.Authenticate(client, cfg.KubeAuthPath, cfg.KubeAuthRole, cfg.KubeTokenFile, cfg.VaultTokenFile)
		if err != nil {
			logger.Fatalf("failed to authenticate with vault: %v", err)
		}

		fmt.Printf("%+v\n\n", token)
		fmt.Printf("%+v\n\n", token.Auth)
		fmt.Printf("%+v\n\n", token.WrapInfo)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

}
