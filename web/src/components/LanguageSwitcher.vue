<script setup>
import { useI18n } from '../i18n'

const { currentLocale, setLocale, t } = useI18n()
</script>

<template>
  <div class="dropdown dropdown-end">
    <label tabindex="0" class="btn btn-ghost btn-sm">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
      </svg>
      <span class="hidden sm:inline">{{ t(currentLocale === 'zh' ? 'language.zh' : 'language.en') }}</span>
    </label>
    <ul tabindex="0" class="menu menu-sm dropdown-content mt-2 z-50 p-2 shadow bg-base-100 rounded-box w-36">
      <li>
        <button :class="{ 'active': currentLocale === 'zh' }" @click="setLocale('zh')">
          🇨🇳 {{ t('language.zh') }}
        </button>
      </li>
      <li>
        <button :class="{ 'active': currentLocale === 'en' }" @click="setLocale('en')">
          🇺🇸 {{ t('language.en') }}
        </button>
      </li>
    </ul>
  </div>
</template>
