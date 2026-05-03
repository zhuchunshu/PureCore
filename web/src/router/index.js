import { createRouter, createWebHistory } from 'vue-router'
import { routes } from './routes'
import { userAPI, adminAPI } from '@/services/api'

export function createSSRRouter(history) {
  const router = createRouter({
    history,
    routes,
  })

  // Client-side only: redirect to login with return URL for protected routes
  if (typeof window !== 'undefined') {
    const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

    router.beforeEach((to, _from, next) => {
      // Check if any matched route requires authentication via meta field
      const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
      if (!requiresAuth) return next()

      // Determine which auth type is needed and the corresponding login path
      const authType = to.matched.find(r => r.meta.authType)?.meta.authType
      const loginPath = authType === 'admin' ? `/${adminPrefix}/login` : '/login'

      // Choose the correct API service based on auth type
      const api = authType === 'admin' ? adminAPI : userAPI

      if (!api.isLoggedIn()) {
        return next({ path: loginPath, query: { redirect: to.fullPath } })
      }

      next()
    })
  }

  return router
}

// Default router instance (for potential direct usage, guards included if client-side)
const router = typeof window !== 'undefined' 
  ? createSSRRouter(createWebHistory())
  : createRouter({ history: null, routes })

export default router
