package testing

import "github.com/hashicorp/vault/api"

// VaultClientLogical implements the vault logical type for testing
type VaultClientLogical struct {
	Result      *api.Secret
	ResultError error
}

// NewVaultClientLogical returns a new VaultClientLogical instance
func NewVaultClientLogical(result *api.Secret, err error) *VaultClientLogical {
	return &VaultClientLogical{
		Result:      result,
		ResultError: err,
	}
}

// Delete just returns the results set on the struct
func (c *VaultClientLogical) Delete(path string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

// List just returns the results set on the struct
func (c *VaultClientLogical) List(path string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

// Read just returns the results set on the struct
func (c *VaultClientLogical) Read(path string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

// ReadWithData just returns the results set on the struct
func (c *VaultClientLogical) ReadWithData(path string, data map[string][]string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

// Unwrap just returns the results set on the struct
func (c *VaultClientLogical) Unwrap(wrappingToken string) (*api.Secret, error) {
	return c.Result, c.ResultError
}

// Write just returns the results set on the struct
func (c *VaultClientLogical) Write(path string, data map[string]interface{}) (*api.Secret, error) {
	return c.Result, c.ResultError
}
