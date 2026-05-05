<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { IconBrandTelegram } from '@tabler/icons-vue'

const props = defineProps({
  botUsername: { type: String, required: true },
  redirectUrl: { type: String, required: true },
  state: { type: String, required: true },
  size: { type: String, default: 'large' },
  borderRadius: { type: Number, default: 12 },
})

const emit = defineEmits(['callback'])
const scriptLoaded = ref(false)
const widgetContainer = ref(null)

// Inject the Telegram widget script
function loadTelegramScript() {
  if (document.getElementById('telegram-widget-script')) {
    scriptLoaded.value = true
    initWidget()
    return
  }
  const script = document.createElement('script')
  script.id = 'telegram-widget-script'
  script.src = 'https://telegram.org/js/telegram-widget.js?22'
  script.async = true
  script.setAttribute('data-telegram-login', props.botUsername)
  script.setAttribute('data-size', props.size)
  script.setAttribute('data-radius', String(props.borderRadius))
  script.setAttribute('data-auth-url', props.redirectUrl + (props.redirectUrl.includes('?') ? '&' : '?') + 'state=' + encodeURIComponent(props.state))
  script.setAttribute('data-request-access', 'write')
  script.onload = () => {
    scriptLoaded.value = true
  }
  if (widgetContainer.value) {
    widgetContainer.value.appendChild(script)
  }
}

// The Telegram widget will render itself inside the container
// and handle the auth flow automatically.

function initWidget() {
  // Telegram widget callback is handled via window.TelegramLoginWidget.onAuth
  // but we rely on the redirect flow (data-auth-url) for server-side verification.
}

onMounted(() => {
  if (typeof window !== 'undefined') {
    loadTelegramScript()
  }
})

onUnmounted(() => {
  // Clean up widget if needed
  const existing = document.getElementById('telegram-widget-script')
  if (scriptLoaded.value && !existing) {
    // Don't remove the script as it may be used by other instances
  }
})
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