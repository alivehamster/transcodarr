<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Skip {
  id: number
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
const adding = ref(false)
const addPath = ref('')
const addError = ref('')

async function loadSkiplist() {
  try {
    const response = await fetch(`/api/skiplist/${props.id}`)
    skiplist.value = await response.json()
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

async function removeItem(id: number) {
  saving.value = true
  try {
    await fetch(`/api/removeSkip/${id}`, {
      method: 'DELETE',
    })
    skiplist.value = skiplist.value.filter(item => item.id !== id)
  } catch (error) {
    console.error('Error saving skiplist:', error)
  } finally {
    saving.value = false
  }
}

async function addItem() {
  const path = addPath.value.trim()
  if (!path) {
    addError.value = 'Path is required.'
    return
  }

  adding.value = true
  addError.value = ''

  try {
    const response = await fetch(`/api/addSkip/${props.id}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        path,
      }),
    })

    if (!response.ok) {
      throw new Error(`Request failed with status ${response.status}`)
    }

    await loadSkiplist()
    addPath.value = ''
  } catch (error) {
    console.error('Error adding skip item:', error)
    addError.value = 'Failed to add skip item.'
  } finally {
    adding.value = false
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

      <div v-if="!loading && skiplist.length === 0" class="text-sm text-gray-500">
        No items in the skip list.
      </div>

      <div class="mb-4 mt-4 rounded-lg border border-gray-200 p-3">
        <div class="grid gap-2 sm:grid-cols-[1fr_auto]">
          <input
            v-model="addPath"
            type="text"
            placeholder="Path to skip"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-gray-500"
          >
          <button
            type="button"
            :disabled="adding"
            class="rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer disabled:opacity-50"
            @click="addItem"
          >
            {{ adding ? 'Adding...' : 'Add Skip' }}
          </button>
        </div>
        <p v-if="addError" class="mt-2 text-xs text-red-600">{{ addError }}</p>
      </div>

      <ul v-if="!loading && skiplist.length > 0" class="space-y-2 max-h-96 overflow-y-auto">
        <li
          v-for="item in skiplist"
          :key="item.id"
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
            @click="removeItem(item.id)"
          >
            ✕
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>
