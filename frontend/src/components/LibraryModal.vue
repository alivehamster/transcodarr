<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'

interface Library {
  id?: number
  name?: string
  cron?: string
  config?: {
    dirs?: string[]
    profile?: string
  }
}

const props = defineProps<{
  id?: number
}>()

const emit = defineEmits<{
  cancel: []
  saved: []
}>()

const profilesByCategory = ref<Record<string, string[]>>({})
const categories = computed(() => Object.keys(profilesByCategory.value))

const name = ref('')
const cron = ref('')
const dirs = ref<string[]>([''])
const selectedCategory = ref('')
const profile = ref('')

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

async function handleSave() {
  const payload: Library = {
  }

  const url = props.id !== undefined ? '/api/editLibrary' : '/api/createLibrary'

  try {
    await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    emit('saved')
  } catch (error) {
    console.error('Error saving library:', error)
  }
}

onMounted(() => {
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
    <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
      <h2 class="mb-5 text-xl font-semibold">
        {{ props.id === undefined ? 'Add Library' : 'Edit Library' }}
      </h2>

      <div class="space-y-4">
        <!-- Name -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Name</label>
          <input v-model="name" type="text" placeholder="My Library"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
        </div>

        <!-- Cron -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Cron Schedule</label>
          <input v-model="cron" type="text" placeholder="0 2 * * *"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none" />
        </div>

        <!-- Directories -->
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Directories</label>
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
          <button type="button" class="mt-2 text-sm text-blue-600 hover:underline cursor-pointer" @click="addDir">
            + Add directory
          </button>
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
      </div>

      <!-- Actions -->
      <div class="mt-6 flex justify-end gap-3">
        <button type="button" class="rounded-lg border border-gray-300 px-4 py-2 text-sm hover:bg-gray-100 cursor-pointer"
          @click="emit('cancel')">
          Cancel
        </button>
        <button type="button" class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 cursor-pointer"
          @click="handleSave">
          Save
        </button>
      </div>
    </div>
  </div>
</template>
