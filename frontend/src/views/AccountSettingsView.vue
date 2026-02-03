<template>
  <div class="space-y-8">
    <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div>
        <p class="text-xs uppercase tracking-widest text-slate-400">Account</p>
        <h2 class="text-2xl font-semibold">Account settings</h2>
        <p class="text-sm text-slate-400 mt-1">Manage your profile and personal API tokens.</p>
      </div>
      <RouterLink
        to="/projects"
        class="inline-flex items-center justify-center rounded-lg border border-slate-800 px-4 py-2 text-sm text-indigo-200 hover:border-indigo-400/80"
      >
        ← Back to projects
      </RouterLink>
    </div>

    <section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-xl font-semibold">Profile</h3>
          <p class="text-sm text-slate-400 mt-1">Update the contact details associated with your account.</p>
        </div>
        <span v-if="profileSuccess" class="text-xs text-emerald-400">{{ profileSuccess }}</span>
      </div>

      <div v-if="profileLoading" class="text-sm text-slate-400">Loading profile…</div>
      <div v-else class="grid gap-4 md:grid-cols-2">
        <div class="space-y-1">
          <label class="text-xs text-slate-400">Login</label>
          <input
            v-model="profileForm.login"
            type="text"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            autocomplete="username"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-slate-400">Email</label>
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
          {{ profileSaving ? 'Saving...' : 'Save profile' }}
        </button>
        <p v-if="profileError" class="text-sm text-red-400">{{ profileError }}</p>
      </div>
    </section>

    <section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 space-y-6">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">Personal API tokens</h3>
          <p class="text-sm text-slate-400 mt-1">
            Authenticate API calls with a bearer token instead of browser cookies.
          </p>
        </div>
      </div>

      <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-5 space-y-4">
        <h4 class="text-sm font-semibold text-slate-200">Create token</h4>
        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Name</label>
            <input
              v-model="tokenForm.name"
              type="text"
              placeholder="cli-token"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
              data-testid="token-name-input"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Expires at (optional)</label>
            <input
              v-model="tokenForm.expires_at"
              type="datetime-local"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
          </div>
        </div>
        <p class="text-xs text-amber-300/80">
          Token value is shown once. Copy it now and store it securely.
        </p>
        <div class="flex items-center gap-3">
          <button
            class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
            :disabled="tokenCreating"
            data-testid="create-token-button"
            @click="handleCreateToken"
          >
            {{ tokenCreating ? 'Creating...' : 'Create token' }}
          </button>
          <p v-if="tokenCreateError" class="text-sm text-red-400">{{ tokenCreateError }}</p>
        </div>
        <div
          v-if="createdToken"
          class="rounded-lg border border-amber-400/50 bg-amber-500/10 px-4 py-3 text-sm text-amber-100"
        >
          <p class="font-semibold">Token created</p>
          <p class="text-xs text-amber-200/80">Copy now. You will not be able to see it again.</p>
          <div class="mt-2 rounded bg-slate-950/80 px-3 py-2 font-mono text-xs break-all">
            {{ createdToken.token }}
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <p class="text-sm font-semibold text-slate-200">Existing tokens</p>
          <p v-if="tokensLoading" class="text-xs text-slate-400">Loading...</p>
        </div>
        <p v-if="tokensError" class="text-sm text-red-400">{{ tokensError }}</p>
        <p v-else-if="tokens.length === 0" class="text-sm text-slate-400">No tokens yet.</p>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="text-xs uppercase text-slate-500">
              <tr>
                <th class="py-2 pr-4">Name</th>
                <th class="py-2 pr-4">Created</th>
                <th class="py-2 pr-4">Last used</th>
                <th class="py-2 pr-4">Status</th>
                <th class="py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800">
              <tr v-for="token in tokens" :key="token.id" class="text-slate-200">
                <td class="py-3 pr-4">{{ token.name }}</td>
                <td class="py-3 pr-4 text-slate-400">{{ formatDate(token.created_at) }}</td>
                <td class="py-3 pr-4 text-slate-400">
                  {{ token.last_used_at ? formatDate(token.last_used_at) : 'Never' }}
                </td>
                <td class="py-3 pr-4">
                  <span
                    v-if="token.revoked_at"
                    class="rounded-full bg-rose-500/20 px-2 py-1 text-xs font-semibold text-rose-200"
                  >
                    Revoked
                  </span>
                  <span
                    v-else
                    class="rounded-full bg-emerald-500/20 px-2 py-1 text-xs font-semibold text-emerald-200"
                  >
                    Active
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
                    {{ revokingTokenId === token.id ? 'Revoking...' : 'Revoke' }}
                  </button>
                  <span v-else class="text-xs text-slate-500">—</span>
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
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  createApiToken,
  fetchAccount,
  listApiTokens,
  revokeApiToken,
  updateAccount,
} from '../api/client'
import { useAuth } from '../stores/auth'

const auth = useAuth()

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

const formatDate = (value) => {
  if (!value) return ''
  const date = typeof value === 'string' ? new Date(value) : value
  return date.toLocaleString()
}

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
    profileError.value = error.message || 'Failed to load profile'
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
    profileSuccess.value = 'Profile updated'
    if (auth) {
      auth.user = updated
    }
  } catch (error) {
    profileError.value = error.message || 'Failed to update profile'
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
    tokensError.value = error.message || 'Failed to load tokens'
  } finally {
    tokensLoading.value = false
  }
}

const handleCreateToken = async () => {
  tokenCreating.value = true
  tokenCreateError.value = ''
  createdToken.value = null
  if (!tokenForm.name || !tokenForm.name.trim()) {
    tokenCreateError.value = 'Name is required'
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
    tokenCreateError.value = error.message || 'Failed to create token'
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
    tokensError.value = error.message || 'Failed to revoke token'
  } finally {
    revokingTokenId.value = null
  }
}

onMounted(async () => {
  await Promise.all([loadProfile(), loadTokens()])
})
</script>
