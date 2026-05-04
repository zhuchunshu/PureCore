<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { useOAuth } from '../../composables/useOAuth'
import { setTokens as setUserTokens } from '../../composables/useUserAuth'
import { IconExclamationCircle, IconCircleCheck, IconUserPlus, IconLogin2, IconLoader2 } from '@tabler/icons-vue'

const { t } = useI18n()
useSEO({ title: t('oauth.callback_title'), description: t('oauth.callback_title') })

const route = useRoute()
const router = useRouter()
const { oauthRegister, bindOAuth, exchangeCode } = useOAuth()

// Query parameters from backend redirect (old flow)
const status = ref(route.query.status || '')
const linkToken = ref(route.query.link_token || '')
const provider = ref(route.query.provider || route.params.provider || '')
const oauthName = ref(route.query.name || '')
const oauthEmail = ref(route.query.email || '')
const oauthAvatar = ref(route.query.avatar_url || '')
const redirectTo = ref(route.query.redirect || '/')

// New exchange flow parameters
const code = ref(route.query.code || '')
const state = ref(route.query.state || '')

const loading = ref(true)
const error = ref('')
const mode = ref('choose') // 'choose' | 'register' | 'login'
const registerName = ref('')
const registerEmail = ref('')
const registerLoading = ref(false)
const registerError = ref('')

onMounted(async () => {
  // Check for bind mode first (coming from login page with mode=bind)
  if (route.query.mode === 'bind') {
    const storedToken = typeof window !== 'undefined' ? sessionStorage.getItem('oauth_link_token') : null
    if (storedToken) {
      linkToken.value = storedToken
      try {
        await bindOAuth(storedToken)
        if (typeof window !== 'undefined') {
          sessionStorage.removeItem('oauth_link_token')
        }
        router.replace(redirectTo.value || '/')
        return
      } catch (err) {
        error.value = err.message || t('oauth.bind_failed')
        loading.value = false
        return
      }
    }
    // If no stored token, fall through — this shouldn't happen but show error
    error.value = t('oauth.bind_failed')
    loading.value = false
    return
  }

  // New exchange flow: OAuth provider redirected back to frontend with code+state
  if (code.value && state.value && !status.value) {
    await handleExchange()
    return
  }

  // Old flow: backend redirected here with status query param
  if (status.value === 'linked') {
    // Backend already set tokens via cookies and redirected here
    readCookiesAndLogin()
  } else if (status.value === 'bound') {
    // OAuth was bound to an existing logged-in session
    router.replace(redirectTo.value)
  } else if (status.value === 'unlinked' && linkToken.value) {
    // New OAuth account — show registration/bind options
    registerName.value = oauthName.value || ''
    registerEmail.value = oauthEmail.value || ''
    loading.value = false
  } else {
    error.value = t('oauth.invalid_callback')
    loading.value = false
  }
})

async function handleExchange() {
  loading.value = true
  error.value = ''
  try {
    const data = await exchangeCode(provider.value, code.value, state.value)
    processExchangeResult(data)
  } catch (err) {
    error.value = err.message || t('oauth.callback_failed')
    loading.value = false
  }
}

function processExchangeResult(data) {
  if (!data) {
    error.value = t('oauth.invalid_callback')
    loading.value = false
    return
  }

  switch (data.status) {
    case 'linked':
      // Already linked: tokens returned, log user in
      if (data.token) {
        setUserTokens(data.token, data.refresh_token || '')
        if (data.name && data.email) {
          localStorage.setItem('user_profile', JSON.stringify({ name: data.name, email: data.email }))
        }
        router.replace(data.redirect || redirectTo.value || '/')
      } else {
        error.value = t('oauth.callback_failed')
        loading.value = false
      }
      break
    case 'bound':
      // Bound to existing logged-in session
      router.replace(data.redirect || redirectTo.value || '/')
      break
    case 'unlinked':
      // New OAuth account: show registration/bind options
      linkToken.value = data.link_token || ''
      provider.value = data.provider || provider.value
      oauthName.value = data.name || ''
      oauthEmail.value = data.email || ''
      oauthAvatar.value = data.avatar_url || ''
      redirectTo.value = data.redirect || redirectTo.value || '/'
      registerName.value = data.name || ''
      registerEmail.value = data.email || ''
      loading.value = false
      mode.value = 'choose'
      break
    default:
      error.value = t('oauth.invalid_callback')
      loading.value = false
  }
}

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
  const callbackPath = '/oauth/' + (route.params.provider || provider.value) + '/callback'
  router.replace({ path: '/login', query: { redirect: callbackPath + '?mode=bind&redirect=' + encodeURIComponent(redirectTo.value) } })
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

const providerDisplay = computed(() => {
  const map = {
    github: 'GitHub',
    google: 'Google',
    apple: 'Apple',
    telegram: 'Telegram',
    discord: 'Discord',
  }
  return map[provider.value] || provider.value || 'OAuth'
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
          <IconExclamationCircle :size="64" class="mx-auto" />
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
            <IconCircleCheck :size="64" class="mx-auto text-success" />
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
            <IconUserPlus :size="20" />
            {{ t('oauth.register_with_provider', { provider: providerDisplay }) }}
          </button>
          <button
            class="btn btn-outline w-full gap-3"
            @click="chooseLogin"
          >
            <IconLogin2 :size="20" />
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
            <IconExclamationCircle :size="20" />
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
