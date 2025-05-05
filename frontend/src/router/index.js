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
const ASR = () => import('../views/asr/ASR.vue')
const ASRDetail = () => import('../views/asr/ASRDetail.vue')
const DigitalHuman = () => import('../views/digital-human/DigitalHuman.vue')
const DigitalHumanDetail = () => import('../views/digital-human/DigitalHumanDetail.vue')
const Accessory = () => import('../views/accessory/Accessory.vue')
const ImageProcessingTask = () => import('../views/workflow/ImageProcessingTask.vue')
const ImageProcessingTaskDetail = () => import('../views/workflow/ImageProcessingTaskDetail.vue')
const UserProfile = () => import('../views/user/UserProfile.vue')
const AdminUsers = () => import('../views/admin/Users.vue')
const AdminMessages = () => import('../views/admin/Messages.vue')
const AdminStatistics = () => import('../views/admin/Statistics.vue')

Vue.use(VueRouter)

// 基础路由配置
const baseRoutes = [
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
        path: '/speech2text',
        component: ASR,
        meta: { requiresAuth: true }
    },
    {
        path: '/speech2text/:id',
        component: ASRDetail,
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
        path: '/admin/messages',
        component: AdminMessages,
        meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
        path: '/admin/statistics',
        component: AdminStatistics,
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

// 动态生成图像处理模块路由
const generateImageProcessingRoutes = (modules) => {
    const routes = []
    if (modules && modules.length > 0) {
        modules.forEach(module => {
            // 添加模块列表页路由
            routes.push({
                path: module.route,
                component: ImageProcessingTask,
                props: { moduleId: module.id },
                meta: { requiresAuth: true }
            })
            // 添加模块详情页路由
            routes.push({
                path: `${module.route}/task/:id`,
                component: ImageProcessingTaskDetail,
                meta: { requiresAuth: true }
            })
        })
    }
    return routes
}

// 添加配饰路由
const AccessoryDetail = () => import('../views/accessory/AccessoryDetail.vue')
baseRoutes.push({
    path: '/accessory',
    component: Accessory,
    meta: { requiresAuth: true }
})
baseRoutes.push({
    path: '/accessory/:id',
    component: AccessoryDetail,
    meta: { requiresAuth: true }
})

// 合并基础路由和动态路由
const routes = baseRoutes.concat(generateImageProcessingRoutes([{
    id: 'image-processing',
    route: '/image-processing'
}]))

const router = new VueRouter({
    mode: 'history',
    base: process.env.BASE_URL,
    routes
})

// 添加generateImageProcessingRoutes方法到router实例
router.options.generateImageProcessingRoutes = generateImageProcessingRoutes

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