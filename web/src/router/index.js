import { createRouter, createWebHistory } from 'vue-router'
import { routes } from './routes'

export function createSSRRouter(history) {
  return createRouter({
    history,
    routes,
  })
}

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
