// src/service/request.js
import axios from 'axios'


// 开发环境用 '/' 触发代理，生产环境用真实 API 地址
const baseURL = import.meta.env.DEV
  ? '/api' // 开发时走 Vite 代理
  : import.meta.env.VITE_API_BASE_URL ;// 生产时用真实地址

// 1. 创建axios实例
const request = axios.create({
  // 开发阶段：使用代理（vite.config.js 中配置）
  // 防止跨域问题（生产环境：配置Nginx反向代理）
  baseURL,
  timeout: 5000, // 超时时间
  headers: {
    'Content-Type': 'application/json;charset=utf-8'
  }
})

// 2. 请求拦截器（发请求前做的事：加token、加loading等）

request.interceptors.request.use(
  (config) => {


    return config
  },
  (error) => {
   
    return Promise.reject(error)
  }
)

// 3. 响应拦截器（接收到响应后做的事：统一解析、错误处理）
request.interceptors.response.use(
  (response) => {
    const res = response.data
    // 接口返回的code非200时，统一提示错误（根据后端约定调整）
    if (res.Code !== 200) {
      return Promise.reject(res)
    }
    return res // 只返回后端的data部分，内容实际是一个封装{Code, Data, Message}
  },
  (error) => {

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
export default request
