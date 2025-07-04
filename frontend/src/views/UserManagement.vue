<template>
  <div class="user-management">
    <div class="header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="openAddDialog">新增用户</el-button>
    </div>

    <el-table :data="users" style="width: 100%">
      <el-table-column prop="id" label="ID" width="180" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="role" label="角色" />
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button size="small" @click="onEdit(row)">编辑</el-button>
          <el-button size="small" type="warning" @click="onResetPassword(row)">重置密码</el-button>
          <el-button size="small" type="danger" @click="onDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showAdd" :title="editId ? '编辑用户' : '新增用户'">
      <el-form :model="form" label-width="80px">
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">用户名<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input v-model="form.username" :disabled="!!editId" />
        </el-form-item>
        <el-form-item v-if="!editId" label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">密码<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input v-model="form.password" type="password" />
        </el-form-item>
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">角色<span style='display:inline-block;width:100%;'></span></div></template>
          <el-select v-model="form.role">
            <el-option label="管理员" value="admin" />
            <el-option label="用户" value="user" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">取消</el-button>
        <el-button type="primary" @click="onSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import api from '../api';

interface User {
  id: number;
  username: string;
  role: 'user' | 'admin';
}

const users = ref<User[]>([]);
const showAdd = ref(false);
const form = ref({
  id: null as number | null,
  username: '',
  password: '',
  role: 'user' as 'user' | 'admin',
});
const editId = ref<number | null>(null);

const load = async () => {
  try {
    const res = await api.listUsers();
    users.value = res.data.users;
  } catch (error) {
    ElMessage.error('获取用户列表失败');
  }
};

onMounted(load);

const openAddDialog = () => {
  editId.value = null;
  form.value = { id: null, username: '', password: '', role: 'user' };
  showAdd.value = true;
};

const onEdit = (user: User) => {
  editId.value = user.id;
  form.value = { ...user, password: '' };
  showAdd.value = true;
};

const onDelete = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定删除该用户吗？此操作不可撤销。', '警告', {
      type: 'warning',
    });
    await api.deleteUser(id);
    ElMessage.success('删除成功');
    await load();
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败');
    }
  }
};

const onResetPassword = async (user: User) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入新密码', `重置用户 ${user.username} 的密码`, {
      inputType: 'password',
    });
    if (value) {
      await api.resetPassword(user.id, { password: value });
      ElMessage.success('密码重置成功');
    }
  } catch (e) {
     if (e !== 'cancel') {
      ElMessage.error('密码重置失败');
    }
  }
};

const onSave = async () => {
  try {
    if (editId.value) {
      await api.updateUser(editId.value, { role: form.value.role });
      ElMessage.success('更新成功');
    } else {
      if (!form.value.password) {
        ElMessage.warning('新增用户时必须填写密码');
        return;
      }
      await api.addUser(form.value);
      ElMessage.success('新增成功');
    }
    showAdd.value = false;
    await load();
  } catch (error) {
    ElMessage.error('保存失败');
  }
};
</script>

<style scoped>
.user-management {
  padding: 20px;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
</style> 