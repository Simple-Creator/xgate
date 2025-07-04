<template>
  <el-row justify="center" align="middle" style="height:100vh;">
    <el-col :span="6">
      <el-card>
        <h2 style="text-align:center;">登录 xGate 堡垒机</h2>
        <el-form :model="form" @submit.prevent="onLogin">
          <el-form-item label-width="56px">
            <template #label><div style="width:100%;text-align:justify;font-weight:500;">用户名<span style='display:inline-block;width:100%;'></span></div></template>
            <el-input v-model="form.username" autocomplete="off" />
          </el-form-item>
          <el-form-item label-width="56px">
            <template #label><div style="width:100%;text-align:justify;font-weight:500;">密码<span style='display:inline-block;width:100%;'></span></div></template>
            <el-input v-model="form.password" type="password" autocomplete="off" @keyup.enter.native="onLogin" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="onLogin" style="width:100%;">登录</el-button>
          </el-form-item>
          <el-form-item>
            <el-link type="primary" @click="router.push('/register')" style="width: 100%; justify-content: center;">还没有账号？立即注册</el-link>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>
  </el-row>
</template>
<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../api'
import { useMainStore } from '../store'

const router = useRouter()
const store = useMainStore()
const form = reactive({ username: '', password: '' })

const onLogin = async () => {
  try {
    const res = await api.login(form)
    store.setToken(res.data.token, res.data.role, res.data.username)
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || '登录失败')
  }
}
</script> 