package vault

import "github.com/hashicorp/vault/api"

// CredentialStore fetches credentials
type CredentialStore struct {
	client *api.Client
}

func New
