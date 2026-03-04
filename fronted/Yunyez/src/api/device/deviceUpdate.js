// 更新设备
import request from '@/utils/request'

export function updateDevice(data) {
  return request({
    url: '/device/update',
    method: 'put', 
    data: data 
  })
}