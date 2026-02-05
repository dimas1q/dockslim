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
              v-if="analysis?.status === 'completed' && previousAnalysis"
              class="rounded-lg border border-slate-700 px-3 py-1 text-xs text-slate-200 hover:border-slate-500"
              @click="handleCompare"
            >
              Compare
            </button>
            <button
              v-if="isOwner"
              class="rounded-lg border border-slate-700 px-3 py-1 text-xs text-slate-200 hover:border-slate-500 disabled:opacity-60"
              :disabled="rerunning"
              @click="handleRerun"
            >
              {{ rerunning ? 'Re-running...' : 'Re-run analysis' }}
            </button>
            <button
              v-if="isOwner"
              class="rounded-lg border border-rose-500/60 px-3 py-1 text-xs text-rose-200 hover:border-rose-400 disabled:opacity-60"
              :disabled="deleting"
              @click="handleDeleteAnalysis"
            >
              {{ deleting ? 'Deleting...' : 'Delete analysis' }}
            </button>
            <span class="rounded-full px-3 py-1 text-xs font-semibold" :class="statusBadgeClass">
              {{ analysis?.status }}
            </span>
          </div>
        </div>
        <p
          v-if="analysis?.status === 'completed' && !previousAnalysis"
          class="text-xs text-slate-400"
        >
          No previous completed analysis to compare.
        </p>
        <p v-if="compareError" class="text-xs text-rose-400">{{ compareError }}</p>
        <p v-if="rerunError" class="text-sm text-rose-400">{{ rerunError }}</p>
        <p v-if="deleteError" class="text-sm text-rose-400">{{ deleteError }}</p>

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

        <div class="mt-6 rounded-xl border border-slate-800 bg-slate-950/50 p-6 space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-semibold text-slate-200">{{ baselineHeading }}</p>
              <p class="text-xs text-slate-400">
                Latest completed analysis on {{ baselineBranchLabel }}.
              </p>
            </div>
            <span
              v-if="baselineCompare"
              class="rounded-full px-3 py-1 text-xs font-semibold"
              :class="baselineStatusClass"
            >
              {{ baselineCompare.status }}
            </span>
          </div>
          <p v-if="baselineLoading" class="text-xs text-slate-400">Loading baseline comparison...</p>
          <p v-else-if="baselineError" class="text-xs text-slate-400">{{ baselineError }}</p>
          <div v-else-if="baselineCompare" class="space-y-4">
            <div class="grid gap-4 md:grid-cols-3 text-sm text-slate-300">
              <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-4">
                <p class="text-xs text-slate-500">Delta size</p>
                <p class="mt-1 font-semibold" :class="deltaClass(baselineCompare.deltas.total_size_bytes)">
                  {{ formatDeltaBytes(baselineCompare.deltas.total_size_bytes) }}
                </p>
              </div>
              <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-4">
                <p class="text-xs text-slate-500">Delta layers</p>
                <p class="mt-1 font-semibold" :class="deltaClass(baselineCompare.deltas.layer_count)">
                  {{ formatDeltaCount(baselineCompare.deltas.layer_count, 'layers') }}
                </p>
              </div>
              <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-4">
                <p class="text-xs text-slate-500">Delta largest layer</p>
                <p class="mt-1 font-semibold" :class="deltaClass(baselineCompare.deltas.largest_layer_bytes)">
                  {{ formatDeltaBytes(baselineCompare.deltas.largest_layer_bytes) }}
                </p>
              </div>
            </div>
            <div class="flex flex-wrap items-center justify-between gap-2 text-xs text-slate-400">
              <span>{{ baselineLabel }}</span>
              <RouterLink
                v-if="baselineLink"
                class="text-indigo-400 hover:text-indigo-300"
                :to="baselineLink"
              >
                View baseline
              </RouterLink>
            </div>
          </div>
          <p v-else class="text-xs text-slate-400">
            Baseline data will appear after the first main analysis completes.
          </p>
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
            <div v-if="recommendations.length" class="flex flex-wrap items-center gap-3 text-xs">
              <span class="rounded-full border px-3 py-1" :class="severityStyles('critical').container">
                Critical: {{ recommendationCounts.critical }}
              </span>
              <span class="rounded-full border px-3 py-1" :class="severityStyles('warning').container">
                Warnings: {{ recommendationCounts.warning }}
              </span>
              <span class="rounded-full border px-3 py-1" :class="severityStyles('info').container">
                Info: {{ recommendationCounts.info }}
              </span>
            </div>
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
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  deleteAnalysis,
  getAnalysis,
  getBaselineCompare,
  getProject,
  listAnalyses,
  rerunAnalysis,
} from '../api/client'

const route = useRoute()
const router = useRouter()
const projectId = route.params.id
const analysisId = route.params.analysisId

const analysis = ref(null)
const loading = ref(true)
const error = ref('')
const project = ref(null)
const polling = ref(false)
const rerunning = ref(false)
const rerunError = ref('')
const deleting = ref(false)
const deleteError = ref('')
const showRaw = ref(false)
const analyses = ref([])
const compareError = ref('')
const baselineCompare = ref(null)
const baselineLoading = ref(false)
const baselineError = ref('')
let pollTimer = null

const fetchAnalyses = async () => {
  compareError.value = ''
  try {
    analyses.value = await listAnalyses(projectId)
  } catch (err) {
    compareError.value = err.message
  }
}

const fetchAnalysis = async ({ silent = false } = {}) => {
  if (!silent) {
    loading.value = true
  }
  error.value = ''
  try {
    analysis.value = await getAnalysis(projectId, analysisId)
    if (!silent) {
      await fetchAnalyses()
    }
    if (!silent) {
      await fetchBaselineCompare()
    }
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

const previousAnalysis = computed(() => {
  if (!analysis.value || !analysis.value.created_at) {
    return null
  }
  const currentCreatedAt = new Date(analysis.value.created_at).getTime()
  if (!Number.isFinite(currentCreatedAt)) {
    return null
  }
  return analyses.value.reduce((latest, item) => {
    if (item.id === analysis.value.id) {
      return latest
    }
    if (item.image !== analysis.value.image || item.status !== 'completed') {
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
})

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

const fetchBaselineCompare = async () => {
  baselineError.value = ''
  baselineCompare.value = null
  if (!analysis.value || analysis.value.status !== 'completed') {
    return
  }
  baselineLoading.value = true
  try {
    baselineCompare.value = await getBaselineCompare(analysisId)
  } catch (err) {
    if (err.status === 404) {
      baselineError.value = 'No baseline on main yet.'
    } else {
      baselineError.value = err.message
    }
  } finally {
    baselineLoading.value = false
  }
}

watch(
  () => analysis.value?.status,
  (status, previous) => {
    if (status === 'completed' && (previous !== 'completed' || !baselineCompare.value)) {
      fetchBaselineCompare()
      return
    }
    if (status !== 'completed') {
      baselineCompare.value = null
      baselineError.value = ''
    }
  },
)

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

const baselineStatusClass = computed(() => {
  switch (baselineCompare.value?.status) {
    case 'OK':
      return 'bg-emerald-500/20 text-emerald-200'
    case 'WARN':
      return 'bg-amber-500/20 text-amber-200'
    case 'FAIL':
      return 'bg-rose-500/20 text-rose-200'
    default:
      return 'bg-slate-700/40 text-slate-200'
  }
})

const formatDeltaBytes = (value) => {
  if (value === null || value === undefined) return '—'
  const sign = value > 0 ? '+' : value < 0 ? '-' : ''
  return `${sign}${formatBytes(Math.abs(value))}`
}

const formatDeltaCount = (value, label) => {
  if (value === null || value === undefined) return '—'
  const sign = value > 0 ? '+' : value < 0 ? '-' : ''
  return `${sign}${Math.abs(value)} ${label}`
}

const deltaClass = (value) => {
  if (value > 0) return 'text-rose-200'
  if (value < 0) return 'text-emerald-200'
  return 'text-slate-200'
}

const baselineLabel = computed(() => {
  if (!baselineCompare.value) return ''
  const baseline = baselineCompare.value.baseline
  const parts = []
  if (baseline.tag) {
    parts.push(`Tag ${baseline.tag}`)
  }
  if (baseline.commit_sha) {
    parts.push(`Commit ${baseline.commit_sha.slice(0, 8)}`)
  }
  if (baseline.analyzed_at) {
    parts.push(new Date(baseline.analyzed_at).toLocaleString())
  }
  return parts.length ? `Baseline: ${parts.join(' · ')}` : 'Baseline: main'
})

const baselineLink = computed(() => {
  if (!baselineCompare.value) return ''
  return `/projects/${projectId}/analyses/${baselineCompare.value.baseline.analysis_id}`
})

const baselineBranchLabel = computed(() => {
  if (!baselineCompare.value) return 'main'
  return baselineCompare.value.baseline.ref_branch || 'main'
})

const baselineHeading = computed(() => `Compared to ${baselineBranchLabel.value}`)

const statusBadgeClass = computed(() => {
  switch (analysis.value?.status) {
    case 'completed':
      return 'bg-emerald-500/20 text-emerald-200'
    case 'running':
      return 'bg-sky-500/20 text-sky-200'
    case 'queued':
      return 'bg-slate-700/40 text-slate-200'
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
const recommendationCounts = computed(() =>
  recommendations.value.reduce(
    (counts, recommendation) => {
      if (counts[recommendation.severity] != null) {
        counts[recommendation.severity] += 1
      }
      return counts
    },
    { critical: 0, warning: 0, info: 0 },
  ),
)
const layerCountDisplay = computed(() => {
  if (result.value?.insights?.layer_count != null) {
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

const handleCompare = () => {
  if (!previousAnalysis.value) {
    return
  }
  router.push({
    path: `/projects/${projectId}/analyses/compare`,
    query: {
      from: previousAnalysis.value.id,
      to: analysisId,
    },
  })
}

const handleDeleteAnalysis = async () => {
  if (!analysis.value?.id) {
    return
  }
  const confirmed = window.confirm('Delete this analysis? This cannot be undone.')
  if (!confirmed) {
    return
  }

  deleteError.value = ''
  deleting.value = true
  try {
    await deleteAnalysis(projectId, analysis.value.id)
    router.push({ path: `/projects/${projectId}`, query: { analysisDeleted: '1' } })
  } catch (err) {
    deleteError.value = err.message
  } finally {
    deleting.value = false
  }
}
</script>
