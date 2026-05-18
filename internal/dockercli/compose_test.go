package dockercli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseComposePS_JSONL(t *testing.T) {
	input := []byte(`{"ID":"abc","Name":"node0","Service":"node0","Image":"x","State":"running","Health":"healthy","ExitCode":0,"Ports":"0.0.0.0:8800->8800/tcp"}
{"ID":"def","Name":"seednode","Service":"seednode","Image":"x","State":"running","Health":"","ExitCode":0,"Ports":""}
`)
	svc, err := parseComposePS(input)
	require.NoError(t, err)
	require.Len(t, svc, 2)
	require.Equal(t, "node0", svc[0].Name)
	require.Equal(t, "healthy", svc[0].Health)
	require.Len(t, svc[0].Ports, 1)
	require.Equal(t, "8800", svc[0].Ports[0].Host)
	require.Equal(t, "8800", svc[0].Ports[0].Container)
}

func TestParseComposePS_Array(t *testing.T) {
	input := []byte(`[{"ID":"abc","Name":"node0","Service":"node0","Image":"x","State":"running","Health":"healthy","ExitCode":0,"Ports":""}]`)
	svc, err := parseComposePS(input)
	require.NoError(t, err)
	require.Len(t, svc, 1)
	require.Equal(t, "running", svc[0].State)
}

func TestParseComposePS_Empty(t *testing.T) {
	svc, err := parseComposePS([]byte(""))
	require.NoError(t, err)
	require.Empty(t, svc)
}

func TestParsePortsString(t *testing.T) {
	p := parsePortsString("0.0.0.0:8800->8800/tcp, :::8800->8800/tcp")
	require.Len(t, p, 2)
	require.Equal(t, "tcp", p[0].Protocol)
	require.Equal(t, "8800", p[0].Host)
	require.Equal(t, "8800", p[0].Container)
	require.Equal(t, "8800", p[1].Container)
}

func TestParseStats(t *testing.T) {
	input := []byte(`{"Container":"abc","Name":"node0","CPUPerc":"8.4%","MemUsage":"512MiB / 2GiB","MemPerc":"25.0%","NetIO":"1kB / 2kB","BlockIO":"0B / 0B","PIDs":"12"}
{"Container":"def","Name":"seednode","CPUPerc":"1.2%","MemUsage":"142MiB / 2GiB","MemPerc":"7.1%","NetIO":"1kB / 2kB","BlockIO":"0B / 0B","PIDs":"8"}
`)
	stats, err := parseStats(input)
	require.NoError(t, err)
	require.Len(t, stats, 2)
	require.InDelta(t, 8.4, stats[0].CPUPct, 0.001)
	require.Equal(t, 12, stats[0].PIDs)
}
