<template>
  <div class="ai-chat-container theme-background">
    <!-- 历史侧栏切换按钮（仅PC端显示，移动端使用自定义图标） -->
    <el-button
      v-if="!isMobile"
      class="sidebar-toggle-button chat-btn theme-button--text"
      :icon="sidebarVisible ? 'el-icon-arrow-left' : 'el-icon-notebook-2'"
      size="small"
      @click="toggleSidebar"
    >
      {{ sidebarVisible ? '隐藏历史' : '' }}
    </el-button>

    <!-- 移动端顶部栏：左侧一长一短两条横线按钮切换历史，顶部中间模型下拉 -->
    <div class="mobile-topbar header-glass" v-if="isMobile">
      <button class="history-toggle" @click="toggleSidebar" aria-label="切换历史侧栏">
        <span class="line long"></span>
        <span class="line short"></span>
      </button>
      <div class="topbar-center">
        <el-select
          v-model="selectedModel"
          placeholder="选择模型"
          size="mini"
          filterable
          default-first-option
          popper-class="model-select-popper"
          :popper-append-to-body="!isMobile">
          <el-option
            v-for="model in availableModels"
            :key="model.name"
            :label="model.name"
            :value="model.name">
            <div class="model-info">
              <div class="model-name">{{ model.name }}</div>
              <div class="model-desc">{{ model.description }}</div>
            </div>
          </el-option>
        </el-select>
      </div>
    </div>

    <div class="chat-main">
      <!-- 左侧会话列表 - 根据sidebarVisible控制显示/隐藏 -->
      <transition name="slide-fade">
        <div class="chat-sidebar glass-card" v-if="sidebarVisible">
          <div class="sidebar-header">
            <div class="section-title compact">
              <div class="section-accent"></div>
              <h3>会话历史</h3>
            </div>
            <div class="sidebar-actions">
              <el-button
                icon="el-icon-plus"
                size="mini"
                class="theme-button--text"
                type="text"
                @click="createNewSession"
                title="新建会话"
              />
            </div>
          </div>
          
          <div class="session-list">
            <!-- 移动端新建会话入口：清空当前聊天，由首条消息自动创建会话 -->
            <div
              v-if="isMobile"
              class="session-item new-session"
              @click="createNewSession"
            >
              <div class="session-info">
                <div class="session-title">
                  <i class="el-icon-plus" style="margin-right:6px;"></i> 新建会话
                </div>
                <div class="session-meta">
                  <span>清空当前聊天，从第一条消息生成标题</span>
                </div>
              </div>
            </div>
            <div
              v-for="session in sessions"
              :key="session.id"
              :class="['session-item', { active: currentSession && currentSession.id === session.id }]"
              @click="selectSession(session)"
            >
              <div class="session-info">
                <div class="session-title">{{ session.title }}</div>
                <div class="session-meta">
                  <span class="message-count">{{ session.message_count }} 条消息</span>
                  <span class="session-time">{{ formatTime(session.updated_at) }}</span>
                </div>
              </div>
              <div class="session-actions">
                <el-button
                  type="text"
                  icon="el-icon-delete"
                  size="mini"
                  class="theme-button--text"
                  @click.stop="deleteSession(session.id)"
                />
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- 移动端遮罩，用于全屏侧栏 -->
      <div v-if="sidebarVisible && isMobile" class="mobile-overlay" @click="toggleSidebar"></div>
      <!-- 右侧聊天区域 -->
      <div class="chat-content">
        <!-- 不再需要这里的侧边栏切换按钮，已移至顶部导航栏 -->
        <!-- 模型选择区域 - 即使在新会话状态下也显示 -->
        <div class="model-selector glass-card">
          <div class="section-title compact">
            <div class="section-accent"></div>
            <h3>对话设置</h3>
          </div>
          <div class="selector-header">
            <div class="model-list">
              <el-select
                v-model="selectedModel"
                placeholder="请选择AI模型"
                style="width: 300px;"
                filterable
                default-first-option
                popper-class="model-select-popper"
                :popper-append-to-body="!isMobile">
                <el-option
                  v-for="model in availableModels"
                  :key="model.name"
                  :label="model.name"
                  :value="model.name">
                  <div class="model-info">
                    <div class="model-name">{{ model.name }}</div>
                    <div class="model-desc">{{ model.description }}</div>
                  </div>
                </el-option>
              </el-select>
            </div>
            
            <div class="selector-controls">
              <el-switch
                v-model="enableStream"
                active-text="流式返回"
                inactive-text="普通返回"
                size="small"
              />
              <span class="stream-status" v-if="enableStream">
                (实时显示AI思考过程)
              </span>
              <span class="stream-status" v-else>
                (等待完整回复)
              </span>
              <el-button
                type="text"
                size="small"
                class="theme-button--text"
                @click="showModelInfo"
              >
                查看模型详情
              </el-button>
            </div>
          </div>
        </div>

        <!-- 聊天消息区域 -->
        <div class="chat-messages glass-card" ref="messageContainer" :style="{ paddingBottom: messagesPadding + 'px' }">
          <div
            v-for="message in messages"
            :key="message.id"
            :class="['message-item', message.role, {'streaming': message.streaming}]"
          >
            <div class="message-avatar">
              <i :class="message.role === 'user' ? 'el-icon-user' : 'el-icon-cpu'"></i>
            </div>
            <div class="message-content">
              <div class="message-header">
                <span class="message-role">
                  {{ message.role === 'user' ? ($store.state.user ? $store.state.user.username : '我') : 'AI助手' }}
                </span>
                <span v-if="message.model" class="message-model">
                  ({{ message.model }})
                </span>
                <span class="message-time">
                  {{ formatTime(message.timestamp) }}
                </span>
              </div>
              <!-- 思考过程（如果有的话） -->
              <div v-if="message.reasoning_content" class="reasoning-content" :class="{ 'collapsed': message.reasoningCollapsed }">
                <div class="reasoning-header" @click="toggleReasoning(message)">
                  <i :class="message.reasoningCollapsed ? 'el-icon-arrow-right' : 'el-icon-arrow-down'"></i>
                  <span>思考过程</span>
                </div>
                <div v-show="!message.reasoningCollapsed" class="reasoning-text">{{ message.reasoning_content }}</div>
              </div>
              <!-- 正式回复内容 -->
              <div class="message-text">{{ message.content }}</div>
              <!-- 使用:class动态绑定预览状态类，图片ID使用索引+URL唯一标识 -->
              <div v-if="message.image_url" class="message-image">
                <img
                  :src="message.image_url"
                  alt="图像"
                  @click="previewImage(message.image_url)"
                  :ref="'img-' + message.id" />
              </div>
              <div v-if="message.video_url" class="message-video">
                <video controls :src="message.video_url" class="message-video-player"></video>
              </div>
            </div>
          </div>
          
          <!-- 加载状态 -->
          <div v-if="loading" class="loading-message">
            <i class="el-icon-loading"></i>
            <span>AI正在思考中...</span>
          </div>
        </div>

        <!-- 底部输入区域 -->
        <div
          class="chat-input glass-card"
          ref="chatInput"
          :style="chatInputStyle"
        >
                  <div class="input-toolbar" :class="{'has-media': !!pendingMedia}">
                    <el-button
                      v-if="!isMobile"
                      icon="el-icon-picture"
                      size="small"
                      class="chat-btn"
                      @click="uploadImage"
                    >
                      上传图片
                    </el-button>
                    <el-button
                      v-if="!isMobile"
                      icon="el-icon-video-camera"
                      size="small"
                      class="chat-btn"
                      @click="uploadVideo"
                    >
                      上传视频
                    </el-button>
                    <input
                      ref="imageInput"
                      type="file"
                      accept="image/*"
                      style="display: none"
                      @change="handleImageUpload"
                    />
                    <input
                      ref="videoInput"
                      type="file"
                      accept="video/*"
                      style="display: none"
                      @change="handleVideoUpload"
                    />
                    
                    <!-- 显示已选媒体预览 -->
                    <div v-if="pendingMedia" class="pending-media">
                      <div class="pending-media-preview">
                        <img v-if="pendingMedia.type.startsWith('image/')"
                          :src="pendingMedia.previewUrl"
                          class="media-preview-thumbnail" />
                        <video v-else-if="pendingMedia.type.startsWith('video/')"
                          :src="pendingMedia.previewUrl"
                          class="media-preview-thumbnail"
                          controls></video>
                      </div>
                      <div class="pending-media-info">
                        <span>{{ pendingMedia.file.name }} ({{ formatFileSize(pendingMedia.file.size) }})</span>
                        <el-button
                          type="text"
                          icon="el-icon-delete"
                          class="theme-button--text"
                          @click="clearPendingMedia"></el-button>
                      </div>
                    </div>
                  </div>
          
          <div class="input-area">
            <el-input
              v-model="inputMessage"
              type="textarea"
              :rows="2"
              placeholder="请输入您的描述... (按Enter发送)"
              @keydown.enter.native="handleEnterSend"
              @keydown.shift.enter="() => {}"
            />
            <!-- 移动端：将“＋”按钮内置到输入框右侧 -->
            <button v-if="isMobile" class="plus-fab" v-popover:plusMenu aria-label="更多操作">＋</button>
            <!-- 桌面端：保留发送按钮；移动端去掉发送按钮，使用键盘发送 -->
            <el-button
              v-if="!isMobile"
              class="chat-btn chat-btn--send"
              :type="loading ? 'danger' : 'primary'"
              :icon="loading ? 'el-icon-close' : 'el-icon-s-promotion'"
              :loading="false"
              :disabled="((!inputMessage.trim() && !pendingMedia) || !selectedModel) && !loading"
              @click="loading ? cancelRequest() : sendMessage()"
            >
              {{ loading ? '中断' : '发送' }}
            </el-button>
          </div>

          <!-- 输入区右下角 + 按钮及菜单（移动端与PC均可用） -->
          <el-popover v-if="isMobile" ref="plusMenu" placement="top-end" trigger="click" @show="updateMessagesPadding" @hide="updateMessagesPadding">
            <div class="plus-menu">
              <el-button type="text" class="theme-button--text" @click="uploadImage">上传图片</el-button>
              <el-button type="text" class="theme-button--text" @click="uploadVideo">上传视频</el-button>
            </div>
          </el-popover>
        </div>
      </div>
    </div>

    <!-- 模型信息对话框 -->
    <el-dialog
      title="AI模型详情"
      :visible.sync="modelInfoVisible"
      width="600px"
    >
      <div class="model-details">
        <div 
          v-for="model in availableModels" 
          :key="model.name"
          class="model-detail-item"
        >
          <h4>{{ model.name }}</h4>
          <p>{{ model.description }}</p>
          <div class="model-specs">
            <span>上下文：{{ model.context }}</span>
            <span>输入价格：{{ model.input_price }}</span>
            <span>输出价格：{{ model.output_price }}</span>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 新建会话对话框 -->
    <el-dialog
      title="创建新会话"
      :visible.sync="newSessionVisible"
      width="400px"
    >
      <el-form :model="newSessionForm" label-width="80px">
        <el-form-item label="会话标题">
          <el-input 
            v-model="newSessionForm.title" 
            placeholder="请输入会话标题"
          />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button class="chat-btn" @click="newSessionVisible = false">取消</el-button>
        <el-button class="chat-btn" type="primary" @click="confirmCreateSession">创建</el-button>
      </div>
    </el-dialog>
    
    <!-- 移动端设置抽屉 -->
    <el-drawer
      title="对话设置"
      :visible.sync="settingsDrawerVisible"
      direction="rtl"
      size="86vw"
      custom-class="settings-drawer"
      :append-to-body="true"
      :wrapper-closable="!drawerSelectOpen"
      :close-on-press-escape="!drawerSelectOpen">
      <div class="drawer-section">
        <div class="section-title compact">
          <div class="section-accent"></div>
          <h3>选择模型</h3>
        </div>
        <el-select
          v-model="selectedModel"
          placeholder="请选择AI模型"
          filterable
          default-first-option
          style="width: 100%;"
          popper-class="model-select-popper"
          :popper-append-to-body="false"
          @visible-change="onDrawerSelectVisibleChange">
          <el-option
            v-for="model in availableModels"
            :key="model.name"
            :label="model.name"
            :value="model.name">
            <div class="model-info">
              <div class="model-name">{{ model.name }}</div>
              <div class="model-desc">{{ model.description }}</div>
            </div>
          </el-option>
        </el-select>

        <div style="margin-top: 14px; display: flex; align-items: center; gap: 10px;">
          <el-switch
            v-model="enableStream"
            active-text="流式返回"
            inactive-text="普通返回"
            size="small" />
          <span class="stream-status" v-if="enableStream">(实时显示AI思考过程)</span>
          <span class="stream-status" v-else>(等待完整回复)</span>
        </div>

        <div style="margin-top: 12px;">
          <el-button type="text" size="small" class="theme-button--text" @click="showModelInfo">查看模型详情</el-button>
        </div>
      </div>
    </el-drawer>

    <!-- 图片预览遮罩：在遮罩内渲染独立的预览图，避免被遮罩层压暗 -->
    <div v-if="imagePreviewVisible" class="image-preview-overlay" @click="closeImagePreview">
      <img :src="currentPreviewUrl" alt="预览图像" class="image-preview-img" @click.stop />
    </div>
  </div>
</template>

<script>
import { chatAPI } from '@/api/chat'

export default {
  name: 'AIChat',
  data() {
    return {
      // 会话相关
      sessions: [],
      currentSession: null,
      messages: [],
      sidebarVisible: false, // 控制侧边栏是否可见，默认隐藏
      
      // 模型相关
      availableModels: [],
      selectedModel: '',
      
      // 输入相关
      inputMessage: '',
      loading: false,
      
      // 多媒体相关
      pendingMedia: null, // {file: File, type: string, previewUrl: string, uploadPath: string, visitUrl: string}
      supportedMediaTypes: {
        image: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
        video: ['video/mp4', 'video/webm', 'video/quicktime']
      },
      
      // 图片预览相关
      imagePreviewVisible: false,
      currentPreviewUrl: null, // 当前预览的URL
      activeImageRef: null, // 当前激活的图片元素引用
      
      // 流式返回设置
      enableStream: true, // 是否启用流式返回
      streamingInProgress: false, // 是否正在流式处理中
      
      // 请求控制
      currentRequest: null, // 当前请求对象，用于取消请求
      
      // 对话框
      modelInfoVisible: false,
      newSessionVisible: false,
      settingsDrawerVisible: false,
      drawerSelectOpen: false,
      newSessionForm: {
        title: ''
      },

      // 设备状态
      isMobile: false,
      // 消息区底部内边距（用于避免被输入区遮挡，移动端动态计算）
      messagesPadding: 0,
      // 软键盘遮挡偏移，仅移动端使用（只移动输入区，不影响菜单栏）
      keyboardOffset: 0,
      // 键盘显隐状态（与 keyboardOffset 配合，便于事件总线触发）
      keyboardShown: false,

      // 观察器
      chatInputObserver: null,
      // iOS/Android 键盘补偿轮询定时器
      keyboardOffsetEnsureTimer: null,
      // 上次已知的键盘抬升高度（用于首次聚焦时的立即回退抬升，避免被遮挡）
      lastKeyboardOffset: 0
    }
  },
  
  async mounted() {
    // 检查认证状态
    console.log('AIChat mounted - 认证状态:', this.$store.getters.isAuthenticated)
    console.log('AIChat mounted - Token:', this.$store.state.token)
    
    // 加载数据但不自动选择会话，保持新会话状态
    await this.loadData()
    
    // 禁用输入框的自动滚动
    this.preventTextareaScroll()
    
    // 从本地存储加载侧边栏状态，默认为隐藏
    const savedSidebarState = localStorage.getItem('aichat_sidebar_visible');
    this.sidebarVisible = savedSidebarState === 'true';
    
    // 确保进入页面时是一个新会话状态
    this.currentSession = null;
    this.messages = [];
    console.log('进入AI聊天页面，显示新会话状态');
    
    // 添加ESC键盘事件监听
    document.addEventListener('keydown', this.handleKeyDown);
    
    // 添加窗口大小变化监听，以保持预览居中
    this.updateIsMobile();
    window.addEventListener('resize', this.handleWindowResize);

    // 移动端键盘发送提示与布局稳定
    this.enhanceMobileTextareaAttributes();

    // 初始化计算消息区底部内边距
    this.updateMessagesPadding();

    // 监听 chatInput 尺寸变化，动态更新消息区底部留白
    this.$nextTick(() => {
      const el = this.$refs.chatInput;
      if (el && 'ResizeObserver' in window) {
        this.chatInputObserver = new ResizeObserver(() => {
          this.updateMessagesPadding();
        });
        this.chatInputObserver.observe(el);
      }
      // 监听移动端可视区域变化（软键盘弹出/收起）
      if (window.visualViewport) {
        window.visualViewport.addEventListener('resize', this.updateMessagesPadding);
        window.visualViewport.addEventListener('scroll', this.updateMessagesPadding);
        window.visualViewport.addEventListener('resize', this.updateKeyboardOffset);
        window.visualViewport.addEventListener('scroll', this.updateKeyboardOffset);
        // 初始化键盘偏移
        this.updateKeyboardOffset();
      }
      // 文本域聚焦/失焦时更新（键盘弹出场景）
      const ta = this.$el && this.$el.querySelector('.el-textarea__inner');
      if (ta) {
        ta.addEventListener('focus', () => {
          setTimeout(() => {
            this.updateMessagesPadding();
            try {
              this.keyboardShown = true;
              this.$eventBus && this.$eventBus.$emit('keyboard-open');
            } catch (e) {
              // eslint-disable-next-line no-console
              console.warn('AIChat textarea focus emit error:', e);
            }
            // 立即尝试计算一次键盘偏移，并开启短时轮询，避免首次聚焦时遮挡
            this.updateKeyboardOffset();
            this.startKeyboardOffsetEnsure();
          }, 50);
        });
        ta.addEventListener('blur', () => {
          setTimeout(() => {
            this.updateMessagesPadding();
            try {
              this.keyboardShown = false;
              this.$eventBus && this.$eventBus.$emit('keyboard-close');
            } catch (e) {
              // eslint-disable-next-line no-console
              console.warn('AIChat textarea blur emit error:', e);
            }
            // 停止键盘偏移轮询；不立即将 keyboardOffset 置 0，避免下次聚焦时无回退值可用
            this.stopKeyboardOffsetEnsure();
            // 让 visualViewport 的变更驱动 updateKeyboardOffset 自行归零
            setTimeout(() => {
              try { this.updateKeyboardOffset(); } catch (err) {
                // eslint-disable-next-line no-console
                console.debug('AIChat blur deferred updateKeyboardOffset error:', err);
              }
            }, 80);
          }, 50);
        });
      }
    });
  },
  
  beforeDestroy() {
    // 移除事件监听
    document.removeEventListener('keydown', this.handleKeyDown);
    window.removeEventListener('resize', this.handleWindowResize);
    // 断开观察器与事件
    try {
      if (this.chatInputObserver) {
        this.chatInputObserver.disconnect();
        this.chatInputObserver = null;
      }
      if (window.visualViewport) {
        window.visualViewport.removeEventListener('resize', this.updateMessagesPadding);
        window.visualViewport.removeEventListener('scroll', this.updateMessagesPadding);
        window.visualViewport.removeEventListener('resize', this.updateKeyboardOffset);
        window.visualViewport.removeEventListener('scroll', this.updateKeyboardOffset);
      }
      if (this.keyboardOffsetEnsureTimer) {
        clearTimeout(this.keyboardOffsetEnsureTimer);
        this.keyboardOffsetEnsureTimer = null;
      }
    } catch (e) {
      // 确保非空catch以通过eslint，并记录清理异常
      // eslint-disable-next-line no-console
      console.warn('AIChat beforeDestroy cleanup error:', e);
    }
  },
  
  watch: {
    isMobile() {
      this.$nextTick(() => {
        this.updateKeyboardOffset();
        this.updateMessagesPadding();
      });
    },
    sidebarVisible() {
      this.$nextTick(() => this.updateMessagesPadding());
    },
    'messages.length'() {
      // 消息数量变化时确保底部留白正确
      this.$nextTick(() => this.updateMessagesPadding());
    }
  },
  
  computed: {
    // 移动端将输入条抬高至底部导航（App.vue .mobile-bottom-nav，高度60px）之上
    // 键盘弹出时（App会隐藏底部导航），自动还原为贴底并跟随 keyboardOffset 上移
    chatInputStyle() {
      if (!this.isMobile) return {};
      const style = {};
      // 键盘打开：贴底并随键盘上移
      if (this.keyboardShown || this.keyboardOffset > 0) {
        // 优先使用实时 keyboardOffset；若暂不可得，使用上次已知高度做“立即抬升”回退，避免一瞬间被遮挡
        const lift = this.keyboardOffset > 0
          ? this.keyboardOffset
          : (this.keyboardShown && this.lastKeyboardOffset > 0 ? this.lastKeyboardOffset : 0);
        if (lift > 0) {
          style.transform = `translateY(-${lift}px)`;
        }
        style.bottom = '0px';
        return style;
      }
      // 键盘关闭：紧贴底部菜单顶边（动态读取 App.vue 中 .mobile-bottom-nav 实际高度）
      let navH = 60;
      try {
        const navEl = document.querySelector('.mobile-bottom-nav');
        if (navEl && navEl.offsetHeight) navH = navEl.offsetHeight;
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('AIChat chatInputStyle read .mobile-bottom-nav height error:', e);
      }
      style.bottom = navH + 'px';
      // 取消额外底部内边距，视觉上“紧贴”菜单栏
      style.paddingBottom = '0px';
      return style;
    }
  },
  
  methods: {
    // 切换思考过程的折叠状态
    toggleReasoning(message) {
      // 为消息对象添加reasoningCollapsed属性（如果不存在），确保默认展开
      if (typeof message.reasoningCollapsed === 'undefined') {
        this.$set(message, 'reasoningCollapsed', false);
      }
      
      // 切换折叠状态
      this.$set(message, 'reasoningCollapsed', !message.reasoningCollapsed);
    },
    
    // 处理窗口大小变化
    handleWindowResize() {
      // 更新移动端状态
      this.updateIsMobile();
      // 更新移动端输入属性
      this.enhanceMobileTextareaAttributes();

      // 如果有预览激活，重新调整预览图片的位置（占位：避免 no-empty）
      if (this.imagePreviewVisible && this.activeImageRef) {
        // eslint-disable-next-line no-console
        console.debug('AIChat handleWindowResize: imagePreview active; no reposition logic needed currently');
      }

      // 更新消息区底部内边距，避免被输入区遮挡
      this.updateMessagesPadding();
    },

    // 计算软键盘遮挡偏移，仅在移动端启用，使仅输入区跟随上移；并通过事件总线通知 App 隐藏底部菜单
    updateKeyboardOffset() {
      try {
        if (!this.isMobile || !window.visualViewport) {
          const wasOpen = this.keyboardShown;
          this.keyboardOffset = 0;
          if (wasOpen) {
            this.keyboardShown = false;
            try {
              this.$eventBus && this.$eventBus.$emit('keyboard-close');
            } catch (err) {
              // eslint-disable-next-line no-console
              console.warn('AIChat emit keyboard-close error:', err);
            }
          }
          return;
        }
        const vv = window.visualViewport;
        // 底部遮挡 = 布局高度 - (可视高度 + 可视偏移Top)
        const bottomOverlap = Math.max(0, window.innerHeight - (vv.height + vv.offsetTop));
        const offset = Math.round(bottomOverlap);
        const wasOpen = this.keyboardShown;
        const isOpen = offset > 0;

        this.keyboardOffset = offset;
        if (offset > 0) {
          // 记录上次已知的键盘高度，供下次首次聚焦时“立即抬升”回退使用
          this.lastKeyboardOffset = offset;
        }
 
        if (isOpen && !wasOpen) {
          this.keyboardShown = true;
          try {
            this.$eventBus && this.$eventBus.$emit('keyboard-open');
          } catch (err) {
            // eslint-disable-next-line no-console
            console.warn('AIChat emit keyboard-open error:', err);
          }
        } else if (!isOpen && wasOpen) {
          this.keyboardShown = false;
          try {
            this.$eventBus && this.$eventBus.$emit('keyboard-close');
          } catch (err) {
            // eslint-disable-next-line no-console
            console.warn('AIChat emit keyboard-close error:', err);
          }
        }
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('updateKeyboardOffset error:', e);
        this.keyboardOffset = 0;
      }
    },
    
    // 在键盘聚焦初期进行多次补偿，直至获得有效的 keyboardOffset 或超时
    startKeyboardOffsetEnsure() {
      try {
        this.stopKeyboardOffsetEnsure();
        let tries = 0;
        const maxTries = 12; // ~12 * 60ms ≈ 720ms
        const tick = () => {
          tries += 1;
          this.updateKeyboardOffset();
          this.updateMessagesPadding();
          if (this.keyboardOffset > 0 || !this.keyboardShown || tries >= maxTries) {
            if (this.keyboardOffsetEnsureTimer) {
              clearTimeout(this.keyboardOffsetEnsureTimer);
              this.keyboardOffsetEnsureTimer = null;
            }
            return;
          }
          this.keyboardOffsetEnsureTimer = setTimeout(tick, 60);
        };
        this.keyboardOffsetEnsureTimer = setTimeout(tick, 0);
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('startKeyboardOffsetEnsure error:', e);
      }
    },
    
    stopKeyboardOffsetEnsure() {
      try {
        if (this.keyboardOffsetEnsureTimer) {
          clearTimeout(this.keyboardOffsetEnsureTimer);
          this.keyboardOffsetEnsureTimer = null;
        }
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('stopKeyboardOffsetEnsure error:', e);
      }
    },
    
    // 计算是否为移动端
    updateIsMobile() {
      try {
        this.isMobile = window.innerWidth <= 768;
      } catch (e) {
        this.isMobile = false;
      }
    },

    // 打开设置抽屉（仅移动端展示按钮）
    openSettingsDrawer() {
      this.settingsDrawerVisible = true;
    },

    // 下拉显隐时，暂时禁用抽屉遮罩点击与 ESC 关闭，避免弹层抖动与误关
    onDrawerSelectVisibleChange(visible) {
      try {
        this.drawerSelectOpen = !!visible;
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('onDrawerSelectVisibleChange error:', e);
      }
    },
    
    // 处理键盘事件 - 用于ESC关闭图片预览
    handleKeyDown(event) {
      if (event.key === 'Escape' && this.imagePreviewVisible) {
        this.closeImagePreview();
      }
    },
    
    // 预览图片 - 使用已经加载的图片元素，不再创建新的img元素
    previewImage(url) {
      if (!url) return;
      this.currentPreviewUrl = url;
      this.imagePreviewVisible = true;
      // 禁止页面滚动
      document.body.style.overflow = 'hidden';
    },
    
    // 关闭图片预览
    closeImagePreview() {
      this.imagePreviewVisible = false;
      this.currentPreviewUrl = null;
      this.activeImageRef = null;
      // 恢复页面滚动
      document.body.style.overflow = '';
    },
    
    // 取消请求
    cancelRequest() {
      if (this.currentRequest) {
        console.log('取消当前请求')
        this.currentRequest.abort()
        this.currentRequest = null
        this.loading = false
        this.$message.info('已取消AI回复')
      }
    },
    
    // 加载数据
    async loadData() {
      try {
        console.log('开始加载数据...')
        await Promise.all([
          this.loadSessions(),
          this.loadModels()
        ])
        console.log('数据加载完成')
      } catch (error) {
        console.error('加载数据失败:', error)
        this.$message.error('加载数据失败: ' + (error.message || '未知错误'))
      }
    },
    
    // 加载会话列表
    async loadSessions() {
      try {
        console.log('开始加载会话列表...')
        const response = await chatAPI.getSessions()
        console.log('会话列表响应:', response)
        this.sessions = response.sessions || []
        
        // 不自动选择会话，保持为新会话状态
        // 注释掉原来的自动选择逻辑，保持currentSession为null表示新会话状态
      } catch (error) {
        console.error('加载会话列表失败:', error)
        throw error
      }
    },
    
    // 加载模型列表
    async loadModels() {
      const response = await chatAPI.getAvailableModels()
      this.availableModels = response.models || []
      
      // 默认选择第一个模型
      if (this.availableModels.length > 0) {
        this.selectedModel = this.availableModels[0].name
      }
    },
    
    // 选择会话
    async selectSession(session) {
      this.currentSession = session
      await this.loadSessionMessages(session.id)
      if (this.isMobile) {
        this.sidebarVisible = false
        localStorage.setItem('aichat_sidebar_visible', 'false')
      }
    },
    
    // 加载会话消息
    async loadSessionMessages(sessionId) {
      try {
        const response = await chatAPI.getSessionMessages(sessionId)
        
        // 处理后端返回的消息，确保时间戳格式正确
        const messages = (response.messages || []).map(msg => {
          // 确保每条消息都有有效的时间戳
          let timestamp = msg.timestamp || msg.created_at || msg.time;
          
          // 如果时间戳无效，使用当前时间
          if (!timestamp || new Date(timestamp).toString() === 'Invalid Date') {
            console.warn('消息时间戳无效:', msg.id, timestamp);
            timestamp = new Date().toISOString();
          }
          
          // 返回处理后的消息对象并初始化思考过程为展开状态
          return {
            ...msg,
            timestamp: timestamp,
            reasoningCollapsed: false // 默认展开思考过程
          };
        });
        
        this.messages = messages;
        
        // 如果有消息并且模型列表已经加载，同步最后一条AI消息的模型到模型选择框
        if (messages.length > 0 && this.availableModels.length > 0) {
          // 过滤出所有AI助手发送的消息
          const assistantMessages = messages.filter(msg => msg.role === 'assistant');
          
          // 如果有助手消息，找到最后一条（按照时间排序，最新的在最后）
          if (assistantMessages.length > 0) {
            const lastAssistantMessage = assistantMessages[assistantMessages.length - 1];
            
            // 如果消息中包含模型信息，更新selectedModel
            if (lastAssistantMessage.model) {
              console.log('同步最后使用的模型:', lastAssistantMessage.model);
              this.selectedModel = lastAssistantMessage.model;
            }
          }
        }
        
        this.$nextTick(() => {
          // 加载历史消息时直接跳转到底部，不使用平滑滚动
          this.scrollToBottom(false)
        })
      } catch (error) {
        console.error('加载消息失败:', error);
        this.$message.error('加载消息失败')
      }
    },
    
    // 发送消息
    async sendMessage() {
      console.log('发送消息方法被调用')
      // 检查是否有内容或待发送媒体，并且已选择模型
      if ((!this.inputMessage.trim() && !this.pendingMedia) || !this.selectedModel) {
        console.log('没有内容或未选择模型，不发送')
        return
      }
      
      if (!this.currentSession) {
        // 首次发送消息时，使用消息前50字作为标题创建会话
        const rawTitle = (this.inputMessage || '').trim();
        const title = rawTitle
          ? (rawTitle.length > 50 ? rawTitle.substring(0, 50) + '...' : rawTitle)
          : '新会话';
        
        try {
          const response = await chatAPI.createSession(title);
          this.currentSession = response.session;
          console.log('自动创建会话成功:', this.currentSession);
        } catch (error) {
          console.error('自动创建会话失败:', error);
          this.$message.error('创建会话失败');
          return;
        }
      }
      
      this.loading = true
      
      try {
        let mediaInfo = null
        let imageUrl = ''
        let videoUrl = ''
        
        // 如果有待上传的媒体，先上传
        if (this.pendingMedia) {
          try {
            this.$message.info('正在上传媒体文件...')
            mediaInfo = await this.uploadMediaFile()
            
            // 设置图片或视频URL
            if (mediaInfo) {
              if (mediaInfo.type.startsWith('image/')) {
                imageUrl = mediaInfo.url
              } else if (mediaInfo.type.startsWith('video/')) {
                videoUrl = mediaInfo.url
              }
            }
          } catch (error) {
            this.$message.error('媒体文件上传失败，将只发送文本')
            console.error('媒体上传失败:', error)
          }
        }
        
        // 添加用户消息到界面
        const userMessage = {
          id: Date.now().toString(),
          role: 'user',
          content: this.inputMessage,
          timestamp: new Date().toISOString()
        }
        
        // 如果有媒体信息，添加到消息中
        if (imageUrl) {
          userMessage.image_url = imageUrl
        }
        if (videoUrl) {
          userMessage.video_url = videoUrl
        }
        
        this.messages.push(userMessage)
        
        // 清空输入
        const messageText = this.inputMessage
        this.inputMessage = ''
        this.clearPendingMedia()

        // 移动端：发送后收起键盘，防止页面放大与布局跳动
        if (this.isMobile) {
          const ta = this.$el && this.$el.querySelector('.el-textarea__inner');
          if (ta) ta.blur();
        }
        
        // 滚动到底部 - 仅在用户发送新消息时
        const shouldScroll = true // 这里可以添加条件判断是否需要滚动
        if (shouldScroll) {
          this.$nextTick(() => {
            this.scrollToBottom()
          })
        }
        
        // 使用流式发送消息
        if (this.enableStream) {
          console.log('使用流式发送消息')
          await this.sendMessageStream(messageText, imageUrl, videoUrl)
        } else {
          console.log('使用非流式发送消息')
          await this.sendMessageNormal(messageText, imageUrl, videoUrl)
        }
        
      } catch (error) {
        console.error('发送消息失败:', error)
        this.$message.error('发送消息失败')
      } finally {
        this.loading = false
      }
    },
    
    // 流式发送消息
    async sendMessageStream(messageText, imageUrl = '', videoUrl = '') {
      const aiMessages = new Map() // 存储每个模型的AI消息
      
      try {
        console.log('开始流式请求，会话ID:', this.currentSession.id, '模型:', this.selectedModel)
        
        // 标记当前正在流式处理
        this.streamingInProgress = true
        
        // 创建请求控制器并保存引用以便于取消
        this.currentRequest = chatAPI.sendMessageStream(
          this.currentSession.id,
          messageText,
          this.selectedModel, // 将单个模型名称包装成数组
          imageUrl, // 图片URL
          videoUrl, // 视频URL
          // onChunk - 处理流式数据块
          (chunkData) => {
            console.log('接收到数据块:', chunkData)
            // 根据实际SSE数据结构解构变量
            const message_id = chunkData.message_id
            const model = chunkData.model
            const content = chunkData.content
            
            // 记录增量内容
            console.log(`收到增量内容: "${content}" (长度: ${content.length})`);
            
            // 记录收到的每个数据块
            console.log(`收到数据块: ID=${message_id}, 内容="${content}", 模型=${model}`)
            
            if (!aiMessages.has(message_id)) {
              console.log('创建新消息:', message_id, model)
              // 创建新消息对象
              const newMessage = {
                id: message_id,
                role: 'assistant',
                model: model,
                content: '',
                timestamp: new Date().toISOString(),
                streaming: true // 标记为流式传输中
              }
              // 保存到Map
              aiMessages.set(message_id, newMessage)
              // 添加到消息列表
              this.messages.push(newMessage)
            }
            
            // 获取当前已有的内容和思考过程
            let currentContent = "";
            let currentReasoningContent = "";
            if (aiMessages.has(message_id)) {
              currentContent = aiMessages.get(message_id).content || "";
              currentReasoningContent = aiMessages.get(message_id).reasoning_content || "";
            }
            
            // 根据SSE格式，每次chunk包含的是增量内容，需要累积
            const updatedContent = currentContent + content;
            const updatedReasoningContent = currentReasoningContent + (chunkData.reasoning_content || "");
            
            console.log(`累积消息: 当前长度=${currentContent.length}, 新增=${content.length}, 总计=${updatedContent.length}`);
            if (chunkData.reasoning_content) {
              console.log(`累积思考过程: 当前长度=${currentReasoningContent.length}, 新增=${chunkData.reasoning_content.length}, 总计=${updatedReasoningContent.length}`);
            }
            
            // 查找对应的消息
            const index = this.messages.findIndex(msg => msg.id === message_id)
            if (index !== -1) {
              // 使用Vue的响应式方法更新累积内容
              // 为了确保响应式更新，先创建一个新对象，然后整体替换
              const updatedMessage = {
                ...this.messages[index],
                content: updatedContent,
                reasoning_content: updatedReasoningContent,
                reasoningCollapsed: this.messages[index].reasoningCollapsed !== undefined ? this.messages[index].reasoningCollapsed : false,
                streaming: true
              }
              
              // 替换整个对象，保证Vue能检测到变化
              this.$set(this.messages, index, updatedMessage)
              
              // 调试输出
              console.log('更新消息:', updatedMessage.id.slice(0, 6),
                '累积内容:', updatedContent.length,
                '最新部分:', content);
            }
            
            // 更新Map中的消息
            if (aiMessages.has(message_id)) {
              const mapMessage = aiMessages.get(message_id)
              mapMessage.content = updatedContent
              mapMessage.reasoning_content = updatedReasoningContent
            }
            
            // 立即滚动到底部
            this.$nextTick(() => {
              this.scrollToBottom()
              
              // 调试信息 - 显示当前streaming状态
              console.log('消息streaming状态:', this.messages.map(m => ({
                id: m.id.slice(0, 6) + '...',
                role: m.role,
                streaming: m.streaming,
                content_length: m.content ? m.content.length : 0
              })))
            })
          },
          // onStart - 处理开始事件
          (startData) => {
            console.log('开始生成回复:', startData)
            const { message_id, model, timestamp } = startData
            
            // 确保有有效的时间戳
            const validTimestamp = timestamp || new Date().toISOString()
            console.log('使用时间戳:', validTimestamp);
            
            // 如果消息不存在，创建一个新消息
            if (!aiMessages.has(message_id)) {
              const newMessage = {
                id: message_id,
                role: 'assistant',
                model: model,
                content: '',
                reasoning_content: '',
                timestamp: validTimestamp,
                reasoningCollapsed: false, // 默认展开思考过程
                streaming: true
              }
              
              aiMessages.set(message_id, newMessage)
              this.messages.push(newMessage)
              console.log('创建初始消息:', message_id, model, '时间戳:', validTimestamp)
            }
          },
          // onComplete - 处理完成事件
          (completeData) => {
            console.log('完成响应:', completeData)
            const { message_id, content, image_url } = completeData
            
            // 找到消息对象
            const index = this.messages.findIndex(msg => msg.id === message_id)
            if (index !== -1) {
              // 替换整个对象以确保响应式更新
              this.$set(this.messages, index, {
                ...this.messages[index],
                content: content || this.messages[index].content,
                reasoning_content: completeData.reasoning_content || this.messages[index].reasoning_content,
                image_url: image_url || this.messages[index].image_url,
                streaming: false // 标记流式结束
              })
            }
            
            // 更新Map中的消息
            const message = aiMessages.get(message_id)
            if (message) {
              message.content = content || message.content
              message.streaming = false
              if (image_url) message.image_url = image_url
            }
          },
          // onError - 处理错误事件
          (errorData) => {
            console.error('AI回复错误:', errorData)
            this.$message.error(errorData.error || 'AI回复失败')
            this.loading = false // 出错时结束加载状态
            this.currentRequest = null
          },
          // onEnd - 处理结束事件
          (endData) => {
            console.log('流式回复结束:', endData)
            
            // 移除所有流式标记
            this.messages.forEach((msg, idx) => {
              if (msg.streaming) {
                this.$set(this.messages, idx, {
                  ...msg,
                  streaming: false
                })
              }
            })
            
            // 刷新会话列表
            this.loadSessions()
            
            // 结束加载状态
            this.loading = false
            this.streamingInProgress = false
            this.currentRequest = null
          }
        )
        
        // 等待请求完成
        const result = await this.currentRequest.requestPromise
        
        // 检查是否被取消
        if (result && result.aborted) {
          console.log('请求已被用户取消')
          return
        }
      } catch (error) {
        console.error('流式发送失败:', error)
        throw error
      }
    },
    
    // 普通发送消息
    async sendMessageNormal(messageText, imageUrl = '', videoUrl = '') {
      try {
        console.log('开始普通请求，会话ID:', this.currentSession.id)
        
        // 创建可取消的请求控制器
        this.currentRequest = chatAPI.sendMessage(
          this.currentSession.id,
          messageText,
          this.selectedModel, // 将单个模型名称包装成数组
          imageUrl,
          videoUrl
        )
        
        // 等待响应
        const response = await this.currentRequest.requestPromise
        
        // 检查是否被取消
        if (response && response.aborted) {
          console.log('请求已被用户取消')
          return
        }
        
        console.log('非流式响应数据:', response)
        
        // 添加AI回复到界面
        if (response.ai_messages && response.ai_messages.length > 0) {
          console.log('接收到AI消息:', response.ai_messages.length, '条')
          for (const aiMessage of response.ai_messages) {
            // 确保时间戳有效
            let timestamp = aiMessage.created_at || aiMessage.timestamp || aiMessage.time;
            
            // 如果时间戳无效，使用当前时间
            if (!timestamp || new Date(timestamp).toString() === 'Invalid Date') {
              console.warn('AI消息时间戳无效:', aiMessage.id, timestamp);
              timestamp = new Date().toISOString();
            }
            
            const newMessage = {
              id: aiMessage.id,
              role: 'assistant',
              model: aiMessage.model,
              content: aiMessage.content,
              reasoning_content: aiMessage.reasoning_content,
              image_url: aiMessage.image_url,
              timestamp: timestamp,
              reasoningCollapsed: false // 默认展开思考过程
            }
            
            console.log('添加消息到界面:', newMessage.id, 'timestamp:', newMessage.timestamp);
            this.messages.push(newMessage)
          }
          
          // 强制更新视图
          this.$forceUpdate()
        } else {
          console.warn('未收到AI消息或消息为空')
          this.$message.warning('未收到AI回复')
        }
        
        // 滚动到底部
        this.$nextTick(() => {
          this.scrollToBottom()
        })
        
        // 刷新会话列表
        await this.loadSessions()
        
      } catch (error) {
        console.error('普通发送失败:', error)
        this.$message.error(`发送消息失败: ${error.message || '未知错误'}`)
      } finally {
        // 无论成功或失败，都结束加载状态
        this.loading = false
        this.currentRequest = null
      }
    },
    
    // 创建新会话（在历史会话中触发，打开对话框）
    createNewSession() {
      // 统一“新建会话”行为：清空当前聊天，由首条消息自动创建会话
      if (!this.currentSession && this.messages.length === 0) {
        // 已处于新会话
        if (this.isMobile) {
          this.sidebarVisible = false;
          localStorage.setItem('aichat_sidebar_visible', 'false');
        }
        // this.$message.info('已在新会话页');
        return;
      }
      this.currentSession = null;
      this.messages = [];
      if (this.isMobile) {
        this.sidebarVisible = false;
        localStorage.setItem('aichat_sidebar_visible', 'false');
      }
      this.$nextTick(() => {
        this.scrollToBottom(false);
      });
      // this.$message.success('已切换到新会话');
    },
    
    // 确认创建会话
    async confirmCreateSession() {
      if (!this.newSessionForm.title.trim()) {
        this.$message.warning('请输入会话标题')
        return
      }
      
      try {
        const response = await chatAPI.createSession(this.newSessionForm.title)
        this.currentSession = response.session
        this.messages = []
        this.newSessionForm.title = ''
        this.newSessionVisible = false
        
        // 刷新会话列表
        await this.loadSessions()
        
        this.$message.success('会话创建成功')
      } catch (error) {
        this.$message.error('创建会话失败')
      }
    },
    
    // 删除会话
    async deleteSession(sessionId) {
      try {
        await this.$confirm('确定要删除这个会话吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        
        await chatAPI.deleteSession(sessionId)
        
        // 如果删除的是当前会话，清空当前会话
        if (this.currentSession && this.currentSession.id === sessionId) {
          this.currentSession = null
          this.messages = []
        }
        
        // 刷新会话列表
        await this.loadSessions()
        
        this.$message.success('会话删除成功')
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('删除会话失败')
        }
      }
    },
    
    // 刷新会话列表
    async refreshSessions() {
      await this.loadSessions()
      this.$message.success('刷新成功')
    },
    
    // 显示模型信息
    showModelInfo() {
      this.modelInfoVisible = true
    },
    
    // 上传图片
    uploadImage() {
      this.$refs.imageInput.click()
    },
    
    // 上传视频
    uploadVideo() {
      this.$refs.videoInput.click()
    },
    
    // 处理图片上传
    async handleImageUpload(event) {
      const file = event.target.files[0]
      if (!file) return
      
      try {
        // 检查文件类型
        if (!this.supportedMediaTypes.image.includes(file.type)) {
          this.$message.error('不支持的图片格式')
          return
        }
        
        // 创建预览URL
        const previewUrl = URL.createObjectURL(file)
        
        // 生成唯一文件路径
        const fileExt = file.name.split('.').pop()
        const uploadPath = `${this.generateUuid()}.${fileExt}`
        
        // 保存待上传的媒体信息
        this.pendingMedia = {
          file,
          type: file.type,
          previewUrl,
          uploadPath,
          visitUrl: null
        }
        
        this.$message.success('图片已选择，发送消息时将自动上传')
        this.$nextTick(() => this.updateMessagesPadding())
      } catch (error) {
        console.error('处理图片上传失败:', error)
        this.$message.error('处理图片失败')
        this.clearPendingMedia()
      }
    },
    
    // 处理视频上传
    async handleVideoUpload(event) {
      const file = event.target.files[0]
      if (!file) return
      
      try {
        // 检查文件类型
        if (!this.supportedMediaTypes.video.includes(file.type)) {
          this.$message.error('不支持的视频格式')
          return
        }
        
        // 检查文件大小，限制为100MB
        if (file.size > 100 * 1024 * 1024) {
          this.$message.error('视频文件不能超过100MB')
          return
        }
        
        // 创建预览URL
        const previewUrl = URL.createObjectURL(file)
        
        // 生成唯一文件路径
        const fileExt = file.name.split('.').pop()
        const uploadPath = `${this.generateUuid()}.${fileExt}`
        
        // 保存待上传的媒体信息
        this.pendingMedia = {
          file,
          type: file.type,
          previewUrl,
          uploadPath,
          visitUrl: null
        }
        
        this.$message.success('视频已选择，发送消息时将自动上传')
        this.$nextTick(() => this.updateMessagesPadding())
      } catch (error) {
        console.error('处理视频上传失败:', error)
        this.$message.error('处理视频失败')
        this.clearPendingMedia()
      }
    },
    
    // 清除待上传媒体
    clearPendingMedia() {
      if (this.pendingMedia && this.pendingMedia.previewUrl) {
        URL.revokeObjectURL(this.pendingMedia.previewUrl)
      }
      this.pendingMedia = null
      this.$nextTick(() => this.updateMessagesPadding())
    },
    
    // 格式化文件大小
    formatFileSize(bytes) {
      if (bytes < 1024) {
        return bytes + ' B'
      } else if (bytes < 1024 * 1024) {
        return (bytes / 1024).toFixed(1) + ' KB'
      } else if (bytes < 1024 * 1024 * 1024) {
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
      } else {
        return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
      }
    },
    
    // 生成UUID
    generateUuid() {
      return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0
        const v = c === 'x' ? r : (r & 0x3 | 0x8)
        return v.toString(16)
      })
    },
    
    // 上传媒体文件
    async uploadMediaFile() {
      if (!this.pendingMedia) return null
      
      try {
        // 获取上传URL和访问URL
        const { fileAPI } = require('@/api/chat')
        const response = await fileAPI.getPublicUploadURL(this.pendingMedia.uploadPath)
        
        console.log('获取上传URL响应:', response)
        
        // 上传文件
        await fileAPI.uploadFile(response.upload_url, response.method, this.pendingMedia.file)
        
        // 直接使用getPublicUploadURL返回的visit_url
        // 返回访问信息
        return {
          type: this.pendingMedia.type,
          url: response.visit_url  // 直接使用upload接口返回的visit_url
        }
      } catch (error) {
        console.error('上传媒体文件失败:', error)
        throw new Error('上传媒体文件失败')
      }
    },
    
    // 分享会话
    shareSession() {
      if (this.currentSession) {
        // TODO: 实现分享功能
        this.$message.info('分享功能待实现')
      } else {
        this.$message.warning('请先选择一个会话')
      }
    },
    
    // 显示更多选项
    showMoreOptions() {
      // TODO: 实现更多选项菜单
      this.$message.info('更多选项功能待实现')
    },
    
    // 切换侧边栏显示状态
    toggleSidebar() {
      this.sidebarVisible = !this.sidebarVisible;
      // 保存状态到本地存储，以便刷新后保持状态
      localStorage.setItem('aichat_sidebar_visible', this.sidebarVisible);
      
      // 侧边栏状态变化后重新计算布局
      this.$nextTick(() => {
        // 可以在这里处理任何布局调整
        console.log("侧边栏状态变更为:", this.sidebarVisible ? "显示" : "隐藏");
      });
    },
    
    // 滚动到底部
    scrollToBottom(smooth = true) {
      const doScroll = () => {
        // 先保证底部留白为最新
        this.updateMessagesPadding();
        const container = this.$refs.messageContainer
        if (container) {
          container.scrollTo({
            top: container.scrollHeight,
            behavior: smooth ? 'smooth' : 'auto'
          });
        }
      };
      // 等待DOM包含新消息后再滚动
      this.$nextTick(doScroll);
    },
    
    // 禁用输入框的自动滚动
    preventTextareaScroll() {
      // 查找所有textarea元素并添加事件监听
      this.$nextTick(() => {
        const textareas = document.querySelectorAll('.chat-input textarea');
        textareas.forEach(textarea => {
          textarea.addEventListener('keydown', (e) => {
            // 阻止Enter键的默认滚动行为
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
            }
          });
        });
      });
    },

    // 移动端输入属性增强：显示键盘“发送”并保持布局稳定
    enhanceMobileTextareaAttributes() {
      this.$nextTick(() => {
        const ta = this.$el && this.$el.querySelector('.el-textarea__inner');
        if (!ta) return;
        if (this.isMobile) {
          ta.setAttribute('enterkeyhint', 'send');
          ta.setAttribute('inputmode', 'text');
          ta.setAttribute('autocapitalize', 'off');
          ta.setAttribute('autocomplete', 'off');
        } else {
          ta.removeAttribute('enterkeyhint');
          ta.removeAttribute('inputmode');
        }
      });
    },

    // 动态计算消息区底部内边距，确保输入区不会遮挡内容
    // 使用“实际重叠量”而非仅仅使用输入区高度，适配不同设备/键盘场景
    updateMessagesPadding() {
      this.$nextTick(() => {
        const container = this.$refs.messageContainer;
        const input = this.$refs.chatInput;
        if (!container || !input) {
          this.messagesPadding = 0;
          return;
        }

        // 非移动端不需要额外留白
        if (!this.isMobile) {
          this.messagesPadding = 0;
          return;
        }

        const containerRect = container.getBoundingClientRect();
        const inputRect = input.getBoundingClientRect();

        // 与输入区的重叠
        const overlapInput = Math.max(containerRect.bottom - inputRect.top, 0);

        // 与右下角“＋”按钮的重叠（如果存在）
        let overlapPlus = 0;
        let plusHeight = 0;
        const plus = this.$el && this.$el.querySelector('.plus-fab');
        if (plus) {
          const plusRect = plus.getBoundingClientRect();
          overlapPlus = Math.max(containerRect.bottom - plusRect.top, 0);
          plusHeight = plusRect.height || 0;
        }

        // 取更大的实际重叠量
        const overlap = Math.max(overlapInput, overlapPlus);

        // 小缓冲，避免视觉贴边
        const extraBuffer = 12;

        // 回退高度：在没有重叠但需要留白的情况下，取输入区/按钮中较大者
        const fallbackHeight = Math.max(inputRect.height || 0, plusHeight || 0);
        const fallback = fallbackHeight > 0 ? fallbackHeight + extraBuffer : 0;

        const padding = overlap > 0 ? (overlap + extraBuffer) : fallback;

        this.messagesPadding = Math.max(0, Math.round(padding));
      });
    },
    
    // 格式化时间
    formatTime(timestamp) {
      // 检查时间戳是否有效
      if (!timestamp) {
        console.warn('收到无效时间戳:', timestamp);
        return '未知时间';
      }
      
      try {
        const date = new Date(timestamp);
        
        // 检查日期是否有效
        if (isNaN(date.getTime())) {
          console.warn('无法解析时间戳:', timestamp);
          return '未知时间';
        }
        
        const now = new Date();
        const diff = now - date;
        
        if (diff < 60000) { // 1分钟内
          return '刚刚';
        } else if (diff < 3600000) { // 1小时内
          return `${Math.floor(diff / 60000)}分钟前`;
        } else if (diff < 86400000) { // 1天内
          return `${Math.floor(diff / 3600000)}小时前`;
        } else {
          // 使用更可靠的格式化方法
          return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
        }
      } catch (error) {
        console.error('格式化时间出错:', error, timestamp);
        return '未知时间';
      }
    },
    
    // 处理回车发送消息
    handleEnterSend(event) {
      // 如果按下了Shift键，不发送消息（允许换行）
      if (event.shiftKey) {
        return;
      }
      
      // 阻止默认行为（换行）
      event.preventDefault();
      event.stopPropagation();
      
      // 保存当前滚动位置
      const messageContainer = this.$refs.messageContainer;
      const scrollPosition = messageContainer ? messageContainer.scrollTop : 0;
      
      // 调用发送消息方法
      this.sendMessage();
      
      // 恢复滚动位置（下一个渲染循环）
      this.$nextTick(() => {
        if (messageContainer) {
          messageContainer.scrollTop = scrollPosition;
        }
      });
      
      // 阻止事件继续传播
      return false;
    }
  }
}
</script>

<style scoped>
/* 思考过程样式 */
.reasoning-content {
  margin-bottom: 10px;
  transition: all 0.3s ease;
}

.reasoning-header {
  display: flex;
  align-items: center;
  padding: 6px 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
  cursor: pointer;
  margin-bottom: 8px;
  font-size: 12px;
  color: #606266;
  user-select: none;
}

.reasoning-header:hover {
  background-color: #ebeef5;
}

.reasoning-header i {
  margin-right: 5px;
}

.reasoning-text {
  background-color: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  font-size: 14px;
  color: #606266;
  white-space: pre-wrap;
  word-break: break-word;
  margin-bottom: 10px;
  border-left: 3px solid #dcdfe6;
}

.reasoning-content.collapsed .reasoning-text {
  display: none;
}
.ai-chat-container {
  height: 100%;
  min-height: 100%;
  display: flex;
  flex-direction: column;
  background: transparent; /* 由 theme-background 提供背景 */
  overflow: hidden; /* 防止页面滑动 */
  padding: 12px 16px; /* 让顶部玻璃卡片有留白圆角 */
  box-sizing: border-box;
}

/* 强制覆盖所有可能的间距 */
.ai-chat-container > * {
  margin-top: 0 !important;
  margin-bottom: 0 !important;
}

/* 移除聊天区域的顶部间距 */
.chat-content > *:first-child {
  margin-top: 0 !important;
  padding-top: 0 !important;
}

/* 调整导航栏与主体间距，营造玻璃卡片层次 */
.chat-header + .chat-main {
  margin-top: 12px !important;
  border-top: 0 !important;
}

.chat-header {
  height: 60px;
  background: transparent;
  border-bottom: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  margin-bottom: 0 !important;
}

/* 顶部栏玻璃质感，与Home风格统一 */
.header-glass {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(200, 200, 200, 0.3);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
  border-radius: 12px;
}

.header-left, .header-right {
  display: flex;
  gap: 10px;
}

.header-center h2 {
  margin: 0;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.chat-main {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 0; /* 允许内部可滚动区域正确收缩 */
  margin-top: 0 !important;
  padding-top: 0 !important;
  border-top: none !important;
}

.chat-sidebar {
  width: 300px;
  background: transparent; /* 由 glass-card 提供背景 */
  border-right: none;
  display: flex;
  flex-direction: column;
  border-radius: 12px;
}

.sidebar-header {
  padding: 15px 20px;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-actions {
  display: flex;
  gap: 5px;
}

.sidebar-toggle-button {
  position: absolute;
  top: 2px;
  left: 0px;
  z-index: 10;
}

.sidebar-header h3 {
  margin: 0;
  color: #303133;
  font-size: 16px;
}

/* 不再需要单独的侧边栏切换按钮样式 */

/* 侧边栏过渡动画 */
.slide-fade-enter-active, .slide-fade-leave-active {
  transition: all 0.3s ease;
}

.slide-fade-enter, .slide-fade-leave-to {
  transform: translateX(-100%);
  opacity: 0;
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  margin-bottom: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.session-item:hover {
  background-color: #f5f7fa;
}

.session-item.active {
  background-color: #ecf5ff;
  border-left: 3px solid #409eff;
}

.session-info {
  flex: 1;
  min-width: 0;
}

.session-title {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-meta {
  font-size: 12px;
  color: #909399;
  display: flex;
  gap: 10px;
}

.session-actions {
  opacity: 0;
  transition: opacity 0.3s;
}

.session-item:hover .session-actions {
  opacity: 1;
}

/* 主体区域内边距，形成统一留白 */
.chat-main {
  padding: 12px 0 12px 0;
  gap: 12px;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0; /* 关键：使内部chat-messages可滚动而不撑破布局 */
  background: transparent; /* 避免与子级glass-card冲突 */
  margin-top: 0 !important;
  padding-top: 0 !important;
}

.model-selector {
  padding: 12px 16px !important;
  border-bottom: none;
  background-color: transparent;
  margin-top: 0 !important;
  margin-bottom: 12px !important;
  flex-shrink: 0; /* 防止被挤压，保持顶部固定高度 */
}

.selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 15px;
}

.model-list {
  display: flex;
  align-items: center;
}

.selector-controls {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

/* 流式返回样式 */
.message-item.assistant.streaming {
  opacity: 0.9;
}

.message-item.assistant.streaming .message-content {
  border-left: 4px solid #409eff;
  background-color: #f0f9ff;
  animation: pulse 2s infinite;
  position: relative;
  overflow: hidden;
}

@keyframes pulse {
  0% { background-color: #f0f9ff; }
  50% { background-color: #ecf5ff; }
  100% { background-color: #f0f9ff; }
}

/* 流式状态提示 */
.stream-status {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}

/* 打字机效果 */
.message-item.assistant .message-text {
  white-space: pre-wrap;
  word-break: break-word;
}

.model-list {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
}

.model-checkbox {
  margin-right: 0;
}

.model-info {
  margin-left: 8px;
}

.model-name {
  font-weight: 500;
  color: #303133;
}

.model-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background-color: transparent; /* 使用glass卡片背景 */
  margin-bottom: 12px;
  overscroll-behavior-y: contain;
  -webkit-overflow-scrolling: touch;
}

.message-item {
  display: flex;
  margin-bottom: 20px;
  gap: 12px;
}

.message-item.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #409eff;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 18px;
  flex-shrink: 0;
}

.message-item.user .message-avatar {
  background: #67c23a;
}

.message-content {
  /* 限制气泡两侧留出头像+间距的空间，避免跨越对侧头像区域 */
  /* 计算：2 * (头像40px + 间距12px) = 104px */
  max-width: calc(100% - 104px);
  background: white;
  border-radius: 12px;
  padding: 12px 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  color: #303133; /* 默认文本颜色为深灰色 */
}

.message-item.assistant .message-content {
  background: white;
  color: #303133; /* AI助手消息的文本颜色 */
}

.message-item.user .message-content {
  background: white;
  color: #303133;
  border: 1px solid #e6e6e6;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 12px;
}

.message-role {
  font-weight: 500;
}

.message-model {
  color: #909399;
}

.message-item.user .message-model {
  color: #909399;
}

.message-time {
  color: #c0c4cc;
}

.message-item.user .message-time {
  color: #c0c4cc;
}

.message-text {
  line-height: 1.5;
  word-break: break-word;
  color: #303133; /* 设置默认文本颜色为深灰色 */
}

/* 用户消息文本样式 */
.message-item.user .message-text {
  color: #303133;
}

/* 确保AI消息文本是深色 */
.message-item.assistant .message-text {
  color: #303133;
}

/* 修改消息图片尺寸和样式 */
.message-image {
  margin-top: 12px;
}

.message-image img {
  max-width: 250px; /* 限制最大宽度 */
  max-height: 200px; /* 限制最大高度 */
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 0.3s, max-width 0.3s, max-height 0.3s;
  object-fit: contain; /* 保持原始比例 */
}

.message-image img:hover {
  transform: scale(1.02);
}

/* 图片预览样式 */
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: rgba(0, 0, 0, 0.8);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}

.image-preview-img {
  max-width: 90vw;
  max-height: 90vh;
  border-radius: 10px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
  object-fit: contain;
}

/* 当图片处于预览模式时应用的样式 */
.message-image img.image-preview-mode {
  max-width: none;
  max-height: none;
  z-index: 10000;
}

/* 预览图片的动画效果 */
.message-image img.preview-active {
  max-width: 90vw;
  max-height: 90vh;
}

.message-video {
  margin-top: 12px;
  max-width: 100%;
}

.message-video-player {
  width: 100%;
  max-width: 480px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.pending-media {
  margin-top: 10px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.media-preview-thumbnail {
  width: 60px;
  height: 60px;
  object-fit: cover;
  border-radius: 4px;
}

.pending-media-info {
  flex: 1;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.loading-message {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #909399;
  font-style: italic;
}

.chat-input {
  padding: 16px;
  border-top: none;
  background: transparent; /* 由 glass-card 提供背景 */
  flex-shrink: 0; /* 固定在底部，不随消息区滚动 */
}

.input-toolbar {
  margin-bottom: 12px;
}

.input-area {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.input-area .el-textarea {
  flex: 1;
}

.model-details {
  max-height: 400px;
  overflow-y: auto;
}

.model-detail-item {
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.model-detail-item:last-child {
  border-bottom: none;
}

.model-detail-item h4 {
  margin: 0 0 8px 0;
  color: #303133;
}

.model-detail-item p {
  margin: 0 0 12px 0;
  color: #606266;
  line-height: 1.5;
}

.model-specs {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #909399;
}

.model-specs span {
  background: #f5f7fa;
  padding: 4px 8px;
  border-radius: 4px;
}

/* 与Home统一的区段标题样式（紧凑版） */
.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 8px 0;
  padding-left: 2px;
}
.section-title.compact {
  margin: 0 0 6px 0;
}
.section-title .section-accent {
  width: 28px;
  height: 3px;
  border-radius: 999px;
  background: linear-gradient(90deg, #2c3e50, #4a6572);
  box-shadow: 0 0 6px rgba(0,0,0,0.06);
}
.section-title h3 {
  font-size: 14px;
  font-weight: 700;
  margin: 0;
  color: #2c3e50;
  letter-spacing: 0.2px;
}

/* removed duplicate section-title styles */
/* 按钮统一风格（与Home保持一致的圆角与轻微悬浮效果） */
.chat-btn.el-button {
  border-radius: 10px;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}
.chat-btn.el-button:not(.el-button--text) {
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}
.chat-btn.el-button:not(.el-button--text):hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 14px rgba(0,0,0,0.12);
}
.chat-btn.el-button.is-disabled,
.theme-button--text.el-button.is-disabled {
  opacity: 0.6;
  box-shadow: none;
  transform: none;
  cursor: not-allowed;
}
/* 发送按钮强调 */
.chat-btn.chat-btn--send {
  font-weight: 600;
  min-width: 88px;
}
/* 文本按钮统一（与 theme.css 的 theme-button--text 呼应） */
.theme-button--text.el-button--text {
  border-radius: 8px;
  padding: 4px 6px;
  color: #444;
}
.theme-button--text.el-button--text:hover {
  color: #000;
  background-color: rgba(0, 0, 0, 0.04);
}
/* 标题栏样式参考自 TTSDetail.vue 的 page-header */
.tts-page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 标题文案样式（与TTSDetail保持一致的视觉权重与字重） */
.tts-page-header .header-center h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 500;
  letter-spacing: 0.3px;
}

/* 主按钮颜色统一为深蓝灰渐变（与TTSDetail的action-button主色一致） */
.chat-btn.el-button--primary {
  background: linear-gradient(90deg, #2c3e50, #4a6572);
  border: none;
  color: #ffffff;
  box-shadow: 0 5px 15px rgba(44, 62, 80, 0.2);
}
.chat-btn.el-button--primary:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(44, 62, 80, 0.3);
}

/* 文本按钮色系微调，贴合深蓝灰主题 */
.theme-button--text.el-button--text {
  color: #2c3e50;
}
.theme-button--text.el-button--text:hover {
  color: #1f2d3d;
  background-color: rgba(0, 0, 0, 0.04);
}
/* 移动端适配（不影响PC端） */
@media (max-width: 768px) {
  .ai-chat-container {
    padding: 0;
    height: 100dvh;
    min-height: 100svh;
    position: fixed;
    inset: 0;
  }

  /* 移动端顶部栏 */
  .mobile-topbar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 52px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 12px 0 56px; /* 预留左侧历史按钮点击区 */
    z-index: 2000;
    border-radius: 0;
    margin: 0;
  }
  .history-toggle {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    width: 44px;
    height: 36px;
    border: none;
    background: transparent;
    padding: 0;
    z-index: 2100; /* 确保在下拉框之上，避免被遮挡 */
  }
  .history-toggle .line {
    display: block;
    height: 3px;
    background: #2c3e50;
    border-radius: 999px;
    margin: 6px 0;
  }
  .history-toggle .line.long { width: 22px; }
  .history-toggle .line.short { width: 14px; }

  .topbar-center { width: 100%; max-width: 260px; }
  .topbar-center .el-select { width: 100%; }

  /* 为顶部栏预留空间 */
  .chat-content {
    padding-top: 60px !important;
    height: 100%;
    position: relative;
  }

  /* 全屏化聊天区域，移除多层背景 */
  .glass-card {
    background: transparent !important;
    box-shadow: none !important;
    border: none !important;
  }
  .chat-main {
    padding: 0 !important;
    gap: 0 !important;
    height: 100%;
  }
  .chat-messages {
    margin-bottom: 0 !important;
  }

  /* 顶部侧栏开关在移动端固定，避免被内容挤压（桌面端按钮已隐藏） */
  .sidebar-toggle-button {
    position: fixed;
    top: 10px;
    left: 10px;
    z-index: 2000;
  }

  /* 输入区右下角 + 按钮 */
  /* 将“＋”按钮内置到输入框内侧的最右边 */
  .input-area .plus-fab {
    position: absolute;
    right: 14px;
    bottom: 14px;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: linear-gradient(90deg, #2c3e50, #4a6572);
    color: #fff;
    font-size: 20px;
    line-height: 36px;
    text-align: center;
    border: none;
    box-shadow: 0 6px 16px rgba(0,0,0,0.2);
    z-index: 2; /* 置于文本域之上，但不遮挡弹层 */
  }
  .input-area .plus-fab:active { transform: scale(0.98); }
  .plus-menu { display: flex; flex-direction: column; gap: 6px; }

  /* 侧栏以浮层覆盖形式出现 */
  .chat-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    width: 66.6667vw;
    height: 100vh;
    overflow-y: auto;
    z-index: 1500;
    border-radius: 0;
    box-shadow: 0 6px 24px rgba(0, 0, 0, 0.2);
    background: rgba(18, 18, 18, 0.92);
  }

  .chat-main {
    padding: 0;
    gap: 0;
  }

  .chat-content {
    margin-left: 0 !important;
  }

  /* 模型选择区域紧凑化，纵向排布 */
  .model-selector {
    padding: 10px 12px !important;
    margin-bottom: 8px !important;
  }

  .selector-header {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }

  .model-list {
    width: 100%;
  }

  /* 覆盖 el-select 的内联宽度（300px） */
  .model-selector .el-select {
    width: 100% !important;
  }

  .selector-controls {
    justify-content: space-between;
    gap: 8px;
    flex-wrap: wrap;
  }

  .stream-status {
    margin-left: 0;
  }

  /* 消息区与消息卡片在小屏下更紧凑 */
  .chat-messages {
    padding: 0;
    margin-bottom: 0;
  }

  .message-avatar {
    width: 32px;
    height: 32px;
    font-size: 14px;
  }

  .message-content {
    /* 移动端：头像32px，间距10px，需为两侧同时预留 => 2 * (32 + 10) = 84px */
    max-width: calc(100% - 84px);
    padding: 10px 12px;
  }

  .message-item {
    gap: 10px;
    margin-bottom: 16px;
  }

  /* 图片、视频在小屏下自适应 */
  .message-image img {
    max-width: 80vw;
    max-height: 60vh;
  }

  .message-video-player {
    max-width: 100%;
  }

  /* 输入区改为上下堆叠，发送键全宽 */
  .input-area {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    position: relative;
  }
  .input-area .el-textarea { width: 100%; }
  .input-area .el-textarea__inner {
    border-radius: 16px;
    background: #ffffff;
    border: 1px solid #e5e7eb;
    padding: 10px 58px 10px 12px; /* 右侧为“＋”按钮留出空间 */
    box-shadow: 0 2px 8px rgba(0,0,0,0.05);
    /* 防止 iOS 聚焦时页面放大（<16px 会触发自动缩放） */
    font-size: 16px;
    line-height: 1.45;
  }

  .chat-btn.chat-btn--send {
    width: 100%;
  }

  .input-toolbar {
    margin-bottom: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .input-toolbar.has-media {
    margin-bottom: 10px;
  }

  .pending-media {
    align-items: flex-start;
  }

  /* Element 对话框移动端宽度适配（scoped 深度选择器） */
  ::v-deep .el-dialog {
    width: 90vw !important;
    max-width: 90vw !important;
    margin-top: 10vh !important;
  }
  ::v-deep .el-dialog__body {
    max-height: 60vh;
    overflow-y: auto;
  }

  /* 覆盖与增强（参考移动端布局需求） */
  .chat-sidebar {
    width: 66.6667vw !important;
    max-width: 66.6667vw !important;
  }
  /* 新建会话入口样式（移动端抽屉） */
  .chat-sidebar .session-item.new-session {
    border: 1px dashed rgba(255,255,255,0.18);
    background: rgba(255,255,255,0.06);
  }
  .chat-sidebar .session-item.new-session .session-title { color: #f9fafb; }
  .chat-sidebar .session-item.new-session .session-meta { color: #cbd5e1; }
  .chat-sidebar .session-item.new-session:hover { background: rgba(255,255,255,0.12); }
  /* 移动端历史侧栏深色主题以提升对比度 */
  .chat-sidebar { color: #f2f3f5; }
  .chat-sidebar .sidebar-header { border-bottom: 1px solid rgba(255,255,255,0.12); }
  .chat-sidebar .sidebar-header h3 { color: #f9fafb; }
  .chat-sidebar .session-title { color: #f3f4f6; }
  .chat-sidebar .session-meta { color: #cbd5e1; }
  .chat-sidebar .session-item:hover { background: rgba(255,255,255,0.06); }
  .chat-sidebar .session-item.active { background: rgba(255,255,255,0.12); border-left: 3px solid #60a5fa; }
  .chat-sidebar .session-actions .el-button { color: #e5e7eb; }

  .mobile-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.45);
    z-index: 1400;
  }

  /* 隐藏顶部模型卡片，改由齿轮打开抽屉 */
  .model-selector {
    display: none !important;
  }

  /* 齿轮按钮固定在右上角，仅移动端显示（模板已用 v-if 控制） */
  .settings-toggle-button {
    position: fixed;
    top: 10px;
    right: 10px;
    z-index: 2000;
  }

  /* 输入区固定底部并适配安全区 */
  .chat-input {
    position: fixed;
    left: 0;
    right: 0;
    bottom: 0;
    width: 100vw;
    padding: 8px 0;
    padding-bottom: calc(16px + env(safe-area-inset-bottom));
    padding-bottom: calc(16px + constant(safe-area-inset-bottom));
    border-radius: 0;
    z-index: 1200;
    background: transparent;
    will-change: transform;
  }

  /* 为固定输入区预留空间，避免消息被遮挡（由JS动态计算） */
  .chat-messages {
    padding-bottom: 0;
    overscroll-behavior-y: contain;
    -webkit-overflow-scrolling: touch;
  }

  /* 气泡样式：圆角（最大宽度在上方规则已计算两侧头像与间距） */
  .message-content {
    border-radius: 14px !important;
  }

  /* 移动端模型下拉宽度与换行优化：宽度等同组件，字体大小继承组件 */
  /* 仅限“未 append-to-body”的下拉：限制在组件容器内，宽度与组件一致 */
  .ai-chat-container ::v-deep .model-select-popper.el-select-dropdown {
    width: 100% !important;
    min-width: 100% !important;
    max-width: 100% !important;
    box-sizing: border-box;
    z-index: 2500 !important; /* 高于顶部栏与历史按钮 */
  }
  /* body 挂载的下拉层同样提升层级，避免被抽屉/遮罩覆盖 */
  ::v-deep .model-select-popper.el-select-dropdown {
    z-index: 2500 !important;
  }
  ::v-deep .model-select-popper .el-select-dropdown__wrap,
  ::v-deep .model-select-popper .el-scrollbar__wrap {
    max-height: 56vh !important;
    overflow-y: auto !important;
    -webkit-overflow-scrolling: touch;
    overscroll-behavior: contain;
    touch-action: pan-y;
  }
  ::v-deep .model-select-popper .el-select-dropdown__item {
    white-space: normal !important;
    line-height: 1.4;
  }
  /* 不覆盖字体大小，保持与组件一致，仅做描述两行省略 */
  ::v-deep .model-select-popper .model-desc {
    display: -webkit-box;
    -webkit-line-clamp: 2;

    -webkit-box-orient: vertical;
    overflow: hidden;
    word-break: break-word;
    margin-top: 4px;
    color: inherit;
  }
}
</style>