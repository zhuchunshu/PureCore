import { watch, getCurrentInstance } from 'vue'
import { useI18n } from '../i18n'
import { adminOption } from './useAdminOption'

// Default SEO values — used as fallbacks
const defaults = {
  title: 'PureCore',
  description: '',
  keywords: '',
}

let globalTitle = ''
let globalDescription = ''
let globalKeywords = ''

/**
 * Initialize global SEO values from admin options.
 * Call this once at app startup.
 */
export async function initSEO() {
  globalTitle = await adminOption('site_name', 'PureCore')
  globalDescription = await adminOption('site_description', '')
  globalKeywords = await adminOption('site_keywords', '')
}

/**
 * Vue composable for setting page-level SEO meta tags.
 * Works both client-side (via document.head) and SSR (via head data injection).
 *
 * @param {object} options
 * @param {string} options.title - Page title (appended to site name)
 * @param {string} options.description - Page meta description
 * @param {string} options.keywords - Page meta keywords
 */
export function useSEO({ title = '', description = '', keywords = '' } = {}) {
  const { t } = useI18n()

  // Build page title: "Page Title - Site Name" or just "Site Name"
  const siteName = globalTitle || defaults.title
  const pageTitle = title ? `${title} - ${siteName}` : siteName
  const pageDescription = description || globalDescription || defaults.description
  const pageKeywords = keywords || globalKeywords || defaults.keywords

  // Client-side: update document head
  if (typeof document !== 'undefined') {
    document.title = pageTitle
    setMeta('description', pageDescription)
    setMeta('keywords', pageKeywords)
  }

  // SSR: provide head data for server-side rendering
  const instance = getCurrentInstance()
  if (instance && instance.appContext) {
    const app = instance.appContext.app
    if (app && app.config && app.config.globalProperties) {
      app.config.globalProperties.$seo = {
        title: pageTitle,
        description: pageDescription,
        keywords: pageKeywords,
      }
    }
  }

  return {
    pageTitle,
    pageDescription,
    pageKeywords,
  }
}

function setMeta(name, content) {
  let el = document.querySelector(`meta[name="${name}"]`)
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute('name', name)
    document.head.appendChild(el)
  }
  el.setAttribute('content', content)
}
