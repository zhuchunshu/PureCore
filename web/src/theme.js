// Reads .env (VITE_THEME) first for the default DaisyUI theme name,
// falling back to theme.config.json if the env var is not set.
// Works in both browser (via Vite's import.meta.env) and Node.js (via process.env).
// Usage: import { themeName } from './theme.js'

import themeConfig from '../theme.config.json' with { type: 'json' }

function getEnvTheme() {
  if (typeof import.meta !== 'undefined' && import.meta.env?.VITE_THEME) {
    return import.meta.env.VITE_THEME
  }
  if (typeof process !== 'undefined' && process.env?.THEME) {
    return process.env.THEME
  }
  return null
}

export const themeName = getEnvTheme() || themeConfig.theme || 'sunset'
