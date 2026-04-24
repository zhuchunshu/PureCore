import { ref, computed } from 'vue'

// 多语言支持模块，与后端共用 lang/ 目录下的 JSON 文件
const currentLocale = ref('zh')
const messages = ref({})
const isServer = typeof window === 'undefined'

// 展平嵌套的 JSON 结构为 "group.key" 格式
function flattenMessages(data) {
  const flat = {}
  for (const [group, obj] of Object.entries(data)) {
    if (typeof obj === 'object') {
      for (const [key, value] of Object.entries(obj)) {
        flat[`${group}.${key}`] = value
      }
    } else {
      flat[group] = obj
    }
  }
  return flat
}

// 客户端：从 /lang/ 路径加载翻译文件（通过 public/lang 软链接）
async function loadLocaleClient(locale) {
  try {
    const response = await fetch(`/lang/${locale}.json`)
    if (response.ok) {
      const data = await response.json()
      messages.value = { ...messages.value, ...flattenMessages(data) }
    }
  } catch (err) {
    console.warn(`Failed to load locale "${locale}":`, err)
  }
}

// 服务端：直接使用预加载的翻译数据
function loadLocaleServer(locale, translations) {
  if (translations && translations[locale]) {
    messages.value = { ...messages.value, ...flattenMessages(translations[locale]) }
  }
}

// 检测浏览器语言（仅客户端）
function detectBrowserLanguage() {
  if (isServer) return 'zh'
  const lang = navigator.language || navigator.userLanguage || 'zh'
  const primary = lang.split('-')[0]
  return primary === 'zh' || primary === 'en' ? primary : 'zh'
}

const STORAGE_KEY = 'purecore-locale'

// 初始化（支持 SSR：传入 translations 参数则直接使用，否则 fetch）
export function initI18n(locale, translations) {
  const targetLocale = locale || (!isServer ? localStorage.getItem(STORAGE_KEY) : null) || detectBrowserLanguage()
  currentLocale.value = targetLocale

  if (isServer || translations) {
    loadLocaleServer(targetLocale, translations)
  } else {
    return loadLocaleClient(targetLocale)
  }
}

// 翻译函数
export function t(key, defaultVal) {
  return messages.value[key] || defaultVal || key
}

// 切换语言
export async function setLocale(locale) {
  currentLocale.value = locale
  if (!isServer) {
    localStorage.setItem(STORAGE_KEY, locale)
    await loadLocaleClient(locale)
  } else {
    // Server-side: locale is set but translations are preloaded
  }
}

// 获取当前语言
export function getLocale() {
  return currentLocale.value
}

// Vue composable
export function useI18n() {
  return {
    t,
    setLocale,
    getLocale,
    locale: computed(() => currentLocale.value),
    currentLocale: computed(() => currentLocale.value),
  }
}

// 默认导出，方便在 main.js 中使用
export default {
  initI18n,
  t,
  setLocale,
  getLocale,
  useI18n,
}
