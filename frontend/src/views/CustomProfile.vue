<script setup lang="ts">
import * as monaco from 'monaco-editor'
import EditorWorker from 'monaco-editor/editor/editor.worker?worker'
import JsonWorker from 'monaco-editor/languages/features/json/json.worker?worker'
import { onBeforeUnmount, onMounted, ref } from 'vue'

self.MonacoEnvironment = {
  getWorker: (_, label) => (label === 'json' ? new JsonWorker() : new EditorWorker()),
}

const container = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | undefined
const profiles = ref('')
const editorValue = ref('')
const errorMessage = ref('')
let errorTimer: ReturnType<typeof setTimeout> | undefined

function showError(message: string) {
  errorMessage.value = message
  clearTimeout(errorTimer)
  errorTimer = setTimeout(() => (errorMessage.value = ''), 10000)
}

async function saveProfile() {
  const response = await fetch('/api/setCustomProfile', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: editorValue.value,
  })
  if (response.ok) {
    profiles.value = editorValue.value
  } else {
    const data = await response.json().catch(() => null)
    showError(data?.error ?? 'Failed to save profile')
  }
}

onMounted(() => {
  editor = monaco.editor.create(container.value!, {
    language: 'json',
    theme: 'vs-dark',
    automaticLayout: true,
  })

  editor.onDidChangeModelContent(() => {
    editorValue.value = editor?.getValue() ?? ''
  })

  fetch('/api/customProfile')
    .then((res) => res.text())
    .then((data) => {
      profiles.value = data
      editorValue.value = data
      editor?.setValue(data)
    })
    .catch((err) => {
      console.error('Failed to fetch profiles:', err)
    })
})

onBeforeUnmount(() => {
  clearTimeout(errorTimer)
  editor?.dispose()
})
</script>

<template>
  <div v-if="errorMessage" class="fixed top-3 left-1/2 -translate-x-1/2 rounded-xl w-60 z-20 bg-red-600 text-white px-4 py-2 flex justify-between items-center">
    <span>{{ errorMessage }}</span>
    <button @click="errorMessage = ''" class="hover:cursor-pointer">✕</button>
  </div>

  <div ref="container" class="h-full w-screen"></div>

  <div v-if="profiles !== editorValue" class="fixed bottom-2 right-2 p-2 z-10">
    <button @click="saveProfile" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 hover:cursor-pointer">Save</button>
  </div>
</template>