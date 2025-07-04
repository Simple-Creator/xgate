import axios from 'axios'
import { useMainStore } from '../store'
import router from '../router'

const instance = axios.create({
  baseURL: '/api',
  timeout: 20000
})

instance.interceptors.request.use(config => {
  const store = useMainStore()
  if (store.token) {
    config.headers = config.headers || {}
    config.headers['Authorization'] = 'Bearer ' + store.token
  }
  return config
})

instance.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // 自动登出逻辑
    if (error.response && error.response.status === 401) {
      const store = useMainStore();
      store.logout();
      if (router.currentRoute.value.path !== '/login') {
        router.replace('/login');
      }
    }
    return Promise.reject(error);
  }
)

export default {
  // 用户相关
  login(data: any) {
    return instance.post('/login', data)
  },
  register(data: any) {
    return instance.post('/register', data)
  },
  // 分组
  getGroups() {
    return instance.get('/groups');
  },
  // 连接管理
  getConnections() {
    return instance.get('/connections')
  },
  addConnection(data: any) {
    return instance.post('/connections', data)
  },
  updateConnection(id: number, data: any) {
    return instance.put(`/connections/${id}`, data)
  },
  deleteConnection(id: number) {
    return instance.delete(`/connections/${id}`)
  },
  testConnection(data: any) {
    return instance.post('/connections/test', data)
  },
  // 终端
  getTerminalWSUrl(connectionId: number) {
    const store = useMainStore()
    // 根据页面协议自动选择 ws 或 wss
    const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    return `${wsProtocol}://${window.location.host}/api/terminal/${connectionId}?token=${store.token}`
  },
  // 文件管理
  listFiles(connectionId: number, path: string, offset = 0, limit = 100) {
    return instance.get(`/files/${connectionId}`, { params: { path, offset, limit } })
  },
  uploadFile(connectionId: number, path: string, file: File) {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('path', path)
    return instance.post(`/files/${connectionId}/upload`, formData)
  },
  downloadFile(connectionId: number, path: string) {
    return instance.get(`/files/${connectionId}/download`, { params: { path }, responseType: 'blob' })
  },
  deleteFile(connectionId: number, path: string) {
    return instance.delete(`/files/${connectionId}`, { params: { path } })
  },
  renameFile(connectionId: number, oldPath: string, newPath: string) {
    return instance.put(`/files/${connectionId}/rename`, { oldPath, newPath })
  },
  editFile(connectionId: number, path: string, content: string) {
    return instance.put(`/files/${connectionId}/edit`, { path, content })
  },
  getHomeDir(connectionId: number) {
    return instance.get(`/files/${connectionId}/home`)
  },
  readFile(connectionId: number, path: string) {
    return instance.get(`/files/${connectionId}/read`, { params: { path } })
  },
  // 用户管理
  listUsers() {
    return instance.get('/users')
  },
  addUser(data: any) {
    return instance.post('/users', data)
  },
  updateUser(id: number, data: any) {
    return instance.put(`/users/${id}`, data)
  },
  deleteUser(id: number) {
    return instance.delete(`/users/${id}`)
  },
  resetPassword(id: number, data: any) {
    return instance.put(`/users/${id}/resetpwd`, data)
  },
  changePassword(data: any) {
    return instance.put('/users/changepwd', data)
  }
} 