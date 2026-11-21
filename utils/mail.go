package utils

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/lejianwen/rustdesk-api/v2/config"
)

const smtpTimeout = 15 * time.Second

// SendMail 通过 SMTP 发送一封 UTF-8 纯文本邮件。
// 支持 587 STARTTLS 和 465 直连 TLS，认证信息来自配置。
func SendMail(cfg *config.Mail, to, subject, body string) error {
	if cfg == nil || !cfg.Ready() {
		return fmt.Errorf("SMTP 未配置")
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	msg := buildMessage(cfg.From, to, subject, body)

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var conn net.Conn
	var err error
	if cfg.ImplicitTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: cfg.Host})
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("连接 SMTP %s: %w", addr, err)
	}
	_ = conn.SetDeadline(time.Now().Add(smtpTimeout))

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("SMTP 握手: %w", err)
	}
	defer c.Close()

	if !cfg.ImplicitTLS {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP 服务器未提供 STARTTLS")
		}
		if err := c.StartTLS(&tls.Config{ServerName: cfg.Host}); err != nil {
			return fmt.Errorf("STARTTLS: %w", err)
		}
	}

	if cfg.User != "" {
		if ok, _ := c.Extension("AUTH"); !ok {
			return fmt.Errorf("SMTP 服务器未提供认证")
		}
		auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP 认证: %w", err)
		}
	}

	if err := c.Mail(envelopeAddr(cfg.From)); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("写入邮件正文: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("结束邮件正文: %w", err)
	}
	return c.Quit()
}

// buildMessage 生成一封 UTF-8 纯文本邮件，正文用 base64 避免折行破坏验证码。
func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.BEncoding.Encode("UTF-8", subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n")
	b.WriteString("\r\n")
	b.WriteString(chunkBase64(body))
	b.WriteString("\r\n")
	return []byte(b.String())
}

// chunkBase64 按 76 字符折行，符合 RFC 2045。
func chunkBase64(s string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(s))
	var b strings.Builder
	for len(enc) > 76 {
		b.WriteString(enc[:76])
		b.WriteString("\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc)
	return b.String()
}

// envelopeAddr 从 "名字 <addr>" 中取出信封地址。
func envelopeAddr(from string) string {
	if i := strings.LastIndex(from, "<"); i >= 0 {
		if j := strings.Index(from[i:], ">"); j > 0 {
			return strings.TrimSpace(from[i+1 : i+j])
		}
	}
	return strings.TrimSpace(from)
}
