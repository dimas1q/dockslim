<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" to="/projects">
      ← Back to projects
    </RouterLink>

    <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
      <p v-if="loading" class="text-sm text-slate-400">Loading project...</p>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>
      <div v-else>
        <h2 class="text-2xl font-semibold">{{ project?.name }}</h2>
        <p class="text-slate-400 mt-2">Project details coming soon.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { getProject } from '../api/client'

const route = useRoute()
const project = ref(null)
const loading = ref(true)
const error = ref('')

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
</script>
