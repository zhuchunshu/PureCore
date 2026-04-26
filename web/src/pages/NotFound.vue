<script setup>
import { RouterLink } from 'vue-router'
import { useI18n } from '../i18n'
import GradientButton from '../components/GradientButton.vue'
import TechBackground from '../components/TechBackground.vue'

const { t } = useI18n()
</script>

<template>
  <div class="relative min-h-[70vh] flex items-center justify-center overflow-hidden">
    <TechBackground variant="purple" />

    <div class="relative z-10 text-center px-4">
      <div class="max-w-md mx-auto">
        <h1 class="text-[10rem] md:text-[12rem] font-black leading-none bg-gradient-to-br from-blue-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">404</h1>
        <h2 class="text-2xl md:text-3xl font-bold text-white mt-2">{{ t('notfound.title') }}</h2>
        <p class="py-6 text-base-content/60 text-lg">{{ t('notfound.description') }}</p>
        <GradientButton to="/" variant="purple" size="lg">
          <template #icon>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
            </svg>
          </template>
          {{ t('notfound.back_home') }}
        </GradientButton>
      </div>
    </div>
  </div>
</template>
