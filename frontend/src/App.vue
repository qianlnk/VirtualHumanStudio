<template>
  <div id="app">
    <!-- 顶部导航栏 -->
    <el-header v-if="isAuthenticated" height="60px" class="app-header">
      <div class="header-logo">
        <router-link to="/">
          <h1>Virtual Human Studio</h1>
        </router-link>
      </div>
      
      <el-menu 
        :default-active="activeIndex" 
        mode="horizontal" 
        router 
        class="header-menu"
        background-color="#ffffff"
        text-color="#303133"
        active-text-color="#409EFF">
        <el-menu-item index="/">首页</el-menu-item>
        <el-submenu index="voice">
          <template slot="title">音频服务</template>
          <el-menu-item index="/voice-clone">音色克隆</el-menu-item>
          <el-menu-item index="/voice-library">音色库</el-menu-item>
          <el-menu-item index="/tts">文本转语音</el-menu-item>
        </el-submenu>
        <el-menu-item index="/digital-human">数字人合成</el-menu-item>
        <el-menu-item v-if="isAdmin" index="/admin/users">用户管理</el-menu-item>
      </el-menu>
      
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
    <el-footer v-if="isAuthenticated" height="50px" class="app-footer">
      <p>© 2023 Virtual Human Studio. All Rights Reserved.</p>
    </el-footer>
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
          this.$router.push('/login')
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
  display: flex;
  flex-direction: column;
}

/* 头部样式 */
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #ebeef5;
  background-color: #ffffff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  padding: 0 20px;
}

.header-logo h1 {
  margin: 0;
  font-size: 20px;
  color: #409EFF;
}

.header-logo a {
  text-decoration: none;
}

.header-menu {
  border-bottom: none;
}

.header-user {
  cursor: pointer;
}

.el-dropdown-link {
  color: #606266;
  display: flex;
  align-items: center;
}

/* 主内容区域样式 */
.app-main {
  flex: 1;
  overflow-y: auto;
  background-color: #f5f7fa;
  padding: 20px;
}

/* 底部样式 */
.app-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #ffffff;
  border-top: 1px solid #ebeef5;
  color: #909399;
  font-size: 12px;
}
</style>