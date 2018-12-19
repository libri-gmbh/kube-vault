// Copyright © 2018 Alexander Pinnecke <alexander.pinnecke@googlemail.com>

package cmd

import (
	"github.com/libri-gmbh/kube-vault/pkg/processor"
	"github.com/libri-gmbh/kube-vault/pkg/siedecar"
	"github.com/libri-gmbh/kube-vault/pkg/vault"
	"github.com/spf13/cobra"
)

// renewCmd represents the renew command
var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew the leases created by the init process",
	Run: func(cmd *cobra.Command, args []string) {
		logger := baseLogger.WithField("cmd", "renew")
		auth := vault.NewAuthenticator(logger)
		_, err := auth.Authenticate(client, false, cfg.KubeAuthPath, cfg.KubeAuthRole, cfg.KubeTokenFile, cfg.VaultTokenFile)
		if err != nil {
			baseLogger.Fatalf("failed to authenticate with vault: %v", err)
		}

		ctx := newExitHandlerContext(logger)
		leaseManager := siedecar.NewLeaseManager(logger, client)
		leaseManager.StartRenew(ctx, processor.LeasesFileName(cfg.EnvFile))
	},
}

func init() {
	RootCmd.AddCommand(renewCmd)
}
