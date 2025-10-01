# Kleverchain LocalNet

A local development environment for running a Kleverchain blockchain network using Docker containers.

## 📋 Requirements

- **Docker** (version 20.10+)
- **Docker Compose** (version 1.29+)
- **Python 3.x** (for configuration generation scripts)
- **Make** (for running build commands)

## 🚀 Quick Start

### Step 1: Generate Validator Keys

Generate cryptographic keys for validators. By default, this creates keys for 1 validator.

```bash
make generate-keys
```

**To create multiple validators:**
```bash
VALIDATORS_NUM=3 make generate-keys
```

> **Note:** You can also modify the `VALIDATORS_NUM` variable in the Makefile if you prefer.

### Step 2: Set Key Permissions

Ensure the generated keys have the correct permissions:

```bash
chmod 600 keys/*/*.pem
```

### Step 3: Generate Network Configuration

Create the local network configuration files and Docker Compose setup:

```bash
make create-localnet
```

**With custom parameters:**
```bash
VALIDATORS_NUM=3 MAX_SUPPLY=20000000000000000 make create-localnet
```

### Step 4: Create Directory Structure

Generate the required directories for logs and databases:

```bash
make generate-dirs
```

### Step 5: Set Directory Permissions

Ensure Docker containers can write to the directories:

```bash
chmod -R 755 logs/ dbs/
```

### Step 6: Start the Network

Launch the blockchain network:

```bash
make compose-up
```

### Step 7: Monitor the Network

Check the logs of the first validator node:

```bash
docker logs --tail 20 -f node0
```

## 📊 Network Information

- **Default Validators:** 1 (configurable)
- **Default Max Supply:** 10,000,000,000,000,000 KLV
- **API Endpoint:** http://localhost:8800
- **Network Type:** Local development network

## 🔧 Configuration Options

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VALIDATORS_NUM` | 1 | Number of validator nodes |
| `MAX_SUPPLY` | 10000000000000000 | Maximum token supply |
| `CONSENSUS_GROUP_SIZE` | Same as `VALIDATORS_NUM` | Size of consensus group |

### Genesis Configuration

To modify the genesis time, edit the `startTime` field in `config/node/nodesSetup.json` after running `make create-localnet`.

## 🛠️ Advanced Usage

### Restarting Nodes with Sync

To restart a node and sync from the beginning, add the `--start-in-sync` flag to the Docker command:

```yaml
command: [
  "--log-level=*:INFO",
  "--use-log-view",
  "--validator-key-pem-file=./config/validatorKey.pem",
  "--rest-api-interface=0.0.0.0:8800",
  "--start-in-sync"
]
```

### Resetting the Blockchain

To completely reset the blockchain state:

```bash
# Stop the network
docker-compose down

# Remove blockchain data
rm -rf dbs/

# Regenerate directories and restart
make generate-dirs
# It's important to recreate the localnet, or you can change the startTime in nodesSetup.json
make create-localnet
make compose-up
```


### Useful Commands

```bash
# Stop the network
docker-compose down

# View all running containers
docker-compose ps

# Follow logs from all nodes
docker-compose logs -f

# Access a specific node's logs
docker logs node0

# Clean up everything (containers, volumes, networks)
docker-compose down --volumes --remove-orphans
```

## 📁 Project Structure

```
kleverchain-localnet/
├── config/           # Node configuration files
├── keys/            # Generated validator keys
├── logs/            # Node log files
├── dbs/             # Blockchain database files
├── scripts/         # Python generation scripts
├── docker-compose.yaml
├── Makefile
└── README.md
```
