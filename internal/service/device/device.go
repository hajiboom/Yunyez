package device

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"yunyez/internal/common/constant"
	"yunyez/internal/model/device"
	"yunyez/internal/pkg/postgre"
	deviceType "yunyez/internal/types/device"
)

var (
	once            sync.Once
	ServiceInstance *service
)

func init() {
	// Initialize the service instance with a default DBProvider.
	ServiceInstance = &service{provider: &PostgreClient{Client: postgre.GetClient()}}
}

// Service defines the device business logic interface.
type Service interface {
	RegisterDevice(ctx context.Context, baseDevice *device.BaseDevice) error
	// 查询设备是否存在
	CheckDeviceExist(ctx context.Context, sn string) (bool, error)
	// 根据序列号查询设备基础信息
	GetDeviceBySN(ctx context.Context, sn string) (*deviceType.DeviceBaseInfo, error)
	UpdateDevice(ctx context.Context, sn string, updates map[string]interface{}) error
	DeleteDevice(ctx context.Context, sn string) error
	// 查询设备列表
	ListDevices(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*deviceType.DeviceListItem, int64, error)
	UpdateDeviceNetwork(ctx context.Context, deviceSN string, updates map[string]interface{}) error
	// 根据序列号查询设备网络信息
	GetDeviceNetworkBySN(ctx context.Context, deviceSN string) (*deviceType.DeviceNetworkInfo, error)
	ActivateDevice(ctx context.Context, deviceID int64, activationTime time.Time) error
}

// DBProvider abstracts database access.
type DBProvider interface {
	DB() *gorm.DB
}

// PostgreClient implements DBProvider using PostgreSQL.
type PostgreClient struct {
	Client *postgre.Client
}

func (p *PostgreClient) DB() *gorm.DB {
	return p.Client.DB
}

// NewService returns a singleton instance of the device service.
// It supports dependency injection via DBProvider for testability.
func NewService(dbProvider DBProvider) Service {
	once.Do(func() {
		ServiceInstance = &service{provider: dbProvider}
	})
	return ServiceInstance
}

// service implements the Service interface.
type service struct {
	provider DBProvider
}

// RegisterDevice creates a new device record.
func (s *service) RegisterDevice(ctx context.Context, baseDevice *device.BaseDevice) error {
	if baseDevice == nil {
		return errors.New("baseDevice is nil")
	}
	return s.provider.DB().WithContext(ctx).Create(baseDevice).Error
}

// CheckDeviceExist 根据序列号查询设备是否存在
// 参数：
//   - ctx context.Context 上下文
//   - sn string 设备序列号
//
// 返回：
//   - bool 设备是否存在
//   - error 查询失败时返回错误
func (s *service) CheckDeviceExist(ctx context.Context, sn string) (bool, error) {
	if sn == "" {
		return false, errors.New("empty serial number")
	}
	var count int64
	err := s.provider.DB().WithContext(ctx).Model(&device.BaseDevice{}).
		Where(&device.BaseDevice{SN: sn}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetDeviceBySN 根据序列号查询设备基础信息
// 参数：
//   - ctx context.Context 上下文
//   - sn string 设备序列号
//
// 返回：
//   - *deviceType.DeviceBaseInfo 设备基础信息
//   - error 查询失败时返回错误
func (s *service) GetDeviceBySN(ctx context.Context, sn string) (*deviceType.DeviceBaseInfo, error) {
	if sn == "" {
		return nil, errors.New("empty serial number")
	}
	// 设备基础信息
	var baseDevice device.BaseDevice
	// Use explicit column name that matches GORM tag: `gorm:"column:device_sn"`
	err := s.provider.DB().WithContext(ctx).Where(&device.BaseDevice{SN: sn}).First(&baseDevice).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("device_base not found by device_sn: %s", sn)
		}
		return nil, err
	}

	baseInfo := &deviceType.DeviceBaseInfo{
		SN:              baseDevice.SN,
		IMEI:            baseDevice.IMEI,
		ICCID:           baseDevice.ICCID,
		DeviceType:      baseDevice.DeviceType,
		VendorName:      baseDevice.VendorName,
		HardwareVersion: baseDevice.HardwareVersion,
		FirmwareVersion: baseDevice.FirmwareVersion,
		ProductModel:    baseDevice.ProductModel,
		ManufactureDate: baseDevice.ManufactureDate,
		ExpireDate:      deviceType.SafeTimePointer(baseDevice.ExpireDate), // 指针类型，可能为 nil
		Status:          baseDevice.Status,
		ActivationTime:  deviceType.SafeTimePointer(baseDevice.ActivationTime), // 指针类型，可能为 nil
		Remark:          baseDevice.Remark,
	}
	return baseInfo, nil
}

// UpdateDevice 更新设备基础信息
// 参数：
//   - ctx context.Context 上下文
//   - sn string 设备序列号
//   - updates map[string]interface{} 更新字段
//
// 返回：
//   - error 更新失败时返回错误
func (s *service) UpdateDevice(ctx context.Context, sn string, updates map[string]interface{}) error {
	if sn == "" {
		return fmt.Errorf("[%d]empty serial number", constant.ErrInvalidParam)
	}
	if len(updates) == 0 {
		return fmt.Errorf("[%d]no fields to update", constant.ErrInvalidParam)
	}
	result := s.provider.DB().WithContext(ctx).Model(&device.BaseDevice{}).Where(&device.BaseDevice{SN: sn}).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteDevice 删除设备（软删除）
// 参数：
//   - ctx context.Context 上下文
//   - sn string 设备序列号
//
// 返回：
//   - error 删除失败时返回错误
func (s *service) DeleteDevice(ctx context.Context, sn string) error {
	if sn == "" {
		return errors.New("empty serial number")
	}
	db := s.provider.DB().WithContext(ctx)

	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除设备基本信息
		record := tx.Where(&device.BaseDevice{SN: sn}).Delete(&device.BaseDevice{})
		if record.Error != nil {
			return record.Error
		}
		if record.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// 删除设备网络信息（网络信息可能不存在）
		if err := tx.Where(&device.DeviceNetwork{SN: sn}).Delete(&device.DeviceNetwork{}).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

// ListDevices returns a paginated list of devices with optional filters.
// Note: keys in `filters` must be safe column names (e.g., "status", "device_sn").
// For production, consider allow-list validation to prevent injection.
func (s *service) ListDevices(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*deviceType.DeviceListItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	db := s.provider.DB().WithContext(ctx).Model(&device.BaseDevice{})
	for key, value := range filters {
		if value == nil {
			continue
		}
		// ⚠️ In production, validate `key` against an allow-list of columns
		db = db.Where(key+" = ?", value)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var devices []*device.BaseDevice
	if err := db.Offset(offset).Limit(pageSize).Find(&devices).Error; err != nil {
		return nil, 0, err
	}

	deviceListItems := make([]*deviceType.DeviceListItem, 0, len(devices))
	for _, dev := range devices {
		deviceListItems = append(deviceListItems, &deviceType.DeviceListItem{
			SN:           dev.SN,
			DeviceType:   dev.DeviceType,
			VendorName:   dev.VendorName,
			Status:       dev.Status,
			CreateTime:   dev.CreateTime,
			ProductModel: dev.ProductModel,
		})
	}

	return deviceListItems, total, nil
}

// UpdateDeviceNetwork updates network info for a device.
func (s *service) UpdateDeviceNetwork(ctx context.Context, deviceSN string, updates map[string]interface{}) error {
	if deviceSN == "" {
		return errors.New("empty device serial number")
	}
	if len(updates) == 0 {
		return nil
	}
	result := s.provider.DB().WithContext(ctx).Model(&device.DeviceNetwork{}).Where(&device.DeviceNetwork{SN: deviceSN}).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetDeviceNetworkBySN 根据序列号查询设备网络信息
// 参数：
//   - ctx context.Context 上下文
//   - deviceSN string 设备序列号
//
// 返回：
//   - *deviceType.DeviceNetworkInfo 设备网络信息
//   - error 查询失败时返回错误
func (s *service) GetDeviceNetworkBySN(ctx context.Context, deviceSN string) (*deviceType.DeviceNetworkInfo, error) {
	if deviceSN == "" {
		return nil, errors.New("empty device serial number")
	}
	var deviceNetwork device.DeviceNetwork
	err := s.provider.DB().WithContext(ctx).Where("device_sn = ?", deviceSN).First(&deviceNetwork).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	networkInfo := &deviceType.DeviceNetworkInfo{
		NetworkType:        deviceNetwork.NetworkType,
		MacAddress:         deviceNetwork.MacAddress,
		IPAddress:          deviceNetwork.IPAddress,
		Port:               deviceNetwork.Port,
		SignalStrength:     deviceNetwork.SignalStrength,
		ConnectStatus:      deviceNetwork.ConnectStatus,
		LastConnectTime:    deviceType.SafeTimePointer(deviceNetwork.LastConnectTime),    // 指针类型，可能为 nil
		LastDisconnectTime: deviceType.SafeTimePointer(deviceNetwork.LastDisconnectTime), // 指针类型，可能为 nil
	}

	return networkInfo, nil
}

// ActivateDevice sets the device to activated state.
func (s *service) ActivateDevice(ctx context.Context, deviceID int64, activationTime time.Time) error {
	if deviceID <= 0 {
		return errors.New("invalid device ID")
	}
	updates := map[string]interface{}{
		"status":          "activated",
		"activation_time": activationTime,
		"update_time":     time.Now(),
	}
	result := s.provider.DB().WithContext(ctx).Model(&device.BaseDevice{}).Where("id = ?", deviceID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
