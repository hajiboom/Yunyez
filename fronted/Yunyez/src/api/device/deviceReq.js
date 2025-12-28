//查询设备列表
import request from '@/utils/request'

export function getDeviceList(data) {
  return request({
    url: '/device/fetch',
    method: 'get',
    params: data
  })
}