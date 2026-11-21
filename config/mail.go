package config

// Mail 发信配置，当前用于注册邮箱验证码。
// 复用甲骨文 OCI Email Delivery（smtp.email.<region>.oci.oraclecloud.com:587 STARTTLS）。
type Mail struct {
	Host        string `mapstructure:"host"`         // 为空时邮件只写日志，便于本地调试
	Port        int    `mapstructure:"port"`         // 默认 587
	User        string `mapstructure:"user"`         // SMTP 用户名
	Pass        string `mapstructure:"pass"`         // SMTP 密码
	From        string `mapstructure:"from"`         // 发件人，必须与 OCI Approved Sender 一致
	ImplicitTLS bool   `mapstructure:"implicit-tls"` // 465 直连 TLS；587 用 STARTTLS
}

func (m *Mail) Init() {
	if m.Port == 0 {
		m.Port = 587
	}
	if m.From == "" {
		m.From = m.User
	}
	if m.Port == 465 {
		m.ImplicitTLS = true
	}
}

// Ready 发信通道是否可用。不可用时注册验证码会退化为写日志（仅限调试）。
func (m *Mail) Ready() bool {
	return m.Host != "" && m.From != ""
}
