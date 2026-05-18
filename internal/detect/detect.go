package detect

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type State int

const (
	Uninitialized State = iota
	KeysOnly
	Ready
	Running
	Partial
)

func (s State) String() string {
	switch s {
	case Uninitialized:
		return "Uninitialized"
	case KeysOnly:
		return "KeysOnly"
	case Ready:
		return "Ready"
	case Running:
		return "Running"
	case Partial:
		return "Partial"
	}
	return fmt.Sprintf("State(%d)", int(s))
}

type Snapshot struct {
	State           State
	Validators      int
	RunningServices int
	HealthyServices int
}

type Service interface {
	ServiceHealth() (state, health string)
}

type DockerProbe interface {
	ListServices(ctx context.Context) ([]DockerService, error)
}

type DockerService struct {
	Name   string
	State  string
	Health string
}

type Detector struct {
	WorkDir string
	Docker  DockerProbe
}

func (d *Detector) Detect(ctx context.Context) Snapshot {
	snap := Snapshot{}

	keysDir := filepath.Join(d.WorkDir, "keys")
	hasRootWallet := fileExists(filepath.Join(keysDir, "walletKey.pem"))
	nodeCount, partialNodes := countNodeKeys(keysDir)
	snap.Validators = nodeCount

	composeExists := fileExists(filepath.Join(d.WorkDir, "docker-compose.yaml")) ||
		fileExists(filepath.Join(d.WorkDir, "docker-compose.yml"))
	genesisExists := fileExists(filepath.Join(d.WorkDir, "config", "node", "genesis.json"))
	nodesSetupExists := fileExists(filepath.Join(d.WorkDir, "config", "node", "nodesSetup.json"))

	hasAnyKey := hasRootWallet || nodeCount > 0 || partialNodes > 0
	hasAllKeys := hasRootWallet && nodeCount > 0 && partialNodes == 0
	hasConfigs := genesisExists && nodesSetupExists
	hasCompose := composeExists

	if d.Docker != nil {
		services, err := d.Docker.ListServices(ctx)
		if err == nil {
			for _, s := range services {
				if s.State == "running" {
					snap.RunningServices++
					if s.Health == "healthy" || s.Health == "" {
						snap.HealthyServices++
					}
				}
			}
		}
	}

	switch {
	case snap.RunningServices > 0:
		snap.State = Running
	case !hasAnyKey && !hasCompose && !hasConfigs:
		snap.State = Uninitialized
	case hasAllKeys && hasConfigs && hasCompose:
		snap.State = Ready
	case hasAllKeys && !hasConfigs && !hasCompose:
		snap.State = KeysOnly
	default:
		snap.State = Partial
	}
	return snap
}

func countNodeKeys(keysDir string) (full int, partial int) {
	entries, err := os.ReadDir(keysDir)
	if err != nil {
		return 0, 0
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if len(e.Name()) < 5 || e.Name()[:5] != "node-" {
			continue
		}
		nodeDir := filepath.Join(keysDir, e.Name())
		v := fileExists(filepath.Join(nodeDir, "validatorKey.pem"))
		w := fileExists(filepath.Join(nodeDir, "walletKey.pem"))
		switch {
		case v && w:
			full++
		case v || w:
			partial++
		}
	}
	return full, partial
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
