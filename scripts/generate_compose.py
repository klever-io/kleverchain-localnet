from const import _COMPOSER_NODES,_COMPOSER_BASE

def generate_compose(validators):
    nodes_data = ""
    for index,_ in enumerate(validators):
        nodes_data += _COMPOSER_NODES % (index, index, index, index, index,index if index >= 10 else f"0{index}")

    compose = _COMPOSER_BASE % nodes_data


    # Dump the dictionary to a .yaml file
    with open('docker-compose.yaml', 'w') as file:
        file.write(compose)