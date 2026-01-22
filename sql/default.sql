-- 用户与账户系统
CREATE SCHEMA auth;

-- 设备与 IoT 通信（设备注册、心跳、指令下发）
CREATE SCHEMA device;

-- 媒体存储元数据（语音、图像的 URL、时间戳、标签等，实际文件存对象存储如 S3/MinIO）
CREATE SCHEMA media;

-- 日志与审计（可选）
CREATE SCHEMA logging;

-- 公共模式，默认模式
CREATE SCHEMA public;