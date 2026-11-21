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

// Allow 记录一次注册尝试，并返回是否允许。
func (r *RegisterLimiter) Allow(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)
	hourAgo := now.Add(-time.Hour)

	r.mu.Lock()
	defer r.mu.Unlock()

	kept := make([]time.Time, 0, len(r.records[ip])+1)
	hourCount, dayCount := 0, 0
	for _, t := range r.records[ip] {
		if t.After(dayAgo) {
			kept = append(kept, t)
			dayCount++
			if t.After(hourAgo) {
				hourCount++
			}
		}
	}

	if dayCount >= r.daily || hourCount >= r.hourly {
		r.records[ip] = kept
		return false
	}

	r.records[ip] = append(kept, now)
	r.cleanupLocked(dayAgo)
	return true
}

// cleanupLocked 清理长时间没有注册记录的 IP，避免 map 无限增长。
func (r *RegisterLimiter) cleanupLocked(dayAgo time.Time) {
	if len(r.records) <= 1000 {
		return
	}
	for ip, list := range r.records {
		if len(list) == 0 || list[len(list)-1].Before(dayAgo) {
			delete(r.records, ip)
		}
	}
}
