<script setup>
import { useI18n } from '../i18n'

const { t } = useI18n()
const currentYear = new Date().getFullYear()
</script>

<template>
  <footer class="flex items-center justify-between px-6 py-4 bg-base-100 border-t border-base-300/50 text-base-content">
    <div class="flex items-center gap-4">
      <span class="text-sm font-semibold"><span class="text-primary">Pure</span>Core</span>
      <span class="text-xs opacity-50">© {{ currentYear }}</span>
    </div>
    <div class="flex items-center gap-4">
      <a href="/docs/zh/README.md" class="link link-hover text-xs opacity-50">{{ t('nav.docs') }}</a>
      <a href="/docs/zh/API.md" class="link link-hover text-xs opacity-50">API</a>
      <a href="https://github.com/zhuchunshu/PureCore" target="_blank" class="link link-hover text-xs opacity-50">GitHub</a>
    </div>
  </footer>
</template>
