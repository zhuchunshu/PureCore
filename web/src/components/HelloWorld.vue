<script setup>
import { ref } from 'vue'
import { useI18n } from '../i18n'
import GradientButton from './GradientButton.vue'
import TechBackground from './TechBackground.vue'

const { t } = useI18n()
const count = ref(0)
</script>

<template>
  <div class="relative min-h-[50vh] flex items-center justify-center overflow-hidden">
    <TechBackground variant="blue" />
    <div class="relative z-10 text-center">
      <h2 class="text-3xl font-black mb-4">
        <span class="bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">{{ t('hello.world') }}</span>
      </h2>
      <GradientButton @click="count++" variant="purple" size="md">
        {{ t('home.get_started') }} {{ count }}
      </GradientButton>
    </div>
  </div>
</template>
