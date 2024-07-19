from const import _COMPOSER_NODES,_COMPOSER_BASE

def generate_compose(validators):
    nodes_data = ""
    index = 0
    for _,value in validators.items():
        index_with_zero = index if index >= 10 else f"0{index}"
        nodes_data += _COMPOSER_NODES % (index, index,index_with_zero, index_with_zero, value['path'], index, index,index_with_zero)
        index = index + 1

    compose = _COMPOSER_BASE % nodes_data


    # Dump the dictionary to a .yaml file
    with open('docker-compose.yaml', 'w') as file:
        file.write(compose)