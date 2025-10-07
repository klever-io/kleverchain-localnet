import os
import platform
from const import _COMPOSER_NODES,_COMPOSER_BASE

def generate_compose(validators):
    nodes_data = ""
    index = 0
    is_windows = platform.system() == 'Windows'

    for _,value in validators.items():
        index_with_zero = index if index >= 10 else f"0{index}"

        # Convert Windows path to Docker-compatible format
        path = value['path']
        if is_windows:
            # Convert backslashes to forward slashes for Docker
            path = path.replace('\\', '/')
            # Handle drive letter (C:/ -> /c/)
            if len(path) > 1 and path[1] == ':':
                drive = path[0].lower()
                path = f"/{drive}{path[2:]}"

        nodes_data += _COMPOSER_NODES % (index, index,index_with_zero, index_with_zero, path, index, index,index_with_zero)
        index = index + 1

    compose = _COMPOSER_BASE % nodes_data


    # Dump the dictionary to a .yaml file
    with open('docker-compose.yaml', 'w') as file:
        file.write(compose)