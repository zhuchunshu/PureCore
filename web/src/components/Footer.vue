<script setup>
import { useI18n } from '../i18n'

const { t } = useI18n()
const currentYear = new Date().getFullYear()
</script>

<template>
  <footer class="relative border-t border-primary/10 bg-base-100/80 backdrop-blur-xl shadow-inner shadow-primary/5">
    <div class="absolute inset-0 bg-gradient-to-r from-primary/3 via-transparent to-secondary/3 opacity-50 pointer-events-none"></div>
    <div class="max-w-6xl mx-auto px-4 py-8 flex flex-col sm:flex-row items-center justify-between gap-6 relative">
      <div class="flex items-center gap-4">
        <div class="w-8 h-8 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-md shadow-purple-500/20 ring-1 ring-white/10">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
          </svg>
        </div>
        <div class="flex flex-col">
          <span class="text-sm font-bold tracking-tight">
            <span class="bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">Pure</span><span class="text-base-content/70">Core</span>
          </span>
          <span class="text-xs text-base-content/40">© {{ currentYear }}</span>
        </div>
      </div>
      <div class="flex items-center gap-6">
        <a href="/docs/zh/README.md" class="relative text-sm text-base-content/60 hover:text-blue-400 transition-all duration-300 before:absolute before:-bottom-0.5 before:left-0 before:w-0 before:h-px before:bg-blue-400 before:transition-all before:duration-300 hover:before:w-full">{{ t('nav.docs') }}</a>
        <a href="/docs/zh/API.md" class="relative text-sm text-base-content/60 hover:text-secondary transition-all duration-300 before:absolute before:-bottom-0.5 before:left-0 before:w-0 before:h-px before:bg-secondary before:transition-all before:duration-300 hover:before:w-full">{{ t('nav.api') }}</a>
        <a href="https://github.com/zhuchunshu/PureCore" target="_blank" class="relative text-sm text-base-content/60 hover:text-accent transition-all duration-300 before:absolute before:-bottom-0.5 before:left-0 before:w-0 before:h-px before:bg-accent before:transition-all before:duration-300 hover:before:w-full">GitHub</a>
      </div>
    </div>
  </footer>
</template>
