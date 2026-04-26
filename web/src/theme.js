// Reads theme.config.json and provides the active DaisyUI theme name.
// Works in both browser (via Vite's JSON import) and Node.js (via fs).
// Usage: import { themeName } from './theme.js'

import themeConfig from '../theme.config.json' with { type: 'json' }

export const themeName = themeConfig.theme || 'sunset'
