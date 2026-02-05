<template>
  <div class="space-y-10">
    <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-subtle">{{ t('account.title') }}</p>
        <h2 class="text-2xl font-semibold text-ink">{{ t('account.heading') }}</h2>
        <p class="text-sm text-muted mt-1">{{ t('account.subtitle') }}</p>
      </div>
      <RouterLink to="/projects" class="btn btn-secondary text-sm">
        {{ t('nav.backToProjects') }}
      </RouterLink>
    </div>

    <section class="panel p-6 space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-xl font-semibold">{{ t('account.profile.title') }}</h3>
          <p class="text-sm text-muted mt-1">{{ t('account.profile.subtitle') }}</p>
        </div>
        <span v-if="profileSuccess" class="text-xs text-success">{{ profileSuccess }}</span>
      </div>

      <div v-if="profileLoading" class="space-y-3">
        <div class="h-10 rounded skeleton"></div>
        <div class="h-10 rounded skeleton"></div>
      </div>
      <div v-else class="grid gap-4 md:grid-cols-2">
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('account.profile.loginLabel') }}</label>
          <input
            v-model="profileForm.login"
            type="text"
            class="input"
            autocomplete="username"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('account.profile.emailLabel') }}</label>
          <input
            v-model="profileForm.email"
            type="email"
            class="input"
            autocomplete="email"
          />
        </div>
      </div>

      <div class="flex items-center gap-3">
        <button
          class="btn btn-primary"
          :disabled="profileSaving"
          data-testid="profile-save"
          @click="handleSaveProfile"
        >
          {{ profileSaving ? t('account.profile.saving') : t('account.profile.save') }}
        </button>
        <p v-if="profileError" class="text-sm text-danger">{{ profileError }}</p>
      </div>
    </section>

    <section class="panel p-6 space-y-4">
      <div>
        <h3 class="text-xl font-semibold">{{ t('language.title') }}</h3>
        <p class="text-sm text-muted mt-1">{{ t('language.subtitle') }}</p>
      </div>
      <div class="space-y-1">
        <label class="text-xs font-medium text-subtle">{{ t('language.label') }}</label>
        <BaseSelect v-model="selectedLocale" :options="localeOptions" />
      </div>
    </section>

    <section class="panel p-6 space-y-6">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">{{ t('account.tokens.title') }}</h3>
          <p class="text-sm text-muted mt-1">
            {{ t('account.tokens.subtitle') }}
          </p>
        </div>
      </div>

      <div class="surface p-5 space-y-4">
        <h4 class="text-sm font-semibold text-ink">{{ t('account.tokens.createTitle') }}</h4>
        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('account.tokens.nameLabel') }}</label>
            <input
              v-model="tokenForm.name"
              type="text"
              :placeholder="t('account.tokens.namePlaceholder')"
              class="input"
              data-testid="token-name-input"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('account.tokens.expiresLabel') }}</label>
            <BaseDatePicker
              v-model="tokenForm.expires_at"
              :locale="locale"
              :placeholder="t('common.datePlaceholder')"
              :clear-label="t('common.clear')"
              :close-label="t('common.close')"
            />
          </div>
        </div>
        <p class="text-xs text-warning">
          {{ t('account.tokens.warning') }}
        </p>
        <div class="flex items-center gap-3">
          <button
            class="btn btn-primary"
            :disabled="tokenCreating"
            data-testid="create-token-button"
            @click="handleCreateToken"
          >
            {{ tokenCreating ? t('account.tokens.createLoading') : t('account.tokens.createButton') }}
          </button>
          <p v-if="tokenCreateError" class="text-sm text-danger">{{ tokenCreateError }}</p>
        </div>
        <div
          v-if="createdToken"
          class="callout callout-warning"
        >
          <p class="font-semibold">{{ t('account.tokens.createdTitle') }}</p>
          <p class="text-xs text-muted">{{ t('account.tokens.createdSubtitle') }}</p>
          <div class="mt-2 rounded-lg border border-border bg-base/60 px-3 py-2 font-mono text-xs text-ink break-all">
            {{ createdToken.token }}
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <p class="text-sm font-semibold text-ink">{{ t('account.tokens.existingTitle') }}</p>
          <p v-if="tokensLoading" class="text-xs text-muted">{{ t('account.tokens.loading') }}</p>
        </div>
        <p v-if="tokensError" class="text-sm text-danger">{{ tokensError }}</p>
        <p v-else-if="tokens.length === 0" class="text-sm text-muted">
          {{ t('account.tokens.empty') }}
        </p>
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th class="py-2 pr-4">{{ t('common.name') }}</th>
                <th class="py-2 pr-4">{{ t('common.created') }}</th>
                <th class="py-2 pr-4">{{ t('common.lastUsed') }}</th>
                <th class="py-2 pr-4">{{ t('common.status') }}</th>
                <th class="py-2 text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="token in tokens" :key="token.id">
                <td class="py-3 pr-4">{{ token.name }}</td>
                <td class="py-3 pr-4 text-muted">{{ formatDate(token.created_at) }}</td>
                <td class="py-3 pr-4 text-muted">
                  {{ token.last_used_at ? formatDate(token.last_used_at) : t('common.never') }}
                </td>
                <td class="py-3 pr-4">
                  <span
                    v-if="token.revoked_at"
                    class="badge badge-danger"
                  >
                    {{ t('common.revoked') }}
                  </span>
                  <span
                    v-else
                    class="badge badge-success"
                  >
                    {{ t('common.active') }}
                  </span>
                </td>
                <td class="py-3 text-right">
                  <button
                    v-if="!token.revoked_at"
                    class="text-xs text-danger hover:text-danger/80"
                    :disabled="revokingTokenId === token.id"
                    :data-testid="`revoke-token-${token.id}`"
                    @click="handleRevokeToken(token)"
                  >
                    {{ revokingTokenId === token.id ? t('account.tokens.revoking') : t('account.tokens.revoke') }}
                  </button>
                  <span v-else class="text-xs text-subtle">{{ t('common.empty') }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  createApiToken,
  fetchAccount,
  listApiTokens,
  revokeApiToken,
  updateAccount,
} from '../api/client'
import BaseSelect from '../components/BaseSelect.vue'
import BaseDatePicker from '../components/BaseDatePicker.vue'
import { SUPPORTED_LOCALES, setLocale } from '../i18n'
import { useAuth } from '../stores/auth'

const auth = useAuth()
const { locale, t } = useI18n()

const profileForm = reactive({ login: '', email: '' })
const profileLoading = ref(true)
const profileSaving = ref(false)
const profileSuccess = ref('')
const profileError = ref('')

const tokens = ref([])
const tokensLoading = ref(true)
const tokensError = ref('')
const tokenForm = reactive({ name: '', expires_at: '' })
const tokenCreating = ref(false)
const tokenCreateError = ref('')
const createdToken = ref(null)
const revokingTokenId = ref(null)
const selectedLocale = ref(locale.value)
const localeOptions = computed(() => {
  const current = locale.value
  return SUPPORTED_LOCALES.map((value) => ({
    value,
    label: value === 'ru' ? t('language.ru') : t('language.en'),
  }))
})

const formatDate = (value) => {
  if (!value) return ''
  const date = typeof value === 'string' ? new Date(value) : value
  return date.toLocaleString(locale.value)
}

watch(
  locale,
  (value) => {
    selectedLocale.value = value
  },
  { immediate: true },
)

watch(selectedLocale, (value) => {
  setLocale(value)
})

const hydrateProfile = (data) => {
  profileForm.login = data.login || ''
  profileForm.email = data.email || ''
}

const loadProfile = async () => {
  profileLoading.value = true
  profileError.value = ''
  try {
    const data = await fetchAccount()
    hydrateProfile(data)
    if (auth) {
      auth.user = data
    }
  } catch (error) {
    profileError.value = error.message || t('account.profile.loadError')
  } finally {
    profileLoading.value = false
  }
}

const handleSaveProfile = async () => {
  profileSaving.value = true
  profileError.value = ''
  profileSuccess.value = ''
  try {
    const payload = {
      login: profileForm.login,
      email: profileForm.email,
    }
    const updated = await updateAccount(payload)
    hydrateProfile(updated)
    profileSuccess.value = t('account.profile.updated')
    if (auth) {
      auth.user = updated
    }
  } catch (error) {
    profileError.value = error.message || t('account.profile.updateError')
  } finally {
    profileSaving.value = false
  }
}

const loadTokens = async () => {
  tokensLoading.value = true
  tokensError.value = ''
  try {
    tokens.value = await listApiTokens()
  } catch (error) {
    tokensError.value = error.message || t('account.tokens.loadError')
  } finally {
    tokensLoading.value = false
  }
}

const handleCreateToken = async () => {
  tokenCreating.value = true
  tokenCreateError.value = ''
  createdToken.value = null
  if (!tokenForm.name || !tokenForm.name.trim()) {
    tokenCreateError.value = t('account.tokens.nameRequired')
    tokenCreating.value = false
    return
  }
  try {
    const payload = { name: tokenForm.name }
    if (tokenForm.expires_at) {
      payload.expires_at = new Date(tokenForm.expires_at).toISOString()
    }
    const token = await createApiToken(payload)
    createdToken.value = token
    tokens.value = [token, ...tokens.value]
    tokenForm.name = ''
    tokenForm.expires_at = ''
  } catch (error) {
    tokenCreateError.value = error.message || t('account.tokens.createError')
  } finally {
    tokenCreating.value = false
  }
}

const handleRevokeToken = async (token) => {
  revokingTokenId.value = token.id
  try {
    await revokeApiToken(token.id)
    const now = new Date().toISOString()
    tokens.value = tokens.value.map((t) =>
      t.id === token.id ? { ...t, revoked_at: now } : t,
    )
  } catch (error) {
    tokensError.value = error.message || t('account.tokens.revokeError')
  } finally {
    revokingTokenId.value = null
  }
}

onMounted(async () => {
  await Promise.all([loadProfile(), loadTokens()])
})
</script>
