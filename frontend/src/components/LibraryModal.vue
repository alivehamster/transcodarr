<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import FiltersModal from './filters/FiltersModal.vue'

import type { OrderedFilter } from './filters/filters'

interface Library {
  id?: number
  name?: string
  cron?: string
  config?: {
    dirs?: string[]
    handbrakeCategory?: string
    handbrakeProfile?: string
    cacheDir?: string
    nodes?: any[]
    edges?: any[]
    order?: OrderedFilter[]
  }
}

const props = defineProps<{
  id?: number
}>()

const emit = defineEmits<{
  cancel: []
  saved: []
}>()

const showModal = ref(false)
const filterOrder = ref<OrderedFilter[] | undefined>(undefined)

const profilesByCategory = ref<Record<string, string[]>>({})
const categories = computed(() => Object.keys(profilesByCategory.value))

const name = ref('')
const cron = ref('')
const dirs = ref<string[]>([''])
const selectedCategory = ref('')
const profile = ref('')
const cacheDir = ref('')

const errors = ref({
  name: false,
  cron: false,
  dirs: false,
})

function validate() {
  errors.value.name = name.value.trim() === ''
  errors.value.cron = cron.value.trim() === ''
  errors.value.dirs = dirs.value.every(d => d.trim() === '')
  return !errors.value.name && !errors.value.cron && !errors.value.dirs
}

const availableProfiles = computed(() =>
  selectedCategory.value ? (profilesByCategory.value[selectedCategory.value] ?? []) : [],
)

function onCategoryChange() {
  profile.value = availableProfiles.value[0] ?? ''
}

function addDir() {
  dirs.value.push('')
}

function removeDir(index: number) {
  if (dirs.value.length > 1) {
    dirs.value.splice(index, 1)
  }
}

function handleFiltersSave(payload: OrderedFilter[]) {
  filterOrder.value = payload
  showModal.value = false
}

async function handleSave() {
  if (!validate()) return

  const payload: Library = {
    id: props.id,
    name: name.value,
    cron: cron.value,
    config: {
      dirs: dirs.value.filter(d => d.trim() !== ''),
      handbrakeCategory: selectedCategory.value,
      handbrakeProfile: profile.value,
      cacheDir: cacheDir.value.trim(),
      order: filterOrder.value ?? [],
    },
  }

  const url = props.id !== undefined ? '/api/editLibrary' : '/api/createLibrary'
  const method = props.id !== undefined ? 'PUT' : 'POST'

  try {
    const saveResponse = await fetch(url, {
      method: method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })

    if (!saveResponse.ok) {
      throw new Error(`Failed to save library: ${saveResponse.status}`)
    }

    emit('saved')
  } catch (error) {
    console.error('Error saving library:', error)
  }
}

onMounted(() => {

  if (props.id !== undefined) {
    fetch(`/api/library/${props.id}`)
      .then(response => response.json())
      .then((data: Library) => {
        name.value = data.name ?? ''
        cron.value = data.cron ?? ''
        dirs.value = data.config?.dirs ?? ['']
        selectedCategory.value = data.config?.handbrakeCategory ?? ''
        profile.value = data.config?.handbrakeProfile ?? ''
        cacheDir.value = data.config?.cacheDir ?? ''
        filterOrder.value = data.config?.order ?? []
      })
      .catch(error => {
        console.error('Error fetching library details:', error)
      })
  }

  fetch('/api/handbrakeProfiles')
    .then(response => response.json())
    .then((data: Record<string, string[]>) => {
      profilesByCategory.value = data
    })
    .catch(error => {
      console.error('Error fetching HandBrake profiles:', error)
    })
})
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl overflow-y-auto max-h-[90vh]">
      <h2 class="mb-5 text-xl font-semibold">
        {{ props.id === undefined ? 'Add Library' : 'Edit Library' }}
      </h2>

      <div class="space-y-4">
        <!-- Name -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Name</label>
          <input v-model="name" type="text" placeholder="My Library"
            class="w-full rounded-lg border px-3 py-2 text-sm focus:outline-none"
            :class="errors.name ? 'border-red-500 focus:border-red-500' : 'border-gray-300 focus:border-blue-500'" />
          <p v-if="errors.name" class="mt-1 text-xs text-red-500">Name is required.</p>
        </div>

        <!-- Cron -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Cron Schedule</label>
          <input v-model="cron" type="text" placeholder="0 2 * * *"
            class="w-full rounded-lg border px-3 py-2 text-sm font-mono focus:outline-none"
            :class="errors.cron ? 'border-red-500 focus:border-red-500' : 'border-gray-300 focus:border-blue-500'" />
          <p v-if="errors.cron" class="mt-1 text-xs text-red-500">Cron schedule is required.</p>
        </div>

        <!-- Directories -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Source Directories</label>
          <div class="space-y-2">
            <div v-for="(_, i) in dirs" :key="i" class="flex gap-2">
              <input v-model="dirs[i]" type="text" placeholder="/path/to/media"
                class="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
              <button type="button" :disabled="dirs.length === 1"
                class="rounded-lg border border-gray-300 px-2 py-1 text-gray-500 hover:bg-gray-100 disabled:opacity-30 cursor-pointer"
                @click="removeDir(i)">
                ✕
              </button>
            </div>
          </div>
          <p v-if="errors.dirs" class="mt-1 text-xs text-red-500">At least one directory is required.</p>
          <button type="button" class="mt-2 text-sm text-blue-600 hover:underline cursor-pointer" @click="addDir">
            + Add directory
          </button>
        </div>

        <!-- Cache Directory -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Transcode Cache Directory</label>
          <input v-model="cacheDir" type="text" placeholder="/path/to/cache"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
          <p class="mt-1 text-xs text-gray-500">Leave blank to keep temporary files next to the source video.</p>
        </div>

        <!-- HandBrake Category -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Category</label>
          <select v-model="selectedCategory"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none cursor-pointer"
            @change="onCategoryChange">
            <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
          </select>
        </div>

        <!-- HandBrake Profile -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">HandBrake Profile</label>
          <select v-model="profile"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none cursor-pointer">
            <option v-for="p in availableProfiles" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>

        <div>
          <button class="w-full h-10 border border-gray-300 rounded-lg cursor-pointer hover:bg-gray-100"
            @click="showModal = true">Filters</button>
        </div>
      </div>



      <!-- Actions -->
      <div class="mt-6 flex justify-end gap-3">
        <button type="button"
          class="rounded-lg border border-gray-300 px-4 py-2 text-sm hover:bg-gray-100 cursor-pointer"
          @click="emit('cancel')">
          Cancel
        </button>
        <button type="button"
          class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 cursor-pointer"
          @click="handleSave">
          Save
        </button>
      </div>
    </div>
  </div>

  <FiltersModal
    v-if="showModal"
    :order="filterOrder"
    @cancel="showModal = false"
    @save="handleFiltersSave"
  />
</template>
