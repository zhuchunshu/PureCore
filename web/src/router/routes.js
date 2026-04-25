import HomePage from '../pages/HomePage.vue'
import NotFound from '../pages/NotFound.vue'
import AdminLogin from '../pages/admin/AdminLogin.vue'
import AdminRegister from '../pages/admin/AdminRegister.vue'
import AdminDashboard from '../pages/admin/AdminDashboard.vue'

// Admin route prefix from .env (default: control-panel)
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

export const routes = [
  {
    path: '/',
    name: 'Home',
    component: HomePage,
  },
  {
    path: `/${adminPrefix}/login`,
    name: 'AdminLogin',
    component: AdminLogin,
  },
  {
    path: `/${adminPrefix}/register`,
    name: 'AdminRegister',
    component: AdminRegister,
  },
  {
    path: `/${adminPrefix}`,
    name: 'AdminDashboard',
    component: AdminDashboard,
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFound,
  },
]
