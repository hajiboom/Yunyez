import { defineStore } from 'pinia'
import { login, logout, fixPassword, getUserInfo, getUserMenusByRoleId } from '@/api/login/login.js'
import { ERROR_CODES, ERROR_MESSAGES } from '@/utils/codes'
import { ElMessage } from 'element-plus'
import { mapMenusToRoutes, firstMenu, mapMenuToPermissions } from '@/utils/map-menus.js'
import router from '@/router'

export const useLoginStore = defineStore('login', {
  state: () => ({
    menuData: [],
    userInfo: {},
    permissions: []
  }),
  actions: {
    async login(data) {
      //先测试登录成功之后获取数据
      try {
        const menuData = await getUserMenusByRoleId(1)
        const userInfo = await getUserInfo(1)
        localStorage.setItem('menuData', JSON.stringify(menuData.data.data))
        localStorage.setItem('userInfo', JSON.stringify(userInfo.data))
        this.permissions = mapMenuToPermissions(menuData.data.data)
        
        
        //动态添加路由，页面刷新之后会丢失，只会执行router里面的内容
        const routes = mapMenusToRoutes(menuData.data.data)
        
        routes.forEach(route => router.addRoute('main', route))


        router.push(firstMenu.url)
      } catch (error) {
        ElMessage.error(error.message || '获取用户菜单失败')
        return
      }


    },
    async logout() {
      const res = await logout()

      if (res.code === ERROR_CODES.SUCCESS) {


        ElMessage.success('登出成功')
        window.location.href = '/login'
      }
    },
    async fixPassword(data) {
      const res = await fixPassword(data)
      if (res.code === ERROR_CODES.SUCCESS) {
        ElMessage.success('密码修改成功')
      } else {
        ElMessage.error(res.message || '密码修改失败')
      }
    },
    loadLocalCacheAction() {
      const userInfo = JSON.parse(localStorage.getItem('userInfo') || '{}') ?? {}
      const userMenus = JSON.parse(localStorage.getItem('menuData') || '[]') ?? []
      this.permissions = mapMenuToPermissions(userMenus)
      if (userInfo.id && userMenus.length) {
        // this.token = token
        this.userInfo = userInfo
        this.menuData = userMenus
        //动态添加路由实时挂到 name 为 main 的菜单下”
        const routes = mapMenusToRoutes(this.menuData)
        routes.forEach((route) =>
          router.addRoute('main', route)
        )
      }
    }
  }
})