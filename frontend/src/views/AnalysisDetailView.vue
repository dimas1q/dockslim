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
              {{ totalSizeDisplay }}
            </p>
          </div>
        </div>

        <div v-if="failedMessage" class="mt-6 rounded-xl border border-rose-500/40 bg-rose-950/40 p-6">
          <p class="text-sm font-semibold text-rose-200">Analysis failed</p>
          <p class="mt-2 text-sm text-rose-300">{{ failedMessage }}</p>
        </div>
        <div v-else-if="!analysis?.result_json" class="mt-6 rounded-xl border border-slate-800 bg-slate-950/50 p-6">
          <p class="text-sm text-slate-400">Layer breakdown coming soon.</p>
        </div>
        <div v-else class="mt-6 space-y-6">
          <div class="flex items-center justify-between text-xs text-slate-400">
            <span>Layer breakdown</span>
            <span v-if="polling">Updating...</span>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-6 space-y-3">
              <p class="text-sm font-semibold text-slate-200">Summary</p>
              <div class="text-sm text-slate-300 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-slate-500">Image</span>
                  <span>{{ analysis?.image }}:{{ analysis?.tag }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-500">Total size</span>
                  <span>{{ totalSizeDisplay }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-500">Layer count</span>
                  <span>{{ layerCountDisplay }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-500">Manifest type</span>
                  <span>{{ manifestTypeLabel }}</span>
                </div>
              </div>
            </div>

            <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-6 space-y-3">
              <p class="text-sm font-semibold text-slate-200">Insights</p>
              <div class="space-y-3 text-sm text-slate-300">
                <div>
                  <p class="text-xs uppercase text-slate-500 tracking-wide">Warnings</p>
                  <ul v-if="warnings.length" class="mt-2 space-y-2">
                    <li
                      v-for="warning in warnings"
                      :key="warning"
                      class="rounded-lg border border-rose-500/40 bg-rose-950/40 px-3 py-2 text-rose-200"
                    >
                      {{ warning }}
                    </li>
                  </ul>
                  <p v-else class="mt-2 text-slate-400">No warnings detected.</p>
                </div>
                <div>
                  <p class="text-xs uppercase text-slate-500 tracking-wide">Largest layers</p>
                  <ul v-if="largestLayers.length" class="mt-2 space-y-2">
                    <li
                      v-for="layer in largestLayers"
                      :key="layer.digest"
                      class="flex items-center justify-between rounded-lg border border-slate-800 bg-slate-900/60 px-3 py-2"
                    >
                      <span class="text-slate-300">{{ shortDigest(layer.digest) }}</span>
                      <span class="text-slate-200">{{ formatBytes(layer.size_bytes) }}</span>
                    </li>
                  </ul>
                  <p v-else class="mt-2 text-slate-400">No layers reported.</p>
                </div>
              </div>
            </div>
          </div>

          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-6 space-y-4">
            <p class="text-sm font-semibold text-slate-200">Optimization Recommendations</p>
            <div v-if="recommendations.length" class="grid gap-3 md:grid-cols-2">
              <div
                v-for="recommendation in recommendations"
                :key="recommendation.id"
                class="rounded-xl border p-4"
                :class="severityStyles(recommendation.severity).container"
              >
                <div class="flex items-center gap-2">
                  <span
                    class="h-2.5 w-2.5 rounded-full"
                    :class="severityStyles(recommendation.severity).icon"
                  ></span>
                  <p class="text-sm font-semibold">{{ recommendation.title }}</p>
                </div>
                <p class="mt-2 text-sm text-slate-200">{{ recommendation.description }}</p>
                <p class="mt-2 text-xs text-slate-400">{{ recommendation.suggested_action }}</p>
              </div>
            </div>
            <p v-else class="text-sm text-slate-400">No optimization issues detected 🎉</p>
          </div>

          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-6 space-y-4">
            <div class="flex items-center justify-between">
              <p class="text-sm font-semibold text-slate-200">Layers</p>
              <button
                class="text-xs text-indigo-400 hover:text-indigo-300"
                type="button"
                @click="showRaw = !showRaw"
              >
                {{ showRaw ? 'Hide raw JSON' : 'Show raw JSON' }}
              </button>
            </div>
            <div class="max-h-80 overflow-y-auto rounded-lg border border-slate-800">
              <table class="min-w-full text-left text-sm text-slate-300">
                <thead class="bg-slate-900/70 text-xs uppercase text-slate-500">
                  <tr>
                    <th class="px-4 py-3">Digest</th>
                    <th class="px-4 py-3">Size</th>
                    <th class="px-4 py-3">Media type</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="layer in layers" :key="layer.digest" class="border-t border-slate-800">
                    <td class="px-4 py-3 font-mono text-xs text-slate-200">
                      {{ shortDigest(layer.digest) }}
                    </td>
                    <td class="px-4 py-3">{{ formatBytes(layer.size_bytes) }}</td>
                    <td class="px-4 py-3 text-xs text-slate-400">{{ layer.media_type || '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <pre
              v-if="showRaw"
              class="whitespace-pre-wrap break-words rounded-lg bg-slate-950/70 p-4 text-xs text-slate-200"
            >
{{ formattedResult }}
            </pre>
          </div>
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
const showRaw = ref(false)
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

const shortDigest = (value) => {
  if (!value) return '—'
  const trimmed = value.trim()
  if (trimmed.includes(':')) {
    const [algo, hash] = trimmed.split(':')
    if (hash) {
      return `${algo}:${hash.slice(0, 12)}`
    }
  }
  return trimmed.slice(0, 12)
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

const result = computed(() => analysis.value?.result_json ?? null)

const failedMessage = computed(() => {
  if (analysis.value?.status !== 'failed') {
    return ''
  }
  if (analysis.value?.result_json?.error) {
    return analysis.value.result_json.error
  }
  return 'Analysis failed.'
})

const totalSizeDisplay = computed(() => {
  if (analysis.value?.total_size_bytes || analysis.value?.total_size_bytes === 0) {
    return formatBytes(analysis.value.total_size_bytes)
  }
  if (result.value?.total_size_bytes || result.value?.total_size_bytes === 0) {
    return formatBytes(result.value.total_size_bytes)
  }
  return '—'
})

const layers = computed(() => result.value?.layers ?? [])
const warnings = computed(() => result.value?.insights?.warnings ?? [])
const largestLayers = computed(() => result.value?.insights?.largest_layers ?? [])
const recommendations = computed(() => result.value?.recommendations ?? [])
const layerCountDisplay = computed(() => {
  if (result.value?.insights?.layer_count) {
    return result.value.insights.layer_count
  }
  if (layers.value.length) {
    return layers.value.length
  }
  return '—'
})

const manifestTypeLabel = computed(() => {
  const mediaType = result.value?.media_type
  if (!mediaType) {
    return '—'
  }
  if (mediaType.includes('docker.distribution.manifest.v2+json')) {
    return 'Docker'
  }
  if (mediaType.includes('oci.image.manifest.v1+json')) {
    return 'OCI'
  }
  return mediaType
})

const severityStyles = (severity) => {
  switch (severity) {
    case 'critical':
      return {
        container: 'border-rose-500/40 bg-rose-950/30',
        icon: 'bg-rose-400',
      }
    case 'warning':
      return {
        container: 'border-amber-500/40 bg-amber-950/30',
        icon: 'bg-amber-400',
      }
    case 'info':
      return {
        container: 'border-sky-500/40 bg-sky-950/30',
        icon: 'bg-sky-400',
      }
    default:
      return {
        container: 'border-slate-700 bg-slate-950/30',
        icon: 'bg-slate-400',
      }
  }
}

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
