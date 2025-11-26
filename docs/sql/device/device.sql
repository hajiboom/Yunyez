DROP TYPE IF EXISTS device_global_status;
DROP TYPE IF EXISTS network_type_enum;
DROP TYPE IF EXISTS connect_status_enum;
DROP TYPE IF EXISTS power_status_enum;
DROP TYPE IF EXISTS working_status_enum;


-- 设备全局状态枚举
CREATE TYPE device_global_status AS ENUM (
    'inactivated',  -- 未激活（对应原 0）
    'activated',    -- 已激活（对应原 1）
    'disabled',     -- 禁用（对应原 2）
    'scrapped'      -- 报废（对应原 3）
);

-- 网络类型枚举
CREATE TYPE network_type_enum AS ENUM (
    'wifi',         -- WiFi（对应原 1）
    'cellular',     -- 蜂窝网络（4G/5G，对应原 2）
    'bluetooth',    -- 蓝牙（对应原 3）
    'ethernet'      -- 以太网（对应原 4）
);

-- 网络连接状态枚举
CREATE TYPE connect_status_enum AS ENUM (
    'disconnected', -- 断开（对应原 0）
    'connecting',   -- 连接中（对应原 1）
    'connected'     -- 已连接（对应原 2)
);

-- 设备供电状态枚举
CREATE TYPE power_status_enum AS ENUM (
    'shutdown',     -- 关机（对应原 0）
    'power_on',     -- 开机（对应原 1）
    'standby',      -- 待机（对应原 2）
    'charging'      -- 充电中（对应原 3）
);

-- 设备工作状态枚举
CREATE TYPE working_status_enum AS ENUM (
    'idle',         -- 空闲（对应原 0）
    'active',       -- 活动中（对应原 1）
    'busy'         -- 忙碌（对应原 2）
);

DROP TABLE IF EXISTS device_base;
CREATE TABLE device_base (
    id BIGSERIAL PRIMARY KEY,
    device_sn VARCHAR(64) NOT NULL UNIQUE,
    imei VARCHAR(32) UNIQUE,
    iccid VARCHAR(32) UNIQUE,
    device_type VARCHAR(32) NOT NULL,
    vendor_id BIGINT NOT NULL,
    vendor_name VARCHAR(64) NOT NULL,
    hardware_version VARCHAR(32) NOT NULL,
    firmware_version VARCHAR(32) NOT NULL,
    product_model VARCHAR(64) NOT NULL,
    manufacture_date TIMESTAMP NOT NULL,
    expire_date TIMESTAMP,
    status device_global_status NOT NULL DEFAULT 'inactivated',
    activation_time TIMESTAMP,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    remark VARCHAR(255)
);

-- 为 device_base 表的列添加注释
COMMENT ON COLUMN device_base.device_sn IS '设备序列号（硬件唯一标识）';
COMMENT ON COLUMN device_base.imei IS '国际移动设备识别码（仅蜂窝网络设备必填）';
COMMENT ON COLUMN device_base.iccid IS '集成电路卡识别码（仅插SIM卡设备必填）';
COMMENT ON COLUMN device_base.device_type IS '设备类型（如 public_dog_mini/vendor_a_pro）';
COMMENT ON COLUMN device_base.vendor_id IS '厂商ID（0=自研公版，其他为定制厂商）';
COMMENT ON COLUMN device_base.vendor_name IS '厂商名称（冗余字段）';
COMMENT ON COLUMN device_base.hardware_version IS '硬件版本（如 V1.0/V2.1）';
COMMENT ON COLUMN device_base.firmware_version IS '固件版本（初始为出厂版本）';
COMMENT ON COLUMN device_base.product_model IS '产品型号（如 DogBot-001/VendorB-Dog-Pro）';
COMMENT ON COLUMN device_base.manufacture_date IS '生产日期';
COMMENT ON COLUMN device_base.expire_date IS '质保到期日（可选）';
COMMENT ON COLUMN device_base.status IS '设备全局状态';
COMMENT ON COLUMN device_base.activation_time IS '激活时间';
COMMENT ON COLUMN device_base.create_time IS '记录创建时间';
COMMENT ON COLUMN device_base.update_time IS '记录更新时间';
COMMENT ON COLUMN device_base.deleted_at IS '软删除标记时间';
COMMENT ON COLUMN device_base.remark IS '备注（定制设备配置说明等）';


-- 设备网络连接信息表
DROP TABLE IF EXISTS device_network;
CREATE TABLE device_network (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL,
    network_type network_type_enum NOT NULL,
    mac_address VARCHAR(64) UNIQUE,
    ip_address VARCHAR(64),
    port INT,
    signal_strength INT,
    connect_status connect_status_enum NOT NULL DEFAULT 'disconnected',
    last_connect_time TIMESTAMP,
    last_disconnect_time TIMESTAMP,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL
);

-- 为 device_network 表的列添加注释
COMMENT ON COLUMN device_network.device_id IS '关联设备主表ID';
COMMENT ON COLUMN device_network.network_type IS '网络类型';
COMMENT ON COLUMN device_network.mac_address IS 'MAC地址（WiFi/以太网必填）';
COMMENT ON COLUMN device_network.ip_address IS '设备IP地址（动态更新）';
COMMENT ON COLUMN device_network.port IS '通信端口（如MQTT端口1883）';
COMMENT ON COLUMN device_network.signal_strength IS '信号强度（-100~0，值越大信号越好）';
COMMENT ON COLUMN device_network.connect_status IS '连接状态';
COMMENT ON COLUMN device_network.last_connect_time IS '最后一次连接时间';
COMMENT ON COLUMN device_network.last_disconnect_time IS '最后一次断开时间';
COMMENT ON COLUMN device_network.create_time IS '记录创建时间';
COMMENT ON COLUMN device_network.update_time IS '记录更新时间';
COMMENT ON COLUMN device_network.deleted_at IS '软删除标记时间';

-- 设备状态表
DROP TABLE IF EXISTS device_status;
CREATE TABLE device_status (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL,
    battery_level INT NOT NULL DEFAULT 100 CHECK (battery_level BETWEEN 0 AND 100),
    power_status power_status_enum NOT NULL DEFAULT 'power_on',
    working_status working_status_enum NOT NULL DEFAULT 'idle',
    last_heartbeat_time TIMESTAMP,
    last_message_time TIMESTAMP,
    cpu_usage NUMERIC(5,2) CHECK (cpu_usage BETWEEN 0 AND 100),
    memory_usage NUMERIC(5,2) CHECK (memory_usage BETWEEN 0 AND 100),
    error_code VARCHAR(32),
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL
);

-- 为 device_status 表的列添加注释
COMMENT ON COLUMN device_status.device_id IS '关联设备主表ID';
COMMENT ON COLUMN device_status.battery_level IS '电池电量（0~100）';
COMMENT ON COLUMN device_status.power_status IS '供电状态';
COMMENT ON COLUMN device_status.working_status IS '工作状态';
COMMENT ON COLUMN device_status.last_heartbeat_time IS '最后一次心跳时间';
COMMENT ON COLUMN device_status.last_message_time IS '最后一次消息传输时间';
COMMENT ON COLUMN device_status.cpu_usage IS 'CPU使用率（0~100）';
COMMENT ON COLUMN device_status.memory_usage IS '内存使用率（0~100）';
COMMENT ON COLUMN device_status.error_code IS '错误码（硬件故障/网络异常等）';
COMMENT ON COLUMN device_status.create_time IS '记录创建时间';
COMMENT ON COLUMN device_status.update_time IS '记录更新时间';
COMMENT ON COLUMN device_status.deleted_at IS '软删除标记时间';


-- ============== 索引 ============== 

-- 设备主表索引（高频查询字段）
CREATE INDEX idx_device_base_sn ON device_base (device_sn);
CREATE INDEX idx_device_base_imei ON device_base (imei);
CREATE INDEX idx_device_base_iccid ON device_base (iccid);
CREATE INDEX idx_device_base_vendor_id ON device_base (vendor_id);
CREATE INDEX idx_device_base_status ON device_base (status);
CREATE INDEX idx_device_base_deleted_at ON device_base (deleted_at);

-- 设备网络表索引（关联查询+状态查询）
CREATE INDEX idx_device_network_device_id ON device_network (device_id);
CREATE INDEX idx_device_network_connect_status ON device_network (connect_status);
CREATE INDEX idx_device_network_network_type ON device_network (network_type);
CREATE INDEX idx_device_network_deleted_at ON device_network (deleted_at);

-- 设备状态表索引（关联查询+心跳/消息时间查询）
CREATE INDEX idx_device_status_device_id ON device_status (device_id);
CREATE INDEX idx_device_status_last_heartbeat ON device_status (last_heartbeat_time);
CREATE INDEX idx_device_status_working_status ON device_status (working_status);
CREATE INDEX idx_device_status_deleted_at ON device_status (deleted_at);
