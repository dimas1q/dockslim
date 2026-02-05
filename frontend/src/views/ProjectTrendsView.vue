<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" :to="`/projects/${projectId}`">
      ← Back to project
    </RouterLink>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-4">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h2 class="text-2xl font-semibold">Trends</h2>
          <p class="text-sm text-slate-400 mt-1">
            Track size and layer growth for your images.
          </p>
        </div>
        <button
          class="text-sm text-slate-300 hover:text-white"
          :disabled="loading"
          @click="fetchTrendsData"
        >
          Refresh
        </button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="metric in metrics"
          :key="metric.key"
          class="rounded-full border px-3 py-1 text-xs font-semibold"
          :class="metric.key === selectedMetric ? 'border-indigo-400 text-indigo-200 bg-indigo-500/10' : 'border-slate-700 text-slate-300 hover:border-slate-500'"
          type="button"
          @click="selectedMetric = metric.key"
        >
          {{ metric.label }}
        </button>
      </div>

      <form class="grid gap-4 md:grid-cols-5" @submit.prevent="applyFilters">
        <div class="space-y-1 md:col-span-2">
          <label class="text-xs text-slate-400">Image</label>
          <input
            v-model="filters.image"
            type="text"
            placeholder="repo/name"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-slate-400">Branch</label>
          <input
            v-model="filters.git_ref"
            type="text"
            placeholder="main"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-slate-400">From</label>
          <input
            v-model="filters.from"
            type="date"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-slate-400">To</label>
          <input
            v-model="filters.to"
            type="date"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          />
        </div>
        <div class="md:col-span-5 flex items-center gap-3">
          <button
            type="submit"
            class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
            :disabled="loading"
          >
            {{ loading ? 'Loading...' : 'Apply filters' }}
          </button>
          <button
            type="button"
            class="text-sm text-slate-400 hover:text-slate-200"
            :disabled="loading"
            @click="clearFilters"
          >
            Clear
          </button>
          <p v-if="error" class="text-sm text-rose-400">{{ error }}</p>
        </div>
      </form>
    </section>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-4">
      <p v-if="loading" class="text-sm text-slate-400">Loading trend data...</p>
      <p v-else-if="error" class="text-sm text-rose-400">{{ error }}</p>
      <p v-else-if="trendPoints.length === 0" class="text-sm text-slate-400">
        No data for this selection yet.
      </p>
      <div v-else class="space-y-4">
        <div class="grid gap-4 md:grid-cols-3">
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Latest</p>
            <p class="mt-1 text-lg font-semibold">{{ formatMetricValue(latestValue) }}</p>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Minimum</p>
            <p class="mt-1 text-lg font-semibold">{{ formatMetricValue(minValue) }}</p>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">Maximum</p>
            <p class="mt-1 text-lg font-semibold">{{ formatMetricValue(maxValue) }}</p>
          </div>
        </div>

        <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
          <div class="flex items-center justify-between text-xs text-slate-400">
            <span>{{ metricLabel }}</span>
            <span>{{ timeRangeLabel }}</span>
          </div>
          <div class="mt-4 h-56">
            <svg
              v-if="chartPoints"
              viewBox="0 0 100 60"
              class="h-full w-full"
              preserveAspectRatio="none"
            >
              <defs>
                <linearGradient id="trendStroke" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="#60a5fa" />
                  <stop offset="100%" stop-color="#a78bfa" />
                </linearGradient>
                <linearGradient id="trendFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#60a5fa" stop-opacity="0.35" />
                  <stop offset="100%" stop-color="#0f172a" stop-opacity="0" />
                </linearGradient>
              </defs>
              <polygon :points="areaPoints" fill="url(#trendFill)" />
              <polyline
                :points="chartPoints"
                fill="none"
                stroke="url(#trendStroke)"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchTrends } from '../api/client'

const route = useRoute()
const projectId = route.params.id

const metrics = [
  { key: 'total_size_bytes', label: 'Total size', format: 'bytes' },
  { key: 'layer_count', label: 'Layer count', format: 'count' },
  { key: 'largest_layer_bytes', label: 'Largest layer', format: 'bytes' },
]

const selectedMetric = ref(metrics[0].key)
const trend = ref([])
const loading = ref(false)
const error = ref('')
const filters = ref({
  image: '',
  git_ref: '',
  from: '',
  to: '',
})

const buildParams = () => {
  const params = { metric: selectedMetric.value }
  if (filters.value.image.trim()) params.image = filters.value.image.trim()
  if (filters.value.git_ref.trim()) params.git_ref = filters.value.git_ref.trim()
  if (filters.value.from) params.from = filters.value.from
  if (filters.value.to) params.to = filters.value.to
  return params
}

const fetchTrendsData = async () => {
  loading.value = true
  error.value = ''
  try {
    trend.value = await fetchTrends(projectId, buildParams())
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  fetchTrendsData()
}

const clearFilters = () => {
  filters.value = { image: '', git_ref: '', from: '', to: '' }
  fetchTrendsData()
}

const trendPoints = computed(() => trend.value || [])

const values = computed(() => trendPoints.value.map((point) => point.value))
const times = computed(() => trendPoints.value.map((point) => new Date(point.ts).getTime()))

const minValue = computed(() => (values.value.length ? Math.min(...values.value) : null))
const maxValue = computed(() => (values.value.length ? Math.max(...values.value) : null))
const latestValue = computed(() => (values.value.length ? values.value[values.value.length - 1] : null))

const chartPoints = computed(() => {
  if (trendPoints.value.length === 0) return ''
  const minVal = minValue.value ?? 0
  const maxVal = maxValue.value ?? 1
  const minTime = Math.min(...times.value)
  const maxTime = Math.max(...times.value)
  const range = maxVal - minVal || 1
  const timeRange = maxTime - minTime || 1

  return trendPoints.value
    .map((point) => {
      const ts = new Date(point.ts).getTime()
      const x = ((ts - minTime) / timeRange) * 100
      const y = 55 - ((point.value - minVal) / range) * 45
      return `${x.toFixed(2)},${y.toFixed(2)}`
    })
    .join(' ')
})

const areaPoints = computed(() => {
  if (!chartPoints.value) return ''
  return `${chartPoints.value} 100,55 0,55`
})

const metricLabel = computed(() => metrics.find((metric) => metric.key === selectedMetric.value)?.label || '')

const timeRangeLabel = computed(() => {
  if (trendPoints.value.length === 0) return ''
  const start = new Date(trendPoints.value[0].ts).toLocaleDateString()
  const end = new Date(trendPoints.value[trendPoints.value.length - 1].ts).toLocaleDateString()
  return `${start} to ${end}`
})

const formatBytes = (value) => {
  if (value === null || value === undefined) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = Number(value)
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

const formatMetricValue = (value) => {
  if (value === null || value === undefined) return '-'
  const config = metrics.find((metric) => metric.key === selectedMetric.value)
  if (config?.format === 'bytes') {
    return formatBytes(value)
  }
  return `${value}`
}

watch(selectedMetric, () => {
  fetchTrendsData()
})

onMounted(fetchTrendsData)
</script>
