<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'

const { t } = useI18n()
const router = useRouter()
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'
const profile = ref(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const token = localStorage.getItem('admin_token')
  if (!token) {
    router.push(`/${adminPrefix}/login`)
    return
  }
  try {
    const resp = await fetch(`/api/v1/${adminPrefix}/auth/profile`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    const json = await resp.json()
    if (json.code === 0) {
      profile.value = json.data
    } else {
      localStorage.removeItem('admin_token')
      router.push(`/${adminPrefix}/login`)
    }
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
})

function logout() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_user')
  router.push(`/${adminPrefix}/login`)
}

const stats = [
  { labelKey: 'admin.users', value: '—', icon: '👥', color: 'text-primary' },
  { labelKey: 'admin.active_sessions', value: '1', icon: '🔑', color: 'text-secondary' },
  { labelKey: 'admin.database', value: 'PostgreSQL', icon: '🗄️', color: 'text-accent' },
  { labelKey: 'admin.framework', value: 'PureCore', icon: '⚡', color: 'text-info' },
]
</script>

<template>
  <!-- Full-page loading spinner until auth check completes -->
  <div v-if="loading" class="min-h-screen flex items-center justify-center bg-base-200">
    <span class="loading loading-spinner loading-lg text-primary"></span>
  </div>

  <!-- Error state -->
  <div v-else-if="error" class="min-h-screen flex items-center justify-center bg-base-200">
    <div class="alert alert-error max-w-md">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>{{ error }}</span>
    </div>
  </div>

  <!-- Dashboard content (only shown after auth is verified) -->
  <div v-else class="min-h-screen bg-base-200">
    <!-- Header -->
    <div class="navbar bg-base-100 shadow-sm border-b border-base-300/20">
      <div class="navbar-start">
        <span class="text-xl font-black">Pure<span class="text-primary">Core</span> Admin</span>
      </div>
      <div class="navbar-end gap-2">
        <div v-if="profile" class="dropdown dropdown-end">
          <label tabindex="0" class="btn btn-ghost btn-sm">
            <span class="text-sm">{{ profile.name }}</span>
            <span class="badge badge-sm badge-ghost">{{ profile.role }}</span>
          </label>
          <ul tabindex="0" class="menu menu-sm dropdown-content mt-2 z-50 p-2 shadow bg-base-100 rounded-box w-40">
            <li><button @click="logout">{{ t('admin.logout') }}</button></li>
          </ul>
        </div>
      </div>
    </div>

    <div class="flex">
      <!-- Sidebar -->
      <aside class="w-64 bg-base-100 min-h-[calc(100vh-64px)] border-r border-base-300/20 hidden lg:block">
        <ul class="menu p-4 gap-1">
          <li><a class="active"><span>📊</span> {{ t('admin.dashboard') }}</a></li>
          <li><a><span>👥</span> {{ t('admin.users') }}</a></li>
          <li><a><span>⚙️</span> {{ t('admin.settings') }}</a></li>
          <li class="mt-auto pt-4 border-t border-base-300/20">
            <a href="/" target="_blank"><span>🏠</span> {{ t('admin.view_site') }}</a>
          </li>
          <li><button @click="logout"><span>🚪</span> {{ t('admin.logout') }}</button></li>
        </ul>
      </aside>

      <!-- Main Content -->
      <main class="flex-1 p-6">
        <div class="mb-6">
          <h1 class="text-2xl font-bold">{{ t('admin.dashboard') }}</h1>
          <p v-if="profile" class="text-base-content/50 mt-1">
            {{ t('admin.welcome') }}, {{ profile.name }}
          </p>
        </div>

        <!-- Stats -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <div v-for="s in stats" :key="s.labelKey" class="card bg-base-100 shadow-sm border border-base-300/20">
            <div class="card-body p-4">
              <div class="flex items-center justify-between">
                <span class="text-2xl">{{ s.icon }}</span>
                <span :class="['text-lg font-bold', s.color]">{{ s.value }}</span>
              </div>
              <p class="text-sm text-base-content/50 mt-2">{{ t(s.labelKey) }}</p>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="card bg-base-100 shadow-sm border border-base-300/20">
          <div class="card-body">
            <h2 class="card-title text-lg mb-4">{{ t('admin.quick_actions') }}</h2>
            <div class="flex flex-wrap gap-2">
              <button class="btn btn-outline btn-sm">{{ t('admin.add_user') }}</button>
              <button class="btn btn-outline btn-sm">{{ t('admin.view_logs') }}</button>
              <button class="btn btn-outline btn-sm">{{ t('admin.backup_db') }}</button>
              <button class="btn btn-outline btn-sm">{{ t('admin.clear_cache') }}</button>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>
