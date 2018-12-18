package testing

import "github.com/hashicorp/vault/api"

type VaultClientLogical struct {
	Result      *api.Secret
	ResultError error
}

func NewVaultClientLogical(result *api.Secret, err error) *VaultClientLogical {
	return &VaultClientLogical{
		Result:      result,
		ResultError: err,
	}
}

func (c *VaultClientLogical) Delete(path string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

func (c *VaultClientLogical) List(path string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

func (c *VaultClientLogical) Read(path string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

func (c *VaultClientLogical) ReadWithData(path string, data map[string][]string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

func (c *VaultClientLogical) Unwrap(wrappingToken string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

func (c *VaultClientLogical) Write(path string, data map[string]interface{}) (*api.Secret, error) {
	return c.Result, c.ResultError
}
