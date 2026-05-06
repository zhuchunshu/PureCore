<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { useOAuth } from '../../composables/useOAuth'
import { setTokens as setUserTokens } from '../../composables/useUserAuth'
import PureCoreLoading from '../../components/PureCoreLoading.vue'
import { IconExclamationCircle, IconCircleCheck, IconUserPlus, IconLogin2, IconLink, IconX } from '@tabler/icons-vue'

const { t } = useI18n()
useSEO({ title: t('oauth.callback_title'), description: t('oauth.callback_title') })

const route = useRoute()
const router = useRouter()
const { exchangeCode, bindOAuth } = useOAuth()

// Query parameters from backend redirect (old flow)
const status = ref(route.query.status || '')
const linkToken = ref(route.query.link_token || '')
const provider = ref(route.query.provider || route.params.provider || '')
const oauthName = ref(route.query.name || '')
const oauthEmail = ref(route.query.email || '')
const oauthAvatar = ref(route.query.avatar_url || '')
const redirectTo = ref(route.query.redirect || '/dashboard')
const currentUserName = ref(route.query.current_user_name || '')
const currentUserEmail = ref(route.query.current_user_email || '')

// New exchange flow parameters
const code = ref(route.query.code || '')
const state = ref(route.query.state || '')

const loading = ref(true)
const bindLoading = ref(false)
const error = ref('')
const bindError = ref('')
const mode = ref('choose')

onMounted(async () => {
  // New exchange flow: OAuth provider redirected back to frontend with code+state
  const isTelegramStateCallback = provider.value === 'telegram' && state.value && route.query.hash
  if ((code.value && state.value && !status.value) || (isTelegramStateCallback && !status.value)) {
    await handleExchange()
    return
  }

  // Old flow: backend redirected here with status query param
  if (status.value === 'linked') {
    // Backend already set tokens via cookies and redirected here
    readCookiesAndLogin()
  } else if (status.value === 'bound') {
    // OAuth was bound to an existing logged-in session (legacy, kept for backward compat)
    router.replace(redirectTo.value)
  } else if (status.value === 'logged_in' && linkToken.value) {
    // User is already logged in — show bind confirmation
    oauthName.value = route.query.name || ''
    oauthEmail.value = route.query.email || ''
    oauthAvatar.value = route.query.avatar_url || ''
    currentUserName.value = route.query.current_user_name || ''
    currentUserEmail.value = route.query.current_user_email || ''
    loading.value = false
    mode.value = 'bind_confirm'
  } else if (status.value === 'unlinked' && linkToken.value) {
    // New OAuth account — show registration/bind options
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
    const callbackData = provider.value === 'telegram'
      ? Object.fromEntries(
          Object.entries(route.query)
            .filter(([key, value]) => key !== 'state' && value !== undefined && value !== null)
            .map(([key, value]) => [key, String(value)])
        )
      : null
    const data = await exchangeCode(provider.value, code.value, state.value, callbackData)
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
        router.replace(data.redirect || redirectTo.value || '/dashboard')
      } else {
        error.value = t('oauth.callback_failed')
        loading.value = false
      }
      break
    case 'bound':
      // Bound to existing logged-in session (legacy, kept for backward compat)
      router.replace(data.redirect || redirectTo.value || '/dashboard')
      break
    case 'logged_in':
      // User is already logged in — show bind confirmation
      linkToken.value = data.link_token || ''
      provider.value = data.provider || provider.value
      oauthName.value = data.name || ''
      oauthEmail.value = data.email || ''
      oauthAvatar.value = data.avatar_url || ''
      redirectTo.value = data.redirect || redirectTo.value || '/dashboard'
      if (data.current_user) {
        currentUserName.value = data.current_user.name || ''
        currentUserEmail.value = data.current_user.email || ''
      }
      loading.value = false
      mode.value = 'bind_confirm'
      break
    case 'unlinked':
      // New OAuth account: show registration/bind options
      linkToken.value = data.link_token || ''
      provider.value = data.provider || provider.value
      oauthName.value = data.name || ''
      oauthEmail.value = data.email || ''
      oauthAvatar.value = data.avatar_url || ''
      redirectTo.value = data.redirect || redirectTo.value || '/dashboard'
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

// --- Bind confirmation (logged-in user) ---

async function bindAccount() {
  bindLoading.value = true
  bindError.value = ''
  try {
    await bindOAuth(linkToken.value)
    router.replace(redirectTo.value || '/dashboard')
  } catch (err) {
    bindError.value = err.message || t('oauth.bind_failed')
    bindLoading.value = false
  }
}

function cancelBind() {
  router.replace(redirectTo.value || '/dashboard')
}

// --- Unlinked account options ---

function chooseRegister() {
  // Store OAuth data in sessionStorage to avoid URL length limits with JWT tokens
  const oauthData = {
    link_token: linkToken.value,
    name: oauthName.value,
    email: oauthEmail.value,
    avatar_url: oauthAvatar.value,
    redirect: redirectTo.value,
  }
  try {
    sessionStorage.setItem('oauth_link_data', JSON.stringify(oauthData))
  } catch (_) { /* ignore quota errors */ }
  const providerName = route.params.provider || provider.value
  router.push(`/oauth/${providerName}/link/register`)
}

function chooseLogin() {
  // Store OAuth data in sessionStorage to avoid URL length limits with JWT tokens
  const oauthData = {
    link_token: linkToken.value,
    email: oauthEmail.value,
    redirect: redirectTo.value,
  }
  try {
    sessionStorage.setItem('oauth_link_data', JSON.stringify(oauthData))
  } catch (_) { /* ignore quota errors */ }
  const providerName = route.params.provider || provider.value
  router.push(`/oauth/${providerName}/link/login`)
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
      <!-- Branded loading state -->
      <PureCoreLoading v-if="loading" :text="t('oauth.callback_processing')" />

      <!-- Error state -->
      <div v-else-if="error" class="card bg-base-100/80 border border-base-300/20 shadow-xl p-8 text-center">
        <div class="text-error mb-4">
          <IconExclamationCircle :size="64" class="mx-auto" />
        </div>
        <h2 class="text-xl font-bold text-base-content mb-2">{{ t('oauth.callback_error') }}</h2>
        <p class="text-base-content/60 mb-6">{{ error }}</p>
        <a href="/login" class="btn btn-primary">{{ t('user.login_title') }}</a>
      </div>

      <!-- Bind confirmation mode (user is logged in, ask whether to bind) -->
      <div v-else-if="mode === 'bind_confirm'" class="card bg-base-100/80 border border-base-300/20 shadow-xl p-8">
        <div class="text-center mb-8">
          <div v-if="oauthAvatar" class="avatar mb-4">
            <div class="w-20 h-20 rounded-full ring ring-primary/20 ring-offset-base-100 ring-offset-2">
              <img :src="oauthAvatar" :alt="oauthName" />
            </div>
          </div>
          <div v-else class="mb-4">
            <IconCircleCheck :size="64" class="mx-auto text-success" />
          </div>
          <h2 class="text-xl font-bold text-base-content">{{ t('oauth.bind_confirm_title') }}</h2>
          <p class="text-base-content/60 mt-2">
            {{ t('oauth.bind_confirm_desc', { provider: providerDisplay, name: oauthName, email: oauthEmail }) }}
          </p>
          <p v-if="currentUserName" class="text-base-content/60 mt-1 text-sm">
            {{ t('oauth.bind_confirm_current_account') }}: {{ currentUserName }}
            <template v-if="currentUserEmail">({{ currentUserEmail }})</template>
          </p>
        </div>

        <div v-if="bindError" class="alert alert-error mb-4 text-sm">
          <IconExclamationCircle :size="18" />
          <span>{{ bindError }}</span>
        </div>

        <div class="space-y-3">
          <button
            class="btn btn-primary w-full gap-3"
            :disabled="bindLoading"
            @click="bindAccount"
          >
            <template v-if="bindLoading">
              <span class="loading loading-spinner-xs"></span>
            </template>
            <IconLink v-else :size="20" />
            {{ t('oauth.bind_confirm_yes') }}
          </button>
          <button
            class="btn btn-outline w-full gap-3"
            :disabled="bindLoading"
            @click="cancelBind"
          >
            <IconX :size="20" />
            {{ t('oauth.bind_confirm_no') }}
          </button>
        </div>
      </div>

      <!-- Choose action mode (unlinked OAuth account) -->
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
    </div>
  </div>
</template>
