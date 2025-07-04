<template>
  <div class="terminal-root">
    <div ref="xterm" class="xterm-box"></div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { Terminal as XTerm } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import api from '../api'

const props = defineProps<{ connectionId: number }>()
const xterm = ref<HTMLElement | null>(null)
let term: XTerm | null = null
let socket: WebSocket | null = null
let fitAddon: FitAddon | null = null
let isDisconnected = false

const connectTerminal = () => {
  if (!xterm.value) return
  term = new XTerm({ fontSize: 14 })
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(xterm.value)
  fitAddon.fit()
  term.write('欢迎使用终端！\r\n')
  const wsUrl = api.getTerminalWSUrl(props.connectionId)
  socket = new WebSocket(wsUrl.replace('http', 'ws'))
  isDisconnected = false
  term.onData(data => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(data)
    } else if (isDisconnected && (data === '\r' || data === '\n')) {
      // 回车重连
      term?.write('\r\n[正在重连...]\r\n')
      term?.dispose()
      socket?.close()
      connectTerminal()
    }
  })
  socket.onmessage = e => {
    term?.write(e.data)
    term?.scrollToBottom()
  }
  socket.onclose = () => {
    term?.write('\r\n[连接已断开] 按回车重连\r\n')
    socket = null
    isDisconnected = true
  }
}

onMounted(() => {
  nextTick(() => {
    connectTerminal()
    window.addEventListener('resize', () => {
      fitAddon?.fit()
    })
  })
})
onBeforeUnmount(() => {
  term?.dispose()
  socket?.close()
})

watch(() => props.connectionId, () => {
  term?.dispose()
  socket?.close()
  connectTerminal()
})
</script>
<style scoped>
.terminal-root {
  height: 100%;
  width: 100%;
}
.xterm-box {
  height: 100%;
  width: 100%;
  background: #000;
}
</style> 