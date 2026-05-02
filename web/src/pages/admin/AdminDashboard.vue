<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { clearTokens, accessToken } from '../../composables/useAuth'
import AdminOverview from '../../components/dashboard/AdminOverview.vue'

const { t } = useI18n()
useSEO({
  title: t('admin.dashboard'),
  description: t('admin.dashboard'),
})
const router = useRouter()
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

const profile = ref(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  if (!accessToken.value) {
    router.push(`/${adminPrefix}/login`)
    return
  }
  try {
    const resp = await fetch(`/api/v1/${adminPrefix}/auth/profile`, {
      headers: { Authorization: `Bearer ${accessToken.value}` },
    })
    const json = await resp.json()
    if (json.code === 0) {
      profile.value = json.data
    } else {
      clearTokens()
      router.push(`/${adminPrefix}/login`)
    }
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <!-- Loading skeleton -->
  <div v-if="loading" class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-6 pb-8">
    <!-- Welcome header skeleton -->
    <div class="mb-6">
      <div class="skeleton h-8 w-56 rounded-lg mb-2"></div>
      <div class="skeleton h-4 w-72 rounded"></div>
    </div>
    <!-- Stats grid skeleton -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <div class="skeleton h-24 rounded-2xl"></div>
      <div class="skeleton h-24 rounded-2xl"></div>
      <div class="skeleton h-24 rounded-2xl"></div>
      <div class="skeleton h-24 rounded-2xl"></div>
    </div>
    <!-- Quick actions skeleton -->
    <div class="skeleton h-32 rounded-2xl"></div>
  </div>

  <!-- Error state -->
  <div v-else-if="error" class="flex items-center justify-center py-20">
    <div class="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 max-w-md flex items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>{{ error }}</span>
    </div>
  </div>

  <!-- Dashboard content -->
  <div v-else class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-6 pb-8">
    <AdminOverview :profile="profile" />
  </div>
</template>
