<template>
  <div class="space-y-8">
    <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div>
        <p class="text-xs uppercase tracking-widest text-slate-400">{{ t('account.title') }}</p>
        <h2 class="text-2xl font-semibold">{{ t('account.heading') }}</h2>
        <p class="text-sm text-slate-400 mt-1">{{ t('account.subtitle') }}</p>
      </div>
      <RouterLink
        to="/projects"
        class="inline-flex items-center justify-center rounded-lg border border-slate-800 px-4 py-2 text-sm text-indigo-200 hover:border-indigo-400/80"
      >
        {{ t('nav.backToProjects') }}
      </RouterLink>
    </div>

    <section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-xl font-semibold">{{ t('account.profile.title') }}</h3>
          <p class="text-sm text-slate-400 mt-1">{{ t('account.profile.subtitle') }}</p>
        </div>
        <span v-if="profileSuccess" class="text-xs text-emerald-400">{{ profileSuccess }}</span>
      </div>

      <div v-if="profileLoading" class="text-sm text-slate-400">{{ t('account.profile.loading') }}</div>
      <div v-else class="grid gap-4 md:grid-cols-2">
        <div class="space-y-1">
          <label class="text-xs text-slate-400">{{ t('account.profile.loginLabel') }}</label>
          <input
            v-model="profileForm.login"
            type="text"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            autocomplete="username"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-slate-400">{{ t('account.profile.emailLabel') }}</label>
          <input
            v-model="profileForm.email"
            type="email"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            autocomplete="email"
          />
        </div>
      </div>

      <div class="flex items-center gap-3">
        <button
          class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
          :disabled="profileSaving"
          data-testid="profile-save"
          @click="handleSaveProfile"
        >
          {{ profileSaving ? t('account.profile.saving') : t('account.profile.save') }}
        </button>
        <p v-if="profileError" class="text-sm text-red-400">{{ profileError }}</p>
      </div>
    </section>

    <section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 space-y-4">
      <div>
        <h3 class="text-xl font-semibold">{{ t('language.title') }}</h3>
        <p class="text-sm text-slate-400 mt-1">{{ t('language.subtitle') }}</p>
      </div>
      <div class="space-y-1">
        <label class="text-xs text-slate-400">{{ t('language.label') }}</label>
        <select
          v-model="selectedLocale"
          class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          @change="handleLocaleChange"
        >
          <option v-for="option in localeOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      </div>
    </section>

    <section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 space-y-6">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">{{ t('account.tokens.title') }}</h3>
          <p class="text-sm text-slate-400 mt-1">
            {{ t('account.tokens.subtitle') }}
          </p>
        </div>
      </div>

      <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-5 space-y-4">
        <h4 class="text-sm font-semibold text-slate-200">{{ t('account.tokens.createTitle') }}</h4>
        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-1">
            <label class="text-xs text-slate-400">{{ t('account.tokens.nameLabel') }}</label>
            <input
              v-model="tokenForm.name"
              type="text"
              :placeholder="t('account.tokens.namePlaceholder')"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
              data-testid="token-name-input"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">{{ t('account.tokens.expiresLabel') }}</label>
            <input
              v-model="tokenForm.expires_at"
              type="datetime-local"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
          </div>
        </div>
        <p class="text-xs text-amber-300/80">
          {{ t('account.tokens.warning') }}
        </p>
        <div class="flex items-center gap-3">
          <button
            class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
            :disabled="tokenCreating"
            data-testid="create-token-button"
            @click="handleCreateToken"
          >
            {{ tokenCreating ? t('account.tokens.createLoading') : t('account.tokens.createButton') }}
          </button>
          <p v-if="tokenCreateError" class="text-sm text-red-400">{{ tokenCreateError }}</p>
        </div>
        <div
          v-if="createdToken"
          class="rounded-lg border border-amber-400/50 bg-amber-500/10 px-4 py-3 text-sm text-amber-100"
        >
          <p class="font-semibold">{{ t('account.tokens.createdTitle') }}</p>
          <p class="text-xs text-amber-200/80">{{ t('account.tokens.createdSubtitle') }}</p>
          <div class="mt-2 rounded bg-slate-950/80 px-3 py-2 font-mono text-xs break-all">
            {{ createdToken.token }}
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <p class="text-sm font-semibold text-slate-200">{{ t('account.tokens.existingTitle') }}</p>
          <p v-if="tokensLoading" class="text-xs text-slate-400">{{ t('account.tokens.loading') }}</p>
        </div>
        <p v-if="tokensError" class="text-sm text-red-400">{{ tokensError }}</p>
        <p v-else-if="tokens.length === 0" class="text-sm text-slate-400">
          {{ t('account.tokens.empty') }}
        </p>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="text-xs uppercase text-slate-500">
              <tr>
                <th class="py-2 pr-4">{{ t('common.name') }}</th>
                <th class="py-2 pr-4">{{ t('common.created') }}</th>
                <th class="py-2 pr-4">{{ t('common.lastUsed') }}</th>
                <th class="py-2 pr-4">{{ t('common.status') }}</th>
                <th class="py-2 text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800">
              <tr v-for="token in tokens" :key="token.id" class="text-slate-200">
                <td class="py-3 pr-4">{{ token.name }}</td>
                <td class="py-3 pr-4 text-slate-400">{{ formatDate(token.created_at) }}</td>
                <td class="py-3 pr-4 text-slate-400">
                  {{ token.last_used_at ? formatDate(token.last_used_at) : t('common.never') }}
                </td>
                <td class="py-3 pr-4">
                  <span
                    v-if="token.revoked_at"
                    class="rounded-full bg-rose-500/20 px-2 py-1 text-xs font-semibold text-rose-200"
                  >
                    {{ t('common.revoked') }}
                  </span>
                  <span
                    v-else
                    class="rounded-full bg-emerald-500/20 px-2 py-1 text-xs font-semibold text-emerald-200"
                  >
                    {{ t('common.active') }}
                  </span>
                </td>
                <td class="py-3 text-right">
                  <button
                    v-if="!token.revoked_at"
                    class="text-xs text-red-300 hover:text-red-200"
                    :disabled="revokingTokenId === token.id"
                    :data-testid="`revoke-token-${token.id}`"
                    @click="handleRevokeToken(token)"
                  >
                    {{ revokingTokenId === token.id ? t('account.tokens.revoking') : t('account.tokens.revoke') }}
                  </button>
                  <span v-else class="text-xs text-slate-500">{{ t('common.empty') }}</span>
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

const handleLocaleChange = () => {
  setLocale(selectedLocale.value)
}

watch(
  locale,
  (value) => {
    selectedLocale.value = value
  },
  { immediate: true },
)

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
