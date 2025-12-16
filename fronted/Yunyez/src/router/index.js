import { createRouter, createWebHashHistory } from 'vue-router'
import login from '@/views/login/index.vue'


const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/login',
        name: 'Login' ,
        component: () => import('@/views/login/index.vue')
    },
    {
        path: '/Dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard/index.vue'),
        children:[
          {     
            path:'deviceManage',
      name:'DeviceManage',
      component: () => import('@/views/Dashboard/deviceManage/index.vue')
          }
      ]

    },
    
  ]
})


export default router