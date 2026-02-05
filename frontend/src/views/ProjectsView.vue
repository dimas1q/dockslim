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
          <button class="btn btn-ghost text-sm" @click="fetchProjects">
            {{ t('common.refresh') }}
          </button>
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
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { createProject, listProjects } from '../api/client'
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

onMounted(fetchProjects)
onMounted(() => {
  deletedNotice.value = route.query.deleted === '1'
})
</script>
