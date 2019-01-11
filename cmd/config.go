package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
)

type config struct {
	KubeAuthRole      string `required:"true" split_words:"true"`
	KubeAuthPath      string `default:"kubernetes" split_words:"true"`
	KubeTokenFile     string `default:"/run/secrets/kubernetes.io/serviceaccount/token" split_words:"true"`
	VaultTokenFile    string `default:"/env/vault-token" split_words:"true"`
	EnvFile           string `default:"/env/secrets" split_words:"true"`
	ProcessorStrategy string `default:"env" split_words:"true"`
	Verbose           bool   `default:"false" split_words:"true"`
}

func newExitHandlerContext(logger *logrus.Entry) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-c
		defer cancel()
		logger.Info("shutting down")
	}()

	return ctx
}
