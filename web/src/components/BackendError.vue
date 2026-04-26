<script setup>
import { useRouter } from 'vue-router'
import { useI18n } from '../i18n'
import { useBackendHealth } from '../composables/useBackendHealth'
import LanguageSwitcher from './LanguageSwitcher.vue'
import ThemeSwitcher from './ThemeSwitcher.vue'
import GradientButton from './GradientButton.vue'
import TechBackground from './TechBackground.vue'

const { t } = useI18n()
const router = useRouter()
const { checkHealth, isBackendReachable } = useBackendHealth()

const retry = () => {
  checkHealth()
  if (isBackendReachable.value) {
    window.location.reload()
  }
}

const goHome = () => {
  router.push('/')
}
</script>

<template>
  <div class="relative min-h-screen flex flex-col items-center justify-center bg-base-200 overflow-hidden">
    <TechBackground variant="blue" />

    <div class="absolute top-0 right-0 z-10 p-2 flex gap-1">
      <ThemeSwitcher />
      <LanguageSwitcher />
    </div>

    <div class="relative z-10 w-full max-w-md mx-4">
      <div class="backdrop-blur-xl bg-base-300/40 rounded-2xl shadow-2xl border border-red-500/20 p-8 md:p-10 text-center">
        <div class="flex justify-center mb-6">
          <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-red-500/20 to-orange-500/20 flex items-center justify-center shadow-lg shadow-red-500/10">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
        </div>
        <h2 class="text-2xl font-bold text-base-content">{{ t('backend.title') }}</h2>
        <p class="text-base-content/60 mt-3 mb-8">{{ t('backend.message') }}</p>
        <div class="flex flex-col sm:flex-row items-center justify-center gap-3">
          <GradientButton @click="retry" variant="blue">
            <template #icon>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </template>
            {{ t('backend.retry') }}
          </GradientButton>
          <GradientButton @click="goHome" variant="purple">
            <template #icon>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
              </svg>
            </template>
            {{ t('backend.back_home') }}
          </GradientButton>
        </div>
      </div>
    </div>
  </div>
</template>
