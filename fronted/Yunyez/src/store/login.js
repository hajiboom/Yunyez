import { defineStore } from 'pinia'
import { login } from '@/api/login/login.js'
import { ERROR_CODES, ERROR_MESSAGES } from '@/utils/codes'
import { ElMessage } from 'element-plus'


export const useLoginStore = defineStore('login', {
  state: () => ({
   
  }),
  actions: {
    async login(data) {
      const res = await login(data)
      if (res.code === ERROR_CODES.SUCCESS) {
        // 登录成功，将token存储到localStorage
        localStorage.setItem('token', res.data.accessToken)
        ElMessage.success('登录成功')
      } 
    },
    async logout() {
      const res = await logout()
      const authStore = useAuthStore()
      if (res.code === ERROR_CODES.SUCCESS) {
        // 登出成功，清除localStorage中的token
        localStorage.removeItem('token')
        ElMessage.success('登出成功')
        window.location.href = '/'
      }
    }
  }
})