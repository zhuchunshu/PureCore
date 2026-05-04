/**
 * Backward-compatible wrapper for user authentication.
 * 
 * This module now delegates to the centralized API service (web/src/services/api.js).
 * All existing imports of useUserAuth.js continue to work unchanged, but the underlying
 * logic is now shared with useAuth.js through a single configurable service.
 * 
 * For new code, prefer importing directly from '@/services/api':
 *   import { userAPI } from '@/services/api'
 *   userAPI.setTokens(access, refresh)
 *   await userAPI.fetch('/some-endpoint')
 */

import { userAPI } from '@/services/api'

// Re-export reactive state for backward compatibility
export const { accessToken, refreshToken, currentUser } = userAPI

// Re-export auth functions with the same names
export const setTokens = userAPI.setTokens.bind(userAPI)
export const clearTokens = userAPI.clearTokens.bind(userAPI)
export const refreshAccessToken = userAPI.refreshAccessToken.bind(userAPI)
export const refreshIfNeeded = userAPI.refreshIfNeeded.bind(userAPI)
export const fetchProfile = userAPI.fetchProfile.bind(userAPI)

// Auth fetch wrapper delegates to the centralized service
export const authFetch = userAPI.fetch.bind(userAPI)

// Check if user is logged in
export const isLoggedIn = userAPI.isLoggedIn.bind(userAPI)

// Initialize on import — backward compatible with existing initUserAuth() calls
export function initUserAuth() {
  // The API service initializes itself on creation, so this is a no-op.
  // Kept for backward compatibility with existing main.js imports.
}

// Composable for user auth state — function-based API used by IntegrationsPage.vue
export function useUserAuth() {
  return {
    accessToken,
    refreshToken,
    currentUser,
    isLoggedIn,
    setTokens,
    clearTokens,
    refreshAccessToken,
    refreshIfNeeded,
    fetchProfile,
    authFetch,
  }
}
