<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- ===== COL 1-2: Main ===== -->
    <div class="lg:col-span-2 space-y-4">
    <!-- Header skeleton -->
    <template v-if="loading">
      <div class="skeleton h-8 w-48 rounded-lg"></div>
      <div class="skeleton h-24 rounded-2xl"></div>
      <div class="space-y-3">
        <div v-for="i in 4" :key="i" class="flex items-center gap-4 rounded-xl bg-base-100 p-4">
          <div class="skeleton w-10 h-10 rounded-xl shrink-0"></div>
          <div class="flex-1 space-y-2">
            <div class="skeleton h-5 w-1/3 rounded"></div>
            <div class="skeleton h-3 w-1/4 rounded"></div>
          </div>
          <div class="skeleton h-9 w-20 rounded-lg shrink-0"></div>
        </div>
      </div>
    </template>

    <!-- Content -->
    <template v-else>
      <div class="flex items-center justify-between mb-4">
        <p class="text-sm text-base-content/40">{{ t('integrations.description') }}</p>
        <span class="text-sm text-base-content/40">{{ accounts.length }} {{ t('integrations.account') }}</span>
      </div>

      <!-- Current login provider -->
      <div v-if="currentLoginProvider" class="bg-base-100 border border-emerald-500/10 rounded-2xl p-5 shadow-sm">
        <h3 class="text-sm font-semibold text-base-content/60 mb-3">{{ t('integrations.current_login') }}</h3>
        <div class="flex items-center gap-4">
          <div class="w-10 h-10 shrink-0 rounded-xl bg-emerald-500/10 flex items-center justify-center">
            <component :is="providerIcon(currentLoginProvider)" class="w-5 h-5 text-emerald-500" />
          </div>
          <div class="flex-1 min-w-0">
            <p class="font-semibold text-base-content">{{ providerDisplay(currentLoginProvider) }}</p>
            <p class="text-sm text-base-content/50">{{ t('integrations.current_login_desc') }}</p>
          </div>
          <span class="inline-flex items-center text-xs font-medium text-emerald-500 bg-emerald-500/10 px-2.5 py-1 rounded-full">
            {{ t('integrations.current_login') }}
          </span>
        </div>
      </div>

      <!-- Bound accounts -->
      <div v-if="accounts.length > 0">
        <h3 class="text-sm font-semibold text-base-content/60 mb-3">{{ t('integrations.linked_at') }}</h3>
        <div class="space-y-2">
          <div
            v-for="account in accounts"
            :key="account.id"
            class="flex items-center gap-4 bg-base-100 border border-base-300/20 rounded-2xl p-4 shadow-sm hover:shadow transition-shadow"
          >
            <div class="w-10 h-10 shrink-0 rounded-xl bg-primary/10 flex items-center justify-center">
              <component :is="providerIcon(account.provider)" class="w-5 h-5 text-primary" />
            </div>
            <div class="flex-1 min-w-0">
              <p class="font-semibold text-base-content text-sm">{{ providerDisplay(account.provider) }}</p>
              <p class="text-sm text-base-content/50 truncate">
                {{ account.email || account.name || account.provider_id }}
              </p>
              <p class="text-xs text-base-content/30">
                {{ t('integrations.linked_at') }}: {{ formatDate(account.updated_at || account.created_at) }}
              </p>
            </div>
            <button
              class="text-red-400 hover:text-red-500 hover:bg-red-500/10 rounded-lg px-3 py-2 text-sm font-medium transition-all duration-200 cursor-pointer disabled:opacity-40"
              :disabled="unlinkingId === account.id"
              @click="unlinkAccount(account)"
            >
              <span v-if="unlinkingId === account.id" class="loading loading-spinner loading-xs"></span>
              <Unlink v-else class="w-4 h-4" />
              <span class="hidden sm:inline ml-1.5">{{ t('integrations.unbind') }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- No accounts -->
      <div v-if="accounts.length === 0 && !currentLoginProvider" class="bg-base-100 border border-base-300/20 rounded-2xl shadow-sm">
        <div class="text-center py-16 px-6">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-base-200 flex items-center justify-center">
            <Link2 class="w-8 h-8 text-base-content/20" />
          </div>
          <p class="text-sm font-medium text-base-content/40">{{ t('integrations.no_accounts') }}</p>
          <p class="mt-1 text-sm text-base-content/30 max-w-md mx-auto">{{ t('integrations.no_accounts_desc') }}</p>
        </div>
      </div>

      <!-- Available providers to connect -->
      <div v-if="availableProviders.length > 0">
        <h3 class="text-sm font-semibold text-base-content/60 mb-3">{{ t('integrations.connect_more') }}</h3>
        <div class="space-y-2">
          <div
            v-for="provider in availableProviders"
            :key="provider.name"
            class="flex items-center gap-4 bg-base-100 border border-base-300/20 rounded-2xl p-4 shadow-sm hover:shadow transition-shadow"
          >
            <div class="w-10 h-10 shrink-0 rounded-xl bg-base-200 flex items-center justify-center">
              <component :is="providerIcon(provider.name)" class="w-5 h-5 text-base-content/40" />
            </div>
            <div class="flex-1 min-w-0">
              <p class="font-semibold text-base-content text-sm">{{ provider.display_name }}</p>
              <p class="text-sm text-base-content/50 truncate">{{ t('integrations.connect_desc', { provider: provider.display_name }) }}</p>
            </div>
            <button
              class="inline-flex items-center gap-1.5 border border-base-300/30 hover:border-primary/30 hover:bg-primary/5 rounded-xl px-4 py-2 text-sm font-medium transition-all duration-200 cursor-pointer disabled:opacity-40"
              :disabled="connectingProvider === provider.name"
              @click="connectProvider(provider)"
            >
              <span v-if="connectingProvider === provider.name" class="loading loading-spinner loading-xs"></span>
              <Link2 v-else class="w-4 h-4" />
              {{ t('integrations.connect') }}
            </button>
          </div>
        </div>
      </div>
    </template>
    </div>

    <!-- ===== COL 3: Sidebar ===== -->
    <div class="space-y-4">
      <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
        <h3 class="text-sm font-bold text-base-content flex items-center gap-2 mb-3">
          <Link2 :size="16" class="text-primary/60" />
          {{ t('user.oauth_info_title') }}
        </h3>
        <ul class="space-y-2 text-xs text-base-content/50">
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.oauth_info_link') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.oauth_info_unlink') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.oauth_info_highlight') }}
          </li>
        </ul>
      </div>
    </div>

    <!-- Confirm dialog -->
    <dialog ref="confirmDialog" class="modal">
      <div class="modal-box">
        <h3 class="text-lg font-bold">{{ t('integrations.unlink_confirm', { provider: confirmProvider }) }}</h3>
        <div class="modal-action">
          <button class="btn" @click="closeConfirmDialog">{{ t('admin.cancel') }}</button>
          <button class="btn btn-error" @click="doUnlink">
            <span v-if="unlinking" class="loading loading-spinner loading-xs"></span>
            {{ t('integrations.unbind') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeConfirmDialog">close</button>
      </form>
    </dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '../../i18n'
import { useRouter } from 'vue-router'
import { Link2, Unlink } from 'lucide-vue-next'
import { IconBrandGithub, IconBrandGoogle, IconBrandApple, IconBrandTelegram, IconBrandDiscord } from '@tabler/icons-vue'
import { userAPI } from '../../services/api'
import { useUserAuth } from '../../composables/useUserAuth'
import { useOAuth } from '../../composables/useOAuth'
import { useSEO } from '../../composables/useSEO'

const { t } = useI18n()
const router = useRouter()
const { accessToken } = useUserAuth()
const { providers, fetchProviders: fetchOAuthProviders, initiateLogin } = useOAuth()

const loading = ref(true)
const accounts = ref([])
const currentLoginProvider = ref('')
const connectingProvider = ref(null)
const confirmDialog = ref(null)
const confirmProvider = ref('')
const selectedAccount = ref(null)
const unlinking = ref(false)
const unlinkingId = ref(null)

useSEO({
  title: t('integrations.title'),
  description: t('integrations.description')
})

const providerIconMap = {
  github: IconBrandGithub,
  google: IconBrandGoogle,
  apple: IconBrandApple,
  telegram: IconBrandTelegram,
  discord: IconBrandDiscord,
}
function providerIcon(provider) {
  return providerIconMap[provider] || Link2
}

const providerDisplayMap = {
  github: 'GitHub',
  google: 'Google',
  apple: 'Apple',
  telegram: 'Telegram',
  discord: 'Discord',
}
function providerDisplay(provider) {
  return providerDisplayMap[provider] || provider.charAt(0).toUpperCase() + provider.slice(1)
}

const availableProviders = computed(() => {
  const linkedNames = new Set(accounts.value.map(a => a.provider))
  return providers.value.filter(p => p.enabled && !linkedNames.has(p.name))
})

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

async function fetchAccounts() {
  try {
    const resp = await userAPI.get('/api/v1/oauth/accounts')
    const json = await resp.json()
    if (json.code === 0) {
      accounts.value = json.data?.accounts || []
      currentLoginProvider.value = json.data?.current_login_provider || ''
    }
  } catch (err) {
    console.error('Failed to fetch OAuth accounts', err)
  }
}

async function connectProvider(provider) {
  if (connectingProvider.value) return
  connectingProvider.value = provider.name
  try {
    if (provider.name === 'telegram') {
      const widgetData = await initiateLogin(provider.name, '/dashboard/integrations')
      if (widgetData && widgetData.type === 'widget') {
        await triggerTelegramAuth(widgetData)
        return
      }
    }
    await initiateLogin(provider.name, '/dashboard/integrations')
  } catch (err) {
    console.error('Failed to initiate OAuth for', provider.name, err)
  } finally {
    connectingProvider.value = null
  }
}

function ensureTelegramScriptLoaded() {
  return new Promise((resolve, reject) => {
    if (typeof window === 'undefined') {
      reject(new Error('Telegram login unavailable in SSR'))
      return
    }
    if (window.Telegram?.Login?.auth) {
      resolve()
      return
    }
    const existing = document.getElementById('telegram-login-sdk')
    if (existing) {
      existing.addEventListener('load', () => resolve(), { once: true })
      existing.addEventListener('error', () => reject(new Error('Failed to load Telegram SDK')), { once: true })
      return
    }
    const script = document.createElement('script')
    script.id = 'telegram-login-sdk'
    script.src = 'https://telegram.org/js/telegram-widget.js?22'
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Failed to load Telegram SDK'))
    document.head.appendChild(script)
  })
}

function triggerTelegramAuth(widgetData) {
  const botId = String(widgetData?.bot_id || '').trim()
  const callbackBase = String(widgetData?.redirect_url || '').trim()
  const state = String(widgetData?.state || '').trim()
  if (!botId || !callbackBase || !state) {
    throw new Error('Telegram widget config incomplete')
  }

  return ensureTelegramScriptLoaded().then(() => {
    window.Telegram.Login.auth({ bot_id: botId, request_access: true }, (data) => {
      if (!data) return
      const callbackURL = new URL(callbackBase, window.location.origin)
      callbackURL.searchParams.set('state', state)
      Object.entries(data).forEach(([k, v]) => {
        if (v !== undefined && v !== null) callbackURL.searchParams.set(k, String(v))
      })
      window.location.href = callbackURL.toString()
    })
  })
}

function unlinkAccount(account) {
  selectedAccount.value = account
  confirmProvider.value = providerDisplay(account.provider)
  confirmDialog.value?.showModal()
}

function closeConfirmDialog() {
  confirmDialog.value?.close()
  selectedAccount.value = null
  confirmProvider.value = ''
}

async function doUnlink() {
  const account = selectedAccount.value
  if (!account) return
  unlinking.value = true
  unlinkingId.value = account.id
  try {
    const resp = await userAPI.delete(`/api/v1/oauth/accounts/${account.id}`)
    const json = await resp.json()
    if (json.code === 0) {
      accounts.value = accounts.value.filter(a => a.id !== account.id)
      if (currentLoginProvider.value === account.provider) {
        currentLoginProvider.value = ''
      }
    } else {
      alert(json.message || t('integrations.unlink_failed'))
    }
  } catch (err) {
    alert(t('integrations.unlink_failed'))
  } finally {
    unlinking.value = false
    unlinkingId.value = null
    closeConfirmDialog()
  }
}

onMounted(async () => {
  if (!accessToken.value) {
    loading.value = false
    router.push('/login')
    return
  }
  try {
    await Promise.all([fetchAccounts(), fetchOAuthProviders()])
  } finally {
    loading.value = false
  }
})
</script>
