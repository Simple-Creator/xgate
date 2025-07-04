<template>
  <div class="change-password">
    <el-card>
      <h2>修改密码</h2>
      <el-form :model="form" label-width="80px">
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">旧密码<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input v-model="form.oldPassword" type="password" />
        </el-form-item>
        <el-form-item label-width="56px">
          <template #label><div style="width:100%;text-align:justify;font-weight:500;">新密码<span style='display:inline-block;width:100%;'></span></div></template>
          <el-input v-model="form.newPassword" type="password" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSubmit">提交</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const form = ref({ oldPassword: '', newPassword: '' })

const handleSubmit = async () => {
  if (!form.value.oldPassword || !form.value.newPassword) {
    ElMessage.error('请输入旧密码和新密码')
    return
  }
  try {
    await api.changePassword(form.value)
    ElMessage.success('密码修改成功')
    form.value = { oldPassword: '', newPassword: '' }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || '修改失败')
  }
}
</script>

<style scoped>
.change-password {
  padding: 20px;
  max-width: 500px;
  margin: 0 auto;
}
</style> 