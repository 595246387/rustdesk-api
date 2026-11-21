package admin

import (
	"regexp"
	"strings"
)

// 简单的邮箱格式校验，与服务端发信无关，只挡明显无效的输入。
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// NormalizeEmail 统一转小写并去掉首尾空白，便于去重与限流。
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidEmail 判断邮箱格式是否合法。
func ValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	return emailPattern.MatchString(email)
}
