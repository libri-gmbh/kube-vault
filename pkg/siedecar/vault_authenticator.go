package siedecar

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v2"
)

type kubeLogin struct {
	JWT  string `json:"jwt"`
	Role string `json:"role"`
}

type VaultAuthenticator struct {
	logger *logrus.Entry
}

var (
	errVaultTokenFileNotFound = errors.New("vault authentication token not found")
	errTokenIsNil             = errors.New("given token is nil or empty")
)

func NewVaultAuthenticator(logger *logrus.Entry) *VaultAuthenticator {
	return &VaultAuthenticator{
		logger: logger,
	}
}

// Authenticate hands over the k8s SA token to vault, receiving the vault authentication token
func (f *VaultAuthenticator) Authenticate(client *api.Client, kubeLoginPath, kubeLoginRole, kubeTokenFilePath, vaultTokenFilePath string) (*api.Secret, error) {
	// first try to read the vault token - if this is successful we are already logged in
	token, err := f.readTokenFile(client, vaultTokenFilePath)
	if err != nil && err != errVaultTokenFileNotFound {
		return nil, err
	} else if err == nil {
		return token, nil
	}

	k8sTokenBytes, err := ioutil.ReadFile(kubeTokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %s", err)
	}

	k8sToken := strings.TrimSpace(string(k8sTokenBytes))
	f.logger.Debugf("loaded JWT token %v from kube token path %q", k8sToken, kubeTokenFilePath)

	req := client.NewRequest(http.MethodPost, fmt.Sprintf("/v1/auth/%s/login", kubeLoginPath))
	req.SetJSONBody(&kubeLogin{JWT: k8sToken, Role: kubeLoginRole})
	resp, err := client.RawRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Error() != nil {
		return nil, resp.Error()
	}

	token = &api.Secret{}
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", err)
	}

	f.logger.Infof("successfully authenticated kube role %s at %s", kubeLoginRole, kubeLoginPath)
	client.SetToken(token.Auth.ClientToken)

	if err := f.writeTokenToFile(token, vaultTokenFilePath); err != nil {
		return nil, fmt.Errorf("failed to save token to file: %v", err)
	}

	return token, nil
}

// readTokenFile reads a vault token from a given path
func (f *VaultAuthenticator) readTokenFile(client *api.Client, vaultTokenFilePath string) (*api.Secret, error) {
	f.logger.Debugf("trying to read token from file %q", vaultTokenFilePath)
	if _, err := os.Stat(vaultTokenFilePath); os.IsNotExist(err) {
		return nil, errVaultTokenFileNotFound
	}

	var token api.Secret
	bytes, err := ioutil.ReadFile(vaultTokenFilePath)
	if err != nil {
		f.logger.Fatal("failed to read token:", err)
	}

	err = yaml.Unmarshal(bytes, &token)
	if err != nil {
		f.logger.Fatal("failed to parse token")
	}

	f.logger.Debugf("Setting secret auth token %q on vault client", token.Auth.Accessor)
	client.SetToken(token.Auth.ClientToken)

	return &token, nil
}

func (f *VaultAuthenticator) writeTokenToFile(token *api.Secret, vaultTokenFilePath string) error {
	if token == nil {
		return errTokenIsNil
	}

	f.logger.Debugf("writing secret auth token content to file %q", vaultTokenFilePath)

	b, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %v", err)
	}

	err = ioutil.WriteFile(vaultTokenFilePath, b, 0640)
	if err != nil {
		return fmt.Errorf("failed to write token to file: %v", err)
	}

	return nil
}
