package dockercli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type Runner interface {
	Run(ctx context.Context, bin string, args []string, opts RunOptions) (RunResult, error)
}

type RunOptions struct {
	Dir    string
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type RunResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

type execRunner struct{}

func NewExecRunner() Runner { return &execRunner{} }

func (r *execRunner) Run(ctx context.Context, bin string, args []string, opts RunOptions) (RunResult, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = opts.Dir
	if len(opts.Env) > 0 {
		cmd.Env = opts.Env
	}
	cmd.Stdin = opts.Stdin

	var outBuf, errBuf bytes.Buffer
	if opts.Stdout != nil {
		cmd.Stdout = io.MultiWriter(&outBuf, opts.Stdout)
	} else {
		cmd.Stdout = &outBuf
	}
	if opts.Stderr != nil {
		cmd.Stderr = io.MultiWriter(&errBuf, opts.Stderr)
	} else {
		cmd.Stderr = &errBuf
	}

	err := cmd.Run()
	res := RunResult{
		Stdout:   outBuf.Bytes(),
		Stderr:   errBuf.Bytes(),
		ExitCode: cmd.ProcessState.ExitCode(),
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
			return res, fmt.Errorf("%s %v exited with code %d: %s", bin, args, res.ExitCode, errBuf.String())
		}
		return res, fmt.Errorf("%s %v: %w", bin, args, err)
	}
	return res, nil
}
