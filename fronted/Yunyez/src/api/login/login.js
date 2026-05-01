// api/auth.js
import request from '@/utils/request'
import service from "@/mock/requestMock"
/**
 * 登录 - 使用 HttpOnly Cookie
 * @param {Object} data - { username, password, rememberMe }
 * @returns {Promise} - 返回用户信息，Token 由 Cookie 自动处理
 */
export function login(data) {
  return request({
    url: '/login',
    method: 'post',
    data: data,                    
  })
}


export function getUserInfo(id) {
  return service.get('/auth/info/', { params: { id } })
}

export function getUserMenusByRoleId(id) {
  return service.get('/auth/menus/', { params: { id } })
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
/**
 * 修改密码
 */

export function fixPassword(data) {
  return request({
    url: '/auth/fixPassword',
    method: 'post',
    data: data,
  })
}
