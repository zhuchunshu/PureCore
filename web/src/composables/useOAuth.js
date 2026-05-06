import { ref } from 'vue'
import { userAPI } from '../services/api'

// Simple unauthenticated JSON request helper for public OAuth endpoints.
async function publicApi(url, options = {}) {
  const resp = await fetch(url, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  })
  const json = await resp.json()
  if (json.code !== 0 && json.code !== undefined) {
    throw new Error(json.message || 'Request failed')
  }
  return json
}

// Shared singleton state to avoid duplicating refs/functions for each component
// instance (e.g. multiple OAuth buttons on auth pages).
const providers = ref([])
const loading = ref(false)
const error = ref('')

// Fetch all OAuth providers and their enabled status
async function fetchProviders() {
  loading.value = true
  error.value = ''
  try {
    const json = await publicApi('/api/v1/oauth/providers')
    providers.value = json.data || []
  } catch (e) {
    error.value = e.message || 'Failed to load OAuth providers'
  } finally {
    loading.value = false
  }
}

// Initiate OAuth login: ask backend for authorization URL, then redirect.
// For non-OAuth2 providers (e.g., Telegram), the backend returns widget config
// instead of a URL — in that case, return the data for the caller to handle.
async function initiateLogin(providerName, redirect = '/') {
  const json = await publicApi(`/api/v1/oauth/${providerName}/authorize?redirect=${encodeURIComponent(redirect)}`)
  if (json.data) {
    // Non-OAuth2 provider (widget type) — return config for frontend widget
    if (json.data.type === 'widget') {
      return json.data
    }
    // OAuth2 provider — redirect to auth URL
    if (json.data.url) {
      window.location.href = json.data.url
    } else {
      throw new Error('No authorization URL received')
    }
  } else {
    throw new Error('No authorization URL received')
  }
}

// Exchange the authorization code returned by the OAuth provider with the backend.
// This is called by the frontend OAuth callback page after receiving code + state.
// Includes the auth token if user is already logged in, so the backend can detect
// a logged-in user and return logged_in status for bind confirmation.
async function exchangeCode(providerName, code, state, callbackData = null) {
  const payload = callbackData
    ? { state, data: callbackData }
    : { code, state }
  const headers = { 'Content-Type': 'application/json' }
  if (userAPI.accessToken && userAPI.accessToken.value) {
    headers['Authorization'] = `Bearer ${userAPI.accessToken.value}`
  }
  const json = await publicApi(`/api/v1/oauth/${providerName}/exchange`, {
    method: 'POST',
    headers,
    body: JSON.stringify(payload),
  })
  return json.data || {}
}

// Complete OAuth register: send link token and user info to create account
async function oauthRegister(linkToken, name, email) {
  const resp = await fetch('/api/v1/oauth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ link_token: linkToken, name, email }),
  })
  const json = await resp.json()
  if (json.code !== 0) throw new Error(json.message || 'Registration failed')
  return json.data
}

// Bind OAuth account to currently logged-in user (requires auth, uses userAPI)
async function bindOAuth(linkToken) {
  const resp = await userAPI.post('/api/v1/oauth/bind', { link_token: linkToken })
  const json = await resp.json()
  if (json.code !== 0) throw new Error(json.message || 'Bind failed')
  return json
}

// Fetch linked OAuth accounts for current user (requires auth)
async function fetchAccounts() {
  const resp = await userAPI.get('/api/v1/oauth/accounts')
  const json = await resp.json()
  return json.data || []
}

// Unlink an OAuth account (requires auth)
async function unlinkAccount(accountId) {
  const resp = await userAPI.delete(`/api/v1/oauth/accounts/${accountId}`)
  const json = await resp.json()
  if (json.code !== 0) throw new Error(json.message || 'Unlink failed')
  return json
}

// Composable for OAuth flows: fetching providers, initiating login,
// handling callbacks, binding/unbinding accounts, and OAuth registration.
export function useOAuth() {
  return {
    providers,
    loading,
    error,
    fetchProviders,
    initiateLogin,
    exchangeCode,
    oauthRegister,
    bindOAuth,
    fetchAccounts,
    unlinkAccount,
  }
}
