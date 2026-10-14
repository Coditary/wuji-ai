# wuji-ai

Frontend CLI — includes **wuji-core** in-process (single `wuji` binary).

**Documentation:** [Wuji Wiki](https://github.com/Coditary/wuji-ai/wiki) (source in [`wuji-ai.wiki/`](wuji-ai.wiki/Home.md) when cloned with the repo).

Related repos in `~/Dev/Coditary/`:

- `core/wuji-core/` — backend daemon
- `plugins/wuji/` — external driver plugins

## Build

```bash
make build    # → bin/wuji
```

## Run

```bash
./bin/wuji text "hello"
```

Config lives in `~/.wuji/config.yaml` (or under `WUJI_ROOT` when set).

For development with a separate daemon: `export WUJI_DAEMON=1`, start `wuji-core`, then run `wuji`.

Driver inventory:

```bash
wuji list drivers              # active + configured plugins (no auto-start)
wuji list drivers --connect    # attach / auto-start plugins, then list
wuji info driver <id>          # details; shows plugin status when not yet registered
```

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

## Install via ReqPack

```bash
rqp install wuji-ai
# or via the Wuji driver catalog:
rqp install wuji wuji-ai
```

The installed command is `wuji` (not `wuji-ai`).
