<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" :to="`/projects/${projectId}`">
      {{ t('nav.backToProject') }}
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
          <h2 class="text-2xl font-semibold">{{ t('analysisCompare.title') }}</h2>
          <p class="text-sm text-slate-400">{{ comparison?.image }}</p>
          <p class="text-xs text-slate-500">
            {{ t('analysisCompare.from', { tag: comparison?.from?.tag, date: formatDate(comparison?.from?.created_at) }) }}
            <span class="mx-2 text-slate-600">→</span>
            {{ t('analysisCompare.to', { tag: comparison?.to?.tag, date: formatDate(comparison?.to?.created_at) }) }}
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-4 mt-6">
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">{{ t('analysisCompare.totalSizeChange') }}</p>
            <p class="mt-1 text-lg font-semibold" :class="sizeChangeClass">
              {{ totalSizeDiffLabel }}
            </p>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">{{ t('analysisCompare.layerCountChange') }}</p>
            <p class="mt-1 text-lg font-semibold text-slate-100">
              {{ layerCountDiffLabel }}
            </p>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4">
            <p class="text-xs text-slate-500">{{ t('analysisCompare.impact') }}</p>
            <span
              class="mt-2 inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold"
              :class="impactBadgeClass"
            >
              {{ impactLabel }}
            </span>
          </div>
          <div class="rounded-xl border border-slate-800 bg-slate-950/50 p-4 space-y-2">
            <div class="flex items-center justify-between">
              <p class="text-xs text-slate-500">{{ t('analysisCompare.budgetVerdict') }}</p>
              <span
                class="inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold"
                :class="budgetBadgeClass"
              >
                {{ budgetStatusLabel }}
              </span>
            </div>
            <div v-if="budgetReasons.length" class="space-y-1">
              <p class="text-xs text-slate-500">{{ t('analysisCompare.reasons') }}</p>
              <ul class="text-xs text-slate-200 list-disc list-inside space-y-0.5">
                <li v-for="reason in budgetReasons" :key="reason">{{ reason }}</li>
              </ul>
            </div>
            <p v-else class="text-xs text-slate-500">{{ t('analysisCompare.noBudgetConfigured') }}</p>
            <p v-if="budgetThresholdLabel" class="text-[11px] text-slate-500">
              {{ t('analysisCompare.thresholds', { label: budgetThresholdLabel }) }}
            </p>
          </div>
        </div>

        <div class="grid gap-6 lg:grid-cols-2 mt-8">
          <div class="rounded-xl border border-emerald-500/30 bg-emerald-950/20 p-5 space-y-4">
            <div class="flex items-center justify-between">
              <p class="text-sm font-semibold text-emerald-200">{{ t('analysisCompare.addedLayers') }}</p>
              <span class="text-xs text-emerald-300">{{ addedLayers.length }}</span>
            </div>
            <div v-if="addedLayers.length" class="overflow-hidden rounded-lg border border-emerald-500/30">
              <table class="min-w-full text-left text-sm text-slate-200">
                <thead class="bg-emerald-950/40 text-xs uppercase text-emerald-300">
                  <tr>
                    <th class="px-4 py-3">{{ t('analysisDetail.digest') }}</th>
                    <th class="px-4 py-3">{{ t('analysisDetail.size') }}</th>
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
            <p v-else class="text-sm text-emerald-200/80">{{ t('analysisCompare.noAddedLayers') }}</p>
          </div>

          <div class="rounded-xl border border-rose-500/30 bg-rose-950/20 p-5 space-y-4">
            <div class="flex items-center justify-between">
              <p class="text-sm font-semibold text-rose-200">{{ t('analysisCompare.removedLayers') }}</p>
              <span class="text-xs text-rose-300">{{ removedLayers.length }}</span>
            </div>
            <div v-if="removedLayers.length" class="overflow-hidden rounded-lg border border-rose-500/30">
              <table class="min-w-full text-left text-sm text-slate-200">
                <thead class="bg-rose-950/40 text-xs uppercase text-rose-300">
                  <tr>
                    <th class="px-4 py-3">{{ t('analysisDetail.digest') }}</th>
                    <th class="px-4 py-3">{{ t('analysisDetail.size') }}</th>
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
            <p v-else class="text-sm text-rose-200/80">{{ t('analysisCompare.noRemovedLayers') }}</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { compareAnalyses } from '../api/client'

const route = useRoute()
const projectId = route.params.id
const { locale, t, tm } = useI18n()

const comparison = ref(null)
const loading = ref(true)
const error = ref('')

const fetchComparison = async () => {
  loading.value = true
  error.value = ''

  const fromId = route.query.from
  const toId = route.query.to
  if (!fromId || !toId) {
    error.value = t('analysisCompare.missingParams')
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

watch(
  () => [route.query.from, route.query.to],
  () => {
    fetchComparison()
  },
  { immediate: true },
)

const formatDate = (value) => {
  if (!value) return t('common.justNow')
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

const shortDigest = (value) => {
  if (!value) return t('common.empty')
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
const budgetResult = computed(() => comparison.value?.budget)

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
  if (totalSizeDiff.value > 0) return t('analysisCompare.impactRegression')
  if (totalSizeDiff.value < 0) return t('analysisCompare.impactImprovement')
  return t('analysisCompare.impactNoChange')
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

const budgetStatusLabel = computed(() => {
  const status = budgetResult.value?.status
  if (!status) return t('analysisCompare.budgetStatusNone')
  switch (status) {
    case 'fail':
      return t('analysisCompare.budgetStatusFail')
    case 'warn':
      return t('analysisCompare.budgetStatusWarn')
    default:
      return t('analysisCompare.budgetStatusOk')
  }
})

const budgetBadgeClass = computed(() => {
  const status = budgetResult.value?.status
  if (status === 'fail') return 'bg-rose-500/20 text-rose-200'
  if (status === 'warn') return 'bg-amber-500/20 text-amber-200'
  if (status === 'ok') return 'bg-emerald-500/20 text-emerald-200'
  return 'bg-slate-700/60 text-slate-200'
})

const budgetReasons = computed(() => budgetResult.value?.reasons ?? [])

const formatMB = (bytes) => {
  if (bytes === null || bytes === undefined) return null
  return Math.round(Number(bytes) / (1024 * 1024))
}

const budgetThresholdLabel = computed(() => {
  if (!budgetResult.value) return ''
  const warn = formatMB(budgetResult.value.warn_delta_bytes)
  const fail = formatMB(budgetResult.value.fail_delta_bytes)
  const hard = formatMB(budgetResult.value.hard_limit_bytes)
  const parts = []
  if (warn !== null) parts.push(t('analysisCompare.budgetWarnDelta', { value: warn }))
  if (fail !== null) parts.push(t('analysisCompare.budgetFailDelta', { value: fail }))
  if (hard !== null) parts.push(t('analysisCompare.budgetHardLimit', { value: hard }))
  return parts.join(' · ')
})
</script>
