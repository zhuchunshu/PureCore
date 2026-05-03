<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { useOAuth } from '../../composables/useOAuth'
import { setTokens as setUserTokens } from '../../composables/useUserAuth'

const { t } = useI18n()
useSEO({ title: t('oauth.callback_title'), description: t('oauth.callback_title') })

const route = useRoute()
const router = useRouter()
const { oauthRegister, bindOAuth } = useOAuth()

// Query parameters from backend redirect
const status = ref(route.query.status || '')
const linkToken = ref(route.query.link_token || '')
const provider = ref(route.query.provider || '')
const oauthName = ref(route.query.name || '')
const oauthEmail = ref(route.query.email || '')
const oauthAvatar = ref(route.query.avatar_url || '')
const redirectTo = ref(route.query.redirect || '/')

const loading = ref(true)
const error = ref('')
const mode = ref('choose') // 'choose' | 'register' | 'login'
const registerName = ref('')
const registerEmail = ref('')
const registerLoading = ref(false)
const registerError = ref('')

onMounted(async () => {
  // Status is set by backend as a query param after the OAuth redirect
  if (status.value === 'linked') {
    // Backend already set tokens via cookies and redirected here
    // We need to read cookies and save to localStorage
    readCookiesAndLogin()
  } else if (status.value === 'bound') {
    // OAuth was bound to an existing logged-in session
    // Just redirect
    router.replace(redirectTo.value)
  } else if (status.value === 'unlinked' && linkToken.value) {
    // New OAuth account — show registration/bind options
    loading.value = false
    registerName.value = oauthName.value || ''
    registerEmail.value = oauthEmail.value || ''
  } else {
    error.value = t('oauth.invalid_callback')
    loading.value = false
  }
})

function readCookiesAndLogin() {
  try {
    const cookies = document.cookie.split(';').reduce((acc, cookie) => {
      const [key, val] = cookie.trim().split('=')
      if (key) acc[key] = val
      return acc
    }, {})
    if (cookies.access_token) {
      setUserTokens(cookies.access_token, cookies.refresh_token || '')
      router.replace(redirectTo.value)
      return
    }
  } catch (_) { /* ignore */ }
  error.value = t('oauth.callback_failed')
  loading.value = false
}

function chooseRegister() {
  mode.value = 'register'
}

function chooseLogin() {
  // User wants to log in first, then bind. Store link token and redirect to login.
  if (typeof window !== 'undefined') {
    sessionStorage.setItem('oauth_link_token', linkToken.value)
  }
  router.replace({ path: '/login', query: { redirect: '/oauth/callback?mode=bind&redirect=' + encodeURIComponent(redirectTo.value) } })
}

async function handleOAuthRegister() {
  if (!registerName.value || !registerEmail.value) {
    registerError.value = t('user.enter_credentials')
    return
  }
  registerLoading.value = true
  registerError.value = ''
  try {
    const data = await oauthRegister(linkToken.value, registerName.value, registerEmail.value)
    setUserTokens(data.token, data.refresh_token)
    if (data.name && data.email) {
      localStorage.setItem('user_profile', JSON.stringify({ name: data.name, email: data.email }))
    }
    router.replace(redirectTo.value)
  } catch (err) {
    registerError.value = err.message || t('oauth.register_failed')
  } finally {
    registerLoading.value = false
  }
}

function goBack() {
  mode.value = 'choose'
}

// Check for bind mode after login
onMounted(async () => {
  if (route.query.mode === 'bind') {
    const storedToken = typeof window !== 'undefined' ? sessionStorage.getItem('oauth_link_token') : null
    if (storedToken) {
      linkToken.value = storedToken
      loading.value = true
      try {
        await bindOAuth(storedToken)
        sessionStorage.removeItem('oauth_link_token')
        router.replace(redirectTo.value || '/')
      } catch (err) {
        error.value = err.message || t('oauth.bind_failed')
        loading.value = false
      }
    }
  }
})

const providerDisplay = computed(() => {
  if (provider.value === 'github') return 'GitHub'
  if (provider.value === 'google') return 'Google'
  return provider.value || 'OAuth'
})
</script>

<template>
  <div class="relative flex min-h-[calc(100vh-4rem)] items-center justify-center overflow-hidden bg-base-200">
    <!-- Animated grid background -->
    <div class="absolute inset-0 opacity-10">
      <div class="absolute inset-0" style="background-image: radial-gradient(circle at 1px 1px, oklch(var(--p)/0.15) 1px, transparent 0); background-size: 40px 40px;"></div>
    </div>

    <!-- Glow orbs -->
    <div class="absolute top-1/4 -left-20 w-72 h-72 bg-primary/20 rounded-full blur-3xl animate-pulse"></div>
    <div class="absolute bottom-1/4 -right-20 w-96 h-96 bg-secondary/15 rounded-full blur-3xl animate-pulse" style="animation-delay: 2s;"></div>

    <div class="relative z-10 w-full max-w-lg mx-auto px-4 py-8">
      <!-- Loading skeleton -->
      <template v-if="loading">
        <div class="skeleton h-28 rounded-2xl mb-8"></div>
        <div class="skeleton h-12 w-2/3 rounded-xl mb-4"></div>
        <div class="skeleton h-12 w-1/2 rounded-xl mb-4"></div>
        <div class="skeleton h-48 rounded-2xl"></div>
      </template>

      <!-- Error state -->
      <div v-else-if="error" class="card bg-base-100/80 border border-base-300/20 shadow-xl p-8 text-center">
        <div class="text-error mb-4">
          <svg class="w-16 h-16 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z"/>
          </svg>
        </div>
        <h2 class="text-xl font-bold text-base-content mb-2">{{ t('oauth.callback_error') }}</h2>
        <p class="text-base-content/60 mb-6">{{ error }}</p>
        <a href="/login" class="btn btn-primary">{{ t('user.login_title') }}</a>
      </div>

      <!-- Choose action mode -->
      <div v-else-if="mode === 'choose'" class="card bg-base-100/80 border border-base-300/20 shadow-xl p-8">
        <div class="text-center mb-8">
          <div v-if="oauthAvatar" class="avatar mb-4">
            <div class="w-20 h-20 rounded-full ring ring-primary/20 ring-offset-base-100 ring-offset-2">
              <img :src="oauthAvatar" :alt="oauthName" />
            </div>
          </div>
          <div v-else class="mb-4">
            <svg class="w-16 h-16 mx-auto text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <h2 class="text-xl font-bold text-base-content">{{ t('oauth.choose_action') }}</h2>
          <p class="text-base-content/60 mt-2">
            {{ t('oauth.choose_action_desc', { provider: providerDisplay, name: oauthName, email: oauthEmail }) }}
          </p>
        </div>

        <div class="space-y-4">
          <button
            class="btn btn-primary w-full gap-3"
            @click="chooseRegister"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"/>
            </svg>
            {{ t('oauth.register_with_provider', { provider: providerDisplay }) }}
          </button>
          <button
            class="btn btn-outline w-full gap-3"
            @click="chooseLogin"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1"/>
            </svg>
            {{ t('oauth.login_then_bind') }}
          </button>
        </div>
      </div>

      <!-- Register form mode -->
      <div v-else-if="mode === 'register'" class="card bg-base-100/80 border border-base-300/20 shadow-xl p-8">
        <div class="text-center mb-8">
          <h2 class="text-xl font-bold text-base-content">{{ t('oauth.complete_registration') }}</h2>
          <p class="text-base-content/60 mt-2">{{ t('oauth.complete_registration_desc', { provider: providerDisplay }) }}</p>
        </div>

        <form @submit.prevent="handleOAuthRegister" class="space-y-5">
          <div>
            <label class="block text-sm font-medium text-base-content/70 mb-2 ml-1">{{ t('user.name') }}</label>
            <input
              v-model="registerName"
              type="text"
              :placeholder="t('user.name_placeholder')"
              class="input input-bordered w-full"
              required
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-base-content/70 mb-2 ml-1">{{ t('user.email') }}</label>
            <input
              v-model="registerEmail"
              type="email"
              :placeholder="t('user.email_placeholder')"
              class="input input-bordered w-full"
              required
            />
          </div>

          <div v-if="registerError" class="alert alert-error text-sm">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z"/></svg>
            <span>{{ registerError }}</span>
          </div>

          <div class="flex gap-3">
            <button type="button" class="btn btn-outline flex-1" @click="goBack">
              {{ t('admin.back') }}
            </button>
            <button type="submit" class="btn btn-primary flex-1" :disabled="registerLoading">
              <span v-if="registerLoading" class="loading loading-spinner loading-xs"></span>
              {{ t('user.create_account') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
