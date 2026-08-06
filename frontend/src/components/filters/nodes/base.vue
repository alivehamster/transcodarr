<script setup lang="ts">
import { Handle, Position} from '@vue-flow/core'
import Tooltip from '../../Tooltip.vue'

const props = defineProps<{
  label: string
  skipFuture?: boolean
}>()

const emit = defineEmits<{
  'update:skipFuture': [value: boolean]
}>()
</script>

<template>
  <div class="custom-node-box border border-gray-300 bg-white p-3 rounded-lg shadow-md min-w-45">
    <Handle type="target" :position="Position.Top" />

    <div class="flex flex-col gap-1">
      <label class="text-xs font-semibold text-gray-600 cursor-grab">
        {{ props.label || 'Node' }}
      </label>

      <slot></slot>

      <div class="flex items-center gap-1 mt-1">
        <input
          :id="`skipFuture-${props.label}`"
          :checked="props.skipFuture"
          type="checkbox"
          class="h-3 w-3 cursor-pointer"
          @change="emit('update:skipFuture', ($event.target as HTMLInputElement).checked)"
        />
        <label :for="`skipFuture-${props.label}`" class="text-xs text-gray-500 cursor-pointer">Skip on future scans</label>
        <Tooltip text="If media is filtered out by this node it will be skipped in future scans (Recommended to enable on properties unlikely to change)"></Tooltip>
      </div>

      <Handle type="source" :position="Position.Bottom" />
    </div>
  </div>
</template>