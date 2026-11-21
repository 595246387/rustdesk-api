package model

import "time"

type User struct {
	IdModel
	Username string `json:"username" gorm:"default:'';not null;uniqueIndex"`
	Email    string `json:"email" gorm:"default:'';not null;index"`
	// Email	string     	`json:"email" `
	Password string     `json:"-" gorm:"default:'';not null;"`
	Nickname string     `json:"nickname" gorm:"default:'';not null;"`
	Avatar   string     `json:"avatar" gorm:"default:'';not null;"`
	GroupId  uint       `json:"group_id" gorm:"default:0;not null;index"`
	IsAdmin  *bool      `json:"is_admin" gorm:"default:0;not null;"`
	Status   StatusCode `json:"status" gorm:"default:1;not null;"`
	Remark   string     `json:"remark" gorm:"default:'';not null;"`
	// DeviceLimit 每用户设备数上限。nil 跟随全局 app.device-limit，0 暂未分配设备（拒绝新设备登录），
	// <0 不限制，>0 为该用户独立上限。用指针是为了让管理端能区分「没提交这个字段」和「显式设为 0」。
	DeviceLimit *int `json:"device_limit" gorm:"default:0"`
	// ExpireAt 账号有效期（Unix 秒）。0 表示永久有效；到期后无法再登录客户端或后台，
	// 已签发的 token 也会被拒绝。管理员改大或清空即可恢复。
	ExpireAt int64 `json:"expire_at" gorm:"default:0;not null;"`
	TimeModel
}

// IsExpired 账号是否已过期（ExpireAt 为 0 表示永久有效）。
func (u *User) IsExpired() bool {
	return u != nil && u.ExpireAt > 0 && u.ExpireAt <= time.Now().Unix()
}

// BeforeSave 钩子用于确保 email 字段有合理的默认值
//func (u *User) BeforeSave(tx *gorm.DB) (err error) {
//	// 如果 email 为空，设置为默认值
//	if u.Email == "" {
//		u.Email = fmt.Sprintf("%s@example.com", u.Username)
//	}
//	return nil
//}

type UserList struct {
	Users []*User `json:"list,omitempty"`
	Pagination
}

var UserRouteNames = []string{
	"MyTagList", "MyAddressBookList", "MyInfo", "MyAddressBookCollection", "MyPeer", "MyShareRecordList", "MyLoginLog",
}
var AdminRouteNames = []string{"*"}
