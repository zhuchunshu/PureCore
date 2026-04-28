<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../i18n'
import { useSEO } from '../composables/useSEO'
import { accessToken, fetchProfile, clearTokens } from '../composables/useUserAuth'
import TechCard from '../components/TechCard.vue'

const { t } = useI18n()
useSEO({
  title: t('user.dashboard'),
  description: t('user.dashboard'),
})
const router = useRouter()

const profile = ref(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  if (!accessToken.value) {
    router.push('/login')
    return
  }
  try {
    const data = await fetchProfile()
    if (data) {
      profile.value = data
    } else {
      clearTokens()
      router.push('/login')
    }
  } catch (err) {
    error.value = t('user.network_error')
  } finally {
    loading.value = false
  }
})

const statCards = [
  { label: 'user.posts', value: '0', icon: '📝', gradient: 'from-blue-500/20 to-blue-600/20', iconBg: 'from-blue-500 to-blue-600' },
  { label: 'user.comments', value: '0', icon: '💬', gradient: 'from-emerald-500/20 to-emerald-600/20', iconBg: 'from-emerald-500 to-emerald-600' },
  { label: 'user.likes', value: '0', icon: '❤️', gradient: 'from-purple-500/20 to-purple-600/20', iconBg: 'from-purple-500 to-purple-600' },
  { label: 'user.followers', value: '0', icon: '👥', gradient: 'from-amber-500/20 to-amber-600/20', iconBg: 'from-amber-500 to-amber-600' },
]

const actions = [
  { label: 'user.edit_profile', icon: '✏️', variant: 'blue' },
  { label: 'user.my_posts', icon: '📄', variant: 'emerald' },
  { label: 'user.settings', icon: '⚙️', variant: 'purple' },
  { label: 'user.help', icon: '❓', variant: 'amber' },
]
</script>

<template>
  <!-- Full-page loading spinner -->
  <div v-if="loading" class="flex items-center justify-center py-20">
    <span class="loading loading-spinner loading-lg text-primary"></span>
  </div>

  <!-- Error state -->
  <div v-else-if="error" class="flex items-center justify-center py-20">
    <div class="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 max-w-md flex items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>{{ error }}</span>
    </div>
  </div>

  <!-- Dashboard content -->
  <div v-else class="space-y-6">
    <!-- Welcome header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl md:text-3xl font-black tracking-tight">
          <span class="bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent">{{ t('user.dashboard') }}</span>
        </h1>
        <p v-if="profile" class="text-base-content/50 mt-1">
          👋 {{ t('user.welcome_back') }}<span v-if="profile.name" class="font-semibold text-base-content/80">, {{ profile.name }}</span>
        </p>
      </div>
    </div>

    <!-- User profile card -->
    <TechCard variant="blue" :hover="true" padded>
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-4">
        <div class="w-16 h-16 rounded-full bg-gradient-to-br from-blue-400 to-emerald-400 flex items-center justify-center text-2xl font-bold text-white shrink-0">
          {{ profile?.name?.charAt(0)?.toUpperCase() || '?' }}
        </div>
        <div class="flex-1 min-w-0">
          <h2 class="text-lg font-bold text-base-content/80">{{ profile?.name || '—' }}</h2>
          <p class="text-sm text-base-content/50">{{ profile?.email || '—' }}</p>
          <div v-if="profile?.bio" class="mt-2 text-sm text-base-content/60">{{ profile.bio }}</div>
        </div>
        <div class="flex items-center gap-2">
          <span v-if="profile?.status === 'active'" class="badge badge-success badge-sm">● {{ t('user.active') }}</span>
          <span v-else-if="profile?.status === 'banned'" class="badge badge-error badge-sm">● {{ t('user.banned') }}</span>
          <span v-else class="badge badge-ghost badge-sm">● {{ t('user.inactive') }}</span>
        </div>
      </div>
      <div v-if="profile?.last_login_at" class="mt-3 text-xs text-base-content/40">
        {{ t('user.last_login') }}: {{ new Date(profile.last_login_at).toLocaleString() }}
      </div>
    </TechCard>

    <!-- Stats grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <TechCard v-for="s in statCards" :key="s.label" variant="blue" :hover="true" :padded="false">
        <div :class="['p-5 rounded-2xl bg-gradient-to-br', s.gradient]">
          <div class="flex items-center justify-between">
            <span class="text-3xl">{{ s.icon }}</span>
            <span :class="['w-10 h-10 rounded-xl bg-gradient-to-br flex items-center justify-center text-white text-xs font-bold shadow-lg', s.iconBg]">
              {{ s.value }}
            </span>
          </div>
          <p class="text-sm font-medium text-base-content/60 mt-3">{{ t(s.label) }}</p>
        </div>
      </TechCard>
    </div>

    <!-- Quick actions -->
    <TechCard variant="emerald" padded>
      <h2 class="text-lg font-bold text-base-content/80 mb-4">⚡ {{ t('user.quick_actions') }}</h2>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="action in actions"
          :key="action.label"
          class="btn btn-sm gap-2 bg-base-200/80 border border-base-300/30 hover:bg-base-300/50 hover:border-base-300/50 transition-colors rounded-xl"
        >
          <span>{{ action.icon }}</span>
          <span>{{ t(action.label) }}</span>
        </button>
      </div>
    </TechCard>
  </div>
</template>
