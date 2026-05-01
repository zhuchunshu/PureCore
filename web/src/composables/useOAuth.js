import { ref } from 'vue'
import { setTokens } from './useUserAuth'

const providers = ref([])
const loaded = ref(false)

export function useOAuth() {
  async function fetchProviders() {
    if (loaded.value) return providers.value
    try {
      const resp = await fetch('/api/v1/oauth/providers')
      const json = await resp.json()
      if (json.code === 0) {
        providers.value = json.data || []
      }
      loaded.value = true
    } catch {
      providers.value = []
      loaded.value = true
    }
    return providers.value
  }

  function getProviderUrl(provider) {
    return `/api/v1/oauth/${provider}`
  }

  /**
   * Handle OAuth callback — called on the /dashboard page when tokens
   * are passed via URL query parameters (e.g. ?token=xxx&refresh_token=xxx&name=xxx)
   */
  function handleOAuthCallback() {
    const params = new URLSearchParams(window.location.search)
    const token = params.get('token')
    const refreshToken = params.get('refresh_token')
    const name = params.get('name')
    const email = params.get('email')

    if (token && refreshToken) {
      setTokens(token, refreshToken)
      if (name && email) {
        localStorage.setItem('user_profile', JSON.stringify({ name, email }))
      }
      // Clean up URL
      const url = new URL(window.location)
      url.searchParams.delete('token')
      url.searchParams.delete('refresh_token')
      url.searchParams.delete('name')
      url.searchParams.delete('email')
      window.history.replaceState({}, '', url.toString())
      return true
    }
    return false
  }

  return {
    providers,
    loaded,
    fetchProviders,
    getProviderUrl,
    handleOAuthCallback,
  }
}
