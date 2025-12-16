package device

import "time"

// DeviceListItem 设备列表项（精简版）
type DeviceListItem struct {
	SN           string    `json:"sn"`
	DeviceType   string    `json:"deviceType"`
	VendorName   string    `json:"vendorName"`
	ProductModel string    `json:"productModel"`
	Status       string    `json:"status"`
	CreateTime   time.Time `json:"createTime"`
}

// DeviceDetail 设备详情（完整版）
type DeviceDetail struct {
	// 基础信息
	*DeviceBaseInfo `json:",inline"`
	// 网络信息
	*DeviceNetworkInfo `json:",inline,omitempty"`
}

// DeviceBaseInfo 设备基础信息
type DeviceBaseInfo struct {
	SN              string     `json:"sn"`
	IMEI            string     `json:"imei,omitempty"`
	ICCID           string     `json:"iccid,omitempty"`
	DeviceType      string     `json:"deviceType"`
	VendorName      string     `json:"vendorName"`
	HardwareVersion string     `json:"hardwareVersion"`
	FirmwareVersion string     `json:"firmwareVersion"`
	ProductModel    string     `json:"productModel"`
	ManufactureDate time.Time  `json:"manufactureDate"`
	ExpireDate      *time.Time `json:"expireDate,omitempty"`
	Status          string     `json:"status"`
	ActivationTime  *time.Time `json:"activationTime,omitempty"`
	Remark          string     `json:"remark,omitempty"`
}

// DeviceNetworkInfo 设备网络信息
type DeviceNetworkInfo struct {
	// SN                 string     `json:"sn"`
	NetworkType        string     `json:"networkType"`
	MacAddress         string     `json:"macAddress,omitempty"`
	IPAddress          string     `json:"ipAddress,omitempty"`
	Port               int        `json:"port,omitempty"`
	SignalStrength     int        `json:"signalStrength,omitempty"`
	ConnectStatus      string     `json:"connectStatus"`
	LastConnectTime    *time.Time `json:"lastConnectTime,omitempty"`
	LastDisconnectTime *time.Time `json:"lastDisconnectTime,omitempty"`
}


// SafeTimePointer 安全地将 time.Time 转换为指针
// 参数：
//   - t time.Time 时间值
// 返回：
//   - *time.Time 指向时间值的指针
func SafeTimePointer(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}	