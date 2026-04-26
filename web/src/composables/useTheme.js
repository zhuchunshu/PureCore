import { ref, watch } from 'vue'
import { themeName as configTheme } from '../theme'

const STORAGE_KEY = 'purecore-theme-mode'
const COOKIE_KEY = 'purecore-theme'

/**
 * Theme modes:
 * - 'auto': follow system preference
 * - 'light': force light mode
 * - 'dark': force dark mode
 * or any DaisyUI theme name (e.g., 'sunset', 'cyberpunk', etc.) for a specific theme override.
 *
 * Priority: localStorage manual > system preference > theme.config.json default
 */

const isServer = typeof window === 'undefined'

// Singleton reactive state
const currentTheme = ref(__loadInitialTheme())

function __loadInitialTheme() {
  if (isServer) return configTheme
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved) return saved
  return 'auto'
}

function __getSystemPreference() {
  if (isServer) return 'dark'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function __resolveTheme(mode) {
  if (mode === 'auto') return __getSystemPreference()
  if (mode === 'light' || mode === 'dark') return mode
  return mode // specific theme name
}

function __setCookie(name, value, days = 365) {
  const expires = new Date(Date.now() + days * 864e5).toUTCString()
  document.cookie = `${name}=${encodeURIComponent(value)};expires=${expires};path=/;SameSite=Lax`
}

function applyTheme(mode) {
  currentTheme.value = mode
  const resolved = __resolveTheme(mode)
  document.documentElement.setAttribute('data-theme', resolved)
  // Store the RESOLVED actual theme name in cookie, so SSR can use it directly
  // (avoids SSR needing to resolve 'auto' → system preference)
  __setCookie(COOKIE_KEY, resolved)
}

export function useTheme() {
  // Apply theme reactively when currentTheme changes
  if (!isServer) {
    watch(currentTheme, (mode) => {
      const resolved = __resolveTheme(mode)
      document.documentElement.setAttribute('data-theme', resolved)
    })

    // On mount, apply the stored/saved theme
    applyTheme(currentTheme.value)

    // Listen for system preference changes (only when in 'auto' mode)
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', () => {
      if (currentTheme.value === 'auto') {
        const resolved = __getSystemPreference()
        document.documentElement.setAttribute('data-theme', resolved)
      }
    })
  }

  function setThemeMode(mode) {
    if (!isServer) {
      localStorage.setItem(STORAGE_KEY, mode)
    }
    applyTheme(mode)
  }

  return {
    theme: currentTheme,
    setThemeMode,
    resolvedTheme: () => __resolveTheme(currentTheme.value),
  }
}
