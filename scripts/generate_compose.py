import os
import platform
from const import _COMPOSER_NODES,_COMPOSER_BASE

def generate_compose(validators):
    nodes_data = []
    index = 0
    is_windows = platform.system() == 'Windows'

    for value in validators.values():
        index_with_zero = index if index >= 10 else f"0{index}"

        # Use relative path so docker-compose works on any OS/user
        path = os.path.relpath(value['path'])
        if is_windows:
            # Convert backslashes to forward slashes for Docker
            path = path.replace('\\', '/')
            # Handle drive letter (C:/ -> /c/)
            if len(path) > 1 and path[1] == ':':
                drive = path[0].lower()
                path = f"/{drive}{path[2:]}"
        else:
            path = f"./{path}"

        nodes_data.append(_COMPOSER_NODES % (index, index, index_with_zero, index_with_zero, path, path, index, index, index_with_zero))
        index += 1

    compose = _COMPOSER_BASE % ''.join(nodes_data)

    with open('docker-compose.yaml', 'w') as file:
        file.write(compose)
