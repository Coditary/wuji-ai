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
