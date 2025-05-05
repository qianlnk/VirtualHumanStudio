<template>
  <div id="app">
    <router-view v-if="$route.path === '/' || $route.path === '/login' || $route.path === '/register' || (!isAuthenticated && $route.path === '/contact')"></router-view>
    <el-container v-else-if="isAuthenticated">
        <!-- 左侧导航栏 -->
        <el-aside :width="isCollapse ? '64px' : '200px'" class="app-aside">
          <div class="aside-logo">
            <router-link to="/">
              <h1 v-if="!isCollapse">Virtual Human Studio</h1>
              <h1 v-else>VHS</h1>
            </router-link>
          </div>
          <div class="collapse-btn" @click="toggleCollapse">
            <i :class="isCollapse ? 'el-icon-s-unfold' : 'el-icon-s-fold'"></i>
          </div>
          <el-menu 
            :default-active="activeIndex" 
            mode="vertical" 
            router 
            :collapse="isCollapse"
            class="aside-menu"
            background-color="transparent"
            text-color="#b3e5fc"
            active-text-color="#64b5f6">
            <el-menu-item index="/">
              <i class="el-icon-s-home"></i>
              <span>首页</span>
            </el-menu-item>
            <el-submenu index="voice-services">
              <template slot="title">
                <i class="el-icon-microphone"></i>
                <span class="submenu-title">音频服务</span>
              </template>
              <el-menu-item index="/voice-clone">音色克隆</el-menu-item>
              <el-menu-item index="/voice-library">音色库</el-menu-item>
              <el-menu-item index="/tts">文本转语音</el-menu-item>
              <el-menu-item index="/speech2text">语音识别</el-menu-item>
            </el-submenu>
            <el-submenu index="digital-human">
              <template slot="title">
                <i class="el-icon-user"></i>

                <span class="submenu-title">数字人合成</span>
              </template>
              <el-menu-item index="/digital-human">数字人制作</el-menu-item>
            </el-submenu>
            <el-submenu index="image-processing">
              <template slot="title">
                <i class="el-icon-picture"></i>
                <span class="submenu-title">图像处理</span>
              </template>
              <el-menu-item v-for="module in imageProcessingModules" :key="module.id" :index="module.route">
                <i :class="module.icon" v-if="module.icon"></i>
                <span>{{ module.name }}</span>
                <el-tooltip v-if="module.description" :content="module.description" placement="right">
                  <i class="el-icon-info" style="margin-left: 5px;"></i>
                </el-tooltip>
              </el-menu-item>
            </el-submenu>
            <el-menu-item v-if="isAdmin" index="/admin/users">
              <i class="el-icon-s-custom"></i>
              <span>用户管理</span>
            </el-menu-item>
            <el-menu-item index="/contact">
              <i class="el-icon-phone"></i>
              <span>联系我们</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        
        <el-container>
          <!-- 顶部用户信息 -->
          <el-header height="50px" class="app-header">
            <div class="mobile-menu-btn" @click="toggleCollapse">
              <i class="el-icon-s-operation"></i>
            </div>
            <div class="header-user">
              <el-dropdown trigger="click" @command="handleCommand">
                <span class="el-dropdown-link">
                  {{ currentUser.username }}<i class="el-icon-arrow-down el-icon--right"></i>
                </span>
                <el-dropdown-menu slot="dropdown">
                  <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                  <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </el-dropdown>
            </div>
          </el-header>
          
          <!-- 主内容区域 -->
          <el-main class="app-main">
            <router-view />
          </el-main>
          
          <!-- 底部版权信息 -->
          <el-footer height="50px" class="app-footer">
            <p>© 2023 Virtual Human Studio. All Rights Reserved.</p>
          </el-footer>
        </el-container>
    </el-container>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { getImageProcessingModules } from './api/modules'

export default {
  name: 'App',
  data() {
    return {
      activeIndex: this.$route.path,
      imageProcessingModules: [],
      isCollapse: window.innerWidth <= 768
    }
  },
  computed: {
    ...mapGetters(['isAuthenticated', 'isAdmin', 'currentUser'])
  },
  watch: {
    // 监听路由变化，更新激活的菜单项
    '$route.path'(newPath) {
      this.activeIndex = newPath
    }
  },
  created() {
    // 获取图像处理模块列表
    this.fetchImageProcessingModules()
    // 监听窗口大小变化
    window.addEventListener('resize', this.handleResize)
  },
  beforeDestroy() {
    // 移除事件监听
    window.removeEventListener('resize', this.handleResize)
  },
  methods: {
    // 获取图像处理模块列表
    async fetchImageProcessingModules() {
      try {
        const response = await getImageProcessingModules()
        if (response.success) {
          // 合并本地和远程模块
          const localModules = [{
            id: 'accessory',
            name: '饰品替换',
            route: '/accessory',
            icon: 'el-icon-magic-stick',
            description: '智能替换人物饰品'
          }]
          this.imageProcessingModules = [...localModules, ...response.modules]
          // 动态添加模块路由
          const moduleRoutes = this.$router.options.generateImageProcessingRoutes(this.imageProcessingModules)
          // 逐个添加模块路由
          moduleRoutes.forEach(route => {
            this.$router.addRoute(route)
          })
        }
      } catch (error) {
        console.error('获取图像处理模块列表失败:', error)
        this.$message.error('获取图像处理模块列表失败')
      }
    },
    // 处理用户下拉菜单命令
    handleCommand(command) {
      if (command === 'logout') {
        this.logout()
      } else if (command === 'profile') {
        this.$router.push('/profile').catch(err => {
          if (err.name !== 'NavigationDuplicated') {
            throw err
          }
        })
      }
    },
    // 退出登录
    logout() {
      this.$store.dispatch('logout')
        .then(() => {
          this.$message.success('已成功退出登录')
          this.$router.push('/')
        })
        .catch(() => {
          this.$message.error('退出登录失败，请重试')
        })
    },
    // 切换侧边栏折叠状态
    toggleCollapse() {
      this.isCollapse = !this.isCollapse
    },
    // 处理窗口大小变化
    handleResize() {
      this.isCollapse = window.innerWidth <= 768
    }
  }
}
</script>

<style>
/* 全局样式 */
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
  -webkit-text-size-adjust: 100%;
  -webkit-tap-highlight-color: transparent;
}

#app {
  height: 100vh;
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #fff;
  overflow-x: hidden;
}

/* 左侧菜单样式 */
.app-aside {
  background-color: rgba(48, 65, 86, 0.7);
  backdrop-filter: blur(10px);
  height: 100vh;
  position: fixed;
  left: 0;
  top: 0;
  z-index: 1001;
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  transition: width 0.3s ease;
}

.aside-logo {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #1f2d3d;
  overflow: hidden;
}

.collapse-btn {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #b3e5fc;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  z-index: 1002;
}

.mobile-menu-btn {
  display: none;
  cursor: pointer;
  font-size: 20px;
  color: #fff;
}

.aside-logo h1 {
  margin: 0;
  font-size: 16px;
  color: #fff;
}

.aside-logo a {
  text-decoration: none;
}

.aside-menu {
  border-right: none;
  height: calc(100vh - 50px);
}

.aside-menu i {
  margin-right: 10px;
}

/* 头部样式 */
.app-header {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background-color: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  padding: 0 20px;
  transition: margin-left 0.3s ease;
}

.header-user {
  cursor: pointer;
}

.el-dropdown-link {
  color: #fff;
  display: flex;
  align-items: center;
}

/* 主内容区域样式 */
.app-main {
  margin-left: 200px;
  margin-top: 50px;
  padding: 20px;
  min-height: calc(100vh - 100px);
  background-color: transparent;
  transition: margin-left 0.3s ease;
}

/* 底部样式 */
.app-footer {
  margin-left: 200px;
  text-align: center;
  color: #b3e5fc;
  font-size: 12px;
  padding: 15px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  background-color: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(5px);
  transition: margin-left 0.3s ease;
}

/* 样式部分 */
.aside-menu .submenu-title {
  font-size: 15px;
}

.aside-menu .el-menu-item {
  font-size: 14px;
}

/* 响应式样式 */
@media screen and (max-width: 768px) {
  .mobile-menu-btn {
    display: block;
    margin-right: auto;
  }
  
  .collapse-btn {
    display: none;
  }
  
  .app-aside {
    transform: translateX(0);
    transition: transform 0.3s ease, width 0.3s ease;
  }
  
  .app-aside.is-collapsed {
    transform: translateX(-100%);
  }
  
  .app-main {
    margin-left: 0;
    padding: 15px;
  }
  
  .app-footer {
    margin-left: 0;
    padding: 10px 0;
  }
  
  .app-header {
    padding: 0 15px;
  }
  
  /* 折叠时的样式 */
  [class*="el-col-"] {
    width: 100%;
  }
  
  .el-form-item {
    margin-bottom: 15px;
  }
  
  .el-form-item__label {
    padding: 0 0 8px;
    display: block;
    text-align: left;
    width: 100% !important;
  }
  
  .el-form-item__content {
    margin-left: 0 !important;
    width: 100%;
  }
  
  .el-input {
    width: 100%;
  }
  
  .el-button {
    display: block;
    width: 100%;
    margin-left: 0 !important;
    margin-top: 8px;
  }
  
  .el-button + .el-button {
    margin-left: 0 !important;
  }
}

/* 侧边栏折叠时的样式 */
.el-menu--collapse .el-submenu__title span,
.el-menu--collapse .el-menu-item span {
  display: none;
}

.el-menu--collapse .el-tooltip {
  display: none;
}

/* 当侧边栏折叠时，调整主内容区域和底部的边距 */
.app-main {
  margin-left: v-bind('isCollapse ? "64px" : "200px"');
}

.app-footer {
  margin-left: v-bind('isCollapse ? "64px" : "200px"');
}
</style>