<template>
  <div class="space-y-10">
    <RouterLink class="link-subtle text-sm" :to="`/projects/${projectId}`">
      {{ t('nav.backToProject') }}
    </RouterLink>

    <section class="panel p-6 space-y-4 ds-reveal">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h2 class="text-2xl font-semibold text-ink">{{ t('projectTrends.title') }}</h2>
          <p class="text-sm text-muted mt-1">
            {{ t('projectTrends.subtitle') }}
          </p>
        </div>
        <button class="btn btn-ghost text-sm" :disabled="loading" @click="fetchTrendsData">
          {{ t('common.refresh') }}
        </button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="metric in metrics"
          :key="metric.key"
          class="rounded-full border px-3 py-1 text-xs font-semibold transition"
          :class="metric.key === selectedMetric ? 'border-primary/50 bg-primary/10 text-primary' : 'border-border text-muted hover:border-primary/40'"
          type="button"
          @click="selectedMetric = metric.key"
        >
          {{ metric.label }}
        </button>
      </div>

      <form class="grid gap-4 md:grid-cols-5" @submit.prevent="applyFilters">
        <div class="space-y-1 md:col-span-2">
          <label class="text-xs font-medium text-subtle">{{ t('projectTrends.filters.image') }}</label>
          <input v-model="filters.image" type="text" placeholder="repo/name" class="input" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('projectTrends.filters.branch') }}</label>
          <input v-model="filters.git_ref" type="text" placeholder="main" class="input" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('projectTrends.filters.from') }}</label>
          <BaseDatePicker
            v-model="filters.from"
            :locale="locale"
            :placeholder="t('common.datePlaceholder')"
            :clear-label="t('common.clear')"
            :close-label="t('common.close')"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('projectTrends.filters.to') }}</label>
          <BaseDatePicker
            v-model="filters.to"
            :locale="locale"
            :placeholder="t('common.datePlaceholder')"
            :clear-label="t('common.clear')"
            :close-label="t('common.close')"
          />
        </div>
        <div class="md:col-span-5 flex items-center gap-3">
          <button type="submit" class="btn btn-primary" :disabled="loading">
            {{ loading ? t('projectTrends.filters.applyLoading') : t('projectTrends.filters.apply') }}
          </button>
          <button type="button" class="btn btn-ghost" :disabled="loading" @click="clearFilters">
            {{ t('common.clear') }}
          </button>
          <p v-if="error" class="text-sm text-danger">{{ error }}</p>
        </div>
      </form>
    </section>

    <section class="panel p-6 space-y-4 ds-reveal">
      <div v-if="loading" class="space-y-3">
        <div class="h-12 rounded-xl skeleton"></div>
        <div class="h-12 rounded-xl skeleton"></div>
        <div class="h-12 rounded-xl skeleton"></div>
      </div>
      <p v-else-if="error" class="text-sm text-danger">{{ error }}</p>
      <p v-else-if="trendPoints.length === 0" class="text-sm text-muted">
        {{ t('projectTrends.empty') }}
      </p>
      <div v-else class="space-y-4">
        <div class="grid gap-4 md:grid-cols-3">
          <div class="stat-card">
            <p class="text-xs text-subtle">{{ t('projectTrends.latest') }}</p>
            <p class="mt-1 text-lg font-semibold">{{ formatMetricValue(latestValue) }}</p>
          </div>
          <div class="stat-card">
            <p class="text-xs text-subtle">{{ t('projectTrends.minimum') }}</p>
            <p class="mt-1 text-lg font-semibold">{{ formatMetricValue(minValue) }}</p>
          </div>
          <div class="stat-card">
            <p class="text-xs text-subtle">{{ t('projectTrends.maximum') }}</p>
            <p class="mt-1 text-lg font-semibold">{{ formatMetricValue(maxValue) }}</p>
          </div>
        </div>

        <div class="surface p-4">
          <div class="flex items-center justify-between text-xs text-muted">
            <span>{{ metricLabel }}</span>
            <span>{{ timeRangeLabel }}</span>
          </div>
          <div class="mt-4 h-56">
            <svg v-if="chartPoints" viewBox="0 0 100 60" class="h-full w-full" preserveAspectRatio="none">
              <defs>
                <linearGradient id="trendStroke" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="rgb(var(--ds-chart-start))" />
                  <stop offset="100%" stop-color="rgb(var(--ds-chart-end))" />
                </linearGradient>
                <linearGradient id="trendFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="rgb(var(--ds-chart-start))" stop-opacity="0.28" />
                  <stop offset="100%" stop-color="rgb(var(--ds-chart-end))" stop-opacity="0" />
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
import { useI18n } from 'vue-i18n'
import { fetchTrends } from '../api/client'
import BaseDatePicker from '../components/BaseDatePicker.vue'

const route = useRoute()
const projectId = route.params.id
const { locale, t, tm } = useI18n()

const metrics = computed(() => [
  { key: 'total_size_bytes', label: t('projectTrends.metrics.totalSize'), format: 'bytes' },
  { key: 'layer_count', label: t('projectTrends.metrics.layerCount'), format: 'count' },
  { key: 'largest_layer_bytes', label: t('projectTrends.metrics.largestLayer'), format: 'bytes' },
])

const selectedMetric = ref('total_size_bytes')
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

const metricLabel = computed(() => metrics.value.find((metric) => metric.key === selectedMetric.value)?.label || '')

const timeRangeLabel = computed(() => {
  if (trendPoints.value.length === 0) return ''
  const start = new Date(trendPoints.value[0].ts).toLocaleDateString(locale.value)
  const end = new Date(trendPoints.value[trendPoints.value.length - 1].ts).toLocaleDateString(locale.value)
  return t('projectTrends.range', { start, end })
})

const formatBytes = (value) => {
  if (value === null || value === undefined) return t('common.empty')
  const units = tm('units.bytes')
  const unitList = Array.isArray(units) && units.length ? units : ['B', 'KB', 'MB', 'GB', 'TB']
  let size = Number(value)
  let unitIndex = 0
  while (size >= 1024 && unitIndex < unitList.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${unitList[unitIndex]}`
}

const formatMetricValue = (value) => {
  if (value === null || value === undefined) return t('common.empty')
  const config = metrics.value.find((metric) => metric.key === selectedMetric.value)
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
