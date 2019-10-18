package processor

import (
	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

// Processor processes env var requirements of an application and renders the result
type Processor interface {
	Process(logger *logrus.Entry, client vaultLogicalClient) error
}

type vaultLogicalClient interface {
	Read(path string) (*api.Secret, error)
}
