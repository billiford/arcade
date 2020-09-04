# arcade

Arcade is meant to be used in tandem with Google's [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) to generate your tokens in a sidecar and make them retrievable through a simple  authenticated API.

### Build

```bash
go build cmd/arcade/arcade.go
```

### Run

```bash
export ARCADE_API_KEY=test
./arcade
```

### Test

```bash
curl localhost:1982/tokens -H "Api-Key: test"
```
