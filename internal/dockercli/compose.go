package dockercli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

type Service struct {
	Name        string
	ID          string
	Image       string
	State       string
	Health      string
	Ports       []PortMapping
	ExitCode    int
	ServiceName string
}

type PortMapping struct {
	Host      string
	Container string
	Protocol  string
}

type ComposeLogOpts struct {
	Follow bool
	Node   string
	Tail   int
	Stdout io.Writer
	Stderr io.Writer
}

const (
	defaultComposeFileYAML = "docker-compose.yaml"
	alternateComposeFile   = "docker-compose.yml"
)

func FindComposeFile(workDir string) (string, error) {
	for _, name := range []string{defaultComposeFileYAML, alternateComposeFile} {
		candidate := filepath.Join(workDir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%w: no docker-compose.yaml or docker-compose.yml in %s", errs.ErrComposeMissing, workDir)
}

func (c *Client) composeArgs(sub ...string) []string {
	args := []string{"compose"}
	args = append(args, sub...)
	return args
}

func (c *Client) ComposeUp(ctx context.Context) (RunResult, error) {
	return c.runDockerCompose(ctx, []string{"up", "-d"})
}

func (c *Client) ComposeDown(ctx context.Context) (RunResult, error) {
	return c.runDockerCompose(ctx, []string{"down"})
}

func (c *Client) ComposeRestart(ctx context.Context) (RunResult, error) {
	return c.runDockerCompose(ctx, []string{"restart"})
}

func (c *Client) runDockerCompose(ctx context.Context, sub []string) (RunResult, error) {
	if _, err := FindComposeFile(c.workDir); err != nil {
		return RunResult{}, err
	}
	res, err := c.run(ctx, c.composeArgs(sub...), os.Stdout, os.Stderr)
	if err != nil {
		return res, fmt.Errorf("%w: %w", errs.ErrDockerFailure, err)
	}
	return res, nil
}

type composePSJSON struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	Service  string `json:"Service"`
	Image    string `json:"Image"`
	State    string `json:"State"`
	Health   string `json:"Health"`
	ExitCode int    `json:"ExitCode"`
	Ports    string `json:"Publishers"`
	PortsStr string `json:"Ports"`
}

func (c *Client) ComposePS(ctx context.Context) ([]Service, error) {
	if _, err := FindComposeFile(c.workDir); err != nil {
		return nil, err
	}
	res, err := c.run(ctx, c.composeArgs("ps", "--all", "--format", "json"), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDockerFailure, err)
	}
	return parseComposePS(res.Stdout)
}

func parseComposePS(b []byte) ([]Service, error) {
	out := make([]Service, 0)
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) == 0 {
		return out, nil
	}

	if trimmed[0] == '[' {
		var arr []composePSJSON
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			return nil, fmt.Errorf("decode compose ps array: %w", err)
		}
		for _, row := range arr {
			out = append(out, rowToService(row))
		}
		return out, nil
	}

	scan := bufio.NewScanner(bytes.NewReader(trimmed))
	scan.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scan.Scan() {
		line := bytes.TrimSpace(scan.Bytes())
		if len(line) == 0 {
			continue
		}
		var row composePSJSON
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, fmt.Errorf("decode compose ps line: %w", err)
		}
		out = append(out, rowToService(row))
	}
	if err := scan.Err(); err != nil {
		return nil, fmt.Errorf("scan compose ps: %w", err)
	}
	return out, nil
}

func rowToService(r composePSJSON) Service {
	ports := parsePortsString(r.PortsStr)
	return Service{
		Name:        r.Name,
		ID:          r.ID,
		Image:       r.Image,
		State:       r.State,
		Health:      r.Health,
		Ports:       ports,
		ExitCode:    r.ExitCode,
		ServiceName: r.Service,
	}
}

func parsePortsString(s string) []PortMapping {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	entries := strings.Split(s, ", ")
	out := make([]PortMapping, 0, len(entries))
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		proto := "tcp"
		if idx := strings.LastIndex(e, "/"); idx != -1 {
			proto = e[idx+1:]
			e = e[:idx]
		}

		var leftPart, rightPart string
		if idx := strings.Index(e, "->"); idx != -1 {
			leftPart = e[:idx]
			rightPart = e[idx+2:]
		} else {
			rightPart = e
		}

		host := ""
		if leftPart != "" {
			if idx := strings.LastIndex(leftPart, ":"); idx != -1 {
				host = leftPart[idx+1:]
			} else {
				host = leftPart
			}
		}
		out = append(out, PortMapping{Host: host, Container: rightPart, Protocol: proto})
	}
	return out
}

func (c *Client) ComposeLogs(ctx context.Context, opts ComposeLogOpts) error {
	if _, err := FindComposeFile(c.workDir); err != nil {
		return err
	}
	args := c.composeArgs("logs")
	if opts.Follow {
		args = append(args, "-f")
	}
	if opts.Tail > 0 {
		args = append(args, "--tail", strconv.Itoa(opts.Tail))
	}
	if opts.Node != "" {
		args = append(args, opts.Node)
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	_, err := c.run(ctx, args, stdout, stderr)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDockerFailure, err)
	}
	return nil
}

func (c *Client) ComposeRestartService(ctx context.Context, service string) (RunResult, error) {
	return c.runDockerCompose(ctx, []string{"restart", service})
}
