<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" to="/projects">
      ← Back to projects
    </RouterLink>

    <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-4">
      <p v-if="loading" class="text-sm text-slate-400">Loading project...</p>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>
      <div v-else>
        <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
          <div>
            <h2 class="text-2xl font-semibold">{{ project?.name }}</h2>
            <p class="text-slate-400 mt-2">Project details coming soon.</p>
          </div>
          <button
            v-if="isOwner"
            class="inline-flex items-center justify-center rounded-lg border border-red-500/60 px-4 py-2 text-sm text-red-300 hover:bg-red-500/10"
            :disabled="deleting"
            @click="handleDelete"
          >
            {{ deleting ? 'Deleting...' : 'Delete project' }}
          </button>
        </div>
        <p v-if="deleteError" class="text-sm text-red-400">{{ deleteError }}</p>
      </div>
    </div>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">Registries</h3>
          <p class="text-sm text-slate-400 mt-1">
            Manage container registries connected to this project.
          </p>
        </div>
        <button
          v-if="isOwner"
          class="inline-flex items-center justify-center rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400"
          @click="toggleForm"
        >
          {{ showForm ? 'Close' : 'Add registry' }}
        </button>
      </div>

      <p v-if="!isOwner && project" class="text-xs text-slate-500">
        Only project owners can manage registries.
      </p>

      <div
        v-if="showForm && isOwner"
        class="rounded-xl border border-slate-800 bg-slate-950/60 p-5 space-y-4"
      >
        <h4 class="text-sm font-semibold text-slate-200">Add a registry</h4>
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleCreateRegistry">
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Name</label>
            <input
              v-model="form.name"
              type="text"
              placeholder="Production registry"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
            <p v-if="fieldErrors.name" class="text-xs text-red-400">{{ fieldErrors.name }}</p>
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Registry URL</label>
            <input
              v-model="form.registry_url"
              type="url"
              placeholder="https://registry.example.com"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
            <p v-if="fieldErrors.registry_url" class="text-xs text-red-400">
              {{ fieldErrors.registry_url }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Username (optional)</label>
            <input
              v-model="form.username"
              type="text"
              placeholder="ci-bot"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Password (optional)</label>
            <input
              v-model="form.password"
              type="password"
              placeholder="••••••••"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
          </div>
          <div class="md:col-span-2 flex items-center gap-3">
            <button
              type="submit"
              class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
              :disabled="creatingRegistry"
            >
              {{ creatingRegistry ? 'Saving...' : 'Save registry' }}
            </button>
            <button
              type="button"
              class="text-sm text-slate-400 hover:text-slate-200"
              :disabled="creatingRegistry"
              @click="resetForm"
            >
              Clear
            </button>
          </div>
        </form>
        <p v-if="createRegistryError" class="text-sm text-red-400">{{ createRegistryError }}</p>
      </div>

      <div>
        <p v-if="registriesLoading" class="text-sm text-slate-400">Loading registries...</p>
        <p v-else-if="registriesError" class="text-sm text-red-400">{{ registriesError }}</p>
        <p v-else-if="registries.length === 0" class="text-sm text-slate-400">
          No registries added yet.
        </p>
        <div v-else class="grid gap-4 md:grid-cols-2">
          <div
            v-for="registry in registries"
            :key="registry.id"
            class="rounded-xl border border-slate-800 bg-slate-950/50 p-4 space-y-3"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-base font-semibold">{{ registry.name }}</p>
                <p class="text-xs text-slate-500 mt-1">{{ registry.registry_url }}</p>
              </div>
              <span class="rounded-full bg-slate-800/70 px-2 py-1 text-xs text-slate-200">
                Generic
              </span>
            </div>
            <p v-if="registry.username" class="text-xs text-slate-400">
              Username: {{ registry.username }}
            </p>
            <div v-if="isOwner" class="pt-1">
              <div class="flex items-center gap-3">
                <button
                  class="text-xs text-indigo-300 hover:text-indigo-200"
                  type="button"
                  @click="openEditRegistry(registry)"
                >
                  Edit
                </button>
                <button
                  class="text-xs text-red-300 hover:text-red-200"
                  :disabled="deletingRegistryId === registry.id"
                  @click="handleDeleteRegistry(registry.id)"
                >
                  {{ deletingRegistryId === registry.id ? 'Deleting...' : 'Delete' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="editingRegistry"
        class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 px-4"
      >
        <div class="w-full max-w-lg rounded-2xl border border-slate-800 bg-slate-900 p-6 space-y-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h4 class="text-lg font-semibold text-slate-100">Edit registry</h4>
              <p class="text-xs text-slate-400">Update the name or registry URL.</p>
            </div>
            <button class="text-slate-400 hover:text-slate-200" type="button" @click="closeEdit">
              ✕
            </button>
          </div>
          <form class="space-y-4" @submit.prevent="handleUpdateRegistry">
            <div class="space-y-1">
              <label class="text-xs text-slate-400">Name</label>
              <input
                v-model="editForm.name"
                type="text"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
              />
              <p v-if="editErrors.name" class="text-xs text-red-400">{{ editErrors.name }}</p>
            </div>
            <div class="space-y-1">
              <label class="text-xs text-slate-400">Registry URL</label>
              <input
                v-model="editForm.registry_url"
                type="url"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
              />
              <p v-if="editErrors.registry_url" class="text-xs text-red-400">
                {{ editErrors.registry_url }}
              </p>
            </div>
            <p v-if="editRegistryError" class="text-sm text-red-400">{{ editRegistryError }}</p>
            <div class="flex items-center justify-end gap-3">
              <button
                type="button"
                class="text-sm text-slate-400 hover:text-slate-200"
                :disabled="savingRegistry"
                @click="closeEdit"
              >
                Cancel
              </button>
              <button
                type="submit"
                class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
                :disabled="savingRegistry"
              >
                {{ savingRegistry ? 'Saving...' : 'Save changes' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </section>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <div class="flex items-center gap-3">
            <h3 class="text-xl font-semibold">Analyses</h3>
            <span v-if="polling" class="text-xs text-slate-400">Updating...</span>
          </div>
          <p class="text-sm text-slate-400 mt-1">
            Track image analysis requests and review their status.
          </p>
        </div>
        <button
          v-if="isOwner"
          class="inline-flex items-center justify-center rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400"
          :disabled="registries.length === 0"
          @click="toggleAnalysisForm"
        >
          {{ showAnalysisForm ? 'Close' : 'New analysis' }}
        </button>
      </div>

      <p v-if="!isOwner && project" class="text-xs text-slate-500">
        Only project owners can create new analyses.
      </p>
      <p v-if="isOwner && registries.length === 0" class="text-xs text-slate-500">
        Create a registry first to run image analyses.
      </p>

      <div
        v-if="showAnalysisForm && isOwner"
        class="rounded-xl border border-slate-800 bg-slate-950/60 p-5 space-y-4"
      >
        <h4 class="text-sm font-semibold text-slate-200">Request a new analysis</h4>
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleCreateAnalysis">
          <div class="space-y-1 md:col-span-2">
            <label class="text-xs text-slate-400">Registry</label>
            <select
              v-model="analysisForm.registry_id"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            >
              <option disabled value="">Select a registry</option>
              <option v-for="registry in registries" :key="registry.id" :value="registry.id">
                {{ registry.name }} · {{ registry.registry_url }}
              </option>
            </select>
            <p v-if="analysisErrors.registry_id" class="text-xs text-red-400">
              {{ analysisErrors.registry_id }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Image</label>
            <input
              v-model="analysisForm.image"
              type="text"
              placeholder="repo/name"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
            <p v-if="analysisErrors.image" class="text-xs text-red-400">
              {{ analysisErrors.image }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Tag</label>
            <input
              v-model="analysisForm.tag"
              type="text"
              placeholder="latest"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
          </div>
          <div class="md:col-span-2 flex items-center gap-3">
            <button
              type="submit"
              class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
              :disabled="creatingAnalysis"
            >
              {{ creatingAnalysis ? 'Submitting...' : 'Start analysis' }}
            </button>
            <button
              type="button"
              class="text-sm text-slate-400 hover:text-slate-200"
              :disabled="creatingAnalysis"
              @click="resetAnalysisForm"
            >
              Clear
            </button>
          </div>
        </form>
        <p v-if="createAnalysisError" class="text-sm text-red-400">{{ createAnalysisError }}</p>
      </div>

      <div>
        <p v-if="analysesLoading" class="text-sm text-slate-400">Loading analyses...</p>
        <p v-else-if="analysesError" class="text-sm text-red-400">{{ analysesError }}</p>
        <p v-else-if="analyses.length === 0" class="text-sm text-slate-400">
          No analyses yet. Kick off your first image inspection.
        </p>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="text-xs uppercase text-slate-500">
              <tr>
                <th class="py-2 pr-4">Image</th>
                <th class="py-2 pr-4">Status</th>
                <th class="py-2 pr-4">Created</th>
                <th class="py-2">Total size</th>
                <th class="py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800">
              <tr v-for="analysis in analyses" :key="analysis.id" class="text-slate-200">
                <td class="py-3 pr-4">
                  <RouterLink
                    class="text-indigo-400 hover:text-indigo-300"
                    :to="`/projects/${project?.id}/analyses/${analysis.id}`"
                  >
                    {{ analysis.image }}:{{ analysis.tag }}
                  </RouterLink>
                </td>
                <td class="py-3 pr-4">
                  <span
                    class="rounded-full px-2 py-1 text-xs font-semibold"
                    :class="statusBadgeClass(analysis.status)"
                  >
                    {{ analysis.status }}
                  </span>
                </td>
                <td class="py-3 pr-4 text-slate-400">
                  {{ formatDate(analysis.created_at) }}
                </td>
                <td class="py-3 text-slate-400">
                  {{ analysis.total_size_bytes ? formatBytes(analysis.total_size_bytes) : '—' }}
                </td>
                <td class="py-3 text-right">
                  <div class="flex items-center justify-end gap-3">
                    <RouterLink
                      v-if="getPreviousCompletedAnalysis(analysis)"
                      class="text-xs text-indigo-400 hover:text-indigo-300"
                      :to="`/projects/${project?.id}/analyses/compare?from=${getPreviousCompletedAnalysis(analysis)?.id}&to=${analysis.id}`"
                    >
                      Compare
                    </RouterLink>
                    <button
                      v-if="isOwner"
                      class="text-xs text-red-300 hover:text-red-200 disabled:opacity-60"
                      :disabled="deletingAnalysisId === analysis.id"
                      @click="handleDeleteAnalysis(analysis)"
                    >
                      {{ deletingAnalysisId === analysis.id ? 'Deleting...' : 'Delete' }}
                    </button>
                  </div>
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
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import {
  createAnalysis,
  createRegistry,
  deleteAnalysis,
  deleteProject,
  deleteRegistry,
  getProject,
  listAnalyses,
  listRegistries,
  updateRegistry,
} from '../api/client'

const route = useRoute()
const router = useRouter()
const project = ref(null)
const loading = ref(true)
const error = ref('')
const deleting = ref(false)
const deleteError = ref('')
const registries = ref([])
const registriesLoading = ref(false)
const registriesError = ref('')
const analyses = ref([])
const analysesLoading = ref(false)
const analysesError = ref('')
const polling = ref(false)
const showForm = ref(false)
const creatingRegistry = ref(false)
const createRegistryError = ref('')
const deletingRegistryId = ref(null)
const deletingAnalysisId = ref(null)
const editingRegistry = ref(null)
const editForm = ref({ name: '', registry_url: '' })
const editErrors = ref({})
const savingRegistry = ref(false)
const editRegistryError = ref('')
const fieldErrors = ref({})
const showAnalysisForm = ref(false)
const creatingAnalysis = ref(false)
const createAnalysisError = ref('')
const analysisErrors = ref({})
let pollTimer = null

const form = ref({
  name: '',
  registry_url: '',
  username: '',
  password: '',
})

const analysisForm = ref({
  registry_id: '',
  image: '',
  tag: 'latest',
})

const isOwner = computed(() => project.value?.role === 'owner')
const hasActiveAnalyses = computed(() =>
  analyses.value.some((analysis) => ['queued', 'running'].includes(analysis.status)),
)

const fetchProject = async () => {
  loading.value = true
  error.value = ''
  try {
    project.value = await getProject(route.params.id)
    await fetchRegistries()
    await fetchAnalyses()
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

onMounted(fetchProject)
onBeforeUnmount(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})

const fetchRegistries = async () => {
  registriesLoading.value = true
  registriesError.value = ''
  try {
    registries.value = await listRegistries(route.params.id)
  } catch (err) {
    registriesError.value = err.message
  } finally {
    registriesLoading.value = false
  }
}

const fetchAnalyses = async ({ silent = false } = {}) => {
  if (!silent) {
    analysesLoading.value = true
  }
  analysesError.value = ''
  try {
    analyses.value = await listAnalyses(route.params.id)
  } catch (err) {
    analysesError.value = err.message
  } finally {
    if (!silent) {
      analysesLoading.value = false
    }
  }
}

const startPolling = () => {
  if (pollTimer) {
    return
  }
  polling.value = true
  pollTimer = setInterval(() => {
    fetchAnalyses({ silent: true })
  }, 3000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  polling.value = false
}

watch(
  hasActiveAnalyses,
  (active) => {
    if (active) {
      startPolling()
      return
    }
    stopPolling()
  },
  { immediate: true },
)

const toggleForm = () => {
  showForm.value = !showForm.value
  if (!showForm.value) {
    resetForm()
  }
}

const resetForm = () => {
  form.value = {
    name: '',
    registry_url: '',
    username: '',
    password: '',
  }
  fieldErrors.value = {}
  createRegistryError.value = ''
}

const toggleAnalysisForm = () => {
  showAnalysisForm.value = !showAnalysisForm.value
  if (!showAnalysisForm.value) {
    resetAnalysisForm()
  }
}

const resetAnalysisForm = () => {
  analysisForm.value = {
    registry_id: '',
    image: '',
    tag: 'latest',
  }
  analysisErrors.value = {}
  createAnalysisError.value = ''
}

const handleCreateRegistry = async () => {
  fieldErrors.value = {}
  createRegistryError.value = ''

  if (!form.value.name.trim()) {
    fieldErrors.value.name = 'Name is required.'
  }
  if (!form.value.registry_url.trim()) {
    fieldErrors.value.registry_url = 'Registry URL is required.'
  }

  if (Object.keys(fieldErrors.value).length > 0) {
    return
  }

  creatingRegistry.value = true
  try {
    await createRegistry(route.params.id, {
      name: form.value.name,
      type: 'generic',
      registry_url: form.value.registry_url,
      username: form.value.username,
      password: form.value.password,
    })
    form.value.password = ''
    showForm.value = false
    resetForm()
    await fetchRegistries()
  } catch (err) {
    createRegistryError.value = err.message
  } finally {
    creatingRegistry.value = false
  }
}

const handleCreateAnalysis = async () => {
  analysisErrors.value = {}
  createAnalysisError.value = ''

  if (!analysisForm.value.registry_id) {
    analysisErrors.value.registry_id = 'Registry is required.'
  }
  if (!analysisForm.value.image.trim()) {
    analysisErrors.value.image = 'Image is required.'
  }

  if (Object.keys(analysisErrors.value).length > 0) {
    return
  }

  creatingAnalysis.value = true
  try {
    await createAnalysis(route.params.id, {
      registry_id: analysisForm.value.registry_id,
      image: analysisForm.value.image,
      tag: analysisForm.value.tag,
    })
    showAnalysisForm.value = false
    resetAnalysisForm()
    await fetchAnalyses()
  } catch (err) {
    createAnalysisError.value = err.message
  } finally {
    creatingAnalysis.value = false
  }
}

const handleDeleteRegistry = async (registryId) => {
  deletingRegistryId.value = registryId
  try {
    await deleteRegistry(route.params.id, registryId)
    await fetchRegistries()
  } catch (err) {
    registriesError.value = err.message
  } finally {
    deletingRegistryId.value = null
  }
}

const handleDeleteAnalysis = async (analysis) => {
  if (!analysis?.id) {
    return
  }
  const confirmed = window.confirm('Delete this analysis? This cannot be undone.')
  if (!confirmed) {
    return
  }

  deletingAnalysisId.value = analysis.id
  try {
    await deleteAnalysis(route.params.id, analysis.id)
    await fetchAnalyses()
  } catch (err) {
    analysesError.value = err.message
  } finally {
    deletingAnalysisId.value = null
  }
}

const formatDate = (value) => {
  if (!value) return 'just now'
  return new Date(value).toLocaleString()
}

const formatBytes = (value) => {
  if (!value && value !== 0) return '—'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = Number(value)
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

const statusBadgeClass = (status) => {
  switch (status) {
    case 'completed':
      return 'bg-emerald-500/20 text-emerald-200'
    case 'running':
      return 'bg-sky-500/20 text-sky-200'
    case 'failed':
      return 'bg-rose-500/20 text-rose-200'
    default:
      return 'bg-amber-500/20 text-amber-200'
  }
}

const getPreviousCompletedAnalysis = (analysis) => {
  if (!analysis || analysis.status !== 'completed') {
    return null
  }
  const currentCreatedAt = new Date(analysis.created_at).getTime()
  if (!Number.isFinite(currentCreatedAt)) {
    return null
  }
  return analyses.value.reduce((latest, item) => {
    if (item.id === analysis.id) {
      return latest
    }
    if (item.image !== analysis.image || item.status !== 'completed') {
      return latest
    }
    const itemCreatedAt = new Date(item.created_at).getTime()
    if (!Number.isFinite(itemCreatedAt) || itemCreatedAt >= currentCreatedAt) {
      return latest
    }
    if (!latest) {
      return item
    }
    const latestCreatedAt = new Date(latest.created_at).getTime()
    if (!Number.isFinite(latestCreatedAt) || itemCreatedAt > latestCreatedAt) {
      return item
    }
    return latest
  }, null)
}

const openEditRegistry = (registry) => {
  editingRegistry.value = registry
  editForm.value = {
    name: registry.name || '',
    registry_url: registry.registry_url || '',
  }
  editErrors.value = {}
  editRegistryError.value = ''
}

const closeEdit = () => {
  editingRegistry.value = null
  editForm.value = { name: '', registry_url: '' }
  editErrors.value = {}
  editRegistryError.value = ''
}

const handleUpdateRegistry = async () => {
  if (!editingRegistry.value) {
    return
  }

  editErrors.value = {}
  editRegistryError.value = ''

  const nameValue = editForm.value.name.trim()
  const urlValue = editForm.value.registry_url.trim()
  const hasNameChange = nameValue !== editingRegistry.value.name
  const hasURLChange = urlValue !== editingRegistry.value.registry_url

  if (hasNameChange && !nameValue) {
    editErrors.value.name = 'Name is required.'
  }
  if (hasURLChange && !urlValue) {
    editErrors.value.registry_url = 'Registry URL is required.'
  }
  if (!hasNameChange && !hasURLChange) {
    editRegistryError.value = 'Make a change before saving.'
    return
  }
  if (Object.keys(editErrors.value).length > 0) {
    return
  }

  const payload = {}
  if (hasNameChange) {
    payload.name = nameValue
  }
  if (hasURLChange) {
    payload.registry_url = urlValue
  }

  savingRegistry.value = true
  try {
    await updateRegistry(route.params.id, editingRegistry.value.id, payload)
    closeEdit()
    await fetchRegistries()
  } catch (err) {
    editRegistryError.value = err.message
  } finally {
    savingRegistry.value = false
  }
}

const handleDelete = async () => {
  if (!project.value) {
    return
  }
  deleteError.value = ''
  const confirmed = window.confirm('Delete this project? This cannot be undone.')
  if (!confirmed) {
    return
  }

  deleting.value = true
  try {
    await deleteProject(project.value.id)
    router.push({ path: '/projects', query: { deleted: '1' } })
  } catch (err) {
    deleteError.value = err.message
  } finally {
    deleting.value = false
  }
}
</script>
