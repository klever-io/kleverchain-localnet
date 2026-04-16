package dockercli_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
)

type fakeCall struct {
	Bin  string
	Args []string
	Dir  string
}

type fakeResponse struct {
	Stdout string
	Stderr string
	Exit   int
	Err    error
}

type fakeRunner struct {
	calls     []fakeCall
	responses map[string]fakeResponse
	defaultRe fakeResponse
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{responses: make(map[string]fakeResponse)}
}

func (f *fakeRunner) expect(key string, resp fakeResponse) {
	f.responses[key] = resp
}

func (f *fakeRunner) Run(_ context.Context, bin string, args []string, opts dockercli.RunOptions) (dockercli.RunResult, error) {
	f.calls = append(f.calls, fakeCall{Bin: bin, Args: append([]string(nil), args...), Dir: opts.Dir})
	key := strings.Join(append([]string{bin}, args...), " ")
	resp, ok := f.responses[key]
	if !ok {
		resp = f.defaultRe
	}
	res := dockercli.RunResult{
		Stdout:   []byte(resp.Stdout),
		Stderr:   []byte(resp.Stderr),
		ExitCode: resp.Exit,
	}
	if opts.Stdout != nil && len(resp.Stdout) > 0 {
		_, _ = opts.Stdout.Write([]byte(resp.Stdout))
	}
	if opts.Stderr != nil && len(resp.Stderr) > 0 {
		_, _ = opts.Stderr.Write([]byte(resp.Stderr))
	}
	if resp.Err != nil {
		return res, resp.Err
	}
	if resp.Exit != 0 {
		return res, fmt.Errorf("exit %d", resp.Exit)
	}
	return res, nil
}
