package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterBlocksAfterMaxRequests(t *testing.T) {
	l := NewLimiter(true, 50*time.Millisecond, 2)
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("second allow: %v", err)
	}
	if err := l.Allow("u:1"); err == nil {
		t.Fatal("third allow should be rate limited")
	}
}

func TestLimiterResetsAfterWindow(t *testing.T) {
	l := NewLimiter(true, 20*time.Millisecond, 1)
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if err := l.Allow("u:1"); err == nil {
		t.Fatal("second allow should be blocked")
	}
	time.Sleep(25 * time.Millisecond)
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("allow after reset: %v", err)
	}
}

func TestLimiterEvictsExpiredKeys(t *testing.T) {
	l := NewLimiter(true, 5*time.Millisecond, 1)
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("allow u:1: %v", err)
	}
	if err := l.Allow("u:2"); err != nil {
		t.Fatalf("allow u:2: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := l.Allow("u:3"); err != nil {
		t.Fatalf("allow u:3: %v", err)
	}
	if len(l.state) != 1 {
		t.Fatalf("state size = %d, want 1", len(l.state))
	}
}
