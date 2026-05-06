<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'
import AdminLayout from './components/AdminLayout.vue'
import Footer from './components/Footer.vue'
import BackendError from './components/BackendError.vue'
import ToastContainer from './components/ToastContainer.vue'
import { useBackendHealth } from './composables/useBackendHealth'

const route = useRoute()
const { isBackendReachable, hasChecked } = useBackendHealth()

const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

const isHomePage = computed(() => route.path === '/')
const isAdminDashboard = computed(() => route.path === `/${adminPrefix}`)
const isAdminPage = computed(() => route.path.startsWith(`/${adminPrefix}`) && !route.path.endsWith('/login') && !route.path.endsWith('/register'))
const isDashboardPage = computed(() => route.path.startsWith('/dashboard'))
const showFooter = computed(() => !isHomePage.value && !isDashboardPage.value)
const showSpinner = computed(() => !isHomePage.value && !hasChecked.value)
const showBackendError = computed(() => !isHomePage.value && hasChecked.value && !isBackendReachable.value)
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <!-- Loading spinner while checking backend health on non-home routes -->
    <template v-if="showSpinner">
      <div class="min-h-screen bg-base-200">
        <!-- Navbar skeleton -->
        <div class="h-16 bg-base-100 border-b border-base-300/20 flex items-center px-6">
          <div class="skeleton h-8 w-32 rounded-lg"></div>
          <div class="hidden sm:flex gap-4 ml-auto">
            <div class="skeleton h-6 w-16 rounded-lg"></div>
            <div class="skeleton h-6 w-16 rounded-lg"></div>
            <div class="skeleton h-6 w-16 rounded-lg"></div>
            <div class="skeleton h-6 w-16 rounded-lg"></div>
          </div>
        </div>
        <!-- Content skeleton -->
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
          <div class="skeleton h-8 w-64 mb-4 rounded-lg"></div>
          <div class="skeleton h-4 w-full max-w-2xl mb-2 rounded"></div>
          <div class="skeleton h-4 w-3/4 max-w-xl mb-8 rounded"></div>
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            <div class="skeleton h-48 rounded-2xl"></div>
            <div class="skeleton h-48 rounded-2xl"></div>
            <div class="skeleton h-48 rounded-2xl"></div>
          </div>
        </div>
      </div>
    </template>
    <!-- Backend is unreachable: show error page -->
    <template v-else-if="showBackendError">
      <BackendError />
    </template>
    <!-- Admin pages (dashboard, settings, and future admin routes) use AdminLayout -->
    <template v-else-if="isAdminPage">
      <AdminLayout>
        <router-view :key="$route.fullPath" />
      </AdminLayout>
    </template>
    <!-- Public pages use plain Navbar + Footer -->
    <template v-else>
      <Navbar />
      <main class="flex-1">
        <router-view :key="$route.fullPath" />
      </main>
      <Footer v-if="showFooter" />
    </template>

    <!-- Global toast notifications -->
    <ToastContainer />
  </div>
</template>
