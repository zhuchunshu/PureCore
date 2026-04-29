import { createSSRApp } from 'vue'
import { renderToString } from '@vue/server-renderer'
import { createRouter as createVueRouter, createMemoryHistory } from 'vue-router'
import App from './App.vue'
import { initI18n, setLocale } from './i18n'
import { routes } from './router/routes'

// Map route names to i18n translation keys for page title and description
const routeSEO = {
  Home: { title: 'home.title', description: 'home.description' },
  Login: { title: 'user.login_title', description: 'user.login_description' },
  Register: { title: 'user.register_title', description: 'user.register_description' },
  AdminLogin: { title: 'admin.title', description: 'admin.title' },
  AdminRegister: { title: 'admin.register_title', description: 'admin.register_title' },
  AdminDashboard: { title: 'admin.dashboard', description: 'admin.dashboard' },
  AdminSettings: { title: 'admin.settings', description: 'admin.settings_description' },
  NotFound: { title: 'notfound.title', description: 'notfound.description' },
  'Docs-404': { title: 'notfound.title', description: 'notfound.description' },
}

// Build API base URL for SSR-side fetch (mirrors config.js logic)
function getApiBaseUrl() {
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL + '/api/v1'
  }
  const protocol = import.meta.env.VITE_API_PROTOCOL
  const host = import.meta.env.VITE_API_HOST
  const port = import.meta.env.VITE_API_PORT
  if (protocol && host) {
    return `${protocol}://${host}${port ? ':' + port : ''}/api/v1`
  }
  return import.meta.env.SSR ? 'http://localhost:9002/api/v1' : '/api/v1'
}

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

  // Determine if this is a 404 page by checking the matched route name
  const currentRoute = router.currentRoute.value
  const routeName = currentRoute.name || 'Home'
  let statusCode = routeName === 'NotFound' ? 404 : 200

  // For Docs routes, verify the document exists via the backend API
  if (routeName === 'Docs' && import.meta.env.SSR) {
    const params = currentRoute.params
    const docLocale = params.locale || 'en'
    const docPage = params.page || 'README'
    try {
      const apiBase = getApiBaseUrl()
      const resp = await fetch(`${apiBase}/docs?locale=${docLocale}&page=${docPage}`)
      const json = await resp.json()
      if (resp.status === 404 || json.code === 404) {
        statusCode = 404
      }
    } catch (e) {
      console.error('[SSR] Docs API check failed:', e.message)
      // Don't treat backend unreachable as 404 — let client-side handle it
    }
  }

  // Use the potentially updated routeName (could be 'Docs-404' from docs check)
  const effectiveRouteName = (statusCode === 404 && routeName === 'Docs') ? 'Docs-404' : routeName

  // Compute page-specific SEO title and description from i18n translations
  const seo = routeSEO[effectiveRouteName] || routeSEO[routeName] || { title: '', description: '' }
  const t = (key, fallback) => {
    // translations is a flat map of key -> value
    const val = translations[locale]
    if (val && val[key]) return val[key]
    return fallback
  }
  const siteName = projectInfo?.name || 'PureCore'
  const pageTitleKey = t(seo.title, '')
  const pageDescKey = t(seo.description, '')
  const pageTitle = pageTitleKey ? `${pageTitleKey} - ${siteName}` : siteName
  const pageDescription = pageDescKey || ''

  const html = await renderToString(app)

  return { html, statusCode, pageTitle, pageDescription }
}
