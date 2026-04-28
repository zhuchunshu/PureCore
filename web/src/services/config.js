/**
 * Centralized frontend configuration service.
 * 
 * This mirrors the backend's core/config.go pattern, providing a single source
 * of truth for all runtime configuration. All environment variables are accessed
 * through this module, making them easy to mock in tests and keeping components
 * decoupled from Vite-specific import.meta.env.
 * 
 * Usage:
 *   import { config } from '@/services/config'
 *   const apiUrl = config.adminRoutePrefix
 */

// All configuration values are loaded from Vite environment variables.
// In production, these are injected at build time or set via the SSR server.
// In development, they come from the .env file in the web/ directory.

export const config = {
  /**
   * Backend API base URL.
   *
   * Constructed from VITE_API_PROTOCOL, VITE_API_HOST, and VITE_API_PORT.
   * Falls back to relative `/api/v1` if any part is missing (SSR proxy mode).
   *
   * Examples:
   *   VITE_API_PROTOCOL=http  VITE_API_HOST=localhost  VITE_API_PORT=9002  → http://localhost:9002/api/v1
   *   VITE_API_PROTOCOL=https VITE_API_HOST=api.example.com              → https://api.example.com/api/v1
   *   (missing)                                                           → /api/v1 (proxy mode)
   */
  get apiBaseUrl() {
    if (import.meta.env.VITE_API_BASE_URL) {
      return import.meta.env.VITE_API_BASE_URL + '/api/v1'
    }
    const protocol = import.meta.env.VITE_API_PROTOCOL
    const host = import.meta.env.VITE_API_HOST
    const port = import.meta.env.VITE_API_PORT
    if (protocol && host) {
      return `${protocol}://${host}${port ? ':' + port : ''}/api/v1`
    }
    return '/api/v1'
  },

  /** Admin route prefix from environment, defaults to 'control-panel' */
  get adminRoutePrefix() {
    return import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'
  },

  /** Full admin API path: /api/v1/{adminRoutePrefix} */
  get adminApiPath() {
    return `/api/v1/${this.adminRoutePrefix}`
  },

  /** Whether the app is running in development mode */
  get isDevelopment() {
    return import.meta.env.DEV === true
  },

  /** Whether the app is running in production mode */
  get isProduction() {
    return import.meta.env.PROD === true
  },

  /** Whether SSR is enabled */
  get isSSR() {
    return import.meta.env.SSR === true
  },

  /** Get site name from environment (may be overridden by admin options at runtime) */
  get siteName() {
    return import.meta.env.VITE_SITE_NAME || 'PureCore'
  },

  /** Get site description */
  get siteDescription() {
    return import.meta.env.VITE_SITE_DESCRIPTION || ''
  },

  /** Get site keywords */
  get siteKeywords() {
    return import.meta.env.VITE_SITE_KEYWORDS || ''
  }
}

export default config
