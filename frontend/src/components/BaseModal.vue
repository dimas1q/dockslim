<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center px-4">
        <div class="absolute inset-0 bg-black/60" @click="handleBackdrop"></div>
        <div class="relative w-full max-w-2xl">
          <div class="card p-6 shadow-card">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { onBeforeUnmount, onMounted, watch } from 'vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  closeOnEsc: { type: Boolean, default: true },
  closeOnOutside: { type: Boolean, default: true },
})

const emit = defineEmits(['update:modelValue', 'close'])

const close = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleBackdrop = () => {
  if (props.closeOnOutside) {
    close()
  }
}

const handleKeydown = (event) => {
  if (event.key === 'Escape' && props.closeOnEsc) {
    close()
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      document.addEventListener('keydown', handleKeydown)
    } else {
      document.removeEventListener('keydown', handleKeydown)
    }
  },
  { immediate: true },
)

onMounted(() => {
  if (props.modelValue) {
    document.addEventListener('keydown', handleKeydown)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>
