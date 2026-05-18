package state

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

const DefaultFileName = ".localnet-state.yaml"

func Load(path string) (domain.LocalnetState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.LocalnetState{}, fmt.Errorf("%w: %w", os.ErrNotExist, err)
		}
		return domain.LocalnetState{}, fmt.Errorf("read %s: %w", path, err)
	}
	var s domain.LocalnetState
	if err := yaml.Unmarshal(b, &s); err != nil {
		return domain.LocalnetState{}, fmt.Errorf("%w: decode %s: %w", errs.ErrInvalidInput, path, err)
	}
	if s.Version == 0 {
		return domain.LocalnetState{}, fmt.Errorf("%w: %s missing version field", errs.ErrInvalidInput, path)
	}
	if s.Version > domain.StateSchemaVersion {
		return domain.LocalnetState{}, fmt.Errorf("%w: %s uses schema version %d; this binary supports %d", errs.ErrInvalidInput, path, s.Version, domain.StateSchemaVersion)
	}
	return s, nil
}

func Save(path string, s domain.LocalnetState) error {
	if s.Version == 0 {
		s.Version = domain.StateSchemaVersion
	}
	b, err := yaml.Marshal(&s)
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
