<template>
  <div class="relative" ref="root">
    <button
      type="button"
      class="input flex items-center justify-between gap-3 text-left"
      :class="disabled ? 'opacity-60 cursor-not-allowed' : ''"
      :disabled="disabled"
      @click="toggle"
    >
      <span class="truncate">
        {{ selectedLabel || placeholder }}
      </span>
      <span class="text-subtle">
        <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <path :d="chevronDownIcon" />
        </svg>
      </span>
    </button>

    <Transition name="modal-fade">
      <div v-if="open" class="absolute right-0 mt-2 w-full min-w-[200px] z-40">
        <div class="surface p-2 shadow-card">
          <div v-if="searchable" class="mb-2">
            <input
              v-model="search"
              type="text"
              class="input"
              :placeholder="searchPlaceholder"
              @keydown.stop
            />
          </div>
          <div class="max-h-60 overflow-y-auto">
            <button
              v-for="option in filteredOptions"
              :key="option.value"
              type="button"
              class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-sm text-ink hover:bg-card/70"
              @click="selectOption(option)"
            >
              <span class="truncate">{{ option.label }}</span>
              <span v-if="option.value === modelValue" class="text-primary">
                <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <path :d="checkIcon" />
                </svg>
              </span>
            </button>
            <div v-if="!filteredOptions.length" class="px-3 py-2 text-xs text-muted">
              {{ emptyLabel }}
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { mdiCheck, mdiChevronDown } from '@mdi/js'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, default: () => [] },
  placeholder: { type: String, default: 'Select' },
  disabled: { type: Boolean, default: false },
  searchable: { type: Boolean, default: false },
  searchPlaceholder: { type: String, default: 'Search…' },
  emptyLabel: { type: String, default: 'No results' },
})

const emit = defineEmits(['update:modelValue'])

const open = ref(false)
const search = ref('')
const root = ref(null)

const selectedLabel = computed(() => {
  const option = props.options.find((item) => item.value === props.modelValue)
  return option ? option.label : ''
})

const chevronDownIcon = mdiChevronDown
const checkIcon = mdiCheck

const filteredOptions = computed(() => {
  if (!props.searchable || !search.value.trim()) return props.options
  const term = search.value.toLowerCase()
  return props.options.filter((option) => String(option.label).toLowerCase().includes(term))
})

const close = () => {
  open.value = false
  search.value = ''
}

const toggle = () => {
  if (props.disabled) return
  open.value = !open.value
}

const selectOption = (option) => {
  emit('update:modelValue', option.value)
  close()
}

const handleClickOutside = (event) => {
  if (!root.value) return
  if (root.value.contains(event.target)) return
  close()
}

const handleEsc = (event) => {
  if (event.key === 'Escape') close()
}

watch(open, (value) => {
  if (value) {
    document.addEventListener('click', handleClickOutside)
    document.addEventListener('keydown', handleEsc)
  } else {
    document.removeEventListener('click', handleClickOutside)
    document.removeEventListener('keydown', handleEsc)
  }
})

onMounted(() => {
  if (open.value) {
    document.addEventListener('click', handleClickOutside)
    document.addEventListener('keydown', handleEsc)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEsc)
})
</script>
