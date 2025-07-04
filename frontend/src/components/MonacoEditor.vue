<template>
  <div ref="container" :style="{height, width: '100%'}"></div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import * as monaco from 'monaco-editor'

const props = defineProps<{
  modelValue: string
  language: string
  height?: string
}>()

const emit = defineEmits(['update:modelValue'])

const container = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null

onMounted(() => {
  editor = monaco.editor.create(container.value!, {
    value: props.modelValue,
    language: props.language,
    automaticLayout: true,
    theme: 'vs-dark'
  })
  editor.onDidChangeModelContent(() => {
    emit('update:modelValue', editor!.getValue())
  })
})

watch(() => props.modelValue, (val) => {
  if (editor && editor.getValue() !== val) {
    editor.setValue(val)
  }
})

watch(() => props.language, (lang) => {
  if (editor) {
    monaco.editor.setModelLanguage(editor.getModel()!, lang)
  }
})

onBeforeUnmount(() => {
  editor?.dispose()
})
</script> 