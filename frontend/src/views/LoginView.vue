<template>
  <div class="mx-auto max-w-md">
    <div class="panel p-8 ds-reveal">
      <h2 class="text-2xl font-semibold">{{ t('auth.login.title') }}</h2>
      <p class="mt-2 text-sm text-muted">{{ t('auth.login.subtitle') }}</p>

      <form class="mt-6 space-y-4" @submit.prevent="handleSubmit">
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('auth.login.identifierLabel') }}</label>
          <input v-model="identifier" type="text" class="input" required />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('auth.login.passwordLabel') }}</label>
          <input v-model="password" type="password" class="input" required />
        </div>

        <p v-if="error" class="text-sm text-danger">{{ error }}</p>

        <button type="submit" class="btn btn-primary w-full" :disabled="loading">
          {{ loading ? t('auth.login.submitLoading') : t('auth.login.submit') }}
        </button>
      </form>

      <p class="mt-6 text-sm text-muted">
        {{ t('auth.login.footerQuestion') }}
        <RouterLink class="link" to="/register">
          {{ t('auth.login.footerAction') }}
        </RouterLink>
      </p>
    </div>
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
