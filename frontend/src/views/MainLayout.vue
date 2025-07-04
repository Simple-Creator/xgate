<template>
  <splitpanes class="default-theme" style="height: 100vh;">
    <pane v-if="showAside" min-size="15" size="20" max-size="30">
      <div style="display:flex; flex-direction: column; height: 100%; background:#f5f7fa; position: relative;">
        <ConnectionList @connect="onConnect" style="flex: 1; overflow-y: auto;"/>
        <div class="aside-bottom">
          <el-dropdown>
            <span class="el-dropdown-link">
              {{ store.username }}
              <el-icon class="el-icon--right el-icon-arrow-down"></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push('/changepwd')">修改密码</el-dropdown-item>
                <el-dropdown-item v-if="store.isAdmin" @click="router.push('/users')">用户管理</el-dropdown-item>
                <el-dropdown-item divided @click="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        <el-button class="aside-collapse" @click="toggleAside" icon="el-icon-arrow-left" circle />
      </div>
    </pane>
    <pane>
      <div style="position: relative; height: 100%;">
        <el-button v-if="!showAside" class="aside-expand" @click="toggleAside" icon="el-icon-arrow-right" circle />
        <router-view v-if="router.currentRoute.value.path !== '/'" />
          <template v-else>
            <el-tabs v-if="tabs.length" v-model="activeTab" type="card" @tab-remove="removeTab">
              <el-tab-pane
                v-for="tab in tabs"
                :key="tab.name"
                :label="tab.label"
                :name="tab.name"
                closable>
                <template v-if="tab.name.startsWith('terminal-')">
                  <splitpanes style="height:calc(100vh - 40px);">
                    <pane :size="terminalMaximized ? 100 : (fileMaximized ? 0 : 50)" min-size="0">
                      <div style="height:100%;display:flex;flex-direction:column;">
                        <div style="padding:4px 8px;display:flex;align-items:center;justify-content:space-between;background:#f6f8fa;">
                          <span>终端</span>
                          <el-button size="small" @click="toggleTerminalMax" icon="el-icon-full-screen">{{ terminalMaximized ? '还原' : '最大化' }}</el-button>
                        </div>
                        <div style="flex:1;">
                          <Terminal :connectionId="tab.props.connectionId" />
                        </div>
                      </div>
                    </pane>
                    <pane :size="fileMaximized ? 100 : (terminalMaximized ? 0 : 50)" min-size="0">
                      <div style="height:100%;display:flex;flex-direction:column;">
                        <div style="padding:4px 8px;display:flex;align-items:center;justify-content:space-between;background:#f6f8fa;">
                          <span>文件管理</span>
                          <el-button size="small" @click="toggleFileMax" icon="el-icon-full-screen">{{ fileMaximized ? '还原' : '最大化' }}</el-button>
                        </div>
                        <div style="flex:1;">
                          <FileManager :connectionId="tab.props.connectionId" />
                        </div>
                      </div>
                    </pane>
                  </splitpanes>
                </template>
                <template v-else>
                  <component :is="tab.component" v-bind="tab.props" />
                </template>
              </el-tab-pane>
            </el-tabs>
            <div v-else class="empty-hint">请选择左侧连接以开始</div>
          </template>
      </div>
    </pane>
  </splitpanes>
</template>

<script setup lang="ts">
import { ref, markRaw } from 'vue'
import { useRouter } from 'vue-router'
import { useMainStore } from '../store'
import ConnectionList from './ConnectionList.vue'
import Terminal from '../components/Terminal.vue'
import FileManager from '../components/FileManager.vue'
import { Splitpanes, Pane } from 'splitpanes'
import 'splitpanes/dist/splitpanes.css'

interface Tab {
  name: string;
  label: string;
  component: any;
  props: Record<string, any>;
}

const store = useMainStore()
const router = useRouter()

const showAside = ref(true)
const toggleAside = () => {
  showAside.value = !showAside.value
}

const tabs = ref<Tab[]>([])
const activeTab = ref('')

const terminalMaximized = ref(false)
const fileMaximized = ref(false)

const toggleTerminalMax = () => {
  terminalMaximized.value = !terminalMaximized.value
  if (terminalMaximized.value) fileMaximized.value = false
}
const toggleFileMax = () => {
  fileMaximized.value = !fileMaximized.value
  if (fileMaximized.value) terminalMaximized.value = false
}

const removeTab = (name: string) => {
  tabs.value = tabs.value.filter(tab => tab.name !== name)
  if (tabs.value.length > 0) {
    activeTab.value = tabs.value[tabs.value.length - 1].name
  }
}
const logout = () => {
  store.logout()
  router.push('/login')
}
const onConnect = (conn: any) => {
  const tabName = `terminal-${conn.id}`
  if (!tabs.value.find(t => t.name === tabName)) {
    tabs.value.push({
      name: tabName,
      label: `终端-${conn.name}`,
      component: markRaw(Terminal),
      props: { connectionId: conn.id }
    })
  }
  activeTab.value = tabName
}
</script>

<style>
/* splitpanes antd theme */
.splitpanes.default-theme .splitpanes__splitter {
  background-color: #f5f7fa;
  box-sizing: border-box;
  position: relative;
  flex-shrink: 0;
}
.splitpanes.default-theme .splitpanes__splitter:before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  transition: background-color 0.3s;
  background-color: rgba(0, 0, 0, 0.15);
  opacity: 1;
  z-index: 1;
}
.splitpanes.default-theme .splitpanes__splitter:hover:before {
  background-color: rgba(0, 0, 0, 0.25);
}
.splitpanes.default-theme.splitpanes--vertical > .splitpanes__splitter,
.splitpanes.default-theme .splitpanes__splitter.splitpanes__splitter--vertical {
  width: 7px;
}
.splitpanes.default-theme.splitpanes--vertical > .splitpanes__splitter:before,
.splitpanes.default-theme .splitpanes__splitter.splitpanes__splitter--vertical:before {
  left: 2px;
  width: 1px;
  height: 100%;
}

.aside-bottom {
  border-top: 1px solid #e0e2e4;
  padding: 16px 0;
  text-align: center;
}
.el-dropdown-link {
  cursor: pointer;
  color: #409EFF;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.empty-hint {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #bbb;
  font-size: 18px;
}

.aside-collapse,
.aside-expand {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  z-index: 10;
  border: 1px solid #dcdfe6;
  background-color: #f4f4f5;
  color: #777;
  transition: all 0.3s;
}
.aside-collapse:hover,
.aside-expand:hover {
  background-color: #409eff;
  color: #fff;
  border-color: #409eff;
}

.aside-collapse {
  right: -15px;
}
.aside-expand {
  left: -15px;
}
</style> 