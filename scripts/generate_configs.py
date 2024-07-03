import json
import os

from datetime import datetime
from const import _GENESIS_JSON_ELEMENT,_NODE_SETUP_JSON_ELEMENT,_GENESIS,_NODE_SETUP

def generate_genesis(wallets,klv_supply,path):
    klvDelegation = 10_000_000_000_000
    kfiSupply = 21_000_000_000_000

    totalStaking = klvDelegation * len(wallets)
    eachKLV = (int(klv_supply) - totalStaking) // len(wallets)
    eachKFI = kfiSupply // len(wallets)

    wallet_text = ''
    for index,wallet in enumerate(wallets):
        wallet_text += _GENESIS_JSON_ELEMENT % (wallet, eachKLV, eachKFI, wallet, klvDelegation)

        if (index != len(wallets)-1):
            wallet_text += ","
    

    genesis_raw = _GENESIS % wallet_text

    genesis = json.loads(genesis_raw)


    with open(f"{path}/config/node/genesis.json","w") as genesis_file:
        json.dump(genesis,genesis_file,ensure_ascii=False, indent=4)

def generate_nodes_setup(wallets, validators, path):
    start_time = datetime.now().strftime("%s")
    start_time = int(start_time) + 21600 - (int(start_time) % 21600)
    chain_id = start_time // 21600
    
    slot_per_epoch = 20
    
    consensusGroupSize = int(os.getenv('CONSENSUS_GROUP_SIZE', 1))
    min_nodes = int(os.getenv('MIN_NODES', 1))
    if consensusGroupSize > min_nodes:
        min_nodes = consensusGroupSize

    node_text = ''
    for index, wallet in enumerate(wallets):
        bls = validators[index]
        node_text += _NODE_SETUP_JSON_ELEMENT % (bls, wallet)

        if (index != len(wallets)-1):
            node_text += ","
    node_raw = _NODE_SETUP % (start_time, slot_per_epoch, consensusGroupSize, min_nodes, f"\"{chain_id}\"", node_text)
 
    nodes_setup = json.loads(node_raw)

    with open(f"{path}/config/node/nodesSetup.json","w") as node_setup_file:
        json.dump(nodes_setup,node_setup_file,ensure_ascii=False, indent=4)
