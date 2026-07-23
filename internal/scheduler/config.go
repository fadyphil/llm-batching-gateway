package scheduler

import "time"

type Config struct {
	IngressCapacity int
	MaxBatchSize    int
	BatchWindow     time.Duration
	SessionTTL      time.Duration
	HeartbeatLimit  time.Duration
}

type Timer interface {
	C() <-chan time.Time
	Stop() bool
}

type TimeSource interface {
	Now() time.Time
	NewTimer(d time.Duration) Timer
}

type realTimer struct {
	t *time.Timer
}

func (r *realTimer) C() <-chan time.Time { return r.t.C }
func (r *realTimer) Stop() bool          { return r.t.Stop() }

type RealTimeSource struct{}

func (RealTimeSource) Now() time.Time                        { return time.Now() }
func (RealTimeSource) NewTimer(d time.Duration) Timer        { return &realTimer{t: time.NewTimer(d)} }

type FakeTimeSource struct {
	now    time.Time
	timers []*fakeTimer
}

type fakeTimer struct {
	c      chan time.Time
	dueAt  time.Time
	active bool
}

func NewFakeTimeSource(start time.Time) *FakeTimeSource {
	return &FakeTimeSource{now: start}
}

func (f *FakeTimeSource) Now() time.Time { return f.now }

func (f *FakeTimeSource) NewTimer(d time.Duration) Timer {
	ft := &fakeTimer{
		c:      make(chan time.Time, 1),
		dueAt:  f.now.Add(d),
		active: true,
	}
	f.timers = append(f.timers, ft)
	return ft
}

func (ft *fakeTimer) C() <-chan time.Time { return ft.c }
func (ft *fakeTimer) Stop() bool {
	was := ft.active
	ft.active = false
	return was
}

// Advance moves the clock forward by d, firing all timers whose dueAt <= new time.
// Timers fire in order. Returns the number of timers that fired.
func (f *FakeTimeSource) Advance(d time.Duration) int {
	f.now = f.now.Add(d)
	fired := 0
	for _, ft := range f.timers {
		if ft.active && !ft.dueAt.After(f.now) {
			ft.active = false
			ft.c <- f.now
			fired++
		}
	}
	return fired
}
