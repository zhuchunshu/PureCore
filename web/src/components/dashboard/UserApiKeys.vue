<script setup>
import { useI18n } from '../../i18n'
import TechCard from '../TechCard.vue'
import { Key, Plus } from 'lucide-vue-next'

const { t } = useI18n()
</script>

<template>
  <TechCard variant="blue" padded>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-5">
      <div>
        <h2 class="text-lg font-bold text-base-content/80 flex items-center gap-2">
          <Key :size="20" />
          {{ t('user.api_keys') }}
        </h2>
        <p class="text-sm text-base-content/50 mt-0.5">{{ t('user.api_keys_description') }}</p>
      </div>
      <button class="btn btn-primary btn-sm rounded-xl">
        <Plus :size="16" />
        {{ t('user.create_api_key') }}
      </button>
    </div>
    <div class="text-center py-10 text-base-content/40">
      <Key :size="48" class="mx-auto mb-3 opacity-30" />
      <span class="text-sm">{{ t('user.no_api_keys') }}</span>
    </div>
  </TechCard>
</template>
