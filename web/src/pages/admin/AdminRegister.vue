<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import TechCard from '../../components/TechCard.vue'
import GradientButton from '../../components/GradientButton.vue'
import TechBackground from '../../components/TechBackground.vue'

const { t } = useI18n()
const router = useRouter()
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'
const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const name = ref('')
const errMsg = ref('')
const loading = ref(false)
const hasAdmins = ref(true)

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
      hasAdmins.value = false
    } else {
      router.push(`/${adminPrefix}/login`)
    }
  } catch (err) {
    errMsg.value = t('admin.network_error')
  }
})

async function register() {
  if (!username.value || !password.value || !name.value) {
    errMsg.value = t('admin.enter_credentials')
    return
  }
  if (password.value !== confirmPassword.value) {
    errMsg.value = t('admin.passwords_not_match')
    return
  }
  loading.value = true
  errMsg.value = ''
  try {
    const resp = await fetch(`/api/v1/${adminPrefix}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value, name: name.value }),
    })
    const json = await resp.json()
    if (json.code === 0) {
      localStorage.setItem('admin_token', json.data.token)
      localStorage.setItem('admin_user', JSON.stringify(json.data))
      router.push(`/${adminPrefix}`)
    } else {
      errMsg.value = json.message || t('admin.register_failed')
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
    <TechBackground variant="emerald" opacity="0.05" />

    <!-- Register card -->
    <div class="relative z-10 w-full max-w-lg mx-4">
      <TechCard variant="emerald">
        <!-- Header -->
        <div class="text-center mb-8">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 mb-6 shadow-lg shadow-teal-500/25">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
            </svg>
          </div>
          <h1 class="text-3xl md:text-4xl font-black tracking-tight">
            Pure<span class="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-teal-400">Core</span>
          </h1>
          <p class="text-base-content/60 mt-3 text-lg">{{ t('admin.register_title') }}</p>
          <p v-if="!hasAdmins" class="text-warning/80 text-sm mt-2 bg-warning/10 py-1.5 px-3 rounded-lg inline-block">{{ t('admin.no_admin_redirect') }}</p>
        </div>

        <!-- Form -->
        <form @submit.prevent="register" class="space-y-5">
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
                class="w-full pl-10 pr-4 py-3 bg-base-200 border border-base-300/50 rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-transparent transition-all"
                autocomplete="username"
              />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-base-content/70 mb-2">{{ t('admin.name') }}</label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0a2 2 0 104 0m-5 8a2 2 0 100-4 2 2 0 000 4zm0 0c1.306 0 2.417.835 2.83 2M9 14a3.001 3.001 0 00-2.83 2M15 11h3m-3 4h2" />
                </svg>
              </div>
              <input
                v-model="name"
                type="text"
                :placeholder="t('admin.name_placeholder')"
                class="w-full pl-10 pr-4 py-3 bg-base-200 border border-base-300/50 rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-transparent transition-all"
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
                class="w-full pl-10 pr-4 py-3 bg-base-200 border border-base-300/50 rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-transparent transition-all"
                autocomplete="new-password"
              />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-base-content/70 mb-2">{{ t('admin.confirm_password') }}</label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
              <input
                v-model="confirmPassword"
                type="password"
                :placeholder="t('admin.confirm_password_placeholder')"
                class="w-full pl-10 pr-4 py-3 bg-base-200 border border-base-300/50 rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-transparent transition-all"
                autocomplete="new-password"
              />
            </div>
          </div>

          <div v-if="errMsg" class="p-3 bg-error/10 border border-error/20 rounded-xl text-error text-sm flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            <span>{{ errMsg }}</span>
          </div>

          <GradientButton type="submit" :loading="loading" variant="emerald" size="md" class="w-full">
            <template #icon>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </template>
            {{ t('admin.register_button') }}
          </GradientButton>
        </form>

        <div class="mt-6 text-center">
          <a href="/" class="inline-flex items-center gap-1 text-base-content/40 hover:text-emerald-400 text-sm transition-colors">
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
