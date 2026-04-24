<script setup>
import { RouterLink } from 'vue-router'
import LanguageSwitcher from './LanguageSwitcher.vue'
import { useI18n } from '../i18n'

const { t } = useI18n()
</script>

<template>
  <div class="navbar bg-base-100 shadow-sm">
    <div class="navbar-start">
      <div class="dropdown">
        <label tabindex="0" class="btn btn-ghost lg:hidden">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </label>
        <ul tabindex="0" class="menu menu-sm dropdown-content mt-3 z-50 p-2 shadow bg-base-100 rounded-box w-52">
          <li><RouterLink to="/">{{ t('nav.home') }}</RouterLink></li>
          <li><a href="/docs/zh/API.md">{{ t('nav.docs') }}</a></li>
          <li><a href="https://github.com/zhuchunshu/PureCore" target="_blank">GitHub</a></li>
        </ul>
      </div>
      <RouterLink to="/" class="btn btn-ghost text-xl font-bold">
        Pure<span class="text-primary">Core</span>
      </RouterLink>
    </div>
    <div class="navbar-center hidden lg:flex">
      <ul class="menu menu-horizontal px-1">
        <li><RouterLink to="/">{{ t('nav.home') }}</RouterLink></li>
        <li><a href="/docs/zh/API.md">{{ t('nav.docs') }}</a></li>
        <li><a href="https://github.com/zhuchunshu/PureCore" target="_blank">GitHub</a></li>
      </ul>
    </div>
    <div class="navbar-end">
      <LanguageSwitcher />
    </div>
  </div>
</template>
