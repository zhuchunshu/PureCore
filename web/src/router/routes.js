import HomePage from '../pages/HomePage.vue'
import NotFound from '../pages/NotFound.vue'

export const routes = [
  {
    path: '/',
    name: 'Home',
    component: HomePage,
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFound,
  },
]
