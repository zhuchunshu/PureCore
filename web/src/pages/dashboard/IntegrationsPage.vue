<template>
  <div class="flex-1 bg-base-200">
    <!-- Header skeleton -->
    <template v-if="loading">
      <div class="skeleton h-28 rounded-2xl"></div>
      <div class="mt-6 space-y-4">
        <div v-for="i in 4" :key="i" class="flex items-center gap-4 rounded-xl bg-base-100/80 p-4">
          <div class="skeleton w-12 h-12 rounded-xl shrink-0"></div>
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
      <!-- Page header -->
      <div class="flex items-center justify-between rounded-2xl bg-base-100/80 p-6 shadow-sm border border-base-300/20">
        <div>
          <h1 class="text-2xl font-bold">{{ t('integrations.title') }}</h1>
          <p class="mt-1 text-sm text-base-content/60">{{ t('integrations.description') }}</p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-sm text-base-content/50">{{ accounts.length }} {{ accounts.length === 1 ? t('integrations.account') : t('integrations.account') }}</span>
        </div>
      </div>

      <!-- No accounts -->
      <div v-if="accounts.length === 0" class="mt-6 text-center py-12 rounded-2xl bg-base-100/50 border border-base-300/20">
        <div class="mx-auto w-16 h-16 rounded-2xl bg-base-200/80 flex items-center justify-center mb-4">
          <Link2 class="w-8 h-8 text-base-content/30" />
        </div>
        <p class="font-semibold text-base-content/60">{{ t('integrations.no_accounts') }}</p>
        <p class="mt-1 text-sm text-base-content/40 max-w-md mx-auto">{{ t('integrations.no_accounts_desc') }}</p>
      </div>

      <!-- Accounts list -->
      <div v-else class="mt-6 space-y-4">
        <div
          v-for="account in accounts"
          :key="account.id"
          class="flex items-center gap-4 rounded-xl bg-base-100/80 p-4 border border-base-300/20 shadow-sm hover:shadow-md transition-shadow"
        >
          <!-- Provider icon -->
          <div class="w-12 h-12 shrink-0 rounded-xl bg-primary/10 flex items-center justify-center">
            <component :is="providerIcon(account.provider)" class="w-6 h-6 text-primary" />
          </div>

          <!-- Account info -->
          <div class="flex-1 min-w-0">
            <p class="font-semibold truncate">{{ providerDisplay(account.provider) }}</p>
            <p class="text-sm text-base-content/60 truncate">
              {{ account.email || account.name || account.provider_id }}
            </p>
            <p class="text-xs text-base-content/40">
              {{ t('integrations.linked_at') }}: {{ formatDate(account.created_at) }}
            </p>
          </div>

          <!-- Unlink button -->
          <button
            class="btn btn-ghost btn-sm text-error shrink-0"
            :disabled="unlinkingId === account.id"
            @click="unlinkAccount(account)"
          >
            <span v-if="unlinkingId === account.id" class="loading loading-spinner loading-xs"></span>
            <Unlink v-else class="w-4 h-4" />
            <span class="hidden sm:inline ml-1">{{ t('user.delete') }}</span>
          </button>
        </div>
      </div>
    </template>

    <!-- Confirm dialog -->
    <dialog ref="confirmDialog" class="modal">
      <div class="modal-box">
        <h3 class="text-lg font-bold">{{ t('integrations.unlink_confirm', { provider: confirmProvider }) }}</h3>
        <div class="modal-action">
          <button class="btn" @click="closeConfirmDialog">{{ t('admin.cancel') }}</button>
          <button class="btn btn-error" @click="doUnlink">
            <span v-if="unlinking" class="loading loading-spinner loading-xs"></span>
            {{ t('user.delete') }}
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
import { ref, onMounted } from 'vue'
import { useI18n } from '../../i18n'
import { useRouter } from 'vue-router'
import { Link2, Unlink } from 'lucide-vue-next'
import { IconBrandGithub, IconBrandGoogle, IconBrandApple, IconBrandTelegram, IconBrandDiscord } from '@tabler/icons-vue'
import { userAPI } from '../../services/api'
import { useUserAuth } from '../../composables/useUserAuth'
import { useSEO } from '../../composables/useSEO'

const { t } = useI18n()
const router = useRouter()
const { accessToken } = useUserAuth()

const loading = ref(true)
const accounts = ref([])
const confirmDialog = ref(null)
const confirmProvider = ref('')
const selectedAccount = ref(null)
const unlinking = ref(false)
const unlinkingId = ref(null)

useSEO({
  title: t('integrations.title'),
  description: t('integrations.description')
})

// Provider icon mapping
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

// Provider display name mapping
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
      accounts.value = json.data || []
    }
  } catch (err) {
    console.error('Failed to fetch OAuth accounts', err)
  }
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
    await fetchAccounts()
  } finally {
    loading.value = false
  }
})
</script>
