import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import  '../font_5088216_vzliewbsbs/iconfont.css'
import router from './router'
// 1. 导入Pinia
import { createPinia } from 'pinia'

// 2. 创建Pinia实例
const pinia = createPinia()
// 3. 挂载Pinia到Vue实例
createApp(App).use(router).use(pinia).mount('#app')
