<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" to="/projects">
      ← Back to projects
    </RouterLink>

    <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-4">
      <p v-if="loading" class="text-sm text-slate-400">Loading project...</p>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>
      <div v-else>
        <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
          <div>
            <h2 class="text-2xl font-semibold">{{ project?.name }}</h2>
            <p class="text-slate-400 mt-2">Project details coming soon.</p>
          </div>
          <button
            class="inline-flex items-center justify-center rounded-lg border border-red-500/60 px-4 py-2 text-sm text-red-300 hover:bg-red-500/10"
            :disabled="deleting"
            @click="handleDelete"
          >
            {{ deleting ? 'Deleting...' : 'Delete project' }}
          </button>
        </div>
        <p v-if="deleteError" class="text-sm text-red-400">{{ deleteError }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { deleteProject, getProject } from '../api/client'

const route = useRoute()
const router = useRouter()
const project = ref(null)
const loading = ref(true)
const error = ref('')
const deleting = ref(false)
const deleteError = ref('')

const fetchProject = async () => {
  loading.value = true
  error.value = ''
  try {
    project.value = await getProject(route.params.id)
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

onMounted(fetchProject)

const handleDelete = async () => {
  if (!project.value) {
    return
  }
  deleteError.value = ''
  const confirmed = window.confirm('Delete this project? This cannot be undone.')
  if (!confirmed) {
    return
  }

  deleting.value = true
  try {
    await deleteProject(project.value.id)
    router.push({ path: '/projects', query: { deleted: '1' } })
  } catch (err) {
    deleteError.value = err.message
  } finally {
    deleting.value = false
  }
}
</script>
