package siedecar

import "github.com/hashicorp/vault/api"

type vaultClient interface {
}

type authenticator interface {
	GetAuthToken() *api.Secret
}
