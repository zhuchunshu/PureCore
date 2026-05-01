<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { fetchProfile, accessToken } from '../../composables/useUserAuth'
import UserProfile from '../../components/dashboard/UserProfile.vue'

const { t } = useI18n()
useSEO({ title: t('user.profile'), description: t('user.profile') })

const profile = ref(null)
const loading = ref(true)
const error = ref('')
const profileRef = ref(null)

const onProfileUpdated = (updated) => {
  profile.value = { ...profile.value, ...updated }
}

onMounted(async () => {
  if (!accessToken.value) return
  try {
    const data = await fetchProfile()
    if (data) {
      profile.value = data
      setTimeout(() => profileRef.value?.initProfileForm(), 0)
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
  <UserProfile
    v-else
    ref="profileRef"
    :profile="profile"
    @profile-updated="onProfileUpdated"
  />
</template>
