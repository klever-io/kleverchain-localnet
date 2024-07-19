import os

from utils import get_pem_header
from generate_configs import generate_genesis,generate_nodes_setup
from generate_compose import generate_compose

num_validators = int(os.getenv("VALIDATORS_NUM", 1))
max_supply = int(os.getenv("MAX_SUPPLY", 10_000_000_000_000_000))

validators = {}
wallets = {}

def read_keys():
    if num_validators == 1:
        base_validators_path = f'{base_dir}/keys'
        for (dirpath, _, _) in  os.walk(base_validators_path):
            node_name = "node-0"

            # Get validator public key
            pubkey = get_pem_header(dirpath + "/validatorKey.pem")
            address = get_pem_header(dirpath + "/walletKey.pem")

            validators[node_name] = {"path": dirpath, "pubkey": pubkey}
            wallets[node_name] = {"path": dirpath, "address": address}
    else:
        base_validators_path = f'{base_dir}/keys/'
        for (dirpath, _, _) in  os.walk(base_validators_path):
            if dirpath == base_validators_path:
                continue

            node_name = dirpath.split("/keys/")[1]

            # Get validator public key
            pubkey = get_pem_header(dirpath + "/validatorKey.pem")

            validators[node_name] = {"path": dirpath, "pubkey": pubkey}

        base_wallet_path = f'{base_dir}/keys'
        for (dirpath, _, _) in  os.walk(base_wallet_path):
            if dirpath == base_wallet_path:
                continue

            node_name = dirpath.split("/keys/")[1]

            # Get address
            address = get_pem_header(dirpath + "/walletKey.pem")

            wallets[node_name] = {"path": dirpath, "address": address}

def main():
    read_keys()

    orderedWallets = []
    orderedValidators = []

    for i in range(num_validators):
        key = f'node-{i}'
        orderedWallets.append(wallets[key]['address'])

    for i in range(num_validators):
        key = f'node-{i}'
        orderedValidators.append(validators[key]['pubkey'])

    print("Generating Genesis...")
    # generate genesis
    generate_genesis(wallets=orderedWallets,
                     klv_supply=max_supply,
                     path=base_dir
                     )

    print("Generating Nodes Setup...")
    # generate nodes setup
    generate_nodes_setup(wallets=orderedWallets,
                         validators=orderedValidators,
                         path=base_dir
                        )
    
    print("Generating Docker-compose...")
    # print(validators)
    generate_compose(validators=validators)

if __name__ == "__main__":
    global base_dir
    base_dir = os.getcwd()
    main()