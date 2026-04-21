import axios from 'axios'
import { ERROR_CODES } from './codes'

// 开发环境用 '/' 触发代理，生产环境用真实 API 地址
const baseURL = import.meta.env.DEV
  ? '/api'
  : import.meta.env.VITE_API_BASE_URL

// 1. 创建axios实例
const request = axios.create({
  baseURL,
  timeout: 5000,
  withCredentials: true, // 让浏览器自动携带 cookie（refresh token 通常放在 httpOnly cookie 中）
  headers: {
    'Content-Type': 'application/json;charset=utf-8'
  }
})

// ========== 错误码分组 ==========
const ACCESS_TOKEN_EXPIRED = [ERROR_CODES.TOKEN_EXPIRED]
const ACCESS_TOKEN_INVALID = [
  ERROR_CODES.TOKEN_INVALID,
  ERROR_CODES.TOKEN_REVOKED,
]
const REFRESH_TOKEN_INVALID = [
  ERROR_CODES.REFRESH_TOKEN_INVALID,
  ERROR_CODES.REFRESH_TOKEN_EXPIRED,
  ERROR_CODES.REFRESH_TOKEN_REVOKED
]

// 获取 access token
const getToken = () => localStorage.getItem('token')
const setToken = (token) => localStorage.setItem('token', token)
const removeToken = () => localStorage.removeItem('token')

// 刷新 token 的状态控制
let isRefreshing = false
let failedQueue = [] // 存储 { resolve, reject, config }

// 处理队列：刷新成功后重试所有请求，失败则全部 reject
const processQueue = (error, newToken = null) => {
  failedQueue.forEach(({ resolve, reject, config }) => {
    if (error) {
      reject(error)
    } else {
      // 用新 token 更新请求头
      config.headers.Authorization = `Bearer ${newToken}`
      // 重新发起请求，并将结果 resolve 出去
      resolve(request(config))
    }
  })
  failedQueue = []
}

// 刷新 token 的函数（使用原生 axios 实例，避免被拦截器循环拦截）
const refreshTokenRequest = () => {
  return axios({
    method: 'GET',
    url: `${baseURL}/login/refresh`, // 或者直接 '/login/refresh' 使用代理
    withCredentials: true, // 携带 cookie
  })
}

// ========== 请求拦截器：添加 accessToken ==========
request.interceptors.request.use(
  config => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// ========== 响应拦截器：处理 token 过期 ==========
request.interceptors.response.use(
  response => {
    const { code } = response.data
    // 如果业务码表示 access token 过期，则进入刷新流程（注意：这里返回的是正常响应，但需要重试）
    if (ACCESS_TOKEN_EXPIRED.includes(code)) {
      // 将原始请求的配置保存下来
      const originalConfig = response.config

      if (!isRefreshing) {
        // 开始刷新 token
        isRefreshing = true

        return refreshTokenRequest()
          .then(refreshRes => {
            const refreshCode = refreshRes.data.code
            if (refreshCode === ERROR_CODES.SUCCESS) {
              const newToken = refreshRes.data.data.accessToken
              setToken(newToken)
              // 处理队列中的请求
              processQueue(null, newToken)
              // 重试当前请求（第一个触发刷新的请求）
              originalConfig.headers.Authorization = `Bearer ${newToken}`
              return request(originalConfig)
            } else if (REFRESH_TOKEN_INVALID.includes(refreshCode)) {
              // refresh token 失效，清空 token，跳转登录
              removeToken()
              processQueue(new Error('Refresh token expired'), null)
              window.location.href = '/'
              return Promise.reject(new Error('请重新登录'))
            } else {
              // 其他刷新失败
              const error = new Error('Refresh token failed')
              processQueue(error, null)
              return Promise.reject(error)
            }
          })
          .catch(err => {
            // 刷新请求本身失败（网络错误、500等）
            removeToken()
            processQueue(err, null)
            window.location.href = '/'
            return Promise.reject(err)
          })
          .finally(() => {
            isRefreshing = false
          })
      } else {
        // 正在刷新中，将当前请求加入队列
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject, config: originalConfig })
        })
      }
    }

    // 其他情况：直接返回响应
    return response
  },
  error => {
    // 处理 HTTP 错误（如 401 Unauthorized）
    // 如果你的后端在 token 过期时返回 HTTP 401，可以在这里处理
    const { response } = error
    if (response && response.status === 401) {
      // 可以复用上面的逻辑，或者统一跳转登录
      removeToken()
      window.location.href = '/'
    }
    return Promise.reject(error)
  }
)

export default request