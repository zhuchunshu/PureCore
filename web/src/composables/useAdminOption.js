import { ref } from 'vue'

const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

// In-memory cache for options
const cache = ref({})
let loaded = false
let loading = false

/**
 * Fetch all options from the backend and cache them.
 * Called automatically on first use.
 */
async function loadOptions() {
  if (loaded || loading) return
  loading = true

  try {
    // Use the public /system/info-like endpoint — actually we need an authenticated endpoint
    // For public reading, we can access the admin options endpoint if it's public,
    // or create a separate public endpoint. For now, try fetching from the admin options API.
    const resp = await fetch(`/api/v1/${adminPrefix}/options`)
    if (resp.ok) {
      const json = await resp.json()
      if (json.code === 0 && json.data) {
        cache.value = { ...json.data }
      }
    } else {
      // If unauthorized (401), silently fail — options will remain empty
      console.warn('Could not load admin options (auth required)')
    }
  } catch (err) {
    console.warn('Failed to load admin options:', err)
  } finally {
    loaded = true
    loading = false
  }
}

/**
 * Get an option value by key. Returns the value or defaultVal if not found.
 * Automatically loads options from the backend on first call.
 *
 * @param {string} key - The option name
 * @param {any} defaultVal - Default value if the option is not found
 * @returns {Promise<any>} The option value
 */
export async function adminOption(key, defaultVal = '') {
  await loadOptions()
  const val = cache.value[key]
  return val !== undefined ? val : defaultVal
}

/**
 * Synchronous version — returns the cached value if already loaded, otherwise the default.
 * Use this in computed properties or templates where async is not possible.
 *
 * @param {string} key - The option name
 * @param {any} defaultVal - Default value if not yet loaded or not found
 * @returns {any} The cached value or default
 */
export function adminOptionSync(key, defaultVal = '') {
  if (!loaded) {
    // Trigger async load in background (non-blocking)
    loadOptions()
  }
  const val = cache.value[key]
  return val !== undefined ? val : defaultVal
}

/**
 * Refresh the options cache by re-fetching from the backend.
 */
export async function refreshOptions() {
  loaded = false
  loading = false
  await loadOptions()
}

/**
 * Set an option value on the backend (requires authentication).
 * Also updates the local cache if successful.
 *
 * @param {string} key - The option name
 * @param {string} value - The option value
 * @returns {Promise<boolean>} Whether the operation succeeded
 */
export async function adminOptionSet(key, value) {
  const token = typeof localStorage !== 'undefined' ? localStorage.getItem('admin_token') : null
  if (!token) {
    console.warn('Cannot set option: not authenticated')
    return false
  }

  try {
    const resp = await fetch(`/api/v1/${adminPrefix}/options`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify({ name: key, value }),
    })
    const json = await resp.json()
    if (json.code === 0) {
      // Update local cache
      cache.value = { ...cache.value, [key]: value }
      return true
    }
  } catch (err) {
    console.error('Failed to set option:', err)
  }
  return false
}

/**
 * Check if options have been loaded
 */
export function isOptionsLoaded() {
  return loaded
}

/**
 * Vue composable for reactive option access
 */
export function useAdminOption() {
  return {
    adminOption,
    adminOptionSync,
    adminOptionSet,
    refreshOptions,
    isOptionsLoaded,
    options: cache,
  }
}
