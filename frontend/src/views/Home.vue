<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ring } from 'ldrs'

import LibraryModal from '../components/LibraryModal.vue'
import SkiplistModal from '../components/SkiplistModal.vue'

interface Library {
  id: number
  name: string
  cron: string
  lastRun: string | undefined
  nextRun: string | undefined
}

const libraries = ref<Library[]>([])
const jobInProgress = ref(0)
const showModal = ref(false)
const editingId = ref<number | undefined>(undefined)
const showSkiplistModal = ref(false)
const skiplistLibraryId = ref<number | undefined>(undefined)

const dateTimeFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

async function fetchCurrentJob() {
    try {
    const response = await fetch('/api/runningJob')
      const data = await response.json()
    if (jobInProgress.value !== data.job) {
      jobInProgress.value = data.job
      fetchLibraries()
    }
    } catch (error) {
      console.error('Error fetching job in progress:', error)
    }
}

function startJobPolling() {
  fetchCurrentJob()
  setInterval(fetchCurrentJob, 10000)
}

async function fetchLibraries() {
  try {
    const response = await fetch('/api/libraries')
    libraries.value = await response.json()
  } catch (error) {
    console.error('Error fetching libraries:', error)
  }
}

onMounted(() => {
  ring.register()
  fetchLibraries()
  startJobPolling()
})

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
    libraries.value = libraries.value.filter((l) => l.id !== id)
  } catch (error) {
    console.error('Error deleting library:', error)
  }
}

async function startScan(id: number) {
  try {
    await fetch(`/api/run/${id}`, {
      method: 'POST',
    })
    fetchCurrentJob()
  } catch (error) {
    console.error('Error starting scan for library:', error)
  }
}

function onSaved() {
  showModal.value = false
  fetchLibraries()
}

function formatDate(dateString: string | undefined) {
  if (!dateString) return ''
  const date = new Date(dateString)
  if (Number.isNaN(date.getTime())) return dateString
  return dateTimeFormatter.format(date)
}

function getLibName(id: number) {
  const lib = libraries.value.find((l) => l.id === id)
  return lib ? lib.name : 'unknown library'
}
</script>

<template>
  <div class="p-6">
    <div class="mb-2 flex items-center justify-between">
      <h1 class="text-2xl font-semibold">Libraries</h1>
      <button
        class="flex items-center gap-1 rounded-lg bg-blue-600 px-3 py-2 text-sm text-white hover:bg-blue-700 cursor-pointer"
        @click="openCreate"
      >
        <span class="text-lg leading-none">+</span> Add Library
      </button>
    </div>

    <div v-if="jobInProgress !== 0" class="flex flex-col mb-4 bg-gray-300 p-4 rounded-lg">
      <h2 class="text-lg">Job in Progress:</h2>
      <span class="flex items-center gap-2">
        <l-ring size="12" stroke="2" bg-opacity="0" speed="2" color="black"></l-ring>
        <p>{{ getLibName(jobInProgress) }}</p>
      </span>
    </div>

    <div v-if="libraries === null || libraries.length === 0" class="text-gray-500">
      No libraries found.
    </div>
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="lib in libraries"
        :key="lib.id"
        class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm"
      >
        <h2 class="text-lg font-medium">{{ lib.name }}</h2>
        <p class="mt-1 text-sm text-gray-500">Schedule: {{ lib.cron }}</p>
        <p v-if="lib.lastRun" class="mt-1 text-sm text-gray-500">
          Last completed run: <time :datetime="lib.lastRun">{{ formatDate(lib.lastRun) }}</time>
        </p>
        <p v-if="lib.nextRun" class="mt-1 text-sm text-gray-500">
          Next run: <time :datetime="lib.nextRun">{{ formatDate(lib.nextRun) }}</time>
        </p>
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
          <button
            class="rounded-lg border border-green-300 px-3 py-1 text-sm text-green-600 hover:bg-green-50 cursor-pointer"
            @click="startScan(lib.id)"
          >
            Run
          </button>
        </div>
      </div>
    </div>
  </div>

  <LibraryModal v-if="showModal" :id="editingId" @cancel="showModal = false" @saved="onSaved" />

  <SkiplistModal
    v-if="showSkiplistModal && skiplistLibraryId !== undefined"
    :id="skiplistLibraryId"
    @close="showSkiplistModal = false"
  />
</template>
