<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from '../i18n'
import { useSEO } from '../composables/useSEO'
import { fetchProfile, clearTokens } from '../composables/useUserAuth'
import AvatarInitials from '../components/AvatarInitials.vue'
import { LayoutDashboard, User, Lock, Key, Smartphone, Link2, LogOut } from 'lucide-vue-next'

const { t } = useI18n()
useSEO({ title: t('user.dashboard'), description: t('user.dashboard') })
const route = useRoute()

const profile = ref(null)
const loading = ref(true)
const error = ref('')

const sidebarItems = [
  { path: '/dashboard',           icon: LayoutDashboard, label: 'user.overview', exact: true },
  { path: '/dashboard/profile',   icon: User,             label: 'user.profile' },
  { path: '/dashboard/security',  icon: Lock,             label: 'user.security' },
  { path: '/dashboard/api-keys',  icon: Key,              label: 'user.api_keys' },
  { path: '/dashboard/sessions',  icon: Smartphone,        label: 'user.sessions' },
  { path: '/dashboard/integrations', icon: Link2,         label: 'user.integrations' },
]

function isActive(item) {
  if (item.exact) return route.path === item.path
  return route.path.startsWith(item.path)
}

// Current page name from sidebar items
const currentPageLabel = computed(() => {
  const item = sidebarItems.find(item => isActive(item))
  return item ? t(item.label) : t('user.dashboard')
})

function handleLogout() {
  clearTokens()
  window.location.href = '/'
}

onMounted(async () => {
  try {
    const data = await fetchProfile()
    if (data) {
      profile.value = data
    } else {
      error.value = t('user.network_error')
    }
  } catch {
    error.value = t('user.network_error')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <!-- Loading skeleton -->
  <div v-if="loading" class="flex h-[calc(100vh-4rem)]">
    <!-- Sidebar skeleton -->
    <aside class="hidden md:flex flex-col w-64 shrink-0 bg-base-100 border-r border-base-300/20 p-5">
      <div class="skeleton h-5 w-24 rounded mb-6"></div>
      <div class="skeleton h-20 rounded-2xl mb-4"></div>
      <div class="space-y-2">
        <div class="skeleton h-10 rounded-xl"></div>
        <div class="skeleton h-10 rounded-xl"></div>
        <div class="skeleton h-10 rounded-xl"></div>
        <div class="skeleton h-10 rounded-xl"></div>
        <div class="skeleton h-10 rounded-xl"></div>
        <div class="skeleton h-10 rounded-xl"></div>
      </div>
    </aside>
    <!-- Content skeleton -->
    <div class="flex-1 overflow-auto bg-base-200 p-6 md:p-8 space-y-5">
      <div class="skeleton h-36 rounded-2xl"></div>
      <div class="skeleton h-28 rounded-2xl"></div>
      <div class="skeleton h-20 rounded-2xl"></div>
    </div>
  </div>

  <!-- Error -->
  <div v-else-if="error" class="flex h-[calc(100vh-4rem)] bg-base-200 items-center justify-center">
    <div class="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 max-w-md flex items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>{{ error }}</span>
    </div>
  </div>

  <!-- Dashboard -->
  <div v-else class="flex h-[calc(100vh-4rem)]">
    <!-- ===== Left sidebar ===== -->
    <aside class="hidden md:flex flex-col w-64 shrink-0 bg-base-100 border-r border-base-300/20">
      <!-- Sidebar header: user info -->
      <div class="px-5 pt-6 pb-4">
        <div class="bg-base-200/60 rounded-2xl p-4 flex items-center gap-3">
          <AvatarInitials :name="profile?.name" size="md" />
          <div class="min-w-0">
            <p class="text-sm font-semibold text-base-content truncate">{{ profile?.name || '—' }}</p>
            <p class="text-xs text-base-content/40 truncate">{{ profile?.email || '—' }}</p>
          </div>
        </div>
      </div>

      <!-- Nav items -->
      <nav class="flex-1 px-3 space-y-0.5">
        <router-link
          v-for="item in sidebarItems"
          :key="item.path"
          :to="item.path"
          :class="[
            'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200',
            isActive(item)
              ? 'bg-primary/10 text-primary'
              : 'text-base-content/50 hover:text-base-content hover:bg-base-200/50'
          ]"
        >
          <component :is="item.icon" :size="19" :stroke-width="isActive(item) ? 2.25 : 1.75" />
          <span>{{ t(item.label) }}</span>
        </router-link>
      </nav>

      <!-- Footer: logout -->
      <div class="p-3 border-t border-base-200">
        <button
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium text-base-content/40 hover:text-error hover:bg-error/5 transition-all duration-200 w-full cursor-pointer"
          @click="handleLogout"
        >
          <LogOut :size="19" />
          <span>Logout</span>
        </button>
      </div>
    </aside>

    <!-- ===== Main content ===== -->
    <main class="flex-1 overflow-auto bg-base-200">
      <div class="p-6 md:p-8">
        <h1 class="text-lg font-bold text-base-content/60 mb-5">{{ currentPageLabel }}</h1>
        <router-view />
      </div>
    </main>
  </div>
</template>
