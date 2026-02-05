<template>
  <div class="space-y-10">
    <RouterLink class="link-subtle text-sm" :to="`/projects/${projectId}`">
      {{ t('nav.backToProject') }}
    </RouterLink>

    <section class="panel p-6 space-y-4 ds-reveal">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h2 class="text-2xl font-semibold text-ink">{{ t('projectHistory.title') }}</h2>
          <p class="text-sm text-muted mt-1">
            {{ t('projectHistory.subtitle') }}
          </p>
        </div>
        <button class="btn btn-ghost text-sm" :disabled="loading" @click="fetchHistoryData">
          {{ t('common.refresh') }}
        </button>
      </div>

      <form class="grid gap-4 md:grid-cols-6" @submit.prevent="applyFilters">
        <div class="space-y-1 md:col-span-2">
          <label class="text-xs font-medium text-subtle">{{ t('projectHistory.filters.image') }}</label>
          <input v-model="filters.image" type="text" placeholder="repo/name" class="input" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('projectHistory.filters.branch') }}</label>
          <input v-model="filters.git_ref" type="text" placeholder="main" class="input" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('projectHistory.filters.from') }}</label>
          <BaseDatePicker
            v-model="filters.from"
            :locale="locale"
            :placeholder="t('common.datePlaceholder')"
            :clear-label="t('common.clear')"
            :close-label="t('common.close')"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('projectHistory.filters.to') }}</label>
          <BaseDatePicker
            v-model="filters.to"
            :locale="locale"
            :placeholder="t('common.datePlaceholder')"
            :clear-label="t('common.clear')"
            :close-label="t('common.close')"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('projectHistory.filters.status') }}</label>
          <BaseSelect v-model="filters.status" :options="statusOptions" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-subtle">{{ t('projectHistory.filters.limit') }}</label>
          <BaseSelect v-model="filters.limit" :options="limitOptions" />
        </div>
        <div class="md:col-span-6 flex items-center gap-3">
          <button type="submit" class="btn btn-primary" :disabled="loading">
            {{ loading ? t('projectHistory.filters.applyLoading') : t('projectHistory.filters.apply') }}
          </button>
          <button type="button" class="btn btn-ghost" :disabled="loading" @click="clearFilters">
            {{ t('common.clear') }}
          </button>
          <p v-if="error" class="text-sm text-danger">{{ error }}</p>
        </div>
      </form>
    </section>

    <section class="panel p-6 ds-reveal">
      <div v-if="loading" class="space-y-3">
        <div class="h-12 rounded-xl skeleton"></div>
        <div class="h-12 rounded-xl skeleton"></div>
        <div class="h-12 rounded-xl skeleton"></div>
      </div>
      <p v-else-if="error" class="text-sm text-danger">{{ error }}</p>
      <p v-else-if="history.length === 0" class="text-sm text-muted">
        {{ t('projectHistory.empty') }}
      </p>
      <div v-else class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr>
              <th class="py-2 pr-4">{{ t('projectHistory.table.time') }}</th>
              <th class="py-2 pr-4">{{ t('projectHistory.filters.image') }}</th>
              <th class="py-2 pr-4">{{ t('projectHistory.table.branch') }}</th>
              <th class="py-2 pr-4">{{ t('projectHistory.table.commit') }}</th>
              <th class="py-2 pr-4">{{ t('projectHistory.table.size') }}</th>
              <th class="py-2 pr-4">{{ t('projectHistory.table.layers') }}</th>
              <th class="py-2">{{ t('common.status') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in history" :key="item.id">
              <td class="py-3 pr-4 text-muted">{{ formatDate(item.analyzed_at) }}</td>
              <td class="py-3 pr-4">
                <RouterLink class="link" :to="`/projects/${projectId}/analyses/${item.id}`">
                  {{ item.image }}
                </RouterLink>
              </td>
              <td class="py-3 pr-4 text-muted">{{ item.git_ref || t('common.empty') }}</td>
              <td class="py-3 pr-4 font-mono text-xs text-muted">{{ shortCommit(item.commit_sha) }}</td>
              <td class="py-3 pr-4 text-muted">
                {{ item.total_size_bytes ? formatBytes(item.total_size_bytes) : t('common.empty') }}
              </td>
              <td class="py-3 pr-4 text-muted">{{ item.layer_count ?? t('common.empty') }}</td>
              <td class="py-3">
                <span class="badge" :class="statusBadgeClass(item.status)">
                  {{ statusLabel(item.status) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { fetchHistory } from '../api/client'
import BaseSelect from '../components/BaseSelect.vue'
import BaseDatePicker from '../components/BaseDatePicker.vue'

const route = useRoute()
const projectId = route.params.id
const { locale, t, tm } = useI18n()

const history = ref([])
const loading = ref(false)
const error = ref('')
const filters = ref({
  image: '',
  git_ref: '',
  from: '',
  to: '',
  status: 'all',
  limit: '100',
})

const statusOptions = computed(() => [
  { value: 'all', label: t('projectHistory.statusAll') },
  { value: 'completed', label: t('status.completed') },
  { value: 'failed', label: t('status.failed') },
  { value: 'running', label: t('status.running') },
  { value: 'queued', label: t('status.queued') },
])

const limitOptions = computed(() => [
  { value: '50', label: '50' },
  { value: '100', label: '100' },
  { value: '200', label: '200' },
  { value: '500', label: '500' },
])

const buildParams = () => {
  const params = {}
  if (filters.value.image.trim()) params.image = filters.value.image.trim()
  if (filters.value.git_ref.trim()) params.git_ref = filters.value.git_ref.trim()
  if (filters.value.from) params.from = filters.value.from
  if (filters.value.to) params.to = filters.value.to
  if (filters.value.status && filters.value.status !== 'all') params.status = filters.value.status
  if (filters.value.limit) params.limit = filters.value.limit
  return params
}

const fetchHistoryData = async () => {
  loading.value = true
  error.value = ''
  try {
    history.value = await fetchHistory(projectId, buildParams())
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  fetchHistoryData()
}

const clearFilters = () => {
  filters.value = { image: '', git_ref: '', from: '', to: '', status: 'all', limit: '100' }
  fetchHistoryData()
}

const formatDate = (value) => {
  if (!value) return t('common.empty')
  return new Date(value).toLocaleString(locale.value)
}

const formatBytes = (value) => {
  if (!value && value !== 0) return t('common.empty')
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

const shortCommit = (value) => {
  if (!value) return t('common.empty')
  return value.slice(0, 8)
}

const statusBadgeClass = (status) => {
  switch (status) {
    case 'completed':
      return 'badge-success'
    case 'failed':
      return 'badge-danger'
    case 'running':
      return 'badge-warning'
    case 'queued':
      return 'badge-neutral'
    default:
      return 'badge-neutral'
  }
}

const statusLabel = (status) => {
  if (!status) {
    return t('common.empty')
  }
  return t(`status.${status}`)
}

onMounted(fetchHistoryData)
</script>
