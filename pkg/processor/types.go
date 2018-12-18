package processor

import (
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

// Processor processes env var requirements of an application and renders the result
type Processor interface {
	Process(logger *logrus.Entry, client vaultLogicalClient) error
}

type vaultLogicalClient interface {
	Read(path string) (*api.Secret, error)
}
