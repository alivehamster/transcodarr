import { fileURLToPath, URL } from 'node:url'

import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

// Stub every monaco language except JSON (languages/features/json is kept)
const monacoJsonOnly = (): Plugin => ({
  name: 'monaco-json-only',
  enforce: 'pre',
  load(id) {
    if (/monaco-editor\/esm\/vs\/languages\/(definitions\/|features\/(css|html|typescript)\/)/.test(id)) {
      return 'export {}'
    }
  },
})

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    monacoJsonOnly(),
    vue(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  build: {
    // This is monaco's fault
    chunkSizeWarningLimit: 4100,
  },
})
