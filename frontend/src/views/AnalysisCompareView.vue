<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" :to="`/projects/${projectId}`">
      ← Back to project
    </RouterLink>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6">
      <div v-if="loading" class="space-y-4 animate-pulse">
        <div class="h-6 w-48 rounded bg-slate-800/80"></div>
        <div class="h-4 w-72 rounded bg-slate-800/60"></div>
        <div class="grid gap-4 md:grid-cols-3">
          <div class="h-20 rounded-xl bg-slate-800/70"></div>
          <div class="h-20 rounded-xl bg-slate-800/70"></div>
          <div class="h-20 rounded-xl bg-slate-800/70"></div>
        </div>
      </div>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>
      <div v-else>
        <div class="space-y-2">
          <h2 class="text-2xl font-semibold">Compare Analyses</h2>
          <p class="text-sm text-slate-400">{{ comparison?.image }}</p>
          <p class="text-xs text-slate-500">
            From {{ comparison?.from?.tag }} · {{ formatDate(comparison?.from?.created_at) }}
            <span class="mx-2 text-slate-600">→</span>
            To {{ comparison?.to?.tag }} · {{ formatDate(comparison?.to?.created_at) }}
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-3 mt-6">
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Total size change</p>
            <p class="mt-1 text-lg font-semibold" :class="sizeChangeClass">
              {{ totalSizeDiffLabel }}
            </p>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Layer count change</p>
            <p class="mt-1 text-lg font-semibold text-slate-100">
              {{ layerCountDiffLabel }}
            </p>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Impact</p>
            <span
              class="mt-2 inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold"
              :class="impactBadgeClass"
            >
              {{ impactLabel }}
            </span>
          </div>
        </div>

        <div class="grid gap-6 lg:grid-cols-2 mt-8">
          <div class="rounded-xl border border-emerald-500/30 bg-emerald-950/20 p-5 space-y-4">
            <div class="flex items-center justify-between">
              <p class="text-sm font-semibold text-emerald-200">Added layers</p>
              <span class="text-xs text-emerald-300">{{ addedLayers.length }}</span>
            </div>
            <div v-if="addedLayers.length" class="overflow-hidden rounded-lg border border-emerald-500/30">
              <table class="min-w-full text-left text-sm text-slate-200">
                <thead class="bg-emerald-950/40 text-xs uppercase text-emerald-300">
                  <tr>
                    <th class="px-4 py-3">Digest</th>
                    <th class="px-4 py-3">Size</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="layer in addedLayers" :key="layer.digest" class="border-t border-emerald-500/20">
                    <td class="px-4 py-3 font-mono text-xs">{{ shortDigest(layer.digest) }}</td>
                    <td class="px-4 py-3">{{ formatBytes(layer.size_bytes) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else class="text-sm text-emerald-200/80">No added layers.</p>
          </div>

          <div class="rounded-xl border border-rose-500/30 bg-rose-950/20 p-5 space-y-4">
            <div class="flex items-center justify-between">
              <p class="text-sm font-semibold text-rose-200">Removed layers</p>
              <span class="text-xs text-rose-300">{{ removedLayers.length }}</span>
            </div>
            <div v-if="removedLayers.length" class="overflow-hidden rounded-lg border border-rose-500/30">
              <table class="min-w-full text-left text-sm text-slate-200">
                <thead class="bg-rose-950/40 text-xs uppercase text-rose-300">
                  <tr>
                    <th class="px-4 py-3">Digest</th>
                    <th class="px-4 py-3">Size</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="layer in removedLayers" :key="layer.digest" class="border-t border-rose-500/20">
                    <td class="px-4 py-3 font-mono text-xs">{{ shortDigest(layer.digest) }}</td>
                    <td class="px-4 py-3">{{ formatBytes(layer.size_bytes) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else class="text-sm text-rose-200/80">No removed layers.</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { compareAnalyses } from '../api/client'

const route = useRoute()
const projectId = route.params.id

const comparison = ref(null)
const loading = ref(true)
const error = ref('')

const fetchComparison = async () => {
  loading.value = true
  error.value = ''

  const fromId = route.query.from
  const toId = route.query.to
  if (!fromId || !toId) {
    error.value = 'Missing comparison parameters.'
    loading.value = false
    return
  }

  try {
    comparison.value = await compareAnalyses(projectId, fromId, toId)
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

onMounted(fetchComparison)

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

const totalSizeDiff = computed(() => comparison.value?.summary?.total_size_diff_bytes ?? 0)
const layerCountDiff = computed(() => comparison.value?.summary?.layer_count_diff ?? 0)

const totalSizeDiffLabel = computed(() => {
  const diff = totalSizeDiff.value
  const sign = diff > 0 ? '+' : diff < 0 ? '-' : ''
  return `${sign}${formatBytes(Math.abs(diff))}`
})

const layerCountDiffLabel = computed(() => {
  const diff = layerCountDiff.value
  const sign = diff > 0 ? '+' : diff < 0 ? '-' : ''
  return `${sign}${Math.abs(diff)}`
})

const impactLabel = computed(() => {
  if (totalSizeDiff.value > 0) return 'Regression'
  if (totalSizeDiff.value < 0) return 'Improvement'
  return 'No change'
})

const impactBadgeClass = computed(() => {
  if (totalSizeDiff.value > 0) return 'bg-rose-500/20 text-rose-200'
  if (totalSizeDiff.value < 0) return 'bg-emerald-500/20 text-emerald-200'
  return 'bg-slate-700/60 text-slate-200'
})

const sizeChangeClass = computed(() => {
  if (totalSizeDiff.value > 0) return 'text-rose-200'
  if (totalSizeDiff.value < 0) return 'text-emerald-200'
  return 'text-slate-100'
})

const addedLayers = computed(() => comparison.value?.layers?.added ?? [])
const removedLayers = computed(() => comparison.value?.layers?.removed ?? [])
</script>
