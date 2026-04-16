package printer

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

type Printer struct {
	Out   io.Writer
	Err   io.Writer
	Quiet bool

	header  *color.Color
	success *color.Color
	info    *color.Color
	warn    *color.Color
	errc    *color.Color
	dim     *color.Color
}

func New(out, errOut io.Writer, noColor bool) *Printer {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	if noColor || !isTTY(out) || os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
	return &Printer{
		Out:     out,
		Err:     errOut,
		header:  color.New(color.FgHiMagenta, color.Bold),
		success: color.New(color.FgGreen, color.Bold),
		info:    color.New(color.FgCyan),
		warn:    color.New(color.FgYellow),
		errc:    color.New(color.FgRed, color.Bold),
		dim:     color.New(color.FgHiBlack),
	}
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

func (p *Printer) Header(title string) {
	if p.Quiet {
		return
	}
	bar := strings.Repeat("─", maxInt(3, len(title)+4))
	_, _ = fmt.Fprintln(p.Out)
	_, _ = p.header.Fprintln(p.Out, bar)
	_, _ = p.header.Fprintln(p.Out, "  "+title)
	_, _ = p.header.Fprintln(p.Out, bar)
}

func (p *Printer) Success(format string, args ...any) {
	if p.Quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	_, _ = p.success.Fprintln(p.Out, "✓ "+msg)
}

func (p *Printer) Info(format string, args ...any) {
	if p.Quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	_, _ = p.info.Fprintln(p.Out, "» "+msg)
}

func (p *Printer) Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	_, _ = p.warn.Fprintln(p.Err, "! "+msg)
}

func (p *Printer) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	_, _ = p.errc.Fprintln(p.Err, "✗ "+msg)
}

func (p *Printer) Plain(format string, args ...any) {
	if p.Quiet {
		return
	}
	_, _ = fmt.Fprintf(p.Out, format+"\n", args...)
}

func (p *Printer) Dim(format string, args ...any) {
	if p.Quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	_, _ = p.dim.Fprintln(p.Out, msg)
}

type Row []string

func (p *Printer) Table(headers Row, rows []Row) {
	if p.Quiet {
		return
	}
	if len(headers) == 0 && len(rows) == 0 {
		return
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, r := range rows {
		for i, cell := range r {
			if i >= len(widths) {
				break
			}
			if l := visibleLen(cell); l > widths[i] {
				widths[i] = l
			}
		}
	}
	padPrint := func(row Row, sep string) {
		parts := make([]string, 0, len(row))
		for i, cell := range row {
			if i >= len(widths) {
				break
			}
			pad := widths[i] - visibleLen(cell)
			if pad < 0 {
				pad = 0
			}
			parts = append(parts, cell+strings.Repeat(" ", pad))
		}
		_, _ = fmt.Fprintln(p.Out, strings.Join(parts, sep))
	}
	padPrint(headers, "   ")
	sepRow := make(Row, len(widths))
	for i, w := range widths {
		sepRow[i] = strings.Repeat("─", w)
	}
	padPrint(sepRow, "   ")
	for _, r := range rows {
		padPrint(r, "   ")
	}
}

func visibleLen(s string) int {
	out := 0
	inEsc := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case inEsc:
		default:
			out++
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
