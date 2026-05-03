<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../i18n'
import { clearTokens } from '../composables/useAuth'
import AdminNavbar from './AdminNavbar.vue'
import {
  LayoutDashboard, Users, Settings, Link,
  Home, LogOut, X
} from 'lucide-vue-next'

const { t } = useI18n()
const router = useRouter()
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'
const sidebarOpen = ref(false)

function toggleSidebar() {
  sidebarOpen.value = !sidebarOpen.value
}

function closeSidebar() {
  sidebarOpen.value = false
}

function logout() {
  clearTokens()
  localStorage.removeItem('admin_user')
  router.push(`/${adminPrefix}/login`)
}

const props = defineProps({
  profile: { type: Object, default: null },
})
</script>

<template>
  <div class="flex flex-col min-h-screen bg-base-200">
    <!-- Top navbar with hamburger for mobile -->
    <AdminNavbar :profile="profile" @toggle-sidebar="toggleSidebar" />

    <div class="flex flex-1">
      <!-- Desktop sidebar: fixed position, independent of content scroll -->
      <aside class="hidden lg:flex lg:flex-col fixed top-16 left-0 bottom-0 w-64 bg-base-100 border-r border-base-300/20 z-20">
        <div class="flex-1 overflow-y-auto">
          <ul class="menu menu-lg w-full p-3 gap-1">
            <li><router-link :to="`/${adminPrefix}`" :class="{ active: $route.path === `/${adminPrefix}` }"><LayoutDashboard :size="20" /> {{ t('admin.dashboard') }}</router-link></li>
            <li><router-link :to="`/${adminPrefix}/users`" :class="{ active: $route.path === `/${adminPrefix}/users` }"><Users :size="20" /> {{ t('admin.users_title') }}</router-link></li>
            <li><router-link :to="`/${adminPrefix}/settings`" :class="{ active: $route.path === `/${adminPrefix}/settings` }"><Settings :size="20" /> {{ t('admin.settings') }}</router-link></li>
            <li><router-link :to="`/${adminPrefix}/oauth`" :class="{ active: $route.path === `/${adminPrefix}/oauth` }"><Link :size="20" /> {{ t('admin.oauth_settings') }}</router-link></li>
          </ul>
        </div>
        <div class="p-4 border-t border-base-300/20">
          <ul class="menu menu-lg w-full gap-1">
            <li>
              <a href="/" target="_blank"><Home :size="20" /> {{ t('admin.view_site') }}</a>
            </li>
            <li><button @click="logout"><LogOut :size="20" /> {{ t('admin.logout') }}</button></li>
          </ul>
        </div>
      </aside>

      <!-- Mobile sidebar overlay -->
      <Transition name="sidebar-slide">
        <div
          v-if="sidebarOpen"
          class="fixed inset-0 z-40 lg:hidden"
          @click="closeSidebar"
        >
          <div class="absolute inset-0 bg-black/50 transition-opacity duration-300"></div>
          <aside
            class="absolute left-0 top-0 h-full w-64 bg-base-100 shadow-xl z-50"
            @click.stop
          >
            <div class="flex items-center justify-between p-4 border-b border-base-300/20">
              <span class="text-lg font-bold text-primary">PureCore</span>
              <button class="btn btn-ghost btn-sm" @click="closeSidebar">
                <X :size="18" />
              </button>
            </div>
            <ul class="menu menu-lg w-full p-3 gap-1">
              <li><router-link :to="`/${adminPrefix}`" @click="closeSidebar" :class="{ active: $route.path === `/${adminPrefix}` }"><LayoutDashboard :size="20" /> {{ t('admin.dashboard') }}</router-link></li>
              <li><router-link :to="`/${adminPrefix}/users`" @click="closeSidebar" :class="{ active: $route.path === `/${adminPrefix}/users` }"><Users :size="20" /> {{ t('admin.users_title') }}</router-link></li>
              <li><router-link :to="`/${adminPrefix}/settings`" @click="closeSidebar" :class="{ active: $route.path === `/${adminPrefix}/settings` }"><Settings :size="20" /> {{ t('admin.settings') }}</router-link></li>
              <li><router-link :to="`/${adminPrefix}/oauth`" @click="closeSidebar" :class="{ active: $route.path === `/${adminPrefix}/oauth` }"><Link :size="20" /> {{ t('admin.oauth_settings') }}</router-link></li>
              <li class="mt-auto pt-4 border-t border-base-300/20">
                <a href="/" target="_blank"><Home :size="20" /> {{ t('admin.view_site') }}</a>
              </li>
              <li><button @click="logout"><LogOut :size="20" /> {{ t('admin.logout') }}</button></li>
            </ul>
          </aside>
        </div>
      </Transition>

      <!-- Main content area: offset by sidebar width on desktop -->
      <main class="flex-1 lg:ml-64 p-4 md:p-6 overflow-y-auto">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.sidebar-slide-enter-active,
.sidebar-slide-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.sidebar-slide-enter-from,
.sidebar-slide-leave-to {
  opacity: 0;
}

.sidebar-slide-enter-from aside,
.sidebar-slide-leave-to aside {
  transform: translateX(-100%);
}

.sidebar-slide-enter-active aside,
.sidebar-slide-leave-active aside {
  transition: transform 0.3s ease;
}
</style>
