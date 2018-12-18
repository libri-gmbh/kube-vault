package processor

import "testing"

func TestLeasesFileName(t *testing.T) {
	secretsFile := "/var/env/test-secrets"
	leasesFile := "/var/env/test-secrets.leases"
	generatedLeasesFile := LeasesFileName(secretsFile)

	if generatedLeasesFile != leasesFile {
		t.Errorf("expected to get %s, got %s", leasesFile, generatedLeasesFile)
	}
}
