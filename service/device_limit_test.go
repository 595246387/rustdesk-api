package service

import (
	"testing"
	"time"

	"github.com/lejianwen/rustdesk-api/v2/config"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDeviceLimitDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(
		sqlite.Open(t.TempDir()+"/device_limit.db"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserToken{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	oldDB, oldConfig := DB, Config
	DB, Config = db, &config.Config{}
	t.Cleanup(func() {
		DB, Config = oldDB, oldConfig
	})
	return db
}

func TestCheckDeviceLimit(t *testing.T) {
	db := setupDeviceLimitDB(t)

	limit := 2
	user := &model.User{Username: "device-limit-user", DeviceLimit: &limit}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	now := time.Now().Unix()
	addToken := func(uuid, deviceId string, expiredAt int64) {
		t.Helper()
		if err := db.Create(&model.UserToken{
			UserId:     user.Id,
			DeviceUuid: uuid,
			DeviceId:   deviceId,
			Token:      uuid + deviceId,
			ExpiredAt:  expiredAt,
		}).Error; err != nil {
			t.Fatalf("create token: %v", err)
		}
	}

	us := &UserService{}
	addToken("uuid-a", "1001", now+3600)
	addToken("uuid-b", "1002", now+3600)

	if err := us.CheckDeviceLimit(user, "uuid-a", "1001"); err != nil {
		t.Fatalf("existing device should pass: %v", err)
	}
	if err := us.CheckDeviceLimit(user, "uuid-c", "1003"); err == nil {
		t.Fatal("new device over limit should fail")
	}
	if err := us.CheckDeviceLimit(user, "", ""); err != nil {
		t.Fatalf("web login (no device) should pass: %v", err)
	}

	// 过期 token 不占额度
	addToken("uuid-old", "1000", now-60)
	if err := us.CheckDeviceLimit(user, "uuid-c", "1003"); err == nil {
		t.Fatal("expired token should not count, but limit is still reached")
	}

	// 用户级 -1 覆盖为不限制
	unlimited := -1
	user.DeviceLimit = &unlimited
	if err := us.CheckDeviceLimit(user, "uuid-c", "1003"); err != nil {
		t.Fatalf("user unlimited override should pass: %v", err)
	}

	// 用户级 0 表示暂未分配设备：拒绝新设备，已有 token 的设备仍可重连
	zero := 0
	user.DeviceLimit = &zero
	if err := us.CheckDeviceLimit(user, "uuid-c", "1003"); err == nil || err.Error() != "DeviceLimitNotAssigned" {
		t.Fatalf("device_limit 0 should block new device, got %v", err)
	}
	if err := us.CheckDeviceLimit(user, "uuid-a", "1001"); err != nil {
		t.Fatalf("existing device should still pass when device_limit is 0: %v", err)
	}

	// 未设置用户值时跟随全局：全局 2 时第 3 台被拦
	user.DeviceLimit = nil
	Config.App.DeviceLimit = 2
	if err := us.CheckDeviceLimit(user, "uuid-c", "1003"); err == nil {
		t.Fatal("global limit 2 should block the third device")
	}
	Config.App.DeviceLimit = 3
	if err := us.CheckDeviceLimit(user, "uuid-c", "1003"); err != nil {
		t.Fatalf("limit 3 should allow third device: %v", err)
	}

	// 全局 0 同样表示未分配设备
	Config.App.DeviceLimit = 0
	if err := us.CheckDeviceLimit(user, "uuid-d", "1004"); err == nil || err.Error() != "DeviceLimitNotAssigned" {
		t.Fatalf("global 0 should block new device, got %v", err)
	}
}

func TestCheckDeviceLimitFallsBackToDeviceId(t *testing.T) {
	db := setupDeviceLimitDB(t)

	limit := 1
	user := &model.User{Username: "id-only-user", DeviceLimit: &limit}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.UserToken{
		UserId:    user.Id,
		DeviceId:  "2001",
		Token:     "token-2001",
		ExpiredAt: time.Now().Unix() + 3600,
	}).Error; err != nil {
		t.Fatalf("create token: %v", err)
	}

	us := &UserService{}
	if err := us.CheckDeviceLimit(user, "", "2001"); err != nil {
		t.Fatalf("existing device_id should pass: %v", err)
	}
	if err := us.CheckDeviceLimit(user, "", "2002"); err == nil {
		t.Fatal("second device_id should exceed limit 1")
	}
}
