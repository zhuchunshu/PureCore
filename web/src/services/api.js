/**
 * Centralized API service layer with built-in token management.
 * 
 * This module consolidates the duplicate logic from useAuth.js and useUserAuth.js
 * into a single, configurable service. It provides:
 * 
 * - JWT token acquisition, storage, and auto-refresh
 * - Authorization header injection
 * - Proactive token refresh based on TTL timers
 * - Single-flight refresh queue (concurrent 401s share one refresh call)
 * - Page visibility handler for tab-switching scenarios
 * 
 * Usage:
 *   import { createApiService } from '@/services/api'
 *   const api = createApiService({ type: 'admin' })  // or 'user'
 *   await api.fetch('/users')  // auto-injects Authorization header
 */

import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { config } from './config'

// Token TTL constants (mirrors backend: 15 min access, 7 day refresh)
const ACCESS_TOKEN_TTL = 15 * 60 * 1000    // 15 minutes
const REFRESH_INTERVAL = 5 * 60 * 1000     // Try to refresh every 5 minutes
const SAFE_REFRESH_MARGIN = 3 * 60 * 1000  // Refresh proactively if token expires within 3 minutes

/**
 * Creates an API service instance for the given auth type.
 * 
 * @param {'admin'|'user'} options.type - Authentication type
 * @param {object} [options.router] - Vue Router instance (auto-imported if not provided)
 * @returns {object} API service with fetch, auth methods, and reactive state
 */
export function createApiService({ type = 'admin' } = {}) {
  // Determine storage keys based on auth type
  const storageKey = type === 'admin' ? 'admin' : 'user'
  const TOKEN_KEY = `${storageKey}_token`
  const REFRESH_KEY = `${storageKey}_refresh_token`
  const PROFILE_KEY = `${storageKey}_profile`

  // Determine refresh endpoint
  const refreshUrl = type === 'admin'
    ? `${config.apiBaseUrl}/${config.adminRoutePrefix}/auth/refresh`
    : `${config.apiBaseUrl}/auth/refresh`

  // Determine login redirect path
  const loginPath = type === 'admin'
    ? `/${config.adminRoutePrefix}/login`
    : '/login'

  // Reactive state — SSR-safe initialization
  const getInitial = (key) => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem(key) || null
    }
    return null
  }

  const accessToken = ref(getInitial(TOKEN_KEY))
  const refreshToken = ref(getInitial(REFRESH_KEY))
  const currentUser = ref(null)

  // Refresh management
  let lastRefreshTime = 0
  let refreshTimer = null
  let isRedirecting = false
  let visibilityListener = null
  let refreshPromise = null

  /**
   * Returns true if the access token is within SAFE_REFRESH_MARGIN of expiring
   */
  function tokenExpiresSoon() {
    if (!lastRefreshTime) return false
    const elapsed = Date.now() - lastRefreshTime
    return elapsed >= (ACCESS_TOKEN_TTL - SAFE_REFRESH_MARGIN)
  }

  /**
   * Sets tokens in localStorage and reactive state.
   * Starts the automatic refresh timer.
   */
  function setTokens(access, refresh) {
    accessToken.value = access
    refreshToken.value = refresh
    lastRefreshTime = Date.now()
    localStorage.setItem(TOKEN_KEY, access)
    if (refresh) {
      localStorage.setItem(REFRESH_KEY, refresh)
    }
    startTokenRefresh()
  }

  /**
   * Clears all stored tokens and cancels timers.
   * Redirects to the login page after a short delay.
   */
  function clearTokens() {
    accessToken.value = null
    refreshToken.value = null
    currentUser.value = null
    lastRefreshTime = 0
    refreshPromise = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
    localStorage.removeItem(PROFILE_KEY)
    if (refreshTimer) {
      clearTimeout(refreshTimer)
      refreshTimer = null
    }
    if (visibilityListener) {
      document.removeEventListener('visibilitychange', visibilityListener)
      visibilityListener = null
    }
    // Prevent redirect storms
    setTimeout(() => { isRedirecting = false }, 100)
  }

  /**
   * Refreshes the access token using the stored refresh token.
   * Single-flight: concurrent callers receive the same promise.
   * Returns the new token on success, null on failure.
   */
  function refreshAccessToken() {
    if (!refreshToken.value) return Promise.resolve(null)
    if (refreshPromise) return refreshPromise

    refreshPromise = (async () => {
      try {
        const resp = await fetch(refreshUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken.value }),
        })
        const json = await resp.json()
        if (json.code === 0) {
          setTokens(json.data.token, json.data.refresh_token)
          return json.data.token
        }
      } catch (err) {
        console.error(`[api:${type}] Token refresh failed:`, err)
      }
      // Refresh failed — clear everything and redirect with return URL + reason
      if (!isRedirecting) {
        isRedirecting = true
        clearTokens()
        const router = useRouter()
        const currentPath = router.currentRoute.value.fullPath
        router.push({
          path: loginPath,
          query: {
            reason: 'session_expired',
            ...(currentPath !== loginPath ? { redirect: currentPath } : {}),
          },
        })
      }
      return null
    })().finally(() => {
      refreshPromise = null
    })

    return refreshPromise
  }

  /**
   * Refreshes the token if it's close to expiring.
   * Called before API calls and on page visibility changes.
   */
  function refreshIfNeeded() {
    if (!accessToken.value || !refreshToken.value) return Promise.resolve()
    if (tokenExpiresSoon()) {
      return refreshAccessToken()
    }
    return Promise.resolve()
  }

  /**
   * Starts the automatic token refresh interval.
   * Runs every REFRESH_INTERVAL ms while the page is open.
   */
  function startTokenRefresh() {
    if (refreshTimer) clearInterval(refreshTimer)
    if (!refreshToken.value) return
    refreshTimer = setInterval(() => {
      if (tokenExpiresSoon()) {
        refreshAccessToken()
      }
    }, REFRESH_INTERVAL)
  }

  /**
   * Handles page visibility changes — refreshes when the user returns to the tab
   */
  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      refreshIfNeeded()
    }
  }

  /**
   * Initializes token refresh on app startup.
   * Called automatically when the service is created.
   */
  function init() {
    if (refreshToken.value) {
      // Load cached profile if available
      const cached = localStorage.getItem(PROFILE_KEY)
      if (cached) {
        try {
          currentUser.value = JSON.parse(cached)
        } catch (_) { /* ignore parse errors */ }
      }
      startTokenRefresh()
      visibilityListener = handleVisibilityChange
      document.addEventListener('visibilitychange', visibilityListener)
    }
  }

  /**
   * Fetches the user profile from the API.
   * Automatically refreshes the token if needed before the request.
   */
  async function fetchProfile() {
    if (!accessToken.value) return null
    await refreshIfNeeded()
    try {
      // Admin and user profiles live at different endpoints:
      //   admin → /api/v1/{adminRoutePrefix}/auth/profile
      //   user  → /api/v1/auth/profile
      const profileUrl = type === 'admin'
        ? `${config.apiBaseUrl}/${config.adminRoutePrefix}/auth/profile`
        : `${config.apiBaseUrl}/auth/profile`
      const resp = await fetch(profileUrl, {
        headers: { Authorization: `Bearer ${accessToken.value}` },
      })
      if (resp.status === 401) {
        clearTokens()
        const router = useRouter()
        const currentPath = router.currentRoute.value.fullPath
        router.push({ path: loginPath, query: { reason: 'session_expired', redirect: currentPath } })
        return null
      }
      const json = await resp.json()
      if (json.code === 0) {
        currentUser.value = json.data
        localStorage.setItem(PROFILE_KEY, JSON.stringify(json.data))
        return json.data
      }
    } catch (err) {
      console.error(`[api:${type}] Failed to fetch profile:`, err)
    }
    return null
  }

  /**
   * Returns true if the user is currently authenticated
   */
  function isLoggedIn() {
    return !!accessToken.value
  }

  /**
   * Make an authenticated API request.
   * 
   * Proactively refreshes the token before the request, and if a 401 is
   * received, queues behind the single-flight refresh promise before retrying.
   * This matches the industry-standard pattern used by Axios and OkHttp.
   * 
   * @param {string} url - API endpoint (e.g., '/users')
   * @param {object} [options] - Fetch options (method, headers, body, etc.)
   * @returns {Promise<Response>} Fetch response
   */
  function authFetch(url, options = {}) {
    const makeRequest = (token) => {
      const headers = { ...options.headers }
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }
      return fetch(url, { ...options, headers })
    }

    return refreshIfNeeded()
      .then(() => makeRequest(accessToken.value))
      .then(async (resp) => {
        if (resp.status === 401 && refreshToken.value) {
          const newToken = await refreshAccessToken()
          if (newToken) {
            return makeRequest(newToken)
          }
        }
        return resp
      })
  }

  // Initialize on creation
  init()

  return {
    // Reactive state
    accessToken,
    refreshToken,
    currentUser,
    isLoggedIn,

    // Auth methods
    setTokens,
    clearTokens,
    refreshAccessToken,
    refreshIfNeeded,
    fetchProfile,

    // HTTP methods
    fetch: authFetch,
    get: (url, opts = {}) => authFetch(url, { ...opts, method: 'GET' }),
    post: (url, data, opts = {}) => authFetch(url, {
      ...opts,
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...opts.headers },
      body: JSON.stringify(data),
    }),
    put: (url, data, opts = {}) => authFetch(url, {
      ...opts,
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...opts.headers },
      body: JSON.stringify(data),
    }),
    delete: (url, opts = {}) => authFetch(url, { ...opts, method: 'DELETE' }),
  }
}

/**
 * Pre-configured admin API service instance.
 * This replaces the old useAuth.js composable.
 */
export const adminAPI = createApiService({ type: 'admin' })

/**
 * Pre-configured user API service instance.
 * This replaces the old useUserAuth.js composable.
 */
export const userAPI = createApiService({ type: 'user' })
