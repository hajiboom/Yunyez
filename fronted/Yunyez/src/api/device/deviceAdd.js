//添加设备
import request from '@/utils/request'

export function addDevice(data) {
  return request({
    url: '/device/add',
    method: 'post',
    params: data
  })
}