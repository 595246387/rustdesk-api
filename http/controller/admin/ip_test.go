package admin

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRealClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name     string
		xff      string
		remote   string
		expected string
	}{
		{
			name:     "客户端伪造的地址在最左边，取右侧真实来源",
			xff:      "1.1.1.1, 203.0.113.9, 127.0.0.1",
			remote:   "192.0.2.10:1234",
			expected: "203.0.113.9",
		},
		{
			name:     "全是内网地址时回退到 gin 的 ClientIP",
			xff:      "10.0.0.1, 127.0.0.1",
			remote:   "192.0.2.10:1234",
			expected: "10.0.0.1",
		},
		{
			name:     "没有 XFF 时用 RemoteAddr",
			xff:      "",
			remote:   "198.51.100.7:5678",
			expected: "198.51.100.7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/", nil)
			c.Request.RemoteAddr = tc.remote
			if tc.xff != "" {
				c.Request.Header.Set("X-Forwarded-For", tc.xff)
			}
			if got := realClientIP(c); got != tc.expected {
				t.Fatalf("realClientIP() = %q, want %q", got, tc.expected)
			}
		})
	}
}
