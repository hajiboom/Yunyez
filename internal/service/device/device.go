package device

import (
	"context"
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"

	"yunyez/internal/model/device"
	"yunyez/pkg/postgre"
)

// Service defines the device business logic interface.
type Service interface {
	RegisterDevice(ctx context.Context, baseDevice *device.BaseDevice) error
	GetDeviceByID(ctx context.Context, id int64) (*device.BaseDevice, error)
	GetDeviceBySN(ctx context.Context, sn string) (*device.BaseDevice, error)
	UpdateDevice(ctx context.Context, id int64, updates map[string]interface{}) error
	DeleteDevice(ctx context.Context, id int64) error
	ListDevices(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*device.BaseDevice, int64, error)
	UpdateDeviceNetwork(ctx context.Context, deviceID int64, updates map[string]interface{}) error
	GetDeviceNetwork(ctx context.Context, deviceID int64) (*device.DeviceNetwork, error)
	UpdateDeviceStatus(ctx context.Context, deviceID int64, updates map[string]interface{}) error
	GetDeviceStatus(ctx context.Context, deviceID int64) (*device.DeviceStatus, error)
	ActivateDevice(ctx context.Context, deviceID int64, activationTime time.Time) error
}

// DBProvider abstracts database access.
type DBProvider interface {
	DB() *gorm.DB
}

// postgreClient implements DBProvider using PostgreSQL.
type postgreClient struct {
	client *postgre.Client
}

func (p *postgreClient) DB() *gorm.DB {
	return p.client.DB
}

var (
	once            sync.Once
	serviceInstance *service
)

// NewService returns a singleton instance of the device service.
// It supports dependency injection via DBProvider for testability.
func NewService(dbProvider DBProvider) Service {
	once.Do(func() {
		serviceInstance = &service{provider: dbProvider}
	})
	return serviceInstance
}

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

// GetDeviceByID retrieves a device by its ID.
func (s *service) GetDeviceByID(ctx context.Context, id int64) (*device.BaseDevice, error) {
	if id <= 0 {
		return nil, errors.New("invalid device ID")
	}
	var baseDevice device.BaseDevice
	err := s.provider.DB().WithContext(ctx).First(&baseDevice, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &baseDevice, nil
}

// GetDeviceBySN retrieves a device by its serial number (DeviceSN).
func (s *service) GetDeviceBySN(ctx context.Context, sn string) (*device.BaseDevice, error) {
	if sn == "" {
		return nil, errors.New("empty serial number")
	}
	var baseDevice device.BaseDevice
	// Use explicit column name that matches GORM tag: `gorm:"column:device_sn"`
	err := s.provider.DB().WithContext(ctx).Where(&device.BaseDevice{DeviceSN: sn}).First(&baseDevice).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &baseDevice, nil
}

// UpdateDevice performs partial update on a device by ID.
func (s *service) UpdateDevice(ctx context.Context, id int64, updates map[string]interface{}) error {
	if id <= 0 {
		return errors.New("invalid device ID")
	}
	if len(updates) == 0 {
		return nil // nothing to update
	}
	result := s.provider.DB().WithContext(ctx).Model(&device.BaseDevice{}).Where(&device.BaseDevice{ID: id}).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteDevice soft-deletes a device by ID.
func (s *service) DeleteDevice(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid device ID")
	}
	result := s.provider.DB().WithContext(ctx).Delete(&device.BaseDevice{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListDevices returns a paginated list of devices with optional filters.
// Note: keys in `filters` must be safe column names (e.g., "status", "device_sn").
// For production, consider allow-list validation to prevent injection.
func (s *service) ListDevices(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*device.BaseDevice, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	db := s.provider.DB().WithContext(ctx).Model(&device.BaseDevice{})
	for key, value := range filters {
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

	return devices, total, nil
}

// UpdateDeviceNetwork updates network info for a device.
func (s *service) UpdateDeviceNetwork(ctx context.Context, deviceID int64, updates map[string]interface{}) error {
	if deviceID <= 0 {
		return errors.New("invalid device ID")
	}
	if len(updates) == 0 {
		return nil
	}
	result := s.provider.DB().WithContext(ctx).Model(&device.DeviceNetwork{}).Where(&device.DeviceNetwork{DeviceID: deviceID}).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetDeviceNetwork retrieves network info by device ID.
func (s *service) GetDeviceNetwork(ctx context.Context, deviceID int64) (*device.DeviceNetwork, error) {
	if deviceID <= 0 {
		return nil, errors.New("invalid device ID")
	}
	var net device.DeviceNetwork
	err := s.provider.DB().WithContext(ctx).Where("device_id = ?", deviceID).First(&net).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &net, nil
}

// UpdateDeviceStatus updates device status fields.
func (s *service) UpdateDeviceStatus(ctx context.Context, deviceID int64, updates map[string]interface{}) error {
	if deviceID <= 0 {
		return errors.New("invalid device ID")
	}
	if len(updates) == 0 {
		return nil
	}

	// Prevent updating protected fields
	protected := []string{"ID", "DeviceID", "CreateTime"}
	for _, field := range protected {
		delete(updates, field)
	}
	updates["update_time"] = time.Now()

	result := s.provider.DB().WithContext(ctx).Model(&device.DeviceStatus{}).Where(&device.DeviceStatus{DeviceID: deviceID}).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetDeviceStatus retrieves status info by device ID.
func (s *service) GetDeviceStatus(ctx context.Context, deviceID int64) (*device.DeviceStatus, error) {
	if deviceID <= 0 {
		return nil, errors.New("invalid device ID")
	}
	var status device.DeviceStatus
	err := s.provider.DB().WithContext(ctx).Where(&device.DeviceStatus{DeviceID: deviceID}).First(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &status, nil
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