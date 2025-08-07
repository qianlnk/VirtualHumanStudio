<template>
  <div class="ai-chat-container">
    <!-- 顶部导航栏 -->
    <div class="chat-header">
      <div class="header-left">
        <!-- 侧边栏切换按钮放在顶部导航栏 -->
        <el-button
          :icon="sidebarVisible ? 'el-icon-arrow-left' : 'el-icon-notebook-2'"
          size="small"
          @click="toggleSidebar"
          :type="sidebarVisible ? 'info' : 'primary'"
        >
          {{ sidebarVisible ? '隐藏历史' : '显示历史' }}
        </el-button>
        <el-button
          type="primary"
          icon="el-icon-plus"
          size="small"
          @click="createNewSession"
        >
          新会话
        </el-button>
        <el-button
          icon="el-icon-time"
          size="small"
          @click="refreshSessions"
        >
          刷新
        </el-button>
      </div>
      
      <div class="header-center">
        <h2>AI绘画聊天</h2>
      </div>
      
      <div class="header-right">
        <el-button
          icon="el-icon-share"
          size="small"
          @click="shareSession"
        >
          分享
        </el-button>
        <el-button
          icon="el-icon-more"
          size="small"
          @click="showMoreOptions"
        >
          更多
        </el-button>
      </div>
    </div>

    <div class="chat-main">
      <!-- 左侧会话列表 - 根据sidebarVisible控制显示/隐藏 -->
      <transition name="slide-fade">
        <div class="chat-sidebar" v-if="sidebarVisible">
          <div class="sidebar-header">
            <h3>会话历史</h3>
            <div class="sidebar-actions">
              <el-button
                icon="el-icon-plus"
                size="mini"
                type="text"
                @click="createNewSession"
                title="新会话"
              />
              <el-button
                icon="el-icon-time"
                size="mini"
                type="text"
                @click="refreshSessions"
                title="刷新"
              />
              <el-button
                icon="el-icon-arrow-left"
                size="mini"
                type="text"
                @click="toggleSidebar"
                title="隐藏会话列表"
              />
            </div>
          </div>
          
          <div class="session-list">
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
                  @click.stop="deleteSession(session.id)"
                />
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- 右侧聊天区域 -->
      <div class="chat-content">
        <!-- 不再需要这里的侧边栏切换按钮，已移至顶部导航栏 -->
        <!-- 模型选择区域 - 即使在新会话状态下也显示 -->
        <div class="model-selector">
          <div class="selector-header">
            <div class="model-list">
              <el-select
                v-model="selectedModel"
                placeholder="请选择AI模型"
                style="width: 300px;"
                filterable
                default-first-option>
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
                @click="showModelInfo"
              >
                查看模型详情
              </el-button>
            </div>
          </div>
        </div>

        <!-- 聊天消息区域 -->
        <div class="chat-messages" ref="messageContainer">
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
                  :class="{'preview-active': currentPreviewUrl === message.image_url}"
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
        <div class="chat-input">
                  <div class="input-toolbar">
                    <el-button
                      icon="el-icon-picture"
                      size="small"
                      @click="uploadImage"
                    >
                      上传图片
                    </el-button>
                    <el-button
                      icon="el-icon-video-camera"
                      size="small"
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
            <el-button
              :type="loading ? 'danger' : 'primary'"
              :icon="loading ? 'el-icon-close' : 'el-icon-s-promotion'"
              :loading="false"
              :disabled="(!inputMessage.trim() || !selectedModel) && !loading"
              @click="loading ? cancelRequest() : sendMessage()"
            >
              {{ loading ? '中断' : '发送' }}
            </el-button>
          </div>
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
        <el-button @click="newSessionVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmCreateSession">创建</el-button>
      </div>
    </el-dialog>
    
    <!-- 简化的图片预览遮罩 -->
    <div v-if="imagePreviewVisible" class="image-preview-overlay" @click="closeImagePreview">
      <!-- 不创建新的图片元素，而是使用CSS样式控制预览效果 -->
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
      newSessionForm: {
        title: ''
      }
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
    window.addEventListener('resize', this.handleWindowResize);
  },
  
  beforeDestroy() {
    // 移除事件监听
    document.removeEventListener('keydown', this.handleKeyDown);
    window.removeEventListener('resize', this.handleWindowResize);
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
      // 如果有预览激活，重新调整预览图片的位置
      if (this.imagePreviewVisible && this.activeImageRef) {
        // 可以在这里添加逻辑来重新定位预览图片
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
      
      // 记录当前预览的URL
      this.currentPreviewUrl = url;
      this.imagePreviewVisible = true;
      
      // 查找对应的消息，从而定位到正确的图片元素
      const messageWithImage = this.messages.find(msg => msg.image_url === url);
      if (messageWithImage) {
        // 获取对应的图片引用
        const imgRef = this.$refs[`img-${messageWithImage.id}`];
        if (imgRef && imgRef[0]) {
          // 设置当前激活的图片元素
          this.activeImageRef = imgRef[0];
          
          // 添加预览样式类
          imgRef[0].classList.add('image-preview-mode');
          
          // 调整图片元素的样式，使其在预览时居中显示
          this.$nextTick(() => {
            imgRef[0].style.position = 'fixed';
            imgRef[0].style.top = '50%';
            imgRef[0].style.left = '50%';
            imgRef[0].style.transform = 'translate(-50%, -50%)';
            imgRef[0].style.maxWidth = '90vw';
            imgRef[0].style.maxHeight = '90vh';
            imgRef[0].style.zIndex = '10000';
            imgRef[0].style.objectFit = 'contain';
            
            // 增加过渡动画
            imgRef[0].style.transition = 'all 0.3s ease';
          });
        }
      }
      
      // 禁止页面滚动
      document.body.style.overflow = 'hidden';
    },
    
    // 关闭图片预览
    closeImagePreview() {
      if (this.activeImageRef) {
        // 恢复原始样式
        this.activeImageRef.classList.remove('image-preview-mode');
        this.activeImageRef.style.position = '';
        this.activeImageRef.style.top = '';
        this.activeImageRef.style.left = '';
        this.activeImageRef.style.transform = '';
        this.activeImageRef.style.maxWidth = '';
        this.activeImageRef.style.maxHeight = '';
        this.activeImageRef.style.zIndex = '';
        this.activeImageRef.style.objectFit = '';
        this.activeImageRef.style.transition = '';
      }
      
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
        const title = this.inputMessage.length > 50
          ? this.inputMessage.substring(0, 50) + '...'
          : this.inputMessage;
        
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
    
    // 创建新会话
    createNewSession() {
      // 判断当前是否已经是空会话
      if (!this.currentSession) {
        // 已经是空会话，不需要再创建
        console.log('当前已经是空会话，无需创建');
        return;
      }
      
      // 不是空会话，清空当前会话状态
      this.currentSession = null;
      this.messages = [];
      console.log('已清空当前会话，等待用户输入');
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
      const container = this.$refs.messageContainer
      if (container) {
        container.scrollTo({
          top: container.scrollHeight,
          behavior: smooth ? 'smooth' : 'auto'
        });
      }
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
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f5f5;
  overflow: hidden; /* 防止页面滑动 */
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

/* 调整导航栏底部无间隙 */
.chat-header + .chat-main {
  margin-top: 0 !important;
  border-top: 0 !important;
}

.chat-header {
  height: 60px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 0 !important;
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
  margin-top: 0 !important;
  padding-top: 0 !important;
  border-top: none !important;
}

.chat-sidebar {
  width: 300px;
  background: white;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
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
  top: 10px;
  left: 10px;
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

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
  margin-top: 0 !important;
  padding-top: 0 !important;
}

.model-selector {
  padding: 5px 20px !important;
  border-bottom: 1px solid #e4e7ed;
  background-color: #fafafa;
  margin-top: 0 !important;
  margin-bottom: 0 !important;
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
  padding: 20px;
  background-color: #fafafa;
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
  max-width: 70%;
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
  padding: 20px;
  border-top: 1px solid #e4e7ed;
  background: white;
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
</style>