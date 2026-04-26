import { ref, onMounted, onUnmounted } from 'vue'

const isBackendReachable = ref(false)  // Start as false - assume unreachable until proven
const hasChecked = ref(false)          // Has the first check completed?
let interval = null

export function useBackendHealth() {
  const checkHealth = async () => {
    try {
      const response = await fetch('/api/v1/ping', { method: 'GET', cache: 'no-store' })
      isBackendReachable.value = response.ok
    } catch {
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
