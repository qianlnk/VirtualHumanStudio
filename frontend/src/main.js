import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import axios from 'axios'
import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'
// 导入全局主题样式
import './assets/css/theme.css'

// 使用ElementUI
Vue.use(ElementUI)

// 配置axios
axios.defaults.baseURL = process.env.VUE_APP_API_URL || 'http://192.168.218.233:8080'

// 如果有token，设置请求头
const token = localStorage.getItem('token')
if (token) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
}

// 添加响应拦截器处理401错误
axios.interceptors.response.use(
    response => response,
    error => {
        if (error.response && error.response.status === 401) {
            // 清除认证信息并跳转到登录页
            store.commit('clearAuth')
            router.push('/login')
        }
        return Promise.reject(error)
    }
)

// 将axios挂载到Vue实例
Vue.prototype.$http = axios

Vue.config.productionTip = false

new Vue({
    router,
    store,
    render: h => h(App),
    created() {
        // 应用启动时，如果已登录，刷新用户信息
        if (store.getters.isAuthenticated) {
            store.dispatch('refreshUserInfo').catch(() => { })
        }
    }
}).$mount('#app')