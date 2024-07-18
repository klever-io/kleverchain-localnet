import os
import subprocess

def generate_keys(volume_path, key_type, num_keys, node_container):
    try:
        path = volume_path
        if num_keys == 1:
            if not os.path.exists(path):
                os.makedirs(path)
            path += "/node-0/"

        command = [
            'docker', 'run', '--rm',
            '-v', f'{os.path.abspath(path)}:/opt/klever-blockchain',
            '--user', f'{os.getuid()}:{os.getgid()}',
            '--name', node_container,
            '--entrypoint', '',
            'kleverapp/klever-go:latest',
            'keygenerator',
            '--num-keys', str(num_keys),
            '--key-type', key_type
        ]
        result = subprocess.run(command, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        print(f"Output for {key_type} keys:\n{result.stdout.decode('utf-8')}")
    except subprocess.CalledProcessError as e:
        raise Exception(f"docker error: {e}")
