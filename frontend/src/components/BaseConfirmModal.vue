<template>
  <BaseModal :model-value="modelValue" @update:model-value="handleUpdate" @close="handleCancel">
    <div class="space-y-4">
      <div class="flex items-start justify-between gap-3">
        <div>
          <h3 class="text-lg font-semibold text-ink">{{ title }}</h3>
          <p v-if="description" class="text-sm text-muted mt-1">{{ description }}</p>
        </div>
        <button class="btn btn-ghost btn-icon" type="button" @click="handleCancel" aria-label="Close">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path :d="closeIcon" />
          </svg>
        </button>
      </div>
      <div class="flex items-center justify-end gap-2">
        <button class="btn btn-ghost" type="button" @click="handleCancel">
          {{ cancelLabel }}
        </button>
        <button :class="confirmClass" type="button" @click="handleConfirm">
          {{ confirmLabel }}
        </button>
      </div>
    </div>
  </BaseModal>
</template>

<script setup>
import { computed } from 'vue'
import BaseModal from './BaseModal.vue'
import { mdiClose } from '@mdi/js'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  confirmLabel: { type: String, default: 'OK' },
  cancelLabel: { type: String, default: 'Cancel' },
  tone: { type: String, default: 'danger' },
})

const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

const confirmClass = computed(() => {
  if (props.tone === 'danger') return 'btn btn-danger'
  if (props.tone === 'secondary') return 'btn btn-secondary'
  return 'btn btn-primary'
})

const closeIcon = mdiClose

const handleCancel = () => {
  emit('update:modelValue', false)
  emit('cancel')
}

const handleUpdate = (value) => {
  emit('update:modelValue', value)
}

const handleConfirm = () => {
  emit('confirm')
  emit('update:modelValue', false)
}
</script>
