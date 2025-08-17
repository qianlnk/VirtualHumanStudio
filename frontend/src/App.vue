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
            <span class="collapse-tooltip" v-if="!isCollapse">收起菜单</span>
            <span class="collapse-tooltip" v-else>展开菜单</span>
          </div>
          <el-menu 
            :default-active="activeIndex" 
            mode="vertical" 
            router 
            :collapse="isCollapse"
            class="aside-menu"
            background-color="transparent"
            text-color="#333333"
            active-text-color="#1a73e8">
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
              <el-menu-item index="/digital-human-template">数字人模版</el-menu-item>
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
            <el-menu-item index="/ai-chat">
              <i class="el-icon-chat-dot-round"></i>
              <span>AI绘画聊天</span>
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
        
        <el-container class="main-container">
          <!-- 顶部用户信息 - 不包含菜单按钮 -->
          <el-header v-if="!isMobile" height="60px" class="app-header">
            <!-- 在这里刻意删除菜单按钮 -->
            <div style="flex: 1;"></div>
            <div class="header-user">
              <el-dropdown trigger="click" @command="handleCommand">
                <span class="el-dropdown-link">
                  {{ currentUser.username }} <i class="el-icon-arrow-down el-icon--right"></i>
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
            <div class="app-main-content">
              <router-view />
            </div>
          </el-main>
          
          
          <!-- 移动端底部导航菜单 -->
          <div v-if="isMobile" class="mobile-bottom-nav" :style="isChatRoute && (keyboardOpen || keyboardOffset > 0) ? 'visibility:hidden;opacity:0' : ''">
            <div class="mobile-nav-item" @click="navigateTo('/inspiration')" :class="{'active': activeIndex === '/inspiration'}">
              <i class="el-icon-magic-stick"></i>
              <span>灵感空间</span>
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
          <div v-if="isMobile && showMobileSubmenu" class="mobile-submenu" :class="mobileSubmenuClass" :style="isChatRoute && (keyboardOpen || keyboardOffset > 0) ? 'visibility:hidden;opacity:0' : ''">
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
              <div class="mobile-submenu-item" @click="navigateTo('/ai-chat')">
                <i class="el-icon-chat-dot-round"></i>
                <span>AI绘画聊天</span>
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
  name: 'VHS APP',
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
      hasLoadedModules: false,
      // 键盘底部遮挡高度，用于锁定底部导航栏
      keyboardOffset: 0,
      // 来自AIChat的键盘显隐信号
      keyboardOpen: false
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
             this.activeIndex === '/ai-chat' ||
             this.activeIndex === '/membership' || 
             this.activeIndex === '/contact' ||
             this.activeIndex.includes('/admin');
    },
    mobileSubmenuClass() {
      return `mobile-submenu-${this.activeMobileSubmenu}`;
    },
    // 是否为AI聊天页
    isChatRoute() {
      try {
        return this.$route.path === '/ai-chat';
      } catch (e) {
        return false;
      }
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
      // 监听来自 AIChat 的键盘事件
      this.$eventBus.$on('keyboard-open', this.onKeyboardOpen);
      this.$eventBus.$on('keyboard-close', this.onKeyboardClose);
    }
    
    // 监听可视区域变化（用于检测软键盘）
    if (window.visualViewport) {
      window.visualViewport.addEventListener('resize', this.updateKeyboardOffset);
      window.visualViewport.addEventListener('scroll', this.updateKeyboardOffset);
      this.updateKeyboardOffset();
    }
  },
  beforeDestroy() {
    // 移除事件监听
    window.removeEventListener('resize', this.handleResize);
    // 移除自定义事件监听
    if (this.$eventBus) {
      this.$eventBus.$off('auth-changed');
      this.$eventBus.$off('keyboard-open', this.onKeyboardOpen);
      this.$eventBus.$off('keyboard-close', this.onKeyboardClose);
    }

    // 清理 visualViewport 监听
    try {
      if (window.visualViewport) {
        window.visualViewport.removeEventListener('resize', this.updateKeyboardOffset);
        window.visualViewport.removeEventListener('scroll', this.updateKeyboardOffset);
      }
    } catch (e) {
      // eslint-disable-next-line no-console
      console.warn('App beforeDestroy visualViewport cleanup error:', e);
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
        
        // 添加折叠类名以便于样式调整
        this.$nextTick(() => {
          const asideEl = document.querySelector('.app-aside')
          if (asideEl) {
            if (this.isCollapse) {
              asideEl.classList.add('is-collapse')
            } else {
              asideEl.classList.remove('is-collapse')
            }
          }
        })
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
    // 计算软键盘遮挡高度（用于抵消底部导航随键盘上移）
    updateKeyboardOffset() {
      try {
        if (!this.isMobile || !window.visualViewport) {
          this.keyboardOffset = 0;
          return;
        }
        const vv = window.visualViewport;
        const bottomOverlap = Math.max(0, window.innerHeight - (vv.height + vv.offsetTop));
        this.keyboardOffset = Math.round(bottomOverlap);
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('updateKeyboardOffset(App) error:', e);
        this.keyboardOffset = 0;
      }
    },
    onKeyboardOpen() {
      this.keyboardOpen = true;
    },
    onKeyboardClose() {
      this.keyboardOpen = false;
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
  background: linear-gradient(135deg, #f7f7f7, #e9ecef, #2c3e50);
  color: #333;
  overflow-x: hidden;
  position: relative;
}

/* 全局背景网格装饰 */
#app::before {
  content: '';
  position: fixed;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background-image:
    linear-gradient(rgba(0,0,0,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0,0,0,0.03) 1px, transparent 1px);
  background-size: 30px 30px;
  transform: rotate(5deg);
  opacity: 0.2;
  pointer-events: none;
  z-index: -1;
}

/* 左侧菜单样式 */
.app-aside {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(248, 248, 248, 0.85));
  backdrop-filter: blur(15px);
  height: 100vh;
  position: fixed;
  left: 0;
  top: 0;
  z-index: 1003; /* 确保在导航栏之上 */
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1), transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  border-radius: 0 30px 30px 0;
  overflow: hidden;
  border: none;
}

/* 添加装饰性边框光效 */
.app-aside::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 2px;
  background: linear-gradient(to bottom,
    rgba(255, 255, 255, 0),
    rgba(255, 255, 255, 0.8),
    rgba(255, 255, 255, 0));
  opacity: 0.6;
  z-index: 2;
}

/* 添加底部装饰图形 */
.app-aside::after {
  content: '';
  position: absolute;
  bottom: 80px;
  right: 0;
  width: 80px;
  height: 80px;
  background: radial-gradient(circle at bottom right,
    rgba(230, 230, 230, 0.8),
    rgba(255, 255, 255, 0));
  border-top-left-radius: 100%;
  opacity: 0.6;
  z-index: 1;
  pointer-events: none;
}

.aside-logo {
  height: 70px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(240, 240, 240, 0.7));
  padding: 0 5px; /* 减少内边距增加可用空间 */
  position: relative;
  margin-bottom: 10px;
  border-bottom: none;
  width: 100%; /* 确保使用全部可用宽度 */
}

/* 添加LOGO区域的底部装饰线 */
.aside-logo::after {
  content: '';
  position: absolute;
  left: 15%;
  right: 15%;
  bottom: 0;
  height: 1px;
  background: linear-gradient(to right,
    rgba(200, 200, 200, 0),
    rgba(200, 200, 200, 0.5),
    rgba(200, 200, 200, 0));
}

.aside-logo h1 {
  background: linear-gradient(120deg, #444444, #000000);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  font-weight: 600;
  letter-spacing: 0.5px;
  font-size: 16px; /* 减小字体大小 */
  white-space: nowrap; /* 防止文字换行 */
  overflow: hidden; /* 超出部分隐藏 */
  text-overflow: ellipsis; /* 超出显示省略号 */
  max-width: 100%; /* 限制最大宽度 */
  margin: 0; /* 移除默认边距 */
}

.collapse-btn {
  position: absolute;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  width: 45px;
  height: 45px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #333;
  background-color: #fff;
  border-radius: 50%;
  z-index: 1002;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.15);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.collapse-btn::after {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg,
    rgba(26, 115, 232, 0.1),
    rgba(26, 115, 232, 0.05),
    rgba(255, 255, 255, 0),
    rgba(26, 115, 232, 0.05),
    rgba(26, 115, 232, 0.1));
  border-radius: 50%;
  z-index: -1;
  animation: rotate 8s linear infinite;
}

@keyframes rotate {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.collapse-btn:hover {
  background-color: #f5f5f5;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
  transform: translateX(-50%) scale(1.05);
}

.collapse-btn i {
  font-size: 18px;
  transition: transform 0.3s ease;
}

.collapse-btn:hover i {
  transform: scale(1.1);
}

.collapse-tooltip {
  position: absolute;
  top: -30px;
  left: 50%;
  transform: translateX(-50%) scale(0);
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  opacity: 0;
  transition: all 0.2s ease;
  pointer-events: none;
}

.collapse-btn:hover .collapse-tooltip {
  transform: translateX(-50%) scale(1);
  opacity: 1;
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

.aside-logo a {
  text-decoration: none;
  display: flex;
  align-items: center;
  width: 100%;
  height: 100%;
  justify-content: center;
}

.aside-menu {
  border-right: none;
  height: calc(100vh - 70px);
  padding: 10px 15px 10px 0;
  overflow-y: auto;
  overflow-x: hidden;
}

/* 设置滚动条样式 */
.aside-menu::-webkit-scrollbar {
  width: 4px;
}

.aside-menu::-webkit-scrollbar-track {
  background: transparent;
}

.aside-menu::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.aside-menu::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.2);
}

.aside-menu i {
  margin-right: 12px;
  transition: all 0.3s ease;
  font-size: 18px;
}

.el-menu-item {
  height: 50px;
  line-height: 50px;
  margin: 6px 0;
  border-radius: 0 25px 25px 0;
  padding-left: 25px !important;
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
}

.el-menu-item:hover {
  background-color: rgba(0, 0, 0, 0.05) !important;
}

.el-menu-item.is-active {
  position: relative;
  background: linear-gradient(to right, rgba(26, 115, 232, 0.1), rgba(26, 115, 232, 0));
}

.el-menu-item.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  height: 35px;
  width: 4px;
  background-color: #1a73e8;
  border-radius: 0 4px 4px 0;
}

.el-submenu__title {
  height: 50px;
  line-height: 50px;
  margin: 6px 0;
  border-radius: 0 25px 25px 0;
  padding-left: 25px !important;
  transition: all 0.3s ease;
}

.el-submenu__title:hover {
  background-color: rgba(0, 0, 0, 0.05) !important;
}

.el-submenu.is-active > .el-submenu__title {
  color: #1a73e8 !important;
  position: relative;
}

.el-submenu.is-active > .el-submenu__title::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  height: 35px;
  width: 4px;
  background-color: #1a73e8;
  border-radius: 0 4px 4px 0;
}

.el-menu--vertical .el-menu {
  background: rgba(245, 245, 245, 0.5) !important;
  margin-left: 8px;
  border-left: 1px dashed rgba(0, 0, 0, 0.1);
}

.el-submenu__title i.el-submenu__icon-arrow {
  transition: transform 0.3s ease;
}

.el-submenu.is-opened > .el-submenu__title i.el-submenu__icon-arrow {
  transform: rotateZ(-90deg);
}

.el-menu--collapse .el-submenu__icon-arrow {
  display: none;
}

/* 头部样式 */
.app-header {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  border: none;
  background: transparent;
  backdrop-filter: blur(6px);
  padding: 0 25px;
  position: fixed; /* 固定导航栏 */
  top: 0;
  left: 200px; /* 与左侧菜单宽度一致 */
  right: 0;
  z-index: 1050; /* 确保在内容和滚动条之上 */
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  height: 60px !important;
}

/* 折叠时的头部样式调整 */
.is-collapse ~ .el-container .app-header {
  left: 64px; /* 当侧边栏折叠时，调整左边距 */
}

.header-user {
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: 6px 15px;
  border-radius: 25px;
  transition: all 0.3s ease;
  background-color: rgba(255, 255, 255, 0.5);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  backdrop-filter: blur(5px);
}

.header-user:hover {
  background-color: rgba(255, 255, 255, 0.85);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  transform: translateY(-1px);
}

.el-dropdown-link {
  color: #333;
  display: flex;
  align-items: center;
  font-weight: 500;
  font-size: 14px;
}

.el-dropdown-link i {
  margin-left: 5px;
  transition: transform 0.3s ease;
  font-size: 12px;
}

.el-dropdown-link:hover i {
  transform: rotate(180deg);
}

/* 下拉菜单样式 */
.el-dropdown-menu {
  border-radius: 12px !important;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15) !important;
  background-color: rgba(255, 255, 255, 0.95) !important;
  backdrop-filter: blur(10px) !important;
  border: 1px solid rgba(200, 200, 200, 0.3) !important;
  padding: 5px !important;
  margin-top: 5px !important;
}

.el-dropdown-menu__item {
  padding: 10px 15px !important;
  border-radius: 8px !important;
  margin: 2px 5px !important;
  transition: all 0.2s ease !important;
  line-height: 1.5 !important;
  font-size: 14px !important;
}

.el-dropdown-menu__item:hover {
  background-color: rgba(0, 0, 0, 0.05) !important;
  color: #1a73e8 !important;
}

.el-dropdown-menu__item i {
  margin-right: 8px !important;
}

.el-dropdown-menu__item.is-disabled {
  color: #999 !important;
}

/* 主内容区域样式 */
.app-main {
  margin-left: 200px;
  padding: 25px;
  padding-top: 85px; /* 为固定的头部导航栏留出足够空间 */
  min-height: calc(100vh - 60px); /* 移除底部版权信息后减少保留的空间 */
  background-color: transparent;
  transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  z-index: 1;
}

/* 创建一个内容容器，完全阻止内容滚动到导航栏区域 */
.el-main.app-main {
  position: relative;
  overflow: visible;
  padding-top: 0 !important; /* 移除顶部内边距，由内部容器控制 */
}

/* 创建一个滚动容器 */
.el-main.app-main .el-scrollbar__wrap,
.el-main.app-main .el-scrollbar__view {
  overflow: visible !important;
}

/* 添加一个内部容器来实现滚动，但不允许滚动到导航栏区域 */
.app-main-content {
  position: absolute;
  top: 60px; /* 从导航栏下方开始 */
  left: 0;
  right: 0;
  bottom: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 25px;
  height: calc(100vh - 60px); /* 只保留导航栏高度的空间，移除底部空间 */
  box-sizing: border-box;
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
  display: flex;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  width: 100%;
  height: 60px;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(15px);
  z-index: 1010;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.08);
  justify-content: space-around;
  align-items: center;
  border-top: 1px solid rgba(200, 200, 200, 0.3);
  margin: 0;
  will-change: transform;
}

.mobile-nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;
  padding: 6px 0;
  flex: 1;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

.mobile-nav-item i {
  font-size: 22px;
  margin-bottom: 3px;
  transition: transform 0.3s ease, color 0.3s ease;
}

.mobile-nav-item span {
  font-size: 12px;
  transition: color 0.3s ease;
  white-space: nowrap;
}

.mobile-nav-item.active {
  color: #1a73e8;
  font-weight: 500;
}

.mobile-nav-item.active i {
  transform: translateY(-2px);
  color: #1a73e8;
}

.mobile-nav-item:active {
  transform: scale(0.95);
}

.mobile-nav-item::after {
  content: '';
  position: absolute;
  bottom: -1px;
  left: 50%;
  transform: translateX(-50%) scaleX(0);
  width: 20px;
  height: 3px;
  background-color: #1a73e8;
  border-radius: 3px 3px 0 0;
  transition: transform 0.3s ease;
  opacity: 0;
}

.mobile-nav-item.active::after {
  transform: translateX(-50%) scaleX(1);
  opacity: 1;
}

/* 移动端样式增强 */
@media screen and (max-width: 768px) {
  .mobile-nav-item {
    padding: 8px 0;
  }
  
  .mobile-nav-item i {
    font-size: 24px;
  }
}

/* 移动端子菜单 */
.mobile-submenu {
  position: fixed;
  bottom: 60px; /* 确保子菜单紧贴底部导航 */
  left: 0;
  width: 100%;
  background-color: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(20px);
  z-index: 1009;
  box-shadow: 0 -8px 24px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  transform: translateY(0);
  max-height: 70vh;
  overflow-y: auto;
  border-top-left-radius: 16px;
  border-top-right-radius: 16px;
}

.mobile-submenu-title {
  display: flex;
  align-items: center;
  padding: 18px 20px;
  border-bottom: 1px solid rgba(200, 200, 200, 0.3);
  position: sticky;
  top: 0;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  z-index: 2;
}

.mobile-submenu-title i {
  margin-right: 15px;
  font-size: 20px;
  cursor: pointer;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background-color: rgba(0, 0, 0, 0.05);
  transition: all 0.2s ease;
}

.mobile-submenu-title i:active {
  background-color: rgba(0, 0, 0, 0.1);
  transform: scale(0.95);
}

.mobile-submenu-title span {
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0.5px;
  background: linear-gradient(120deg, #444444, #000000);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.mobile-submenu-item {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  margin: 2px 10px;
  border-radius: 12px;
  color: #333;
  transition: all 0.25s ease;
  position: relative;
}

.mobile-submenu-item:active {
  background-color: rgba(0, 0, 0, 0.05);
  transform: scale(0.98);
}

.mobile-submenu-item i {
  margin-right: 15px;
  font-size: 20px;
  width: 35px;
  height: 35px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.05);
  border-radius: 10px;
  transition: all 0.3s ease;
}

.mobile-submenu-item span {
  font-weight: 500;
  font-size: 15px;
  transition: color 0.3s ease;
}

.mobile-submenu-content {
  padding-bottom: 20px; /* 确保底部有足够的空间 */
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
    position: fixed !important; /* 确保在移动端也固定 */
    left: 0 !important;
    right: 0 !important;
    top: 0 !important;
    z-index: 1001 !important;
    height: 60px !important;
    background-color: rgba(255, 255, 255, 0.3) !important; /* 移动端使用更低的不透明度 */
    border: none !important;
  }
  
  .collapse-btn {
    display: none;
  }
  
  .app-main {
    margin-left: 0 !important;
    min-height: calc(100vh - 60px); /* 移除底部版权信息的空间，只保留导航栏空间 */
  }
  
  .app-main-content {
    padding: 20px;
    top: 0; /* 移动端取消顶部导航栏占位 */
    padding-bottom: 80px;
    height: calc(100vh - 60px); /* 仅减去底部导航栏的高度 */
  }
  
  .mobile-main {
    margin-bottom: 0; /* 移除底部边距，让内容区域占据更多空间 */
    padding-bottom: 0;
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


/* 遮罩层 */
.menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.2);
  z-index: 1005;
  display: none;
}

.menu-overlay.active {
  display: block;
}
</style>