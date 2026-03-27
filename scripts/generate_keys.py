import os
import shutil
import subprocess

num_validators = int(os.getenv("VALIDATORS_NUM", 1))

def generate_keys(volume_path, key_type, num_keys, node_container, validator=True):
    try:
        if not os.path.exists(volume_path):
            os.makedirs(volume_path)

        command = [
            'docker', 'run', '--rm',
            '-v', f'{os.path.abspath(volume_path)}:/opt/klever-blockchain',
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

        if validator and num_keys == 1:
            node_dir = os.path.join(volume_path, "node-0")
            os.makedirs(node_dir, exist_ok=True)
            for file in os.listdir(volume_path):
                src = os.path.join(volume_path, file)
                if os.path.isfile(src):
                    shutil.move(src, os.path.join(node_dir, file))
    except subprocess.CalledProcessError as e:
        raise Exception(f"docker error: {e}")

if __name__ == "__main__":
    # Generate the wallets and validators keys
    base_dir = os.getcwd()
    print("Generating Keys...")
    generate_keys(f'{base_dir}/keys', 'validator', num_validators, "generate_keys_validator")
    generate_keys(f'{base_dir}/keys', 'wallet', num_validators, "generate_keys_wallet")
    generate_keys(f'{base_dir}/keys', 'wallet', 1, "generate_keys_root", validator=False)