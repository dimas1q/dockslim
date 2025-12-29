<template>
  <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-8 shadow-xl">
    <h2 class="text-2xl font-semibold mb-2">Create your account</h2>
    <p class="text-slate-400 mb-6">Start tracking Docker image sizes in minutes.</p>

    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div>
        <label class="text-sm text-slate-300">Email</label>
        <input
          v-model="email"
          type="email"
          class="mt-1 w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          required
        />
      </div>
      <div>
        <label class="text-sm text-slate-300">Password</label>
        <input
          v-model="password"
          type="password"
          class="mt-1 w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          minlength="8"
          required
        />
      </div>

      <p v-if="success" class="text-sm text-emerald-400">{{ success }}</p>
      <p v-if="error" class="text-sm text-red-400">{{ error }}</p>

      <button
        type="submit"
        class="w-full rounded-lg bg-indigo-500 py-2 text-sm font-semibold hover:bg-indigo-400"
        :disabled="loading"
      >
        {{ loading ? 'Creating account...' : 'Create account' }}
      </button>
    </form>

    <p class="text-sm text-slate-400 mt-6">
      Already have an account?
      <RouterLink class="text-indigo-400 hover:text-indigo-300" to="/login">Sign in</RouterLink>
    </p>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { registerUser } from '../api/client'

const router = useRouter()
const email = ref('')
const password = ref('')
const error = ref('')
const success = ref('')
const loading = ref(false)

const handleSubmit = async () => {
  error.value = ''
  success.value = ''
  loading.value = true
  try {
    await registerUser({ email: email.value, password: password.value })
    success.value = 'Account created. Redirecting to login...'
    setTimeout(() => router.push('/login'), 800)
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>
