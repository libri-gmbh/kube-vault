package siedecar

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type LeaseRenewer struct {
}

func NewLeaseRenewer() *LeaseRenewer {
	return &LeaseRenewer{}
}

func (l *LeaseRenewer) Renew(client vaultClient, leaseFile string) error {
	content, err := ioutil.ReadFile(leaseFile)
	if err != nil {
		return fmt.Errorf("failed to read written env file: %v", err)
	}

	leases := strings.Split(string(content), "\n")

	if len(leases) == 0 {
		return nil
	}

	return nil
}
