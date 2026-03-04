import { defineConfig, loadEnv } from 'vite' // 新增 loadEnv
import path from 'path'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// 接收 mode 参数，显式加载环境变量
export default defineConfig(({ mode }) => {
  // 加载根目录下对应环境的 .env 文件
  const env = loadEnv(mode, process.cwd())
  // 打印验证（启动项目时看控制台，确认变量读到了）
  console.log('后端地址：', env.VITE_API_BASE_URL)

  return {
    plugins: [
      vue(),
      AutoImport({ resolvers: [ElementPlusResolver()] }),
      Components({ resolvers: [ElementPlusResolver()] }),
    ],
    resolve: {
      alias: { '@': path.resolve(__dirname, './src') }
    },
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_BASE_URL, // 用加载的环境变量
          changeOrigin: true,
        },
      },
    },
  }
})