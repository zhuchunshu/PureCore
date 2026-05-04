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
  if (!accessToken.value) {
    loading.value = false
    return
  }
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
  <!-- Loading skeleton -->
  <div v-if="loading" class="space-y-6">
    <!-- Avatar + name skeleton -->
    <div class="flex items-center gap-4 mb-4">
      <div class="skeleton w-16 h-16 rounded-full"></div>
      <div class="space-y-2">
        <div class="skeleton h-5 w-32 rounded"></div>
        <div class="skeleton h-4 w-48 rounded"></div>
      </div>
    </div>
    <!-- Form fields skeleton -->
    <div class="space-y-4">
      <div class="skeleton h-12 rounded-xl"></div>
      <div class="skeleton h-12 rounded-xl"></div>
      <div class="skeleton h-24 rounded-xl"></div>
      <div class="skeleton h-12 rounded-xl"></div>
    </div>
    <!-- Save button skeleton -->
    <div class="skeleton h-10 w-28 rounded-xl mt-6"></div>
  </div>
  <div v-else-if="error" class="text-center py-20 text-red-400">{{ error }}</div>
  <UserProfile
    v-else
    ref="profileRef"
    :profile="profile"
    @profile-updated="onProfileUpdated"
  />
</template>
