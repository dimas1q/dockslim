import { createRouter, createWebHistory } from 'vue-router'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import ProjectsView from './views/ProjectsView.vue'
import ProjectDetailView from './views/ProjectDetailView.vue'
import { loadCurrentUser, useAuth } from './stores/auth'

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
  const auth = useAuth()

  if (!auth.initialized) {
    await loadCurrentUser()
  }

  if (requiresAuth && !auth.user) {
    return { path: '/login' }
  }

  if ((to.path === '/login' || to.path === '/register') && auth.user) {
    return { path: '/projects' }
  }

  return true
})

export default router
