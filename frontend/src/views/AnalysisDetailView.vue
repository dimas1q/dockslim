<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" :to="`/projects/${projectId}`">
      ← Back to project
    </RouterLink>

    <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-4">
      <p v-if="loading" class="text-sm text-slate-400">Loading analysis...</p>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>
      <div v-else>
        <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
          <div>
            <h2 class="text-2xl font-semibold">{{ analysis?.image }}:{{ analysis?.tag }}</h2>
            <p class="text-sm text-slate-400 mt-1">
              Created {{ formatDate(analysis?.created_at) }}
            </p>
          </div>
          <div class="flex items-center gap-3">
            <button
              v-if="isOwner"
              class="rounded-lg border border-slate-700 px-3 py-1 text-xs text-slate-200 hover:border-slate-500 disabled:opacity-60"
              :disabled="rerunning"
              @click="handleRerun"
            >
              {{ rerunning ? 'Re-running...' : 'Re-run analysis' }}
            </button>
            <span class="rounded-full px-3 py-1 text-xs font-semibold" :class="statusBadgeClass">
              {{ analysis?.status }}
            </span>
          </div>
        </div>
        <p v-if="rerunError" class="text-sm text-rose-400">{{ rerunError }}</p>

        <div class="grid gap-4 md:grid-cols-3 text-sm text-slate-300 mt-4">
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Created</p>
            <p class="mt-1">{{ formatDate(analysis?.created_at) }}</p>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Updated</p>
            <p class="mt-1">{{ formatDate(analysis?.updated_at) }}</p>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Total size</p>
            <p class="mt-1">
              {{ analysis?.total_size_bytes ? formatBytes(analysis.total_size_bytes) : '—' }}
            </p>
          </div>
        </div>

        <div class="mt-6 rounded-xl border border-slate-800 bg-slate-950/50 p-6 space-y-3">
          <div class="flex items-center justify-between">
            <p class="text-sm font-semibold">Result</p>
            <span v-if="polling" class="text-xs text-slate-400">Updating...</span>
          </div>
          <p v-if="failedMessage" class="text-sm text-rose-300">
            {{ failedMessage }}
          </p>
          <p v-else-if="!analysis?.result_json" class="text-sm text-slate-400">
            Layer breakdown coming soon.
          </p>
          <pre
            v-else
            class="whitespace-pre-wrap break-words rounded-lg bg-slate-950/70 p-4 text-xs text-slate-200"
          >
{{ formattedResult }}
          </pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { getAnalysis, getProject, rerunAnalysis } from '../api/client'

const route = useRoute()
const projectId = route.params.id
const analysisId = route.params.analysisId

const analysis = ref(null)
const loading = ref(true)
const error = ref('')
const project = ref(null)
const polling = ref(false)
const rerunning = ref(false)
const rerunError = ref('')
let pollTimer = null

const fetchAnalysis = async ({ silent = false } = {}) => {
  if (!silent) {
    loading.value = true
  }
  error.value = ''
  try {
    analysis.value = await getAnalysis(projectId, analysisId)
  } catch (err) {
    error.value = err.message
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

const fetchProject = async () => {
  try {
    project.value = await getProject(projectId)
  } catch (err) {
    // ignore role fetch errors
  }
}

onMounted(() => {
  fetchProject()
  fetchAnalysis()
})
onBeforeUnmount(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})

const isActive = computed(() => ['queued', 'running'].includes(analysis.value?.status))
const isOwner = computed(() => project.value?.role === 'owner')

const startPolling = () => {
  if (pollTimer) {
    return
  }
  polling.value = true
  pollTimer = setInterval(() => {
    fetchAnalysis({ silent: true })
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
  isActive,
  (active) => {
    if (active) {
      startPolling()
      return
    }
    stopPolling()
  },
  { immediate: true },
)

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

const statusBadgeClass = computed(() => {
  switch (analysis.value?.status) {
    case 'completed':
      return 'bg-emerald-500/20 text-emerald-200'
    case 'running':
      return 'bg-sky-500/20 text-sky-200'
    case 'failed':
      return 'bg-rose-500/20 text-rose-200'
    default:
      return 'bg-amber-500/20 text-amber-200'
  }
})

const failedMessage = computed(() => {
  if (analysis.value?.status !== 'failed') {
    return ''
  }
  if (analysis.value?.result_json?.error) {
    return analysis.value.result_json.error
  }
  return 'Analysis failed.'
})

const formattedResult = computed(() => {
  if (!analysis.value?.result_json) {
    return ''
  }
  try {
    return JSON.stringify(analysis.value.result_json, null, 2)
  } catch (err) {
    return String(analysis.value.result_json)
  }
})

const handleRerun = async () => {
  const confirmed = window.confirm('Re-run this analysis? It will overwrite the current result.')
  if (!confirmed) {
    return
  }

  rerunError.value = ''
  rerunning.value = true
  try {
    await rerunAnalysis(projectId, analysisId)
    await fetchAnalysis()
  } catch (err) {
    if (err.status === 409) {
      rerunError.value = 'Analysis is already running.'
    } else {
      rerunError.value = err.message
    }
  } finally {
    rerunning.value = false
  }
}
</script>
