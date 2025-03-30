import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

Vue.use(Vuex)

export default new Vuex.Store({
    state: {
        token: localStorage.getItem('token') || '',
        user: JSON.parse(localStorage.getItem('user') || '{}')
    },
    getters: {
        isAuthenticated: state => !!state.token,
        isAdmin: state => state.user && state.user.role === 'admin',
        currentUser: state => state.user,
        user: state => state.user,
        userRole: state => state.user ? (state.user.role === 'admin' ? '管理员' : '普通用户') : ''
    },
    mutations: {
        setToken(state, token) {
            state.token = token
            localStorage.setItem('token', token)
            // 设置axios默认请求头
            axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
        },
        setUser(state, user) {
            state.user = user
            localStorage.setItem('user', JSON.stringify(user))
        },
        clearAuth(state) {
            state.token = ''
            state.user = {}
            localStorage.removeItem('token')
            localStorage.removeItem('user')
            // 清除axios默认请求头
            delete axios.defaults.headers.common['Authorization']
        }
    },
    actions: {
        login({ commit }, credentials) {
            return new Promise((resolve, reject) => {
                axios.post('/api/login', credentials)
                    .then(response => {
                        const { token, user } = response.data
                        commit('setToken', token)
                        commit('setUser', user)
                        resolve(response)
                    })
                    .catch(error => {
                        reject(error)
                    })
            })
        },
        register({ commit }, userData) {
            return new Promise((resolve, reject) => {
                axios.post('/api/register', userData)
                    .then(response => {
                        const { token, user } = response.data
                        commit('setToken', token)
                        commit('setUser', user)
                        resolve(response)
                    })
                    .catch(error => {
                        reject(error)
                    })
            })
        },
        logout({ commit }) {
            return new Promise((resolve, reject) => {
                axios.post('/api/logout')
                    .then(() => {
                        commit('clearAuth')
                        resolve()
                    })
                    .catch(error => {
                        // 即使请求失败，也清除本地认证信息
                        commit('clearAuth')
                        reject(error)
                    })
            })
        },
        // 刷新用户信息
        refreshUserInfo({ commit }) {
            return new Promise((resolve, reject) => {
                axios.get('/api/user')
                    .then(response => {
                        commit('setUser', response.data.user)
                        resolve(response)
                    })
                    .catch(error => {
                        // 如果获取用户信息失败，可能是token过期
                        if (error.response && error.response.status === 401) {
                            commit('clearAuth')
                        }
                        reject(error)
                    })
            })
        }
    }
})