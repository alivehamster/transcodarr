<script setup lang="ts">
import { ref } from 'vue'

defineProps<{ text: string }>()

const anchor = ref<HTMLElement | null>(null)
const visible = ref(false)
const x = ref(0)
const y = ref(0)

function show() {
  if (!anchor.value) return
  const rect = anchor.value.getBoundingClientRect()
  x.value = rect.left + rect.width / 2
  y.value = rect.top - 8
  visible.value = true
}

function hide() {
  visible.value = false
}
</script>

<template>
  <span ref="anchor" class="inline-flex" @mouseenter="show" @mouseleave="hide">
    <span
      class="inline-flex h-4 w-4 items-center justify-center rounded-full border border-gray-400 text-gray-400 hover:border-gray-600 hover:text-gray-600 cursor-pointer text-xs font-bold leading-none"
    >?</span>
  </span>
  <Teleport to="body">
    <span
      v-if="visible"
      class="fixed z-9999 -translate-x-1/2 -translate-y-full whitespace-nowrap rounded-lg bg-gray-800 px-3 py-2 text-xs text-white shadow-lg pointer-events-none"
      :style="{ left: x + 'px', top: y + 'px' }"
    >{{ text }}</span>
  </Teleport>
</template>
