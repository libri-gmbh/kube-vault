package testing

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

// CreateTempFile creates a temporary file and returns the path and a callback to delete the file (should be called
// with defer by the caller)
func CreateTempFile(logger *logrus.Entry) (string, func(), error) {
	tmpfile, err := ioutil.TempFile("", "kube_vault_sidecar_test")
	if err != nil {
		log.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		return "", nil, fmt.Errorf("failed to close the file handle: %v", err)
	}

	clean := func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			logger.Fatalf("failed to clean up temp file: %v", err)
		}
	}

	return tmpfile.Name(), clean, nil
}
