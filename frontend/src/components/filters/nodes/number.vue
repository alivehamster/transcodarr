<script setup lang="ts">
import { type NodeProps } from '@vue-flow/core'

import baseNode from './base.vue'

const props = defineProps<NodeProps>()

</script>

<template>
  <baseNode :label="props.data.label" v-model:skipFuture="props.data.skipFuture">

    <div class="flex items-center gap-2">
      <input v-model.number="props.data.num" type="number" min="1" :placeholder="props.data.placeholder"
        class="w-28 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
      <span class="text-sm text-gray-500">{{ props.data.placeholder }}</span>
    </div>

  </baseNode>
</template>