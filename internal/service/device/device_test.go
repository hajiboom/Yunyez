// internal/service/device/device_test.go

package device

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"yunyez/internal/pkg/logger"
	"yunyez/internal/common/tools"
	"yunyez/internal/model/device"
	"yunyez/internal/pkg/postgre"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// randomMAC 生成随机 MAC 地址用于测试（格式 xx:xx:xx:xx:xx:xx）
func randomMAC() string {
	uuid := uuid.New()
	// 使用 UUID 的前 6 个字节生成 MAC 地址
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		uuid[0], uuid[1], uuid[2],
		uuid[3], uuid[4], uuid[5])
}

// DeviceServiceTestSuite 设备服务测试套件
type DeviceServiceTestSuite struct {
	suite.Suite
	service Service
	db      *gorm.DB
	ctx     context.Context
}

// SetupSuite 初始化测试套件：重建干净表
func (suite *DeviceServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// 记录测试开始日志
	logger.Info(suite.ctx, "开始设置设备服务测试套件", map[string]interface{}{
		"action": "setup_suite",
	})
	
	client := postgre.GetClient()
	suite.db = client.DB

	// 🔥 先删除旧表（如果存在）
	logger.Info(suite.ctx, "删除现有设备相关表", map[string]interface{}{
		"action": "drop_tables",
	})
	_ = suite.db.Migrator().DropTable(
		&device.BaseDevice{},
		&device.DeviceNetwork{},
		&device.DeviceStatus{},
	)

	// 🆕 再自动创建新表（由 GORM 管理索引/约束）
	logger.Info(suite.ctx, "创建设备相关表", map[string]interface{}{
		"action": "auto_migrate",
	})
	err := suite.db.AutoMigrate(
		&device.BaseDevice{},
		&device.DeviceNetwork{},
		&device.DeviceStatus{},
	)
	if err != nil {
		logger.Error(suite.ctx, "自动迁移表失败", map[string]interface{}{
			"action": "auto_migrate",
			"error":  err.Error(),
		})
		suite.T().Fatalf("AutoMigrate failed: %v", err)
	}

	// 创建服务实例
	dbProvider := &PostgreClient{Client: client}
	suite.service = NewService(dbProvider)
	
	logger.Info(suite.ctx, "设备服务测试套件设置完成", map[string]interface{}{
		"action": "setup_complete",
	})
}

// TearDownSuite 清理测试数据（可选，因为每次测试都重建表）
func (suite *DeviceServiceTestSuite) TearDownSuite() {
	// 记录测试结束日志
	logger.Info(suite.ctx, "设备服务测试套件清理完成", map[string]interface{}{
		"action": "teardown_complete",
	})
}

// createTestDevice 创建测试设备
func (suite *DeviceServiceTestSuite) createTestDevice() *device.BaseDevice {
	deviceSN := "test_" + uuid.New().String()
	logger.Info(suite.ctx, "创建测试设备", map[string]interface{}{
		"action":    "create_test_device",
		"device_sn": deviceSN,
	})
	
	baseDevice := &device.BaseDevice{
		DeviceSN:        deviceSN,
		IMEI:            "test_imei_" + uuid.New().String()[:10],
		ICCID:           "test_iccid_" + uuid.New().String()[:10],
		DeviceType:      "sensor",
		VendorID:        1,
		VendorName:      "TestVendor",
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0",
		ProductModel:    "TestModel",
		ManufactureDate: time.Now(),
		ExpireDate:      time.Now().AddDate(1, 0, 0),
		Status:          "inactivated",
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}

	err := suite.service.RegisterDevice(suite.ctx, baseDevice)
	if err != nil {
		logger.Error(suite.ctx, "注册测试设备失败", map[string]interface{}{
			"action":    "register_device",
			"device_sn": deviceSN,
			"error":     err.Error(),
		})
		suite.NoError(err)
		return nil
	}
	
	logger.Info(suite.ctx, "测试设备创建成功", map[string]interface{}{
		"action":    "create_test_device_success",
		"device_sn": deviceSN,
		"device_id": baseDevice.ID,
	})
	
	return baseDevice
}

// ==================== 测试用例 ====================

func (suite *DeviceServiceTestSuite) TestRegisterDevice() {
	logger.Info(suite.ctx, "开始执行设备注册测试", map[string]interface{}{
		"action": "test_register_device",
	})
	
	deviceSN := "test_register_" + uuid.New().String()
	baseDevice := &device.BaseDevice{
		DeviceSN:        deviceSN,
		DeviceType:      "sensor",
		VendorID:        1,
		VendorName:      "TestVendor",
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0",
		ProductModel:    "TestModel",
		ManufactureDate: time.Now(),
		Status:          "inactivated",
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}

	err := suite.service.RegisterDevice(suite.ctx, baseDevice)
	suite.NoError(err)
	suite.NotZero(baseDevice.ID)

	retrievedDevice, err := suite.service.GetDeviceBySN(suite.ctx, deviceSN)
	suite.NoError(err)
	suite.Equal(deviceSN, retrievedDevice.DeviceSN)
	suite.Equal("sensor", retrievedDevice.DeviceType)
	
	logger.Info(suite.ctx, "设备注册测试完成", map[string]interface{}{
		"action":    "test_register_device_complete",
		"device_sn": deviceSN,
		"device_id": baseDevice.ID,
	})
}

func (suite *DeviceServiceTestSuite) TestGetDeviceByID() {
	logger.Info(suite.ctx, "开始执行根据ID获取设备测试", map[string]interface{}{
		"action": "test_get_device_by_id",
	})
	
	testDevice := suite.createTestDevice()
	retrievedDevice, err := suite.service.GetDeviceByID(suite.ctx, testDevice.ID)
	suite.NoError(err)
	suite.NotNil(retrievedDevice)
	suite.Equal(testDevice.DeviceSN, retrievedDevice.DeviceSN)
	
	logger.Info(suite.ctx, "根据ID获取设备测试完成", map[string]interface{}{
		"action":    "test_get_device_by_id_complete",
		"device_id": testDevice.ID,
	})
}

func (suite *DeviceServiceTestSuite) TestGetDeviceBySN() {
	logger.Info(suite.ctx, "开始执行根据SN获取设备测试", map[string]interface{}{
		"action": "test_get_device_by_sn",
	})
	
	testDevice := suite.createTestDevice()
	retrievedDevice, err := suite.service.GetDeviceBySN(suite.ctx, testDevice.DeviceSN)
	suite.NoError(err)
	suite.NotNil(retrievedDevice)
	suite.Equal(testDevice.ID, retrievedDevice.ID)
	
	logger.Info(suite.ctx, "根据SN获取设备测试完成", map[string]interface{}{
		"action":    "test_get_device_by_sn_complete",
		"device_sn": testDevice.DeviceSN,
	})
}

func (suite *DeviceServiceTestSuite) TestUpdateDevice() {
	logger.Info(suite.ctx, "开始执行更新设备测试", map[string]interface{}{
		"action": "test_update_device",
	})
	
	testDevice := suite.createTestDevice()
	updates := map[string]interface{}{
		"device_type": "updated_sensor",
		"vendor_name": "UpdatedVendor",
	}
	err := suite.service.UpdateDevice(suite.ctx, testDevice.ID, updates)
	suite.NoError(err)

	updatedDevice, err := suite.service.GetDeviceByID(suite.ctx, testDevice.ID)
	suite.NoError(err)
	suite.Equal("updated_sensor", updatedDevice.DeviceType)
	suite.Equal("UpdatedVendor", updatedDevice.VendorName)
	
	logger.Info(suite.ctx, "更新设备测试完成", map[string]interface{}{
		"action":    "test_update_device_complete",
		"device_id": testDevice.ID,
	})
}

func (suite *DeviceServiceTestSuite) TestDeleteDevice() {
	logger.Info(suite.ctx, "开始执行删除设备测试", map[string]interface{}{
		"action": "test_delete_device",
	})
	
	testDevice := suite.createTestDevice()
	err := suite.service.DeleteDevice(suite.ctx, testDevice.ID)
	suite.NoError(err)

	_, err = suite.service.GetDeviceByID(suite.ctx, testDevice.ID)
	suite.Error(err)
	suite.Contains(err.Error(), "record not found")
	
	logger.Info(suite.ctx, "删除设备测试完成", map[string]interface{}{
		"action":    "test_delete_device_complete",
		"device_id": testDevice.ID,
	})
}

func (suite *DeviceServiceTestSuite) TestListDevices() {
	logger.Info(suite.ctx, "开始执行列出设备测试", map[string]interface{}{
		"action": "test_list_devices",
	})
	
	for i := 0; i < 5; i++ {
		suite.createTestDevice()
	}
	devices, total, err := suite.service.ListDevices(suite.ctx, 1, 10, map[string]interface{}{})
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(5))
	suite.Len(devices, int(total)) // 或 >=5
	
	logger.Info(suite.ctx, "列出设备测试完成", map[string]interface{}{
		"action":     "test_list_devices_complete",
		"device_num": total,
	})
}

func (suite *DeviceServiceTestSuite) TestUpdateDeviceNetwork() {
	logger.Info(suite.ctx, "开始执行更新设备网络测试", map[string]interface{}{
		"action": "test_update_device_network",
	})
	
	testDevice := suite.createTestDevice()

	mac := randomMAC()
	logger.Info(suite.ctx, "使用随机MAC地址创建网络信息", map[string]interface{}{
		"action":      "create_network_info",
		"mac_address": mac,
	})
	
	networkInfo := &device.DeviceNetwork{
		DeviceID:      testDevice.ID,
		NetworkType:   "wifi",
		MacAddress:    mac,
		ConnectStatus: "connected",
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}
	err := suite.db.Create(networkInfo).Error
	suite.NoError(err)

	updates := map[string]interface{}{
		"ip_address":     "192.168.1.100",
		"connect_status": "disconnected",
	}
	err = suite.service.UpdateDeviceNetwork(suite.ctx, testDevice.ID, updates)
	suite.NoError(err)

	updatedNetwork, err := suite.service.GetDeviceNetwork(suite.ctx, testDevice.ID)
	suite.NoError(err)
	suite.NotNil(updatedNetwork)
	suite.Equal("192.168.1.100", updatedNetwork.IPAddress)
	suite.Equal("disconnected", updatedNetwork.ConnectStatus)
	
	logger.Info(suite.ctx, "更新设备网络测试完成", map[string]interface{}{
		"action":    "test_update_device_network_complete",
		"device_id": testDevice.ID,
	})
}

func (suite *DeviceServiceTestSuite) TestGetDeviceNetwork() {
	logger.Info(suite.ctx, "开始执行获取设备网络测试", map[string]interface{}{
		"action": "test_get_device_network",
	})
	
	testDevice := suite.createTestDevice()
	networkInfo := &device.DeviceNetwork{
		DeviceID:      testDevice.ID,
		NetworkType:   "wifi",
		MacAddress:    randomMAC(), // 使用随机MAC地址避免冲突
		IPAddress:     "192.168.1.100",
		ConnectStatus: "connected",
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}
	err := suite.db.Create(networkInfo).Error
	suite.NoError(err)

	retrievedNetwork, err := suite.service.GetDeviceNetwork(suite.ctx, testDevice.ID)
	suite.NoError(err)
	suite.NotNil(retrievedNetwork)
	suite.Equal("192.168.1.100", retrievedNetwork.IPAddress)
	
	logger.Info(suite.ctx, "获取设备网络测试完成", map[string]interface{}{
		"action":    "test_get_device_network_complete",
		"device_id": testDevice.ID,
	})
}

func (suite *DeviceServiceTestSuite) TestUpdateDeviceStatus() {
	logger.Info(suite.ctx, "开始执行更新设备状态测试", map[string]interface{}{
		"action": "test_update_device_status",
	})
	
	testDevice := suite.createTestDevice()
	statusInfo := &device.DeviceStatus{
		DeviceID:      testDevice.ID,
		BatteryLevel:  80,
		PowerStatus:   "power_on",
		WorkingStatus: "idle",
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}
	err := suite.db.Create(statusInfo).Error
	suite.NoError(err)

	updates := map[string]interface{}{
		"battery_level":  90,
		"working_status": "idle", // 修改为有效的枚举值
	}
	err = suite.service.UpdateDeviceStatus(suite.ctx, testDevice.ID, updates)
	suite.NoError(err)

	updatedStatus, err := suite.service.GetDeviceStatus(suite.ctx, testDevice.ID)
	suite.NoError(err)
	suite.Equal(90, updatedStatus.BatteryLevel)
	suite.Equal("idle", updatedStatus.WorkingStatus) // 验证为有效的枚举值
	
	logger.Info(suite.ctx, "更新设备状态测试完成", map[string]interface{}{
		"action":    "test_update_device_status_complete",
		"device_id": testDevice.ID,
	})
}

func (suite *DeviceServiceTestSuite) TestGetDeviceStatus() {
	logger.Info(suite.ctx, "开始执行获取设备状态测试", map[string]interface{}{
		"action": "test_get_device_status",
	})
	
	testDevice := suite.createTestDevice()
	statusInfo := &device.DeviceStatus{
		DeviceID:      testDevice.ID,
		BatteryLevel:  85,
		PowerStatus:   "power_on",
		WorkingStatus: "idle",
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}
	err := suite.db.Create(statusInfo).Error
	suite.NoError(err)

	retrievedStatus, err := suite.service.GetDeviceStatus(suite.ctx, testDevice.ID)
	suite.NoError(err)
	suite.NotNil(retrievedStatus)
	suite.Equal(85, retrievedStatus.BatteryLevel)
	
	logger.Info(suite.ctx, "获取设备状态测试完成", map[string]interface{}{
		"action":    "test_get_device_status_complete",
		"device_id": testDevice.ID,
	})
}

func (suite *DeviceServiceTestSuite) TestActivateDevice() {
	logger.Info(suite.ctx, "开始执行激活设备测试", map[string]interface{}{
		"action": "test_activate_device",
	})
	
	testDevice := suite.createTestDevice()
	activationTime := time.Now()
	err := suite.service.ActivateDevice(suite.ctx, testDevice.ID, activationTime)
	suite.NoError(err)

	activatedDevice, err := suite.service.GetDeviceByID(suite.ctx, testDevice.ID)
	suite.NoError(err)
	suite.Equal("activated", activatedDevice.Status)
	suite.WithinDuration(activationTime, activatedDevice.ActivationTime, time.Second)
	
	logger.Info(suite.ctx, "激活设备测试完成", map[string]interface{}{
		"action":    "test_activate_device_complete",
		"device_id": testDevice.ID,
	})
}

// ==================== TestMain ====================

func TestMain(t *testing.T) {
	defer logger.DefaultLogger.Sync()
	projectRoot := tools.GetRootDir()
	os.Chdir(projectRoot)

	// 注意：确保 config 已初始化（如果你的 postgre.GetClient() 依赖 config）
	// 如果尚未初始化，取消下面注释：
	// if err := config.Init(); err != nil {
	//     t.Fatalf("Failed to initialize config: %v", err)
	// }

	suite.Run(t, new(DeviceServiceTestSuite))
}

// ==================== 基准测试（可选保留）====================

// BenchmarkGetDeviceByID 基准测试：根据ID获取设备
func BenchmarkGetDeviceByID(b *testing.B) {
	// 设置工作目录为项目根目录
	projectRoot := tools.GetRootDir()
	os.Chdir(projectRoot)
	
	// 创建带trace_id的上下文
	ctx := context.Background()
	
	// 记录基准测试开始日志
	logger.Info(ctx, "开始执行根据ID获取设备基准测试", map[string]interface{}{
		"action": "benchmark_get_device_by_id",
	})

	// 初始化服务
	client := postgre.GetClient()
	dbProvider := &PostgreClient{Client: client}
	service := NewService(dbProvider)
	
	// 创建测试设备用于基准测试
	deviceSN := "bench_get_" + uuid.New().String()
	baseDevice := &device.BaseDevice{
		DeviceSN:        deviceSN,
		DeviceType:      "sensor",
		VendorID:        1,
		VendorName:      "BenchmarkVendor",
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0",
		ProductModel:    "BenchmarkModel",
		ManufactureDate: time.Now(),
		Status:          "inactivated",
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}
	
	err := service.RegisterDevice(ctx, baseDevice)
	assert.NoError(b, err)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetDeviceByID(ctx, baseDevice.ID)
		assert.NoError(b, err)
	}
	
	// 记录基准测试完成日志
	logger.Info(ctx, "根据ID获取设备基准测试完成", map[string]interface{}{
		"action": "benchmark_get_device_by_id_complete",
		"rounds": b.N,
	})
}

// BenchmarkUpdateDevice 基准测试：更新设备
func BenchmarkUpdateDevice(b *testing.B) {
	// 设置工作目录为项目根目录
	projectRoot := tools.GetRootDir()
	os.Chdir(projectRoot)
	
	// 创建带trace_id的上下文
	ctx := context.Background()
	
	// 记录基准测试开始日志
	logger.Info(ctx, "开始执行更新设备基准测试", map[string]interface{}{
		"action": "benchmark_update_device",
	})
	
	// 初始化服务
	client := postgre.GetClient()
	dbProvider := &PostgreClient{Client: client}
	service := NewService(dbProvider)
	
	// 创建测试设备用于基准测试
	deviceSN := "bench_update_" + uuid.New().String()
	baseDevice := &device.BaseDevice{
		DeviceSN:        deviceSN,
		DeviceType:      "sensor",
		VendorID:        1,
		VendorName:      "BenchmarkVendor",
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0",
		ProductModel:    "BenchmarkModel",
		ManufactureDate: time.Now(),
		Status:          "inactivated",
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}
	
	err := service.RegisterDevice(ctx, baseDevice)
	assert.NoError(b, err)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updates := map[string]interface{}{
			"device_type": "updated_sensor",
			"vendor_name": "UpdatedVendor",
		}
		
		err := service.UpdateDevice(ctx, baseDevice.ID, updates)
		assert.NoError(b, err)
	}
	
	// 记录基准测试完成日志
	logger.Info(ctx, "更新设备基准测试完成", map[string]interface{}{
		"action": "benchmark_update_device_complete",
		"rounds": b.N,
	})
}
