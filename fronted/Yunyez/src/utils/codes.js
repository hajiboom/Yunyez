// utils/errorCodes.js
export const ERROR_CODES = {
  // 成功
  SUCCESS: 0,
  
  // 登录相关错误 (400)2
  USER_NOT_FOUND: 40001,      // 用户不存在
  PASSWORD_ERROR: 40002,      // 密码错误
  ACCOUNT_LOCKED: 40003,      // 账号已锁定
  ACCOUNT_DISABLED: 40004,    // 账号已禁用
  PARAM_INVALID: 40005,       // 参数校验失败
  
  // Token相关错误 (401)
  TOKEN_INVALID: 40101,       // Token无效（accessToken）
  TOKEN_EXPIRED: 40102,       // Token已过期（accessToken）
  TOKEN_REVOKED: 40103,       // Token已吊销（accessToken）
  
  // 刷新Token相关错误 (401) - 新增
  REFRESH_TOKEN_INVALID: 40111,   // 刷新令牌无效
  REFRESH_TOKEN_EXPIRED: 40112,   // 刷新令牌已过期
  REFRESH_TOKEN_REVOKED: 40113,   // 刷新令牌已被吊销
  
  // 权限错误 (403)
  PERMISSION_DENIED: 40301,   // 权限不足
  
  // 限流错误 (429)
  RATE_LIMIT: 42901,          // 请求过于频繁
  
  // 服务器错误 (500)
  SERVER_ERROR: 50000         // 服务器内部错误
}

// 错误码对应提示消息
export const ERROR_MESSAGES = {
  [ERROR_CODES.USER_NOT_FOUND]: '用户名或密码错误',
  [ERROR_CODES.PASSWORD_ERROR]: '用户名或密码错误',
  [ERROR_CODES.ACCOUNT_LOCKED]: '账号已被锁定，请 15 分钟后重试',
  [ERROR_CODES.ACCOUNT_DISABLED]: '账号已被禁用，请联系管理员',
  [ERROR_CODES.PARAM_INVALID]: '请求参数错误',
  [ERROR_CODES.TOKEN_INVALID]: '无效的令牌',
  [ERROR_CODES.TOKEN_EXPIRED]: '令牌已过期',
  [ERROR_CODES.TOKEN_REVOKED]: '令牌已被吊销',
  [ERROR_CODES.REFRESH_TOKEN_INVALID]: '无效的刷新令牌',
  [ERROR_CODES.REFRESH_TOKEN_EXPIRED]: '刷新令牌已过期',
  [ERROR_CODES.REFRESH_TOKEN_REVOKED]: '刷新令牌已被吊销',
  [ERROR_CODES.PERMISSION_DENIED]: '权限不足',
  [ERROR_CODES.RATE_LIMIT]: '请求过于频繁，请稍后再试',
  [ERROR_CODES.SERVER_ERROR]: '服务器内部错误'
}