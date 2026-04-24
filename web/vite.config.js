import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [
      vue(),
      tailwindcss(),
    ],
    server: {
      port: parseInt(env.FRONTEND_PORT) || 9001,
      proxy: {
        '/api': {
          target: `http://localhost:${env.BACKEND_PORT || 9002}`,
          changeOrigin: true,
        },
      },
    },
  }
})
