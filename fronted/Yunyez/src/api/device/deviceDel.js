//删除设备
import request from '@/utils/request'

export function delDevice(data) {
  return request({
    url: '/device/delete',
    method: 'post',
    params:data
  })
}
