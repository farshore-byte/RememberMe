import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      // 记忆服务代理到端口9100
      '/api/memory': {
        target: 'http://localhost:9100',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/memory/, '/memory')
      },
      // AI响应服务代理到端口8444
      '/api/response': {
        target: 'http://localhost:8444',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/response/, '/v1/response')
      }
    }
  }
})
