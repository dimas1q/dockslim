<template>
  <div class="space-y-8">
    <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-subtle">{{ t('account.title') }}</p>
        <h2 class="text-2xl font-semibold text-ink">{{ t('account.billing.title') }}</h2>
        <p class="text-sm text-muted mt-1">{{ t('account.billing.subtitle') }}</p>
      </div>
      <div class="flex items-center">
        <RouterLink to="/account/settings" class="btn btn-secondary text-sm">
          {{ t('account.billing.backToSettings') }}
        </RouterLink>
      </div>
    </div>

    <section class="panel p-6 space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-xl font-semibold text-ink">{{ t('account.billing.currentPlan') }}</h3>
      </div>

      <div v-if="loading" class="space-y-3">
        <div class="h-10 rounded skeleton"></div>
      </div>
      <div v-else-if="error" class="text-sm text-danger">
        {{ error }}
      </div>
      <div v-else class="surface p-5 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <p class="text-xs uppercase tracking-[0.2em] text-subtle">{{ t('account.billing.activePlan') }}</p>
          <p class="mt-1 text-xl font-semibold text-ink">{{ currentPlanLabel }}</p>
          <p class="mt-1 text-xs text-muted">
            {{ t('account.billing.statusLabel') }}: {{ currentStatusLabel }}
          </p>
        </div>
        <span class="badge badge-success">{{ currentPlanCode }}</span>
      </div>
    </section>

    <section class="panel p-6 space-y-5">
      <div>
        <h3 class="text-xl font-semibold text-ink">{{ t('account.billing.comparisonTitle') }}</h3>
        <p class="text-sm text-muted mt-1">{{ t('account.billing.comparisonSubtitle') }}</p>
      </div>

      <div class="overflow-x-auto">
        <table class="table min-w-[760px]">
          <thead>
            <tr>
              <th class="py-2 pr-4">{{ t('account.billing.featureColumn') }}</th>
              <th v-for="planId in planOrder" :key="`head-${planId}`" class="py-2 pr-4">
                {{ t(`account.billing.planNames.${planId}`) }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in featureRows" :key="row.key">
              <td class="py-3 pr-4 text-sm text-ink">{{ t(`account.billing.features.${row.key}`) }}</td>
              <td
                v-for="planId in planOrder"
                :key="`${row.key}-${planId}`"
                class="py-3 pr-4 text-sm"
              >
                <span :class="cellClass(row.values[planId])">
                  {{ formatFeatureValue(row.values[planId]) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="grid gap-4 md:grid-cols-3">
        <div v-for="planId in planOrder" :key="`card-${planId}`" class="surface p-4 space-y-3">
          <div class="flex items-center justify-between">
            <p class="font-semibold text-ink">{{ t(`account.billing.planNames.${planId}`) }}</p>
            <span v-if="planId === currentPlanCode" class="badge badge-success">
              {{ t('account.billing.currentBadge') }}
            </span>
          </div>
          <button class="btn btn-primary w-full" :disabled="true">
            {{
              planId === currentPlanCode
                ? t('account.billing.currentButton')
                : t('account.billing.upgradeButton')
            }}
          </button>
          <p class="text-xs text-muted">
            {{ t('account.billing.upgradeHint') }}
          </p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { fetchSubscription } from '../api/client'

const { t } = useI18n()

const loading = ref(true)
const error = ref('')
const subscription = ref(null)

const planOrder = ['free', 'pro', 'team']

const featureRows = [
  {
    key: 'basic_analysis',
    values: { free: true, pro: true, team: true },
  },
  {
    key: 'history_days_limit',
    values: { free: 30, pro: null, team: null },
  },
  {
    key: 'advanced_insights',
    values: { free: false, pro: true, team: true },
  },
  {
    key: 'export_pdf',
    values: { free: false, pro: true, team: true },
  },
  {
    key: 'export_json',
    values: { free: false, pro: true, team: true },
  },
  {
    key: 'ci_comments',
    values: { free: 'limited', pro: true, team: true },
  },
  {
    key: 'baseline_sla',
    values: { free: false, pro: true, team: true },
  },
  {
    key: 'team_management',
    values: { free: false, pro: false, team: true },
  },
  {
    key: 'shared_projects',
    values: { free: false, pro: false, team: true },
  },
  {
    key: 'advanced_trends',
    values: { free: false, pro: false, team: true },
  },
]

const currentPlanCode = computed(() => subscription.value?.plan?.id || 'free')
const currentPlanLabel = computed(() =>
  t(`account.billing.planNames.${currentPlanCode.value}`),
)
const currentStatusLabel = computed(() => {
  const status = subscription.value?.plan?.status || 'active'
  return t(`account.billing.status.${status}`)
})

const formatFeatureValue = (value) => {
  if (value === true) return t('account.billing.valueEnabled')
  if (value === false) return t('account.billing.valueDisabled')
  if (value === null) return t('account.billing.valueUnlimited')
  if (value === 'limited') return t('account.billing.valueLimited')
  return String(value)
}

const cellClass = (value) => {
  if (value === true || value === null) {
    return 'badge badge-success'
  }
  if (value === 'limited') {
    return 'badge badge-warning'
  }
  return 'badge badge-danger'
}

const loadSubscription = async () => {
  loading.value = true
  error.value = ''
  try {
    subscription.value = await fetchSubscription()
  } catch (err) {
    error.value = err.message || t('account.billing.loadError')
  } finally {
    loading.value = false
  }
}

onMounted(loadSubscription)
</script>
