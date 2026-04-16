package printer

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Spinner struct {
	out     io.Writer
	msg     string
	frames  []string
	active  bool
	done    chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	enabled bool
}

func (p *Printer) Spinner(msg string) *Spinner {
	enabled := !p.Quiet && isTTY(p.Out)
	return &Spinner{
		out:     p.Out,
		msg:     msg,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		done:    make(chan struct{}),
		enabled: enabled,
	}
}

func (s *Spinner) Start() {
	if !s.enabled {
		if s.out != nil {
			_, _ = fmt.Fprintln(s.out, "» "+s.msg)
		}
		return
	}
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()
	s.wg.Add(1)
	go s.loop()
}

func (s *Spinner) loop() {
	defer s.wg.Done()
	i := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			_, _ = fmt.Fprintf(s.out, "\r%s %s", s.frames[i%len(s.frames)], s.msg)
			i++
		}
	}
}

func (s *Spinner) Stop(finalMsg string) {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		if finalMsg != "" && s.out != nil {
			_, _ = fmt.Fprintln(s.out, finalMsg)
		}
		return
	}
	s.active = false
	s.mu.Unlock()
	close(s.done)
	s.wg.Wait()
	if !s.enabled {
		if finalMsg != "" && s.out != nil {
			_, _ = fmt.Fprintln(s.out, finalMsg)
		}
		return
	}
	_, _ = fmt.Fprint(s.out, "\r\033[K")
	if finalMsg != "" {
		_, _ = fmt.Fprintln(s.out, finalMsg)
	}
}

func DefaultWriter() io.Writer { return os.Stdout }
