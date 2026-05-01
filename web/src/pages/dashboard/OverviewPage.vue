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
  if (!accessToken.value) return
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
  <div v-if="loading" class="text-center py-20">
    <span class="loading loading-spinner loading-lg text-primary"></span>
  </div>
  <div v-else-if="error" class="text-center py-20 text-red-400">{{ error }}</div>
  <UserOverview v-else :profile="profile" />
</template>
