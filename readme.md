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
5. ✓ Starts Docker containers

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

## 📊 Network Information

- **Default Validators:** 1 (configurable)
- **Default Max Supply:** 10,000,000,000,000,000 KLV
- **API Endpoint:** http://localhost:8800
- **Network Type:** Local development network

## 🔧 Available Commands

### Setup Commands

```bash
# Complete automated setup
python setup.py setup-all

# Individual steps
python setup.py check-requirements   # Verify dependencies
python setup.py generate-keys -n 3   # Generate keys for 3 validators
python setup.py generate-dirs -n 3   # Create directories
python setup.py create-localnet      # Generate configs
```

### Container Management

```bash
# Start containers
python setup.py compose-up

# Stop containers
python setup.py compose-down

# Restart containers
python setup.py compose-restart

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

## 🔧 Configuration Options

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-n, --validators-num NUM` | 1 | Number of validator nodes |
| `-s, --max-supply AMOUNT` | 10000000000000000 | Maximum token supply |
| `--no-follow` | false | Don't follow logs in real-time |

### Examples

```bash
# Setup single validator (default)
python setup.py setup-all

# Setup 5 validators
python setup.py setup-all -n 5

# Just generate keys for 3 validators
python setup.py generate-keys -n 3

# View logs without following
python setup.py logs --no-follow
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
python setup.py compose-up
```

### Shell Access to Container

```bash
# Access node-0 container
docker compose exec node-0 sh
```

### Rebuild Containers

```bash
docker compose up -d --build --force-recreate
```

### Pull Latest Images

```bash
docker compose pull
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
├── Makefile          # Alternative Make-based setup (Linux only)
└── README.md
```

## 🐛 Troubleshooting

### "Docker is not installed or not in PATH"

**Solution:**
1. Install Docker Desktop
2. Make sure Docker is running
3. Restart your terminal
4. Run `docker --version` to verify

### "Python is not installed or not in PATH"

**Solution:**
1. Install Python 3.7+
2. On Windows, ensure "Add Python to PATH" was checked during installation
3. Restart your terminal
4. Run `python --version` or `python3 --version` to verify

### "docker compose is not available"

**Solution:**
- Update Docker Desktop to the latest version
- Docker Compose V2 is included with Docker Desktop
- On Linux, you may need: `sudo apt install docker-compose-plugin`

### Permission Issues (Linux/macOS)

If you get permission errors:

```bash
# Add your user to the docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

### Port Already in Use

```bash
# Check what's using the ports
python setup.py status

# Stop existing containers
python setup.py compose-down
```

## 🌐 Platform-Specific Notes

### Linux
- May require adding user to docker group
- Uses UID/GID mapping for proper file permissions
- Best performance

### macOS
- Uses Docker Desktop VM
- Fully compatible with the setup script

### Windows
- Requires WSL2 backend for Docker Desktop
- Paths automatically converted to Docker format
- No UID/GID mapping (not applicable)

## 💡 Why Python Script Instead of Makefile?

This Python script provides:
- ✅ True cross-platform compatibility (Linux, macOS, Windows)
- ✅ No additional dependencies (just Python and Docker)
- ✅ Better error handling and user feedback
- ✅ Colored output for better readability
- ✅ Automatic OS detection and path handling
- ✅ Easy to extend and customize
