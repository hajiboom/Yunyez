// api/auth.js
import request from '@/utils/request'

/**
 * 登录 - 使用 HttpOnly Cookie
 * @param {Object} data - { username, password, rememberMe }
 * @returns {Promise} - 返回用户信息，Token 由 Cookie 自动处理
 */
export function login(data) {
  return request({
    url: '/login',
    method: 'post',
    data: data,                    // 改为 data，params 用于 GET
  })
}

/**
 * 获取当前登录用户信息（用于页面刷新后恢复状态）
 */
export function getUserInfo() {
  return request({
    url: '/auth/info',
    method: 'get',
    
  })
}
/**
 * 登出
 */
export function logout() {
  return request({
    url: '/auth/logout',
    method: 'post',
  })
}