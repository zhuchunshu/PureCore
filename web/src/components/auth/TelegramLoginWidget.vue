<script setup>
import { ref, onMounted, nextTick } from 'vue'

const props = defineProps({
  botUsername: { type: String, required: true },
  redirectUrl: { type: String, required: true },
  state: { type: String, required: true },
  size: { type: String, default: 'large' },
  borderRadius: { type: Number, default: 12 },
  autoTrigger: { type: Boolean, default: false },
})

const scriptLoaded = ref(false)
const widgetContainer = ref(null)

// Inject the Telegram widget script
function loadTelegramScript() {
  if (!widgetContainer.value) {
    return
  }

  widgetContainer.value.innerHTML = ''
  const script = document.createElement('script')
  script.src = 'https://telegram.org/js/telegram-widget.js?22'
  script.async = true
  script.setAttribute('data-telegram-login', props.botUsername)
  script.setAttribute('data-size', props.size)
  script.setAttribute('data-radius', String(props.borderRadius))
  script.setAttribute('data-auth-url', props.redirectUrl + (props.redirectUrl.includes('?') ? '&' : '?') + 'state=' + encodeURIComponent(props.state))
  script.setAttribute('data-request-access', 'write')
  script.onload = () => {
    scriptLoaded.value = true
    if (props.autoTrigger) {
      triggerTelegramAuth()
    }
  }
  widgetContainer.value.appendChild(script)
}

onMounted(() => {
  if (typeof window !== 'undefined') {
    loadTelegramScript()
  }
})

async function triggerTelegramAuth() {
  await nextTick()
  setTimeout(() => {
    if (!widgetContainer.value) return
    const clickable = widgetContainer.value.querySelector('iframe, a, button')
    if (clickable && typeof clickable.click === 'function') {
      clickable.click()
    }
  }, 50)
  setTimeout(() => {
    if (!widgetContainer.value) return
    const clickable = widgetContainer.value.querySelector('iframe, a, button')
    if (clickable && typeof clickable.click === 'function') {
      clickable.click()
    }
  }, 250)
}
</script>

<template>
  <div ref="widgetContainer" class="telegram-login-widget flex justify-center">
    <!-- Loading state while script loads -->
    <div v-if="!scriptLoaded" class="flex items-center justify-center gap-2">
      <span class="loading loading-spinner loading-xs"></span>
      <span class="text-sm text-base-content/50">Telegram...</span>
    </div>
  </div>
</template>