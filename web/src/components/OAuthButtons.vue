<script setup>
import { onMounted, ref } from 'vue'
import { useOAuth } from '../composables/useOAuth'
import { useI18n } from '../i18n'
import { IconBrandGithub, IconBrandGoogle, IconBrandDiscord, IconBrandApple, IconBrandTelegram } from '@tabler/icons-vue'

const props = defineProps({
  mode: {
    type: String,
    default: 'login', // 'login' or 'register'
  },
})

const { t } = useI18n()
const { providers, fetchProviders, getProviderUrl } = useOAuth()
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  await fetchProviders()
  loading.value = false
})

function getProviderLabel(providerKey) {
  // Map provider key to i18n key
  const keyMap = {
    github: 'auth.oauth.github',
    google: 'auth.oauth.google',
    discord: 'auth.oauth.discord',
    apple: 'auth.oauth.apple',
    telegram: 'auth.oauth.telegram',
  }
  if (keyMap[providerKey]) return t(keyMap[providerKey])
  return providerKey
}

const providerIcons = {
  github: IconBrandGithub,
  google: IconBrandGoogle,
  discord: IconBrandDiscord,
  apple: IconBrandApple,
  telegram: IconBrandTelegram,
}

function getButtonClass(providerKey) {
  const classMap = {
    github: 'bg-neutral-800 hover:bg-neutral-700 text-white border-neutral-700',
    google: 'bg-white hover:bg-gray-100 text-gray-700 border-gray-300',
    discord: 'bg-indigo-600 hover:bg-indigo-500 text-white border-indigo-500',
    apple: 'bg-black hover:bg-gray-900 text-white border-gray-800',
    telegram: 'bg-sky-500 hover:bg-sky-400 text-white border-sky-400',
  }
  if (classMap[providerKey]) return classMap[providerKey]
  return 'bg-base-100 hover:bg-base-200 text-base-content border-base-content/20'
}

function loginWith(providerKey) {
  window.location.href = getProviderUrl(providerKey)
}
</script>

<template>
  <div v-if="!loading && providers.length > 0" class="space-y-3">
    <!-- Divider -->
    <div class="relative my-4">
      <div class="absolute inset-0 flex items-center">
        <div class="w-full border-t border-base-content/10"></div>
      </div>
      <div class="relative flex justify-center text-xs">
        <span class="px-3 bg-base-200/40 backdrop-blur-xl rounded-full text-base-content/40 font-medium">
          {{ mode === 'login' ? 'or sign in with' : 'or sign up with' }}
        </span>
      </div>
    </div>

    <!-- Buttons -->
    <div class="space-y-2.5">
      <button
        v-for="provider in providers"
        :key="provider.key"
        @click="loginWith(provider.key)"
        type="button"
        :class="[
          'w-full flex items-center justify-center gap-3 py-3 px-4 rounded-2xl border font-medium text-sm',
          'transition-all duration-200 hover:scale-[1.02] active:scale-[0.98] shadow-sm',
          getButtonClass(provider.key)
        ]"
      >
        <component :is="providerIcons[provider.key]" :size="20" />
        <span>{{ getProviderLabel(provider.key) }}</span>
      </button>
    </div>
  </div>
</template>
