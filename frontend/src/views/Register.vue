<template>
  <el-row justify="center" align="middle" style="height: 100vh">
    <el-col :span="6">
      <el-card>
        <template #header>
          <div style="text-align: center">
            <h2>注册新用户</h2>
          </div>
        </template>
        <div style="text-align: center; color: #e6a23c; margin-bottom: 20px;">
          第一个注册的用户将成为系统管理员
        </div>
        <el-form :model="form" @submit.prevent="onRegister">
          <el-form-item label-width="56px">
            <template #label><div style="width:100%;text-align:justify;font-weight:500;">用户名<span style='display:inline-block;width:100%;'></span></div></template>
            <el-input v-model="form.username" autocomplete="username" />
          </el-form-item>
          <el-form-item label-width="56px">
            <template #label><div style="width:100%;text-align:justify;font-weight:500;">密码<span style='display:inline-block;width:100%;'></span></div></template>
            <el-input v-model="form.password" type="password" autocomplete="new-password" show-password />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="onRegister" style="width: 100%">注册</el-button>
          </el-form-item>
          <el-form-item>
            <el-link type="primary" @click="router.push('/login')" style="width: 100%; justify-content: center;">已有账号？前往登录</el-link>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup lang="ts">
import { reactive } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import api from '../api';

const router = useRouter();
const form = reactive({
  username: '',
  password: '',
});

const onRegister = async () => {
  if (!form.username || !form.password) {
    ElMessage.warning('用户名和密码不能为空');
    return;
  }
  try {
    await api.register(form);
    ElMessage.success('注册成功！您现在可以登录了。');
    router.push('/login');
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || '注册失败');
  }
};
</script> 