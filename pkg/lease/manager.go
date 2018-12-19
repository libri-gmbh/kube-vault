package lease

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

// Manager handles leases and cares about automatic renewal of them
type Manager struct {
	logger *logrus.Entry
	client *api.Client
}

// NewManager returns a new Manager instance
func NewManager(logger *logrus.Entry, client *api.Client) *Manager {
	return &Manager{
		logger: logger,
		client: client,
	}
}

// StartRenew kicks of the renew processes - one for the auth token and one per leased secret
func (m *Manager) StartRenew(ctx context.Context, leaseFile string) {
	leases, err := m.loadLeasesFromFile(leaseFile)
	if err != nil {
		m.logger.Fatal(err)
	}

	if len(leases) == 0 {
		m.logger.Infof("No leases found in file %q", leaseFile)
		return
	}

	go m.renewAuthToken(ctx)
	go m.renewLeases(ctx, leases)

	defer m.revokeAuthToken()
	defer m.revokeLeases(leases)

	<-ctx.Done()
}

func (m *Manager) backOff(ctx context.Context, ttl int, handle func()) {
	t := time.NewTimer(time.Second * time.Duration(ttl))

	select {
	case <-ctx.Done():
		return
	case <-t.C:
		handle()
	}
}

func (m *Manager) renewAuthToken(ctx context.Context) {
	secret, err := m.client.Auth().Token().RenewSelf(1800)
	if err != nil {
		m.logger.Errorf("failed to renew auth token from accessor: %v", err)
		return
	}

	m.logger.Infof("Auth token renewed, backing off for %d seconds", secret.Auth.LeaseDuration/2)

	m.backOff(ctx, secret.Auth.LeaseDuration/2, func() {
		m.renewAuthToken(ctx)
	})
}

func (m *Manager) revokeAuthToken() {
	err := m.client.Auth().Token().RevokeSelf("")
	if err != nil {
		m.logger.Errorf("failed to revoke self token: %v", err)
		return
	}

	m.logger.Info("Auth token revoked")
}

func (m *Manager) loadLeasesFromFile(leaseFile string) ([]*api.Secret, error) {
	m.logger.Debugf("Loading leases from file %s", leaseFile)

	// nolint: gosec
	content, err := ioutil.ReadFile(leaseFile)
	if err != nil {
		return []*api.Secret{}, fmt.Errorf("failed to read written env file: %v", err)
	}

	var leases []*api.Secret
	if err := json.Unmarshal(content, &leases); err != nil {
		return []*api.Secret{}, fmt.Errorf("failed to unmarshal json leases file: %v", err)
	}

	m.logger.Debugf("Found %d leases in file %s", len(leases), leaseFile)

	return leases, nil
}

func (m *Manager) renewLeases(ctx context.Context, secrets []*api.Secret) {
	for _, secret := range secrets {
		go m.renewLease(ctx, secret)
	}
}

func (m *Manager) renewLease(ctx context.Context, currentSecret *api.Secret) {
	secret, err := m.client.Sys().Renew(currentSecret.LeaseID, currentSecret.LeaseDuration)
	if err != nil {
		m.logger.Errorf("failed to renew lease %q: %v", currentSecret.LeaseID, err)
		return
	}

	m.logger.Infof("Lease %q renewed, backing off for %d seconds", secret.LeaseID, secret.LeaseDuration/2)

	m.backOff(ctx, secret.LeaseDuration/2, func() {
		m.renewLease(ctx, secret)
	})
}

func (m *Manager) revokeLeases(secrets []*api.Secret) {
	for _, secret := range secrets {
		go m.revokeLease(secret)
	}
}

func (m *Manager) revokeLease(secret *api.Secret) {
	err := m.client.Sys().Revoke(secret.LeaseID)
	if err != nil {
		m.logger.Errorf("failed to renew lease %q: %v", secret.LeaseID, err)
		return
	}

	m.logger.Infof("Lease %q revoked", secret.LeaseID)
}
