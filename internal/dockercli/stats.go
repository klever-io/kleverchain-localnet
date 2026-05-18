package dockercli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

type Stat struct {
	Container string
	Name      string
	CPUPct    float64
	MemUsage  string
	MemPct    float64
	NetIO     string
	BlockIO   string
	PIDs      int
}

type statsRawJSON struct {
	Container string `json:"Container"`
	Name      string `json:"Name"`
	CPUPerc   string `json:"CPUPerc"`
	MemUsage  string `json:"MemUsage"`
	MemPerc   string `json:"MemPerc"`
	NetIO     string `json:"NetIO"`
	BlockIO   string `json:"BlockIO"`
	PIDs      string `json:"PIDs"`
}

func (c *Client) Stats(ctx context.Context) ([]Stat, error) {
	res, err := c.run(ctx, []string{"stats", "--no-stream", "--format", "{{json .}}"}, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDockerFailure, err)
	}
	return parseStats(res.Stdout)
}

func parseStats(b []byte) ([]Stat, error) {
	out := make([]Stat, 0)
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) == 0 {
		return out, nil
	}
	scan := bufio.NewScanner(bytes.NewReader(trimmed))
	scan.Buffer(make([]byte, 0, 512*1024), 512*1024)
	for scan.Scan() {
		line := bytes.TrimSpace(scan.Bytes())
		if len(line) == 0 {
			continue
		}
		var row statsRawJSON
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, fmt.Errorf("decode stats line: %w", err)
		}
		pids, _ := strconv.Atoi(strings.TrimSpace(row.PIDs))
		out = append(out, Stat{
			Container: row.Container,
			Name:      row.Name,
			CPUPct:    parsePct(row.CPUPerc),
			MemUsage:  row.MemUsage,
			MemPct:    parsePct(row.MemPerc),
			NetIO:     row.NetIO,
			BlockIO:   row.BlockIO,
			PIDs:      pids,
		})
	}
	if err := scan.Err(); err != nil {
		return nil, fmt.Errorf("scan stats: %w", err)
	}
	return out, nil
}

func parsePct(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
