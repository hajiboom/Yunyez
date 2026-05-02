import { createRouter, createWebHashHistory } from 'vue-router'
import { firstMenu } from '@/utils/map-menus'
import { useTabsStore } from '@/store/tabs'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue')
    },
    {
      path: '/',
      redirect: '/main'
    },
    {
      path: '/main',
      name: 'main',
      component: () => import('../views/main/main.vue'),
      children:[
        {
          path: 'personpage',
          component: () => import('@/views/main/personpage/index.vue'),
          meta: { title: '个人中心' }
        }
      ]
    }
   
  ]
})

// router.beforeEach(async (to, from, next) => {
//   const token = localStorage.getItem('token')
//   // 如果目标路径是登录页，直接放行，不要尝试刷新
//   if (to.path === '/login' || to.name === 'Login') {
//     next()
//     return
//   }
//   //如果是登录页面，且有token，则重定向到首页第一个菜单
//     if (to.path === '/main' && token && firstMenu) {
//       return firstMenu.url
//     }
//   // 2. 如果已有 Access Token，直接放行
//   if (token) {
//     next()
//     return
//   }else{
//     next('/login')  // 或 next({ name: 'Login' })
//     return
//   }

// })

router.afterEach((to) => {
  const tabsStore = useTabsStore()
  if (to.meta?.title) {  // 只处理有标题的路由（排除404等）
    tabsStore.addView({
      path: to.path,
      title: to.meta.title,
      name: to.name,
      query: to.query,
      params: to.params
    })
  }
})
export default router