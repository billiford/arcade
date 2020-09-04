# arcade

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
