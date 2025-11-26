-- 1. 插入未激活的公版设备
INSERT INTO device_base (
    device_sn, device_type, vendor_id, vendor_name, 
    hardware_version, firmware_version, product_model, manufacture_date
) VALUES (
    'DOG-SN-20250001', 'public_dog_mini', 0, '自研',
    'V1.0', 'V1.0.0', 'DogBot-001', '2025-01-01 00:00:00'
);

-- 2. 激活设备（更新状态和激活时间）
UPDATE device_base 
SET status = 'activated', activation_time = CURRENT_TIMESTAMP 
WHERE device_sn = 'DOG-SN-20250001';

-- 3. 插入设备WiFi网络信息（已连接状态）
INSERT INTO device_network (
    device_id, network_type, mac_address, ip_address, port, 
    connect_status, last_connect_time, signal_strength
) VALUES (
    1, 'wifi', 'AA:BB:CC:DD:EE:FF', '192.168.1.100', 1883,
    'connected', CURRENT_TIMESTAMP, -50
);

-- 4. 初始化设备状态
INSERT INTO device_status (device_id) VALUES (1);

-- 5. Mock 设备断网
UPDATE device_network 
SET connect_status = 'disconnected', last_disconnect_time = CURRENT_TIMESTAMP 
WHERE device_id = 1 AND network_type = 'wifi';

-- 6. Mock 设备重新联网+上报心跳+消息传输（语音对话中）
UPDATE device_network 
SET connect_status = 'connected', ip_address = '192.168.1.101', 
    last_connect_time = CURRENT_TIMESTAMP, signal_strength = -45 
WHERE device_id = 1 AND network_type = 'wifi';

UPDATE device_status 
SET last_heartbeat_time = CURRENT_TIMESTAMP, battery_level = 85,
    working_status = 'voice_chat', cpu_usage = 30.5, memory_usage = 45.2,
    last_message_time = CURRENT_TIMESTAMP 
WHERE device_id = 1;