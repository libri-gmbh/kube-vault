package processor

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	internalTesting "github.com/libri-gmbh/kube-vault/pkg/internal/testing"
)

func TestEnv_FormatKey(t *testing.T) {
	_, logger := internalTesting.NewLogger()
	env := &Env{
		logger: logger,
	}

	exp := "ASDF_QWERTZ"
	res := env.formatKey("asdf", "qwertz")
	if res != exp {
		t.Errorf("Expected to get %s, got %s", exp, res)
	}
}

func TestEnv_FormatExport(t *testing.T) {
	_, logger := internalTesting.NewLogger()
	env := &Env{
		logger: logger,
	}

	exp := "export ASDF_QWERTZ=test1234"
	res := env.formatExport("ASDF_QWERTZ", "test1234")
	if res != exp {
		t.Errorf("Expected to get %s, got %s", exp, res)
	}
}

func TestEnv_SplitAndCleanEnv(t *testing.T) {
	_, logger := internalTesting.NewLogger()
	env := &Env{
		logger: logger,
	}

	expKey := "ASDF_QWERTZ"
	expVal := "test1234"
	key, val := env.splitAndCleanEnv("SECRET_ASDF_QWERTZ=/test1234/")
	if expKey != key {
		t.Errorf("Expected to get key %s, got %s", expKey, key)
	}
	if expVal != val {
		t.Errorf("Expected to get val %s, got %s", expVal, val)
	}
}

func TestEnv_FormatExportsString(t *testing.T) {
	_, logger := internalTesting.NewLogger()
	env := &Env{
		logger: logger,
	}
	exp := []string{
		"export ASDF_QWERTZ=test1234",
	}

	res := env.formatExports("ASDF_QWERTZ", "test1234")
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Expected to get %s, got %s", exp, res)
	}
}

func TestEnv_FormatExportsStringSlice(t *testing.T) {
	_, logger := internalTesting.NewLogger()
	env := &Env{
		logger: logger,
	}
	exp := []string{
		"export ASDF_QWERTZ=test1234,test5678,teeeest",
	}

	res := env.formatExports("ASDF_QWERTZ", []string{"test1234", "test5678", "teeeest"})
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Expected to get %s, got %s", exp, res)
	}
}

func TestEnv_FormatExportsMapInterface(t *testing.T) {
	_, logger := internalTesting.NewLogger()
	env := &Env{
		logger: logger,
	}
	exp := []string{
		"export ASDF_QWERTZ_PASSWORD=test5678",
		"export ASDF_QWERTZ_USERNAME=test1234",
	}

	res := env.formatExports("ASDF_QWERTZ", map[string]interface{}{"username": "test1234", "password": "test5678"})
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Expected to get %s, got %s", exp, res)
	}
}

func TestEnv_FormatExportsMapString(t *testing.T) {
	_, logger := internalTesting.NewLogger()
	env := &Env{
		logger: logger,
	}
	exp := []string{
		"export ASDF_QWERTZ_PASSWORD=test5678",
		"export ASDF_QWERTZ_USERNAME=test1234",
	}

	res := env.formatExports("ASDF_QWERTZ", map[string]string{"username": "test1234", "password": "test5678"})
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Expected to get %s, got %s", exp, res)
	}
}

func TestEnv_FormatExportsMapInterfaceNested(t *testing.T) {
	_, logger := internalTesting.NewLogger()
	env := &Env{
		logger: logger,
	}
	exp := []string{
		"export ASDF_QWERTZ_ENDPOINT_URL=http://asdf.net/",
		"export ASDF_QWERTZ_PASSWORD=test5678",
		"export ASDF_QWERTZ_USERNAME=test1234",
	}

	res := env.formatExports("ASDF_QWERTZ", map[string]interface{}{
		"username": "test1234",
		"password": "test5678",
		"endpoint": map[string]interface{}{
			"url": "http://asdf.net/",
		},
	})
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Expected to get %s, got %s", exp, res)
	}
}

func TestEnv_Process(t *testing.T) {
	secret := &api.Secret{
		Data: map[string]interface{}{
			"username": "test1234",
			"password": "test5678",
			"endpoint": map[string]interface{}{
				"url": "http://asdf.net/",
			},
		},
	}

	_, logger := internalTesting.NewLogger()
	client := internalTesting.NewVaultClientLogical(secret, nil)

	tmpfile, cleanup, err := internalTesting.CreateTempFile(logger)
	if err != nil {
		t.Fatalf("failed to create tmpfile: %v", err)
	}
	defer cleanup()

	env := &Env{
		values: []string{
			"SECRET_ASDF_QWERTZ=secrets/asdf/qwertz",
		},
		targetFile: tmpfile,
		logger:     logger,
	}
	exp := []string{
		"export ASDF_QWERTZ_ENDPOINT_URL=http://asdf.net/",
		"export ASDF_QWERTZ_PASSWORD=test5678",
		"export ASDF_QWERTZ_USERNAME=test1234",
	}

	err = env.Process(client)
	if err != nil {
		t.Fatalf("Got unexpected error from Process(): %v", err)
	}

	bValues, err := ioutil.ReadFile(tmpfile)
	if err != nil {
		t.Fatalf("failed to read written env file: %v", err)
	}

	content := strings.Split(string(bValues), "\n")
	if !reflect.DeepEqual(exp, content) {
		t.Errorf("Expected to get %s, got %s", exp, content)
	}

	bLeases, err := ioutil.ReadFile(LeasesFileName(tmpfile))
	if err != nil {
		t.Fatalf("failed to read written env file: %v", err)
	}

	var leases []*api.Secret

	if err := json.Unmarshal(bLeases, &leases); err != nil {
		t.Fatalf("failed to unmarshal json lease file content: %v", err)
	}

	if len(leases) != 1 {
		t.Errorf("Invalid amount of leases, expected %d, got %d", 1, len(leases))
	}
}
