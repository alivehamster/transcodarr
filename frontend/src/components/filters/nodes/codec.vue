<script setup lang="ts">
import { type NodeProps } from '@vue-flow/core'
import { ref } from 'vue'

import baseNode from './base.vue'

interface CodecData {
  id: string
  label: string
  mediaCodecs: string[]
  skipFuture: boolean
}

const props = defineProps<NodeProps<CodecData>>()

const availableCodecs = [
  'avc',
  'hevc',
  'av1',
  'vp9',
  'vp8',
  'mpeg-4 visual',
  'mpeg video',
  'theora',
  'vc-1',
  'prores',
];

const selectedCodec = ref<string | null>(null)

function addCodec() {
  if (selectedCodec.value && !props.data.mediaCodecs.includes(selectedCodec.value)) {
    props.data.mediaCodecs.push(selectedCodec.value)
  }
}

function removeCodec(codec: string) {
  props.data.mediaCodecs = props.data.mediaCodecs.filter(c => c !== codec)
}

</script>

<template>
  <baseNode :label="props.data.label" v-model:skipFuture="props.data.skipFuture">

    <div class="flex gap-2">
      <select v-model="selectedCodec"
        class="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none cursor-pointer">
        <option v-for="codec in availableCodecs" :key="codec" :value="codec">{{ codec }}</option>
      </select>
      <button type="button" class="rounded-lg border border-gray-300 px-3 py-2 text-sm hover:bg-gray-100 cursor-pointer"
        @click="addCodec">
        + Add
      </button>
    </div>
    <div v-if="props.data.mediaCodecs.length > 0" class="flex flex-wrap gap-1">
      <span v-for="codec in props.data.mediaCodecs" :key="codec"
        class="flex items-center gap-1 rounded-full bg-blue-100 px-3 py-1 text-xs text-blue-700">
        {{ codec }}
        <button type="button" class="hover:text-blue-900 cursor-pointer" @click="removeCodec(codec)">✕</button>
      </span>
    </div>
  </baseNode>
</template>