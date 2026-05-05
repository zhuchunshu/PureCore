import HomePage from '../pages/HomePage.vue'
import LoginPage from '../pages/LoginPage.vue'
import RegisterPage from '../pages/RegisterPage.vue'
import UserDashboard from '../pages/UserDashboard.vue'
import OverviewPage from '../pages/dashboard/OverviewPage.vue'
import ProfilePage from '../pages/dashboard/ProfilePage.vue'
import SecurityPage from '../pages/dashboard/SecurityPage.vue'
import ApiKeysPage from '../pages/dashboard/ApiKeysPage.vue'
import SessionsPage from '../pages/dashboard/SessionsPage.vue'
import IntegrationsPage from '../pages/dashboard/IntegrationsPage.vue'
import DocsPage from '../pages/DocsPage.vue'
import NotFound from '../pages/NotFound.vue'
import AdminLogin from '../pages/admin/AdminLogin.vue'
import AdminRegister from '../pages/admin/AdminRegister.vue'
import AdminDashboard from '../pages/admin/AdminDashboard.vue'
import AdminSettings from '../pages/admin/AdminSettings.vue'
import AdminUsers from '../pages/admin/AdminUsers.vue'
import OAuthCallback from '../pages/auth/OAuthCallback.vue'
import AdminOAuthSettings from '../pages/admin/AdminOAuthSettings.vue'

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
    component: UserDashboard,
    meta: { requiresAuth: true, authType: 'user' },
    children: [
      {
        path: '',
        name: 'DashboardOverview',
        component: OverviewPage,
      },
      {
        path: 'profile',
        name: 'DashboardProfile',
        component: ProfilePage,
      },
      {
        path: 'security',
        name: 'DashboardSecurity',
        component: SecurityPage,
      },
      {
        path: 'api-keys',
        name: 'DashboardApiKeys',
        component: ApiKeysPage,
      },
      {
        path: 'sessions',
        name: 'DashboardSessions',
        component: SessionsPage,
      },
      {
        path: 'integrations',
        name: 'DashboardIntegrations',
        component: IntegrationsPage,
      },
    ],
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
    meta: { requiresAuth: true, authType: 'admin' },
  },
  {
    path: `/${adminPrefix}/settings`,
    name: 'AdminSettings',
    component: AdminSettings,
    meta: { requiresAuth: true, authType: 'admin' },
  },
  {
    path: `/${adminPrefix}/users`,
    name: 'AdminUsers',
    component: AdminUsers,
    meta: { requiresAuth: true, authType: 'admin' },
  },
  {
    path: `/${adminPrefix}/oauth`,
    name: 'AdminOAuthSettings',
    component: AdminOAuthSettings,
    meta: { requiresAuth: true, authType: 'admin' },
  },
  {
    path: '/oauth/:provider/callback',
    name: 'OAuthCallback',
    component: OAuthCallback,
    props: true,
  },
  {
    path: '/oauth/:provider/link/register',
    name: 'OAuthLinkRegister',
    component: () => import('../pages/auth/OAuthLinkRegister.vue'),
    props: true,
  },
  {
    path: '/oauth/:provider/link/login',
    name: 'OAuthLinkLogin',
    component: () => import('../pages/auth/OAuthLinkLogin.vue'),
    props: true,
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFound,
  },
]
