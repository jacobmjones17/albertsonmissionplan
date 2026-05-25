import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '..', '')
  const backendPort = env.PORT || '8080'
  const backend = `http://127.0.0.1:${backendPort}`

  return {
    plugins: [react()],
    build: {
      outDir: '../backend/web/dist',
      emptyOutDir: true,
    },
    server: {
      port: 5173,
      strictPort: true,
      proxy: {
        '/api': { target: backend, changeOrigin: true },
        '/auth': { target: backend, changeOrigin: true },
      },
    },
  }
})
