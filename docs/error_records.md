# 错误码定义文档

## 1. 通用错误 (1000-1999)

| 错误码 | 常量名 | 描述 | 解决方案 |
|-------|--------|------|---------|
| 0 | Success | 成功 | 无需处理 |
| 1001 | ErrUnknown | 未知错误 | 检查系统日志 |
| 1002 | ErrInvalidParam | 参数无效 | 检查请求参数是否符合规范 |
| 1003 | ErrUnauthorized | 未授权 | 需要提供有效的身份凭证 |
| 1004 | ErrForbidden | 禁止访问 | 当前用户无权限执行此操作 |
| 1005 | ErrNotFound | 资源不存在 | 检查请求的资源ID是否正确 |
| 1006 | ErrInternalServer | 服务器内部错误 | 联系技术支持 |
| 1007 | ErrServiceUnavailable | 服务不可用 | 稍后重试 |
| 1008 | ErrTimeout | 请求超时 | 检查网络连接或稍后重试 |

## 2. 数据库相关错误 (2000-2999)

| 错误码 | 常量名 | 描述 | 解决方案 |
|-------|--------|------|---------|
| 2001 | ErrDBConnectFailed | 数据库连接失败 | 检查数据库配置和服务状态 |
| 2002 | ErrDBQuery | 数据库查询错误 | 检查SQL语句和参数 |
| 2003 | ErrDBExec | 数据库执行错误 | 检查SQL语句和数据完整性 |
| 2004 | ErrDBRecordNotFound | 记录未找到 | 检查查询条件是否正确 |
| 2005 | ErrDBDuplicateKey | 重复键值 | 修改数据避免唯一性冲突 |
| 2006 | ErrDBTransactionFail | 事务执行失败 | 检查事务中的操作是否符合约束 |

## 3. 设备相关错误 (3000-3999)

| 错误码 | 常量名 | 描述 | 解决方案 |
|-------|--------|------|---------|
| 3001 | ErrDeviceOffline | 设备离线 | 检查设备网络连接 |
| 3002 | ErrDeviceNotExists | 设备不存在 | 检查设备ID是否正确 |
| 3003 | ErrDeviceAuthFailed | 设备认证失败 | 检查设备认证信息 |
| 3004 | ErrDeviceCmdNotSupported | 不支持的设备命令 | 查阅设备支持的指令集 |
| 3005 | ErrDeviceRespTimeout | 设备响应超时 | 检查设备状态或网络延迟 |

## 4. 用户相关错误 (4000-4999)

| 错误码 | 常量名 | 描述 | 解决方案 |
|-------|--------|------|---------|
| 4001 | ErrUserNotExists | 用户不存在 | 检查用户名或用户ID |
| 4002 | ErrUserAuthFailed | 用户认证失败 | 检查用户名和密码 |
| 4003 | ErrUserDisabled | 用户已被禁用 | 联系管理员启用账户 |
| 4004 | ErrUserAlreadyExist | 用户已存在 | 使用其他用户名注册 |

## 5. 语音相关错误 (5000-5999)

| 错误码 | 常量名 | 描述 | 解决方案 |
|-------|--------|------|---------|
| 5001 | ErrVoiceProcessFailed | 语音处理失败 | 检查音频文件格式和内容 |
| 5002 | ErrVoiceTooLong | 语音时间过长 | 缩短语音长度至允许范围内 |
| 5003 | ErrVoiceFormatNotSupport | 语音格式不支持 | 转换为支持的音频格式 |
