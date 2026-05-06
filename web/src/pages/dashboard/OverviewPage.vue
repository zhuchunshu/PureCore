<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { fetchProfile, accessToken } from '../../composables/useUserAuth'
import UserOverview from '../../components/dashboard/UserOverview.vue'

const { t } = useI18n()
useSEO({ title: t('user.overview'), description: t('user.dashboard') })

const profile = ref(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  if (!accessToken.value) {
    loading.value = false
    return
  }
  try {
    const data = await fetchProfile()
    if (data) {
      profile.value = data
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
  <div v-if="loading" class="space-y-6">
    <div class="skeleton h-8 w-40 rounded-lg mb-2"></div>
    <div class="skeleton h-24 rounded-2xl"></div>
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="skeleton h-32 rounded-2xl"></div>
      <div class="skeleton h-32 rounded-2xl"></div>
      <div class="skeleton h-32 rounded-2xl"></div>
      <div class="skeleton h-32 rounded-2xl"></div>
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div class="skeleton h-36 rounded-2xl"></div>
      <div class="skeleton h-36 rounded-2xl"></div>
      <div class="skeleton h-36 rounded-2xl"></div>
    </div>
    <div class="skeleton h-40 rounded-2xl"></div>
  </div>

  <div v-else-if="error" class="text-center py-20 text-red-400">{{ error }}</div>

  <UserOverview v-else :profile="profile" />
</template>
