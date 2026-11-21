package service

import (
	"testing"
	"time"

	"github.com/lejianwen/rustdesk-api/v2/model"
)

func TestUserExpire(t *testing.T) {
	setupDeviceLimitDB(t)

	now := time.Now().Unix()
	u := &model.User{
		Username: "expire-user",
		Status:   model.COMMON_STATUS_ENABLE,
		ExpireAt: now - 60,
	}
	if err := DB.Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	us := &UserService{}

	if !u.IsExpired() {
		t.Fatal("user with past expire_at should be expired")
	}
	if us.CheckUserEnable(u) {
		t.Fatal("expired user should not be enabled")
	}

	// 有效期在未来
	u.ExpireAt = now + 3600
	if u.IsExpired() {
		t.Fatal("user with future expire_at should not be expired")
	}
	if !us.CheckUserEnable(u) {
		t.Fatal("enabled user with future expire_at should pass")
	}

	// 0 表示永久有效
	u.ExpireAt = 0
	if u.IsExpired() {
		t.Fatal("expire_at 0 means forever")
	}
	if !us.CheckUserEnable(u) {
		t.Fatal("enabled user without expire_at should pass")
	}

	// 禁用仍然优先于有效期
	u.Status = model.COMMON_STATUS_DISABLED
	if us.CheckUserEnable(u) {
		t.Fatal("disabled user should not be enabled")
	}
}

func TestLoginTokenCappedByUserExpire(t *testing.T) {
	setupDeviceLimitDB(t)
	if err := DB.AutoMigrate(&model.LoginLog{}); err != nil {
		t.Fatalf("migrate login log: %v", err)
	}

	now := time.Now().Unix()
	u := &model.User{
		Username: "expire-token-user",
		Status:   model.COMMON_STATUS_ENABLE,
		ExpireAt: now + 600,
	}
	if err := DB.Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	us := &UserService{}
	ut := us.Login(u, &model.LoginLog{})
	if ut.ExpiredAt != u.ExpireAt {
		t.Fatalf("token expired_at = %d, want capped at %d", ut.ExpiredAt, u.ExpireAt)
	}

	// 没有有效期时使用默认 token 有效期
	u.ExpireAt = 0
	if got := us.capByUserExpire(u, now+604800); got != now+604800 {
		t.Fatalf("capByUserExpire without expire_at = %d, want %d", got, now+604800)
	}
}