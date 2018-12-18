package processor

import "fmt"

// LeasesFileName generates the name to store the lease IDs in, based on the secrets file name.
func LeasesFileName(secretsFile string) string {
	return fmt.Sprintf("%v.leases", secretsFile)
}
