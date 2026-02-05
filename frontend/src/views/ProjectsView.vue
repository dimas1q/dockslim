<template>
  <div class="space-y-8">
    <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
      <p class="text-sm text-slate-400">{{ t('projects.loggedInAs') }}</p>
      <p class="text-lg font-semibold">{{ auth.user?.email || t('projects.loadingAccount') }}</p>
    </div>

    <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
      <h2 class="text-xl font-semibold mb-4">{{ t('projects.createTitle') }}</h2>
      <form class="flex flex-col gap-3 md:flex-row" @submit.prevent="handleCreate">
        <input
          v-model="newName"
          type="text"
          :placeholder="t('projects.createPlaceholder')"
          class="flex-1 rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
        />
        <button
          type="submit"
          class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400"
          :disabled="creating"
        >
          {{ creating ? t('projects.createButtonLoading') : t('projects.createButton') }}
        </button>
      </form>
      <p v-if="createError" class="text-sm text-red-400 mt-3">{{ createError }}</p>
    </div>

    <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold">{{ t('projects.listTitle') }}</h2>
        <button class="text-sm text-slate-300 hover:text-white" @click="fetchProjects">
          {{ t('common.refresh') }}
        </button>
      </div>

      <p v-if="deletedNotice" class="text-sm text-emerald-400 mb-3">
        {{ t('projects.deletedNotice') }}
      </p>
      <p v-if="loading" class="text-sm text-slate-400">{{ t('projects.loadingProjects') }}</p>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>
      <p v-else-if="projects.length === 0" class="text-sm text-slate-400">
        {{ t('projects.empty') }}
      </p>

      <ul v-else class="space-y-3">
        <li
          v-for="project in projects"
          :key="project.id"
          class="flex items-center justify-between rounded-lg border border-slate-800 bg-slate-950/40 px-4 py-3"
        >
          <div>
            <p class="font-medium">{{ project.name }}</p>
            <p class="text-xs text-slate-500">
              {{ t('projects.createdAt', { date: formatDate(project.created_at) }) }}
            </p>
          </div>
          <RouterLink
            class="text-sm text-indigo-400 hover:text-indigo-300"
            :to="`/projects/${project.id}`"
          >
            {{ t('projects.viewDetails') }}
          </RouterLink>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { createProject, listProjects } from '../api/client'
import { useAuth } from '../stores/auth'

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
