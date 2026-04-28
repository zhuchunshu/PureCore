import { defineConfig, loadEnv } from 'vite'
import { fileURLToPath, URL } from 'node:url'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

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
          target: `http://localhost:${env.BACKEND_PORT || 9002}`,
          changeOrigin: true,
        },
      },
    },
    build: {
      ssr: 'src/entry-server.js',
    },
  }
})
