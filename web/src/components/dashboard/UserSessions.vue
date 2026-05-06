<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../../i18n'
import { userAPI } from '../../services/api'
import { useToast } from '../../composables/useToast'
import { IconBrandGithub, IconBrandGoogle, IconBrandApple, IconBrandTelegram, IconBrandDiscord } from '@tabler/icons-vue'
import {
  Smartphone, Tablet, Monitor,
  Globe, Compass,
  Apple, Terminal,
  Trash2, ShieldX, Key, ShieldCheck
} from 'lucide-vue-next'

const { t } = useI18n()
const { toastSuccess, toastError } = useToast()

const sessions = ref([])
const loading = ref(true)
const revokingId = ref(null)
const revokingAll = ref(false)

const deviceIcon = (type) => {
  switch (type) {
    case 'mobile': return Smartphone
    case 'tablet': return Tablet
    default: return Monitor
  }
}

const browserIcon = (browser) => {
  switch (browser?.toLowerCase()) {
    case 'safari': return Compass
    default: return Globe
  }
}

const osIcon = (os) => {
  switch (os?.toLowerCase()) {
    case 'macos':
    case 'ios': return Apple
    case 'android': return Smartphone
    case 'linux': return Terminal
    default: return Monitor
  }
}

const deviceModelDisplay = (session) => {
  const parts = []
  if (session.device_brand) parts.push(session.device_brand)
  if (session.device_model) parts.push(session.device_model)
  return parts.length > 0 ? parts.join(' ') : t('user.session_unknown_device')
}

const browserDisplay = (session) => {
  return session.browser || t('user.session_unknown_browser')
}

const osDisplay = (session) => {
  return session.os || t('user.session_unknown_os')
}

const providerIconMap = {
  github: IconBrandGithub,
  google: IconBrandGoogle,
  apple: IconBrandApple,
  telegram: IconBrandTelegram,
  discord: IconBrandDiscord,
}
const providerDisplayMap = {
  github: 'GitHub',
  google: 'Google',
  apple: 'Apple',
  telegram: 'Telegram',
  discord: 'Discord',
}

const loginProviderIcon = (provider) => {
  return providerIconMap[provider] || Key
}

const loginProviderDisplay = (provider) => {
  return providerDisplayMap[provider] || provider
}

const tp = (key, replacements) => {
  let text = t(key)
  for (const [k, v] of Object.entries(replacements)) {
    text = text.replace(`{${k}}`, v)
  }
  return text
}

const formatRelativeTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = date - now
  const absDiff = Math.abs(diff)

  if (absDiff < 60000) return t('user.just_now')

  if (diff > 0) {
    const hours = Math.ceil(absDiff / 3600000)
    if (hours < 48) return tp('user.in_hours', { n: hours })
    const days = Math.ceil(absDiff / 86400000)
    return tp('user.in_days', { n: days })
  }

  if (absDiff < 3600000) return Math.floor(absDiff / 60000) + 'm ago'
  if (absDiff < 86400000) return Math.floor(absDiff / 3600000) + 'h ago'
  return date.toLocaleDateString()
}

const fetchSessions = async () => {
  loading.value = true
  try {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
    const resp = await userAPI.get(`${baseUrl}/api/v1/sessions`)
    const json = await resp.json()
    if (json.code === 0) {
      sessions.value = json.data || []
    } else {
      toastError(json.message || t('user.session_revoke_failed'))
    }
  } catch (err) {
    console.error('Failed to fetch sessions:', err)
    toastError(t('user.network_error'))
  } finally {
    loading.value = false
  }
}

const revokeSession = async (id) => {
  if (revokingId.value) return
  if (!window.confirm(t('user.session_revoke_confirm'))) return

  revokingId.value = id
  try {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
    const resp = await userAPI.delete(`${baseUrl}/api/v1/sessions/${id}`)
    const json = await resp.json()
    if (json.code === 0) {
      sessions.value = sessions.value.filter(s => s.id !== id)
      toastSuccess(t('user.session_revoke_success'))
    } else {
      toastError(json.message || t('user.session_revoke_failed'))
    }
  } catch (err) {
    console.error('Failed to revoke session:', err)
    toastError(t('user.session_revoke_failed'))
  } finally {
    revokingId.value = null
  }
}

const revokeAll = async () => {
  if (revokingAll.value) return
  if (!window.confirm(t('user.session_revoke_all_confirm'))) return

  revokingAll.value = true
  try {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
    const resp = await userAPI.delete(`${baseUrl}/api/v1/sessions`)
    const json = await resp.json()
    if (json.code === 0) {
      sessions.value = sessions.value.filter(s => s.is_current)
      toastSuccess(t('user.session_revoke_all_success'))
    } else {
      toastError(json.message || t('user.session_revoke_all_failed'))
    }
  } catch (err) {
    console.error('Failed to revoke all sessions:', err)
    toastError(t('user.session_revoke_all_failed'))
  } finally {
    revokingAll.value = false
  }
}

onMounted(fetchSessions)
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- ===== COL 1-2: Main ===== -->
    <div class="lg:col-span-2 space-y-4">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <p class="text-sm text-base-content/40">{{ t('user.sessions_description') }}</p>
      <button
        class="inline-flex items-center gap-2 text-red-400 hover:text-red-500 hover:bg-red-500/10 rounded-xl px-4 py-2.5 text-sm font-medium transition-all duration-200 cursor-pointer disabled:opacity-40"
        :disabled="revokingAll || sessions.filter(s => !s.is_current).length === 0"
        @click="revokeAll"
      >
        <span v-if="revokingAll" class="loading loading-spinner loading-xs"></span>
        <ShieldX v-else :size="15" />
        {{ t('user.revoke_all') }}
      </button>
    </div>

    <!-- Card -->
    <div class="bg-base-100 border border-base-300/20 rounded-2xl shadow-sm overflow-hidden">
      <div class="p-6">
        <!-- Loading skeleton -->
        <div v-if="loading" class="space-y-3">
          <div v-for="i in 4" :key="i" class="flex items-center gap-3 p-3 rounded-xl bg-base-200/40">
            <div class="skeleton w-10 h-10 rounded-lg shrink-0"></div>
            <div class="flex-1 space-y-2">
              <div class="skeleton h-4 w-3/5 rounded"></div>
              <div class="flex gap-2">
                <div class="skeleton h-3 w-20 rounded"></div>
                <div class="skeleton h-3 w-24 rounded"></div>
              </div>
              <div class="skeleton h-3 w-16 rounded"></div>
            </div>
            <div class="skeleton h-8 w-8 rounded-lg shrink-0"></div>
          </div>
        </div>

        <!-- Empty state -->
        <div v-else-if="sessions.length === 0" class="text-center py-16">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-base-200 flex items-center justify-center">
            <Smartphone :size="32" class="text-base-content/20" />
          </div>
          <p class="text-sm font-medium text-base-content/30">{{ t('user.no_sessions') }}</p>
        </div>

        <!-- Session list -->
        <div v-else class="space-y-2">
          <div
            v-for="session in sessions"
            :key="session.id"
            class="flex items-center gap-3 p-3 rounded-xl bg-base-200/40 hover:bg-base-200 transition-colors group"
            :class="{ 'ring-1 ring-primary/20 bg-primary/5': session.is_current }"
          >
            <!-- Device type icon -->
            <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-base-300/40 flex items-center justify-center">
              <component :is="deviceIcon(session.device_type)" :size="20" class="text-base-content/40" />
            </div>

            <!-- Session details -->
            <div class="flex-1 min-w-0">
              <!-- Top row: browser · OS + provider + current badge -->
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-sm font-medium text-base-content/80">
                  {{ browserDisplay(session) }}
                  <span class="text-base-content/30 mx-1">·</span>
                  {{ osDisplay(session) }}
                </span>
                <span
                  v-if="session.login_provider"
                  class="inline-flex items-center gap-1 text-xs text-base-content/40 bg-base-200 px-2 py-0.5 rounded-md"
                >
                  <component :is="loginProviderIcon(session.login_provider)" :size="10" />
                  {{ loginProviderDisplay(session.login_provider) }}
                </span>
                <span
                  v-else
                  class="inline-flex items-center gap-1 text-xs text-base-content/40 bg-base-200 px-2 py-0.5 rounded-md"
                >
                  <Key :size="10" />
                  {{ t('user.login_method_password') }}
                </span>
                <span
                  v-if="session.is_current"
                  class="inline-flex items-center text-xs font-medium text-primary bg-primary/10 px-2 py-0.5 rounded-md"
                >
                  {{ t('user.current_session') }}
                </span>
              </div>

              <!-- Second row: IP, device model -->
              <div class="flex items-center gap-3 mt-1 text-xs text-base-content/40 flex-wrap">
                <span class="inline-flex items-center gap-1">
                  <Globe :size="11" />
                  <code class="font-mono bg-base-200 px-1.5 py-0.5 rounded text-xs">{{ session.ip_address }}</code>
                </span>
                <span class="inline-flex items-center gap-1">
                  <component :is="deviceIcon(session.device_type)" :size="11" />
                  <span>{{ deviceModelDisplay(session) }}</span>
                </span>
                <span class="inline-flex items-center gap-1">
                  <component :is="browserIcon(session.browser)" :size="11" />
                  <span>{{ browserDisplay(session) }}</span>
                </span>
                <span class="inline-flex items-center gap-1">
                  <component :is="osIcon(session.os)" :size="11" />
                  <span>{{ osDisplay(session) }}</span>
                </span>
              </div>

              <!-- Third row: timestamps -->
              <div class="flex items-center gap-2 mt-1 text-xs text-base-content/25">
                <span>{{ formatRelativeTime(session.last_activity) }}</span>
                <span>·</span>
                <span>{{ t('user.session_expires') }}: {{ formatRelativeTime(session.expires_at) }}</span>
              </div>
            </div>

            <!-- Revoke button -->
            <button
              v-if="!session.is_current"
              class="opacity-0 group-hover:opacity-100 transition-opacity text-red-400 hover:text-red-500 hover:bg-red-500/10 p-2 rounded-lg cursor-pointer disabled:opacity-40"
              :disabled="revokingId === session.id"
              @click="revokeSession(session.id)"
            >
              <span v-if="revokingId === session.id" class="loading loading-spinner loading-xs"></span>
              <Trash2 v-else :size="15" />
            </button>
          </div>
        </div>
      </div>
    </div>
    </div>

    <!-- ===== COL 3: Sidebar ===== -->
    <div class="space-y-4">
      <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
        <h3 class="text-sm font-bold text-base-content flex items-center gap-2 mb-3">
          <ShieldCheck :size="16" class="text-primary/60" />
          {{ t('user.session_info') }}
        </h3>
        <ul class="space-y-2 text-xs text-base-content/50">
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.session_info_current') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.session_info_revoke') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.session_info_revoke_all') }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
