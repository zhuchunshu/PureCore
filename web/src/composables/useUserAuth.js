import { ref } from 'vue'
import { useRouter } from 'vue-router'

// Reactive state - check for browser environment (SSR-safe)
const getInitialAccessToken = () => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('user_token') || null
  }
  return null
}

const getInitialRefreshToken = () => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('user_refresh_token') || null
  }
  return null
}

const accessToken = ref(getInitialAccessToken())
const refreshToken = ref(getInitialRefreshToken())
const currentUser = ref(null)

// Track last refresh time as a simple timestamp
let lastRefreshTime = 0
let refreshTimer = null
let isRedirecting = false
let visibilityListener = null

// Refresh promise queue — ensures only one refresh at a time
let refreshPromise = null

// Token TTL constants
const ACCESS_TOKEN_TTL = 15 * 60 * 1000    // 15 minutes
const REFRESH_INTERVAL = 5 * 60 * 1000     // Try to refresh every 5 minutes
const SAFE_REFRESH_MARGIN = 3 * 60 * 1000  // Refresh proactively if token expires within 3 minutes

/**
 * Set authentication tokens in localStorage and state
 */
export function setTokens(access, refresh) {
  accessToken.value = access
  refreshToken.value = refresh
  lastRefreshTime = Date.now()
  localStorage.setItem('user_token', access)
  if (refresh) {
    localStorage.setItem('user_refresh_token', refresh)
  }
  startTokenRefresh()
}

/**
 * Clear all tokens and redirect to login
 */
export function clearTokens() {
  accessToken.value = null
  refreshToken.value = null
  currentUser.value = null
  lastRefreshTime = 0
  refreshPromise = null
  localStorage.removeItem('user_token')
  localStorage.removeItem('user_refresh_token')
  localStorage.removeItem('user_profile')
  if (refreshTimer) {
    clearTimeout(refreshTimer)
    refreshTimer = null
  }
  if (visibilityListener) {
    document.removeEventListener('visibilitychange', visibilityListener)
    visibilityListener = null
  }
  // Keep isRedirecting = true to prevent re-entry; reset after redirect completes
  setTimeout(() => { isRedirecting = false }, 100)
}

/**
 * Check if token is close to expiring and should be refreshed proactively
 */
function tokenExpiresSoon() {
  if (!lastRefreshTime) return false
  const elapsed = Date.now() - lastRefreshTime
  return elapsed >= (ACCESS_TOKEN_TTL - SAFE_REFRESH_MARGIN)
}

/**
 * Refresh the access token using the refresh token.
 * Uses a single promise queue — concurrent callers receive the same promise.
 * Returns the new token on success, null on failure.
 */
export function refreshAccessToken() {
  if (!refreshToken.value) return Promise.resolve(null)
  // If a refresh is already in progress, return the existing promise
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    try {
      const resp = await fetch('/api/v1/auth/refresh', {
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
      console.error('Token refresh failed:', err)
    }
    // If refresh failed, clear and redirect
    if (!isRedirecting) {
      isRedirecting = true
      clearTokens()
      const router = useRouter()
      router.push('/login')
    }
    return null
  })().finally(() => {
    refreshPromise = null
  })

  return refreshPromise
}

/**
 * Refresh token if needed (proactive check based on elapsed time).
 * Called before API calls and when page becomes visible.
 */
export function refreshIfNeeded() {
  if (!isLoggedIn() || !refreshToken.value) return Promise.resolve()
  if (tokenExpiresSoon()) {
    return refreshAccessToken()
  }
  return Promise.resolve()
}

/**
 * Start automatic token refresh timer.
 * Runs every 5 minutes while the page is open.
 */
export function startTokenRefresh() {
  if (refreshTimer) clearInterval(refreshTimer)
  if (!refreshToken.value) return

  refreshTimer = setInterval(() => {
    if (tokenExpiresSoon()) {
      refreshAccessToken()
    }
  }, REFRESH_INTERVAL)
}

/**
 * Handle page visibility change — refresh token when user returns to the page
 */
function handleVisibilityChange() {
  if (document.visibilityState === 'visible') {
    refreshIfNeeded()
  }
}

/**
 * Initialize token refresh on app startup
 */
export function initUserAuth() {
  if (refreshToken.value) {
    startTokenRefresh()
    // Listen for page visibility changes — refresh when user returns
    visibilityListener = handleVisibilityChange
    document.addEventListener('visibilitychange', visibilityListener)
  }
}

/**
 * Fetch user profile and cache it.
 * Automatically refreshes token if needed before the request.
 */
export async function fetchProfile() {
  if (!accessToken.value) return null
  await refreshIfNeeded()
  try {
    const resp = await fetch('/api/v1/auth/profile', {
      headers: { Authorization: `Bearer ${accessToken.value}` },
    })
    const json = await resp.json()
    if (json.code === 0) {
      currentUser.value = json.data
      localStorage.setItem('user_profile', JSON.stringify(json.data))
      return json.data
    }
  } catch (err) {
    console.error('Failed to fetch profile:', err)
  }
  return null
}

/**
 * Check if user is logged in
 */
export function isLoggedIn() {
  return !!accessToken.value
}

/**
 * Create a fetch wrapper with built-in token refresh on 401.
 * 
 * When a 401 is received, this queues behind the single refresh promise (no matter
 * how many requests fail concurrently), then retries with the new token once it's
 * available. This is the industry-standard approach used by Axios, OkHttp, etc.
 */
export function authFetch(url, options = {}) {
  const makeRequest = (token) => {
    const headers = { ...options.headers }
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    return fetch(url, { ...options, headers })
  }

  return refreshIfNeeded().then(() => makeRequest(accessToken.value)).then(async (resp) => {
    if (resp.status === 401 && refreshToken.value) {
      const newToken = await refreshAccessToken()
      if (newToken) {
        return makeRequest(newToken)
      }
    }
    return resp
  })
}

export { accessToken, refreshToken, currentUser }
