# Klever Fast Node Setup

Cross-platform setup script for running Klever blockchain nodes locally. Works on **Linux**, **macOS**, and **Windows**.

## 📋 Requirements

### All Platforms

- **Docker Desktop** (version 20.10+)
  - Linux: [Install Docker Engine](https://docs.docker.com/engine/install/)
  - macOS: [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/)
  - Windows: [Docker Desktop for Windows](https://docs.docker.com/desktop/install/windows-install/)
- **Python 3.7+** (for configuration generation scripts)
  - Verify: `python --version` or `python3 --version`

## 🚀 Quick Start

### For Linux users

Maybe you need to give permissions to dbs/ logs/ and keys/ to the nodes have permissions after setup localnet.

```bash
  chmod -R 777 dbs/* keys/* logs/*
```

### Complete Setup (One Command)

The easiest way to get started:

```bash
# Linux/macOS
python3 setup.py setup-all

# Windows
python setup.py setup-all
```

This automatically:
1. ✓ Checks all requirements
2. ✓ Generates validator and wallet keys
3. ✓ Creates necessary directories
4. ✓ Generates configuration files

after execution of setup-all you can run your localnet

```bash
# Linux/macOS
python3 setup.py start

# Windows
python setup.py start
```


### With Custom Configuration

```bash
# Setup with 3 validators
python setup.py setup-all -n 3

# Setup with custom max supply
python setup.py setup-all -n 5 -s 10000000000000000
```

### Monitor the Network

```bash
# Check container status
python setup.py status

# View logs (CTRL+C to exit)
python setup.py logs

# View logs of specific node
docker logs -f node-0
```

## 🔧 Available Commands

### Setup Commands

```bash
# Complete automated setup
python setup.py setup-all

# Individual steps
python setup.py check-requirements    # Verify dependencies
python setup.py generate-keys -n 3    # Generate keys for 3 validators
python setup.py generate-dirs -n 3    # Create directories
python setup.py create-localnet -n 3  # Generate configs
```

### Container Management

```bash
# Start containers
python setup.py start

# Stop containers
python setup.py down

# Restart containers
python setup.py restart

# Check status
python setup.py status

# View logs
python setup.py logs

# View logs without following
python setup.py logs --no-follow
```

### Cleanup

```bash
# Remove generated configs only
python setup.py clean

# Remove EVERYTHING (keys, dbs, logs, configs) - DESTRUCTIVE!
python setup.py clean-all
```

### Help

```bash
# Show all commands and options
python setup.py -h
```

## 🛠️ Advanced Usage

### Resetting the Blockchain

To completely reset the blockchain state:

```bash
# Clean everything and start fresh
python setup.py clean-all
python setup.py setup-all
```

### Custom Configuration

You can modify generated files before starting:

```bash
# Generate everything but don't start
python setup.py generate-keys -n 3
python setup.py generate-dirs -n 3
python setup.py create-localnet -n 3

# Modify configs/genesis.json or docker-compose.yaml as needed
# Edit configs/nodesSetup.json to change startTime if needed

# Then start
python setup.py start
```

## 📁 Project Structure

```
fast-node-setup/
├── keys/              # Generated validator and wallet keys
│   ├── node-0/
│   │   ├── validatorKey.pem
│   │   └── walletKey.pem
│   ├── node-1/
│   └── ...
├── dbs/               # Blockchain databases
│   ├── node-0/
│   ├── node-1/
│   └── ...
├── logs/              # Node logs
│   ├── node-0/
│   ├── node-1/
│   └── ...
├── configs/           # Generated configuration files
│   ├── genesis.json
│   ├── nodesSetup.json
│   └── ...
├── scripts/           # Helper scripts
├── docker-compose.yaml   # Generated Docker Compose file
├── setup.py          # Cross-platform setup script
└── readme.md
```
