package processor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

const envPrefix = "SECRET_"

// Env handles variables consumed to and written to env vars / a file containing env vars
type Env struct {
	logger     *logrus.Entry
	values     []string
	envFile    string
	leasesFile string
}

// NewEnv returns a new Env processor instance
func NewEnv(logger *logrus.Entry, env []string, envFile, leasesFile string) *Env {
	return &Env{
		logger:     logger,
		values:     env,
		envFile:    envFile,
		leasesFile: leasesFile,
	}
}

// Process reads a list of environment variables and fetches the referenced secrets from vault,
// storing the results in a file using the bash export syntax.
func (p *Env) Process(logicalClient vaultLogicalClient) error {
	var values []string
	var secrets []*api.Secret

	for _, envVar := range p.values {
		if !strings.HasPrefix(envVar, envPrefix) {
			p.logger.Debugf("Skipping %q, not prefixed with SECRET_", strings.Split(envVar, "=")[0])
			continue
		}

		envVarName, uri := p.splitAndCleanEnv(envVar)
		p.logger.Debugf("Loading env var %q from %q", envVarName, uri)

		secret, err := logicalClient.Read(uri)
		if err != nil {
			return fmt.Errorf("failed to read secret endpoint %q: %v", uri, err)
		}

		values = append(values, p.formatExports(envVarName, secret.Data)...)
		secrets = append(secrets, secret)
	}

	valuesBytes := []byte(strings.Join(values, "\n"))
	if err := p.writeFile(valuesBytes, p.envFile); err != nil {
		return fmt.Errorf("failed to write secrets file: %v", err)
	}

	if err := p.writeJSONFile(secrets, p.leasesFile); err != nil {
		return fmt.Errorf("failed to write secrets leases file: %v", err)
	}

	return nil
}

// splitAndCleanEnv receives an env var and splits it into key and value, which gets trimmed
// prefixes and slashes respectively
func (p *Env) splitAndCleanEnv(env string) (string, string) {
	parts := strings.Split(env, "=")
	return strings.Replace(parts[0], envPrefix, "", 1), strings.Trim(parts[1], "/")
}

// formatExports renders the export statements for the given secret, doing recursive calls if values are nested.
// This method is using reflect to analyze the type of the secrets data in order to handle the values correctly.
// If you think there is a better / more efficient / cleaner way to do so please open a PR and contribute the solution.
func (p *Env) formatExports(envVarName string, secret interface{}) []string {
	var values []string

	secretType := reflect.ValueOf(secret)
	if secret == nil || (secretType.Kind() == reflect.Ptr && secretType.IsNil()) {
		p.logger.Debugf("Skipping %q as its value is nil", envVarName)
		return []string{}
	}

	switch secretType.Kind() {
	case reflect.Map:
		for _, key := range secretType.MapKeys() {
			nestedKey := p.formatKey(envVarName, key.String())
			values = append(values, p.formatExports(nestedKey, secretType.MapIndex(key).Interface())...)
		}

	case reflect.Slice:
		return []string{p.formatExport(envVarName, strings.Join(secret.([]string), ","))}

	case reflect.String:
		return []string{p.formatExport(envVarName, secret.(string))}

	default:
		p.logger.Fatalf("Unknown type %q of secret value in env processor: %v", secretType.Kind().String(), secret)
	}

	sort.Strings(values)

	return values
}

// formatKey joins the given pieces of a key and converts it to upper case
func (p *Env) formatKey(values ...string) string {
	return strings.ToUpper(strings.Join(values, "_"))
}

// formatExport formats a bash compatible export statement with the given key and value
func (p *Env) formatExport(key, value string) string {
	return fmt.Sprintf("export %v=%v", strings.ToUpper(key), value)
}

func (p *Env) writeJSONFile(content interface{}, filePath string) error {
	b, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to encode file content: %v", err)
	}

	return p.writeFile(b, filePath)
}

func (p *Env) writeFile(content []byte, filePath string) error {
	if err := ioutil.WriteFile(filePath, content, 0700); err != nil {
		return fmt.Errorf("failed to write the env secrets file to %q: %v", p.envFile, err)
	}

	return nil
}
