import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      // 管理员 API 路由 - 仅代理非 GET 请求（API 调用），GET 请求走 SPA
      '/admin': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        bypass: (req) => {
          if (req.method === 'GET' && !req.headers.accept?.includes('application/json')) {
            return req.url
          }
        },
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
