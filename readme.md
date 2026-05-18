# Klever Localnet (`localnet`)

Single-binary tool to run a multi-validator Klever blockchain on your laptop. Linux, macOS, and Windows.

## Requirements

- Docker 20.10+ and Docker Compose v2 (`docker compose`, not the deprecated `docker-compose`).
- ~5 GiB free disk space.

## Install

### Homebrew (macOS / Linux)

```bash
brew install klever-io/tap/localnet
```

### Direct download

Grab the archive for your OS/arch from the [Releases page](https://github.com/klever-io/kleverchain-localnet/releases), extract `localnet`, and place it on your `PATH`.

### Build from source

```bash
git clone https://github.com/klever-io/kleverchain-localnet.git
cd kleverchain-localnet
make build   # produces ./bin/localnet
```

## Quick start

```bash
# One-shot: generate keys, configs, compose, then start containers
localnet run -n 3

# Or do it step by step:
localnet setup-all -n 3
localnet start

# Inspect
localnet status
localnet logs --node 0

# Stop and clean
localnet stop
localnet clean-all
```

Invoking `localnet` on a terminal with no arguments opens the **TUI dashboard** (splash, main menu, and a live monitor for container health and resource usage). Press `?` in any screen for help. Piping `localnet` to a file instead prints the CLI help — CI-friendly.

## Port layout

| Service   | REST API port |
|-----------|---------------|
| seednode  | 8799          |
| node0     | 8800          |
| node1     | 8801          |
| node-N    | 88NN          |

All services live on the `klever` bridge network (`172.25.0.0/24`).

## Command reference

Run `localnet --help` for the full tree. Key commands:

- `localnet run` — zero-to-running (doctor + keys + configs + compose + start).
- `localnet setup-all` — same as `run` without starting.
- `localnet start` / `stop` / `restart` — lifecycle.
- `localnet status` — container state + ports as a table.
- `localnet logs --node N` — tail logs for a specific node (use `--no-follow` for one-shot output).
- `localnet monitor` — open the live TUI dashboard directly.
- `localnet doctor` — verify Docker + disk + group membership.
- `localnet clean` / `clean-all` — remove generated artifacts (`clean-all` also wipes keys/dbs/logs and prompts for confirmation).
