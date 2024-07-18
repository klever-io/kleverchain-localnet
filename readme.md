# Kleverchain LocalNet

### Requirements

* Docker and Docker-compose
* Python 3.X

<hr>

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