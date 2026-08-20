# Archive Orchestrator

Archive Orchestrator is a Go service for scheduling file archives, verifying archive batches, restoring data, and retrying failed background jobs. It serves a small static dashboard from `web/`.

## Requirements

- Go 1.26 or later
- Docker Desktop, only for container validation

## Local Commands

```bash
go vet ./...
go test ./...
go test -race ./...
go build ./...
go run ./cmd/server
```

The default HTTP address is `:8080`. Use `ARCHIVE_ADDR` to override it. Runtime state is written to `./data/state.json`; set `ARCHIVE_DATA` and `ARCHIVE_RESTORE_ROOT` to customize the state file and allowed restore directory.

```bash
ARCHIVE_ADDR=127.0.0.1:8080 go run ./cmd/server
curl http://127.0.0.1:8080/healthz
```

## Make Targets

```bash
make check
make run
make build
```

`make build` writes the executable to `bin/archive-orchestrator`.

## Container Validation

The provided evaluation image retains the Go toolchain and supports both `linux/arm64` and `linux/amd64` through the official multi-architecture Go image.

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh archive-orchestrator linux/arm64
./build_benzhi_docker.sh archive-orchestrator linux/amd64
docker run --rm -it archive-orchestrator:latest go build ./...
```
