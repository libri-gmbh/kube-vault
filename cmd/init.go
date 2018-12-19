// Copyright © 2018 Alexander Pinnecke <alexander.pinnecke@googlemail.com>

package cmd

import (
	"os"

	"github.com/libri-gmbh/kube-vault/pkg/processor"
	"github.com/libri-gmbh/kube-vault/pkg/vault"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run the sidecar as init container to fetch secrets and store credentials",
	Run: func(cmd *cobra.Command, args []string) {
		logger := baseLogger.WithField("cmd", "init")
		auth := vault.NewAuthenticator(logger)
		_, err := auth.Authenticate(client, true, cfg.KubeAuthPath, cfg.KubeAuthRole, cfg.KubeTokenFile, cfg.VaultTokenFile)
		if err != nil {
			baseLogger.Fatalf("failed to authenticate with vault: %v", err)
		}

		switch cfg.ProcessorStrategy {
		case "env":
			env := processor.NewEnv(logger, os.Environ(), cfg.EnvFile)
			err = env.Process(client.Logical())
			if err != nil {
				logger.Fatal(err)
			}

		default:
			logger.Fatalf("Undefined strategy %q. Possible values: [env]", cfg.ProcessorStrategy)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

}
