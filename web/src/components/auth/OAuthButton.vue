<script setup>
import { ref, computed } from 'vue'
import { useOAuth } from '../../composables/useOAuth'
import {
  IconBrandGithub,
  IconBrandGoogle,
  IconBrandApple,
  IconBrandTelegram,
  IconBrandDiscord,
  IconWorld,
} from '@tabler/icons-vue'

const props = defineProps({
  provider: { type: Object, required: true },
  redirect: { type: String, default: '/' },
})

const { initiateLogin } = useOAuth()

const providerName = computed(() => props.provider?.display_name || props.provider?.name || 'OAuth')
const providerId = computed(() => props.provider?.name || '')
const isLoading = ref(false)

// Map provider names to Tabler brand icons
const providerIcon = computed(() => {
  const map = {
    github: IconBrandGithub,
    google: IconBrandGoogle,
    apple: IconBrandApple,
    telegram: IconBrandTelegram,
    discord: IconBrandDiscord,
  }
  return map[providerId.value] || IconWorld
})

async function handleClick() {
  if (isLoading.value || !providerId.value) return
  isLoading.value = true
  try {
    // initiateLogin will trigger a full page redirect to the provider
    await initiateLogin(providerId.value, props.redirect)
  } catch (err) {
    console.error('OAuth login error:', err)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <button
    type="button"
    :title="providerName"
    :aria-label="providerName"
    class="btn btn-circle btn-outline"
    :class="isLoading ? 'btn-disabled' : ''"
    @click="handleClick"
  >
    <span v-if="isLoading" class="loading loading-spinner loading-xs"></span>
    <component
      v-else
      :is="providerIcon"
      :size="20"
      aria-hidden="true"
    />
  </button>
</template>
