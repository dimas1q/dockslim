<template>
  <div class="space-y-6">
    <RouterLink class="text-sm text-indigo-400 hover:text-indigo-300" to="/projects">
      ← Back to projects
    </RouterLink>

    <div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-4">
      <p v-if="loading" class="text-sm text-slate-400">Loading project...</p>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>
      <div v-else class="space-y-4">
        <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
          <div>
            <h2 class="text-2xl font-semibold">Project settings</h2>
            <p class="text-slate-400 mt-2">
              Update the project name and description for your team.
            </p>
          </div>
          <button
            v-if="isOwner"
            class="inline-flex items-center justify-center rounded-lg border border-red-500/60 px-4 py-2 text-sm text-red-300 hover:bg-red-500/10"
            :disabled="deleting"
            @click="handleDelete"
          >
            {{ deleting ? 'Deleting...' : 'Delete project' }}
          </button>
        </div>
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleUpdateProject">
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Project name</label>
            <input
              v-model="settingsForm.name"
              type="text"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm disabled:opacity-60"
              :disabled="!isOwner"
            />
            <p v-if="settingsErrors.name" class="text-xs text-red-400">
              {{ settingsErrors.name }}
            </p>
          </div>
          <div class="space-y-1 md:col-span-2">
            <label class="text-xs text-slate-400">Description (optional)</label>
            <textarea
              v-model="settingsForm.description"
              rows="3"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm disabled:opacity-60"
              :disabled="!isOwner"
              placeholder="Add context for the team or notes about this project."
            ></textarea>
          </div>
          <div class="md:col-span-2 flex items-center gap-3">
            <button
              v-if="isOwner"
              type="submit"
              class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
              :disabled="savingProject || !settingsDirty"
            >
              {{ savingProject ? 'Saving...' : 'Save settings' }}
            </button>
            <span v-if="settingsSuccess" class="text-xs text-emerald-400">
              {{ settingsSuccess }}
            </span>
          </div>
        </form>
        <p v-if="settingsError" class="text-sm text-red-400">{{ settingsError }}</p>
        <p v-if="!isOwner && project" class="text-xs text-slate-500">
          Only project owners can update settings.
        </p>
        <p v-if="deleteError" class="text-sm text-red-400">{{ deleteError }}</p>
      </div>
    </div>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">Budgets</h3>
          <p class="text-sm text-slate-400 mt-1">
            Set default thresholds and per-image overrides to catch size regressions early.
          </p>
        </div>
        <div class="flex items-center gap-3">
          <button
            v-if="isOwner"
            class="inline-flex items-center justify-center rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400"
            @click="openOverrideModal()"
          >
            Add override
          </button>
        </div>
      </div>

      <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-5 space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="text-sm font-semibold text-slate-200">Project default</h4>
            <p class="text-xs text-slate-400">Applies when no image-specific override exists.</p>
          </div>
          <span v-if="defaultBudgetSuccess" class="text-xs text-emerald-400">{{ defaultBudgetSuccess }}</span>
        </div>
        <div class="grid gap-4 md:grid-cols-3">
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Warn delta (MB)</label>
            <input
              v-model="defaultBudgetForm.warn_delta_mb"
              type="number"
              min="0"
              step="1"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm disabled:opacity-60"
              :disabled="!isOwner"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Fail delta (MB)</label>
            <input
              v-model="defaultBudgetForm.fail_delta_mb"
              type="number"
              min="0"
              step="1"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm disabled:opacity-60"
              :disabled="!isOwner"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Hard limit (MB)</label>
            <input
              v-model="defaultBudgetForm.hard_limit_mb"
              type="number"
              min="0"
              step="1"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm disabled:opacity-60"
              :disabled="!isOwner"
            />
          </div>
        </div>
        <div class="flex items-center gap-3">
          <button
            v-if="isOwner"
            class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
            :disabled="defaultBudgetSaving"
            @click="handleSaveDefaultBudget"
          >
            {{ defaultBudgetSaving ? 'Saving...' : 'Save default' }}
          </button>
          <p v-if="budgetsError" class="text-sm text-red-400">{{ budgetsError }}</p>
          <p v-else-if="budgetsLoading" class="text-sm text-slate-400">Loading budgets...</p>
          <p v-else-if="!isOwner" class="text-xs text-slate-500">Read-only (owner can edit).</p>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <p class="text-sm font-semibold text-slate-200">Per-image overrides</p>
          <p class="text-xs text-slate-500">Exact image match (e.g. company/app)</p>
        </div>
        <p v-if="budgetOverrides.length === 0" class="text-sm text-slate-400">No overrides yet.</p>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="text-xs uppercase text-slate-500">
              <tr>
                <th class="py-2 pr-4">Image</th>
                <th class="py-2 pr-4">Warn Δ</th>
                <th class="py-2 pr-4">Fail Δ</th>
                <th class="py-2 pr-4">Hard limit</th>
                <th class="py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800">
              <tr v-for="budget in budgetOverrides" :key="budget.id" class="text-slate-200">
                <td class="py-3 pr-4 font-mono text-xs">{{ budget.image }}</td>
                <td class="py-3 pr-4 text-slate-300">
                  {{ budget.warn_delta_mb !== null && budget.warn_delta_mb !== undefined ? `${budget.warn_delta_mb} MB` : '—' }}
                </td>
                <td class="py-3 pr-4 text-slate-300">
                  {{ budget.fail_delta_mb !== null && budget.fail_delta_mb !== undefined ? `${budget.fail_delta_mb} MB` : '—' }}
                </td>
                <td class="py-3 pr-4 text-slate-300">
                  {{ budget.hard_limit_mb !== null && budget.hard_limit_mb !== undefined ? `${budget.hard_limit_mb} MB` : '—' }}
                </td>
                <td class="py-3 text-right">
                  <div class="flex items-center justify-end gap-3">
                    <button
                      v-if="isOwner"
                      class="text-xs text-indigo-300 hover:text-indigo-200"
                      type="button"
                      @click="openOverrideModal(budget)"
                    >
                      Edit
                    </button>
                    <button
                      v-if="isOwner"
                      class="text-xs text-red-300 hover:text-red-200"
                      type="button"
                      @click="handleDeleteOverride(budget.id)"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div
        v-if="showOverrideModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 px-4"
      >
        <div class="w-full max-w-lg rounded-2xl border border-slate-800 bg-slate-900 p-6 space-y-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h4 class="text-lg font-semibold text-slate-100">
                {{ editingOverride ? 'Edit override' : 'Add override' }}
              </h4>
              <p class="text-xs text-slate-400">Exact image name match.</p>
            </div>
            <button class="text-slate-400 hover:text-slate-200" type="button" @click="closeOverrideModal">
              ✕
            </button>
          </div>
          <div class="space-y-4">
            <div class="space-y-1">
              <label class="text-xs text-slate-400">Image</label>
              <input
                v-model="overrideForm.image"
                type="text"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
                :disabled="overrideSaving"
              />
            </div>
            <div class="grid gap-3 md:grid-cols-3">
              <div class="space-y-1">
                <label class="text-xs text-slate-400">Warn delta (MB)</label>
              <input
                v-model="overrideForm.warn_delta_mb"
                type="number"
                min="0"
                step="1"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
                :disabled="overrideSaving"
              />
              </div>
              <div class="space-y-1">
                <label class="text-xs text-slate-400">Fail delta (MB)</label>
              <input
                v-model="overrideForm.fail_delta_mb"
                type="number"
                min="0"
                step="1"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
                :disabled="overrideSaving"
              />
              </div>
              <div class="space-y-1">
                <label class="text-xs text-slate-400">Hard limit (MB)</label>
              <input
                v-model="overrideForm.hard_limit_mb"
                type="number"
                min="0"
                step="1"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
                :disabled="overrideSaving"
              />
              </div>
            </div>
            <p v-if="overrideError" class="text-sm text-red-400">{{ overrideError }}</p>
            <div class="flex items-center justify-end gap-3">
              <button
                type="button"
                class="text-sm text-slate-400 hover:text-slate-200"
                :disabled="overrideSaving"
                @click="closeOverrideModal"
              >
                Cancel
              </button>
              <button
                type="button"
                class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
                :disabled="overrideSaving"
                @click="handleSaveOverride"
              >
                {{ overrideSaving ? 'Saving...' : 'Save override' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">Registries</h3>
          <p class="text-sm text-slate-400 mt-1">
            Manage container registries connected to this project.
          </p>
        </div>
        <button
          v-if="isOwner"
          class="inline-flex items-center justify-center rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400"
          @click="toggleForm"
        >
          {{ showForm ? 'Close' : 'Add registry' }}
        </button>
      </div>

      <p v-if="!isOwner && project" class="text-xs text-slate-500">
        Only project owners can manage registries.
      </p>

      <div
        v-if="showForm && isOwner"
        class="rounded-xl border border-slate-800 bg-slate-950/60 p-5 space-y-4"
      >
        <h4 class="text-sm font-semibold text-slate-200">Add a registry</h4>
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleCreateRegistry">
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Name</label>
            <input
              v-model="form.name"
              type="text"
              placeholder="Production registry"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
            <p v-if="fieldErrors.name" class="text-xs text-red-400">{{ fieldErrors.name }}</p>
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Registry URL</label>
            <input
              v-model="form.registry_url"
              type="url"
              placeholder="https://registry.example.com"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
            <p v-if="fieldErrors.registry_url" class="text-xs text-red-400">
              {{ fieldErrors.registry_url }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Username (optional)</label>
            <input
              v-model="form.username"
              type="text"
              placeholder="ci-bot"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Password (optional)</label>
            <input
              v-model="form.password"
              type="password"
              placeholder="••••••••"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
          </div>
          <div class="md:col-span-2 flex items-center gap-3">
            <button
              type="submit"
              class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
              :disabled="creatingRegistry"
            >
              {{ creatingRegistry ? 'Saving...' : 'Save registry' }}
            </button>
            <button
              type="button"
              class="text-sm text-slate-400 hover:text-slate-200"
              :disabled="creatingRegistry"
              @click="resetForm"
            >
              Clear
            </button>
          </div>
        </form>
        <p v-if="createRegistryError" class="text-sm text-red-400">{{ createRegistryError }}</p>
      </div>

      <div>
        <p v-if="registriesLoading" class="text-sm text-slate-400">Loading registries...</p>
        <p v-else-if="registriesError" class="text-sm text-red-400">{{ registriesError }}</p>
        <p v-else-if="registries.length === 0" class="text-sm text-slate-400">
          No registries added yet.
        </p>
        <div v-else class="grid gap-4 md:grid-cols-2">
          <div
            v-for="registry in registries"
            :key="registry.id"
            class="rounded-xl border border-slate-800 bg-slate-950/50 p-4 space-y-3"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-base font-semibold">{{ registry.name }}</p>
                <p class="text-xs text-slate-500 mt-1">{{ registry.registry_url }}</p>
              </div>
              <span class="rounded-full bg-slate-800/70 px-2 py-1 text-xs text-slate-200">
                Generic
              </span>
            </div>
            <p v-if="registry.username" class="text-xs text-slate-400">
              Username: {{ registry.username }}
            </p>
            <div v-if="isOwner" class="pt-1">
              <div class="flex items-center gap-3">
                <button
                  class="text-xs text-indigo-300 hover:text-indigo-200"
                  type="button"
                  @click="openEditRegistry(registry)"
                >
                  Edit
                </button>
                <button
                  class="text-xs text-red-300 hover:text-red-200"
                  :disabled="deletingRegistryId === registry.id"
                  @click="handleDeleteRegistry(registry.id)"
                >
                  {{ deletingRegistryId === registry.id ? 'Deleting...' : 'Delete' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="editingRegistry"
        class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 px-4"
      >
        <div class="w-full max-w-lg rounded-2xl border border-slate-800 bg-slate-900 p-6 space-y-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h4 class="text-lg font-semibold text-slate-100">Edit registry</h4>
              <p class="text-xs text-slate-400">Update the name or registry URL.</p>
            </div>
            <button class="text-slate-400 hover:text-slate-200" type="button" @click="closeEdit">
              ✕
            </button>
          </div>
          <form class="space-y-4" @submit.prevent="handleUpdateRegistry">
            <div class="space-y-1">
              <label class="text-xs text-slate-400">Name</label>
              <input
                v-model="editForm.name"
                type="text"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
              />
              <p v-if="editErrors.name" class="text-xs text-red-400">{{ editErrors.name }}</p>
            </div>
            <div class="space-y-1">
              <label class="text-xs text-slate-400">Registry URL</label>
              <input
                v-model="editForm.registry_url"
                type="url"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
              />
              <p v-if="editErrors.registry_url" class="text-xs text-red-400">
                {{ editErrors.registry_url }}
              </p>
            </div>
            <div class="space-y-1">
              <label class="text-xs text-slate-400">Username</label>
              <input
                v-model="editForm.username"
                type="text"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
                placeholder="ci-bot"
              />
              <p v-if="editErrors.username" class="text-xs text-red-400">
                {{ editErrors.username }}
              </p>
            </div>
            <div class="space-y-1">
              <label class="text-xs text-slate-400">Token</label>
              <input
                v-model="editForm.token"
                type="password"
                class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
                placeholder="••••••••"
              />
              <p class="text-xs text-slate-500">Leave token empty to keep existing credentials.</p>
              <p v-if="editErrors.token" class="text-xs text-red-400">
                {{ editErrors.token }}
              </p>
            </div>
            <p v-if="editRegistryError" class="text-sm text-red-400">{{ editRegistryError }}</p>
            <p v-if="editRegistrySuccess" class="text-sm text-emerald-400">
              {{ editRegistrySuccess }}
            </p>
            <div class="flex items-center justify-end gap-3">
              <button
                type="button"
                class="text-sm text-slate-400 hover:text-slate-200"
                :disabled="savingRegistry"
                @click="closeEdit"
              >
                Cancel
              </button>
              <button
                type="submit"
                class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
                :disabled="savingRegistry"
              >
                {{ savingRegistry ? 'Saving...' : 'Save changes' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </section>

    <section class="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <div class="flex items-center gap-3">
            <h3 class="text-xl font-semibold">Analyses</h3>
            <span v-if="polling" class="text-xs text-slate-400">Updating...</span>
          </div>
          <p class="text-sm text-slate-400 mt-1">
            Track image analysis requests and review their status.
          </p>
        </div>
        <button
          v-if="isOwner"
          class="inline-flex items-center justify-center rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400"
          :disabled="registries.length === 0"
          @click="toggleAnalysisForm"
        >
          {{ showAnalysisForm ? 'Close' : 'New analysis' }}
        </button>
      </div>

      <p v-if="!isOwner && project" class="text-xs text-slate-500">
        Only project owners can create new analyses.
      </p>
      <p v-if="isOwner && registries.length === 0" class="text-xs text-slate-500">
        Create a registry first to run image analyses.
      </p>

      <div
        v-if="showAnalysisForm && isOwner"
        class="rounded-xl border border-slate-800 bg-slate-950/60 p-5 space-y-4"
      >
        <h4 class="text-sm font-semibold text-slate-200">Request a new analysis</h4>
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleCreateAnalysis">
          <div class="space-y-1 md:col-span-2">
            <label class="text-xs text-slate-400">Registry</label>
            <select
              v-model="analysisForm.registry_id"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            >
              <option disabled value="">Select a registry</option>
              <option v-for="registry in registries" :key="registry.id" :value="registry.id">
                {{ registry.name }} · {{ registry.registry_url }}
              </option>
            </select>
            <p v-if="analysisErrors.registry_id" class="text-xs text-red-400">
              {{ analysisErrors.registry_id }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Image</label>
            <input
              v-model="analysisForm.image"
              type="text"
              placeholder="repo/name"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
            <p v-if="analysisErrors.image" class="text-xs text-red-400">
              {{ analysisErrors.image }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs text-slate-400">Tag</label>
            <input
              v-model="analysisForm.tag"
              type="text"
              placeholder="latest"
              class="w-full rounded-lg bg-slate-950 border border-slate-800 px-3 py-2 text-sm"
            />
          </div>
          <div class="md:col-span-2 flex items-center gap-3">
            <button
              type="submit"
              class="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-60"
              :disabled="creatingAnalysis"
            >
              {{ creatingAnalysis ? 'Submitting...' : 'Start analysis' }}
            </button>
            <button
              type="button"
              class="text-sm text-slate-400 hover:text-slate-200"
              :disabled="creatingAnalysis"
              @click="resetAnalysisForm"
            >
              Clear
            </button>
          </div>
        </form>
        <p v-if="createAnalysisError" class="text-sm text-red-400">{{ createAnalysisError }}</p>
      </div>

      <div>
        <p v-if="analysesLoading" class="text-sm text-slate-400">Loading analyses...</p>
        <p v-else-if="analysesError" class="text-sm text-red-400">{{ analysesError }}</p>
        <p v-else-if="analyses.length === 0" class="text-sm text-slate-400">
          No analyses yet. Kick off your first image inspection.
        </p>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="text-xs uppercase text-slate-500">
              <tr>
                <th class="py-2 pr-4">Image</th>
                <th class="py-2 pr-4">Status</th>
                <th class="py-2 pr-4">Created</th>
                <th class="py-2">Total size</th>
                <th class="py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800">
              <tr v-for="analysis in analyses" :key="analysis.id" class="text-slate-200">
                <td class="py-3 pr-4">
                  <RouterLink
                    class="text-indigo-400 hover:text-indigo-300"
                    :to="`/projects/${project?.id}/analyses/${analysis.id}`"
                  >
                    {{ analysis.image }}:{{ analysis.tag }}
                  </RouterLink>
                </td>
                <td class="py-3 pr-4">
                  <span
                    class="rounded-full px-2 py-1 text-xs font-semibold"
                    :class="statusBadgeClass(analysis.status)"
                  >
                    {{ analysis.status }}
                  </span>
                </td>
                <td class="py-3 pr-4 text-slate-400">
                  {{ formatDate(analysis.created_at) }}
                </td>
                <td class="py-3 text-slate-400">
                  {{ analysis.total_size_bytes ? formatBytes(analysis.total_size_bytes) : '—' }}
                </td>
                <td class="py-3 text-right">
                  <div class="flex items-center justify-end gap-3">
                    <RouterLink
                      v-if="getPreviousCompletedAnalysis(analysis)"
                      class="text-xs text-indigo-400 hover:text-indigo-300"
                      :to="`/projects/${project?.id}/analyses/compare?from=${getPreviousCompletedAnalysis(analysis)?.id}&to=${analysis.id}`"
                    >
                      Compare
                    </RouterLink>
                    <button
                      v-if="isOwner"
                      class="text-xs text-red-300 hover:text-red-200 disabled:opacity-60"
                      :disabled="deletingAnalysisId === analysis.id"
                      @click="handleDeleteAnalysis(analysis)"
                    >
                      {{ deletingAnalysisId === analysis.id ? 'Deleting...' : 'Delete' }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import {
  createAnalysis,
  createRegistry,
  deleteAnalysis,
  deleteProject,
  deleteRegistry,
  getBudgets,
  getProject,
  listAnalyses,
  listRegistries,
  updateProject,
  updateRegistry,
  upsertDefaultBudget,
  createBudgetOverride,
  updateBudgetOverride,
  deleteBudgetOverride,
} from '../api/client'

const route = useRoute()
const router = useRouter()
const project = ref(null)
const loading = ref(true)
const error = ref('')
const deleting = ref(false)
const deleteError = ref('')
const registries = ref([])
const registriesLoading = ref(false)
const registriesError = ref('')
const budgetsLoading = ref(false)
const budgetsError = ref('')
const budgetsDefault = ref(null)
const budgetOverrides = ref([])
const defaultBudgetForm = ref({ warn_delta_mb: '', fail_delta_mb: '', hard_limit_mb: '' })
const defaultBudgetSaving = ref(false)
const defaultBudgetSuccess = ref('')
const overrideForm = ref({ image: '', warn_delta_mb: '', fail_delta_mb: '', hard_limit_mb: '' })
const showOverrideModal = ref(false)
const editingOverride = ref(null)
const overrideSaving = ref(false)
const overrideError = ref('')
const savingProject = ref(false)
const settingsError = ref('')
const settingsSuccess = ref('')
const settingsErrors = ref({})
const analyses = ref([])
const analysesLoading = ref(false)
const analysesError = ref('')
const polling = ref(false)
const showForm = ref(false)
const creatingRegistry = ref(false)
const createRegistryError = ref('')
const deletingRegistryId = ref(null)
const deletingAnalysisId = ref(null)
const editingRegistry = ref(null)
const editForm = ref({ name: '', registry_url: '', username: '', token: '' })
const editErrors = ref({})
const savingRegistry = ref(false)
const editRegistryError = ref('')
const editRegistrySuccess = ref('')
const fieldErrors = ref({})
const showAnalysisForm = ref(false)
const creatingAnalysis = ref(false)
const createAnalysisError = ref('')
const analysisErrors = ref({})
const settingsForm = ref({ name: '', description: '' })
let pollTimer = null

const form = ref({
  name: '',
  registry_url: '',
  username: '',
  password: '',
})

const analysisForm = ref({
  registry_id: '',
  image: '',
  tag: 'latest',
})

const isOwner = computed(() => project.value?.role === 'owner')
const hasActiveAnalyses = computed(() =>
  analyses.value.some((analysis) => ['queued', 'running'].includes(analysis.status)),
)
const settingsDirty = computed(() => {
  if (!project.value) {
    return false
  }
  const nameValue = settingsForm.value.name.trim()
  const descriptionValue = settingsForm.value.description.trim()
  const currentDescription = project.value.description || ''
  return nameValue !== project.value.name || descriptionValue !== currentDescription
})

const syncSettingsForm = () => {
  if (!project.value) {
    settingsForm.value = { name: '', description: '' }
    return
  }
  settingsForm.value = {
    name: project.value.name || '',
    description: project.value.description || '',
  }
}

const fetchProject = async () => {
  loading.value = true
  error.value = ''
  try {
    project.value = await getProject(route.params.id)
    syncSettingsForm()
    await fetchBudgets()
    await fetchRegistries()
    await fetchAnalyses()
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

onMounted(fetchProject)
onBeforeUnmount(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})

const fetchRegistries = async () => {
  registriesLoading.value = true
  registriesError.value = ''
  try {
    registries.value = await listRegistries(route.params.id)
  } catch (err) {
    registriesError.value = err.message
  } finally {
    registriesLoading.value = false
  }
}

const syncDefaultBudgetForm = () => {
  if (budgetsDefault.value) {
    defaultBudgetForm.value = {
      warn_delta_mb: bytesToMB(budgetsDefault.value.warn_delta_bytes),
      fail_delta_mb: bytesToMB(budgetsDefault.value.fail_delta_bytes),
      hard_limit_mb: bytesToMB(budgetsDefault.value.hard_limit_bytes),
    }
  } else {
    defaultBudgetForm.value = { warn_delta_mb: '', fail_delta_mb: '', hard_limit_mb: '' }
  }
}

const fetchBudgets = async () => {
  budgetsLoading.value = true
  budgetsError.value = ''
  try {
    const data = await getBudgets(route.params.id)
    budgetsDefault.value = data?.default || null
    budgetOverrides.value = data?.overrides || []
    syncDefaultBudgetForm()
  } catch (err) {
    budgetsError.value = err.message
  } finally {
    budgetsLoading.value = false
  }
}

const fetchAnalyses = async ({ silent = false } = {}) => {
  if (!silent) {
    analysesLoading.value = true
  }
  analysesError.value = ''
  try {
    analyses.value = await listAnalyses(route.params.id)
  } catch (err) {
    analysesError.value = err.message
  } finally {
    if (!silent) {
      analysesLoading.value = false
    }
  }
}

const startPolling = () => {
  if (pollTimer) {
    return
  }
  polling.value = true
  pollTimer = setInterval(() => {
    fetchAnalyses({ silent: true })
  }, 3000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  polling.value = false
}

watch(
  hasActiveAnalyses,
  (active) => {
    if (active) {
      startPolling()
      return
    }
    stopPolling()
  },
  { immediate: true },
)

const toggleForm = () => {
  showForm.value = !showForm.value
  if (!showForm.value) {
    resetForm()
  }
}

const resetForm = () => {
  form.value = {
    name: '',
    registry_url: '',
    username: '',
    password: '',
  }
  fieldErrors.value = {}
  createRegistryError.value = ''
}

const toggleAnalysisForm = () => {
  showAnalysisForm.value = !showAnalysisForm.value
  if (!showAnalysisForm.value) {
    resetAnalysisForm()
  }
}

const resetAnalysisForm = () => {
  analysisForm.value = {
    registry_id: '',
    image: '',
    tag: 'latest',
  }
  analysisErrors.value = {}
  createAnalysisError.value = ''
}

const handleCreateRegistry = async () => {
  fieldErrors.value = {}
  createRegistryError.value = ''

  const nameValue = form.value.name.trim()
  const urlValue = form.value.registry_url.trim()

  if (!nameValue) {
    fieldErrors.value.name = 'Name is required.'
  }
  if (!urlValue) {
    fieldErrors.value.registry_url = 'Registry URL is required.'
  }

  const duplicate = registries.value.find((r) => r.name === nameValue)
  if (!fieldErrors.value.name && duplicate) {
    fieldErrors.value.name = 'Registry with this name already exists.'
  }

  if (Object.keys(fieldErrors.value).length > 0) {
    return
  }

  creatingRegistry.value = true
  try {
    await createRegistry(route.params.id, {
      name: nameValue,
      type: 'generic',
      registry_url: urlValue,
      username: form.value.username,
      password: form.value.password,
    })
    form.value.password = ''
    showForm.value = false
    resetForm()
    await fetchRegistries()
  } catch (err) {
    if (err.status === 409) {
      createRegistryError.value = 'Registry with this name already exists.'
    } else {
      createRegistryError.value = err.message
    }
  } finally {
    creatingRegistry.value = false
  }
}

const buildBudgetPayload = (form) => {
  const payload = {}
  let invalid = false
  const mapField = (key) => {
    const value = form[key]
    if (value === '' || value === null || value === undefined) {
      payload[key] = null
      return
    }
    const numeric = Math.trunc(Number(value))
    if (!Number.isFinite(numeric)) {
      invalid = true
      return
    }
    payload[key] = numeric
  }
  mapField('warn_delta_mb')
  mapField('fail_delta_mb')
  mapField('hard_limit_mb')
  return { payload, invalid }
}

const handleSaveDefaultBudget = async () => {
  if (!isOwner.value) return
  defaultBudgetSaving.value = true
  defaultBudgetSuccess.value = ''
  budgetsError.value = ''
  try {
    const { payload, invalid } = buildBudgetPayload(defaultBudgetForm.value)
    if (invalid) {
      budgetsError.value = 'Enter whole numbers for MB values.'
      defaultBudgetSaving.value = false
      return
    }
    const saved = await upsertDefaultBudget(route.params.id, payload)
    budgetsDefault.value = saved
    syncDefaultBudgetForm()
    defaultBudgetSuccess.value = 'Budget saved.'
  } catch (err) {
    budgetsError.value = err.message
  } finally {
    defaultBudgetSaving.value = false
  }
}

const resetOverrideForm = () => {
  overrideForm.value = { image: '', warn_delta_mb: '', fail_delta_mb: '', hard_limit_mb: '' }
  overrideError.value = ''
  editingOverride.value = null
}

const openOverrideModal = (budget = null) => {
  if (budget) {
    editingOverride.value = budget
    overrideForm.value = {
      image: budget.image || '',
      warn_delta_mb: bytesToMB(budget.warn_delta_bytes),
      fail_delta_mb: bytesToMB(budget.fail_delta_bytes),
      hard_limit_mb: bytesToMB(budget.hard_limit_bytes),
    }
  } else {
    resetOverrideForm()
  }
  overrideError.value = ''
  showOverrideModal.value = true
}

const closeOverrideModal = () => {
  showOverrideModal.value = false
  resetOverrideForm()
}

const handleSaveOverride = async () => {
  if (!isOwner.value) return
  overrideSaving.value = true
  overrideError.value = ''
  try {
    const imageValue = overrideForm.value.image.trim()
    if (!imageValue) {
      overrideError.value = 'Image is required.'
      overrideSaving.value = false
      return
    }
    const duplicate = budgetOverrides.value.find(
      (item) => item.image === imageValue && (!editingOverride.value || item.id !== editingOverride.value.id),
    )
    if (duplicate) {
      overrideError.value = 'Override for this image already exists.'
      overrideSaving.value = false
      return
    }
    const { payload, invalid } = buildBudgetPayload(overrideForm.value)
    if (invalid) {
      overrideError.value = 'Enter whole numbers for MB values.'
      overrideSaving.value = false
      return
    }
    payload.image = imageValue
    let saved
    if (editingOverride.value) {
      saved = await updateBudgetOverride(route.params.id, editingOverride.value.id, payload)
      budgetOverrides.value = budgetOverrides.value.map((item) =>
        item.id === saved.id ? saved : item,
      )
    } else {
      saved = await createBudgetOverride(route.params.id, payload)
      budgetOverrides.value = [...budgetOverrides.value, saved]
    }
    closeOverrideModal()
  } catch (err) {
    if (err.status === 409) {
      overrideError.value = 'Override for this image already exists.'
    } else {
      overrideError.value = err.message
    }
  } finally {
    overrideSaving.value = false
  }
}

const handleDeleteOverride = async (budgetId) => {
  if (!isOwner.value) return
  const confirmed = window.confirm('Delete this override?')
  if (!confirmed) return
  try {
    await deleteBudgetOverride(route.params.id, budgetId)
    budgetOverrides.value = budgetOverrides.value.filter((b) => b.id !== budgetId)
  } catch (err) {
    budgetsError.value = err.message
  }
}

const handleCreateAnalysis = async () => {
  analysisErrors.value = {}
  createAnalysisError.value = ''

  if (!analysisForm.value.registry_id) {
    analysisErrors.value.registry_id = 'Registry is required.'
  }
  if (!analysisForm.value.image.trim()) {
    analysisErrors.value.image = 'Image is required.'
  }

  if (Object.keys(analysisErrors.value).length > 0) {
    return
  }

  creatingAnalysis.value = true
  try {
    await createAnalysis(route.params.id, {
      registry_id: analysisForm.value.registry_id,
      image: analysisForm.value.image,
      tag: analysisForm.value.tag,
    })
    showAnalysisForm.value = false
    resetAnalysisForm()
    await fetchAnalyses()
  } catch (err) {
    createAnalysisError.value = err.message
  } finally {
    creatingAnalysis.value = false
  }
}

const handleDeleteRegistry = async (registryId) => {
  deletingRegistryId.value = registryId
  try {
    await deleteRegistry(route.params.id, registryId)
    await fetchRegistries()
  } catch (err) {
    registriesError.value = err.message
  } finally {
    deletingRegistryId.value = null
  }
}

const handleDeleteAnalysis = async (analysis) => {
  if (!analysis?.id) {
    return
  }
  const confirmed = window.confirm('Delete this analysis? This cannot be undone.')
  if (!confirmed) {
    return
  }

  deletingAnalysisId.value = analysis.id
  try {
    await deleteAnalysis(route.params.id, analysis.id)
    await fetchAnalyses()
  } catch (err) {
    analysesError.value = err.message
  } finally {
    deletingAnalysisId.value = null
  }
}

const formatDate = (value) => {
  if (!value) return 'just now'
  return new Date(value).toLocaleString()
}

const formatBytes = (value) => {
  if (!value && value !== 0) return '—'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = Number(value)
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

const bytesToMB = (value) => {
  if (value === null || value === undefined) return ''
  const mb = Math.round(Number(value) / (1024 * 1024))
  return Number.isFinite(mb) ? mb : ''
}

const statusBadgeClass = (status) => {
  switch (status) {
    case 'completed':
      return 'bg-emerald-500/20 text-emerald-200'
    case 'running':
      return 'bg-sky-500/20 text-sky-200'
    case 'failed':
      return 'bg-rose-500/20 text-rose-200'
    default:
      return 'bg-amber-500/20 text-amber-200'
  }
}

const getPreviousCompletedAnalysis = (analysis) => {
  if (!analysis || analysis.status !== 'completed') {
    return null
  }
  const currentCreatedAt = new Date(analysis.created_at).getTime()
  if (!Number.isFinite(currentCreatedAt)) {
    return null
  }
  return analyses.value.reduce((latest, item) => {
    if (item.id === analysis.id) {
      return latest
    }
    if (item.image !== analysis.image || item.status !== 'completed') {
      return latest
    }
    const itemCreatedAt = new Date(item.created_at).getTime()
    if (!Number.isFinite(itemCreatedAt) || itemCreatedAt >= currentCreatedAt) {
      return latest
    }
    if (!latest) {
      return item
    }
    const latestCreatedAt = new Date(latest.created_at).getTime()
    if (!Number.isFinite(latestCreatedAt) || itemCreatedAt > latestCreatedAt) {
      return item
    }
    return latest
  }, null)
}

const openEditRegistry = (registry) => {
  editingRegistry.value = registry
  editForm.value = {
    name: registry.name || '',
    registry_url: registry.registry_url || '',
    username: registry.username || '',
    token: '',
  }
  editErrors.value = {}
  editRegistryError.value = ''
  editRegistrySuccess.value = ''
}

const closeEdit = () => {
  editingRegistry.value = null
  editForm.value = { name: '', registry_url: '', username: '', token: '' }
  editErrors.value = {}
  editRegistryError.value = ''
  editRegistrySuccess.value = ''
}

const handleUpdateRegistry = async () => {
  if (!editingRegistry.value) {
    return
  }

  editErrors.value = {}
  editRegistryError.value = ''
  editRegistrySuccess.value = ''

  const nameValue = editForm.value.name.trim()
  const urlValue = editForm.value.registry_url.trim()
  const usernameValue = editForm.value.username.trim()
  const tokenValue = editForm.value.token.trim()
  const hasNameChange = nameValue !== editingRegistry.value.name
  const hasURLChange = urlValue !== editingRegistry.value.registry_url
  const currentUsername = editingRegistry.value.username || ''
  const hasUsernameChange = usernameValue !== currentUsername
  const wantsCredentialUpdate = tokenValue !== '' || hasUsernameChange

  if (hasNameChange && !nameValue) {
    editErrors.value.name = 'Name is required.'
  }
  if (hasURLChange && !urlValue) {
    editErrors.value.registry_url = 'Registry URL is required.'
  }
  if (wantsCredentialUpdate && !usernameValue) {
    editErrors.value.username = 'Username is required to update credentials.'
  }
  if (wantsCredentialUpdate && !tokenValue) {
    editErrors.value.token = 'Token is required to update credentials.'
  }
  if (!hasNameChange && !hasURLChange) {
    if (!wantsCredentialUpdate) {
      editRegistryError.value = 'Make a change before saving.'
      return
    }
  }
  if (Object.keys(editErrors.value).length > 0) {
    return
  }

  const payload = {}
  if (hasNameChange) {
    payload.name = nameValue
  }
  if (hasURLChange) {
    payload.registry_url = urlValue
  }
  if (wantsCredentialUpdate) {
    payload.username = usernameValue
    payload.token = tokenValue
  }

  savingRegistry.value = true
  try {
    const updated = await updateRegistry(route.params.id, editingRegistry.value.id, payload)
    editingRegistry.value = updated
    editForm.value = {
      name: updated.name || '',
      registry_url: updated.registry_url || '',
      username: updated.username || '',
      token: '',
    }
    editRegistrySuccess.value = 'Saved.'
    await fetchRegistries()
  } catch (err) {
    if (err.status === 409) {
      editRegistryError.value = 'Registry with this name already exists.'
    } else {
      editRegistryError.value = err.message
    }
  } finally {
    savingRegistry.value = false
  }
}

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

const handleUpdateProject = async () => {
  if (!project.value || !isOwner.value) {
    return
  }
  settingsErrors.value = {}
  settingsError.value = ''
  settingsSuccess.value = ''

  const nameValue = settingsForm.value.name.trim()
  const descriptionValue = settingsForm.value.description.trim()

  if (!nameValue) {
    settingsErrors.value.name = 'Project name is required.'
  }

  if (Object.keys(settingsErrors.value).length > 0) {
    return
  }

  const payload = {}
  if (nameValue !== project.value.name) {
    payload.name = nameValue
  }
  if (descriptionValue !== (project.value.description || '')) {
    payload.description = descriptionValue
  }

  if (Object.keys(payload).length === 0) {
    settingsError.value = 'Make a change before saving.'
    return
  }

  savingProject.value = true
  try {
    const updated = await updateProject(project.value.id, payload)
    project.value = { ...project.value, ...updated }
    syncSettingsForm()
    settingsSuccess.value = 'Saved.'
  } catch (err) {
    settingsError.value = err.message
  } finally {
    savingProject.value = false
  }
}
</script>
