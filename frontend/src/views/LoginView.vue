<template>
  <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-8 shadow-xl">
    <h2 class="text-2xl font-semibold mb-2">{{ t('auth.login.title') }}</h2>
    <p class="text-slate-400 mb-6">{{ t('auth.login.subtitle') }}</p>

    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div>
        <label class="text-sm text-slate-300">{{ t('auth.login.identifierLabel') }}</label>
        <input
          v-model="identifier"
          type="text"
          class="mt-1 w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          required
        />
      </div>
      <div>
        <label class="text-sm text-slate-300">{{ t('auth.login.passwordLabel') }}</label>
        <input
          v-model="password"
          type="password"
          class="mt-1 w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          required
        />
      </div>

      <p v-if="error" class="text-sm text-red-400">{{ error }}</p>

      <button
        type="submit"
        class="w-full rounded-lg bg-indigo-500 py-2 text-sm font-semibold hover:bg-indigo-400"
        :disabled="loading"
      >
        {{ loading ? t('auth.login.submitLoading') : t('auth.login.submit') }}
      </button>
    </form>

    <p class="text-sm text-slate-400 mt-6">
      {{ t('auth.login.footerQuestion') }}
      <RouterLink class="text-indigo-400 hover:text-indigo-300" to="/register">
        {{ t('auth.login.footerAction') }}
      </RouterLink>
    </p>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { loginUser } from '../api/client'
import { loadCurrentUser } from '../stores/auth'

const router = useRouter()
const { t } = useI18n()
const identifier = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const handleSubmit = async () => {
  error.value = ''
  loading.value = true
  try {
    await loginUser({ identifier: identifier.value, password: password.value })
    await loadCurrentUser()
    router.push('/projects')
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>
