<script setup>
import { RouterLink } from 'vue-router'
import LanguageSwitcher from './LanguageSwitcher.vue'
import ThemeSwitcher from './ThemeSwitcher.vue'
import { useI18n } from '../i18n'

const { t } = useI18n()
</script>

<template>
  <nav class="sticky top-0 z-30 backdrop-blur-xl bg-base-100/80 border-b border-primary/10 shadow-lg shadow-primary/5">
    <div class="absolute inset-0 bg-gradient-to-r from-primary/5 via-transparent to-secondary/5 opacity-50 pointer-events-none"></div>
    <div class="max-w-7xl mx-auto px-4 relative">
      <div class="flex items-center justify-between h-16">
        <!-- Logo -->
        <RouterLink to="/" class="flex items-center gap-2 group">
          <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg shadow-purple-500/30 group-hover:shadow-purple-500/50 group-hover:scale-105 transition-all duration-300 ring-1 ring-white/10">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          </div>
          <span class="text-xl font-black tracking-tight">
            <span class="bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">Pure</span><span class="text-base-content/80">Core</span>
          </span>
        </RouterLink>

        <!-- Desktop nav links -->
        <div class="hidden lg:flex items-center gap-1">
          <RouterLink to="/" class="relative px-4 py-2 text-sm font-medium text-base-content/70 hover:text-primary rounded-lg hover:bg-primary/5 transition-all duration-300 before:absolute before:bottom-0 before:left-1/2 before:-translate-x-1/2 before:w-0 before:h-0.5 before:bg-primary before:transition-all before:duration-300 hover:before:w-3/4">{{ t('nav.home') }}</RouterLink>
          <a href="/docs/zh/API.md" class="relative px-4 py-2 text-sm font-medium text-base-content/70 hover:text-secondary rounded-lg hover:bg-secondary/5 transition-all duration-300 before:absolute before:bottom-0 before:left-1/2 before:-translate-x-1/2 before:w-0 before:h-0.5 before:bg-secondary before:transition-all before:duration-300 hover:before:w-3/4">{{ t('nav.docs') }}</a>
          <a href="https://github.com/zhuchunshu/PureCore" target="_blank" class="relative px-4 py-2 text-sm font-medium text-base-content/70 hover:text-accent rounded-lg hover:bg-accent/5 transition-all duration-300 before:absolute before:bottom-0 before:left-1/2 before:-translate-x-1/2 before:w-0 before:h-0.5 before:bg-accent before:transition-all before:duration-300 hover:before:w-3/4">GitHub</a>
        </div>

        <!-- Right side controls -->
        <div class="flex items-center gap-1">
          <ThemeSwitcher />
          <LanguageSwitcher />

          <!-- Mobile hamburger -->
          <div class="dropdown dropdown-end lg:hidden">
            <label tabindex="0" class="btn btn-ghost btn-sm hover:bg-primary/10">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </label>
            <ul tabindex="0" class="menu menu-sm dropdown-content mt-2 z-50 p-2 shadow-xl shadow-primary/10 bg-base-100/95 backdrop-blur-xl rounded-box w-52 border border-primary/20">
              <li><RouterLink to="/" class="hover:text-primary hover:bg-primary/5 rounded-lg transition-colors">{{ t('nav.home') }}</RouterLink></li>
              <li><a href="/docs/zh/API.md" class="hover:text-secondary hover:bg-secondary/5 rounded-lg transition-colors">{{ t('nav.docs') }}</a></li>
              <li><a href="https://github.com/zhuchunshu/PureCore" target="_blank" class="hover:text-accent hover:bg-accent/5 rounded-lg transition-colors">GitHub</a></li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>
