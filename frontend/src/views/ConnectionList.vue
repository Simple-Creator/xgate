<template>
  <el-card>
    <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
      <span style="font-size: 18px; font-weight: bold;">SSH连接管理</span>
      <el-button type="primary" @click="openAddDialog">新增连接</el-button>
    </div>

    <el-collapse v-model="activeCollapse" v-if="groupedConnections.length > 0">
      <el-collapse-item v-for="group in groupedConnections" :key="group.groupName" :title="group.groupName" :name="group.groupName">
        <div v-for="conn in group.connections" :key="conn.id" class="connection-item">
          <span>{{ conn.name }} ({{ conn.username }}@{{ conn.host }})</span>
          <div class="actions">
            <el-button size="small" type="primary" @click="onConnect(conn)">连接</el-button>
            <el-button size="small" @click="onEdit(conn)">编辑</el-button>
            <el-button size="small" type="danger" @click="onDelete(conn.id)">删除</el-button>
          </div>
        </div>
      </el-collapse-item>
    </el-collapse>
    <el-empty v-else description="暂无连接，请先新增一个"></el-empty>

    <!-- 新增/编辑弹窗 -->
    <el-dialog :title="editId ? '编辑连接' : '新增连接'" v-model="showAdd" @close="resetForm">
      <el-form :model="form" label-width="80px">
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">名称<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">分组<span style='display:inline-block;width:100%;'></span></div></template>
          <el-select v-model="form.group" filterable allow-create default-first-option placeholder="选择或创建分组" style="width: 100%;">
            <el-option v-for="item in groupOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">主机<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input v-model="form.host" />
        </el-form-item>
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">端口<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">用户名<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">密码<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">取消</el-button>
        <el-button type="primary" @click="onSave">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<style scoped>
.connection-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #eee;
}
.connection-item:last-child {
  border-bottom: none;
}
</style>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import api from '../api';

const connections = ref<any[]>([]); // This will now hold the grouped data
const groupOptions = ref<string[]>([]); // To hold the list of groups for the select input
const showAdd = ref(false);
const form = ref({
  id: null,
  name: '',
  group: '默认分组',
  host: '',
  port: 22,
  username: 'root',
  password: '',
});
const editId = ref<number | null>(null);
const activeCollapse = ref<string[]>([]); // To control which groups are expanded

const groupedConnections = computed(() => connections.value);

const emit = defineEmits(['connect']);

const load = async () => {
  try {
    const [connRes, groupRes] = await Promise.all([
      api.getConnections(),
      api.getGroups()
    ]);
    connections.value = connRes.data;
    groupOptions.value = groupRes.data;
    
    // Automatically expand all groups by default
    activeCollapse.value = connRes.data.map((g: any) => g.groupName);
  } catch (error) {
    ElMessage.error('加载连接列表失败');
    console.error(error);
  }
};
onMounted(load);

const resetForm = () => {
  form.value = {
    id: null,
    name: '',
    group: '默认分组',
    host: '',
    port: 22,
    username: 'root',
    password: '',
  };
  editId.value = null;
};

const openAddDialog = () => {
  resetForm();
  showAdd.value = true;
};

const onEdit = (row: any) => {
  form.value = { ...row };
  editId.value = row.id;
  showAdd.value = true;
};

const onDelete = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这个连接吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    });
    await api.deleteConnection(id);
    ElMessage.success('删除成功');
    load();
  } catch (error) {
    // User clicked cancel or an API error occurred
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
    }
  }
};

const onSave = async () => {
  if (!form.value.name || !form.value.host || !form.value.username) {
    ElMessage.warning('名称、主机和用户名不能为空');
    return;
  }
  try {
    if (editId.value) {
      await api.updateConnection(editId.value, form.value);
      ElMessage.success('更新成功');
    } else {
      await api.addConnection(form.value);
      ElMessage.success('新增成功');
    }
    showAdd.value = false;
    load();
  } catch (error) {
    ElMessage.error('保存失败');
  }
};

const onConnect = (row: any) => {
  emit('connect', row);
};
</script> 