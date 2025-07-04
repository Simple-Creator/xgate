import { defineStore } from 'pinia'

export const useMainStore = defineStore('main', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    username: localStorage.getItem('username') || '',
    role: localStorage.getItem('role') || '',
    user: null as null | { username: string }
  }),
  getters: {
    isAdmin: (state) => state.role === 'admin',
  },
  actions: {
    setToken(token: string, role: string, username: string) {
      this.token = token
      this.role = role
      this.username = username
      localStorage.setItem('token', token)
      localStorage.setItem('role', role)
      localStorage.setItem('username', username)
    },
    setUser(user: any) {
      this.user = user
    },
    logout() {
      this.token = ''
      this.role = ''
      this.username = ''
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('role')
      localStorage.removeItem('username')
    }
  }
}) 