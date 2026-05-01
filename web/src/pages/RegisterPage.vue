<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../i18n'
import { useSEO } from '../composables/useSEO'
import { setTokens, accessToken } from '../composables/useUserAuth'
import TurnstileWidget from '../components/TurnstileWidget.vue'
import OAuthButtons from '../components/OAuthButtons.vue'

const { t } = useI18n()
useSEO({
  title: t('user.register_title'),
  description: t('user.register_title'),
})
const router = useRouter()
const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const errMsg = ref('')
const loading = ref(false)
const turnstileToken = ref('')
const turnstileRef = ref(null)
const eyeClosed = ref(false)

onMounted(() => {
  if (accessToken.value) {
    router.push('/')
  }
})

function onPasswordFocus() {
  eyeClosed.value = true
}

function onPasswordBlur() {
  eyeClosed.value = false
}

async function register() {
  if (!name.value || !email.value || !password.value) {
    errMsg.value = t('user.enter_credentials')
    return
  }
  if (password.value !== confirmPassword.value) {
    errMsg.value = t('admin.passwords_not_match')
    return
  }
  if (turnstileRef.value && turnstileRef.value.isEnabled && !turnstileRef.value.verified) {
    errMsg.value = 'Please complete the CAPTCHA verification before submitting'
    return
  }
  loading.value = true
  errMsg.value = ''
  try {
    const resp = await fetch('/api/v1/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: name.value, email: email.value, password: password.value, turnstile_token: turnstileToken.value }),
    })
    const json = await resp.json()
    if (json.code === 0) {
      setTokens(json.data.token, json.data.refresh_token)
      localStorage.setItem('user_profile', JSON.stringify({ name: json.data.name, email: json.data.email }))
      router.push('/')
    } else {
      errMsg.value = json.message || t('user.register_failed')
    }
  } catch (err) {
    errMsg.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="relative flex min-h-[calc(100vh-4rem)] items-center justify-center overflow-hidden bg-base-200">
    <!-- Animated grid background -->
    <div class="absolute inset-0 opacity-10">
      <div class="absolute inset-0" style="background-image: radial-gradient(circle at 1px 1px, oklch(var(--p)/0.15) 1px, transparent 0); background-size: 40px 40px;"></div>
    </div>

    <!-- Glow orbs - shifted positions for variety -->
    <div class="absolute top-1/3 -right-20 w-80 h-80 bg-secondary/20 rounded-full blur-3xl animate-pulse"></div>
    <div class="absolute bottom-1/4 -left-20 w-96 h-96 bg-accent/15 rounded-full blur-3xl animate-pulse" style="animation-delay: 2s;"></div>
    <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-64 h-64 bg-primary/10 rounded-full blur-3xl"></div>

    <div class="relative z-10 w-full max-w-5xl mx-auto px-4 py-8">
      <div class="flex flex-col lg:flex-row items-center gap-8 lg:gap-16">
        <!-- Left: Illustration -->
        <div class="flex-1 hidden lg:flex items-center justify-center">
          <div :class="['relative w-80 h-80 transition-all duration-300', { 'scale-105': eyeClosed }]">
            <img src="/assets/img/undraw_enter-password_1kl4.svg" alt="Register illustration" class="w-full h-full object-contain" />
          </div>
        </div>

        <!-- Right: Form -->
        <div class="flex-1 w-full max-w-md mx-auto lg:mx-0">
          <!-- Site name -->
          <div class="text-center lg:text-left mb-8">
            <h1 class="text-4xl lg:text-5xl font-black tracking-tight text-base-content">
              Pure<span class="bg-gradient-to-r from-primary via-secondary to-accent bg-clip-text text-transparent">Core</span>
            </h1>
            <p class="text-base-content/60 mt-2 text-lg font-light">{{ t('user.register_title') }}</p>
          </div>

          <!-- Form card -->
          <div class="backdrop-blur-xl bg-base-200/40 border border-base-content/10 rounded-3xl p-8 shadow-2xl">
            <form @submit.prevent="register" class="space-y-5">
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-2 ml-1">{{ t('user.name') }}</label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <svg class="w-5 h-5 text-base-content/30 group-focus-within:text-primary transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z"/>
                    </svg>
                  </div>
                  <input
                    v-model="name"
                    type="text"
                    :placeholder="t('user.name_placeholder')"
                    class="w-full pl-12 pr-4 py-3.5 bg-base-100/50 border border-base-content/10 rounded-2xl text-base-content placeholder:text-base-content/30 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-transparent transition-all text-sm"
                    autocomplete="name"
                  />
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-2 ml-1">{{ t('user.email') }}</label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <svg class="w-5 h-5 text-base-content/30 group-focus-within:text-primary transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75"/>
                    </svg>
                  </div>
                  <input
                    v-model="email"
                    type="email"
                    :placeholder="t('user.email_placeholder')"
                    class="w-full pl-12 pr-4 py-3.5 bg-base-100/50 border border-base-content/10 rounded-2xl text-base-content placeholder:text-base-content/30 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-transparent transition-all text-sm"
                    autocomplete="email"
                  />
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-2 ml-1">{{ t('admin.password') }}</label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <svg class="w-5 h-5 text-base-content/30 group-focus-within:text-primary transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z"/>
                    </svg>
                  </div>
                  <input
                    v-model="password"
                    type="password"
                    :placeholder="t('admin.password_placeholder')"
                    class="w-full pl-12 pr-4 py-3.5 bg-base-100/50 border border-base-content/10 rounded-2xl text-base-content placeholder:text-base-content/30 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-transparent transition-all text-sm"
                    autocomplete="new-password"
                    @focus="onPasswordFocus"
                    @blur="onPasswordBlur"
                  />
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-2 ml-1">{{ t('admin.confirm_password') }}</label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <svg class="w-5 h-5 text-base-content/30 group-focus-within:text-primary transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"/>
                    </svg>
                  </div>
                  <input
                    v-model="confirmPassword"
                    type="password"
                    :placeholder="t('admin.confirm_password_placeholder')"
                    class="w-full pl-12 pr-4 py-3.5 bg-base-100/50 border border-base-content/10 rounded-2xl text-base-content placeholder:text-base-content/30 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-transparent transition-all text-sm"
                    autocomplete="new-password"
                    @focus="onPasswordFocus"
                    @blur="onPasswordBlur"
                  />
                </div>
              </div>

              <div v-if="errMsg" class="flex items-center gap-3 p-4 bg-error/10 border border-error/20 rounded-2xl text-error text-sm">
                <svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z"/></svg>
                <span>{{ errMsg }}</span>
              </div>

              <TurnstileWidget ref="turnstileRef" context="turnstile_public_login" @verified="turnstileToken = $event"/>

              <button
                type="submit"
                :disabled="loading"
                class="relative w-full py-3.5 px-4 bg-gradient-to-r from-primary via-secondary to-accent rounded-2xl text-primary-content font-semibold text-sm shadow-lg shadow-primary/25 hover:shadow-primary/40 hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
              >
                <span v-if="loading" class="flex items-center justify-center gap-2">
                  <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
                  {{ t('user.register_button') }}
                </span>
                <span v-else>{{ t('user.register_button') }}</span>
              </button>
            </form>

            <OAuthButtons mode="register" />

            <div class="mt-6 text-center lg:text-left space-y-4">
              <p class="text-sm text-base-content/40">
                {{ t('user.have_account') }}
                <a href="/login" class="text-primary hover:text-primary/80 font-semibold transition-colors underline decoration-primary/30 underline-offset-4"> {{ t('user.sign_in') }} </a>
              </p>
              <a href="/" class="inline-flex items-center gap-1.5 text-base-content/30 hover:text-primary text-sm transition-colors">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
                </svg>
                {{ t('admin.back_home') }}
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
