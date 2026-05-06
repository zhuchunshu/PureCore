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

function ensureTelegramScriptLoaded() {
  return new Promise((resolve, reject) => {
    if (typeof window === 'undefined') {
      reject(new Error('Telegram login unavailable in SSR'))
      return
    }
    if (window.Telegram?.Login?.auth) {
      resolve()
      return
    }
    const existing = document.getElementById('telegram-login-sdk')
    if (existing) {
      existing.addEventListener('load', () => resolve(), { once: true })
      existing.addEventListener('error', () => reject(new Error('Failed to load Telegram SDK')), { once: true })
      return
    }
    const script = document.createElement('script')
    script.id = 'telegram-login-sdk'
    script.src = 'https://telegram.org/js/telegram-widget.js?22'
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Failed to load Telegram SDK'))
    document.head.appendChild(script)
  })
}

async function triggerTelegramAuth(widgetData) {
  const botId = String(widgetData?.bot_id || '').trim()
  const callbackBase = String(widgetData?.redirect_url || '').trim()
  const state = String(widgetData?.state || '').trim()
  if (!botId || !callbackBase || !state) {
    throw new Error('Telegram widget config incomplete')
  }

  await ensureTelegramScriptLoaded()
  window.Telegram.Login.auth({ bot_id: botId, request_access: true }, (data) => {
    if (!data) return
    const callbackURL = new URL(callbackBase, window.location.origin)
    callbackURL.searchParams.set('state', state)
    Object.entries(data).forEach(([k, v]) => {
      if (v !== undefined && v !== null) callbackURL.searchParams.set(k, String(v))
    })
    window.location.href = callbackURL.toString()
  })
}

async function handleClick() {
  if (isLoading.value || !providerId.value) return
  isLoading.value = true
  try {
    // Special handling for Telegram (non-OAuth2 widget-based login)
    if (providerId.value === 'telegram') {
      const widgetData = await initiateLogin(providerId.value, props.redirect)
      if (widgetData && widgetData.type === 'widget') {
        await triggerTelegramAuth(widgetData)
        isLoading.value = false
        return
      }
    }
    // For OAuth2 providers, initiateLogin will redirect to the provider
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
