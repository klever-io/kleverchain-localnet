_GENESIS_JSON_ELEMENT = """
    {
        "address": "%s",
        "balance": %d,
        "kfiBalance": %d,
        "delegation": {
            "address": "%s",
            "value": %d
        }
    }
"""

_GENESIS_ROOT_ELEMENT = """
    {
        "address": "%s",
        "balance": %d,
        "kfiBalance": %d
    }
"""

_GENESIS = """
    [
    %s
    ]
"""

_NODE_SETUP_JSON_ELEMENT = """{
        "pubkey": "%s",
        "address": "%s"
    }"""

_NODE_SETUP = """{
  "startTime": %d,
  "slotInterval": 4000,
  "slotsPerEpoch": %d,
  "consensusGroupSize": %d,
  "minNodes": %d,
  "chainID": %s,
  "minTransactionVersion": 1,
  "klvDenomination": 6,
  "initialNodes": [
    %s
  ]
}
"""


_COMPOSER_BASE = """
version: '3'
services:
  seednode:
    image: kleverapp/klever-go:v1.7.16-0-gcf9f612c
    container_name: seednode
    restart: unless-stopped
    volumes:
      - ./config/seednode:/opt/klever-blockchain/config/seednode
    entrypoint: /usr/local/bin/seednode
    command: [
      "--rest-api-interface=0.0.0.0:8799",
    ]
    networks:
      klever:
        ipv4_address: 172.25.0.5
%s
networks:
  klever:
    driver: bridge
    ipam:
      config:
        - subnet: 172.25.0.0/24
"""

_COMPOSER_NODES = """
  node%s:
      image: kleverapp/klever-go:v1.7.16-0-gcf9f612c
      container_name: node%s
      restart: unless-stopped
      networks:
        - klever
      ports:
        - 88%s:88%s
      volumes:
        - ./config/node:/opt/klever-blockchain/config/node
        - %s/validatorKey.pem:/opt/klever-blockchain/config/validatorKey.pem
        - %s/walletKey.pem:/opt/klever-blockchain/config/walletKey.pem
        - ./dbs/node-%s/:/opt/klever-blockchain/db
        - ./logs/node-%s:/opt/klever-blockchain/logs
      entrypoint: /usr/local/bin/validator
      command: [
        "--log-level=*:INFO",
        "--use-log-view",
        "--validator-key-pem-file=./config/validatorKey.pem",
        "--rest-api-interface=0.0.0.0:88%s"
      ]"""
