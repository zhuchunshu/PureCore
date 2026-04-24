import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import { initI18n } from './i18n'

async function bootstrap() {
  await initI18n()

  const app = createApp(App)
  app.use(router)
  app.mount('#app')
}

bootstrap()
