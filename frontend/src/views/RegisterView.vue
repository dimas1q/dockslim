<template>
  <div class="mx-auto max-w-md">
    <div class="panel p-8 ds-reveal">
      <h2 class="text-2xl font-semibold">{{ t('auth.register.title') }}</h2>
      <p class="mt-2 text-sm text-muted">{{ t('auth.register.subtitle') }}</p>

      <form class="mt-6 space-y-4" @submit.prevent="handleSubmit">
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('auth.register.loginLabel') }}</label>
          <input v-model="login" type="text" class="input" minlength="3" required />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('auth.register.emailLabel') }}</label>
          <input v-model="email" type="email" class="input" required />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('auth.register.passwordLabel') }}</label>
          <input v-model="password" type="password" class="input" minlength="8" required />
        </div>

        <p v-if="success" class="text-sm text-success">{{ success }}</p>
        <p v-if="error" class="text-sm text-danger">{{ error }}</p>

        <button type="submit" class="btn btn-primary w-full" :disabled="loading">
          {{ loading ? t('auth.register.submitLoading') : t('auth.register.submit') }}
        </button>
      </form>

      <p class="mt-6 text-sm text-muted">
        {{ t('auth.register.footerQuestion') }}
        <RouterLink class="link" to="/login">
          {{ t('auth.register.footerAction') }}
        </RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { registerUser } from '../api/client'

const router = useRouter()
const { t } = useI18n()
const login = ref('')
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
    await registerUser({ login: login.value, email: email.value, password: password.value })
    success.value = t('auth.register.success')
    setTimeout(() => router.push('/login'), 800)
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>
