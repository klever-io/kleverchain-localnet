# Kleverchain LocalNet

### Requirements

* Docker and Docker-compose
* Python 3.X


### Instructions

1. Generate Keys

If you want more validators change the **VALIDATORS_NUM** const in Makefile

```bash
    make generate_keys
```

2. Give permissions to keys if needed.

3. Generate LocalNet configs and Docker-compose

```bash
    make create-localnet
```

4. Run docker-compose !!

```bash
    make compose-up
```

5. Checking logs of node 0

```bash
    docker logs --tail 5 -f node0
```

<hr>

* if **make create-localnet** gives permission error, you need to give permission to the keys folder
* if you want to change the genesis time, just change the  **startTime** in nodesSetup.json after generate.
* if you want reset blockchain you can delete the dbs generated after you started the blockchain.
* if you want restart the node just add --start-in-sync to the node startup

```bash
    command: [
    "--log-level=*:INFO",
    "--use-log-view",
    "--validator-key-pem-file=./config/validatorKey.pem",
    "--rest-api-interface=0.0.0.0:8800",
    "--start-in-sync"
    ]
```