import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import monacoEditorPluginImport from 'vite-plugin-monaco-editor'

const monacoEditorPlugin = (monacoEditorPluginImport as any).default || monacoEditorPluginImport as any

export default defineConfig({
  plugins: [
    vue(),
    monacoEditorPlugin({
      languageWorkers: ['editorWorkerService', 'typescript', 'json', 'css', 'html'],
    }),
  ],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        ws: true
      }
    }
  }
}) 