<template>
  <div class="file-manager">
    <div class="path-breadcrumb" style="display: flex; align-items: center;">
      <el-button
        size="small"
        @click="toggleShowHidden"
        style="margin-right:8px;vertical-align:middle;"
        title="显示/隐藏隐藏文件"
      >
        <el-icon style="font-size: 18px;">
          <component :is="showHidden ? View : Hide" />
        </el-icon>
      </el-button>
      <el-button
        size="small"
        :icon="Edit"
        @click="startEditPath"
        style="margin-right:8px;"
        circle
        title="编辑路径"
      />
      <template v-if="!editingPath">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item v-for="crumb in breadcrumbs" :key="crumb.path" @click="handleBreadcrumbClick(crumb.path)">
            <a href="#">{{ crumb.name }}</a>
          </el-breadcrumb-item>
        </el-breadcrumb>
      </template>
      <template v-else>
        <el-input v-model="editPathValue" size="small" style="width: 400px;" @keyup.enter="confirmEditPath" @blur="cancelEditPath" />
      </template>
    </div>
    <div style="margin: 8px 0;">
      <el-upload
        :show-file-list="false"
        :before-upload="beforeUpload"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        :auto-upload="true"
        :action="uploadAction"
        :data="uploadData"
      >
        <el-button size="small" type="primary">上传文件</el-button>
      </el-upload>
    </div>
    <div class="file-table-wrapper">
      <el-table
        ref="elTableRef"
        :data="files"
        style="width: 100%"
        height="100%"
      >
        <el-table-column label="Name" sortable prop="name">
          <template #default="{ row }">
            <div v-if="editingItem?.originalName !== row.name" class="file-item">
              <i :class="row.isDir ? 'el-icon-folder' : 'el-icon-document'" />
              <span
                v-if="row.isDir"
                class="file-link"
                @click="enterDir(row)"
              >{{ row.name }}</span>
              <span v-else @click="handleFileClick(row)" style="cursor:pointer;">{{ row.name }}</span>
            </div>
            <div v-else-if="editingItem">
              <el-input
                v-model="editingItem.name"
                class="rename-input"
                :data-name="row.name"
                @blur="rename"
                @keyup.enter="rename"
                @keyup.esc="cancelRename"
                size="small"
              />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Size" prop="size" width="120" />
        <el-table-column label="Modified" prop="modTime" width="200" />
        <el-table-column label="Actions" width="150">
          <template #default="{ row }">
            <el-button size="small" @click.stop="startRename(row)">Rename</el-button>
            <el-button size="small" type="danger" @click.stop="deleteFile(row)">Delete</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="loadingMore" style="text-align:center;padding:8px;">加载中...</div>
      <div v-else-if="rawFiles.length < total" style="text-align:center;padding:8px;cursor:pointer;color:#409EFF;" @click="loadMore">加载更多</div>
    </div>

    <el-dialog v-model="showEdit" title="Edit File" width="70%">
      <MonacoEditor
        v-model="editContent"
        :language="editLanguage"
        height="500px"
      />
      <template #footer>
        <el-button @click="showEdit = false">Cancel</el-button>
        <el-button type="primary" @click="saveFile">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick, onMounted, onBeforeUnmount } from 'vue';
import api from '../api';
import { ElMessage, ElMessageBox } from 'element-plus';
import MonacoEditor from './MonacoEditor.vue';
import { Edit, View, Hide } from '@element-plus/icons-vue';

interface FileItem {
  name: string;
  isDir: boolean;
  size: number;
  modTime: string;
  originalName?: string;
}

// 文件扩展名到 Monaco 语言的简单映射
const ext2lang: Record<string, string> = {
  js: 'javascript', ts: 'typescript', py: 'python', java: 'java', c: 'c', cpp: 'cpp', h: 'cpp',
  json: 'json', md: 'markdown', vue: 'vue', html: 'html', css: 'css', scss: 'scss', less: 'less',
  go: 'go', rs: 'rust', sh: 'shell', yaml: 'yaml', yml: 'yaml', xml: 'xml', php: 'php',
  txt: 'plaintext', log: 'plaintext', conf: 'ini', ini: 'ini', sql: 'sql', rb: 'ruby', swift: 'swift',
  kt: 'kotlin', dart: 'dart', cs: 'csharp', dockerfile: 'dockerfile', makefile: 'makefile',
};

const props = defineProps<{ connectionId?: number }>();

const files = ref<FileItem[]>([]);
const rawFiles = ref<FileItem[]>([]);
const currentPath = ref('/');
const editingItem = ref<FileItem | null>(null);
const total = ref(0);
const offset = ref(0);
const limit = ref(10);
const loadingMore = ref(false);
const showHidden = ref(false);

const showEdit = ref(false);
const editFile = ref<{ path: string; content: string } | null>(null);
const editContent = ref('');
const editLanguage = ref('text');

const editingPath = ref(false);
const editPathValue = ref('');

const uploading = ref(false)

const uploadAction = computed(() => `/api/files/${props.connectionId}/upload`)
const uploadData = computed(() => ({ path: currentPath.value }))

const getLanguageByFilename = (filename: string) => {
  const ext = filename.split('.').pop()?.toLowerCase() || '';
  return ext2lang[ext] || 'plaintext';
};

const loadFiles = async (reset = true) => {
  if (!props.connectionId || !currentPath.value) return;
  try {
    const res = await api.listFiles(props.connectionId, currentPath.value, offset.value, limit.value);
    const newFiles = (res.data.files || []).map((f: any) => ({
      ...f,
      isDir: f.isDir !== undefined ? f.isDir : f.type === 'dir'
    }));
    if (reset) {
      rawFiles.value = newFiles;
    } else {
      rawFiles.value = rawFiles.value.concat(newFiles);
    }
    total.value = res.data.total || 0;
    updateFilesView();
  } catch (error: any) {
    ElMessage.error('Failed to load files: ' + (error && error.message ? error.message : ''));
  }
};

watch(
  () => props.connectionId,
  (newId) => {
    if (newId) {
      currentPath.value = '/';
      loadFiles();
    }
  },
  { immediate: true }
);

onMounted(async () => {
  if (props.connectionId !== undefined) {
    try {
      const res = await api.getHomeDir(props.connectionId);
      currentPath.value = res.data.home.endsWith('/') ? res.data.home : res.data.home + '/';
    } catch {
      currentPath.value = '/';
    }
    await loadFiles();
    // 初次加载后主动触发一次 loadMore，确保内容未撑满时能自动加载
    nextTick(() => {
      if (rawFiles.value.length < total.value) {
        loadMore();
      }
    });
  }
});

const breadcrumbs = computed(() => {
  const segs = currentPath.value.split('/').filter(Boolean);
  const result = [{ name: 'Home', path: '/' }];
  segs.forEach((seg, i) => {
    result.push({ name: seg, path: `/${segs.slice(0, i + 1).join('/')}` });
  });
  return result;
});

const handleBreadcrumbClick = (path: string) => {
  currentPath.value = path;
  offset.value = 0;
  loadFiles(true);
};

const handleFileClick = async (file: FileItem) => {
  if (file.isDir) {
    return;
  } else {
    const filePath = joinPath(currentPath.value, file.name);
    const res = await api.readFile(props.connectionId!, filePath);
    editFile.value = { path: filePath, content: res.data.content };
    // base64 解码
    try {
      editContent.value = atob(res.data.content);
    } catch {
      editContent.value = '';
    }
    editLanguage.value = getLanguageByFilename(file.name);
    showEdit.value = true;
    nextTick(() => {
      // 仅触发 DOM 更新，monaco-editor-vue3 会自动适配
    });
  };
};

const startRename = (file: FileItem) => {
  editingItem.value = { ...file, originalName: file.name };
  nextTick(() => {
    const inputEl = document.querySelector(`.rename-input[data-name="${file.name}"]`);
    if (inputEl) {
      (inputEl as HTMLInputElement).focus();
    }
  });
};

const cancelRename = () => {
  editingItem.value = null;
};

const rename = async () => {
  if (editingItem.value && editingItem.value.originalName) {
    try {
      const oldPath = joinPath(currentPath.value, editingItem.value.originalName);
      const newPath = joinPath(currentPath.value, editingItem.value.name);
      await api.renameFile(props.connectionId!, oldPath, newPath);
      ElMessage.success('Renamed successfully');
      editingItem.value = null;
      loadFiles();
    } catch (error) {
      ElMessage.error('Failed to rename');
    }
  }
};

const saveFile = async () => {
  if (editFile.value) {
    try {
      await api.editFile(props.connectionId!, editFile.value.path, editContent.value);
      ElMessage.success('File saved');
      showEdit.value = false;
    } catch (error) {
      ElMessage.error('Failed to save file');
    }
  }
};

const deleteFile = async (file: FileItem) => {
  try {
    await ElMessageBox.confirm(`Are you sure you want to delete ${file.name}?`, 'Warning', {
      type: 'warning',
    });
    const filePath = joinPath(currentPath.value, file.name);
    await api.deleteFile(props.connectionId!, filePath);
    ElMessage.success('Deleted successfully');
    loadFiles();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Failed to delete');
    }
  }
};

const startEditPath = () => {
  editingPath.value = true;
  editPathValue.value = currentPath.value;
  nextTick(() => {
    const input = document.querySelector('.path-breadcrumb input');
    if (input) (input as HTMLInputElement).focus();
  });
};

const confirmEditPath = () => {
  currentPath.value = editPathValue.value.endsWith('/') ? editPathValue.value : editPathValue.value + '/';
  editingPath.value = false;
  loadFiles();
};

const cancelEditPath = () => {
  editingPath.value = false;
};

const enterDir = (row: FileItem) => {
  if (!row.isDir) return;
  currentPath.value = joinPath(currentPath.value, row.name);
  offset.value = 0;
  loadFiles(true);
};

function beforeUpload() {
  return !uploading.value
}

function handleUploadSuccess() {
  ElMessage.success('上传成功')
  loadFiles()
}

function handleUploadError() {
  ElMessage.error('上传失败')
}

const elTableRef = ref();
let scrollTimer: any = null;

function stopScrollTimer() {
  if (scrollTimer) {
    clearInterval(scrollTimer);
    scrollTimer = null;
  }
}

onMounted(() => {
  nextTick(() => {
    const tableEl = elTableRef.value?.$el;
    let tableBody = tableEl?.querySelector('.el-scrollbar__wrap');
    if (!tableBody) {
      tableBody = tableEl?.querySelector('.el-table__body-wrapper');
    }
    if (tableBody) {
      let lastScrollTop = tableBody.scrollTop;
      scrollTimer = setInterval(() => {
        // 自动加载更多
        if (rawFiles.value.length < total.value && !loadingMore.value) {
          if (tableBody.scrollHeight <= tableBody.clientHeight + 10) {
            loadMore();
          }
        }
        // 监听滚动到底部
        if (tableBody.scrollTop !== lastScrollTop) {
          lastScrollTop = tableBody.scrollTop;
          if (tableBody.scrollTop + tableBody.clientHeight >= tableBody.scrollHeight - 10) {
            loadMore();
          }
        }
        // 修复：所有文件已加载，自动停止定时器
        if (rawFiles.value.length >= total.value) {
          stopScrollTimer();
        }
      }, 100);
    }
  });
});

onBeforeUnmount(() => {
  stopScrollTimer();
});

const loadMore = async () => {
  if (loadingMore.value) return;
  if (rawFiles.value.length >= total.value) return;
  loadingMore.value = true;
  offset.value += limit.value;
  await loadFiles(false);
  loadingMore.value = false;
};

if (typeof window !== 'undefined') {
  (window as any).loadMore = loadMore;
  (window as any).files = files;
}

const toggleShowHidden = () => {
  showHidden.value = !showHidden.value;
  updateFilesView();
};

function updateFilesView() {
  // 目录优先、同类按名称排序
  let arr = [...rawFiles.value];
  arr = arr.filter(f => showHidden.value || !f.name.startsWith('.'));
  arr.sort((a, b) => {
    if (a.isDir && !b.isDir) return -1;
    if (!a.isDir && b.isDir) return 1;
    return a.name.localeCompare(b.name);
  });
  files.value = arr;
}

// 统一拼接文件路径
function joinPath(dir: string, name: string): string {
  if (!dir.endsWith('/')) dir += '/';
  return dir + name;
}

</script>

<style scoped>
.file-manager {
  padding: 10px;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.path-breadcrumb {
  margin-bottom: 10px;
  cursor: pointer;
}
.file-item {
  display: flex;
  align-items: center;
  gap: 5px;
}
.file-link {
  color: #409EFF !important;
  cursor: pointer !important;
  transition: color 0.2s;
}
.file-link:hover {
  color: #1867c0 !important;
  text-decoration: underline !important;
}
.file-table-wrapper {
  flex: 1 1 0%;
  min-height: 0;
  height: 0;
  max-height: 100%;
  overflow: auto !important;
  background: #fff;
}
</style> 