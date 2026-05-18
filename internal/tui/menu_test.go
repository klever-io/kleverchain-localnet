package tui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/detect"
)

func TestMenu_BadgeEachState(t *testing.T) {
	cases := []struct {
		name string
		snap detect.Snapshot
		want string
	}{
		{"uninitialized", detect.Snapshot{State: detect.Uninitialized}, "Not initialized"},
		{"keys-only", detect.Snapshot{State: detect.KeysOnly}, "Keys generated"},
		{"ready", detect.Snapshot{State: detect.Ready}, "Ready to start"},
		{"running", detect.Snapshot{State: detect.Running, HealthyServices: 3, RunningServices: 3}, "3/3 nodes healthy"},
		{"partial", detect.Snapshot{State: detect.Partial}, "Partial state"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			theme := NewTheme()
			out := renderBadge(theme, c.snap)
			require.True(t, strings.Contains(out, c.want), "badge %q should contain %q", out, c.want)
		})
	}
}

func TestMenu_NextStepBannerEachState(t *testing.T) {
	theme := NewTheme()
	cases := []struct {
		state detect.State
		want  string
	}{
		{detect.Uninitialized, "Press [s]"},
		{detect.KeysOnly, "Press [s]"},
		{detect.Ready, "Press [u]"},
		{detect.Running, "Press [m]"},
		{detect.Partial, "Press [X]"},
	}
	for _, c := range cases {
		got := renderNextStepBanner(theme, detect.Snapshot{State: c.state})
		require.True(t, strings.Contains(got, c.want), "state %v banner %q should contain %q", c.state, got, c.want)
	}
}

func TestMenu_DisabledItems(t *testing.T) {
	theme := NewTheme()
	keymap := DefaultKeyMap()

	m := newMenu(theme, keymap, detect.Snapshot{State: detect.Uninitialized})
	startItem := menuItem{action: actionStart, enabledFor: map[detect.State]bool{detect.Ready: true}}
	require.False(t, startItem.enabled(m.snap.State))

	m = m.updateSnapshot(detect.Snapshot{State: detect.Ready})
	require.True(t, startItem.enabled(m.snap.State))
}

func TestSplash_Completes(t *testing.T) {
	theme := NewTheme()
	s := newSplash(theme)
	require.Greater(t, s.totalChars, 0)
	for i := 0; i < 500 && !s.done; i++ {
		s, _ = s.Update(splashTickMsg{})
	}
	require.True(t, s.done, "splash must complete in bounded ticks")
}

func TestSetupForm_ValidationErrors(t *testing.T) {
	theme := NewTheme()
	m := newSetupForm(theme)

	m.inputs[0].SetValue("0")
	m, _ = m.submit()
	require.NotEmpty(t, m.err)

	m.inputs[0].SetValue("3")
	m.inputs[1].SetValue("0")
	m, _ = m.submit()
	require.NotEmpty(t, m.err)

	m.inputs[1].SetValue("10000000000000000")
	m.inputs[2].SetValue("10")
	m, _ = m.submit()
	require.NotEmpty(t, m.err)

	m.inputs[2].SetValue("2")
	m, _ = m.submit()
	require.Empty(t, m.err)
	require.NotNil(t, m.result)
	require.Equal(t, 3, m.result.Validators)
	require.Equal(t, 2, m.result.ConsensusGroupSize)
}
