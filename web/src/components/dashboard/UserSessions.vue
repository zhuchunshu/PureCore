<script setup>
import { useI18n } from '../../i18n'
import TechCard from '../TechCard.vue'

const { t } = useI18n()
</script>

<template>
  <TechCard variant="blue" padded>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-5">
      <div>
        <h2 class="text-lg font-bold text-base-content/80">📱 {{ t('user.sessions') }}</h2>
        <p class="text-sm text-base-content/50 mt-0.5">{{ t('user.sessions_description') }}</p>
      </div>
      <button class="btn btn-ghost btn-sm rounded-xl text-red-400 hover:bg-red-500/10">
        {{ t('user.revoke_all') }}
      </button>
    </div>
    <div class="text-center py-10 text-base-content/40">
      <span class="text-4xl block mb-2">📱</span>
      <span class="text-sm">{{ t('user.no_sessions') }}</span>
    </div>
  </TechCard>
</template>
