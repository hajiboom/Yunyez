import { defineStore } from 'pinia'
import { login,logout,fixPassword,getUserInfo,getUserMenusByRoleId } from '@/api/login/login.js'
import { ERROR_CODES, ERROR_MESSAGES } from '@/utils/codes'
import { ElMessage } from 'element-plus'


export const useLoginStore = defineStore('login', {
  state: () => ({
   userMenus: [], 
    userInfo: {},
    permissions: []
  }),
  actions: {
    async login(data) {
      const res = await login(data)
      if (res.code === ERROR_CODES.SUCCESS) {
        // 登录成功，将token存储到localStorage
        localStorage.setItem('token', res.data.accessToken)
      //   //获取用户资料
      //   const userInfo = await getUserInfo()
      //   this.userInfo = userInfo.data
      //   //根据用户角色请求用户权限
      // const userMenusResult = await getUserMenusByRoleId(this.userInfo.role.id)
      // this.userMenus = userMenusResult.data
      // //本地存储
      // localCache.setCache('userInfo', this.userInfo)
      // localCache.setCache('userMenus', this.userMenus)
      
      ElMessage.success('登录成功')
      }else{
        ElMessage.error(res.message)
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
    },
    async fixPassword(data) {
      const res = await fixPassword(data)
      if (res.code === ERROR_CODES.SUCCESS) {
        ElMessage.success('密码修改成功')
      } else {
        ElMessage.error(res.message || '密码修改失败')
      }
    }
  }
})