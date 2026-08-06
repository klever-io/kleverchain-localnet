package compose

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

type ValidatorService struct {
	Index            int
	RESTPort         int
	ValidatorKeyPath string
}

type Resources struct {
	MemLimit   string
	CPUs       string
	LogMaxSize string
	LogMaxFile string
	Restart    bool
}

func DefaultResources() Resources {
	return Resources{
		MemLimit:   "2g",
		CPUs:       "1.0",
		LogMaxSize: "50m",
		LogMaxFile: "5",
		Restart:    true,
	}
}

type templateData struct {
	Image            string
	SeednodeImage    string
	NetworkName      string
	NetworkSubnet    string
	SeednodeIP       string
	SeednodeRESTPort int
	MemLimit         string
	CPUs             string
	LogMaxSize       string
	LogMaxFile       string
	Restart          bool
	Validators       []ValidatorService
}

func Build(state domain.LocalnetState, services []ValidatorService, res Resources) (string, error) {
	if err := state.Validate(); err != nil {
		return "", err
	}
	if len(services) != state.Validators {
		return "", fmt.Errorf("%w: services=%d does not match validators=%d", errs.ErrInvalidInput, len(services), state.Validators)
	}
	for _, s := range services {
		if err := validateMountPath(s.ValidatorKeyPath); err != nil {
			return "", err
		}
	}

	ordered := make([]ValidatorService, len(services))
	copy(ordered, services)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Index < ordered[j].Index })

	data := templateData{
		Image:            state.KleverImage,
		SeednodeImage:    domain.KleverImage,
		NetworkName:      domain.NetworkName,
		NetworkSubnet:    domain.NetworkSubnet,
		SeednodeIP:       domain.SeednodeStaticIP,
		SeednodeRESTPort: domain.SeednodeRESTPort,
		MemLimit:         res.MemLimit,
		CPUs:             res.CPUs,
		LogMaxSize:       res.LogMaxSize,
		LogMaxFile:       res.LogMaxFile,
		Restart:          res.Restart,
		Validators:       ordered,
	}

	tmpl, err := template.New("compose").Parse(composeTemplate)
	if err != nil {
		return "", fmt.Errorf("parse compose template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute compose template: %w", err)
	}
	return buf.String(), nil
}

func BuildValidatorServices(state domain.LocalnetState, keysDir string) ([]ValidatorService, error) {
	out := make([]ValidatorService, 0, state.Validators)
	for i := 0; i < state.Validators; i++ {
		rel := filepath.ToSlash(filepath.Join(keysDir, fmt.Sprintf("node-%d", i), "validatorKey.pem"))
		if !strings.HasPrefix(rel, "./") && !strings.HasPrefix(rel, "/") {
			rel = "./" + rel
		}
		out = append(out, ValidatorService{
			Index:            i,
			RESTPort:         domain.ValidatorPortBase + i,
			ValidatorKeyPath: rel,
		})
	}
	return out, nil
}

func validateMountPath(p string) error {
	if p == "" {
		return fmt.Errorf("%w: validator key mount path is empty", errs.ErrInvalidInput)
	}
	normalized := filepath.ToSlash(p)
	if !strings.HasPrefix(normalized, "./") && !strings.HasPrefix(normalized, "/") {
		return fmt.Errorf("%w: validator key mount path %q must be absolute or prefixed with ./ (compose named-volume hazard)", errs.ErrInvalidInput, p)
	}
	return nil
}
