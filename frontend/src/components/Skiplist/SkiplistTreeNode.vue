<script setup lang="ts">
import { checkboxChange } from './utils'
import type { FileNode } from './utils'

defineOptions({ name: 'SkiplistTreeNode' })

const props = defineProps<{
  node: FileNode
}>()

function handleCheckboxChange(event: Event) {
  if (event.target instanceof HTMLInputElement) {
    checkboxChange(props.node.id, props.node.path, event.target.checked)
  }
}
</script>

<template>
  <li class="odd:[&>div]:bg-slate-50 even:[&>div]:bg-slate-100">
    <details v-if="node.type === 'folder'">
      <summary>{{ node.name }}</summary>
      <ul v-if="node.children && node.children.length > 0" class="pl-4">
        <SkiplistTreeNode v-for="child in node.children" :key="child.name" :node="child" />
      </ul>
    </details>
    <div v-else class="p-1 rounded-md">
      <span>
        <input type="checkbox" :checked="node.id !== null" @change="handleCheckboxChange" />
        {{ node.name }}
      </span>
      <p>{{ node.description }}</p>
      <p v-if="!node.exists" class="text-red-500 text-sm">File does not exist in directory</p>
    </div>
  </li>
</template>
