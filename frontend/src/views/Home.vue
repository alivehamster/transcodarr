<script setup lang="ts">
import { onMounted, ref } from 'vue'
import LibraryModal from '../components/LibraryModal.vue'
import SkiplistModal from '../components/SkiplistModal.vue'

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
const editingId = ref<number | undefined>(undefined)
const showSkiplistModal = ref(false)
const skiplistLibraryId = ref<number | undefined>(undefined)

async function fetchLibraries() {
  try {
    const response = await fetch('/api/libraries')
    libraries.value = await response.json()
  } catch (error) {
    console.error('Error fetching libraries:', error)
  }
}

onMounted(fetchLibraries)

function openCreate() {
  editingId.value = undefined
  showModal.value = true
}

function openEdit(id: number) {
  editingId.value = id
  showModal.value = true
}

function openSkiplist(id: number) {
  skiplistLibraryId.value = id
  showSkiplistModal.value = true
}

async function deleteLibrary(id: number) {
  try {
    await fetch(`/api/deleteLibrary/${id}`, {
      method: 'DELETE',
    })
    libraries.value = libraries.value.filter(l => l.id !== id)
  } catch (error) {
    console.error('Error deleting library:', error)
  }
}

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
        @click="openCreate"
      >
        <span class="text-lg leading-none">+</span> Add Library
      </button>
    </div>

    <div v-if="libraries === null || libraries.length === 0" class="text-gray-500">No libraries found.</div>
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="lib in libraries"
        :key="lib.id"
        class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm"
      >
        <h2 class="text-lg font-medium">{{ lib.name }}</h2>
        <p class="mt-1 text-sm text-gray-500">Schedule: {{ lib.cron }}</p>
        <div class="mt-3 flex gap-2">
          <button
            class="rounded-lg border border-gray-300 px-3 py-1 text-sm hover:bg-gray-100 cursor-pointer"
            @click="openEdit(lib.id)"
          >
            Edit
          </button>
          <button
            class="rounded-lg border border-gray-300 px-3 py-1 text-sm hover:bg-gray-100 cursor-pointer"
            @click="openSkiplist(lib.id)"
          >
            Skip List
          </button>
          <button
            class="rounded-lg border border-red-300 px-3 py-1 text-sm text-red-600 hover:bg-red-50 cursor-pointer"
            @click="deleteLibrary(lib.id)"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>

  <LibraryModal
    v-if="showModal"
    :id="editingId"
    @cancel="showModal = false"
    @saved="onSaved"
  />

  <SkiplistModal
    v-if="showSkiplistModal && skiplistLibraryId !== undefined"
    :id="skiplistLibraryId"
    @close="showSkiplistModal = false"
  />
</template>