# wuji-ai

Frontend CLI — talks to **wuji-core** over gRPC.

Related repos in `~/Dev/Coditary/`:

- `core/wuji-core/` — backend daemon
- `plugins/wuji/` — external driver plugins

## Build

```bash
make build    # → bin/wuji
```

## Run

```bash
# Start core (separate terminal)
cd ~/Dev/Coditary/core/wuji-core && ./bin/wuji-core

# CLI (auto-finds core when .wuji exists under Coditary layout)
./bin/wuji text "hello"
```

Override backend location: `export WUJI_ROOT=~/Dev/Coditary/core/wuji-core`

## CI

GitHub Actions runs on pushes and pull requests to `main`/`master`:

| Job | Command |
|-----|---------|
| fmt | `make fmt` |
| lint | `golangci-lint run ./...` |
| build-test | `make build`, `go vet ./...`, `go test ./...` |
| cover-check | `make cover-check` (≥ 80% on `./internal/clix`) |

CI checks out **wuji-core** alongside this repo and rewrites the `go.mod` replace to `./wuji-core`.

Local equivalent:

```bash
make ci
```

## Release

Pushing a tag `v*` triggers a release build for four platforms:

- `linux-x86_64`, `linux-aarch64`
- `macos-x86_64`, `macos-aarch64`

Each matrix leg builds the `wuji` binary (version from the tag via `-ldflags`), then packages:

- `wuji-{version}-{platform}-{arch}.tar.gz` — binary + README
- `wuji-{version}-{platform}-{arch}.rqp` — ReqPack installable package

The release job uploads all tarballs, `.rqp` files, and a combined `index.json` ReqPack repository index to the GitHub release.
