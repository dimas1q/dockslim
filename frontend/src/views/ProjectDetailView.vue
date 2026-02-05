<template>
  <div class="space-y-10">
    <RouterLink class="link-subtle text-sm" to="/projects">
      {{ t('nav.backToProjects') }}
    </RouterLink>

    <div class="panel p-6 space-y-4 ds-reveal">
      <div v-if="loading" class="space-y-3">
        <div class="h-5 w-40 rounded skeleton"></div>
        <div class="h-4 w-72 rounded skeleton"></div>
        <div class="h-24 rounded-xl skeleton"></div>
      </div>
      <p v-else-if="error" class="text-sm text-danger">{{ error }}</p>
      <div v-else class="space-y-4">
        <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
          <div>
            <h2 class="text-2xl font-semibold text-ink">{{ t('projectDetail.settingsTitle') }}</h2>
            <p class="text-sm text-muted mt-2">{{ t('projectDetail.settingsSubtitle') }}</p>
          </div>
          <button
            v-if="isOwner"
            class="btn btn-danger"
            :disabled="deleting"
            @click="handleDelete"
          >
            {{ deleting ? t('common.deleting') : t('projectDetail.deleteProject') }}
          </button>
        </div>
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleUpdateProject">
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('projectDetail.nameLabel') }}</label>
            <input
              v-model="settingsForm.name"
              type="text"
              class="input"
              :disabled="!isOwner"
            />
            <p v-if="settingsErrors.name" class="text-xs text-danger">
              {{ settingsErrors.name }}
            </p>
          </div>
          <div class="space-y-1 md:col-span-2">
            <label class="text-xs font-medium text-subtle">{{ t('projectDetail.descriptionLabel') }}</label>
            <textarea
              v-model="settingsForm.description"
              rows="3"
              class="textarea"
              :disabled="!isOwner"
              :placeholder="t('projectDetail.descriptionPlaceholder')"
            ></textarea>
          </div>
          <div class="md:col-span-2 flex items-center gap-3">
            <button
              v-if="isOwner"
              type="submit"
              class="btn btn-primary"
              :disabled="savingProject || !settingsDirty"
            >
              {{ savingProject ? t('common.saving') : t('common.saveSettings') }}
            </button>
            <span v-if="settingsSuccess" class="text-xs text-success">
              {{ settingsSuccess }}
            </span>
          </div>
        </form>
        <p v-if="settingsError" class="text-sm text-danger">{{ settingsError }}</p>
        <p v-if="!isOwner && project" class="text-xs text-subtle">
          {{ t('projectDetail.ownerOnlySettings') }}
        </p>
        <p v-if="deleteError" class="text-sm text-danger">{{ deleteError }}</p>
      </div>
    </div>

    <section class="panel p-6 space-y-6 ds-reveal">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">{{ t('projectDetail.budgetsTitle') }}</h3>
          <p class="text-sm text-muted mt-1">{{ t('projectDetail.budgetsSubtitle') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <button
            v-if="isOwner"
            class="btn btn-primary"
            @click="openOverrideModal()"
          >
            {{ t('projectDetail.addOverride') }}
          </button>
        </div>
      </div>

      <div class="surface p-5 space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="text-sm font-semibold text-ink">{{ t('projectDetail.projectDefault') }}</h4>
            <p class="text-xs text-muted">{{ t('projectDetail.projectDefaultSubtitle') }}</p>
          </div>
          <span v-if="defaultBudgetSuccess" class="text-xs text-success">{{ defaultBudgetSuccess }}</span>
        </div>
        <div class="grid gap-4 md:grid-cols-3">
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('projectDetail.warnDelta') }}</label>
            <input
              v-model="defaultBudgetForm.warn_delta_mb"
              type="number"
              min="0"
              step="1"
              class="input"
              :disabled="!isOwner"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('projectDetail.failDelta') }}</label>
            <input
              v-model="defaultBudgetForm.fail_delta_mb"
              type="number"
              min="0"
              step="1"
              class="input"
              :disabled="!isOwner"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('projectDetail.hardLimit') }}</label>
            <input
              v-model="defaultBudgetForm.hard_limit_mb"
              type="number"
              min="0"
              step="1"
              class="input"
              :disabled="!isOwner"
            />
          </div>
        </div>
        <div class="flex items-center gap-3">
          <button
            v-if="isOwner"
            class="btn btn-primary"
            :disabled="defaultBudgetSaving"
            @click="handleSaveDefaultBudget"
          >
            {{ defaultBudgetSaving ? t('common.saving') : t('common.saveDefault') }}
          </button>
          <p v-if="budgetsError" class="text-sm text-danger">{{ budgetsError }}</p>
          <p v-else-if="budgetsLoading" class="text-sm text-muted">{{ t('projectDetail.budgetsLoading') }}</p>
          <p v-else-if="!isOwner" class="text-xs text-subtle">{{ t('projectDetail.readOnlyOwner') }}</p>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <p class="text-sm font-semibold text-ink">{{ t('projectDetail.overridesTitle') }}</p>
          <p class="text-xs text-subtle">{{ t('projectDetail.overridesHint') }}</p>
        </div>
        <p v-if="budgetOverrides.length === 0" class="text-sm text-muted">
          {{ t('projectDetail.noOverrides') }}
        </p>
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th class="py-2 pr-4">{{ t('common.image') }}</th>
                <th class="py-2 pr-4">{{ t('projectDetail.warnDelta') }}</th>
                <th class="py-2 pr-4">{{ t('projectDetail.failDelta') }}</th>
                <th class="py-2 pr-4">{{ t('projectDetail.hardLimit') }}</th>
                <th class="py-2 text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="budget in budgetOverrides" :key="budget.id">
                <td class="py-3 pr-4 font-mono text-xs">{{ budget.image }}</td>
                <td class="py-3 pr-4 text-muted">
                  {{
                    budget.warn_delta_mb !== null && budget.warn_delta_mb !== undefined
                      ? `${budget.warn_delta_mb} ${t('units.mb')}`
                      : t('common.empty')
                  }}
                </td>
                <td class="py-3 pr-4 text-muted">
                  {{
                    budget.fail_delta_mb !== null && budget.fail_delta_mb !== undefined
                      ? `${budget.fail_delta_mb} ${t('units.mb')}`
                      : t('common.empty')
                  }}
                </td>
                <td class="py-3 pr-4 text-muted">
                  {{
                    budget.hard_limit_mb !== null && budget.hard_limit_mb !== undefined
                      ? `${budget.hard_limit_mb} ${t('units.mb')}`
                      : t('common.empty')
                  }}
                </td>
                <td class="py-3 text-right">
                  <div class="flex items-center justify-end gap-3">
                    <button
                      v-if="isOwner"
                      class="text-xs text-primary hover:text-primary-strong"
                      type="button"
                      @click="openOverrideModal(budget)"
                    >
                      {{ t('common.edit') }}
                    </button>
                    <button
                      v-if="isOwner"
                      class="text-xs text-danger hover:text-danger/80"
                      type="button"
                      @click="handleDeleteOverride(budget.id)"
                    >
                      {{ t('common.delete') }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <Transition name="modal-fade">
        <div
          v-if="showOverrideModal"
          class="fixed inset-0 z-50 flex items-center justify-center bg-base/80 backdrop-blur-sm px-4"
        >
          <Transition name="modal-panel">
            <div class="panel w-full max-w-lg p-6 space-y-4">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h4 class="text-lg font-semibold text-ink">
                    {{ editingOverride ? t('projectDetail.editOverrideTitle') : t('projectDetail.addOverrideTitle') }}
                  </h4>
                  <p class="text-xs text-muted">{{ t('projectDetail.overrideHint') }}</p>
                </div>
              <button class="btn btn-ghost btn-icon" type="button" @click="closeOverrideModal" aria-label="Close">
                <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <path :d="closeIcon" />
                </svg>
              </button>
              </div>
              <div class="space-y-4">
                <div class="space-y-1">
                  <label class="text-xs font-medium text-subtle">{{ t('common.image') }}</label>
                  <input
                    v-model="overrideForm.image"
                    type="text"
                    class="input"
                    :disabled="overrideSaving"
                  />
                </div>
                <div class="grid gap-3 md:grid-cols-3">
                  <div class="space-y-1">
                    <label class="text-xs font-medium text-subtle">{{ t('projectDetail.warnDelta') }}</label>
                    <input
                      v-model="overrideForm.warn_delta_mb"
                      type="number"
                      min="0"
                      step="1"
                      class="input"
                      :disabled="overrideSaving"
                    />
                  </div>
                  <div class="space-y-1">
                    <label class="text-xs font-medium text-subtle">{{ t('projectDetail.failDelta') }}</label>
                    <input
                      v-model="overrideForm.fail_delta_mb"
                      type="number"
                      min="0"
                      step="1"
                      class="input"
                      :disabled="overrideSaving"
                    />
                  </div>
                  <div class="space-y-1">
                    <label class="text-xs font-medium text-subtle">{{ t('projectDetail.hardLimit') }}</label>
                    <input
                      v-model="overrideForm.hard_limit_mb"
                      type="number"
                      min="0"
                      step="1"
                      class="input"
                      :disabled="overrideSaving"
                    />
                  </div>
                </div>
                <p v-if="overrideError" class="text-sm text-danger">{{ overrideError }}</p>
                <div class="flex items-center justify-end gap-3">
                  <button type="button" class="btn btn-ghost" :disabled="overrideSaving" @click="closeOverrideModal">
                    {{ t('common.cancel') }}
                  </button>
                  <button type="button" class="btn btn-primary" :disabled="overrideSaving" @click="handleSaveOverride">
                    {{ overrideSaving ? t('common.saving') : t('common.saveOverride') }}
                  </button>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </section>

    <section class="panel p-6 space-y-6 ds-reveal">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">{{ t('projectDetail.ciTokensTitle') }}</h3>
          <p class="text-sm text-muted mt-1">{{ t('projectDetail.ciTokensSubtitle') }}</p>
        </div>
      </div>

      <p v-if="!isOwner && project" class="text-xs text-subtle">{{ t('projectDetail.ownerOnlyTokens') }}</p>

      <div v-if="isOwner" class="surface p-5 space-y-4">
        <h4 class="text-sm font-semibold text-ink">{{ t('projectDetail.ciCreateTitle') }}</h4>
        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('common.name') }}</label>
            <input
              v-model="ciTokenForm.name"
              type="text"
              :placeholder="t('projectDetail.ciTokenPlaceholder')"
              class="input"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('common.expiresAtOptional') }}</label>
            <BaseDatePicker
              v-model="ciTokenForm.expires_at"
              :locale="locale"
              :placeholder="t('common.datePlaceholder')"
              :clear-label="t('common.clear')"
              :close-label="t('common.close')"
            />
          </div>
        </div>
        <p class="text-xs text-subtle">{{ t('projectDetail.ciTokenOnce') }}</p>
        <div class="flex items-center gap-3">
          <button
            class="btn btn-primary"
            :disabled="ciTokenCreating"
            @click="handleCreateCIToken"
          >
            {{ ciTokenCreating ? t('common.creating') : t('account.tokens.createButton') }}
          </button>
          <p v-if="ciTokenCreateError" class="text-sm text-danger">{{ ciTokenCreateError }}</p>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <p class="text-sm font-semibold text-ink">{{ t('account.tokens.existingTitle') }}</p>
          <p v-if="ciTokensLoading" class="text-xs text-muted">{{ t('account.tokens.loading') }}</p>
        </div>
        <p v-if="ciTokensError" class="text-sm text-danger">{{ ciTokensError }}</p>
        <p v-else-if="ciTokens.length === 0" class="text-sm text-muted">{{ t('projectDetail.ciTokensEmpty') }}</p>
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th class="py-2 pr-4">{{ t('common.name') }}</th>
                <th class="py-2 pr-4">{{ t('common.created') }}</th>
                <th class="py-2 pr-4">{{ t('common.lastUsed') }}</th>
                <th class="py-2 pr-4">{{ t('common.expires') }}</th>
                <th class="py-2 pr-4">{{ t('common.status') }}</th>
                <th class="py-2 text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="token in ciTokens" :key="token.id">
                <td class="py-3 pr-4">{{ token.name }}</td>
                <td class="py-3 pr-4 text-muted">{{ formatDate(token.created_at) }}</td>
                <td class="py-3 pr-4 text-muted">
                  {{ token.last_used_at ? formatDate(token.last_used_at) : t('common.never') }}
                </td>
                <td class="py-3 pr-4 text-muted">
                  {{ token.expires_at ? formatDate(token.expires_at) : t('common.empty') }}
                </td>
                <td class="py-3 pr-4">
                  <span
                    v-if="token.revoked_at"
                    class="badge badge-danger"
                  >
                    {{ t('common.revoked') }}
                  </span>
                  <span
                    v-else
                    class="badge badge-success"
                  >
                    {{ t('common.active') }}
                  </span>
                </td>
                <td class="py-3 text-right">
                  <button
                    v-if="isOwner && !token.revoked_at"
                    class="text-xs text-danger hover:text-danger/80"
                    :disabled="revokingTokenId === token.id"
                    @click="handleRevokeCIToken(token)"
                  >
                    {{ revokingTokenId === token.id ? t('account.tokens.revoking') : t('account.tokens.revoke') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

    <section class="panel p-6 space-y-6 ds-reveal">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-xl font-semibold">{{ t('projectDetail.registriesTitle') }}</h3>
          <p class="text-sm text-muted mt-1">{{ t('projectDetail.registriesSubtitle') }}</p>
        </div>
        <button
          v-if="isOwner"
          class="btn btn-primary"
          @click="toggleForm"
        >
          {{ showForm ? t('projectDetail.closeRegistryForm') : t('projectDetail.addRegistry') }}
        </button>
      </div>

      <p v-if="!isOwner && project" class="text-xs text-subtle">
        {{ t('projectDetail.ownerOnlyRegistries') }}
      </p>

      <div
        v-if="showForm && isOwner"
        class="surface p-5 space-y-4"
      >
        <h4 class="text-sm font-semibold text-ink">{{ t('projectDetail.addRegistryTitle') }}</h4>
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleCreateRegistry">
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('common.name') }}</label>
            <input
              v-model="form.name"
              type="text"
              :placeholder="t('projectDetail.registryNamePlaceholder')"
              class="input"
            />
            <p v-if="fieldErrors.name" class="text-xs text-danger">{{ fieldErrors.name }}</p>
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('projectDetail.registryUrlLabel') }}</label>
            <input
              v-model="form.registry_url"
              type="url"
              :placeholder="t('projectDetail.registryUrlPlaceholder')"
              class="input"
            />
            <p v-if="fieldErrors.registry_url" class="text-xs text-danger">
              {{ fieldErrors.registry_url }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('projectDetail.registryUsernameLabel') }}</label>
            <input
              v-model="form.username"
              type="text"
              :placeholder="t('projectDetail.registryUsernamePlaceholder')"
              class="input"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('projectDetail.registryPasswordLabel') }}</label>
            <input
              v-model="form.password"
              type="password"
              :placeholder="t('projectDetail.registryPasswordPlaceholder')"
              class="input"
            />
          </div>
          <div class="md:col-span-2 flex items-center gap-3">
            <button
              type="submit"
              class="btn btn-primary"
              :disabled="creatingRegistry"
            >
              {{ creatingRegistry ? t('common.saving') : t('common.save') }}
            </button>
            <button
              type="button"
              class="btn btn-ghost"
              :disabled="creatingRegistry"
              @click="resetForm"
            >
              {{ t('common.clear') }}
            </button>
          </div>
        </form>
        <p v-if="createRegistryError" class="text-sm text-danger">{{ createRegistryError }}</p>
      </div>

      <div>
        <p v-if="registriesLoading" class="text-sm text-muted">{{ t('projectDetail.registriesLoading') }}</p>
        <p v-else-if="registriesError" class="text-sm text-danger">{{ registriesError }}</p>
        <p v-else-if="registries.length === 0" class="text-sm text-muted">
          {{ t('projectDetail.registriesEmpty') }}
        </p>
        <div v-else class="grid gap-4 md:grid-cols-2">
          <div
            v-for="registry in registries"
            :key="registry.id"
            class="surface p-4 space-y-3 transition hover:border-primary/40"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-base font-semibold text-ink">{{ registry.name }}</p>
                <p class="text-xs text-subtle mt-1">{{ registry.registry_url }}</p>
              </div>
              <span class="badge badge-neutral">
                {{ t('projectDetail.registryGeneric') }}
              </span>
            </div>
            <p v-if="registry.username" class="text-xs text-muted">
              {{ t('projectDetail.registryUsernamePrefix', { value: registry.username }) }}
            </p>
            <div v-if="isOwner" class="pt-1">
              <div class="flex items-center gap-3">
                <button
                  class="text-xs text-primary hover:text-primary-strong"
                  type="button"
                  @click="openEditRegistry(registry)"
                >
                  {{ t('common.edit') }}
                </button>
                <button
                  class="text-xs text-danger hover:text-danger/80"
                  :disabled="deletingRegistryId === registry.id"
                  @click="handleDeleteRegistry(registry.id)"
                >
                  {{ deletingRegistryId === registry.id ? t('common.deleting') : t('common.delete') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Transition name="modal-fade">
        <div
          v-if="editingRegistry"
          class="fixed inset-0 z-50 flex items-center justify-center bg-base/80 backdrop-blur-sm px-4"
        >
          <Transition name="modal-panel">
            <div class="panel w-full max-w-lg p-6 space-y-4">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h4 class="text-lg font-semibold text-ink">{{ t('projectDetail.editRegistryTitle') }}</h4>
                  <p class="text-xs text-muted">{{ t('projectDetail.editRegistrySubtitle') }}</p>
                </div>
              <button class="btn btn-ghost btn-icon" type="button" @click="closeEdit" aria-label="Close">
                <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <path :d="closeIcon" />
                </svg>
              </button>
              </div>
              <form class="space-y-4" @submit.prevent="handleUpdateRegistry">
                <div class="space-y-1">
                  <label class="text-xs font-medium text-subtle">{{ t('common.name') }}</label>
                  <input v-model="editForm.name" type="text" class="input" />
                  <p v-if="editErrors.name" class="text-xs text-danger">{{ editErrors.name }}</p>
                </div>
                <div class="space-y-1">
                  <label class="text-xs font-medium text-subtle">{{ t('projectDetail.registryUrlLabel') }}</label>
                  <input v-model="editForm.registry_url" type="url" class="input" />
                  <p v-if="editErrors.registry_url" class="text-xs text-danger">
                    {{ editErrors.registry_url }}
                  </p>
                </div>
                <div class="space-y-1">
                  <label class="text-xs font-medium text-subtle">{{ t('projectDetail.registryUsernameLabel') }}</label>
                  <input
                    v-model="editForm.username"
                    type="text"
                    class="input"
                    :placeholder="t('projectDetail.registryUsernamePlaceholder')"
                  />
                  <p v-if="editErrors.username" class="text-xs text-danger">
                    {{ editErrors.username }}
                  </p>
                </div>
                <div class="space-y-1">
                  <label class="text-xs font-medium text-subtle">{{ t('projectDetail.editRegistryTokenLabel') }}</label>
                  <input
                    v-model="editForm.token"
                    type="password"
                    class="input"
                    :placeholder="t('projectDetail.registryPasswordPlaceholder')"
                  />
                  <p class="text-xs text-subtle">{{ t('projectDetail.editRegistryTokenHint') }}</p>
                  <p v-if="editErrors.token" class="text-xs text-danger">
                    {{ editErrors.token }}
                  </p>
                </div>
                <p v-if="editRegistryError" class="text-sm text-danger">{{ editRegistryError }}</p>
                <p v-if="editRegistrySuccess" class="text-sm text-success">
                  {{ editRegistrySuccess }}
                </p>
                <div class="flex items-center justify-end gap-3">
                  <button type="button" class="btn btn-ghost" :disabled="savingRegistry" @click="closeEdit">
                    {{ t('common.cancel') }}
                  </button>
                  <button type="submit" class="btn btn-primary" :disabled="savingRegistry">
                    {{ savingRegistry ? t('common.saving') : t('common.saveChanges') }}
                  </button>
                </div>
              </form>
            </div>
          </Transition>
        </div>
      </Transition>
    </section>

    <section class="panel p-6 space-y-6 ds-reveal">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <div class="flex items-center gap-3">
            <h3 class="text-xl font-semibold">{{ t('projectDetail.analysesTitle') }}</h3>
            <span v-if="polling" class="text-xs text-subtle">{{ t('analysisDetail.updating') }}</span>
          </div>
          <p class="text-sm text-muted mt-1">{{ t('projectDetail.analysesSubtitle') }}</p>
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <RouterLink
            class="btn btn-secondary px-3 py-1.5 text-xs"
            :to="`/projects/${project?.id}/history`"
          >
            {{ t('projectDetail.history') }}
          </RouterLink>
          <RouterLink
            class="btn btn-secondary px-3 py-1.5 text-xs"
            :to="`/projects/${project?.id}/trends`"
          >
            {{ t('projectDetail.trends') }}
          </RouterLink>
          <button
            v-if="isOwner"
            class="btn btn-primary"
            :disabled="registries.length === 0"
            @click="toggleAnalysisForm"
          >
            {{ showAnalysisForm ? t('common.close') : t('projectDetail.newAnalysis') }}
          </button>
        </div>
      </div>

      <p v-if="!isOwner && project" class="text-xs text-subtle">
        {{ t('projectDetail.ownerOnlyAnalyses') }}
      </p>
      <p v-if="isOwner && registries.length === 0" class="text-xs text-subtle">
        {{ t('projectDetail.analysisRequiresRegistry') }}
      </p>

      <div
        v-if="showAnalysisForm && isOwner"
        class="surface p-5 space-y-4"
      >
        <h4 class="text-sm font-semibold text-ink">{{ t('projectDetail.requestAnalysisTitle') }}</h4>
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleCreateAnalysis">
          <div class="space-y-1 md:col-span-2">
            <label class="text-xs font-medium text-subtle">{{ t('common.registry') }}</label>
            <BaseSelect
              v-model="analysisForm.registry_id"
              :options="registryOptions"
              :placeholder="t('projectDetail.selectRegistry')"
              searchable
              :search-placeholder="t('common.search')"
              :empty-label="t('common.noResults')"
            />
            <p v-if="analysisErrors.registry_id" class="text-xs text-danger">
              {{ analysisErrors.registry_id }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('common.image') }}</label>
            <input
              v-model="analysisForm.image"
              type="text"
              :placeholder="t('projectDetail.imagePlaceholder')"
              class="input"
            />
            <p v-if="analysisErrors.image" class="text-xs text-danger">
              {{ analysisErrors.image }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-subtle">{{ t('common.tag') }}</label>
            <input
              v-model="analysisForm.tag"
              type="text"
              :placeholder="t('projectDetail.tagPlaceholder')"
              class="input"
            />
          </div>
          <div class="md:col-span-2 flex items-center gap-3">
            <button
              type="submit"
              class="btn btn-primary"
              :disabled="creatingAnalysis"
            >
              {{ creatingAnalysis ? t('common.submitting') : t('projectDetail.startAnalysis') }}
            </button>
            <button
              type="button"
              class="btn btn-ghost"
              :disabled="creatingAnalysis"
              @click="resetAnalysisForm"
            >
              {{ t('common.clear') }}
            </button>
          </div>
        </form>
        <p v-if="createAnalysisError" class="text-sm text-danger">{{ createAnalysisError }}</p>
      </div>

      <div>
        <p v-if="analysesLoading" class="text-sm text-muted">{{ t('projectDetail.analysesLoading') }}</p>
        <p v-else-if="analysesError" class="text-sm text-danger">{{ analysesError }}</p>
        <p v-else-if="analyses.length === 0" class="text-sm text-muted">
          {{ t('projectDetail.analysisEmpty') }}
        </p>
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th class="py-2 pr-4">{{ t('common.image') }}</th>
                <th class="py-2 pr-4">{{ t('common.status') }}</th>
                <th class="py-2 pr-4">{{ t('common.created') }}</th>
                <th class="py-2">{{ t('common.totalSize') }}</th>
                <th class="py-2 text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="analysis in analyses" :key="analysis.id">
                <td class="py-3 pr-4">
                  <RouterLink
                    class="link"
                    :to="`/projects/${project?.id}/analyses/${analysis.id}`"
                  >
                    {{ analysis.image }}:{{ analysis.tag }}
                  </RouterLink>
                </td>
                <td class="py-3 pr-4">
                  <span
                    class="badge"
                    :class="statusBadgeClass(analysis.status)"
                  >
                    {{ statusLabel(analysis.status) }}
                  </span>
                </td>
                <td class="py-3 pr-4 text-muted">
                  {{ formatDate(analysis.created_at) }}
                </td>
                <td class="py-3 text-muted">
                  {{ analysis.total_size_bytes ? formatBytes(analysis.total_size_bytes) : t('common.empty') }}
                </td>
                <td class="py-3 text-right">
                  <div class="flex items-center justify-end gap-3">
                    <RouterLink
                      v-if="getPreviousCompletedAnalysis(analysis)"
                      class="text-xs text-primary hover:text-primary-strong"
                      :to="`/projects/${project?.id}/analyses/compare?from=${getPreviousCompletedAnalysis(analysis)?.id}&to=${analysis.id}`"
                    >
                      {{ t('projectDetail.analysisCompare') }}
                    </RouterLink>
                    <button
                      v-if="isOwner"
                      class="text-xs text-danger hover:text-danger/80 disabled:opacity-60"
                      :disabled="deletingAnalysisId === analysis.id"
                      @click="handleDeleteAnalysis(analysis)"
                    >
                      {{ deletingAnalysisId === analysis.id ? t('common.deleting') : t('common.delete') }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

    <Transition name="modal-fade">
      <div
        v-if="showTokenModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-base/80 backdrop-blur-sm px-4"
      >
        <Transition name="modal-panel">
          <div class="panel w-full max-w-xl p-6 space-y-4">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h4 class="text-lg font-semibold text-ink">{{ t('projectDetail.newTokenTitle') }}</h4>
                <p class="text-xs text-warning">{{ t('projectDetail.newTokenSubtitle') }}</p>
              </div>
            <button class="btn btn-ghost btn-icon" type="button" @click="showTokenModal = false" aria-label="Close">
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path :d="closeIcon" />
              </svg>
            </button>
            </div>
            <div class="surface p-4 space-y-3">
              <p class="text-sm text-ink break-all font-mono">{{ createdTokenValue }}</p>
              <div class="flex items-center gap-3">
                <button type="button" class="btn btn-primary" @click="copyToken">
                  {{ t('projectDetail.copyToken') }}
                </button>
                <button type="button" class="btn btn-secondary" @click="showTokenModal = false">
                  {{ t('projectDetail.tokenCopiedClose') }}
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>

    <BaseConfirmModal
      v-model="confirmRevokeOpen"
      :title="t('projectDetail.revokeTokenConfirm', { name: pendingRevokeToken?.name || '' })"
      :confirm-label="t('account.tokens.revoke')"
      :cancel-label="t('common.cancel')"
      tone="danger"
      @confirm="confirmRevokeToken"
    />
    <BaseConfirmModal
      v-model="confirmOverrideDeleteOpen"
      :title="t('projectDetail.overrideDeleteConfirm')"
      :confirm-label="t('common.delete')"
      :cancel-label="t('common.cancel')"
      tone="danger"
      @confirm="confirmDeleteOverride"
    />
    <BaseConfirmModal
      v-model="confirmAnalysisDeleteOpen"
      :title="t('projectDetail.analysisDeleteConfirm')"
      :confirm-label="t('common.delete')"
      :cancel-label="t('common.cancel')"
      tone="danger"
      @confirm="confirmDeleteAnalysis"
    />
    <BaseConfirmModal
      v-model="confirmProjectDeleteOpen"
      :title="t('projectDetail.deleteConfirm')"
      :confirm-label="t('projectDetail.deleteProject')"
      :cancel-label="t('common.cancel')"
      tone="danger"
      @confirm="confirmDeleteProject"
    />
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  createAnalysis,
  createRegistry,
  createCIToken,
  listCITokens,
  deleteAnalysis,
  deleteProject,
  deleteRegistry,
  getBudgets,
  getProject,
  listAnalyses,
  listRegistries,
  revokeCIToken,
  updateProject,
  updateRegistry,
  upsertDefaultBudget,
  createBudgetOverride,
  updateBudgetOverride,
  deleteBudgetOverride,
} from '../api/client'
import BaseConfirmModal from '../components/BaseConfirmModal.vue'
import BaseSelect from '../components/BaseSelect.vue'
import BaseDatePicker from '../components/BaseDatePicker.vue'
import { mdiClose } from '@mdi/js'

const route = useRoute()
const router = useRouter()
const { locale, t, tm } = useI18n()
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
const confirmRevokeOpen = ref(false)
const confirmOverrideDeleteOpen = ref(false)
const confirmAnalysisDeleteOpen = ref(false)
const confirmProjectDeleteOpen = ref(false)
const pendingRevokeToken = ref(null)
const pendingOverrideId = ref(null)
const pendingAnalysis = ref(null)
const closeIcon = mdiClose
const ciTokens = ref([])
const ciTokensLoading = ref(false)
const ciTokensError = ref('')
const ciTokenForm = ref({ name: '', expires_at: '' })
const ciTokenCreateError = ref('')
const ciTokenCreating = ref(false)
const showTokenModal = ref(false)
const createdTokenValue = ref('')
const revokingTokenId = ref(null)
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

const registryOptions = computed(() =>
  registries.value.map((registry) => ({
    value: registry.id,
    label: `${registry.name} · ${registry.registry_url}`,
  })),
)

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
    if (isOwner.value) {
      await fetchCITokens()
    } else {
      ciTokens.value = []
    }
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

const fetchCITokens = async () => {
  if (!isOwner.value) return
  ciTokensLoading.value = true
  ciTokensError.value = ''
  try {
    ciTokens.value = await listCITokens(route.params.id)
  } catch (err) {
    ciTokensError.value = err.message
  } finally {
    ciTokensLoading.value = false
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
    fieldErrors.value.name = t('projectDetail.editRegistryNameRequired')
  }
  if (!urlValue) {
    fieldErrors.value.registry_url = t('projectDetail.editRegistryUrlRequired')
  }

  const duplicate = registries.value.find((r) => r.name === nameValue)
  if (!fieldErrors.value.name && duplicate) {
    fieldErrors.value.name = t('projectDetail.registryDuplicate')
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
      createRegistryError.value = t('projectDetail.registryDuplicate')
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
      budgetsError.value = t('projectDetail.budgetInvalid')
      defaultBudgetSaving.value = false
      return
    }
    const saved = await upsertDefaultBudget(route.params.id, payload)
    budgetsDefault.value = saved
    syncDefaultBudgetForm()
    defaultBudgetSuccess.value = t('projectDetail.budgetSaved')
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

const toRFC3339 = (value) => {
  if (!value) return null
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return null
  return date.toISOString()
}

const resetCITokenForm = () => {
  ciTokenForm.value = { name: '', expires_at: '' }
  ciTokenCreateError.value = ''
}

const handleCreateCIToken = async () => {
  if (!isOwner.value) return
  ciTokenCreateError.value = ''
  ciTokenCreating.value = true
  try {
    const nameValue = ciTokenForm.value.name.trim()
    if (!nameValue) {
      ciTokenCreateError.value = t('projectDetail.ciTokenNameRequired')
      ciTokenCreating.value = false
      return
    }
    const expiresValue = ciTokenForm.value.expires_at ? toRFC3339(ciTokenForm.value.expires_at) : null
    const payload = { name: nameValue }
    if (expiresValue) {
      payload.expires_at = expiresValue
    }
    const created = await createCIToken(route.params.id, payload)
    createdTokenValue.value = created.token
    showTokenModal.value = true
    resetCITokenForm()
    await fetchCITokens()
  } catch (err) {
    if (err.status === 409) {
      ciTokenCreateError.value = t('projectDetail.ciTokenDuplicate')
    } else {
      ciTokenCreateError.value = err.message
    }
  } finally {
    ciTokenCreating.value = false
  }
}

const handleRevokeCIToken = async (token) => {
  if (!isOwner.value || token.revoked_at) return
  pendingRevokeToken.value = token
  confirmRevokeOpen.value = true
}

const confirmRevokeToken = async () => {
  if (!pendingRevokeToken.value) return
  revokingTokenId.value = pendingRevokeToken.value.id
  try {
    await revokeCIToken(route.params.id, pendingRevokeToken.value.id)
    await fetchCITokens()
  } catch (err) {
    ciTokensError.value = err.message
  } finally {
    revokingTokenId.value = null
    pendingRevokeToken.value = null
  }
}

const copyToken = async () => {
  if (!createdTokenValue.value) return
  try {
    await navigator.clipboard?.writeText(createdTokenValue.value)
  } catch (err) {
    console.warn('clipboard copy failed', err)
  }
}

const handleSaveOverride = async () => {
  if (!isOwner.value) return
  overrideSaving.value = true
  overrideError.value = ''
  try {
    const imageValue = overrideForm.value.image.trim()
    if (!imageValue) {
      overrideError.value = t('projectDetail.analysisImageRequired')
      overrideSaving.value = false
      return
    }
    const duplicate = budgetOverrides.value.find(
      (item) => item.image === imageValue && (!editingOverride.value || item.id !== editingOverride.value.id),
    )
    if (duplicate) {
      overrideError.value = t('projectDetail.overrideDuplicate')
      overrideSaving.value = false
      return
    }
    const { payload, invalid } = buildBudgetPayload(overrideForm.value)
    if (invalid) {
      overrideError.value = t('projectDetail.budgetInvalid')
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
      overrideError.value = t('projectDetail.overrideDuplicate')
    } else {
      overrideError.value = err.message
    }
  } finally {
    overrideSaving.value = false
  }
}

const handleDeleteOverride = async (budgetId) => {
  if (!isOwner.value) return
  pendingOverrideId.value = budgetId
  confirmOverrideDeleteOpen.value = true
}

const confirmDeleteOverride = async () => {
  if (!pendingOverrideId.value) return
  try {
    await deleteBudgetOverride(route.params.id, pendingOverrideId.value)
    budgetOverrides.value = budgetOverrides.value.filter((b) => b.id !== pendingOverrideId.value)
  } catch (err) {
    budgetsError.value = err.message
  } finally {
    pendingOverrideId.value = null
  }
}

const handleCreateAnalysis = async () => {
  analysisErrors.value = {}
  createAnalysisError.value = ''

  if (!analysisForm.value.registry_id) {
    analysisErrors.value.registry_id = t('projectDetail.analysisRegistryRequired')
  }
  if (!analysisForm.value.image.trim()) {
    analysisErrors.value.image = t('projectDetail.analysisImageRequired')
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
  pendingAnalysis.value = analysis
  confirmAnalysisDeleteOpen.value = true
}

const confirmDeleteAnalysis = async () => {
  if (!pendingAnalysis.value) return
  deletingAnalysisId.value = pendingAnalysis.value.id
  try {
    await deleteAnalysis(route.params.id, pendingAnalysis.value.id)
    await fetchAnalyses()
  } catch (err) {
    analysesError.value = err.message
  } finally {
    deletingAnalysisId.value = null
    pendingAnalysis.value = null
  }
}

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

const bytesToMB = (value) => {
  if (value === null || value === undefined) return ''
  const mb = Math.round(Number(value) / (1024 * 1024))
  return Number.isFinite(mb) ? mb : ''
}

const statusBadgeClass = (status) => {
  switch (status) {
    case 'completed':
      return 'badge-success'
    case 'running':
      return 'badge-info'
    case 'queued':
      return 'badge-neutral'
    case 'failed':
      return 'badge-danger'
    default:
      return 'badge-warning'
  }
}

const statusLabel = (status) => {
  if (!status) {
    return t('common.empty')
  }
  return t(`status.${status}`)
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
    editErrors.value.name = t('projectDetail.editRegistryNameRequired')
  }
  if (hasURLChange && !urlValue) {
    editErrors.value.registry_url = t('projectDetail.editRegistryUrlRequired')
  }
  if (wantsCredentialUpdate && !usernameValue) {
    editErrors.value.username = t('projectDetail.editRegistryUsernameRequired')
  }
  if (wantsCredentialUpdate && !tokenValue) {
    editErrors.value.token = t('projectDetail.editRegistryTokenRequired')
  }
  if (!hasNameChange && !hasURLChange) {
    if (!wantsCredentialUpdate) {
      editRegistryError.value = t('projectDetail.editRegistryNoChanges')
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
    editRegistrySuccess.value = t('projectDetail.editRegistrySaved')
    await fetchRegistries()
  } catch (err) {
    if (err.status === 409) {
      editRegistryError.value = t('projectDetail.registryDuplicate')
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
  confirmProjectDeleteOpen.value = true
}

const confirmDeleteProject = async () => {
  if (!project.value) return
  deleteError.value = ''
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
    settingsErrors.value.name = t('projectDetail.projectNameRequired')
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
    settingsError.value = t('projectDetail.projectNoChanges')
    return
  }

  savingProject.value = true
  try {
    const updated = await updateProject(project.value.id, payload)
    project.value = { ...project.value, ...updated }
    syncSettingsForm()
    settingsSuccess.value = t('projectDetail.editRegistrySaved')
  } catch (err) {
    settingsError.value = err.message
  } finally {
    savingProject.value = false
  }
}
</script>
