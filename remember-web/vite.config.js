import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { getPorts } from './read-config.js'

// 获取端口配置
const ports = getPorts();

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',  // 监听所有地址
    port: ports.web,   // 从config.yaml读取端口
    allowedHosts: [
        'u459706-b934-cfb53c08.bjb1.seetacloud.com',
        '.bjb1.seetacloud.com',
        'localhost',
        '127.0.0.1'
    ], // 允许所有 host，避免容器域名被阻止
    proxy: {
      // 记忆服务代理到主服务端口
      '/api/memory': {
        target: `http://localhost:${ports.main}`,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/memory/, '/memory')
      },
      // AI响应服务代理到OpenAI服务端口
      '/api/response': {
        target: `http://localhost:${ports.openai}`,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/response/, '/v1/response')
      }
    }
  }
})
