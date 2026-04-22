import { createRouter, createWebHashHistory } from 'vue-router'
import request from '@/utils/request'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'Login',
      component: () => import('@/views/login/index.vue')
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('@/views/Dashboard/index.vue'),
      children: [
        {
          path: '/deviceManage',
          name: 'DeviceManage',
          component: () => import('@/views/Dashboard/deviceManage/index.vue')
        },
        {
          path: '/personPage',
          name: 'personPage',
          component: () => import('@/views/Dashboard/personPage/index.vue')
        }
      ]
    },
  ]
})

router.beforeEach(async (to, from, next) => {
  const token = localStorage.getItem('token')
  // 如果目标路径是登录页，直接放行，不要尝试刷新
  if (to.path === '/' || to.name === 'Login') {
    next()
    return
  }
  
  // 2. 如果已有 Access Token，直接放行
  if (token) {
    next()
    return
  }else{
    next('/')  // 或 next({ name: 'Login' })
    return
  }
})

export default router