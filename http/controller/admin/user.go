package admin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	adResp "github.com/lejianwen/rustdesk-api/v2/http/response/admin"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"github.com/lejianwen/rustdesk-api/v2/utils"
	"gorm.io/gorm"
)

type User struct {
}

// Detail 管理员
// @Tags 用户
// @Summary 管理员详情
// @Description 管理员详情
// @Accept  json
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} response.Response{data=model.User}
// @Failure 500 {object} response.Response
// @Router /admin/user/detail/{id} [get]
// @Security token
func (ct *User) Detail(c *gin.Context) {
	id := c.Param("id")
	iid, _ := strconv.Atoi(id)
	u := service.AllService.UserService.InfoById(uint(iid))
	if u.Id > 0 {
		response.Success(c, u)
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
	return
}

// Create 管理员
// @Tags 用户
// @Summary 创建管理员
// @Description 创建管理员
// @Accept  json
// @Produce  json
// @Param body body admin.UserForm true "管理员信息"
// @Success 200 {object} response.Response{data=model.User}
// @Failure 500 {object} response.Response
// @Router /admin/user/create [post]
// @Security token
func (ct *User) Create(c *gin.Context) {
	f := &admin.UserForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := f.ToUser()
	err := service.AllService.UserService.Create(u)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// List 列表
// @Tags 用户
// @Summary 管理员列表
// @Description 管理员列表
// @Accept  json
// @Produce  json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Param username query int false "账户"
// @Success 200 {object} response.Response{data=model.UserList}
// @Failure 500 {object} response.Response
// @Router /admin/user/list [get]
// @Security token
func (ct *User) List(c *gin.Context) {
	query := &admin.UserQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.UserService.List(query.Page, query.PageSize, func(tx *gorm.DB) {
		if query.Username != "" {
			tx.Where("username like ?", "%"+query.Username+"%")
		}
	})
	response.Success(c, res)
}

// Update 编辑
// @Tags 用户
// @Summary 管理员编辑
// @Description 管理员编辑
// @Accept  json
// @Produce  json
// @Param body body admin.UserForm true "用户信息"
// @Success 200 {object} response.Response{data=model.User}
// @Failure 500 {object} response.Response
// @Router /admin/user/update [post]
// @Security token
func (ct *User) Update(c *gin.Context) {
	f := &admin.UserForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if f.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := f.ToUser()
	err := service.AllService.UserService.Update(u)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// Delete 删除
// @Tags 用户
// @Summary 管理员删除
// @Description 管理员编删除
// @Accept  json
// @Produce  json
// @Param body body admin.UserForm true "用户信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user/delete [post]
// @Security token
func (ct *User) Delete(c *gin.Context) {
	f := &admin.UserForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	id := f.Id
	errList := global.Validator.ValidVar(c, id, "required,gt=0")
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := service.AllService.UserService.InfoById(f.Id)
	if u.Id > 0 {
		err := service.AllService.UserService.Delete(u)
		if err == nil {
			response.Success(c, nil)
			return
		}
		response.Fail(c, 101, err.Error())
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
}

// UpdatePassword 修改密码
// @Tags 用户
// @Summary 修改密码
// @Description 修改密码
// @Accept  json
// @Produce  json
// @Param body body admin.UserPasswordForm true "用户信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user/updatePassword [post]
// @Security token
func (ct *User) UpdatePassword(c *gin.Context) {
	f := &admin.UserPasswordForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := service.AllService.UserService.InfoById(f.Id)
	if u.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
		return
	}
	err := service.AllService.UserService.UpdatePassword(u, f.Password)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// Current 当前用户
// @Tags 用户
// @Summary 当前用户
// @Description 当前用户
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response{data=adResp.LoginPayload}
// @Failure 500 {object} response.Response
// @Router /admin/user/current [get]
// @Security token
func (ct *User) Current(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	token, _ := c.Get("token")
	t := token.(string)
	responseLoginSuccess(c, u, t)
}

// ChangeCurPwd 修改当前用户密码
// @Tags 用户
// @Summary 修改当前用户密码
// @Description 修改当前用户密码
// @Accept  json
// @Produce  json
// @Param body body admin.ChangeCurPasswordForm true "用户信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user/changeCurPwd [post]
// @Security token
func (ct *User) ChangeCurPwd(c *gin.Context) {
	f := &admin.ChangeCurPasswordForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}

	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := service.AllService.UserService.CurUser(c)
	// Verify the old password only when the account already has one set
	if !service.AllService.UserService.IsPasswordEmptyByUser(u) {
		ok, _, err := utils.VerifyPassword(u.Password, f.OldPassword)
		if err != nil || !ok {
			response.Fail(c, 101, response.TranslateMsg(c, "OldPasswordError"))
			return
		}
	}
	err := service.AllService.UserService.UpdatePassword(u, f.NewPassword)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// MyOauth
// @Tags 用户
// @Summary 我的授权
// @Description 我的授权
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response{data=[]adResp.UserOauthItem}
// @Failure 500 {object} response.Response
// @Router /admin/user/myOauth [get]
// @Security token
func (ct *User) MyOauth(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	oal := service.AllService.OauthService.List(1, 100, nil)
	ops := make([]string, 0)
	for _, oa := range oal.Oauths {
		ops = append(ops, oa.Op)
	}
	uts := service.AllService.UserService.UserThirdsByUserId(u.Id)
	var res []*adResp.UserOauthItem
	for _, oa := range oal.Oauths {
		item := &adResp.UserOauthItem{
			Op: oa.Op,
		}
		for _, ut := range uts {
			if ut.Op == oa.Op {
				item.Status = 1
				break
			}
		}
		res = append(res, item)
	}
	response.Success(c, res)
}

// groupUsers
func (ct *User) GroupUsers(c *gin.Context) {
	aG := service.AllService.GroupService.List(1, 999, nil)
	aU := service.AllService.UserService.List(1, 9999, nil)
	response.Success(c, gin.H{
		"groups": aG.Groups,
		"users":  aU.Users,
	})
}

// SendRegisterCode 发送注册邮箱验证码
// @Tags 用户
// @Summary 发送注册邮箱验证码
// @Accept  json
// @Produce  json
// @Param body body admin.SendRegisterCodeForm true "邮箱"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user/register-code [post]
func (ct *User) SendRegisterCode(c *gin.Context) {
	if !global.Config.App.Register {
		response.Fail(c, 101, response.TranslateMsg(c, "RegisterClosed"))
		return
	}
	f := &admin.SendRegisterCodeForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	email := NormalizeEmail(f.Email)
	if email == "" {
		response.Fail(c, 101, response.TranslateMsg(c, "EmailRequired"))
		return
	}
	if !ValidEmail(email) {
		response.Fail(c, 101, response.TranslateMsg(c, "EmailInvalid"))
		return
	}
	if service.AllService.UserService.InfoByEmail(email).Id != 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "EmailExists"))
		return
	}
	if global.EmailCodeStore == nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed"))
		return
	}

	code, err := global.EmailCodeStore.Issue(email, realClientIP(c))
	if err != nil {
		switch err {
		case utils.ErrEmailCodeTooSoon:
			wait := global.EmailCodeStore.ResendAfter(email)
			response.Fail(c, 101, response.TranslateTempMsg(c, "EmailCodeResendTooSoon", map[string]interface{}{
				"Seconds": int(wait.Seconds()) + 1,
			}))
		case utils.ErrEmailCodeTooOften:
			response.Fail(c, 101, response.TranslateMsg(c, "RegisterTooFrequent"))
		default:
			global.Logger.Warn(fmt.Sprintf("生成注册验证码失败: %v", err))
			response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed"))
		}
		return
	}

	subject := response.TranslateTempMsg(c, "EmailCodeSubject", nil)
	body := response.TranslateTempMsg(c, "EmailCodeBody", map[string]interface{}{
		"Code":    code,
		"Minutes": int(utils.EmailCodeTTL.Minutes()),
	})
	if !global.Config.Mail.Ready() {
		// 调试模式：SMTP 未配置时把验证码写进日志；生产环境直接报错，避免用户收不到邮件。
		if global.Config.Gin.Mode == "release" {
			global.EmailCodeStore.Discard(email)
			response.Fail(c, 101, response.TranslateMsg(c, "MailNotConfigured"))
			return
		}
		global.Logger.Warn(fmt.Sprintf("SMTP 未配置，注册验证码仅写入日志 email=%s code=%s", email, code))
	} else if err := utils.SendMail(&global.Config.Mail, email, subject, body); err != nil {
		global.Logger.Warn(fmt.Sprintf("发送注册验证码失败 email=%s err=%v", email, err))
		global.EmailCodeStore.Discard(email)
		response.Fail(c, 101, response.TranslateMsg(c, "EmailSendFailed"))
		return
	}

	response.Success(c, gin.H{
		"resend_after": int(utils.EmailCodeResendAfter.Seconds()),
		"expires_in":   int(utils.EmailCodeTTL.Seconds()),
	})
}

// Register
func (ct *User) Register(c *gin.Context) {
	if !global.Config.App.Register {
		response.Fail(c, 101, response.TranslateMsg(c, "RegisterClosed"))
		return
	}
	f := &admin.RegisterForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	if f.ConfirmPassword != f.Password {
		response.Fail(c, 101, response.TranslateMsg(c, "PasswordNotMatchConfirmPassword"))
		return
	}

	// 邮箱：开启验证时必填，任何情况下填写了都要格式正确且未被占用。
	verifyEmail := global.Config.App.RegisterEmailVerify
	email := NormalizeEmail(f.Email)
	if email == "" {
		if verifyEmail {
			response.Fail(c, 101, response.TranslateMsg(c, "EmailRequired"))
			return
		}
	} else {
		if !ValidEmail(email) {
			response.Fail(c, 101, response.TranslateMsg(c, "EmailInvalid"))
			return
		}
		if service.AllService.UserService.InfoByEmail(email).Id != 0 {
			response.Fail(c, 101, response.TranslateMsg(c, "EmailExists"))
			return
		}
	}

	// 注册限流：只检查不占名额，实际创建成功后才记录，避免验证码输错消耗配额。
	clientIP := realClientIP(c)
	if global.RegisterLimiter != nil && !global.RegisterLimiter.Check(clientIP) {
		response.Fail(c, 101, response.TranslateMsg(c, "RegisterTooFrequent"))
		return
	}
	if service.AllService.UserService.IsUsernameExists(f.Username) {
		response.Fail(c, 101, response.TranslateMsg(c, "UsernameExists"))
		return
	}

	if verifyEmail {
		code := strings.TrimSpace(f.Code)
		if code == "" {
			response.Fail(c, 101, response.TranslateMsg(c, "EmailCodeRequired"))
			return
		}
		if global.EmailCodeStore == nil {
			response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed"))
			return
		}
		switch err := global.EmailCodeStore.Verify(email, code); err {
		case nil:
			// 校验通过
		case utils.ErrEmailCodeExpired:
			response.Fail(c, 101, response.TranslateMsg(c, "EmailCodeExpired"))
			return
		case utils.ErrEmailCodeTooMany:
			response.Fail(c, 101, response.TranslateMsg(c, "EmailCodeTooManyAttempts"))
			return
		default:
			response.Fail(c, 101, response.TranslateMsg(c, "EmailCodeInvalid"))
			return
		}
	}

	// register-status: 1 注册后直接启用；2/3 注册后待管理员审核。
	regStatus := model.StatusCode(global.Config.App.RegisterStatus)
	switch regStatus {
	case model.COMMON_STATUS_ENABLE:
		// 直接启用
	case model.COMMON_STATUS_DISABLED, model.COMMON_STATUS_PENDING:
		regStatus = model.COMMON_STATUS_PENDING
	default:
		regStatus = model.COMMON_STATUS_ENABLE
	}

	u := service.AllService.UserService.Register(f.Username, email, f.Password, regStatus)
	if u == nil || u.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed"))
		return
	}
	if global.RegisterLimiter != nil {
		global.RegisterLimiter.Record(clientIP)
	}
	if regStatus == model.COMMON_STATUS_PENDING {
		// 需要管理员审核：不返回 token，前端提示后跳转登录页。
		response.Success(c, gin.H{"pending": true})
		return
	}
	// 注册成功后自动登录
	ut := service.AllService.UserService.Login(u, &model.LoginLog{
		UserId: u.Id,
		Client: model.LoginLogClientWebAdmin,
		Uuid:   "",
		Ip:     c.ClientIP(),
		Type:   model.LoginLogTypeAccount,
	})
	responseLoginSuccess(c, u, ut.Token)
}
