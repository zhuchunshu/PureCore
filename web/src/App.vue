<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'
import AdminLayout from './components/AdminLayout.vue'
import Footer from './components/Footer.vue'
import BackendError from './components/BackendError.vue'
import { useBackendHealth } from './composables/useBackendHealth'

const route = useRoute()
const { isBackendReachable, hasChecked } = useBackendHealth()

const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

const isHomePage = computed(() => route.path === '/')
const isAdminDashboard = computed(() => route.path === `/${adminPrefix}`)
const isAdminPage = computed(() => route.path.startsWith(`/${adminPrefix}`))
const showSpinner = computed(() => !isHomePage.value && !hasChecked.value)
const showBackendError = computed(() => !isHomePage.value && hasChecked.value && !isBackendReachable.value)
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <!-- Loading spinner while checking backend health on non-home routes -->
    <template v-if="showSpinner">
      <div class="min-h-screen flex items-center justify-center bg-base-200">
        <span class="loading loading-spinner loading-lg text-primary"></span>
      </div>
    </template>
    <!-- Backend is unreachable: show error page -->
    <template v-else-if="showBackendError">
      <BackendError />
    </template>
    <!-- Admin dashboard uses AdminLayout (own navbar + sidebar) -->
    <template v-else-if="isAdminDashboard">
      <AdminLayout>
        <router-view />
      </AdminLayout>
    </template>
    <!-- Public pages use plain Navbar + Footer -->
    <template v-else>
      <Navbar />
      <main class="flex-1">
        <router-view />
      </main>
      <Footer />
    </template>
  </div>
</template>
