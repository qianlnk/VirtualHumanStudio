<template>
  <div class="workflow-detail-container">
    <div class="image-preview-overlay" v-if="isPreviewActive" @click="closePreview" @wheel.prevent="handleZoom" @mousedown="startDrag" @mousemove="onDrag" @mouseup="stopDrag" @mouseleave="stopDrag">
      <img :src="previewUrl" :style="imageStyle" class="preview-image" ref="previewImage">
    </div>
    
    <!-- PC端顶部导航栏 -->
    <div class="page-header" v-show="!isMobile">
      <h2>{{ currentModule ? currentModule.name : '图像处理' }}任务详情</h2>
      <div>
        <el-button @click="goBack" icon="el-icon-back">返回列表</el-button>
        <el-button type="primary" @click="retryTask" :loading="retrying" v-if="task && task.status !== 'processing'">重试任务</el-button>
        <el-button type="primary" @click="refreshTask" icon="el-icon-refresh">刷新</el-button>
      </div>
    </div>
    
    <!-- 移动端顶部导航栏 - 只在移动端显示 -->
    <div class="mobile-header-bar" v-show="isMobile">
      <div class="header-back" @click="goBack">
        <i class="el-icon-arrow-left"></i>
        <span>返回</span>
      </div>
      <h2 class="header-title">{{ currentModule ? currentModule.name : '图像处理' }}任务</h2>
    </div>
    
    <!-- 移动端头部占位 - 只在移动端显示 -->
    <div class="mobile-header-placeholder" v-show="isMobile"></div>

    <div class="detail-content-wrapper">
      <!-- PC端展示 -->
      <el-card v-loading="loading" v-show="!isMobile">
        <div v-if="task" class="task-detail">
          <div class="task-header">
            <div class="task-info">
              <h3>任务信息</h3>
              <div class="info-item">
                <span class="label">任务名称:</span>
                <span>{{ task.name }}</span>
              </div>
              <div class="info-item">
                <span class="label">创建时间:</span>
                <span>{{ formatDate(task.created_at) }}</span>
              </div>
              <div class="info-item">
                <span class="label">状态:</span>
                <el-tag :type="getStatusType(task.status)">{{ getStatusText(task.status) }}</el-tag>
              </div>
            </div>
          </div>

          <!-- 输入参数 -->
          <div class="task-section">
            <h3>输入参数</h3>
            <div class="params-container">
              <template v-if="task && task.input_params">
                <div v-for="(param, index) in parseInputParams()" :key="'input-'+index" class="param-item">
                  <div class="param-header">
                    <span class="param-name">{{ param.alias || param.key }}:</span>
                  </div>
                  <div class="param-content">
                    <!-- 图片类型参数 -->
                    <div v-if="param.type === 'image' || param.type === 'mask'" class="image-param">
                      <img :src="param.value" :alt="param.alias || param.key" class="param-image" @click="previewImage(param.value)">
                    </div>
                    <!-- 其他类型参数 -->
                    <div v-else class="text-param">
                      {{ param.value }}
                    </div>
                  </div>
                </div>
              </template>
              <div v-else class="no-params">
                <el-empty description="无输入参数" :image-size="100"></el-empty>
              </div>
            </div>
          </div>

          <!-- 处理结果 -->
          <div class="task-section" v-if="task.status === 'completed'">
            <h3>处理结果</h3>
            <div class="result-container">
              <template v-if="task && task.output_params">
                <div v-for="(param, index) in parseOutputParams()" :key="'output-'+index" class="result-item">
                  <div class="result-header">
                    <span class="result-name">{{ param.alias || param.key }}:</span>
                  </div>
                  <div class="result-content">
                    <!-- 图片类型结果 -->
                    <template v-if="param.type === 'image' && param.value">
                      <img :src="param.value" :alt="param.alias || param.key" class="result-image" @click="previewImage(param.value)">
                      <div class="result-actions">
                        <el-button type="primary" size="small" @click="downloadImage(param.value, param.key)">
                          下载
                        </el-button>
                        <el-button 
                          v-if="index === 0" 
                          type="success" 
                          size="small" 
                          @click="shareTask" 
                          :disabled="isShared">
                          {{ getShareButtonText() }}
                        </el-button>
                      </div>
                    </template>
                    <!-- 视频类型结果 -->
                    <template v-else-if="param.type === 'video' && param.value">
                      <div class="video-preview">
                        <video
                          controls
                          style="width: 100%; max-height: 400px"
                          ref="videoPlayer"
                        >
                          <source :src="param.value" type="video/mp4">
                          您的浏览器不支持视频播放
                        </video>
                      </div>
                      <div class="result-actions">
                        <el-button type="primary" size="small" @click="downloadVideo(param.value, param.key)">
                          下载
                        </el-button>
                        <el-button 
                          v-if="index === 0" 
                          type="success" 
                          size="small" 
                          @click="shareTask" 
                          :disabled="isShared">
                          {{ getShareButtonText() }}
                        </el-button>
                      </div>
                    </template>
                    <!-- 其他类型结果 -->
                    <div v-else class="text-result">
                      {{ param.value }}
                    </div>
                  </div>
                </div>
              </template>
              <div v-else class="no-results">
                <el-empty description="无处理结果" :image-size="100"></el-empty>
              </div>
            </div>
          </div>

          <!-- 错误信息 -->
          <div class="task-section" v-if="task.status === 'failed' && task.error_msg">
            <h3>错误信息</h3>
            <div class="error-container">
              <pre class="error-message">{{ task.error_msg }}</pre>
            </div>
          </div>
        </div>

        <div v-else-if="!loading" class="no-task">
          <el-empty description="任务不存在或已被删除"></el-empty>
          <div style="color: #999; font-size: 12px; margin-top: 10px; text-align: center;">
            调试信息: task={{ task ? '存在' : '不存在' }}, loading={{ loading }}
          </div>
        </div>
      </el-card>
      
      <!-- 移动端展示 -->
      <div class="mobile-content-view" v-show="isMobile" v-loading="loading">
        <div v-if="task" class="mobile-content-inner">
          <!-- 基本信息 - 简化版 -->
          <div class="basic-info">
            <div class="status-tag">
              <el-tag :type="getStatusType(task.status)">{{ getStatusText(task.status) }}</el-tag>
            </div>
            <div class="create-time">创建时间：{{ formatDate(task.created_at) }}</div>
          </div>
          
          <!-- 任务信息 -->
          <div class="mobile-task-info">
            <div class="info-item">
              <span class="label">任务名称：</span>
              <span>{{ task.name }}</span>
            </div>
          </div>
          
          <!-- 输入参数 -->
          <div class="mobile-section">
            <h4 class="section-title">输入参数</h4>
            <div class="mobile-params">
              <template v-if="task && task.input_params">
                <div v-for="(param, index) in parseInputParams()" :key="'mobile-input-'+index" class="mobile-param-item">
                  <div class="mobile-param-header">
                    <span class="mobile-param-name">{{ param.alias || param.key }}:</span>
                  </div>
                  <!-- 图片类型参数 -->
                  <div v-if="param.type === 'image' || param.type === 'mask'" class="mobile-image-param">
                    <img :src="param.value" :alt="param.alias || param.key" class="mobile-param-image" @click="previewImage(param.value)">
                  </div>
                  <!-- 其他类型参数 -->
                  <div v-else class="mobile-text-param">
                    {{ param.value }}
                  </div>
                </div>
              </template>
              <div v-else class="mobile-no-params">
                <el-empty description="无输入参数" :image-size="60"></el-empty>
              </div>
            </div>
          </div>
          
          <!-- 处理结果 -->
          <div class="mobile-section" v-if="task.status === 'completed'">
            <h4 class="section-title">处理结果</h4>
            <div class="mobile-results">
              <template v-if="task && task.output_params">
                <div v-for="(param, index) in parseOutputParams()" :key="'mobile-output-'+index" class="mobile-result-item">
                  <div class="mobile-result-header">
                    <span class="mobile-result-name">{{ param.alias || param.key }}:</span>
                  </div>
                  <!-- 图片类型结果 -->
                  <template v-if="param.type === 'image' && param.value">
                    <div class="mobile-image-result">
                      <img :src="param.value" :alt="param.alias || param.key" class="mobile-result-image" @click="previewImage(param.value)">
                    </div>
                      <div class="mobile-result-actions">
                      <el-button type="primary" size="mini" @click="downloadImage(param.value, param.key)">下载</el-button>
                      <el-button 
                        v-if="index === 0" 
                        type="success" 
                        size="mini" 
                        @click="shareTask" 
                        :disabled="isShared">
                        {{ getShareButtonText() }}
                        </el-button>
                    </div>
                  </template>
                  <!-- 视频类型结果 -->
                  <template v-else-if="param.type === 'video' && param.value">
                    <div class="mobile-video-preview">
                      <video
                        controls
                        style="width: 100%; max-height: 200px"
                        class="mobile-video"
                      >
                        <source :src="param.value" type="video/mp4">
                        您的浏览器不支持视频播放
                      </video>
                      <div class="mobile-result-actions">
                        <el-button type="primary" size="mini" @click="downloadVideo(param.value, param.key)">
                          下载
                        </el-button>
                        <el-button 
                          v-if="index === 0" 
                          type="success" 
                          size="mini" 
                          @click="shareTask" 
                          :disabled="isShared">
                          {{ getShareButtonText() }}
                        </el-button>
                      </div>
                    </div>
                  </template>
                  <!-- 其他类型结果 -->
                  <div v-else class="mobile-text-result">
                    {{ param.value }}
                  </div>
                </div>
              </template>
              <div v-else class="mobile-no-results">
                <el-empty description="无处理结果" :image-size="60"></el-empty>
              </div>
            </div>
          </div>
          
          <!-- 移动端错误信息 -->
          <div class="mobile-section" v-if="task.status === 'failed' && task.error_msg">
            <h4 class="section-title">错误信息</h4>
            <div class="mobile-error-container">
              <pre class="mobile-error-message">{{ task.error_msg }}</pre>
            </div>
          </div>
          
          <!-- 移动端操作按钮 -->
          <div class="mobile-actions">
            <el-button type="primary" @click="refreshTask" icon="el-icon-refresh" size="small">刷新</el-button>
            <el-button type="primary" @click="retryTask" :loading="retrying" size="small" v-if="task && task.status !== 'processing'">重试任务</el-button>
          </div>
        </div>
        
        <div v-else-if="!loading" class="mobile-no-task">
          <el-empty description="任务不存在或已被删除"></el-empty>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getImageProcessingModules, getImageProcessingTaskDetail, retryImageProcessingTask } from '@/utils/imageProcessingApi'
import { downloadFile, getImageUrl } from '@/utils/fileAccess'
import { shareTask } from '@/api/share'

export default {
  name: 'ImageProcessingTaskDetail',
  data() {
    return {
      loading: false,
      task: null,
      currentModule: null,
      refreshInterval: null,
      outputFilePaths: {},
      retrying: false,
      // 图片预览相关
      isPreviewActive: false,
      previewUrl: '',
      zoomLevel: 1,
      isDragging: false,
      dragStartX: 0,
      dragStartY: 0,
      translateX: 0,
      translateY: 0,
      // 移动端适配
      screenWidth: window.innerWidth,
      isMobile: false,
      globalStyleEl: null,
      // 分享相关状态
      sharing: false
    }
  },
  computed: {
    imageStyle() {
      return {
        transform: `scale(${this.zoomLevel}) translate(${this.translateX}px, ${this.translateY}px)`,
        transition: this.isDragging ? 'none' : 'transform 0.3s'
      }
    },
    // 是否已分享
    isShared() {
      return this.task && (this.task.is_shared || this.sharing)
    }
  },
  
  created() {
    this.initModule()
    // 初始化检测设备类型
    this.checkDeviceType()
    // 添加窗口大小变化监听
    window.addEventListener('resize', this.handleResize)
  },
  mounted() {
    // 不需要在mounted中调用initModule，因为已经在created中调用过了
    
    // 修复全局样式，确保菜单在合适的层级
    this.fixGlobalStyles()
  },
  beforeDestroy() {
    this.clearRefreshInterval()
    // 移除窗口大小变化监听
    window.removeEventListener('resize', this.handleResize)
    // 恢复全局样式
    this.restoreGlobalStyles()
  },
  methods: {
    // 初始化模块
    
    async initModule() {
      try {
        console.log('初始化模块，路由参数:', this.$route.params)
        const response = await getImageProcessingModules()
        if (response) {
          // 从路由路径中提取模块ID
          const pathParts = this.$route.path.split('/')
          const moduleId = pathParts[pathParts.length - 3]
          console.log('模块ID:', moduleId)
          this.currentModule = response.modules.find(m => m.id === moduleId)
          if (this.currentModule) {
            console.log('当前模块:', this.currentModule)
            this.fetchTask()
          } else {
            console.warn('未找到对应的模块:', moduleId)
          }
        }
      } catch (error) {
        this.$message.error('初始化模块失败：' + error.message)
      }
    },

    // 重试任务
    async retryTask() {
      if (!this.task || !this.currentModule) return

      this.retrying = true
      try {
        const response = await retryImageProcessingTask(this.currentModule.id, this.task.id)
        if (response && response.success) {
          this.$message.success('任务重试成功')
          // 刷新当前任务状态
          await this.refreshTask()
        } else {
          throw new Error('重试任务失败')
        }
      } catch (error) {
        console.error('重试任务失败:', error)
        this.$message.error('重试任务失败：' + error.message)
      } finally {
        this.retrying = false
      }
    },

    // 获取任务详情
    // 加载图片URL
    async loadImageUrls() {
      if (!this.task) return

      // 缓存已加载的URL，避免重复请求
      const urlCache = {}

      const updateParam = async (param) => {
        if (!param.value) return param
        
        // 检查URL是否已经是blob URL，如果是则不需要再次请求
        if (param.value.startsWith('blob:')) {
          return param
        }

        // 检查是否已经缓存过该URL
        const cacheKey = param.value
        if (urlCache[cacheKey]) {
          param.value = urlCache[cacheKey]
          return param
        }

        try {
          const fullUrl = param.value
          param.value = await getImageUrl(fullUrl)
          // 缓存成功加载的URL
          urlCache[cacheKey] = param.value
          
          // 根据参数类型更新DOM
          this.$nextTick(() => {
            if (param.type === 'image' || param.type === 'mask') {
              const imgElement = this.$el.querySelector(`img[alt="${param.alias || param.key}"]`)
              if (imgElement && imgElement.src !== param.value) {
                imgElement.src = param.value
              }
            } else if (param.type === 'video') {
              const videoElement = this.$el.querySelector(`video source[type="video/mp4"]`)
              if (videoElement && videoElement.src !== param.value) {
                videoElement.src = param.value
                // 重新加载视频
                videoElement.parentElement.load()
              }
            }
          })
        } catch (error) {
          // 401错误不影响用户体验，只记录日志不显示错误提示
          if (error.message && error.message.includes('401')) {
            console.warn(`文件认证失败 (${param.key}): 可能需要登录或刷新token`)
          } else {
            console.warn(`加载文件失败 (${param.key}):`, error)
          }
        }
        return param
      }

      try {
        // 处理输入参数中的图片和视频
        const inputParams = this.parseInputParams()
        for (const param of inputParams) {
          if (param.type === 'image' || param.type === 'mask' || param.type === 'video') {
            await updateParam(param)
          }
        }

        // 处理输出参数中的图片和视频，只有当任务完成且有输出参数时才处理
        if (this.task.status === 'completed' && this.task.output_params) {
          const outputParams = this.parseOutputParams()
          for (const param of outputParams) {
            if (param.type === 'image' || param.type === 'video') {
              await updateParam(param)
            }
          }
        }
      } catch (error) {
        console.error('加载图片URL失败:', error)
        // 只有非401错误才显示错误提示
        if (!error.message || !error.message.includes('401')) {
          this.$message.error('加载图片文件失败: ' + error.message)
        }
      }
    },

    // 下载图片
    async downloadImage(url, filename) {
      try {
        await downloadFile(url, `${filename || 'result'}.png`)
        this.$message.success('下载成功')
      } catch (error) {
        console.error('下载文件失败:', error)
        this.$message.error('下载文件失败')
      }
    },

    // 下载视频
    async downloadVideo(url, filename) {
      try {
        await downloadFile(url, `${filename || 'result'}.mp4`)
        this.$message.success('下载成功')
      } catch (error) {
        console.error('下载视频失败:', error)
        this.$message.error('下载视频失败')
      }
    },

    // 预览图片
    async previewImage(url) {
      if (!url) return
      try {
        const fullUrl = url
        this.previewUrl = await getImageUrl(fullUrl)
        this.isPreviewActive = true
        this.resetZoom()
      } catch (error) {
        console.error('预览图片失败:', error)
        this.$message.error('预览图片失败：' + error.message)
      }
    },

    // 关闭预览
    closePreview() {
      if (this.previewUrl) {
        URL.revokeObjectURL(this.previewUrl)
      }
      this.previewUrl = ''
      this.isPreviewActive = false
      this.resetZoom()
    },

    // 重置缩放和位置
    resetZoom() {
      this.zoomLevel = 1
      this.translateX = 0
      this.translateY = 0
    },

    // 放大图片
    zoomIn() {
      if (this.zoomLevel < 3) {
        this.zoomLevel += 0.2
      }
    },

    // 缩小图片
    zoomOut() {
      if (this.zoomLevel > 0.5) {
        this.zoomLevel -= 0.2
      }
    },

    // 处理鼠标滚轮缩放
    handleZoom(event) {
      if (event.deltaY < 0) {
        this.zoomIn()
      } else {
        this.zoomOut()
      }
    },

    // 开始拖动
    startDrag(event) {
      if (this.zoomLevel > 1) {
        this.isDragging = true
        this.dragStartX = event.clientX - this.translateX
        this.dragStartY = event.clientY - this.translateY
        event.preventDefault()
      }
    },

    // 拖动中
    onDrag(event) {
      if (this.isDragging) {
        const newTranslateX = event.clientX - this.dragStartX
        const newTranslateY = event.clientY - this.dragStartY
        
        this.translateX = newTranslateX
        this.translateY = newTranslateY
        event.preventDefault()
      }
    },

    // 停止拖动
    stopDrag() {
      this.isDragging = false
    },

    async fetchTask() {
      if (!this.currentModule) {
        console.warn('当前模块未初始化，无法获取任务详情')
        return
      }
      
      console.log('开始获取任务详情:', this.currentModule.id, this.$route.params.id)
      this.loading = true
      try {
        const taskId = this.$route.params.id
        const response = await getImageProcessingTaskDetail(this.currentModule.id, taskId)
        
        if (!response) {
          throw new Error('获取任务详情响应为空')
        }

        if (response.success && response.task) {
          // 获取当前任务状态，用于判断是否需要重新设置定时器
          const currentStatus = this.task ? this.task.status : null
          const newStatus = response.task.status || 'unknown'
          
          // 确保task对象包含必要的属性
          const taskData = {
            id: response.task.id || taskId,
            name: response.task.name || '未知任务',
            status: newStatus,
            created_at: response.task.created_at,
            input_params: response.task.input_params || '[]',
            output_params: response.task.output_params || '[]',
            error_msg: response.task.error_msg || '',
            task_type: response.task.task_type || 'unknown'
          }
          
          // 使用Vue.set确保响应式更新
          this.task = taskData
          this.parseOutputFilePaths()
          await this.loadImageUrls()
          
          // 只有当状态发生变化时才处理自动刷新
          if (currentStatus !== newStatus) {
            console.log('任务状态发生变化:', currentStatus, '->', newStatus)
            // 根据新状态决定是否需要自动刷新
            if (['processing', 'pending'].includes(newStatus)) {
              console.log('任务处于进行中状态，开启自动刷新')
              this.setupAutoRefresh()
            } else {
              console.log('任务已完成或失败，关闭自动刷新')
              this.clearRefreshInterval()
            }
          } else {
            console.log('任务状态未变化:', newStatus)
          }
        } else {
          // 如果没有任务数据，将task设置为null
          console.warn('获取任务详情失败:', response.success ? '任务数据不存在' : '请求失败')
          this.task = null
          this.$message.warning(response.success ? '任务数据不存在' : '获取任务详情失败')
        }
      } catch (error) {
        console.error('获取任务详情失败:', error)
        this.$message.error('获取任务详情失败：' + error.message)
        this.task = null
      } finally {
        this.loading = false
      }
    },

    // 解析输出文件路径
    parseOutputFilePaths() {
      if (!this.task) return
      
      try {
        // 尝试从output_file_paths解析
        if (this.task.output_file_paths) {
          this.outputFilePaths = JSON.parse(this.task.output_file_paths)
        } 
        // 如果output_file_paths为空或解析失败，尝试从output_params解析
        else if (this.task.output_params) {
          const outputParams = JSON.parse(this.task.output_params)
          if (Array.isArray(outputParams)) {
            outputParams.forEach(param => {
              if (param.type === 'image' && param.value) {
                this.outputFilePaths[param.key] = param.value
              }
            })
          }
        }
      } catch (error) {
        console.error('解析输出文件路径失败：', error)
        this.outputFilePaths = {}
      }
    },

    // 获取参数值
    getParamValue(paramsStr, key) {
      if (!paramsStr) return ''
      
      try {
        const params = JSON.parse(paramsStr)
        // 检查params是否为数组（新格式）
        if (Array.isArray(params)) {
          const param = params.find(p => p.key === key)
          return param ? param.value : ''
        } else {
          // 兼容旧格式（直接的键值对对象）
          return params[key] || ''
        }
      } catch (error) {
        console.error('解析参数失败:', error, paramsStr)
        return ''
      }
    },

    // 返回列表
    goBack() {
      this.$router.push(`/image-processing/${this.currentModule ? this.currentModule.id : ''}`)
    },

    // 刷新任务
    refreshTask() {
      this.fetchTask()
    },

    // 设置自动刷新
    setupAutoRefresh() {
      // 先清除已存在的定时器
      this.clearRefreshInterval()
      
      // 只有当任务存在且状态为处理中或待处理时才设置定时器
      if (this.task && ['processing', 'pending'].includes(this.task.status)) {
        console.log('设置定时刷新任务:', this.task.id)
        this.refreshInterval = setInterval(() => {
          this.refreshTask()
        }, 5000) // 每5秒刷新一次
      } else {
        console.log('任务状态不需要定时刷新:', this.task ? this.task.status : 'undefined')
      }
    },

    // 清除自动刷新
    clearRefreshInterval() {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval)
        this.refreshInterval = null
      }
    },

    // 格式化日期
    formatDate(date) {
      if (!date || date === '0001-01-01T00:00:00Z') {
        return this.task && this.task.CreatedAt ? new Date(this.task.CreatedAt).toLocaleString() : '未知'
      }
      return new Date(date).toLocaleString()
    },
    
    // 解析输入参数
    parseInputParams() {
      if (!this.task || !this.task.input_params) return []
      
      try {
        const params = JSON.parse(this.task.input_params)
        if (Array.isArray(params)) {
          return params
        }
        // 如果不是数组，转换为数组格式
        return Object.keys(params).map(key => ({
          key,
          value: params[key],
          alias: key,
          type: this.guessParamType(params[key])
        }))
      } catch (error) {
        console.error('解析输入参数失败:', error)
        return []
      }
    },
    
    // 解析输出参数
    parseOutputParams() {
      if (!this.task || !this.task.output_params) return []
      
      try {
        const params = JSON.parse(this.task.output_params)
        if (Array.isArray(params)) {
          return params
        }
        // 如果不是数组，转换为数组格式
        return Object.keys(params).map(key => ({
          key,
          value: params[key],
          alias: key,
          type: this.guessParamType(params[key])
        }))
      } catch (error) {
        console.error('解析输出参数失败:', error)
        return []
      }
    },
    
    // 猜测参数类型
    guessParamType(value) {
      if (typeof value === 'string') {
        // 检查图片类型
        if (value.endsWith('.jpg') || value.endsWith('.jpeg') || value.endsWith('.png') || value.endsWith('.gif')) {
          return 'image'
        }
        // 检查视频类型
        if (value.endsWith('.mp4') || value.endsWith('.webm') || value.endsWith('.mov')) {
          return 'video'
        }
        // 检查音频类型
        if (value.endsWith('.mp3') || value.endsWith('.wav') || value.endsWith('.ogg')) {
          return 'audio'
        }
      }
      return 'text'
    },
    
    // 获取状态类型
    getStatusType(status) {
      const statusMap = {
        pending: 'info',
        processing: 'warning',
        completed: 'success',
        failed: 'danger',
        unknown: 'info'
      }
      return statusMap[status] || 'info'
    },

    // 获取状态文本
    getStatusText(status) {
      const statusMap = {
        pending: '待处理',
        processing: '处理中',
        completed: '已完成',
        failed: '失败',
        unknown: '未知'
      }
      return statusMap[status] || status
    },
    
    // 处理窗口大小变化
    handleResize() {
      this.screenWidth = window.innerWidth
      // 更新设备类型
      this.checkDeviceType()
    },
    
    // 检测设备类型
    checkDeviceType() {
      this.isMobile = window.innerWidth <= 768 // 假设小于等于768px为移动端
    },
    
    // 修复全局样式，确保菜单在合适的层级
    fixGlobalStyles() {
      // 动态创建样式元素
      this.globalStyleEl = document.createElement('style')
      this.globalStyleEl.type = 'text/css'
      this.globalStyleEl.innerHTML = `
        /* 底部菜单和侧边栏保持高层级 */
        .app-footer, .app-sidebar {
          z-index: 3000 !important;
          position: relative !important;
        }
        
        /* 顶部菜单放在较低层级，让详情页内容可以覆盖它 */
        .el-menu, .el-submenu, .el-menu-item, .app-header, .app-navbar {
          z-index: 1 !important;
        }
        
        /* 视频元素保持低层级 */
        video, iframe, embed, object {
          z-index: 1 !important;
        }
      `
      document.head.appendChild(this.globalStyleEl)
    },
    
    // 恢复全局样式
    restoreGlobalStyles() {
      if (this.globalStyleEl && this.globalStyleEl.parentNode) {
        this.globalStyleEl.parentNode.removeChild(this.globalStyleEl)
        this.globalStyleEl = null
      }
    },

    // 分享任务
    async shareTask() {
      try {
        if (!this.task || this.task.status !== 'completed') {
          this.$message.warning('只能分享已完成的任务')
          return
        }
        
        this.$confirm('确定要分享此任务到灵感页吗?', '分享确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'info'
        }).then(async () => {
          this.sharing = true
          this.$message({
            type: 'info',
            message: '正在提交分享请求...',
            duration: 2000
          })
          
          const result = await shareTask({
            taskId: this.task.id,
            mode: 'comfyui',
            taskType: this.task.task_type,
          })
          
          if (result.success) {
            this.$message.success(result.message)
            // 更新本地任务状态
            this.task.is_shared = true
            this.task.share_status = 'pending_review'
            this.task.share_time = new Date().toISOString()
          } else {
            this.$message.error('分享失败: ' + (result.message || '未知错误'))
          }
        }).catch(error => {
          if (error === 'cancel') {
            // 用户取消分享，不执行任何操作
          } else {
            this.$message.error('分享操作失败: ' + (error.message || '未知错误'))
          }
        }).finally(() => {
          this.sharing = false
        })
      } catch (error) {
        this.sharing = false
        this.$message.error('分享过程中发生错误: ' + error.message)
      }
    },
    
    // 获取分享按钮文本
    getShareButtonText() {
      if (!this.task) return '分享到灵感页'
      
      // 根据分享状态返回不同文本
      if (this.task.is_shared) {
        switch (this.task.share_status) {
          case 'pending_review':
            return '审核中'
          case 'rejected':
            return '分享被拒绝'
          case 'approved':
            return '已分享'
          default:
            return '分享到灵感页'
        }
      }
      return '分享到灵感页'
    }
  }
}
</script>

<style scoped>
/* 现有的PC端样式 */
.error-actions {
  margin-top: 16px;
  text-align: right;
}

/* 图片预览样式 */
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.9);
  z-index: 1000;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: grab;
}

.image-preview-overlay:active {
  cursor: grabbing;
}

.preview-image {
  max-width: 90%;
  max-height: 90vh;
  object-fit: contain;
  transform-origin: center;
  user-select: none;
  will-change: transform;
}

.workflow-detail-container {
  padding: 20px;
  overflow-x: hidden;
  transform: translateZ(0);
  isolation: isolate;
  position: relative;
  z-index: 100 !important;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.task-header {
  margin-bottom: 20px;
  background-color: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.info-item {
  margin: 10px 0;
  display: flex;
  align-items: center;
}

.label {
  font-weight: bold;
  margin-right: 10px;
  min-width: 80px;
  color: #606266;
}

/* 视频预览样式 */
.video-preview {
  margin-bottom: 15px;
  border-radius: 4px;
  overflow: hidden;
  background: #000;
}

.video-preview video {
  display: block;
  width: 100%;
  height: auto;
}

.task-section {
  margin-top: 30px;
  position: relative;
}

.task-section h3 {
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
  color: #303133;
}

.params-container,
.result-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.param-item,
.result-item {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 15px;
  transition: all 0.3s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.param-item:hover,
.result-item:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.param-header,
.result-header {
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #ebeef5;
}

.param-name,
.result-name {
  font-weight: bold;
  color: #409EFF;
}

.param-image,
.result-image,
.result-video {
  max-width: 100%;
  max-height: 300px;
  border-radius: 4px;
  display: block;
  margin: 0 auto;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s;
}

.param-image:hover,
.result-image:hover {
  transform: scale(1.02);
}

.result-actions {
  margin-top: 15px;
  text-align: center;
}

.error-container {
  background-color: #fef0f0;
  padding: 15px;
  border-radius: 4px;
  margin-top: 10px;
  border-left: 4px solid #f56c6c;
}

.error-message {
  color: #f56c6c;
  margin: 0;
  white-space: pre-wrap;
  font-family: monospace;
}

.no-params,
.no-results {
  padding: 20px;
  text-align: center;
}

.text-param,
.text-result {
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
  word-break: break-all;
}

/* 移动端新增样式 */
.workflow-detail-container {
  transform: translateZ(0);
  isolation: isolate;
  position: relative;
  z-index: 100 !important;
  overflow-x: hidden !important;
  width: 100% !important;
  max-width: 100% !important;
}

/* 强制所有视频元素不能超出其父容器的堆叠上下文 */
.workflow-detail-container video, 
.workflow-detail-container audio, 
.workflow-detail-container .video-container, 
.workflow-detail-container .video-preview, 
.workflow-detail-container .result-preview,
.workflow-detail-container .video-player {
  transform: translateZ(0);
  position: relative !important;
  z-index: 0 !important;
  isolation: isolate;
  max-width: 100%;
  max-height: 100%;
}

/* 移动端顶部导航栏 */
.mobile-header-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background-color: #409EFF;
  display: flex;
  align-items: center;
  padding: 0 12px;
  z-index: 2000 !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  color: #fff;
  width: 100%;
  box-sizing: border-box;
}

.header-back {
  display: flex;
  align-items: center;
  font-size: 16px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
  z-index: 2100 !important;
  position: relative;
}

.header-back:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.header-back i {
  margin-right: 4px;
  font-size: 16px;
}

.header-title {
  margin: 0 0 0 10px;
  font-size: 16px;
  font-weight: 500;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mobile-header-placeholder {
  height: 56px;
  width: 100%;
  flex-shrink: 0;
}

/* 移动端内容样式 */
.mobile-content-view {
  padding: 0;
  width: 100%;
  background-color: #fff;
  min-height: 100%;
  box-sizing: border-box;
}

.mobile-content-inner {
  min-height: calc(100vh - 56px);
  width: 100%;
  box-sizing: border-box;
}

.detail-content-wrapper {
  width: 100%;
}

/* 移动端基本信息 */
.basic-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #ebeef5;
  width: 100%;
  box-sizing: border-box;
}

.create-time {
  font-size: 12px;
  color: #909399;
}

.mobile-task-info {
  padding: 8px 12px;
  box-sizing: border-box;
  width: 100%;
}

.mobile-section {
  padding: 8px 12px;
  border-top: 1px solid #ebeef5;
  margin-top: 8px;
  box-sizing: border-box;
  width: 100%;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin: 5px 0;
  color: #303133;
  width: 100%;
  box-sizing: border-box;
}

/* 移动端参数和结果样式 */
.mobile-params, .mobile-results {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 10px;
}

.mobile-param-item, .mobile-result-item {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 10px;
  margin-bottom: 10px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.mobile-param-header, .mobile-result-header {
  margin-bottom: 8px;
  padding-bottom: 5px;
  border-bottom: 1px dashed #ebeef5;
}

.mobile-param-name, .mobile-result-name {
  font-weight: bold;
  color: #409EFF;
  font-size: 14px;
}

.mobile-param-image, .mobile-result-image {
  max-width: 100%;
  max-height: 200px;
  border-radius: 4px;
  display: block;
  margin: 0 auto;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

.mobile-text-param, .mobile-text-result {
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
  word-break: break-all;
}

.mobile-video-preview {
  margin-bottom: 10px;
  border-radius: 4px;
  overflow: hidden;
  background: #000;
}

.mobile-video {
  display: block;
  width: 100%;
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.mobile-result-actions {
  margin-top: 10px;
  text-align: center;
}

.mobile-error-container {
  background-color: #fef0f0;
  padding: 10px;
  border-radius: 4px;
  margin-top: 5px;
  border-left: 4px solid #f56c6c;
}

.mobile-error-message {
  color: #f56c6c;
  margin: 0;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 12px;
}

.mobile-no-params, .mobile-no-results {
  padding: 15px;
  text-align: center;
}

.mobile-actions {
  display: flex;
  gap: 10px;
  padding: 15px 12px;
  margin-top: 20px;
  margin-bottom: 30px;
  justify-content: center;
}

.mobile-actions .el-button {
  flex: 1;
}

.mobile-no-task {
  padding: 40px 20px;
  text-align: center;
}

/* 移动端适配 */
@media screen and (max-width: 768px) {
  .workflow-detail-container {
    padding: 0;
    width: 100%;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 5 !important;
    background-color: #fff;
    color: #333;
    height: 100vh;
    overflow-x: hidden;
    box-sizing: border-box;
    transform: translateZ(0);
    isolation: isolate;
  }
  
  .detail-content-wrapper {
    transform: translateZ(0);
    position: absolute;
    top: 56px;
    left: 0;
    right: 0;
    bottom: 0;
    overflow-y: auto;
    overflow-x: hidden;
    box-sizing: border-box;
    padding-bottom: 120px !important;
    z-index: 1 !important;
    isolation: isolate;
  }
  
  .mobile-image-param, .mobile-image-result {
    margin: 10px 0;
    text-align: center;
  }
  
  .mobile-video-preview {
    margin: 10px 0;
  }
}
</style>