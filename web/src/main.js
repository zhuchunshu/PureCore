import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import { initI18n } from './i18n'
import { useTheme } from './composables/useTheme'

async function bootstrap() {
  // Apply theme from localStorage or system preference before mounting the app
  // useTheme() initializes itself immediately (reads localStorage on module load)
  // The watch in useTheme will handle applying and listening for system changes
  const { theme } = useTheme()
  // Force the initial apply - the composable watches for changes automatically
  // eslint-disable-next-line no-unused-expressions
  theme.value

  await initI18n()

  const app = createApp(App)
  app.use(router)
  app.mount('#app')
}

bootstrap()
