package constant

// 错误定义

// 通用错误
const (
	Success               = 0   // 成功
	ErrUnknown            = 1001 // 未知错误
	ErrInvalidParam       = 1002 // 参数无效
	ErrUnauthorized       = 1003 // 未授权
	ErrForbidden          = 1004 // 禁止访问
	ErrNotFound           = 1005 // 资源不存在
	ErrInternalServer     = 1006 // 服务器内部错误
	ErrServiceUnavailable = 1007 // 服务不可用
	ErrTimeout            = 1008 // 请求超时
)

// 数据库相关错误
const (
	ErrDBConnectFailed   = 2001 // 数据库连接失败
	ErrDBQuery           = 2002 // 数据库查询错误
	ErrDBExec            = 2003 // 数据库执行错误
	ErrDBRecordNotFound  = 2004 // 记录未找到
	ErrDBDuplicateKey    = 2005 // 重复键值
	ErrDBTransactionFail = 2006 // 事务执行失败
)

// 设备相关错误
const (
	ErrDeviceOffline      = 3001 // 设备离线
	ErrDeviceNotExists    = 3002 // 设备不存在
	ErrDeviceAuthFailed   = 3003 // 设备认证失败
	ErrDeviceCmdNotSupported = 3004 // 不支持的设备命令
	ErrDeviceRespTimeout  = 3005 // 设备响应超时
)

// 用户相关错误
const (
	ErrUserNotExists    = 4001 // 用户不存在
	ErrUserAuthFailed   = 4002 // 用户认证失败
	ErrUserDisabled     = 4003 // 用户已被禁用
	ErrUserAlreadyExist = 4004 // 用户已存在
)

// 语音相关错误
const (
	ErrVoiceProcessFailed = 5001 // 语音处理失败
	ErrVoiceTooLong       = 5002 // 语音时间过长
	ErrVoiceFormatNotSupport = 5003 // 语音格式不支持
)
