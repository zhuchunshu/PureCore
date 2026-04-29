import { defineConfig, loadEnv } from 'vite'
import { fileURLToPath, URL } from 'node:url'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  // Build API proxy target from environment variables
  const apiProtocol = env.VITE_API_PROTOCOL || 'http'
  const apiHost = env.VITE_API_HOST || 'localhost'
  const apiPort = env.VITE_API_PORT || env.BACKEND_PORT || '9002'
  const apiTarget = `${apiProtocol}://${apiHost}:${apiPort}`

  return {
    plugins: [
      vue(),
      tailwindcss(),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: parseInt(env.FRONTEND_PORT) || 9001,
      proxy: {
        '/api': {
          target: apiTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
