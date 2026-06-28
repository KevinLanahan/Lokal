# cidb

A step-through debugger for GitHub Actions pipelines.

Instead of commit → push → wait → fail → repeat, `cidb` runs your workflow locally in Docker and pauses before each step — so you can inspect the environment, skip steps, drop into a live shell, and debug in real time.

## Requirements

- [Go 1.22+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

## Install

```bash
git clone https://github.com/KevinLanahan/cidb.git
cd cidb
go build -o cidb .
```

## Usage

Run a specific workflow file:
```bash
./cidb run .github/workflows/ci.yml
```

Or let cidb auto-discover a workflow in the current directory:
```bash
./cidb run
```

## Controls

At each step pause prompt:

| Key | Action |
|-----|--------|
| `c` | Run the step |
| `s` | Skip the step |
| `sh` | Drop into a shell inside the container |
| `a` | Abort the run |

If a step fails, cidb pauses again and lets you drop into a shell to inspect the environment before deciding what to do next.

## Status

Early v1 — supports `run:` steps only. `uses:` steps (e.g. `actions/checkout`) are skipped with a notice. Secrets, matrix builds, and expression syntax coming later.
