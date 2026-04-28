/**
 * Backward-compatible wrapper for admin authentication.
 * 
 * This module now delegates to the centralized API service (web/src/services/api.js).
 * All existing imports of useAuth.js continue to work unchanged, but the underlying
 * logic is now shared with userAuth.js through a single configurable service.
 * 
 * For new code, prefer importing directly from '@/services/api':
 *   import { adminAPI } from '@/services/api'
 *   adminAPI.setTokens(access, refresh)
 *   await adminAPI.fetch('/some-endpoint')
 */

import { adminAPI } from '@/services/api'

// Re-export reactive state for backward compatibility
export const { accessToken, refreshToken } = adminAPI

// Re-export auth functions with the same names
export const setTokens = adminAPI.setTokens.bind(adminAPI)
export const clearTokens = adminAPI.clearTokens.bind(adminAPI)
export const refreshAccessToken = adminAPI.refreshAccessToken.bind(adminAPI)
export const refreshIfNeeded = adminAPI.refreshIfNeeded.bind(adminAPI)

// Auth fetch wrapper delegates to the centralized service
export const authFetch = adminAPI.fetch.bind(adminAPI)

// Initialize on import — backward compatible with existing initAuth() calls
export function initAuth() {
  // The API service initializes itself on creation, so this is a no-op.
  // Kept for backward compatibility with existing main.js imports.
}
