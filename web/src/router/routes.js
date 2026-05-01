import HomePage from '../pages/HomePage.vue'
import LoginPage from '../pages/LoginPage.vue'
import RegisterPage from '../pages/RegisterPage.vue'
import UserDashboard from '../pages/UserDashboard.vue'
import DocsPage from '../pages/DocsPage.vue'
import NotFound from '../pages/NotFound.vue'
import AdminLogin from '../pages/admin/AdminLogin.vue'
import AdminRegister from '../pages/admin/AdminRegister.vue'
import AdminDashboard from '../pages/admin/AdminDashboard.vue'
import AdminSettings from '../pages/admin/AdminSettings.vue'
import AdminUsers from '../pages/admin/AdminUsers.vue'

// Admin route prefix from .env (default: control-panel)
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

export const routes = [
  {
    path: '/',
    name: 'Home',
    component: HomePage,
  },
  {
    path: '/login',
    name: 'Login',
    component: LoginPage,
  },
  {
    path: '/register',
    name: 'Register',
    component: RegisterPage,
  },
  {
    path: '/dashboard',
    name: 'UserDashboard',
    component: UserDashboard,
  },
  {
    path: '/docs/:locale/:page',
    name: 'Docs',
    component: DocsPage,
    props: true,
  },
  {
    path: '/docs/:locale?',
    redirect: to => {
      return { path: `/docs/${to.params.locale || 'en'}/README` }
    },
  },
  {
    path: '/docs',
    redirect: '/docs/en/README',
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
    path: `/${adminPrefix}/settings`,
    name: 'AdminSettings',
    component: AdminSettings,
  },
  {
    path: `/${adminPrefix}/users`,
    name: 'AdminUsers',
    component: AdminUsers,
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFound,
  },
]
