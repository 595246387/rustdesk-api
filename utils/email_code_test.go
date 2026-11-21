package utils

import (
	"errors"
	"testing"
)

func TestEmailCodeIssueAndVerify(t *testing.T) {
	s := NewEmailCodeStore()
	code, err := s.Issue("a@example.com", "203.0.113.9")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("code = %q, want 6 digits", code)
	}
	if err := s.Verify("a@example.com", code); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	// 验证码只能使用一次
	if err := s.Verify("a@example.com", code); !errors.Is(err, ErrEmailCodeExpired) {
		t.Fatalf("second Verify() error = %v, want ErrEmailCodeExpired", err)
	}
}

func TestEmailCodeResendInterval(t *testing.T) {
	s := NewEmailCodeStore()
	if _, err := s.Issue("b@example.com", "203.0.113.9"); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if _, err := s.Issue("b@example.com", "203.0.113.9"); !errors.Is(err, ErrEmailCodeTooSoon) {
		t.Fatalf("second Issue() error = %v, want ErrEmailCodeTooSoon", err)
	}
	// Discard 后允许立即重发（用于邮件发送失败）
	s.Discard("b@example.com")
	if _, err := s.Issue("b@example.com", "203.0.113.9"); err != nil {
		t.Fatalf("Issue() after Discard error = %v", err)
	}
}

func TestEmailCodeTooManyAttempts(t *testing.T) {
	s := NewEmailCodeStore()
	code, err := s.Issue("c@example.com", "203.0.113.9")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}
	for i := 0; i < emailCodeMaxAttempts-1; i++ {
		if err := s.Verify("c@example.com", wrong); !errors.Is(err, ErrEmailCodeInvalid) {
			t.Fatalf("attempt %d error = %v, want ErrEmailCodeInvalid", i+1, err)
		}
	}
	if err := s.Verify("c@example.com", wrong); !errors.Is(err, ErrEmailCodeTooMany) {
		t.Fatalf("last attempt error = %v, want ErrEmailCodeTooMany", err)
	}
	// 达到上限后即使输入正确也拒绝，必须重新获取
	if err := s.Verify("c@example.com", code); !errors.Is(err, ErrEmailCodeTooMany) {
		t.Fatalf("correct code after lockout error = %v, want ErrEmailCodeTooMany", err)
	}
}

func TestRegisterLimiterCheckDoesNotConsume(t *testing.T) {
	l := NewRegisterLimiter(2, 5)
	if !l.Check("203.0.113.9") {
		t.Fatal("Check() should allow")
	}
	if !l.Check("203.0.113.9") {
		t.Fatal("Check() must not consume quota")
	}
	l.Record("203.0.113.9")
	l.Record("203.0.113.9")
	if l.Check("203.0.113.9") {
		t.Fatal("Check() should reject after reaching hourly limit")
	}
}
