import { createRouter, createWebHistory } from 'vue-router'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import ProjectsView from './views/ProjectsView.vue'
import ProjectDetailView from './views/ProjectDetailView.vue'
import ProjectHistoryView from './views/ProjectHistoryView.vue'
import ProjectTrendsView from './views/ProjectTrendsView.vue'
import AnalysisDetailView from './views/AnalysisDetailView.vue'
import AnalysisCompareView from './views/AnalysisCompareView.vue'
import AccountSettingsView from './views/AccountSettingsView.vue'
import AccountBillingView from './views/AccountBillingView.vue'
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
  {
    path: '/projects/:id/history',
    component: ProjectHistoryView,
    meta: { requiresAuth: true },
  },
  {
    path: '/projects/:id/trends',
    component: ProjectTrendsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/projects/:id/analyses/:analysisId',
    component: AnalysisDetailView,
    meta: { requiresAuth: true },
  },
  {
    path: '/projects/:id/analyses/compare',
    component: AnalysisCompareView,
    meta: { requiresAuth: true },
  },
  {
    path: '/account/settings',
    component: AccountSettingsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/account/billing',
    component: AccountBillingView,
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
