package api

import "github.com/hashicorp/vault/api"

type CredentialFetcher interface {
	Fetch(path string) (*api.Secret, error)
}

type LeaseRenewer interface {
	RenewLease()
}
