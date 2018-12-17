package cmd

type config struct {
	KubeAuthRole   string `required:"true" split_words:"true"`
	KubeAuthPath   string `default:"kubernetes" split_words:"true"`
	KubeTokenFile  string `default:"/run/secrets/kubernetes.io/serviceaccount/token" split_words:"true"`
	VaultTokenFile string `default:"/env/vault-token" split_words:"true"`
	Verbose        bool   `default:"false" split_words:"true"`
}
