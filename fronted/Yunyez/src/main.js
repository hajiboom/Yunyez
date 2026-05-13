import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import  '../font_5088216_0mvq8ei02qgj/iconfont.css'
import router from './router'
// 1. 导入Pinia
import { createPinia } from 'pinia'
import { useLoginStore } from './store/login.js'


// 2. 创建Pinia实例
const pinia = createPinia()
const app = createApp(App)
app.use(pinia)
const loginStore = useLoginStore()
loginStore.loadLocalCacheAction()
app.use(router)

app.mount('#app')
