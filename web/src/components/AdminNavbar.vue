<script setup>
import { useRouter } from 'vue-router'
import { useI18n } from '../i18n'
import LanguageSwitcher from './LanguageSwitcher.vue'
import ThemeSwitcher from './ThemeSwitcher.vue'

const { t } = useI18n()
const router = useRouter()
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

const emit = defineEmits(['toggle-sidebar'])

function logout() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_user')
  router.push(`/${adminPrefix}/login`)
}

defineProps({
  profile: { type: Object, default: null },
})
</script>

<template>
  <nav class="sticky top-0 z-30 backdrop-blur-xl bg-base-100/80 border-b border-base-300/20 shadow-sm">
    <div class="px-4">
      <div class="flex items-center justify-between h-16">
        <!-- Left: hamburger + brand -->
        <div class="flex items-center gap-3">
          <button class="btn btn-ghost btn-sm lg:hidden" @click="emit('toggle-sidebar')">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-md shadow-teal-500/20">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <span class="text-lg font-black tracking-tight">
              <span class="bg-gradient-to-r from-emerald-400 to-teal-400 bg-clip-text text-transparent">Pure</span><span class="text-base-content/80">Core</span>
            </span>
          </div>
        </div>

        <!-- Right: theme, language, profile -->
        <div class="flex items-center gap-2">
          <ThemeSwitcher />
          <LanguageSwitcher />
          <div v-if="profile" class="dropdown dropdown-end">
            <label tabindex="0" class="btn btn-ghost btn-sm gap-2">
              <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-xs font-bold text-white shadow-sm">
                {{ (profile.name || 'A')[0].toUpperCase() }}
              </div>
              <span class="hidden sm:inline text-sm">{{ profile.name }}</span>
              <span class="badge badge-xs bg-emerald-500/20 text-emerald-400 border-emerald-500/30">{{ profile.role }}</span>
            </label>
            <ul tabindex="0" class="menu menu-sm dropdown-content mt-2 z-50 p-2 shadow-xl bg-base-100 rounded-box w-48 border border-base-300/20">
              <li><a href="/" target="_blank" class="hover:text-blue-400"><span>🏠</span> {{ t('admin.view_site') }}</a></li>
              <li><button @click="logout" class="hover:text-red-400"><span>🚪</span> {{ t('admin.logout') }}</button></li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>
