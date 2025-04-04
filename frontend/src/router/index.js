import Vue from 'vue'
import VueRouter from 'vue-router'
import store from '../store'

// 导入视图组件
import Home from '../views/Home.vue'
import Landing from '../views/Landing.vue'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import Contact from '../views/Contact.vue'

// 懒加载其他视图组件
const VoiceClone = () => import('../views/voice/VoiceClone.vue')
const VoiceCloneDetail = () => import('../views/voice/VoiceCloneDetail.vue')
const VoiceLibrary = () => import('../views/voice/VoiceLibrary.vue')
const TTS = () => import('../views/tts/TTS.vue')
const TTSDetail = () => import('../views/tts/TTSDetail.vue')
const DigitalHuman = () => import('../views/digital-human/DigitalHuman.vue')
const DigitalHumanDetail = () => import('../views/digital-human/DigitalHumanDetail.vue')
const UserProfile = () => import('../views/user/UserProfile.vue')
const AdminUsers = () => import('../views/admin/Users.vue')

Vue.use(VueRouter)

const routes = [
    {
        path: '/',
        name: 'Landing',
        component: Landing,
        beforeEnter: (to, from, next) => {
            const token = localStorage.getItem('token')
            if (token) {
                next('/home')
            } else {
                next()
            }
        }
    },
    {
        path: '/home',
        name: 'Home',
        component: Home,
        meta: { requiresAuth: true }
    },
    {
        path: '/login',
        component: Login,
        meta: { guest: true }
    },
    {
        path: '/register',
        component: Register,
        meta: { guest: true }
    },
    {
        path: '/voice-clone',
        component: VoiceClone,
        meta: { requiresAuth: true }
    },
    {
        path: '/voice-clone/:id',
        component: VoiceCloneDetail,
        meta: { requiresAuth: true }
    },
    {
        path: '/voice-library',
        component: VoiceLibrary,
        meta: { requiresAuth: true }
    },
    {
        path: '/tts',
        component: TTS,
        meta: { requiresAuth: true }
    },
    {
        path: '/tts/:id',
        component: TTSDetail,
        meta: { requiresAuth: true }
    },
    {
        path: '/digital-human',
        component: DigitalHuman,
        meta: { requiresAuth: true }
    },
    {
        path: '/digital-human/:id',
        component: DigitalHumanDetail,
        meta: { requiresAuth: true }
    },
    {
        path: '/profile',
        component: UserProfile,
        meta: { requiresAuth: true }
    },
    {
        path: '/admin/users',
        component: AdminUsers,
        meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
        path: '/contact',
        component: Contact
    },
    {
        path: '*',
        redirect: '/'
    }
]

const router = new VueRouter({
    mode: 'history',
    base: process.env.BASE_URL,
    routes
})

// 导航守卫
router.beforeEach((to, from, next) => {
    const isAuthenticated = store.getters.isAuthenticated
    const isAdmin = store.getters.isAdmin

    // 需要认证的路由
    if (to.matched.some(record => record.meta.requiresAuth)) {
        if (!isAuthenticated) {
            next({
                path: '/login',
                query: { redirect: to.fullPath }
            })
        } else if (to.matched.some(record => record.meta.requiresAdmin) && !isAdmin) {
            // 需要管理员权限的路由
            next({ path: '/' })
        } else {
            next()
        }
    } else if (to.matched.some(record => record.meta.guest) && isAuthenticated) {
        // 已登录用户不能访问游客页面
        next({ path: '/' })
    } else {
        next()
    }
})

export default router