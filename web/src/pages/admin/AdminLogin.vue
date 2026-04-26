<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import TechCard from '../../components/TechCard.vue'
import GradientButton from '../../components/GradientButton.vue'
import TechBackground from '../../components/TechBackground.vue'

const { t } = useI18n()
const router = useRouter()
const username = ref('')
const password = ref('')
const errMsg = ref('')
const loading = ref(false)
const checkingAdmins = ref(true)
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

onMounted(async () => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    router.push(`/${adminPrefix}`)
    return
  }

  try {
    const resp = await fetch(`/api/v1/${adminPrefix}/auth/check`)
    const json = await resp.json()
    if (json.code === 0 && !json.data.exists) {
      router.push(`/${adminPrefix}/register`)
    }
  } catch (err) {
    // If check fails, stay on login page (might be network issue)
  } finally {
    checkingAdmins.value = false
  }
})

async function login() {
  if (!username.value || !password.value) {
    errMsg.value = t('admin.enter_credentials')
    return
  }
  loading.value = true
  errMsg.value = ''
  try {
    const resp = await fetch(`/api/v1/${adminPrefix}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value }),
    })
    const json = await resp.json()
    if (json.code === 0) {
      localStorage.setItem('admin_token', json.data.token)
      localStorage.setItem('admin_user', JSON.stringify(json.data))
      router.push(`/${adminPrefix}`)
    } else {
      errMsg.value = json.message || t('admin.login_failed')
    }
  } catch (err) {
    errMsg.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="relative min-h-screen flex items-center justify-center bg-base-200 overflow-hidden">
    <TechBackground variant="blue" opacity="0.05" />

    <!-- Login card -->
    <div class="relative z-10 w-full max-w-lg mx-4">
      <TechCard variant="blue">
        <!-- Header -->
        <div class="text-center mb-8">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 mb-6 shadow-lg shadow-purple-500/25">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          </div>
          <h1 class="text-3xl md:text-4xl font-black tracking-tight">
            Pure<span class="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">Core</span>
          </h1>
          <p class="text-base-content/60 mt-3 text-lg">{{ t('admin.title') }}</p>
        </div>

        <!-- Form -->
        <form @submit.prevent="login" class="space-y-5">
          <div>
            <label class="block text-sm font-medium text-base-content/70 mb-2">{{ t('admin.username') }}</label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              </div>
              <input
                v-model="username"
                type="text"
                :placeholder="t('admin.username_placeholder')"
                class="w-full pl-10 pr-4 py-3 bg-base-200 border border-base-300/50 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
                autocomplete="username"
              />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-base-content/70 mb-2">{{ t('admin.password') }}</label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <input
                v-model="password"
                type="password"
                :placeholder="t('admin.password_placeholder')"
                class="w-full pl-10 pr-4 py-3 bg-base-200 border border-base-300/50 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
                autocomplete="current-password"
              />
            </div>
          </div>

          <div v-if="errMsg" class="p-3 bg-error/10 border border-error/20 rounded-xl text-error text-sm flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            <span>{{ errMsg }}</span>
          </div>

          <GradientButton type="submit" :loading="loading" variant="blue" size="md" class="w-full">
            <template #icon>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
              </svg>
            </template>
            {{ t('admin.sign_in') }}
          </GradientButton>
        </form>

        <div class="mt-6 text-center">
          <a href="/" class="inline-flex items-center gap-1 text-base-content/40 hover:text-blue-400 text-sm transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            {{ t('admin.back_home') }}
          </a>
        </div>
      </TechCard>
    </div>
  </div>
</template>
