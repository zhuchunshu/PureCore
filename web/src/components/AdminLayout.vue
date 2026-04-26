<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../i18n'
import AdminNavbar from './AdminNavbar.vue'

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
  localStorage.removeItem('admin_token')
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
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <ul class="menu p-4 gap-1">
              <li><a class="active" @click="closeSidebar"><span>📊</span> {{ t('admin.dashboard') }}</a></li>
              <li><a @click="closeSidebar"><span>👥</span> {{ t('admin.users') }}</a></li>
              <li><a @click="closeSidebar"><span>⚙️</span> {{ t('admin.settings') }}</a></li>
              <li class="mt-auto pt-4 border-t border-base-300/20">
                <a href="/" target="_blank"><span>🏠</span> {{ t('admin.view_site') }}</a>
              </li>
              <li><button @click="logout"><span>🚪</span> {{ t('admin.logout') }}</button></li>
            </ul>
          </aside>
        </div>
      </Transition>

      <!-- Desktop sidebar (always visible on lg+) -->
      <aside class="w-64 bg-base-100 border-r border-base-300/20 hidden lg:flex lg:flex-col">
        <div class="flex-1 overflow-y-auto">
          <ul class="menu p-4 gap-1">
            <li><a class="active"><span>📊</span> {{ t('admin.dashboard') }}</a></li>
            <li><a><span>👥</span> {{ t('admin.users') }}</a></li>
            <li><a><span>⚙️</span> {{ t('admin.settings') }}</a></li>
          </ul>
        </div>
        <div class="p-4 border-t border-base-300/20">
          <ul class="menu gap-1">
            <li>
              <a href="/" target="_blank"><span>🏠</span> {{ t('admin.view_site') }}</a>
            </li>
            <li><button @click="logout"><span>🚪</span> {{ t('admin.logout') }}</button></li>
          </ul>
        </div>
      </aside>

      <!-- Main content area -->
      <main class="flex-1 p-4 md:p-6 overflow-y-auto">
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
