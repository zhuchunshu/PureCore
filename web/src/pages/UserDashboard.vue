<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../i18n'
import { useSEO } from '../composables/useSEO'
import { accessToken, fetchProfile, clearTokens } from '../composables/useUserAuth'
import UserOverview from '../components/dashboard/UserOverview.vue'
import UserProfile from '../components/dashboard/UserProfile.vue'
import UserSecurity from '../components/dashboard/UserSecurity.vue'
import UserApiKeys from '../components/dashboard/UserApiKeys.vue'
import UserSessions from '../components/dashboard/UserSessions.vue'

const { t } = useI18n()
useSEO({
  title: t('user.dashboard'),
  description: t('user.dashboard'),
})
const router = useRouter()

const profile = ref(null)
const loading = ref(true)
const error = ref('')

const sidebarItems = [
  { key: 'overview',   icon: '📊', label: 'user.overview' },
  { key: 'account',    icon: '👤', label: 'user.profile',   children: ['profile', 'security'] },
  { key: 'developer',  icon: '🔑', label: 'user.api_keys',  children: ['api_keys', 'sessions'] },
]
const activeSidebar = ref('overview')

const tabs = [
  { key: 'profile',    icon: '👤', label: 'user.profile' },
  { key: 'security',   icon: '🔒', label: 'user.security' },
  { key: 'api_keys',   icon: '🔑', label: 'user.api_keys' },
  { key: 'sessions',   icon: '📱', label: 'user.sessions' },
]
const activeTab = ref('overview')

const subTabs = computed(() => {
  const item = sidebarItems.find(s => s.key === activeSidebar.value)
  if (!item || !item.children) return []
  return tabs.filter(t => item.children.includes(t.key))
})

const profileRef = ref(null)

function selectSidebar(key) {
  activeSidebar.value = key
  const item = sidebarItems.find(s => s.key === key)
  if (item && item.children) {
    activeTab.value = item.children[0]
    if (item.children[0] === 'profile') {
      // Wait for next tick so the component is mounted
      setTimeout(() => profileRef.value?.initProfileForm(), 0)
    }
  } else {
    activeTab.value = 'overview'
  }
}

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

function onProfileUpdated(updated) {
  profile.value = { ...profile.value, ...updated }
}
</script>

<template>
  <!-- ── Loading ──────────────────────────────────────── -->
  <div v-if="loading" class="flex items-center justify-center py-20">
    <span class="loading loading-spinner loading-lg text-primary"></span>
  </div>

  <!-- ── Error ────────────────────────────────────────── -->
  <div v-else-if="error" class="flex items-center justify-center py-20">
    <div class="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 max-w-md flex items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>{{ error }}</span>
    </div>
  </div>

  <!-- ── Dashboard ────────────────────────────────────── -->
  <div v-else class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-6 pb-8">
    <!-- Header -->
    <div class="mb-6">
      <h1 class="text-2xl md:text-3xl font-black tracking-tight">
        <span class="bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent">{{ t('user.dashboard') }}</span>
      </h1>
      <p v-if="profile" class="text-base-content/50 mt-1 text-sm">
        {{ t('user.welcome_back') }}<span v-if="profile.name" class="font-semibold text-base-content/80">, {{ profile.name }}</span>
      </p>
    </div>

    <!-- Layout: Sidebar + Main -->
    <div class="flex gap-6">
      <!-- Left sidebar -->
      <aside class="hidden md:flex flex-col w-56 shrink-0 space-y-1">
        <button
          v-for="item in sidebarItems"
          :key="item.key"
          @click="selectSidebar(item.key)"
          :class="[
            'flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-colors',
            activeSidebar === item.key
              ? 'bg-primary/10 text-primary border border-primary/20'
              : 'hover:bg-base-200/50 text-base-content/70 hover:text-base-content'
          ]"
        >
          <span class="text-lg">{{ item.icon }}</span>
          <span>{{ t(item.label) }}</span>
        </button>
      </aside>

      <!-- Main content area -->
      <div class="flex-1 min-w-0">
        <!-- Sub-tab bar (only if category has children) -->
        <div v-if="subTabs.length > 0" class="tabs tabs-boxed bg-base-200/50 border border-base-300/30 p-1 rounded-xl mb-6 overflow-x-auto">
          <button
            v-for="tab in subTabs"
            :key="tab.key"
            @click="activeTab = tab.key; if (tab.key === 'profile') setTimeout(() => profileRef?.initProfileForm(), 0)"
            :class="[
              'tab gap-2 whitespace-nowrap transition-colors',
              activeTab === tab.key ? 'tab-active bg-base-100 shadow-sm' : 'hover:text-base-content/70'
            ]"
          >
            <span class="text-lg">{{ tab.icon }}</span>
            <span class="text-sm font-medium hidden sm:inline">{{ t(tab.label) }}</span>
          </button>
        </div>

        <UserOverview v-if="activeTab === 'overview'" :profile="profile" />
        <UserProfile v-if="activeTab === 'profile'" ref="profileRef" :profile="profile" @profile-updated="onProfileUpdated" />
        <UserSecurity v-if="activeTab === 'security'" />
        <UserApiKeys v-if="activeTab === 'api_keys'" />
        <UserSessions v-if="activeTab === 'sessions'" />
      </div>
    </div>
  </div>
</template>
