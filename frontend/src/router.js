import { createRouter, createWebHistory } from 'vue-router'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import ProjectsView from './views/ProjectsView.vue'
import ProjectDetailView from './views/ProjectDetailView.vue'
import { hasToken, loadCurrentUser } from './stores/auth'

const routes = [
  { path: '/', redirect: '/projects' },
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  {
    path: '/projects',
    component: ProjectsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/projects/:id',
    component: ProjectDetailView,
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)

  if (requiresAuth && !hasToken()) {
    return { path: '/login' }
  }

  if ((to.path === '/login' || to.path === '/register') && hasToken()) {
    return { path: '/projects' }
  }

  if (requiresAuth && hasToken()) {
    await loadCurrentUser()
  }

  return true
})

export default router
