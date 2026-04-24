import { ref, computed } from 'vue'

// 多语言支持模块，与后端共用 lang/ 目录下的 JSON 文件
const currentLocale = ref('zh')
const messages = ref({})
// 从 /lang/ 路径加载翻译文件（通过 public/lang 软链接）
async function loadLocale(locale) {
  try {
    const response = await fetch(`/lang/${locale}.json`)
    if (response.ok) {
      const data = await response.json()
      // 展平嵌套的 JSON 结构为 "group.key" 格式
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
      messages.value = { ...messages.value, ...flat }
    }
  } catch (err) {
    console.warn(`Failed to load locale "${locale}":`, err)
  }
}

// 检测浏览器语言
function detectBrowserLanguage() {
  const lang = navigator.language || navigator.userLanguage || 'zh'
  const primary = lang.split('-')[0]
  return primary === 'zh' || primary === 'en' ? primary : 'zh'
}

const STORAGE_KEY = 'purecore-locale'

// 初始化
export async function initI18n(locale) {
  const targetLocale = locale || localStorage.getItem(STORAGE_KEY) || detectBrowserLanguage()
  currentLocale.value = targetLocale
  await loadLocale(targetLocale)
}

// 翻译函数
export function t(key, defaultVal) {
  return messages.value[key] || defaultVal || key
}

// 切换语言
export async function setLocale(locale) {
  currentLocale.value = locale
  localStorage.setItem(STORAGE_KEY, locale)
  await loadLocale(locale)
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
