package admin

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// realClientIP 返回用于限流的客户端 IP。
//
// 请求链路是：客户端 -> 外层 nginx -> 本机 9443 -> API。
// 每一层都会把上一跳地址追加到 X-Forwarded-For 末尾，而客户端可以伪造
// 前面的内容，所以从右往左取第一个公网地址，避免用伪造值绕过限流。
func realClientIP(c *gin.Context) string {
	xff := c.GetHeader("X-Forwarded-For")
	if xff == "" {
		return c.ClientIP()
	}
	parts := strings.Split(xff, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		ipStr := strings.TrimSpace(parts[i])
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
			continue
		}
		return ipStr
	}
	return c.ClientIP()
}
