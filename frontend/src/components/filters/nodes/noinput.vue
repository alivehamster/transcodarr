<script setup lang="ts">
import { type NodeProps } from '@vue-flow/core'

import baseNode from './base.vue'

const props = defineProps<NodeProps>()

</script>

<template>
  <baseNode :label="props.data.label" v-model:skipFuture="props.data.skipFuture">
  </baseNode>
</template>