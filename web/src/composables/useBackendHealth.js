import { ref, onMounted, onUnmounted } from 'vue'

const isBackendReachable = ref(false)  // Start as false - assume unreachable until proven
const hasChecked = ref(false)          // Has the first check completed?
let interval = null

// Timeout in ms for the health check request (10 seconds)
const HEALTH_CHECK_TIMEOUT = 10000

export function useBackendHealth() {
  const checkHealth = async () => {
    try {
      // Always use the relative proxy path so the request goes through
      // the Vite dev proxy / SSR proxy / reverse proxy, never directly
      // to the backend. This tests the full stack reachability.
      const controller = typeof AbortController !== 'undefined' ? new AbortController() : null
      const timeoutId = controller ? setTimeout(() => controller.abort(), HEALTH_CHECK_TIMEOUT) : null
      try {
        const response = await fetch(`/api/v1/ping`, {
          method: 'GET',
          cache: 'no-store',
          signal: controller?.signal,
        })
        isBackendReachable.value = response.ok
      } finally {
        if (timeoutId) clearTimeout(timeoutId)
      }
    } catch {
      // This also catches AbortError from the timeout — treat as unreachable
      isBackendReachable.value = false
    } finally {
      if (!hasChecked.value) {
        hasChecked.value = true
      }
    }
  }

  const startHealthCheck = (intervalMs = 100000) => {
    if (interval) clearInterval(interval)
    checkHealth() // immediate check
    interval = setInterval(checkHealth, intervalMs)
  }

  const stopHealthCheck = () => {
    if (interval) {
      clearInterval(interval)
      interval = null
    }
  }

  onMounted(() => {
    startHealthCheck()
  })

  onUnmounted(() => {
    stopHealthCheck()
  })

  return {
    isBackendReachable,
    hasChecked,
    checkHealth,
    startHealthCheck,
    stopHealthCheck,
  }
}
