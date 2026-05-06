<script setup lang="ts">
import { onMounted, ref } from 'vue'

const history = ref<string[]>([])
const loading = ref(true)

async function fetchHistory() {
  try {
    const response = await fetch('/api/history')
    history.value = await response.json()
  } catch (error) {
    console.error('Error fetching history:', error)
  } finally {
    loading.value = false
  }
}

onMounted(fetchHistory)
</script>

<template>
  <div class="p-6">
    <h1 class="mb-4 text-2xl font-semibold">History</h1>

    <div v-if="loading" class="text-gray-500">Loading...</div>
    <div v-else-if="history.length === 0" class="text-gray-500">No history yet.</div>
    <ul v-else class="space-y-1 font-mono text-sm">
      <li
        v-for="(entry, i) in history"
        :key="i"
        class="rounded bg-gray-50 px-3 py-2 text-gray-700"
      >
        {{ entry }}
      </li>
    </ul>
  </div>
</template>
