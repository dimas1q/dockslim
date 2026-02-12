<template>
  <div class="app-shell">
    <div class="page">
      <header class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
        <div class="flex items-center gap-4">
          <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/15 text-primary">
            <svg class="h-6 w-6" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <path :d="logoIcon" />
            </svg>
          </div>
          <div>
            <p class="text-xs uppercase tracking-[0.35em] text-subtle">{{ t('app.brand') }}</p>
            <h1 class="text-2xl font-semibold text-ink">{{ t('app.tagline') }}</h1>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <div class="theme-switch" role="group" aria-label="Theme">
            <span class="theme-switch__indicator" :class="theme === 'dark' ? 'is-dark' : ''"></span>
            <button
              type="button"
              :data-active="theme === 'light'"
              :aria-pressed="theme === 'light'"
              @click="applyTheme('light')"
            >
              {{ t('theme.light') }}
            </button>
            <button
              type="button"
              :data-active="theme === 'dark'"
              :aria-pressed="theme === 'dark'"
              @click="applyTheme('dark')"
            >
              {{ t('theme.dark') }}
            </button>
          </div>
          <div v-if="auth.user" class="flex items-center gap-2">
            <RouterLink
              to="/account/settings"
              class="header-pill"
              :class="isAccountActive ? 'header-pill-active' : ''"
            >
              {{ t('app.accountFallback') }}
            </RouterLink>
            <button class="header-pill" @click="handleLogout">
              {{ t('app.logout') }}
            </button>
          </div>
        </div>
      </header>

      <main class="mt-10">
        <Transition name="route-fade" mode="out-in">
          <RouterView />
        </Transition>
      </main>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { computed } from 'vue'
import { loadCurrentUser, logout, useAuth } from './stores/auth'
import { applyTheme, theme } from './stores/theme'
import { mdiDocker } from '@mdi/js'

const auth = useAuth()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const isAccountActive = computed(() => route.path.startsWith('/account/'))
const logoIcon = mdiDocker

const handleLogout = async () => {
  await logout()
  router.push('/login')
}

onMounted(() => {
  loadCurrentUser()
})
</script>
