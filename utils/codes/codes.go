// Package codes 统一业务状态码与业务错误类型定义。
//
// 说明：
//  1. code=0 代表成功；
//  2. 40xxx 代表通用参数/调用方错误；
//  3. 400xx 代表用户认证相关错误；
//  4. 41xxx 代表后台管理业务错误；
//  5. 42xxx 代表后台管理认证/权限错误；
//  6. 429xx 代表限流/频控错误；
//  7. 5xxxx 代表网关/依赖/系统错误；
//  8. 各服务新增码值时优先在此集中维护，避免散落硬编码。
package codes

// ─────────────────────────────────────────────────────────────────────────────
// BizError 业务错误，携带面向客户端的错误码与消息。
//
// 用法：
//   - 服务层返回 *BizError，例如 return nil, codes.ErrAlreadyVoted
//   - RPC/HTTP 层用 errors.As 识别 *BizError，将 Code/Msg 直接写入响应
//   - 非 *BizError 类型的错误统一以 500 兜底
// ─────────────────────────────────────────────────────────────────────────────

// BizError 业务错误，携带面向客户端的错误码与消息。
type BizError struct {
	Code int32  // 面向前端的业务错误码
	Msg  string // 面向前端的错误描述
}

// Error 实现 error 接口。
func (e *BizError) Error() string { return e.Msg }

// NewBizError 创建一个新的 BizError。
func NewBizError(code int32, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// ═════════════════════════════════════════════════════════════════════════════
// General — 通用状态码 (0, 40xxx)
// ═════════════════════════════════════════════════════════════════════════════

const (
	// OK 请求成功。
	OK = 0

	// BadRequest 请求参数不合法（通用）。
	BadRequest = 40000
	// InvalidRequestBody 请求体 JSON 非法（通用）。
	InvalidRequestBody = 40001
	// OptionIDRequired 投票选项必填。
	OptionIDRequired = 40002
	// InvalidID 路径或参数 ID 非法。
	InvalidID = 40011
)

// ═════════════════════════════════════════════════════════════════════════════
// Auth — 用户认证相关业务码 (400xx)
// ═════════════════════════════════════════════════════════════════════════════

const (
	// UserAuthInvalidBodySendSMS 发送短信请求体非法。
	UserAuthInvalidBodySendSMS = 40041
	// UserAuthPhoneRequired 手机号必填。
	UserAuthPhoneRequired = 40042
	// UserAuthInvalidBodyReg 注册请求体非法。
	UserAuthInvalidBodyReg = 40043
	// UserAuthPhonePwdRequired 手机号与密码必填。
	UserAuthPhonePwdRequired = 40044
	// UserAuthInvalidBodyLogin 登录请求体非法。
	UserAuthInvalidBodyLogin = 40045
	// UserAuthPhonePwdNeed 缺少手机号或密码。
	UserAuthPhonePwdNeed = 40046
	// UserAuthInvalidBodySMS 短信验证码请求体非法。
	UserAuthInvalidBodySMS = 40047
	// UserAuthPhoneCodeRequired 手机号与验证码必填。
	UserAuthPhoneCodeRequired = 40048
	// UserAuthAccessTokenNeed 缺少 AccessToken。
	UserAuthAccessTokenNeed = 40049
)

// ═════════════════════════════════════════════════════════════════════════════
// Admin — 后台管理业务码 (41xxx)
// ═════════════════════════════════════════════════════════════════════════════

const (
	// AdminBizInvalidRequest 后台请求参数不合法。
	AdminBizInvalidRequest = 41001
	// AdminBizEmptyUpdate 更新内容为空。
	AdminBizEmptyUpdate = 41002
	// AdminBizResourceNotFound 业务资源不存在。
	AdminBizResourceNotFound = 41004
)

// ═════════════════════════════════════════════════════════════════════════════
// Admin Auth — 后台管理认证与权限码 (42xxx)
// ═════════════════════════════════════════════════════════════════════════════

const (
	// AdminAuthUnauthorized 未登录或会话无效。
	AdminAuthUnauthorized = 42001
	// AdminAuthTokenInvalid Token 无效或已过期。
	AdminAuthTokenInvalid = 42002
	// AdminAuthForbidden 权限不足。
	AdminAuthForbidden = 42003
	// AdminAuthUserDisabled 账号被禁用或不存在。
	AdminAuthUserDisabled = 42004
	// AdminAuthRateLimited 认证相关请求触发限流。
	AdminAuthRateLimited = 42005
)

// ═════════════════════════════════════════════════════════════════════════════
// Business — 业务逻辑错误码 (40xxx, 429xx)
// ═════════════════════════════════════════════════════════════════════════════

const (
	// BizAlreadyVoted 该设备已对此图纸投过票。
	BizAlreadyVoted = 40033
	// BizInvalidVoterFingerprint 无法识别投票设备指纹。
	BizInvalidVoterFingerprint = 40034
	// BizVoteTooFrequent 同一设备短时间内投票过于频繁。
	BizVoteTooFrequent = 42931
)

// 预定义 BizError 实例（业务层可直接返回）。
var (
	// ErrAlreadyVoted 该设备已对此图纸投过票。
	ErrAlreadyVoted = &BizError{Code: BizAlreadyVoted, Msg: "already voted"}
	// ErrVoteTooFrequent 同一设备短时间内投票过于频繁。
	ErrVoteTooFrequent = &BizError{Code: BizVoteTooFrequent, Msg: "vote too frequent"}
	// ErrInvalidVoterFingerprint 无法识别投票设备指纹。
	ErrInvalidVoterFingerprint = &BizError{Code: BizInvalidVoterFingerprint, Msg: "invalid voter fingerprint"}
)

// ═════════════════════════════════════════════════════════════════════════════
// Upstream — 网关/依赖服务错误码 (50xxx)
// ═════════════════════════════════════════════════════════════════════════════

const (
	// UpstreamUnavailable 上游服务不可达。
	UpstreamUnavailable = 50201
	// UpstreamEmptyReply 上游服务返回空响应。
	UpstreamEmptyReply = 50202
	// UpstreamBadPayload 上游服务返回非法数据。
	UpstreamBadPayload = 50203
)

// ═════════════════════════════════════════════════════════════════════════════
// Admin System — 后台系统/基础设施错误码 (51xxx)
// ═════════════════════════════════════════════════════════════════════════════

const (
	// AdminSysInternalError 服务内部错误。
	AdminSysInternalError = 51001
	// AdminSysDatabaseError 数据库错误。
	AdminSysDatabaseError = 51002
	// AdminSysRedisError Redis 错误。
	AdminSysRedisError = 51003
)
