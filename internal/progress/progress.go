package progress

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Option func(*Reporter)

func WithMinInterval(d time.Duration) Option {
	return func(r *Reporter) {
		r.minInterval = d
	}
}

type Reporter struct {
	out         io.Writer
	minInterval time.Duration

	mu    sync.Mutex
	last  time.Time
	count int
	total int
}

func New(out io.Writer, opts ...Option) *Reporter {
	r := &Reporter{
		out:         out,
		minInterval: 120 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Reporter) Tick(_ string) {
	if r == nil || r.out == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++
	now := time.Now()
	if r.count > 1 && now.Sub(r.last) < r.minInterval {
		return
	}
	r.last = now
	r.printLocked(r.count)
}

func (r *Reporter) Done() {
	if r == nil || r.out == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.count == 0 {
		return
	}
	r.printLocked(r.count)
	fmt.Fprint(r.out, "\n")
}

func (r *Reporter) SetTotal(total int) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if total < 0 {
		total = 0
	}
	r.total = total
}

func (r *Reporter) printLocked(count int) {
	if r.total > 0 {
		percent := count * 100 / r.total
		if percent > 100 {
			percent = 100
		}
		fmt.Fprintf(r.out, "\rtranslating... %d/%d (%d%%)", count, r.total, percent)
		return
	}
	fmt.Fprintf(r.out, "\rtranslating... %d", count)
}
