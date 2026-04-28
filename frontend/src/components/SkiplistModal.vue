<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Skip {
  path: string
  description: string
}

const props = defineProps<{
  id: number
}>()

const emit = defineEmits<{
  close: []
}>()

const skiplist = ref<Skip[]>([])
const loading = ref(true)
const saving = ref(false)

onMounted(async () => {
  try {
    const response = await fetch(`/api/skiplist/${props.id}`)
    skiplist.value = await response.json()
  } catch (error) {
    console.error('Error fetching skiplist:', error)
  } finally {
    loading.value = false
  }
})

async function removeItem(index: number) {
  skiplist.value.splice(index, 1)
  saving.value = true
  try {
    await fetch('/api/editSkiplist', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: props.id, skips: skiplist.value }),
    })
  } catch (error) {
    console.error('Error saving skiplist:', error)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="w-full rounded-2xl bg-white p-6 shadow-xl max-w-3/4">
      <div class="mb-5 flex items-center justify-between">
        <h2 class="text-xl font-semibold">Skip List</h2>
        <button
          type="button"
          class="rounded-lg border border-gray-300 px-2 py-1 text-gray-500 hover:bg-gray-100 cursor-pointer"
          @click="emit('close')"
        >
          ✕
        </button>
      </div>

      <div v-if="loading" class="text-sm text-gray-500">Loading...</div>

      <div v-else-if="skiplist.length === 0" class="text-sm text-gray-500">
        No items in the skip list.
      </div>

      <ul v-else class="space-y-2 max-h-96 overflow-y-auto">
        <li
          v-for="(item, i) in skiplist"
          :key="i"
          class="flex items-center justify-between rounded-lg border border-gray-200 px-3 py-2 text-sm"
        >
          <div class="min-w-0 flex-1 overflow-x-auto">
            <p class="whitespace-nowrap font-medium text-gray-800">{{ item.path }}</p>
            <p v-if="item.description" class="whitespace-nowrap text-xs text-gray-500">{{ item.description }}</p>
          </div>
          <button
            type="button"
            :disabled="saving"
            class="ml-3 shrink-0 rounded-lg border border-red-300 px-2 py-1 text-xs text-red-600 hover:bg-red-50 cursor-pointer disabled:opacity-50"
            @click="removeItem(i)"
          >
            ✕
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>
