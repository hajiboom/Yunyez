//查询设备列表
import request from '@/utils/request'
import service from '@/mock/requestMock'
export function getDeviceList(data) {
  return service.get('/device/fetch', data)
}
