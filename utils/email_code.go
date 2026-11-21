package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// 注册邮箱验证码策略
const (
	EmailCodeTTL          = 15 * time.Minute
	emailCodeMaxAttempts  = 5
	EmailCodeResendAfter  = 60 * time.Second
	emailCodePerEmailHour = 5
	emailCodePerIPHour    = 10
)

var (
	ErrEmailCodeExpired  = errors.New("EmailCodeExpired")
	ErrEmailCodeInvalid  = errors.New("EmailCodeInvalid")
	ErrEmailCodeTooMany  = errors.New("EmailCodeTooManyAttempts")
	ErrEmailCodeTooSoon  = errors.New("EmailCodeResendTooSoon")
	ErrEmailCodeTooOften = errors.New("RegisterTooFrequent")
)

type emailCodeEntry struct {
	Code      string
	ExpiresAt time.Time
	Attempts  int
}

// EmailCodeStore 保存注册邮箱验证码与发送频率记录。
//
// 单实例内存存储：验证码 15 分钟有效、每个码最多尝试 5 次；
// 同一邮箱 60 秒内不可重发、每小时最多 5 封；同一 IP 每小时最多 10 封。
type EmailCodeStore struct {
	mu         sync.Mutex
	codes      map[string]*emailCodeEntry
	emailSends map[string][]time.Time
	ipSends    map[string][]time.Time
}

func NewEmailCodeStore() *EmailCodeStore {
	return &EmailCodeStore{
		codes:      make(map[string]*emailCodeEntry),
		emailSends: make(map[string][]time.Time),
		ipSends:    make(map[string][]time.Time),
	}
}

func newEmailCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// Issue 生成并暂存一个验证码，同时记录发送频率。
// 调用方发送邮件失败时应调用 Discard，让用户能立即重试。
func (s *EmailCodeStore) Issue(email, ip string) (string, error) {
	code, err := newEmailCode()
	if err != nil {
		return "", err
	}
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)

	if last := s.lastSendLocked(s.emailSends, email); !last.IsZero() && now.Sub(last) < EmailCodeResendAfter {
		return "", ErrEmailCodeTooSoon
	}
	if len(s.recentLocked(s.emailSends, email, now)) >= emailCodePerEmailHour {
		return "", ErrEmailCodeTooOften
	}
	if ip != "" && len(s.recentLocked(s.ipSends, ip, now)) >= emailCodePerIPHour {
		return "", ErrEmailCodeTooOften
	}

	s.codes[email] = &emailCodeEntry{Code: code, ExpiresAt: now.Add(EmailCodeTTL)}
	s.emailSends[email] = append(s.emailSends[email], now)
	if ip != "" {
		s.ipSends[ip] = append(s.ipSends[ip], now)
	}
	return code, nil
}

// Verify 校验验证码。成功后验证码立即失效。
func (s *EmailCodeStore) Verify(email, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry := s.codes[email]
	if entry == nil || now.After(entry.ExpiresAt) {
		delete(s.codes, email)
		return ErrEmailCodeExpired
	}
	if entry.Attempts >= emailCodeMaxAttempts {
		return ErrEmailCodeTooMany
	}
	if subtle.ConstantTimeCompare([]byte(entry.Code), []byte(code)) != 1 {
		entry.Attempts++
		if entry.Attempts >= emailCodeMaxAttempts {
			return ErrEmailCodeTooMany
		}
		return ErrEmailCodeInvalid
	}
	delete(s.codes, email)
	return nil
}

// Discard 撤销刚发出的验证码，并允许同一邮箱立即重发（邮件发送失败时使用）。
func (s *EmailCodeStore) Discard(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.codes, email)
	if list := s.emailSends[email]; len(list) > 0 {
		s.emailSends[email] = list[:len(list)-1]
	}
}

// ResendAfter 返回同一邮箱还需等待多久才能重发，0 表示可以立即重发。
func (s *EmailCodeStore) ResendAfter(email string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	last := s.lastSendLocked(s.emailSends, email)
	if last.IsZero() {
		return 0
	}
	if wait := EmailCodeResendAfter - time.Since(last); wait > 0 {
		return wait
	}
	return 0
}

func (s *EmailCodeStore) lastSendLocked(m map[string][]time.Time, key string) time.Time {
	list := m[key]
	if len(list) == 0 {
		return time.Time{}
	}
	return list[len(list)-1]
}

// recentLocked 返回一小时内有效的记录，并顺带清理过期项。
func (s *EmailCodeStore) recentLocked(m map[string][]time.Time, key string, now time.Time) []time.Time {
	cutoff := now.Add(-time.Hour)
	list := m[key]
	kept := list[:0]
	for _, t := range list {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(m, key)
	} else {
		m[key] = kept
	}
	return kept
}

func (s *EmailCodeStore) cleanupLocked(now time.Time) {
	for email, entry := range s.codes {
		if now.After(entry.ExpiresAt) {
			delete(s.codes, email)
		}
	}
	if len(s.emailSends) > 2000 {
		for k := range s.emailSends {
			s.recentLocked(s.emailSends, k, now)
		}
	}
	if len(s.ipSends) > 2000 {
		for k := range s.ipSends {
			s.recentLocked(s.ipSends, k, now)
		}
	}
}
