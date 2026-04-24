import { createSSRApp } from 'vue'
import { renderToString } from '@vue/server-renderer'
import { createRouter as createVueRouter, createMemoryHistory } from 'vue-router'
import App from './App.vue'
import { initI18n, setLocale } from './i18n'
import { routes } from './router/routes'

export async function render(url, { locale = 'zh', translations = {}, projectInfo = null } = {}) {
  // Initialize i18n with preloaded translations (no fetch needed on server)
  initI18n(locale, translations)

  const app = createSSRApp(App)

  // Provide project info to all components (eliminates client-side fetch flash)
  if (projectInfo) {
    app.provide('projectInfo', projectInfo)
  }

  const router = createVueRouter({
    history: createMemoryHistory(),
    routes,
  })

  app.use(router)

  // Navigate to the requested URL
  await router.push(url)
  await router.isReady()

  const html = await renderToString(app)

  return { html }
}
