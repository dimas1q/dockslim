<template>
  <div class="min-h-screen bg-slate-950 text-slate-100">
    <div class="max-w-4xl mx-auto px-6 py-10">
      <header class="flex items-center justify-between mb-10">
        <div>
          <p class="text-xs uppercase tracking-widest text-slate-400">DockSlim</p>
          <h1 class="text-2xl font-semibold">Container insights</h1>
        </div>
        <button
          v-if="auth.user"
          class="text-sm text-slate-300 hover:text-white"
          @click="handleLogout"
        >
          Logout
        </button>
      </header>
      <RouterView />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import { loadCurrentUser, logout, useAuth } from './stores/auth'

const auth = useAuth()
const router = useRouter()

const handleLogout = async () => {
  await logout()
  router.push('/login')
}

onMounted(() => {
  loadCurrentUser()
})
</script>
