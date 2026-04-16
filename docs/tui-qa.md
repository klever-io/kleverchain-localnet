# TUI Manual QA Checklist

Use this when cutting a release to verify the terminal UI works across the supported platforms.

## Platforms to cover

- [ ] Linux, `xterm-256color`, font with Powerline glyphs
- [ ] Linux, minimal `xterm`, `TERM=xterm` (no truecolor)
- [ ] macOS Terminal.app
- [ ] macOS iTerm2 (default profile + Solarized dark profile)
- [ ] Windows Terminal (PowerShell + Command Prompt profiles)
- [ ] WSL2 Ubuntu (via Windows Terminal)
- [ ] Minimum terminal size: **80×24**

## Bare invocation

- [ ] `localnet` on a TTY opens the splash within 50 ms.
- [ ] Splash animation completes within ~2 seconds.
- [ ] Any keypress during the splash jumps straight to the menu.
- [ ] `localnet --no-tui` prints `--help` and exits 0 (on the same TTY).
- [ ] `localnet | cat` prints help and exits 0 (stdin not a TTY).

## Menu screen

- [ ] State badge reflects reality:
  - Fresh dir → `○ Not initialized`
  - After `keys generate` → `◐ Keys generated`
  - After `setup-all` → `◑ Ready to start`
  - After `start` → `● Running N/M nodes healthy`
  - Partial keys directory → `⚠ Partial state — run setup-all`
- [ ] The "Suggested next" banner matches the state.
- [ ] Disabled items are visually dimmed and reject their hotkey with a warning status line.
- [ ] Arrow keys navigate; `enter` selects; hotkeys jump directly.
- [ ] `q` exits.

## Monitor screen

- [ ] `localnet monitor` skips splash+menu and opens directly.
- [ ] Table headers remain aligned at 80-col and at wider widths.
- [ ] Rows refresh every 2 s without the UI stuttering.
- [ ] Health color: green for healthy, yellow for starting, red for unhealthy/exited.
- [ ] `↑/↓` selects rows; the selection marker updates immediately.
- [ ] `q` returns to the menu and cancels in-flight refresh calls.

## Logs screen

- [ ] Logs stream live after pressing `[l]` from the menu.
- [ ] `g` jumps to the top; `G` to the bottom; `f` toggles follow.
- [ ] Buffer is bounded (scroll back should stop around 5,000 lines).
- [ ] `q` exits cleanly; no goroutine leak (verify via `localnet monitor` afterwards).

## Setup form

- [ ] Invalid validators (0 or 11) show an inline error.
- [ ] Invalid max-supply (0 or non-numeric) shows an inline error.
- [ ] Consensus group size > validators shows an inline error.
- [ ] `esc` cancels back to the menu.
- [ ] `enter` on the last field submits and runs `setup-all`, returning to the menu with the new state badge.

## Destructive actions

- [ ] `[d]` Stop shows a confirmation modal; `n` cancels, `y` runs.
- [ ] `[X]` Clean-all shows the modal and only proceeds on `y`.

## Accessibility / compatibility

- [ ] `NO_COLOR=1 localnet` renders with no ANSI colors.
- [ ] `TERM=dumb localnet` prints help instead of opening the TUI.
- [ ] Resizing the terminal re-layouts within a single frame.
- [ ] `ctrl+c` cleanly exits from every screen.
