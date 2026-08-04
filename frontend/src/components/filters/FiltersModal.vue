<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  VueFlow,
  useVueFlow,
  type Node,
  type Edge,
} from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { save, type OrderedFilter } from './filters'

import codec from './nodes/codec.vue'
import number from './nodes/number.vue'
import noinput from './nodes/noinput.vue'

const props = defineProps<{
  nodes?: Node[]
  edges?: Edge[]
}>()

const emit = defineEmits<{
  cancel: []
  save: [payload: { nodes: Node[]; edges: Edge[]; order: OrderedFilter[] }]
}>()

const { onConnect, onEdgeUpdateStart, onEdgeUpdate, onEdgeUpdateEnd, addEdges, updateEdge, removeEdges, addNodes, toObject, getEdges } = useVueFlow()

const filters = [
  { type: 'number', data: { id: 'fileAge', label: 'File Age', placeholder: 'Days', skipFuture: false } },
  { type: 'number', data: { id: 'minimumFileSize', label: 'Minimum Size', placeholder: 'MB', skipFuture: true } },
  { type: 'noinput', data: { id: 'hardlinks', label: 'Hardlinks', skipFuture: false } },
  { type: 'noinput', data: { id: 'newFileSize', label: 'Original File Size', skipFuture: true } },
  { type: 'codec', data: { id: 'mediaCodec', label: 'Media Codec', mediaCodecs: [], skipFuture: true } },
]

const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])
const saveError = ref('')

onMounted(async () => {
  nodes.value = props.nodes || [
    { id: '1', type: 'input', position: { x: 100, y: 5 }, deletable: false, data: { label: 'Start', id: 'start' } },
    { id: '2', type: 'default', position: { x: 100, y: 100 }, deletable: false, style: { border: '2px solid #22c55e' }, data: { label: 'Transcode', id: 'transcode' } },
    { id: '3', type: 'output', position: { x: 100, y: 200 }, deletable: false, data: { label: 'End', id: 'end' } },
  ]
  edges.value = props.edges || [
    { "id": "1-2", "source": "1", "target": "2", "sourceHandle": null, "targetHandle": null, },
    { "id": "2-3", "source": "2", "target": "3", "sourceHandle": null, "targetHandle": null, }
  ]
})

function newNode(type?: string, data?: any) {
  const newNode: Node = {
    id: `node-${Date.now()}`,
    type: type ?? 'default',
    position: { x: 250, y: 250 },
    data: data
  }

  addNodes([newNode])
}

// Add edge when a connection is made, removing any existing edge on the same handles
onConnect((params) => {
  const existing = getEdges.value.filter(
    (e) => e.source === params.source && e.sourceHandle === params.sourceHandle
      || e.target === params.target && e.targetHandle === params.targetHandle
  )
  removeEdges(existing)
  addEdges([params])
})

let edgeReconnectSuccessful = false

onEdgeUpdateStart(() => {
  edgeReconnectSuccessful = false
})

// allow reconnecting if dragged off
onEdgeUpdate(({ edge, connection }) => {
  edgeReconnectSuccessful = true
  updateEdge(edge, connection)
})

// remove edge if not reconnected
onEdgeUpdateEnd(({ edge }) => {
  if (!edgeReconnectSuccessful) {
    removeEdges([edge])
  }
})

function log() {
  console.log(save(toObject()))
}

function handleSave() {
  const result = save(toObject())
  if ('error' in result) {
    saveError.value = result.error
    return
  }

  saveError.value = ''
  emit('save', {
    nodes: result.nodes as Node[],
    edges: result.edges as Edge[],
    order: result.order,
  })
}
</script>

<template>
  <div class="fixed inset-0 z-60 flex items-center justify-center bg-black/50">

    <div class="flex flex-col w-4/5 h-4/5 bg-white rounded-lg shadow-lg overflow-hidden">
      <div v-if="saveError" class="mx-4 mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
        {{ saveError }}
      </div>

      <div class="flex flex-row min-h-0 flex-1">
        <div class="w-4/5">
          <VueFlow v-model:nodes="nodes" v-model:edges="edges" edges-updatable="target"
            :default-viewport="{ x: 100, y: 50, zoom: 1.2 }" :default-zoom="1.5" :min-zoom="0.2" :max-zoom="4">
            <Background pattern-color="#aaa" :gap="16" />
            <template #node-codec="nodeProps">
              <codec v-bind="nodeProps" />
            </template>
            <template #node-number="nodeProps">
              <number v-bind="nodeProps" />
            </template>
            <template #node-noinput="nodeProps">
              <noinput v-bind="nodeProps" />
            </template>
          </VueFlow>
        </div>
        <div class="w-1/5 flex flex-col items-center justify-between">

          <div class="w-full flex flex-col items-center gap-4 pt-10">
            <template v-for="filter in filters" :key="filter.data.id">
              <button class="bg-blue-500 hover:bg-blue-400 w-3/5 h-15 rounded-lg cursor-pointer"
                @click="newNode(filter.type, filter.data)">
                {{
                  filter.data.label }}
              </button>
            </template>
            <!-- <button class="bg-gray-500 hover:bg-gray-400 w-3/5 h-15 rounded-lg cursor-pointer" @click="log">Log</button> -->


          </div>
          <div class="w-full flex flex-col items-center gap-4 pb-10">
            <button class="bg-red-500 hover:bg-red-400 w-3/5 h-15 rounded-lg cursor-pointer"
              @click="emit('cancel')">Cancel</button>
            <button class="bg-green-500 hover:bg-green-400 w-3/5 h-15 rounded-lg cursor-pointer"
              @click="handleSave">Save</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>