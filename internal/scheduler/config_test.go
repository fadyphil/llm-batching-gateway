package scheduler

import (
	"testing"
	"time"
)

func TestFakeTimeSource_Advance_FiresSingleTimer(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fts := NewFakeTimeSource(baseTime)

	timer := fts.NewTimer(10 * time.Second)

	if n := fts.Advance(5 * time.Second); n != 0 {
		t.Fatalf("expected 0 timers to fire after +5s, got %d", n)
	}

	select {
	case <-timer.C():
		t.Fatal("timer fired too early at +5s, due at +10s")
	default:
	}

	if n := fts.Advance(5 * time.Second); n != 1 {
		t.Fatalf("expected 1 timer to fire after +10s, got %d", n)
	}

	select {
	case <-timer.C():
	default:
		t.Fatal("timer did not fire after due time")
	}
}

func TestFakeTimeSource_Stop_PreventsFire(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fts := NewFakeTimeSource(baseTime)

	timer := fts.NewTimer(10 * time.Second)
	timer.Stop()

	if n := fts.Advance(15 * time.Second); n != 0 {
		t.Fatalf("expected 0 timers to fire after Stop, got %d", n)
	}

	select {
	case <-timer.C():
		t.Fatal("timer fired after Stop")
	default:
	}
}

func TestFakeTimeSource_Now_ReturnsCurrent(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fts := NewFakeTimeSource(baseTime)

	if !fts.Now().Equal(baseTime) {
		t.Fatalf("expected Now() to return %v, got %v", baseTime, fts.Now())
	}

	fts.Advance(30 * time.Second)
	expected := baseTime.Add(30 * time.Second)
	if !fts.Now().Equal(expected) {
		t.Fatalf("expected Now() to return %v, got %v", expected, fts.Now())
	}
}
