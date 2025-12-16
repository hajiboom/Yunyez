import { defineConfig } from 'vite'
import path from 'path'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(),
     AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
  ],
  resolve: {
    // 2. 配置别名：@ 指向项目根目录下的 src 文件夹
    alias: {
      '@': path.resolve(__dirname, './src') 
      // __dirname 是当前文件（vite.config.js）的目录（即项目根目录）
      // ./src 是相对根目录的路径，path.resolve 转为绝对路径，兼容所有系统
    }
  },
  // CORS 配置
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
  
})
