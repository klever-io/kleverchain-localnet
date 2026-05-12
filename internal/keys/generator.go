package keys

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
)

type Result struct {
	Validators []domain.Validator
	Wallets    []domain.Wallet
	Root       *domain.Wallet
}

type Generator struct {
	Runner  dockercli.Runner
	Binary  string
	Image   string
	IsLinux bool
	UID     int
	GID     int
}

func NewGenerator(runner dockercli.Runner, image string) *Generator {
	if image == "" {
		image = domain.KleverImage
	}
	g := &Generator{
		Runner:  runner,
		Binary:  "docker",
		Image:   image,
		IsLinux: runtime.GOOS == "linux",
		UID:     os.Getuid(),
		GID:     os.Getgid(),
	}
	return g
}

func (g *Generator) Generate(ctx context.Context, state domain.LocalnetState, keysDir string, force bool) (Result, error) {
	if err := state.Validate(); err != nil {
		return Result{}, err
	}
	if err := fsutil.EnsureDir(keysDir, 0o700); err != nil {
		return Result{}, err
	}

	present := countPresentKeyFiles(keysDir, state.Validators)
	expected := 2*state.Validators + 1
	switch {
	case present == expected && !force:
		return g.readExisting(state, keysDir)
	case present == expected && force:
		if err := removeExistingKeys(keysDir, state.Validators); err != nil {
			return Result{}, err
		}
	case present > 0 && present < expected && !force:
		return Result{}, fmt.Errorf("%w: only %d/%d key files in %s; re-run with --force to regenerate", errs.ErrKeysIncomplete, present, expected, keysDir)
	case present > 0 && present < expected && force:
		if err := removeExistingKeys(keysDir, state.Validators); err != nil {
			return Result{}, err
		}
	}

	for i := 0; i < state.Validators; i++ {
		nodeDir := filepath.Join(keysDir, fmt.Sprintf("node-%d", i))
		if err := fsutil.EnsureDir(nodeDir, 0o700); err != nil {
			return Result{}, err
		}
		mount := nodeDir + ":/opt/klever-blockchain"
		name := fmt.Sprintf("klever-keygen-node-%d", i)
		args := g.dockerRunArgs(mount, name, "--num-keys", "1", "--key-type", "both")
		if _, err := g.Runner.Run(ctx, g.Binary, args, dockercli.RunOptions{Stdout: os.Stdout, Stderr: os.Stderr}); err != nil {
			return Result{}, fmt.Errorf("%w: keygen node-%d: %w", errs.ErrDockerFailure, i, err)
		}
		_ = tightenPerms(nodeDir)
	}

	rootMount := keysDir + ":/opt/klever-blockchain"
	rootArgs := g.dockerRunArgs(rootMount, "klever-keygen-root-wallet", "--num-keys", "1", "--key-type", "wallet")
	if _, err := g.Runner.Run(ctx, g.Binary, rootArgs, dockercli.RunOptions{Stdout: os.Stdout, Stderr: os.Stderr}); err != nil {
		return Result{}, fmt.Errorf("%w: root wallet keygen: %w", errs.ErrDockerFailure, err)
	}
	_ = tightenPerms(keysDir)

	return g.readExisting(state, keysDir)
}

func (g *Generator) dockerRunArgs(volumeMount, name string, keygenArgs ...string) []string {
	args := []string{"run", "--rm",
		"-v", volumeMount,
		"--name", name,
	}
	if g.IsLinux {
		args = append(args,
			"--user", strconv.Itoa(g.UID)+":"+strconv.Itoa(g.GID),
			"--group-add", "klever",
		)
	}
	args = append(args,
		"--entrypoint", "",
		g.Image,
		"keygenerator",
	)
	args = append(args, keygenArgs...)
	return args
}

func existsAllKeyFiles(keysDir string, validators int) bool {
	return countPresentKeyFiles(keysDir, validators) == 2*validators+1
}

func countPresentKeyFiles(keysDir string, validators int) int {
	n := 0
	if fsutil.Exists(filepath.Join(keysDir, "walletKey.pem")) {
		n++
	}
	for i := 0; i < validators; i++ {
		nodeDir := filepath.Join(keysDir, fmt.Sprintf("node-%d", i))
		if fsutil.Exists(filepath.Join(nodeDir, "validatorKey.pem")) {
			n++
		}
		if fsutil.Exists(filepath.Join(nodeDir, "walletKey.pem")) {
			n++
		}
	}
	return n
}

func removeExistingKeys(keysDir string, validators int) error {
	if err := fsutil.RemoveIfExists(filepath.Join(keysDir, "walletKey.pem")); err != nil {
		return err
	}
	for i := 0; i < validators; i++ {
		nodeDir := filepath.Join(keysDir, fmt.Sprintf("node-%d", i))
		if err := fsutil.RemoveIfExists(nodeDir); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) readExisting(state domain.LocalnetState, keysDir string) (Result, error) {
	if !existsAllKeyFiles(keysDir, state.Validators) {
		return Result{}, fmt.Errorf("%w: keys in %s are missing files; re-run with --force", errs.ErrKeysIncomplete, keysDir)
	}

	validators := make([]domain.Validator, 0, state.Validators)
	wallets := make([]domain.Wallet, 0, state.Validators)
	for i := 0; i < state.Validators; i++ {
		nodeDir := filepath.Join(keysDir, fmt.Sprintf("node-%d", i))
		vpath := filepath.Join(nodeDir, "validatorKey.pem")
		wpath := filepath.Join(nodeDir, "walletKey.pem")

		vb, err := os.ReadFile(vpath)
		if err != nil {
			return Result{}, fmt.Errorf("read %s: %w", vpath, err)
		}
		pub, err := ParsePEMHeader(vb)
		if err != nil {
			return Result{}, fmt.Errorf("parse %s: %w", vpath, err)
		}

		wb, err := os.ReadFile(wpath)
		if err != nil {
			return Result{}, fmt.Errorf("read %s: %w", wpath, err)
		}
		addr, err := ParsePEMHeader(wb)
		if err != nil {
			return Result{}, fmt.Errorf("parse %s: %w", wpath, err)
		}

		validators = append(validators, domain.Validator{
			Index:     i,
			PubKey:    pub,
			KeyPath:   vpath,
			WalletKey: wpath,
		})
		wallets = append(wallets, domain.Wallet{Address: addr, Path: wpath})
	}

	rootPath := filepath.Join(keysDir, "walletKey.pem")
	rb, err := os.ReadFile(rootPath)
	if err != nil {
		return Result{}, fmt.Errorf("read %s: %w", rootPath, err)
	}
	rootAddr, err := ParsePEMHeader(rb)
	if err != nil {
		return Result{}, fmt.Errorf("parse %s: %w", rootPath, err)
	}

	return Result{
		Validators: validators,
		Wallets:    wallets,
		Root:       &domain.Wallet{Address: rootAddr, Path: rootPath},
	}, nil
}

func tightenPerms(path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.Chmod(p, 0o700)
		}
		if filepath.Ext(p) == ".pem" {
			return os.Chmod(p, 0o600)
		}
		return nil
	})
}
