<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" :to="`/projects/${projectId}`">
      ← Back to project
    </RouterLink>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-4">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h2 class="text-2xl font-semibold">History</h2>
          <p class="text-sm text-slate-400 mt-1">
            Browse completed analyses over time with quick filters.
          </p>
        </div>
        <button
          class="text-sm text-slate-300 hover:text-white"
          :disabled="loading"
          @click="fetchHistoryData"
        >
          Refresh
        </button>
      </div>

      <form class="grid gap-4 md:grid-cols-6" @submit.prevent="applyFilters">
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
        <div class="space-y-1">
          <label class="text-xs text-slate-400">Status</label>
          <select
            v-model="filters.status"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          >
            <option value="all">All</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="running">Running</option>
            <option value="queued">Queued</option>
          </select>
        </div>
        <div class="space-y-1">
          <label class="text-xs text-slate-400">Limit</label>
          <select
            v-model="filters.limit"
            class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
          >
            <option value="50">50</option>
            <option value="100">100</option>
            <option value="200">200</option>
            <option value="500">500</option>
          </select>
        </div>
        <div class="md:col-span-6 flex items-center gap-3">
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

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
      <p v-if="loading" class="text-sm text-slate-400">Loading history...</p>
      <p v-else-if="error" class="text-sm text-rose-400">{{ error }}</p>
      <p v-else-if="history.length === 0" class="text-sm text-slate-400">
        No analyses match these filters.
      </p>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="text-xs uppercase text-slate-500">
            <tr>
              <th class="py-2 pr-4">Time</th>
              <th class="py-2 pr-4">Image</th>
              <th class="py-2 pr-4">Branch</th>
              <th class="py-2 pr-4">Commit</th>
              <th class="py-2 pr-4">Size</th>
              <th class="py-2 pr-4">Layers</th>
              <th class="py-2">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800">
            <tr v-for="item in history" :key="item.id" class="text-slate-200">
              <td class="py-3 pr-4 text-slate-400">{{ formatDate(item.analyzed_at) }}</td>
              <td class="py-3 pr-4">
                <RouterLink
                  class="text-indigo-400 hover:text-indigo-300"
                  :to="`/projects/${projectId}/analyses/${item.id}`"
                >
                  {{ item.image }}
                </RouterLink>
              </td>
              <td class="py-3 pr-4 text-slate-300">{{ item.git_ref || '-' }}</td>
              <td class="py-3 pr-4 font-mono text-xs text-slate-300">{{ shortCommit(item.commit_sha) }}</td>
              <td class="py-3 pr-4 text-slate-300">
                {{ item.total_size_bytes ? formatBytes(item.total_size_bytes) : '-' }}
              </td>
              <td class="py-3 pr-4 text-slate-300">{{ item.layer_count ?? '-' }}</td>
              <td class="py-3">
                <span class="rounded-full px-2 py-1 text-xs font-semibold" :class="statusBadgeClass(item.status)">
                  {{ item.status }}
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
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchHistory } from '../api/client'

const route = useRoute()
const projectId = route.params.id

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
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

const formatBytes = (value) => {
  if (!value && value !== 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = Number(value)
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

const shortCommit = (value) => {
  if (!value) return '-'
  return value.slice(0, 8)
}

const statusBadgeClass = (status) => {
  switch (status) {
    case 'completed':
      return 'bg-emerald-500/20 text-emerald-200'
    case 'failed':
      return 'bg-rose-500/20 text-rose-200'
    case 'running':
      return 'bg-amber-500/20 text-amber-200'
    case 'queued':
      return 'bg-slate-700/40 text-slate-200'
    default:
      return 'bg-slate-700/40 text-slate-200'
  }
}

onMounted(fetchHistoryData)
</script>
