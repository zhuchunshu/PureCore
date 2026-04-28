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
  /** Backend API base URL (proxied via Vite/SSR server in development) */
  get apiBaseUrl() {
    return import.meta.env.VITE_API_BASE_URL || '/api/v1'
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
