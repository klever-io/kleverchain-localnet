#!/usr/bin/env python3
"""
Klever Fast Node Setup - Cross-Platform Setup Script
Works on Linux, macOS, and Windows
"""

import os
import sys
import subprocess
import argparse
import platform
import shutil
from pathlib import Path


class Colors:
    """ANSI color codes for terminal output"""
    RESET = '\033[0m'
    BOLD = '\033[1m'
    RED = '\033[91m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    MAGENTA = '\033[95m'
    CYAN = '\033[96m'

    @staticmethod
    def disable():
        """Disable colors (for Windows cmd.exe or when piping)"""
        Colors.RESET = ''
        Colors.BOLD = ''
        Colors.RED = ''
        Colors.GREEN = ''
        Colors.YELLOW = ''
        Colors.BLUE = ''
        Colors.MAGENTA = ''
        Colors.CYAN = ''


# Disable colors on Windows unless using Windows Terminal or ConEmu
if platform.system() == 'Windows' and not os.getenv('WT_SESSION') and not os.getenv('ConEmuPID'):
    Colors.disable()


class SetupManager:
    """Main setup manager for Klever node setup"""

    def __init__(self, validators_num=1, max_supply=10_000_000_000_000_000):
        self.validators_num = validators_num
        self.max_supply = max_supply
        self.consensus_group_size = validators_num
        self.base_dir = Path.cwd()
        self.is_windows = platform.system() == 'Windows'
        self.is_linux = platform.system() == 'Linux'
        self.is_macos = platform.system() == 'Darwin'

    def print_header(self, text):
        """Print a formatted header"""
        print(f"\n{Colors.CYAN}{Colors.BOLD}{'='*60}{Colors.RESET}")
        print(f"{Colors.CYAN}{Colors.BOLD}{text.center(60)}{Colors.RESET}")
        print(f"{Colors.CYAN}{Colors.BOLD}{'='*60}{Colors.RESET}\n")

    def print_success(self, text):
        """Print success message"""
        symbol = "[+]" if self.is_windows else "✓"
        print(f"{Colors.GREEN}{symbol} {text}{Colors.RESET}")

    def print_error(self, text):
        """Print error message"""
        symbol = "[x]" if self.is_windows else "✗"
        print(f"{Colors.RED}{symbol} {text}{Colors.RESET}")

    def print_info(self, text):
        """Print info message"""
        symbol = "[i]" if self.is_windows else "ℹ"
        print(f"{Colors.BLUE}{symbol} {text}{Colors.RESET}")

    def print_warning(self, text):
        """Print warning message"""
        symbol = "[!]" if self.is_windows else "⚠"
        print(f"{Colors.YELLOW}{symbol} {text}{Colors.RESET}")

    def run_command(self, cmd, capture_output=False, check=True, shell=False):
        """Run a shell command with proper error handling"""
        try:
            if capture_output:
                result = subprocess.run(
                    cmd,
                    check=check,
                    capture_output=True,
                    text=True,
                    shell=shell
                )
                return result
            else:
                result = subprocess.run(cmd, check=check, shell=shell)
                return result
        except subprocess.CalledProcessError as e:
            if check:
                self.print_error(f"Command failed: {' '.join(cmd) if isinstance(cmd, list) else cmd}")
                if capture_output and e.stderr:
                    print(e.stderr)
                sys.exit(1)
            return e

    def check_requirements(self):
        """Check if all required tools are installed"""
        self.print_header("Checking Requirements")

        requirements = {
            'docker': ['docker', '--version'],
            'python': [sys.executable, '--version'],
        }

        all_satisfied = True

        for name, cmd in requirements.items():
            try:
                result = self.run_command(cmd, capture_output=True, check=False)
                if result.returncode == 0:
                    version = result.stdout.strip().split('\n')[0]
                    self.print_success(f"{name}: {version}")
                else:
                    self.print_error(f"{name} is not installed or not in PATH")
                    all_satisfied = False
            except Exception as e:
                self.print_error(f"{name} check failed: {e}")
                all_satisfied = False

        # Check docker compose
        try:
            result = self.run_command(['docker', 'compose', 'version'], capture_output=True, check=False)
            if result.returncode == 0:
                version = result.stdout.strip()
                self.print_success(f"docker compose: {version}")
            else:
                self.print_error("docker compose is not available")
                all_satisfied = False
        except Exception:
            self.print_error("docker compose check failed")
            all_satisfied = False

        if not all_satisfied:
            self.print_error("\nPlease install missing requirements before continuing")
            sys.exit(1)

        self.print_success("\nAll requirements satisfied!")
        return True

    def generate_keys(self):
        """Generate validator and wallet keys using Docker"""
        self.print_header(f"Generating Keys for {self.validators_num} Validator(s)")

        keys_dir = self.base_dir / 'keys'
        keys_dir.mkdir(exist_ok=True)

        # Docker command - different for Windows vs Unix
        volume_mount = f"{keys_dir.absolute()}:/opt/klever-blockchain"

        # On Windows, we need to convert the path format
        if self.is_windows:
            # Convert Windows path to Docker-compatible format
            volume_path = str(keys_dir.absolute()).replace('\\', '/')
            # Handle drive letter (C: -> /c/)
            if ':' in volume_path:
                drive, path = volume_path.split(':', 1)
                volume_path = f"/{drive.lower()}{path}"
            volume_mount = f"{volume_path}:/opt/klever-blockchain"

        cmd = [
            'docker', 'run', '--rm',
            '-v', volume_mount,
            '--entrypoint', '',
            'kleverapp/klever-go:latest',
            'keygenerator',
            '--num-keys', str(self.validators_num),
            '--key-type', 'both'
        ]

        self.print_info(f"Running: docker run ... keygenerator --num-keys {self.validators_num}")
        self.run_command(cmd)
        self.print_success("Keys generated successfully!")

    def generate_dirs(self):
        """Create node directories for databases and logs"""
        self.print_header(f"Creating Directories for {self.validators_num} Validator(s)")

        for i in range(self.validators_num):
            db_dir = self.base_dir / 'dbs' / f'node-{i}'
            log_dir = self.base_dir / 'logs' / f'node-{i}'

            db_dir.mkdir(parents=True, exist_ok=True)
            log_dir.mkdir(parents=True, exist_ok=True)

            self.print_info(f"Created directories for node-{i}")

        self.print_success("Directories created successfully!")

    def create_localnet(self):
        """Generate genesis and configuration files"""
        self.print_header("Generating Localnet Configuration")

        # Set environment variables for the generate.py script
        env = os.environ.copy()
        env['VALIDATORS_NUM'] = str(self.validators_num)
        env['MAX_SUPPLY'] = str(self.max_supply)
        env['CONSENSUS_GROUP_SIZE'] = str(self.consensus_group_size)

        # Run the generate.py script
        generate_script = self.base_dir / 'scripts' / 'generate.py'

        if not generate_script.exists():
            self.print_error(f"Generate script not found: {generate_script}")
            sys.exit(1)

        self.print_info("Running configuration generator...")
        cmd = [sys.executable, str(generate_script)]

        result = subprocess.run(cmd, env=env, check=True)

        self.print_success("Configuration generated successfully!")

    def start(self):
        """Start Docker containers"""
        self.print_header("Starting Docker Containers")

        compose_file = self.base_dir / 'docker-compose.yaml'
        if not compose_file.exists():
            compose_file = self.base_dir / 'docker-compose.yml'

        if not compose_file.exists():
            self.print_error("docker-compose.yaml not found. Run 'create-localnet' first.")
            sys.exit(1)

        self.print_info("Starting containers in detached mode...")
        self.run_command(['docker', 'compose', 'up', '-d'])
        self.print_success("Containers started successfully!")
        self.print_info("Run 'python setup.py status' to check container status")

    def down(self):
        """Stop Docker containers"""
        self.print_header("Stopping Docker Containers")

        self.print_info("Stopping containers...")
        self.run_command(['docker', 'compose', 'down'])
        self.print_success("Containers stopped!")

    def restart(self):
        """Restart Docker containers"""
        self.print_header("Restarting Docker Containers")

        self.print_info("Restarting containers...")
        self.run_command(['docker', 'compose', 'restart'])
        self.print_success("Containers restarted!")

    def status(self):
        """Show container status"""
        self.print_header("Container Status")
        self.run_command(['docker', 'compose', 'ps'])

    def logs(self, follow=True):
        """Show container logs"""
        self.print_header("Container Logs")
        cmd = ['docker', 'compose', 'logs']
        if follow:
            cmd.append('-f')
        self.run_command(cmd)

    def clean(self):
        """Remove generated configuration files"""
        self.print_header("Cleaning Generated Configs")

        items_to_remove = [
            self.base_dir / 'configs',
            self.base_dir / 'docker-compose.yml',
            self.base_dir / 'docker-compose.yaml',
        ]

        for item in items_to_remove:
            if item.exists():
                if item.is_dir():
                    shutil.rmtree(item)
                    self.print_info(f"Removed directory: {item.name}")
                else:
                    item.unlink()
                    self.print_info(f"Removed file: {item.name}")

        self.print_success("Configs cleaned!")

    def clean_all(self):
        """Remove all generated files"""
        self.print_header("Cleaning Everything")

        self.print_warning("WARNING: This will delete all keys, databases, and logs!")
        response = input("Are you sure you want to continue? (yes/no): ")

        if response.lower() not in ['yes', 'y']:
            self.print_info("Cleanup cancelled")
            return

        items_to_remove = [
            self.base_dir / 'keys',
            self.base_dir / 'dbs',
            self.base_dir / 'logs',
            self.base_dir / 'configs',
            self.base_dir / 'docker-compose.yml',
            self.base_dir / 'docker-compose.yaml',
        ]

        for item in items_to_remove:
            if item.exists():
                if item.is_dir():
                    shutil.rmtree(item)
                    self.print_info(f"Removed directory: {item.name}")
                else:
                    item.unlink()
                    self.print_info(f"Removed file: {item.name}")

        self.print_success("Everything cleaned!")

    def setup_all(self):
        """Complete setup from scratch"""
        self.print_header("Starting Complete Setup")

        print(f"{Colors.BOLD}Configuration:{Colors.RESET}")
        print(f"  Validators: {self.validators_num}")
        print(f"  Max Supply: {self.max_supply:,}")
        print(f"  Platform: {platform.system()} {platform.release()}")
        print()

        # Run all setup steps
        self.check_requirements()
        self.generate_keys()
        self.generate_dirs()
        self.create_localnet()

        self.print_header("Setup Complete!")
        print(f"{Colors.GREEN}Your Klever localnet is now running!{Colors.RESET}\n")
        print(f"Next steps:")
        print(f"  • Run Blockchain status: {Colors.CYAN}python setup.py start{Colors.RESET}")
        print(f"After run:")
        print(f"  • Check status: {Colors.CYAN}python setup.py status{Colors.RESET}")
        print(f"  • View logs:    {Colors.CYAN}python setup.py logs{Colors.RESET}")
        print(f"  • Stop nodes:   {Colors.CYAN}python setup.py stop{Colors.RESET}")
        print()


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description='Klever Fast Node Setup - Cross-Platform Setup Script',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s setup-all                          # Complete setup with defaults
  %(prog)s setup-all -n 3                     # Setup with 3 validators
  %(prog)s generate-keys -n 5                 # Generate keys for 5 validators
  %(prog)s start                              # Start containers
  %(prog)s status                             # Check status
  %(prog)s logs                               # View logs
  %(prog)s clean-all                          # Clean everything

Available commands:
  setup-all         Complete setup (keys + dirs + localnet + compose up)
  check-requirements Check if all dependencies are installed
  generate-keys     Generate validator and wallet keys
  generate-dirs     Generate node directories (dbs, logs)
  create-localnet   Generate configuration files
  start        Start Docker containers
  stop     Stop Docker containers
  restart   Restart Docker containers
  status            Show container status
  logs              Show container logs (use --no-follow to disable following)
  clean             Clean generated configs
  clean-all         Clean everything (configs, keys, dbs, logs)
        """
    )

    parser.add_argument(
        'command',
        help='Command to execute'
    )

    parser.add_argument(
        '-n', '--validators-num',
        type=int,
        default=1,
        help='Number of validator nodes (default: 1)'
    )

    parser.add_argument(
        '-s', '--max-supply',
        type=int,
        default=10_000_000_000_000_000,
        help='Maximum token supply (default: 10000000000000000)'
    )

    parser.add_argument(
        '--no-follow',
        action='store_true',
        help='Do not follow logs (for logs command)'
    )

    args = parser.parse_args()

    # Create setup manager
    manager = SetupManager(
        validators_num=args.validators_num,
        max_supply=args.max_supply
    )

    # Command mapping
    commands = {
        'setup-all': manager.setup_all,
        'check-requirements': manager.check_requirements,
        'generate-keys': manager.generate_keys,
        'generate-dirs': manager.generate_dirs,
        'create-localnet': manager.create_localnet,
        'start': manager.start,
        'stop': manager.stop,
        'restart': manager.restart,
        'status': manager.status,
        'logs': lambda: manager.logs(follow=not args.no_follow),
        'clean': manager.clean,
        'clean-all': manager.clean_all,
    }

    # Execute command
    if args.command not in commands:
        print(f"{Colors.RED}Error: Unknown command '{args.command}'{Colors.RESET}\n")
        parser.print_help()
        sys.exit(1)

    try:
        commands[args.command]()
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Interrupted by user{Colors.RESET}")
        sys.exit(0)
    except Exception as e:
        print(f"{Colors.RED}Error: {e}{Colors.RESET}")
        sys.exit(1)


if __name__ == '__main__':
    main()
