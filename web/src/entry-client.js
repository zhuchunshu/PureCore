import './style.css'
import { createSSRApp } from 'vue'
import { createWebHistory } from 'vue-router'
import App from './App.vue'
import { createSSRRouter } from './router'
import { initI18n } from './i18n'

async function bootstrap() {
  // Initialize i18n on the client (fetches from /lang/ via network)
  await initI18n()

  const app = createSSRApp(App)
  const router = createSSRRouter(createWebHistory())
  app.use(router)

  // Wait for router to be ready before mounting (for initial navigation)
  await router.isReady()

  app.mount('#app')
}

bootstrap()
