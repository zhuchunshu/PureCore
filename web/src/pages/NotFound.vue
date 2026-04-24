<script setup>
import { RouterLink } from 'vue-router'
import { useI18n } from '../i18n'

const { t } = useI18n()
</script>

<template>
  <div class="hero min-h-[70vh]">
    <div class="hero-content text-center">
      <div class="max-w-md">
        <h1 class="text-8xl font-black text-primary">404</h1>
        <h2 class="text-2xl font-bold mt-4">{{ t('notfound.title') }}</h2>
        <p class="py-6 opacity-60">{{ t('notfound.description') }}</p>
        <RouterLink to="/" class="btn btn-primary btn-lg">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
          {{ t('notfound.back_home') }}
        </RouterLink>
      </div>
    </div>
  </div>
</template>
