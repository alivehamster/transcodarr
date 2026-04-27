<script setup lang="ts">
import { onMounted, ref } from 'vue'
import LibraryModal from '../components/LibraryModal.vue'

interface Library {
  id: number
  name: string
  cron: string
  config: {
    dirs: string[]
    profile: string
  }
}

const libraries = ref<Library[]>([])
const showModal = ref(false)

async function fetchLibraries() {
  try {
    const response = await fetch('/api/libraries')
    libraries.value = await response.json()
  } catch (error) {
    console.error('Error fetching libraries:', error)
  }
}

onMounted(fetchLibraries)

function onSaved() {
  showModal.value = false
  fetchLibraries()
}
</script>

<template>
  <div class="p-6">
    <div class="mb-4 flex items-center justify-between">
      <h1 class="text-2xl font-semibold">Libraries</h1>
      <button
        class="flex items-center gap-1 rounded-lg bg-blue-600 px-3 py-2 text-sm text-white hover:bg-blue-700 cursor-pointer"
        @click="showModal = true"
      >
        <span class="text-lg leading-none">+</span> Add Library
      </button>
    </div>

    <div v-if="libraries === null" class="text-gray-500">No libraries found.</div>
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="lib in libraries"
        :key="lib.id"
        class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm"
      >
        <h2 class="text-lg font-medium">{{ lib.name }}</h2>
        <p class="mt-1 text-sm text-gray-500">Schedule: {{ lib.cron }}</p>
      </div>
    </div>
  </div>

  <LibraryModal
    v-if="showModal"
    @cancel="showModal = false"
    @saved="onSaved"
  />
</template>