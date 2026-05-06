<script setup>
import { computed } from 'vue'
import { RouterLink, useRouter, useRoute } from 'vue-router'
import { useI18n } from '../i18n'
import { isLoggedIn, currentUser, fetchProfile, clearTokens } from '../composables/useUserAuth'
import AvatarInitials from './AvatarInitials.vue'
import { LayoutDashboard, LogOut } from 'lucide-vue-next'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'
const isAdminRoute = computed(() => route.path.startsWith(`/${adminPrefix}`))
const showUserAuth = computed(() => !isAdminRoute.value)

// Fetch profile if logged in but profile not yet loaded
if (isLoggedIn() && !currentUser.value) {
  fetchProfile()
}
</script>

<template>
  <template v-if="showUserAuth">
    <!-- Logged in: user dropdown (visible on desktop and mobile) -->
    <div v-if="isLoggedIn()" class="dropdown dropdown-end">
      <label tabindex="0" class="btn btn-ghost btn-sm hover:bg-primary/10 gap-2">
        <AvatarInitials :name="currentUser?.name" size="xs" :rounded="true" />
        <span class="hidden sm:inline text-sm font-medium text-base-content/70">{{ currentUser?.name || '' }}</span>
      </label>
      <ul tabindex="0" class="menu menu-sm dropdown-content mt-2 z-50 p-2 shadow-xl shadow-primary/10 bg-base-100/95 backdrop-blur-xl rounded-box w-48 border border-primary/20">
        <li><RouterLink to="/dashboard" class="hover:text-primary hover:bg-primary/5 rounded-lg transition-colors"><LayoutDashboard :size="16" class="inline mr-2" />{{ t('user.dashboard') }}</RouterLink></li>
        <li><button @click="clearTokens(); router.push('/')" class="hover:text-error hover:bg-error/5 rounded-lg transition-colors w-full text-left"><LogOut :size="16" class="inline mr-2" />{{ t('admin.logout') }}</button></li>
      </ul>
    </div>
    <!-- Not logged in: login link only (register removed to avoid mobile overlap) -->
    <RouterLink v-else to="/login" class="btn btn-ghost btn-sm text-sm font-medium text-base-content/70 hover:text-primary hover:bg-primary/5 transition-colors">{{ t('user.login_title') }}</RouterLink>
  </template>
</template>
