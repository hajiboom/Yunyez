// src/api/device.js
import request from '../index'

/**
 * 查询设备列表（适配你的查询参数）
 * @param {Object} params - 查询参数
 * @param {number} params.pageNum - 页码
 * @param {number} params.pageSize - 每页条数
 * @param {string} params.deviceSn - 设备序列号
 * @param {string} params.vendorName - 厂商名称
 * @param {string} params.status - 设备状态
 * @param {string} params.startTime - 开始时间
 * @param {string} params.endTime - 结束时间
 * @returns {Promise}
 */
export function getDeviceListApi(params) {
  return request({
    url: '/device/fetch', // 后端设备列表接口路径（替换为真实路径）
    method: 'get',
    params 
  })
}

// 其他设备接口（新增/编辑/删除）也放这里，示例：
export function addDeviceApi(data) {
  return request({
    url: '/device/add',
    method: 'post',
    data // post请求用data传参
  })
}

export function deleteDeviceApi(id) {
  return request({
    url: `/device/${id}`,
    method: 'delete'
  })
}