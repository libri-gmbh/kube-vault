package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

type kubeAuth struct {
	JWT  string `json:"jwt"`
	Role string `json:"role"`
}

type vaultClient interface {
	NewRequest(method, requestPath string) *api.Request
	RawRequest(request *api.Request) (*api.Response, error)
	SetToken(v string)
}

// Authenticator handles vault kubernetes authentication
type Authenticator struct {
	logger *logrus.Entry
	client vaultClient
	token  *api.Secret
}

var (
	errVaultTokenFileNotFound = errors.New("vault authentication token not found")
	errTokenIsNil             = errors.New("given token is nil or empty")
)

// NewAuthenticator returns a new Authenticator instance
func NewAuthenticator(logger *logrus.Entry, client vaultClient) *Authenticator {
	return &Authenticator{
		logger: logger,
		client: client,
	}
}

// Authenticate hands over the k8s SA token to vault, receiving the vault authentication token
func (f *Authenticator) Authenticate(forceLogin bool, kubeAuthPath, kubeAuthRole, kubeTokenFilePath, vaultTokenFilePath string) (*api.Secret, error) {
	if !forceLogin {
		// first try to read the vault token - if this is successful we are already logged in
		token, err := f.readTokenFile(vaultTokenFilePath)
		if err != nil && err != errVaultTokenFileNotFound {
			return nil, err
		} else if err == nil {
			return token, nil
		}
	}

	token, err := f.fetchKubeAuthToken(kubeTokenFilePath, kubeAuthPath, kubeAuthRole)
	if err != nil {
		return nil, err
	}

	f.logger.Infof("successfully authenticated kube role %s at %s", kubeAuthRole, kubeAuthPath)

	if err := f.writeTokenToFile(token, vaultTokenFilePath); err != nil {
		return nil, fmt.Errorf("failed to save token to file: %v", err)
	}

	f.token = token
	f.client.SetToken(f.token.Auth.ClientToken)

	return token, nil
}

func (f *Authenticator) fetchKubeAuthToken(kubeTokenFilePath, kubeAuthPath, kubeAuthRole string) (*api.Secret, error) {
	// nolint: gosec
	k8sTokenBytes, err := ioutil.ReadFile(kubeTokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %s", err)
	}

	k8sToken := strings.TrimSpace(string(k8sTokenBytes))
	f.logger.Debugf("loaded JWT token %v from kube token path %q", k8sToken, kubeTokenFilePath)

	req := f.client.NewRequest(http.MethodPost, fmt.Sprintf("/v1/auth/%s/login", kubeAuthPath))
	err = req.SetJSONBody(&kubeAuth{JWT: k8sToken, Role: kubeAuthRole})
	if err != nil {
		return nil, fmt.Errorf("failed to set json body on auth request: %v", err)
	}

	resp, err := f.client.RawRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Error() != nil {
		return nil, resp.Error()
	}

	token := &api.Secret{}
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", err)
	}

	return token, nil
}

// readTokenFile reads a vault token from a given path
func (f *Authenticator) readTokenFile(vaultTokenFilePath string) (*api.Secret, error) {
	f.logger.Debugf("trying to read token from file %v", vaultTokenFilePath)
	if _, err := os.Stat(vaultTokenFilePath); os.IsNotExist(err) {
		return nil, errVaultTokenFileNotFound
	}

	// nolint: gosec
	bytes, err := ioutil.ReadFile(vaultTokenFilePath)
	if err != nil {
		f.logger.Fatalf("failed to read token: %v", err)
	}

	token := &api.Secret{}
	err = json.Unmarshal(bytes, token)
	if err != nil {
		f.logger.Fatal("failed to parse token")
	}

	f.token = token
	f.logger.Debugf("Setting secret auth token %q on vault client", f.token.Auth.Accessor)
	f.client.SetToken(f.token.Auth.ClientToken)

	return f.token, nil
}

// writeTokenToFile writes an authentication token to given file
func (f *Authenticator) writeTokenToFile(token *api.Secret, vaultTokenFilePath string) error {
	if token == nil {
		return errTokenIsNil
	}

	f.logger.Debugf("writing secret auth token content to file %s", vaultTokenFilePath)

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
