<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from '../i18n'
import { useSEO } from '../composables/useSEO'
import { accessToken, fetchProfile, clearTokens } from '../composables/useUserAuth'
import { useOAuth } from '../composables/useOAuth'

const { t } = useI18n()
useSEO({ title: t('user.dashboard'), description: t('user.dashboard') })
const router = useRouter()
const route = useRoute()

const profile = ref(null)
const loading = ref(true)
const error = ref('')

const sidebarItems = [
  { path: '/dashboard',          icon: '📊', label: 'user.overview', exact: true },
  { path: '/dashboard/profile',  icon: '👤', label: 'user.profile' },
  { path: '/dashboard/security', icon: '🔒', label: 'user.security' },
  { path: '/dashboard/api-keys', icon: '🔑', label: 'user.api_keys' },
  { path: '/dashboard/sessions', icon: '📱', label: 'user.sessions' },
]

function isActive(item) {
  if (item.exact) return route.path === item.path
  return route.path.startsWith(item.path)
}

onMounted(async () => {
  // Handle OAuth callback (tokens passed via URL query params)
  const { handleOAuthCallback } = useOAuth()
  if (handleOAuthCallback()) {
    // OAuth login successful — refresh the page to clean up and load fresh state
    // The handleOAuthCallback already sets tokens and cleans the URL
    // We reload to ensure proper state initialization
    window.location.href = '/dashboard'
    return
  }

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
  } catch {
    error.value = t('user.network_error')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <!-- Loading -->
  <div v-if="loading" class="flex items-center justify-center py-20">
    <span class="loading loading-spinner loading-lg text-primary"></span>
  </div>

  <!-- Error -->
  <div v-else-if="error" class="flex items-center justify-center py-20">
    <div class="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 max-w-md flex items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>{{ error }}</span>
    </div>
  </div>

  <!-- Dashboard -->
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
        <router-link
          v-for="item in sidebarItems"
          :key="item.path"
          :to="item.path"
          :class="[
            'flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-colors',
            isActive(item)
              ? 'bg-primary/10 text-primary border border-primary/20'
              : 'hover:bg-base-200/50 text-base-content/70 hover:text-base-content'
          ]"
        >
          <span class="text-lg">{{ item.icon }}</span>
          <span>{{ t(item.label) }}</span>
        </router-link>
      </aside>

      <!-- Main content area -->
      <div class="flex-1 min-w-0">
        <router-view />
      </div>
    </div>
  </div>
</template>
