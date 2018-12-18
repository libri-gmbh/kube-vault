package processor

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
)

const envPrefix = "SECRET_"

// Env handles variables consumed to and written to env vars / a file containing env vars
type Env struct {
	values     []string
	targetFile string
}

func NewEnv(env []string) *Env {
	return &Env{
		values:     env,
		targetFile: "/env/secrets",
	}
}

// SetTargetSecretsFile sets the file path where the rendered env vars should be stored.
func (p *Env) SetTargetSecretsFile(targetFile string) {
	p.targetFile = targetFile
}

// Process reads a list of environment variables and fetches the referenced secrets from vault,
// storing the results in a file using the bash export syntax.
func (p *Env) Process(logger *logrus.Entry, logicalClient vaultLogicalClient) error {
	var values []string
	var leases []string

	for _, envVar := range p.values {
		if !strings.HasPrefix(envVar, envPrefix) {
			continue
		}

		envVarName, path := p.splitAndCleanEnv(envVar)
		secret, err := logicalClient.Read(path)
		if err != nil {
			return fmt.Errorf("failed to read secret endpoint %q: %v", path, err)
		}

		values = append(values, p.formatExports(logger, envVarName, secret.Data)...)
		leases = append(leases, secret.LeaseID)
	}

	if err := p.writeFile(values, p.targetFile); err != nil {
		return fmt.Errorf("failed to write secrets file: %v", err)
	}

	if err := p.writeFile(leases, LeasesFileName(p.targetFile)); err != nil {
		return fmt.Errorf("failed to write secrets file: %v", err)
	}

	return nil
}

// splitAndCleanEnv receives an env var and splits it into key and value, which gets trimmed
// prefixes and slashes respectively
func (p *Env) splitAndCleanEnv(env string) (string, string) {
	parts := strings.Split(env, "=")
	return strings.Trim(parts[0], envPrefix), strings.Trim(parts[1], "/")
}

// formatExports renders the export statements for the given secret, doing recursive calls if values are nested.
// This method is using reflect to analyze the type of the secrets data in order to handle the values correctly.
// If you think there is a better / more efficient / cleaner way to do so please open a PR and contribute the solution.
func (p *Env) formatExports(logger *logrus.Entry, envVarName string, secret interface{}) []string {
	var values []string

	secretType := reflect.ValueOf(secret)
	switch secretType.Kind() {
	case reflect.Map:
		for _, key := range secretType.MapKeys() {
			nestedKey := p.formatKey(envVarName, key.String())
			values = append(values, p.formatExports(logger, nestedKey, secretType.MapIndex(key).Interface())...)
		}
	case reflect.Slice:
		return []string{p.formatExport(envVarName, strings.Join(secret.([]string), ","))}

	case reflect.String:
		return []string{p.formatExport(envVarName, secret.(string))}

	default:
		logger.Fatalf("Unknown type of secret value in env processor: %v", reflect.TypeOf(secret).Elem().String())
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

func (p *Env) writeFile(values []string, filePath string) error {
	content := strings.Join(values, "\n")
	if err := ioutil.WriteFile(filePath, []byte(content), 0640); err != nil {
		return fmt.Errorf("failed to write the env secrets file to %q: %v", p.targetFile, err)
	}

	return nil
}
