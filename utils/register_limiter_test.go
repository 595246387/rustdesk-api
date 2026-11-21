package utils

import "testing"

func TestRegisterLimiterHourly(t *testing.T) {
	l := NewRegisterLimiter(2, 5)
	if !l.Allow("203.0.113.9") {
		t.Fatal("first register should be allowed")
	}
	if !l.Allow("203.0.113.9") {
		t.Fatal("second register should be allowed")
	}
	if l.Allow("203.0.113.9") {
		t.Fatal("third register within an hour should be rejected")
	}
	// 其他 IP 不受影响
	if !l.Allow("203.0.113.10") {
		t.Fatal("other ip should not be limited")
	}
}
