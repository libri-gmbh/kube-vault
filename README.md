# kube-vault

A slim sidecar / init container to fetch and renew vault secret leases written in golang. this image get's build on [Circle.CI](https://circleci.com/gh/libri-gmbh/workflows/kube-vault) and pushed as docker image to Docker Hub as [libri/kube-vault](http://hub.docker.com/r/libri/kube-vault/).

## Background & Inspiration

This project is used to couple vault to k8s, using this one as an init container and as a sidecar, fetching vault secrets (init) and renewing the auth token as well as the leases (renew).

This project is highly inspired by the following projects:

* [WealthWizardsEngineering/kube-vault-auth-init](https://github.com/WealthWizardsEngineering/kube-vault-auth-init) & [WealthWizardsEngineering/kube-vault-auth-renewer](https://github.com/WealthWizardsEngineering/kube-vault-auth-renewer), written in bash and are excessively lacking tests  
* [uswitch/vault-creds](https://github.com/uswitch/vault-creds): Only possible to handle a single secret, as well as (at time writing this) having not a single test

## Usage

This project requires a properly configured [Vault kubernetes auth method](https://www.vaultproject.io/docs/auth/kubernetes.html) and one or multiple configured secret endpoints to fetch the secrets from, as well as a k8s service account token to authenticate at vault.

An example deployment may look like follows:

```yaml
apiVersion: v1
kind: List
items:
- apiVersion: v1
  automountServiceAccountToken: true
  kind: ServiceAccount
  metadata:
    name: dev-example
    namespace: dev
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: dev-example-kube-vault
    namespace: dev
  data:
    KUBE_AUTH_PATH: "dev/example/k8s"
    KUBE_AUTH_ROLE: "dev-example-write"
    VAULT_ADDR: "http://dev-vault:8200"
    VERBOSE: "true"
    SECRET_AWS: "dev/example/aws/creds/write"
    SECRET_MYSQL: "dev/example/mysql/creds/write"
- apiVersion: extensions/v1beta1
  kind: Deployment
  metadata:
    labels:
      run: dev-example
    name: dev-example
    namespace: dev
  spec:
    progressDeadlineSeconds: 600
    replicas: 1
    revisionHistoryLimit: 2
    template:
      metadata:
        labels:
          run: dev-example
      spec:
        volumes:
        - name: env
          emptyDir: {}
        initContainers:
        - name: vault-init
          imagePullPolicy: Always
          image: libri/kube-vault
          args: ["init"]
          envFrom:
            - configMapRef:
                name: dev-example-kube-vault
          volumeMounts:
          - name: env
            mountPath: /env
        containers:
        - name: vault-renew
          image: libri/kube-vault
          envFrom:
            - configMapRef:
                name: dev-example-kube-vault
          args: ["renew"]
          volumeMounts:
          - name: env
            mountPath: /env
        - command:
          - /bin/sh
          args:
            - -c
            - |
              source /env/secrets
              export AWS_SECRET_ACCESS_KEY=$AWS_SECRET_KEY
              export AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY
              export AWS_DEFAULT_REGION=eu-central-1
              aws sns publish --topic-arn arn:aws:sns:eu-central-1:1234567890:some-topic  --message "hello from the other side"
              sleep 86400
          image: mesosphere/aws-cli
          imagePullPolicy: Always
          name: app
          resources: {}
          stdin: true
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          tty: true
          volumeMounts:
          - name: env
            mountPath: /env
            readOnly: true
        serviceAccount: dev-example
        serviceAccountName: dev-example
        automountServiceAccountToken: true
        terminationGracePeriodSeconds: 30

```

## Configuration

You may configure the app using environment variables, as shown in the example. The following variables are supported:

* `VERBOSE`: Enables logging on `debug` level, otherwise logging is done on `info` level
* `KUBE_AUTH_PATH`: The path where the k8s auth method is mounted (used as `fmt.Sprintf("/v1/auth/%s/login", kubeAuthPath)`)
* `KUBE_AUTH_ROLE`: Used to tell the kubernetes auth method which role to assume (has to be defined in vault)
* `KubeTokenFile`: Where to load the k8s auth token from, useful for local development & testing (defaults to `/run/secrets/kubernetes.io/serviceaccount/token`)
* `VAULT_TOKEN_FILE`: Where to store the vault auth token fetched at ``, used to handover the token from `init` to `renew` container (defaults to `/env/vault-token`)
* `ENV_FILE`: Where to store the generated credentials in env format (defaults to `/env/secrets`)
* `PROCESSOR_STRATEGY`: Which config processor to use (means where to store the generated creds). Currently the only supported option is `env`, but this may be extended in the future

This container is logging in JSON format by default, using https://github.com/sirupsen/logrus. 

## License

    MIT License
    
    Copyright (c) 2018 Libri GmbH
