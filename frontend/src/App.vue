<template>
  <div id="app">
    <router-view v-if="$route.path === '/' || $route.path === '/login' || $route.path === '/register' || (!isAuthenticated && $route.path === '/contact')"></router-view>
    <el-container v-else-if="isAuthenticated">
        <!-- 左侧导航栏 -->
        <el-aside width="200px" class="app-aside">
          <div class="aside-logo">
            <router-link to="/">
              <h1>Virtual Human Studio</h1>
            </router-link>
          </div>
          <el-menu 
            :default-active="activeIndex" 
            mode="vertical" 
            router 
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
            </el-submenu>
            <el-submenu index="digital-human">
              <template slot="title">
                <i class="el-icon-user"></i>
                <span class="submenu-title">数字人合成</span>
              </template>
              <el-menu-item index="/digital-human">数字人制作</el-menu-item>
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

export default {
  name: 'App',
  data() {
    return {
      activeIndex: this.$route.path
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
  methods: {
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
}

#app {
  height: 100vh;
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #fff;
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
}

.aside-logo {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #1f2d3d;
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
}

/* 样式部分 */
.aside-menu .submenu-title {
  font-size: 15px;
}

.aside-menu .el-menu-item {
  font-size: 14px;
}
</style>