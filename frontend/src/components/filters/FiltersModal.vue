<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  VueFlow,
  useVueFlow,
  type Node,
  type Edge,
} from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { filters, genOrder, genNodes, genEdges, createNodeId, type OrderedFilter } from './filters'

// nodes
import codec from './nodes/codec.vue'
import number from './nodes/number.vue'
import noinput from './nodes/noinput.vue'

import Tooltip from '../Tooltip.vue'

const props = defineProps<{
  order?: OrderedFilter[]
}>()

const emit = defineEmits<{
  cancel: []
  save: [payload: OrderedFilter[]]
}>()

const { onConnect, onEdgeUpdateStart, onEdgeUpdate, onEdgeUpdateEnd, addEdges, updateEdge, removeEdges, addNodes, toObject, getEdges } = useVueFlow()



const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])
const saveError = ref('')

onMounted(async () => {
  nodes.value = genNodes(props.order ?? [])
  edges.value = genEdges(nodes.value)
})

function newNode(type?: string, data?: any) {
  const newNode: Node = {
    id: createNodeId(),
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
  console.log(genOrder(toObject()))
}

function handleSave() {
  const result = genOrder(toObject())
  if (typeof result === 'string') {
    saveError.value = result
    return
  }

  saveError.value = ''
  emit('save', result)
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
          <VueFlow v-model:nodes="nodes" v-model:edges="edges" edges-updatable="target" fit-view-on-init
            :default-zoom="1.5" :min-zoom="0.2" :max-zoom="4">
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
              <div v-for="filter in filters" :key="filter.data.id" class="flex w-full items-center justify-center gap-2">
              <button class="bg-blue-500 hover:bg-blue-400 w-3/5 h-15 rounded-lg cursor-pointer"
                @click="newNode(filter.type, filter.data)">
                {{
                  filter.data.label }}
              </button>
              <Tooltip v-if="filter.tooltip" :text="filter.tooltip"></Tooltip>
              </div>
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