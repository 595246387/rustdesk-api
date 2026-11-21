package utils

import (
	"sync"
	"time"
)

// RegisterLimiter 限制同一 IP 的自助注册频率，避免开放注册后被批量刷号。
//
// 只统计最近 24 小时的注册时间点：同一 IP 每小时最多 hourly 个、每天最多 daily 个。
type RegisterLimiter struct {
	mu      sync.Mutex
	records map[string][]time.Time
	hourly  int
	daily   int
}

func NewRegisterLimiter(hourly, daily int) *RegisterLimiter {
	if hourly <= 0 {
		hourly = 5
	}
	if daily <= 0 || daily < hourly {
		daily = 20
	}
	return &RegisterLimiter{
		records: make(map[string][]time.Time),
		hourly:  hourly,
		daily:   daily,
	}
}

// Check 只判断是否还允许注册，不占用名额。
// 配合 Record 使用，可以让校验失败（如验证码输错）不消耗配额。
func (l *RegisterLimiter) Check(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.checkLocked(ip, time.Now())
}

// Record 记录一次成功注册。
func (l *RegisterLimiter) Record(ip string) {
	if ip == "" {
		ip = "unknown"
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records[ip] = append(l.records[ip], time.Now())
	l.cleanupLocked(time.Now().Add(-24 * time.Hour))
}

// Allow 检查并记录一次注册，返回是否允许。
func (l *RegisterLimiter) Allow(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.checkLocked(ip, now) {
		return false
	}
	l.records[ip] = append(l.records[ip], now)
	l.cleanupLocked(now.Add(-24 * time.Hour))
	return true
}

func (l *RegisterLimiter) checkLocked(ip string, now time.Time) bool {
	dayAgo := now.Add(-24 * time.Hour)
	hourAgo := now.Add(-time.Hour)

	kept := make([]time.Time, 0, len(l.records[ip])+1)
	hourCount, dayCount := 0, 0
	for _, t := range l.records[ip] {
		if t.After(dayAgo) {
			kept = append(kept, t)
			dayCount++
			if t.After(hourAgo) {
				hourCount++
			}
		}
	}
	l.records[ip] = kept

	return dayCount < l.daily && hourCount < l.hourly
}

// cleanupLocked 清理长时间没有注册记录的 IP，避免 map 无限增长。
func (l *RegisterLimiter) cleanupLocked(cutoff time.Time) {
	if len(l.records) <= 1000 {
		return
	}
	for ip, list := range l.records {
		if len(list) == 0 || list[len(list)-1].Before(cutoff) {
			delete(l.records, ip)
		}
	}
}
