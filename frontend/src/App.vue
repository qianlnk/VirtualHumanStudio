<template>
  <div id="app">
    <router-view v-if="$route.path === '/' || $route.path === '/login' || $route.path === '/register' || (!isAuthenticated && $route.path === '/contact')"></router-view>
    <el-container v-else-if="isAuthenticated">
        <!-- 左侧导航栏 - 在非移动端显示 -->
        <el-aside v-if="!isMobile" :width="isCollapse ? '64px' : '200px'" class="app-aside" :class="{'is-mobile': isMobile, 'is-hidden': isMobile}">
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
            <el-submenu v-if="isAdmin" index="admin">
              <template slot="title">
                <i class="el-icon-s-tools"></i>
                <span class="submenu-title">后台管理</span>
              </template>
              <el-menu-item index="/admin/statistics">
                <i class="el-icon-data-analysis"></i>
                <span>数据统计</span>
              </el-menu-item>
              <el-menu-item index="/admin/users">
                <i class="el-icon-s-custom"></i>
                <span>用户管理</span>
              </el-menu-item>
              <el-menu-item index="/admin/messages">
                <i class="el-icon-message"></i>
                <span>留言管理</span>
              </el-menu-item>
              <el-menu-item index="/admin/membership-orders">
                <i class="el-icon-s-order"></i>
                <span>会员订单</span>
              </el-menu-item>
              <el-menu-item index="/admin/review-tasks">
                <i class="el-icon-check"></i>
                <span>分享审核</span>
              </el-menu-item>
            </el-submenu>
            <el-menu-item index="/inspiration">
              <i class="el-icon-magic-stick"></i>
              <span>灵感空间</span>
            </el-menu-item>
            <el-menu-item index="/contact">
              <i class="el-icon-phone"></i>
              <span>联系我们</span>
            </el-menu-item>
            
            <!-- 添加会员中心菜单项 -->
            <el-menu-item index="/membership">
              <i class="el-icon-medal"></i>
              <span>会员中心</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        
        <el-container>
          <!-- 顶部用户信息 - 不包含菜单按钮 -->
          <el-header height="50px" class="app-header mobile-header">
            <!-- 在这里刻意删除菜单按钮 -->
            <div style="flex: 1;"></div>
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
          <el-main class="app-main" :class="{'mobile-main': isMobile}">
            <router-view />
          </el-main>
          
          <!-- 底部版权信息 -->
          <el-footer height="50px" class="app-footer" :class="{'mobile-footer': isMobile}">
            <p>© 2023 Virtual Human Studio. All Rights Reserved.</p>
          </el-footer>
          
          <!-- 移动端底部导航菜单 -->
          <div v-if="isMobile" class="mobile-bottom-nav">
            <div class="mobile-nav-item" @click="navigateTo('/')" :class="{'active': activeIndex === '/'}">
              <i class="el-icon-s-home"></i>
              <span>首页</span>
            </div>
            <div class="mobile-nav-item" @click="toggleMobileSubmenu('voice')" :class="{'active': isVoiceActive}">
              <i class="el-icon-microphone"></i>
              <span>音频</span>
            </div>
            <div class="mobile-nav-item" @click="navigateTo('/digital-human')" :class="{'active': activeIndex === '/digital-human'}">
              <i class="el-icon-user"></i>
              <span>数字人</span>
            </div>
            <div class="mobile-nav-item" @click="toggleMobileSubmenu('image')" :class="{'active': isImageActive}">
              <i class="el-icon-picture"></i>
              <span>图像</span>
            </div>
            <div class="mobile-nav-item" @click="toggleMobileSubmenu('more')" :class="{'active': isMoreActive}">
              <i class="el-icon-more"></i>
              <span>更多</span>
            </div>
          </div>
          
          <!-- 移动端子菜单 -->
          <div v-if="isMobile && showMobileSubmenu" class="mobile-submenu" :class="mobileSubmenuClass">
            <!-- 音频服务子菜单 -->
            <div v-if="activeMobileSubmenu === 'voice'" class="mobile-submenu-content">
              <div class="mobile-submenu-title">
                <i class="el-icon-back" @click="closeMobileSubmenu"></i>
                <span>音频服务</span>
              </div>
              <div class="mobile-submenu-item" v-for="(item, index) in voiceMenuItems" :key="index" @click="navigateTo(item.route)">
                <i :class="item.icon"></i>
                <span>{{ item.name }}</span>
              </div>
            </div>
            
            <!-- 图像处理子菜单 -->
            <div v-if="activeMobileSubmenu === 'image'" class="mobile-submenu-content">
              <div class="mobile-submenu-title">
                <i class="el-icon-back" @click="closeMobileSubmenu"></i>
                <span>图像处理</span>
              </div>
              <div class="mobile-submenu-item" v-for="module in imageProcessingModules" :key="module.id" @click="navigateTo(module.route)">
                <i :class="module.icon" v-if="module.icon"></i>
                <span>{{ module.name }}</span>
              </div>
            </div>
            
            <!-- 更多选项子菜单 -->
            <div v-if="activeMobileSubmenu === 'more'" class="mobile-submenu-content">
              <div class="mobile-submenu-title">
                <i class="el-icon-back" @click="closeMobileSubmenu"></i>
                <span>更多功能</span>
              </div>
              <div class="mobile-submenu-item" @click="navigateTo('/inspiration')">
                <i class="el-icon-magic-stick"></i>
                <span>灵感空间</span>
              </div>
              <div class="mobile-submenu-item" @click="navigateTo('/membership')">
                <i class="el-icon-medal"></i>
                <span>会员中心</span>
              </div>
              <div class="mobile-submenu-item" @click="navigateTo('/contact')">
                <i class="el-icon-phone"></i>
                <span>联系我们</span>
              </div>
              <div v-if="isAdmin" class="mobile-submenu-item" @click="navigateTo('/admin/review-tasks')">
                <i class="el-icon-check"></i>
                <span>分享审核</span>
              </div>
            </div>
          </div>
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
      isCollapse: window.innerWidth <= 768,
      isMobile: window.innerWidth <= 768,
      showMobileMenu: false,
      showMobileSubmenu: false,
      activeMobileSubmenu: '',
      voiceMenuItems: [
        { name: '音色克隆', route: '/voice-clone', icon: 'el-icon-microphone' },
        { name: '音色库', route: '/voice-library', icon: 'el-icon-headset' },
        { name: '文本转语音', route: '/tts', icon: 'el-icon-reading' },
        { name: '语音识别', route: '/speech2text', icon: 'el-icon-mic' }
      ],
      hasLoadedModules: false
    }
  },
  computed: {
    ...mapGetters(['isAuthenticated', 'isAdmin', 'currentUser']),
    isVoiceActive() {
      return this.activeIndex.includes('/voice-') || this.activeIndex === '/tts' || this.activeIndex === '/speech2text';
    },
    isImageActive() {
      const imageRoutes = this.imageProcessingModules.map(module => module.route);
      return imageRoutes.some(route => this.activeIndex.includes(route)) || this.activeIndex === '/accessory';
    },
    isMoreActive() {
      return this.activeIndex === '/inspiration' || 
             this.activeIndex === '/membership' || 
             this.activeIndex === '/contact' ||
             this.activeIndex.includes('/admin');
    },
    mobileSubmenuClass() {
      return `mobile-submenu-${this.activeMobileSubmenu}`;
    }
  },
  watch: {
    // 监听路由变化，更新激活的菜单项
    '$route.path'(newPath) {
      this.activeIndex = newPath
      if (this.isMobile) {
        this.closeMobileSubmenu();
      }
      
      // 在路由变化时，如果用户已登录且还没加载模块，则加载模块
      if (this.$store.getters.isAuthenticated && !this.hasLoadedModules) {
        console.log('检测到路由变化且用户已登录，正在加载模块...');
        this.fetchImageProcessingModules();
      }
    },
    
    // 监听登录状态变化
    '$store.getters.isAuthenticated': {
      immediate: true,
      handler(isAuthenticated) {
        if (isAuthenticated && !this.hasLoadedModules) {
          console.log('检测到登录状态变化，用户已登录，正在加载模块...');
          this.fetchImageProcessingModules();
        }
      }
    }
  },
  created() {
    // 获取图像处理模块列表
    this.fetchImageProcessingModules()
    // 监听窗口大小变化
    window.addEventListener('resize', this.handleResize)
    // 如果用户已登录，刷新用户信息
    if (this.$store.getters.isAuthenticated) {
      this.$store.dispatch('refreshUserInfo').catch(err => {
        console.error('刷新用户信息失败:', err)
      })
    }
  },
  mounted() {
    // 在DOM加载完成后，强制移除菜单按钮
    this.$nextTick(() => {
      // 确保菜单按钮不显示
      const menuButtons = document.querySelectorAll('.mobile-menu-btn, .el-icon-s-operation');
      menuButtons.forEach(btn => {
        if (btn) {
          btn.style.display = 'none';
          btn.style.visibility = 'hidden';
          btn.style.opacity = '0';
          btn.style.pointerEvents = 'none';
          btn.style.position = 'absolute';
          btn.style.left = '-9999px';
          if (btn.parentNode) {
            btn.parentNode.removeChild(btn);
          }
        }
      });
      
      // 确保侧边栏在移动端总是隐藏
      if (this.isMobile) {
        const asideElement = document.querySelector('.app-aside');
        if (asideElement) {
          asideElement.style.display = 'none';
          asideElement.style.visibility = 'hidden';
        }
      }
    });
    
    // 监听登录事件
    if (this.$eventBus) {
      this.$eventBus.$on('auth-changed', () => {
        console.log('收到auth-changed事件，加载模块');
        this.fetchImageProcessingModules();
      });
    }
  },
  beforeDestroy() {
    // 移除事件监听
    window.removeEventListener('resize', this.handleResize);
    // 移除自定义事件监听
    if (this.$eventBus) {
      this.$eventBus.$off('auth-changed');
    }
  },
  methods: {
    // 获取图像处理模块列表
    async fetchImageProcessingModules() {
      try {
        // 确保用户已登录
        if (this.$store.getters.isAuthenticated) {
          console.log('开始获取图像处理模块列表');
          const response = await getImageProcessingModules();
          if (response.success) {
            // // 合并本地和远程模块
            // const localModules = [{
            //   id: 'accessory',
            //   name: '饰品替换',
            //   route: '/accessory',
            //   icon: 'el-icon-magic-stick',
            //   description: '智能替换人物饰品'
            // }];
            // this.imageProcessingModules = [...localModules, ...response.modules];
            this.imageProcessingModules = [...response.modules];
            
            // 动态添加模块路由
            const moduleRoutes = this.$router.options.generateImageProcessingRoutes(this.imageProcessingModules);
            
            // 逐个添加模块路由
            moduleRoutes.forEach(route => {
              try {
                // 尝试添加路由
                this.$router.addRoute(route);
              } catch (e) {
                console.warn('添加路由失败，可能已存在:', route.path, e);
              }
            });
            
            console.log('图像处理模块加载成功:', this.imageProcessingModules);
            
            // 标记模块已加载
            this.hasLoadedModules = true;
            
            // 强制重新渲染菜单
            this.$nextTick(() => {
              this.$forceUpdate();
            });
          }
        } else {
          console.log('用户未登录，不加载图像处理模块');
        }
      } catch (error) {
        console.error('获取图像处理模块列表失败:', error);
        this.$message.error('获取图像处理模块列表失败');
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
    // 切换侧边栏折叠状态 - 仅在非移动端使用
    toggleCollapse() {
      if (!this.isMobile) {
        this.isCollapse = !this.isCollapse
      }
    },
    // 处理窗口大小变化
    handleResize() {
      const mobile = window.innerWidth <= 768;
      this.isMobile = mobile;
      this.isCollapse = mobile;
      
      // 移动端总是隐藏侧边栏
      if (mobile) {
        this.showMobileMenu = false;
        // 关闭移动端子菜单
        this.closeMobileSubmenu();
      }
    },
    // 移动端切换主菜单 - 在移动端实际上不再使用
    toggleMobileMenu() {
      // 移动端禁用侧边栏
      if (this.isMobile) return;
      
      this.showMobileMenu = !this.showMobileMenu;
      const asideElement = document.querySelector('.app-aside');
      if (asideElement) {
        asideElement.classList.toggle('is-hidden', !this.showMobileMenu);
      }
    },
    // 移动端切换子菜单
    toggleMobileSubmenu(type) {
      if (this.activeMobileSubmenu === type && this.showMobileSubmenu) {
        this.closeMobileSubmenu();
      } else {
        this.showMobileSubmenu = true;
        this.activeMobileSubmenu = type;
      }
    },
    // 关闭移动端子菜单
    closeMobileSubmenu() {
      this.showMobileSubmenu = false;
    },
    // 页面导航
    navigateTo(route) {
      this.$router.push(route).catch(err => {
        if (err.name !== 'NavigationDuplicated') {
          throw err;
        }
      });
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
  transition: width 0.3s ease, transform 0.3s ease;
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

.mobile-menu-btn, 
div.mobile-menu-btn, 
.el-icon-s-operation, 
div.el-icon-s-operation,
i.el-icon-s-operation,
[class*="mobile-menu-btn"],
.app-header > div:first-child:not(.header-user) {
  display: none !important;
  opacity: 0 !important;
  visibility: hidden !important;
  width: 0 !important;
  height: 0 !important;
  padding: 0 !important;
  margin: 0 !important;
  position: absolute !important;
  left: -9999px !important;
  top: -9999px !important;
  pointer-events: none !important;
}

/* 明确设置头部 */
.mobile-header {
  justify-content: flex-end !important;
  padding-left: 20px !important;
  padding-right: 20px !important;
}

.mobile-header::before {
  display: none !important;
  content: none !important;
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

/* 移动端底部导航菜单 */
.mobile-bottom-nav {
  display: none;
  position: fixed;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 56px;
  background-color: rgba(48, 65, 86, 0.9);
  backdrop-filter: blur(10px);
  z-index: 1010;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.2);
  display: flex;
  justify-content: space-around;
  align-items: center;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.mobile-nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #b3e5fc;
  padding: 4px 0;
  flex: 1;
  transition: all 0.3s;
}

.mobile-nav-item i {
  font-size: 20px;
  margin-bottom: 2px;
}

.mobile-nav-item span {
  font-size: 12px;
}

.mobile-nav-item.active {
  color: #64b5f6;
  font-weight: bold;
}

/* 移动端样式增强 */
@media screen and (max-width: 768px) {
  .mobile-nav-item {
    padding: 6px 0;
  }
  
  .mobile-nav-item i {
    font-size: 22px;
  }
  
  .mobile-nav-item.active {
    background-color: rgba(100, 181, 246, 0.2);
    border-top: 2px solid #64b5f6;
    margin-top: -2px;
  }
}

/* 移动端子菜单 */
.mobile-submenu {
  position: fixed;
  bottom: 56px;
  left: 0;
  width: 100%;
  background-color: rgba(48, 65, 86, 0.95);
  backdrop-filter: blur(15px);
  z-index: 1009;
  box-shadow: 0 -5px 15px rgba(0, 0, 0, 0.2);
  transition: transform 0.3s ease;
  transform: translateY(0);
  max-height: 70vh;
  overflow-y: auto;
}

.mobile-submenu-title {
  display: flex;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.mobile-submenu-title i {
  margin-right: 10px;
  font-size: 18px;
  cursor: pointer;
}

.mobile-submenu-title span {
  font-size: 16px;
  font-weight: bold;
}

.mobile-submenu-item {
  display: flex;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  color: #b3e5fc;
  transition: background-color 0.2s;
}

.mobile-submenu-item:active {
  background-color: rgba(255, 255, 255, 0.1);
}

.mobile-submenu-item i {
  margin-right: 10px;
  font-size: 18px;
}

/* 响应式样式 */
@media screen and (max-width: 768px) {
  /* 移动端完全隐藏侧边栏和菜单按钮 */
  html body #app .app-aside, 
  html body #app .mobile-menu-btn,
  html body #app .el-header .mobile-menu-btn,
  html body #app .el-header > div:first-child:not(.header-user) {
    display: none !important;
    opacity: 0 !important;
    visibility: hidden !important;
    width: 0 !important;
    height: 0 !important;
    position: absolute !important;
    pointer-events: none !important;
  }
  
  html body #app .app-header {
    justify-content: flex-end !important;
    padding: 0 15px !important;
  }
  
  .collapse-btn {
    display: none;
  }
  
  .app-main {
    margin-left: 0 !important;
    padding: 15px;
    padding-bottom: 76px;
    min-height: calc(100vh - 106px);
  }
  
  .mobile-main {
    margin-bottom: 56px;
  }
  
  .app-footer {
    margin-left: 0 !important;
    padding: 10px 0;
    margin-bottom: 56px;
  }
  
  .mobile-footer {
    margin-bottom: 56px;
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
  margin-left: v-bind('isCollapse && !isMobile ? "64px" : isMobile ? "0" : "200px"');
}

.app-footer {
  margin-left: v-bind('isCollapse && !isMobile ? "64px" : isMobile ? "0" : "200px"');
}

/* 遮罩层 */
.menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1005;
  display: none;
}

.menu-overlay.active {
  display: block;
}
</style>