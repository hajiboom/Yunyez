// src/utils/request.js
import axios from 'axios'
import { ElMessage, ElLoading } from 'element-plus'


// ✅ 智能判断：开发环境用 '/' 触发代理，生产环境用真实 API 地址
const baseURL = import.meta.env.DEV
  ? '/api' // 开发时走 Vite 代理
  : import.meta.env.VITE_API_BASE_URL // 生产时用真实地址

// 1. 创建axios实例
const service = axios.create({
  // 开发阶段：使用代理（vite.config.js 中配置）
  // 防止跨域问题（生产环境：配置Nginx反向代理）
  baseURL,
  timeout: 10000, // 超时时间
  headers: {
    'Content-Type': 'application/json;charset=utf-8'
  }
})



// 2. 请求拦截器（发请求前做的事：加token、加loading等）
let loadingInstance = null
service.interceptors.request.use(
  (config) => {
    // 可选：添加加载动画（全局）
    loadingInstance = ElLoading.service({
      lock: true,
      text: '加载中...',
      background: 'rgba(0,0,0,0.1)'
    })

    // 可选：添加token（登录后存在localStorage/ Pinia中）
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    loadingInstance?.close() // 失败关闭loading
    ElMessage.error('请求异常：' + error.message)
    return Promise.reject(error)
  }
)

// 3. 响应拦截器（接收到响应后做的事：统一解析、错误处理）
service.interceptors.response.use(
  (response) => {
    loadingInstance?.close() // 成功关闭loading
    const res = response.data

    // 接口返回的code非200时，统一提示错误（根据后端约定调整）
    if (res.Code !== 200) {
      ElMessage.error(res.Message || '请求失败')
      return Promise.reject(res)
    }
    return res // 只返回后端的data部分，内容实际是一个封装{Code, Data, Message}
  },
  (error) => {
    loadingInstance?.close() // 失败关闭loading
    // 统一处理网络错误/401/403/500等
    let msg = ''
    if (error.response) {
      switch (error.response.status) {
        case 401:
          msg = '登录过期，请重新登录'
          // 可选：跳登录页
          // router.push('/login')
          break
        case 403:
          msg = '暂无权限访问'
          break
        case 500:
          msg = '服务器内部错误'
          break
        default:
          msg = '请求失败：' + error.response.data.msg
      }
    } else {
      msg = '网络异常，请检查网络'
    }
    ElMessage.error(msg)
    return Promise.reject(error)
  }
)

// 4. 导出封装好的axios实例
export default service