package siedecar

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

type LeaseManager struct {
	logger *logrus.Entry
	client *api.Client
}

func NewLeaseManager(logger *logrus.Entry, client *api.Client) *LeaseManager {
	return &LeaseManager{
		logger: logger,
		client: client,
	}
}

func (m *LeaseManager) StartRenew(ctx context.Context, leaseFile string) {
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

	//defer m.revokeAuthToken()
	//defer m.revokeLeases(secrets)

	<-ctx.Done()
}

func (m *LeaseManager) backOff(ctx context.Context, ttl int, handle func()) {
	t := time.NewTimer(time.Second * time.Duration(ttl))

	select {
	case <-ctx.Done():
		return
	case <-t.C:
		handle()
	}
}

func (m *LeaseManager) renewAuthToken(ctx context.Context) {
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

func (m *LeaseManager) revokeAuthToken() {
	err := m.client.Auth().Token().RevokeSelf("")
	if err != nil {
		m.logger.Errorf("failed to revoke self token: %v", err)
		return
	}

	m.logger.Info("Auth token revoked")
}

func (m *LeaseManager) loadLeasesFromFile(leaseFile string) ([]*api.Secret, error) {
	m.logger.Debugf("Loading leases from file %s", leaseFile)
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

func (m *LeaseManager) fetchLeases(leases []string) ([]*api.Secret, error) {
	var secrets []*api.Secret
	for _, lease := range leases {
		m.logger.Debugf("Fetching details for lease %q", lease)
		secret, err := m.client.Logical().Write("sys/leases/lookup", map[string]interface{}{"lease_id": lease})
		if err != nil {
			return []*api.Secret{}, fmt.Errorf("failed to fetch lease %q: %v", lease, err)
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func (m *LeaseManager) renewLeases(ctx context.Context, secrets []*api.Secret) {
	for _, secret := range secrets {
		go m.renewLease(ctx, secret)
	}
}

func (m *LeaseManager) renewLease(ctx context.Context, currentSecret *api.Secret) {
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

func (m *LeaseManager) revokeLeases(secrets []*api.Secret) {
	for _, secret := range secrets {
		go m.revokeLease(secret)
	}
}

func (m *LeaseManager) revokeLease(secret *api.Secret) {
	err := m.client.Sys().Revoke(secret.LeaseID)
	if err != nil {
		m.logger.Errorf("failed to renew lease %q: %v", secret.LeaseID, err)
		return
	}

	m.logger.Infof("Lease %q revoked", secret.LeaseID)
}
