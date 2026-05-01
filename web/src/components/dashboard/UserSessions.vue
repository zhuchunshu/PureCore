<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../../i18n'
import { userAPI } from '../../services/api'
import { useToast } from '../../composables/useToast'
import TechCard from '../TechCard.vue'
import {
  Smartphone, Tablet, Monitor,
  Globe, Compass,
  Apple, Terminal,
  Trash2, ShieldX
} from 'lucide-vue-next'

const { t } = useI18n()
const { toastSuccess, toastError } = useToast()

const sessions = ref([])
const loading = ref(true)
const revokingId = ref(null)
const revokingAll = ref(false)

// Icons for device type
const deviceIcon = (type) => {
  switch (type) {
    case 'mobile': return Smartphone
    case 'tablet': return Tablet
    default: return Monitor
  }
}

// Icons for browser — most browsers map to Globe, Safari to Compass
const browserIcon = (browser) => {
  switch (browser?.toLowerCase()) {
    case 'safari': return Compass
    default: return Globe
  }
}

// Icons for OS
const osIcon = (os) => {
  switch (os?.toLowerCase()) {
    case 'macos':
    case 'ios': return Apple
    case 'android': return Smartphone
    case 'linux': return Terminal
    default: return Monitor
  }
}

// Build device model display string
const deviceModelDisplay = (session) => {
  const parts = []
  if (session.device_brand) parts.push(session.device_brand)
  if (session.device_model) parts.push(session.device_model)
  return parts.length > 0 ? parts.join(' ') : t('user.session_unknown_device')
}

// Browser name or fallback
const browserDisplay = (session) => {
  return session.browser || t('user.session_unknown_browser')
}

// OS name or fallback
const osDisplay = (session) => {
  return session.os || t('user.session_unknown_os')
}

// Simple template replacement for {n} (i18n does not support parameter interpolation)
const tp = (key, replacements) => {
  let text = t(key)
  for (const [k, v] of Object.entries(replacements)) {
    text = text.replace(`{${k}}`, v)
  }
  return text
}

// Format relative time (past: "3m ago", future: "6天后")
const formatRelativeTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = date - now                     // positive = future, negative = past
  const absDiff = Math.abs(diff)

  if (absDiff < 60000) return t('user.just_now')

  if (diff > 0) {
    // Future
    const hours = Math.ceil(absDiff / 3600000)
    if (hours < 48) return tp('user.in_hours', { n: hours })
    const days = Math.ceil(absDiff / 86400000)
    return tp('user.in_days', { n: days })
  }

  // Past
  if (absDiff < 3600000) return Math.floor(absDiff / 60000) + 'm ago'
  if (absDiff < 86400000) return Math.floor(absDiff / 3600000) + 'h ago'
  return date.toLocaleDateString()
}

// Fetch sessions from API
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

// Revoke a single session
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

// Revoke all sessions except current
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
  <TechCard variant="blue" padded>
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-5">
      <div>
        <h2 class="text-lg font-bold text-base-content/80 flex items-center gap-2">
          <Smartphone :size="20" />
          {{ t('user.sessions') }}
        </h2>
        <p class="text-sm text-base-content/50 mt-0.5">{{ t('user.sessions_description') }}</p>
      </div>
      <button
        class="btn btn-ghost btn-sm rounded-xl text-red-400 hover:bg-red-500/10"
        :disabled="revokingAll || sessions.filter(s => !s.is_current).length === 0"
        @click="revokeAll"
      >
        <span v-if="revokingAll" class="loading loading-spinner loading-xs"></span>
        <ShieldX v-else :size="14" class="mr-1" />
        {{ t('user.revoke_all') }}
      </button>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="text-center py-10 text-base-content/40">
      <span class="loading loading-spinner loading-md mb-2"></span>
      <span class="text-sm block">{{ t('user.session_loading') }}</span>
    </div>

    <!-- Empty state -->
    <div v-else-if="sessions.length === 0" class="text-center py-10 text-base-content/40">
      <Smartphone :size="48" class="mx-auto mb-2 opacity-30" />
      <span class="text-sm">{{ t('user.no_sessions') }}</span>
    </div>

    <!-- Session list -->
    <div v-else class="space-y-3">
      <div
        v-for="session in sessions"
        :key="session.id"
        class="flex items-center gap-3 p-3 rounded-xl bg-base-200/50 hover:bg-base-200 transition-colors group"
        :class="{ 'ring-1 ring-primary/20': session.is_current }"
      >
        <!-- Device type icon -->
        <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-base-300/50 flex items-center justify-center">
          <component :is="deviceIcon(session.device_type)" :size="20" class="text-base-content/50" />
        </div>

        <!-- Session details -->
        <div class="flex-1 min-w-0">
          <!-- Top row: browser · OS + current badge -->
          <div class="flex items-center gap-2 flex-wrap">
            <span class="text-sm font-medium text-base-content/80 truncate">
              {{ browserDisplay(session) }}
              <span class="text-base-content/40 mx-1">·</span>
              {{ osDisplay(session) }}
            </span>
            <span
              v-if="session.is_current"
              class="badge badge-xs badge-primary !rounded-md"
            >
              {{ t('user.current_session') }}
            </span>
          </div>

          <!-- Second row: IP, device model, browser, OS icons -->
          <div class="flex items-center gap-3 mt-1 text-xs text-base-content/40 flex-wrap">
            <span class="inline-flex items-center gap-1">
              <Globe :size="12" />
              <code class="font-mono bg-base-300/50 px-1.5 py-0.5 rounded">{{ session.ip_address }}</code>
            </span>
            <span class="inline-flex items-center gap-1">
              <component :is="deviceIcon(session.device_type)" :size="12" />
              <span>{{ deviceModelDisplay(session) }}</span>
            </span>
            <span class="inline-flex items-center gap-1">
              <component :is="browserIcon(session.browser)" :size="12" />
              <span>{{ browserDisplay(session) }}</span>
            </span>
            <span class="inline-flex items-center gap-1">
              <component :is="osIcon(session.os)" :size="12" />
              <span>{{ osDisplay(session) }}</span>
            </span>
          </div>

          <!-- Third row: timestamps -->
          <div class="flex items-center gap-2 mt-1 text-xs text-base-content/30">
            <span>{{ formatRelativeTime(session.last_activity) }}</span>
            <span>·</span>
            <span>{{ t('user.session_expires') }}: {{ formatRelativeTime(session.expires_at) }}</span>
          </div>
        </div>

        <!-- Revoke button (hidden for current session) -->
        <button
          v-if="!session.is_current"
          class="btn btn-ghost btn-xs rounded-lg text-red-400 hover:bg-red-500/10 opacity-0 group-hover:opacity-100 transition-opacity"
          :disabled="revokingId === session.id"
          @click="revokeSession(session.id)"
        >
          <span v-if="revokingId === session.id" class="loading loading-spinner loading-xs"></span>
          <Trash2 v-else :size="14" />
          <span class="ml-1">{{ t('user.revoke') }}</span>
        </button>
      </div>
    </div>
  </TechCard>
</template>
