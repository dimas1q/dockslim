<template>
  <div class="min-h-screen bg-slate-950 text-slate-100">
    <div class="max-w-4xl mx-auto px-6 py-10">
      <header class="flex items-center justify-between mb-10">
        <div>
          <p class="text-xs uppercase tracking-widest text-slate-400">{{ t('app.brand') }}</p>
          <h1 class="text-2xl font-semibold">{{ t('app.tagline') }}</h1>
        </div>
        <div v-if="auth.user" class="flex items-center gap-3">
          <RouterLink
            to="/account/settings"
            class="rounded-lg border border-slate-800 px-3 py-1.5 text-sm text-indigo-200 hover:border-indigo-400/80"
          >
            {{ auth.user.login || t('app.accountFallback') }}
          </RouterLink>
          <button class="text-sm text-slate-300 hover:text-white" @click="handleLogout">
            {{ t('app.logout') }}
          </button>
        </div>
      </header>
      <RouterView />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { loadCurrentUser, logout, useAuth } from './stores/auth'

const auth = useAuth()
const router = useRouter()
const { t } = useI18n()

const handleLogout = async () => {
  await logout()
  router.push('/login')
}

onMounted(() => {
  loadCurrentUser()
})
</script>
