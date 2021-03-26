# arcade

Arcade is meant to run as a sidecar to generate authorization tokens to be used as `kubectl` credentials and make them retrievable through a simple authenticated API.

## Providers

Arcade supports two authorization token providers:

1. Google
2. Rancher

### Google

Using google's [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity), Arcade retrieves the token of the active GCP account.

To reduce the number of calls to google, Arcade caches the token for 1 minute.

### Rancher

Use these variables to configure rancher

```sh
RANCHER_ENABLED= # Set to TRUE if rancher is a supported token provider
RANCHER_URL=     # Set to the 'login' endpoint of your rancher instance
RANCHER_USERNAME # Set to your rancher username
RANCHER_PASSWORD # Set to your rancher password
```

Rancher kubeconfig tokens have an expiration time and Arcade will cache the token until it has expired before calling rancher for a new one.

## Run Locally

You'll need default credentials set up. Run the following commands to build and generate a token.

### Build

```bash
go build cmd/arcade/arcade.go
```

### Run

```bash
export ARCADE_API_KEY=test
export RANCHER_ENABLED=TRUE
export RANCHER_URL=https://rancher.example.com/v3/activeDirectoryProviders/activedirectory?action=login
export RANCHER_USERNAME=myUsername
export RANCHER_PASSWORD=myPassword
./arcade
```

### Test

**Google**

```bash
curl localhost:1982/tokens?provider=google -H "Api-Key: test"
```

The default token provider is google, so this is equivalent to the call above
```bash
curl localhost:1982/tokens -H "Api-Key: test"
```

**Rancher**

```bash
curl localhost:1982/tokens?provider=rancher -H "Api-Key: test"
```

