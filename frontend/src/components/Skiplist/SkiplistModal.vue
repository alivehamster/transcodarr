<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { buildFileTree, combineStuff, addSkip, removeSkip } from './utils'
import type { FileNode } from './utils'
import SkiplistTreeNode from './SkiplistTreeNode.vue'

const props = defineProps<{
  id: number
}>()

const emit = defineEmits<{
  close: []
}>()

const fileTree = ref<FileNode[]>([])
const loading = ref(true)

async function loadSkiplist() {
  try {
    const response = await fetch(`/api/skiplist/${props.id}`)
    const response2 = await fetch(`/api/libItems/${props.id}`)

    const skips = await response.json()
    const paths = (await response2.json()) || []
    fileTree.value = buildFileTree(combineStuff(skips, paths))
  } catch (error) {
    console.error('Error fetching skiplist:', error)
  }
}

onMounted(async () => {
  try {
    await loadSkiplist()
  } finally {
    loading.value = false
  }
})

async function updateSkips() {
  await fetch(`/api/updateSkiplist/${props.id}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      add: addSkip.value,
      remove: removeSkip.value,
    }),
  })
  emit('close')
}

onUnmounted(() => {
  fileTree.value = []
  addSkip.value = []
  removeSkip.value = []
  loading.value = true
})
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="relative w-full rounded-2xl bg-white p-6 shadow-xl max-w-3/4">
      <div class="mb-5 flex items-center justify-between">
        <h2 class="text-xl font-semibold">Skip List:</h2>
        <button
          type="button"
          class="rounded-lg border border-gray-300 px-2 py-1 text-gray-500 hover:bg-gray-100 cursor-pointer"
          @click="emit('close')"
        >
          ✕
        </button>
      </div>

      <div v-if="loading" class="text-sm text-gray-500">Loading...</div>

      <div v-if="!loading && fileTree.length === 0" class="text-sm text-gray-500">
        No items in the skip list.
      </div>

      <div v-if="!loading && fileTree.length > 0" class="max-h-[60vh] overflow-y-auto">
        <ul class="pl-4">
          <SkiplistTreeNode v-for="node in fileTree" :key="node.name" :node="node" />
        </ul>
      </div>
      <div v-if="addSkip.length > 0 || removeSkip.length > 0" class="absolute bottom-2 right-2 p-2 z-10">
        <button
          @click="updateSkips"
          class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 hover:cursor-pointer"
        >
          Save
        </button>
      </div>
    </div>
  </div>
</template>
