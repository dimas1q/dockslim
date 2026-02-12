<template>
  <div class="space-y-10">
    <section class="panel p-6 ds-reveal">
      <p class="text-xs uppercase tracking-[0.2em] text-subtle">{{ t('projects.loggedInAs') }}</p>
      <p class="text-lg font-semibold text-ink">{{ auth.user?.email || t('projects.loadingAccount') }}</p>
    </section>

    <section class="panel p-6 space-y-4 ds-reveal">
      <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
        <h2 class="text-xl font-semibold">{{ t('projects.listTitle') }}</h2>
        <div class="flex items-center gap-2">
          <button class="btn btn-secondary text-sm" @click="openCreateModal">
            {{ t('projects.createButton') }}
          </button>
        </div>
      </div>

      <div v-if="deletedNotice" class="callout callout-success">
        {{ t('projects.deletedNotice') }}
      </div>
      <div v-if="loading" class="space-y-3">
        <div class="h-14 rounded-xl skeleton"></div>
        <div class="h-14 rounded-xl skeleton"></div>
        <div class="h-14 rounded-xl skeleton"></div>
      </div>
      <p v-else-if="error" class="text-sm text-danger">{{ error }}</p>
      <p v-else-if="projects.length === 0" class="text-sm text-muted">
        {{ t('projects.empty') }}
      </p>

      <ul v-else class="space-y-3">
        <li
          v-for="project in projects"
          :key="project.id"
          class="flex items-center justify-between rounded-xl border border-border bg-card/60 px-4 py-3 transition hover:border-primary/40"
        >
          <div>
            <p class="font-medium text-ink">{{ project.name }}</p>
            <p class="text-xs text-subtle">
              {{ t('projects.createdAt', { date: formatDate(project.created_at) }) }}
            </p>
          </div>
          <RouterLink class="link text-sm" :to="`/projects/${project.id}`">
            {{ t('projects.viewDetails') }}
          </RouterLink>
        </li>
      </ul>
    </section>

    <section class="panel p-6 space-y-6 ds-reveal">
      <div class="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
        <div>
          <h2 class="text-xl font-semibold text-ink">{{ t('projects.dashboard.title') }}</h2>
          <p class="text-sm text-muted mt-1">{{ t('projects.dashboard.subtitle') }}</p>
        </div>
      </div>

      <div v-if="dashboardLoading" class="space-y-4">
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div v-for="item in 4" :key="`dash-skeleton-${item}`" class="h-24 rounded-xl skeleton"></div>
        </div>
        <div class="grid gap-4 xl:grid-cols-2">
          <div class="h-48 rounded-xl skeleton"></div>
          <div class="h-48 rounded-xl skeleton"></div>
        </div>
      </div>

      <div v-else class="space-y-5">
        <div v-if="dashboardError" class="callout callout-warning">
          {{ t('projects.dashboard.loadError') }}: {{ dashboardError }}
        </div>

        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <article v-for="card in summaryCards" :key="card.key" class="surface p-4 space-y-1">
            <p class="text-xs uppercase tracking-[0.14em] text-subtle">{{ card.label }}</p>
            <p class="text-2xl font-semibold text-ink">{{ card.value }}</p>
            <p class="text-xs text-muted">{{ card.hint }}</p>
          </article>
        </div>

        <div class="grid gap-4 xl:grid-cols-2">
          <article class="surface p-5 space-y-3">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-ink">{{ t('projects.dashboard.heatmapTitle') }}</p>
                <p class="text-xs text-muted mt-1">
                  {{ t('projects.dashboard.heatmapSubtitle', { count: formatNumber(totalContributions) }) }}
                </p>
              </div>
            </div>

            <div v-if="activityPoints.length" class="space-y-3">
              <div class="overflow-x-auto">
                <div class="min-w-[420px]">
                  <div class="grid grid-flow-col auto-cols-[minmax(12px,1fr)] grid-rows-7 gap-1">
                    <div
                      v-for="point in activityPoints"
                      :key="`activity-${point.date}`"
                      class="h-3 w-3 rounded-[4px] transition-opacity hover:opacity-80"
                      :class="heatmapCellClass(point.level)"
                      :title="t('projects.dashboard.heatmapCellTitle', { date: formatDay(point.date), count: point.count })"
                    ></div>
                  </div>
                </div>
              </div>
              <div class="flex items-center justify-between text-xs text-muted">
                <span>{{ t('projects.dashboard.less') }}</span>
                <div class="flex items-center gap-1">
                  <span
                    v-for="level in [0, 1, 2, 3, 4]"
                    :key="`legend-${level}`"
                    class="h-3 w-3 rounded-[4px]"
                    :class="heatmapCellClass(level)"
                  ></span>
                </div>
                <span>{{ t('projects.dashboard.more') }}</span>
              </div>
            </div>

            <p v-else class="text-sm text-muted">{{ t('projects.dashboard.noActivity') }}</p>
          </article>

          <article class="surface p-5 space-y-3">
            <div>
              <p class="text-sm font-semibold text-ink">{{ t('projects.dashboard.recentTitle') }}</p>
              <p class="text-xs text-muted mt-1">{{ t('projects.dashboard.recentSubtitle') }}</p>
            </div>

            <p v-if="recentEvents.length === 0" class="text-sm text-muted">{{ t('projects.dashboard.noEvents') }}</p>

            <ul v-else class="space-y-2">
              <li
                v-for="event in recentEvents"
                :key="`${event.type}-${event.occurred_at}-${event.project_id}-${event.analysis_id || 'none'}`"
                class="rounded-lg border border-border bg-card/60 px-3 py-2"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="space-y-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="badge text-[10px]" :class="eventBadgeClass(event.type)">
                        {{ t(`projects.dashboard.eventTypes.${event.type}`) }}
                      </span>
                      <RouterLink class="link text-sm" :to="eventLink(event)">
                        {{ event.project_name }}
                      </RouterLink>
                    </div>
                    <p class="text-xs text-muted">{{ eventDescription(event) }}</p>
                  </div>
                  <span class="text-xs text-subtle whitespace-nowrap">{{ formatDate(event.occurred_at) }}</span>
                </div>
              </li>
            </ul>
          </article>
        </div>
      </div>
    </section>

    <Transition name="modal-fade">
      <div
        v-if="showCreateModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4"
      >
        <div class="card w-full max-w-lg p-6 space-y-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h3 class="text-lg font-semibold">{{ t('projects.createTitle') }}</h3>
              <p class="text-sm text-muted">{{ t('projects.createSubtitle') }}</p>
            </div>
            <button class="btn btn-ghost btn-icon" type="button" @click="closeCreateModal" aria-label="Close">
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path :d="closeIcon" />
              </svg>
            </button>
          </div>
          <form class="flex flex-col gap-3 md:flex-row md:items-center" @submit.prevent="handleCreate">
            <input
              v-model="newName"
              type="text"
              :placeholder="t('projects.createPlaceholder')"
              class="input flex-1"
            />
            <button type="submit" class="btn btn-primary" :disabled="creating">
              {{ creating ? t('projects.createButtonLoading') : t('projects.createButton') }}
            </button>
          </form>
          <p v-if="createError" class="text-sm text-danger">{{ createError }}</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { createProject, fetchAccountDashboard, listProjects } from '../api/client'
import { useAuth } from '../stores/auth'
import { mdiClose } from '@mdi/js'

const auth = useAuth()
const route = useRoute()
const { locale, t } = useI18n()
const projects = ref([])
const loading = ref(true)
const error = ref('')
const newName = ref('')
const creating = ref(false)
const createError = ref('')
const deletedNotice = ref(false)
const showCreateModal = ref(false)
const closeIcon = mdiClose

const dashboard = ref(null)
const dashboardLoading = ref(true)
const dashboardError = ref('')

const activityPoints = computed(() => dashboard.value?.activity?.last_35_days || [])
const recentEvents = computed(() => dashboard.value?.activity?.recent_events || [])

const totalContributions = computed(() =>
  activityPoints.value.reduce((sum, item) => sum + (Number(item.count) || 0), 0),
)

const summaryCards = computed(() => {
  const summary = dashboard.value?.summary || {}
  return [
    {
      key: 'projects_total',
      label: t('projects.dashboard.kpis.projectsTotal'),
      value: formatNumber(summary.projects_total),
      hint: t('projects.dashboard.kpis.projectsHint'),
    },
    {
      key: 'analyses_total',
      label: t('projects.dashboard.kpis.analysesTotal'),
      value: formatNumber(summary.analyses_total),
      hint: t('projects.dashboard.kpis.analysesHint'),
    },
    {
      key: 'success_ratio',
      label: t('projects.dashboard.kpis.successRate'),
      value: successRate(summary),
      hint: t('projects.dashboard.kpis.successHint'),
    },
    {
      key: 'analyses_last_30_days',
      label: t('projects.dashboard.kpis.last30Days'),
      value: formatNumber(summary.analyses_last_30_days),
      hint: t('projects.dashboard.kpis.last30Hint'),
    },
  ]
})

const openCreateModal = () => {
  createError.value = ''
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
  createError.value = ''
  newName.value = ''
}

const fetchProjects = async () => {
  loading.value = true
  error.value = ''
  try {
    projects.value = await listProjects()
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const fetchDashboard = async () => {
  dashboardLoading.value = true
  dashboardError.value = ''
  try {
    dashboard.value = await fetchAccountDashboard()
  } catch (err) {
    dashboardError.value = err.message
    dashboard.value = null
  } finally {
    dashboardLoading.value = false
  }
}

const handleCreate = async () => {
  createError.value = ''
  const trimmed = newName.value.trim()
  if (!trimmed) {
    createError.value = t('projects.createNameRequired')
    return
  }
  const duplicate = projects.value.find((p) => p.name === trimmed)
  if (duplicate) {
    createError.value = t('projects.createNameDuplicate')
    return
  }

  creating.value = true
  try {
    const project = await createProject({ name: trimmed })
    projects.value = [project, ...projects.value]
    newName.value = ''
    showCreateModal.value = false
    fetchDashboard()
  } catch (err) {
    if (err.status === 409) {
      createError.value = t('projects.createNameDuplicate')
    } else {
      createError.value = err.message
    }
  } finally {
    creating.value = false
  }
}

const formatDate = (value) => {
  if (!value) return t('common.justNow')
  return new Date(value).toLocaleString(locale.value)
}

const formatDay = (value) => {
  if (!value) return t('common.justNow')
  return new Date(`${value}T00:00:00Z`).toLocaleDateString(locale.value)
}

const formatNumber = (value) => {
  const num = Number(value) || 0
  return new Intl.NumberFormat(locale.value).format(num)
}

const successRate = (summary) => {
  const total = Number(summary.analyses_total) || 0
  if (total === 0) {
    return '0%'
  }
  const completed = Number(summary.completed_total) || 0
  const ratio = Math.round((completed / total) * 100)
  return `${ratio}%`
}

const heatmapCellClass = (level) => {
  switch (Number(level)) {
    case 1:
      return 'bg-primary/25 border border-primary/30'
    case 2:
      return 'bg-primary/45 border border-primary/50'
    case 3:
      return 'bg-primary/70 border border-primary/70'
    case 4:
      return 'bg-primary border border-primary'
    default:
      return 'bg-base border border-border'
  }
}

const eventBadgeClass = (type) => {
  switch (type) {
    case 'analysis_completed':
      return 'badge-success'
    case 'analysis_failed':
      return 'badge-danger'
    case 'analysis_running':
      return 'badge-info'
    case 'analysis_queued':
      return 'badge-warning'
    default:
      return 'badge-neutral'
  }
}

const eventLink = (event) => {
  if (event.analysis_id) {
    return `/projects/${event.project_id}/analyses/${event.analysis_id}`
  }
  return `/projects/${event.project_id}`
}

const eventImageTag = (event) => {
  if (event.image && event.tag) {
    return `${event.image}:${event.tag}`
  }
  if (event.image) {
    return event.image
  }
  return t('common.empty')
}

const eventDescription = (event) => {
  const payload = {
    image: eventImageTag(event),
    status: event.analysis_status || t('common.empty'),
  }

  switch (event.type) {
    case 'analysis_completed':
      return t('projects.dashboard.eventDescriptions.analysis_completed', payload)
    case 'analysis_failed':
      return t('projects.dashboard.eventDescriptions.analysis_failed', payload)
    case 'analysis_running':
      return t('projects.dashboard.eventDescriptions.analysis_running', payload)
    case 'analysis_queued':
      return t('projects.dashboard.eventDescriptions.analysis_queued', payload)
    default:
      return t('projects.dashboard.eventDescriptions.project_created')
  }
}

onMounted(fetchProjects)
onMounted(fetchDashboard)
onMounted(() => {
  deletedNotice.value = route.query.deleted === '1'
})
</script>
