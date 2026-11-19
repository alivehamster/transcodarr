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
  <li>
    <details v-if="node.type === 'folder'">
      <summary>{{ node.name }}</summary>
      <ul v-if="node.children && node.children.length > 0" class="pl-4">
        <SkiplistTreeNode v-for="child in node.children" :key="child.name" :node="child" />
      </ul>
    </details>
    <span v-else>
      <p v-if="!node.exists" class="text-red-500 text-sm">File does not exist in library:</p>
      <input type="checkbox" :checked="node.id !== null" @change="handleCheckboxChange"/> {{ node.name }}
    </span>
  </li>
</template>
