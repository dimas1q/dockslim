<template>
  <div class="relative" ref="root">
    <button
      type="button"
      class="input flex items-center justify-between gap-3 text-left"
      :class="disabled ? 'opacity-60 cursor-not-allowed' : ''"
      :disabled="disabled"
      ref="trigger"
      @click="toggle"
    >
      <span class="truncate">
        {{ displayValue || placeholder }}
      </span>
      <span class="text-subtle">
        <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <path :d="calendarIcon" />
        </svg>
      </span>
    </button>

    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="open" ref="popup" class="fixed z-[9999]" :style="floatingStyle">
          <div class="surface p-3 shadow-card max-h-full overflow-auto">
          <div class="flex items-center justify-between mb-3">
            <button class="btn btn-ghost btn-icon" type="button" @click="prevMonth" aria-label="Previous">
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path :d="chevronLeftIcon" />
              </svg>
            </button>
            <p class="text-sm font-semibold text-ink">{{ monthLabel }}</p>
            <button class="btn btn-ghost btn-icon" type="button" @click="nextMonth" aria-label="Next">
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path :d="chevronRightIcon" />
              </svg>
            </button>
          </div>
          <div class="grid grid-cols-7 gap-1 text-xs text-subtle mb-2">
            <span v-for="day in weekdayLabels" :key="day" class="text-center">{{ day }}</span>
          </div>
          <div class="grid grid-cols-7 gap-1 text-sm">
            <button
              v-for="cell in calendarCells"
              :key="cell.key"
              type="button"
              class="h-9 w-full rounded-lg text-center"
              :class="cellClass(cell)"
              :disabled="cell.isOutside"
              @click="selectDate(cell)"
            >
              {{ cell.day }}
            </button>
          </div>
          <div class="mt-3 flex items-center justify-between">
            <button class="btn btn-ghost text-xs" type="button" @click="clear">
              {{ clearLabel }}
            </button>
            <button class="btn btn-secondary text-xs" type="button" @click="close">
              {{ closeLabel }}
            </button>
          </div>
        </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { mdiCalendarMonth, mdiChevronLeft, mdiChevronRight } from '@mdi/js'

const props = defineProps({
  modelValue: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  placeholder: { type: String, default: '' },
  locale: { type: String, default: 'ru' },
  weekStart: { type: Number, default: 1 },
  clearLabel: { type: String, default: 'Clear' },
  closeLabel: { type: String, default: 'Close' },
})

const emit = defineEmits(['update:modelValue'])

const open = ref(false)
const root = ref(null)
const trigger = ref(null)
const floatingStyle = ref({})
const popup = ref(null)
const viewDate = ref(new Date())

const parseValue = (value) => {
  if (!value) return null
  const parts = value.split('-').map((num) => Number(num))
  if (parts.length !== 3) return null
  const [year, month, day] = parts
  if (!year || !month || !day) return null
  return new Date(year, month - 1, day)
}

const formatValue = (date) => {
  if (!date) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const displayValue = computed(() => {
  const date = parseValue(props.modelValue)
  if (!date) return ''
  return new Intl.DateTimeFormat(props.locale, { year: 'numeric', month: 'short', day: 'numeric' }).format(date)
})

const calendarIcon = mdiCalendarMonth
const chevronLeftIcon = mdiChevronLeft
const chevronRightIcon = mdiChevronRight

const monthLabel = computed(() =>
  new Intl.DateTimeFormat(props.locale, { month: 'long', year: 'numeric' }).format(viewDate.value),
)

const weekdayLabels = computed(() => {
  const base = new Date(2023, 0, 1)
  return Array.from({ length: 7 }, (_, index) => {
    const date = new Date(base)
    const shift = (index + props.weekStart) % 7
    date.setDate(base.getDate() + shift)
    return new Intl.DateTimeFormat(props.locale, { weekday: 'short' }).format(date)
  })
})

const calendarCells = computed(() => {
  const year = viewDate.value.getFullYear()
  const month = viewDate.value.getMonth()
  const firstDay = new Date(year, month, 1)
  const startOffset = (firstDay.getDay() - props.weekStart + 7) % 7
  const daysInMonth = new Date(year, month + 1, 0).getDate()

  const cells = []
  for (let i = 0; i < startOffset; i += 1) {
    const day = new Date(year, month, -startOffset + i + 1)
    cells.push({
      key: `prev-${i}`,
      day: day.getDate(),
      date: day,
      isOutside: true,
    })
  }
  for (let day = 1; day <= daysInMonth; day += 1) {
    const date = new Date(year, month, day)
    cells.push({
      key: `day-${day}`,
      day,
      date,
      isOutside: false,
    })
  }
  const remainder = cells.length % 7
  if (remainder !== 0) {
    const extra = 7 - remainder
    for (let i = 1; i <= extra; i += 1) {
      const day = new Date(year, month + 1, i)
      cells.push({
        key: `next-${i}`,
        day: day.getDate(),
        date: day,
        isOutside: true,
      })
    }
  }
  return cells
})

const cellClass = (cell) => {
  if (cell.isOutside) return 'text-subtle opacity-40 cursor-default'
  const selected = props.modelValue && formatValue(cell.date) === props.modelValue
  if (selected) return 'bg-primary/15 text-primary font-semibold'
  return 'text-ink hover:bg-card/60'
}

const toggle = () => {
  if (props.disabled) return
  open.value = !open.value
}

const close = () => {
  open.value = false
}

const clear = () => {
  emit('update:modelValue', '')
  close()
}

const selectDate = (cell) => {
  if (cell.isOutside) return
  emit('update:modelValue', formatValue(cell.date))
  close()
}

const prevMonth = () => {
  viewDate.value = new Date(viewDate.value.getFullYear(), viewDate.value.getMonth() - 1, 1)
}

const nextMonth = () => {
  viewDate.value = new Date(viewDate.value.getFullYear(), viewDate.value.getMonth() + 1, 1)
}

const handleClickOutside = (event) => {
  if (!root.value) return
  if (root.value.contains(event.target)) return
  if (popup.value && popup.value.contains(event.target)) return
  close()
}

const handleEsc = (event) => {
  if (event.key === 'Escape') close()
}

const updatePosition = async () => {
  if (!trigger.value) return
  const rect = trigger.value.getBoundingClientRect()
  await nextTick()
  const popupHeight = popup.value ? popup.value.getBoundingClientRect().height : 0
  const viewportHeight = window.innerHeight || 0
  let top = rect.bottom + 8
  if (popupHeight && top + popupHeight > viewportHeight - 8) {
    top = rect.top - popupHeight - 8
  }
  if (top < 8) {
    top = 8
  }
  let left = rect.left
  const width = Math.max(rect.width, 260)
  if (left + width > window.innerWidth - 8) {
    left = Math.max(8, window.innerWidth - width - 8)
  }
  floatingStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
    width: `${width}px`,
    maxHeight: `${Math.max(240, window.innerHeight - 16)}px`,
  }
}

watch(
  () => props.modelValue,
  (value) => {
    const date = parseValue(value)
    viewDate.value = date || new Date()
  },
  { immediate: true },
)

watch(open, (value) => {
  if (value) {
    updatePosition()
    window.addEventListener('scroll', updatePosition, true)
    window.addEventListener('resize', updatePosition)
    document.addEventListener('click', handleClickOutside)
    document.addEventListener('keydown', handleEsc)
  } else {
    window.removeEventListener('scroll', updatePosition, true)
    window.removeEventListener('resize', updatePosition)
    document.removeEventListener('click', handleClickOutside)
    document.removeEventListener('keydown', handleEsc)
  }
})

onMounted(() => {
  if (open.value) {
    updatePosition()
    window.addEventListener('scroll', updatePosition, true)
    window.addEventListener('resize', updatePosition)
    document.addEventListener('click', handleClickOutside)
    document.addEventListener('keydown', handleEsc)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', updatePosition, true)
  window.removeEventListener('resize', updatePosition)
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEsc)
})
</script>
