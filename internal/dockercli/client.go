package dockercli

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

type Client struct {
	runner  Runner
	workDir string
	binary  string
}

type Option func(*Client)

func WithRunner(r Runner) Option  { return func(c *Client) { c.runner = r } }
func WithWorkDir(p string) Option { return func(c *Client) { c.workDir = p } }
func WithDockerBin(b string) Option {
	return func(c *Client) {
		if b != "" {
			c.binary = b
		}
	}
}

func New(opts ...Option) *Client {
	c := &Client{
		runner: NewExecRunner(),
		binary: "docker",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) run(ctx context.Context, args []string, stdout, stderr io.Writer) (RunResult, error) {
	return c.runner.Run(ctx, c.binary, args, RunOptions{
		Dir:    c.workDir,
		Stdout: stdout,
		Stderr: stderr,
	})
}

type Version struct {
	Client string
	Major  int
	Minor  int
	Patch  int
}

var versionRe = regexp.MustCompile(`(?m)Docker version (\d+)\.(\d+)\.(\d+)`)

func (c *Client) Version(ctx context.Context) (Version, error) {
	res, err := c.run(ctx, []string{"--version"}, nil, nil)
	if err != nil {
		return Version{}, fmt.Errorf("%w: %w", errs.ErrPrereq, err)
	}
	out := string(res.Stdout)
	m := versionRe.FindStringSubmatch(out)
	if len(m) != 4 {
		return Version{}, fmt.Errorf("%w: unexpected docker --version output: %q", errs.ErrPrereq, strings.TrimSpace(out))
	}
	maj, _ := strconv.Atoi(m[1])
	mnr, _ := strconv.Atoi(m[2])
	pat, _ := strconv.Atoi(m[3])
	return Version{Client: strings.TrimSpace(out), Major: maj, Minor: mnr, Patch: pat}, nil
}

func (c *Client) DaemonReachable(ctx context.Context) error {
	_, err := c.run(ctx, []string{"info"}, io.Discard, io.Discard)
	if err != nil {
		return fmt.Errorf("%w: docker daemon not reachable: %w", errs.ErrPrereq, err)
	}
	return nil
}

type ComposeInfo struct {
	Raw   string
	Major int
	Minor int
	Patch int
}

var composeVersionRe = regexp.MustCompile(`(?m)v?(\d+)\.(\d+)\.(\d+)`)

func (c *Client) ComposeVersion(ctx context.Context) (ComposeInfo, error) {
	res, err := c.run(ctx, []string{"compose", "version"}, nil, nil)
	if err != nil {
		return ComposeInfo{}, fmt.Errorf("%w: docker compose v2 not available: %w", errs.ErrPrereq, err)
	}
	out := string(res.Stdout)
	m := composeVersionRe.FindStringSubmatch(out)
	if len(m) != 4 {
		return ComposeInfo{Raw: strings.TrimSpace(out)}, fmt.Errorf("%w: cannot parse compose version: %q", errs.ErrPrereq, out)
	}
	maj, _ := strconv.Atoi(m[1])
	mnr, _ := strconv.Atoi(m[2])
	pat, _ := strconv.Atoi(m[3])
	return ComposeInfo{Raw: strings.TrimSpace(out), Major: maj, Minor: mnr, Patch: pat}, nil
}

func (c *Client) ImageExists(ctx context.Context, image string) bool {
	_, err := c.run(ctx, []string{"image", "inspect", image}, io.Discard, io.Discard)
	return err == nil
}

func (c *Client) Run(ctx context.Context, args []string) (RunResult, error) {
	res, err := c.run(ctx, args, nil, nil)
	if err != nil {
		return res, fmt.Errorf("%w: %w", errs.ErrDockerFailure, err)
	}
	return res, nil
}
