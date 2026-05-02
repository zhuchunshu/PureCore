<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { adminOption } from '../composables/useAdminOption'

const props = defineProps({
  context: {
    type: String,
    required: true,
  },
  theme: {
    type: String,
    default: 'auto',
  },
})

const emit = defineEmits(['verified', 'error', 'expired'])

const widgetId = ref(null)
const isEnabled = ref(false)
const loading = ref(true)
const verified = ref(false)
const widgetError = ref('')
const siteKeyPrefix = ref('')

let siteKey = ''

const elementId = `turnstile-${props.context}-${Math.random().toString(36).slice(2, 9)}`

function loadTurnstileScript() {
  return new Promise((resolve, reject) => {
    if (document.getElementById('turnstile-script')) {
      resolve()
      return
    }
    const script = document.createElement('script')
    script.id = 'turnstile-script'
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.defer = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Failed to load Turnstile script'))
    const timeout = setTimeout(() => reject(new Error('Turnstile script load timed out')), 10000)
    script.addEventListener('load', () => clearTimeout(timeout))
    document.head.appendChild(script)
  })
}

function initWidget() {
  if (!window.turnstile || !isEnabled.value) return
  const container = document.getElementById(elementId)
  if (!container) return
  try {
    const id = window.turnstile.render(container, {
      sitekey: siteKey,
      theme: props.theme,
      callback: (token) => {
        verified.value = true
        emit('verified', token)
      },
      'error-callback': () => {
        widgetError.value = `CAPTCHA verification failed`
        emit('error')
      },
      'expired-callback': () => {
        verified.value = false
        emit('expired')
      },
    })
    widgetId.value = id
  } catch (err) {
    console.error('Turnstile render error:', err)
    widgetError.value = 'Failed to initialize CAPTCHA'
  }
}

onMounted(async () => {
  try {
    const rawSiteKey = await adminOption('turnstile_site_key', '')
    const toggleValue = await adminOption(props.context, '')
    isEnabled.value = rawSiteKey !== '' && toggleValue === '1'
    loading.value = false
    if (!isEnabled.value) return

    if (rawSiteKey.startsWith('1x')) {
      siteKey = '1x00000000000000000000AA'
      siteKeyPrefix.value = '1x...test'
    } else {
      siteKey = rawSiteKey
      if (siteKey.length >= 6) siteKeyPrefix.value = siteKey.substring(0, 6) + '...'
    }
    if (siteKey.length < 10) {
      widgetError.value = `Invalid Turnstile site key configured`
      return
    }
    await loadTurnstileScript()
    await new Promise(resolve => setTimeout(resolve, 300))
    initWidget()
  } catch (err) {
    console.error('Turnstile initialization error:', err)
    widgetError.value = err.message || 'Failed to initialize CAPTCHA'
    loading.value = false
  }
})

onUnmounted(() => {
  if (widgetId.value !== null && window.turnstile) {
    try { window.turnstile.remove(widgetId.value) } catch (err) { /* ignore */ }
  }
})

function reset() {
  verified.value = false
  if (widgetId.value !== null && window.turnstile) {
    try { window.turnstile.reset(widgetId.value) } catch (err) { /* ignore */ }
  }
}

function getToken() {
  if (!isEnabled.value || !window.turnstile) return ''
  const container = document.getElementById(elementId)
  if (!container) return ''
  const input = container.querySelector('input[name="cf-turnstile-response"]')
  return input ? input.value : ''
}

defineExpose({ reset, getToken, verified, isEnabled })
</script>

<template>
  <div v-if="loading" class="flex justify-center py-4">
    <div class="skeleton h-16 w-72 rounded-xl"></div>
  </div>

  <div v-else-if="verified" class="py-4">
    <div class="flex items-center justify-center gap-2.5 py-3 px-5 rounded-2xl bg-success/10 border border-success/20 text-success text-sm font-medium backdrop-blur-sm shadow-inner shadow-success/5">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>Verification successful</span>
    </div>
  </div>

  <div v-else-if="widgetError" class="py-4">
    <div class="p-3.5 rounded-2xl bg-warning/10 border border-warning/20 text-warning text-sm text-center space-y-1.5 backdrop-blur-sm">
      <div class="flex items-center justify-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>{{ widgetError }}</span>
      </div>
      <p class="text-xs opacity-60">Add your domain in Cloudflare Dashboard → Turnstile to fix this</p>
    </div>
  </div>

  <div v-else-if="!loading && isEnabled" :id="elementId" class="turnstile-widget flex justify-center py-4"></div>
</template>
