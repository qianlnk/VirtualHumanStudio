<template>
  <div class="review-tasks-container">
    <div class="page-header">
      <h2>审核分享任务</h2>
      <div class="actions">
        <el-button type="primary" @click="fetchTasks">刷新</el-button>
      </div>
    </div>

    <div class="content-wrapper">
      <!-- 任务列表 -->
      <el-card v-loading="loading" shadow="never" class="task-list-card">
        <!-- 空状态 -->
        <el-empty v-if="tasks.length === 0 && !loading" description="暂无待审核的任务"></el-empty>

        <!-- 任务卡片列表 -->
        <div v-if="tasks.length > 0" class="task-grid">
          <el-card v-for="task in tasks" :key="`${task.type}-${task.id}`" class="task-card" shadow="hover">
            <!-- 卡片头部 -->
            <div slot="header" class="task-card-header">
              <div class="task-info">
                <span class="task-title">{{ task.name || '未命名任务' }}</span>
                <el-tag size="small" :type="getTaskTypeTagType(task.type)">
                  {{ getTaskTypeText(task.type) }}: {{ task.task_type }}
                </el-tag>
              </div>
              <div class="task-actions">
                <el-button 
                  type="success" 
                  size="mini" 
                  @click="openReviewDialog(task, 'approve')"
                  :loading="reviewLoading && currentTaskId === task.id">
                  通过
                </el-button>
                <el-button 
                  type="danger" 
                  size="mini" 
                  @click="openReviewDialog(task, 'reject')"
                  :loading="reviewLoading && currentTaskId === task.id">
                  拒绝
                </el-button>
              </div>
            </div>

            <!-- 卡片内容 -->
            <div class="task-content">
              <!-- 作者信息 -->
              <div class="task-creator">
                <i class="el-icon-user"></i>
                <span>{{ task.username || '未知用户' }}</span>
                <span class="task-time">{{ task.created_at }}</span>
              </div>

              <!-- 任务描述 -->
              <div v-if="task.description" class="task-description">
                <div class="description-label">描述：</div>
                <div class="description-content">{{ task.description }}</div>
              </div>

              <!-- 结果预览 -->
              <div class="result-preview">
                <div v-if="task.type === 'digital_human'" class="video-container">
                  <video 
                    :src="task.result_url" 
                    controls 
                    class="preview-video"
                    @error="handleVideoError($event, task.result_url)"
                    @canplay="videoLoaded($event, task.result_url)"
                    preload="metadata"
                    playsinline
                    :key="`main-video-${task.id}`">
                    您的浏览器不支持视频播放
                  </video>
                  <div v-if="!isVideoLoaded(task.result_url)" class="video-loading">
                    <i class="el-icon-loading"></i>
                    <span>加载中...</span>
                    <el-button type="text" @click="forceLoadVideo(task.result_url)" class="force-load-btn">
                      点击显示
                    </el-button>
                  </div>
                </div>
                <div v-else class="image-container">
                  <img :src="task.result_url" :alt="task.name" class="preview-image">
                </div>
              </div>

              <!-- 任务参数 -->
              <div class="params-section" v-if="task.input_params && task.input_params.length > 0">
                <el-collapse>
                  <el-collapse-item title="输入参数" name="input">
                    <div class="params-list">
                      <div v-for="(param, index) in task.input_params" :key="`input-${index}`" class="param-item">
                        <span class="param-label">{{ param.label || param.key }}:</span>
                        <span class="param-value">
                          <span v-if="param.type === 'image'">
                            <el-image 
                              style="width: 60px; height: 60px"
                              :src="param.value" 
                              :preview-src-list="[param.value]">
                            </el-image>
                          </span>
                          <span v-else-if="param.type === 'video'">
                            <div class="param-video-container">
                              <video 
                                style="width: 100%; max-width: 280px; height: auto; max-height: 160px"
                                :src="param.value" 
                                controls
                                @error="handleVideoError($event, param.value)"
                                @canplay="videoLoaded($event, param.value)"
                                preload="metadata"
                                playsinline
                                :key="`param-${param.key}-${task.id}`"
                                class="param-video">
                                您的浏览器不支持视频播放
                              </video>
                              <div v-if="!isVideoLoaded(param.value)" class="param-video-loading">
                                <i class="el-icon-loading"></i>
                                <span>加载中...</span>
                                <el-button type="text" @click="forceLoadVideo(param.value)" class="force-load-btn">
                                  点击显示
                                </el-button>
                              </div>
                            </div>
                          </span>
                          <span v-else-if="param.type === 'audio'">
                            <div class="param-audio-container">
                              <audio 
                                controls
                                style="width: 100%; max-width: 250px"
                                :src="param.value"
                                @error="handleAudioError($event, param.value)">
                              </audio>
                              <span class="param-audio-filename">{{ getFileName(param.value) }}</span>
                            </div>
                          </span>
                          <span v-else>{{ param.value }}</span>
                        </span>
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>

              <div class="params-section" v-if="task.output_params && task.output_params.length > 0">
                <el-collapse>
                  <el-collapse-item title="输出参数" name="output">
                    <div class="params-list">
                      <div v-for="(param, index) in task.output_params" :key="`output-${index}`" class="param-item">
                        <span class="param-label">{{ param.label || param.key }}:</span>
                        <span class="param-value">
                          <span v-if="param.type === 'image'">
                            <el-image 
                              style="width: 60px; height: 60px"
                              :src="param.value" 
                              :preview-src-list="[param.value]">
                            </el-image>
                          </span>
                          <span v-else-if="param.type === 'video'">
                            <div class="param-video-container">
                              <video 
                                style="width: 100%; max-width: 280px; height: auto; max-height: 160px"
                                :src="param.value" 
                                controls
                                @error="handleVideoError($event, param.value)"
                                @canplay="videoLoaded($event, param.value)"
                                preload="metadata"
                                playsinline
                                :key="`param-${param.key}-${task.id}`"
                                class="param-video">
                                您的浏览器不支持视频播放
                              </video>
                              <div v-if="!isVideoLoaded(param.value)" class="param-video-loading">
                                <i class="el-icon-loading"></i>
                                <span>加载中...</span>
                                <el-button type="text" @click="forceLoadVideo(param.value)" class="force-load-btn">
                                  点击显示
                                </el-button>
                              </div>
                            </div>
                          </span>
                          <span v-else-if="param.type === 'audio'">
                            <div class="param-audio-container">
                              <audio 
                                controls
                                style="width: 100%; max-width: 250px"
                                :src="param.value"
                                @error="handleAudioError($event, param.value)">
                              </audio>
                              <span class="param-audio-filename">{{ getFileName(param.value) }}</span>
                            </div>
                          </span>
                          <span v-else>{{ param.value }}</span>
                        </span>
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
            </div>
          </el-card>
        </div>

        <!-- 分页 -->
        <div class="pagination-container">
          <el-pagination
            background
            layout="prev, pager, next"
            :total="total"
            :page-size="pageSize"
            @current-change="handlePageChange"
            :current-page.sync="currentPage">
          </el-pagination>
        </div>
      </el-card>
    </div>

    <!-- 审核对话框 -->
    <el-dialog 
      :title="reviewAction === 'approve' ? '通过审核' : '拒绝审核'" 
      :visible.sync="reviewDialogVisible"
      width="500px">
      <div class="review-dialog-content">
        <div v-if="currentTask" class="task-info-summary">
          <div class="review-task-title">任务: {{ currentTask.name || '未命名任务' }}</div>
          <div>类型: {{ getTaskTypeText(currentTask.type) }} - {{ currentTask.task_type }}</div>
          <div>提交者: {{ currentTask.username || '未知用户' }}</div>
        </div>

        <div v-if="reviewAction === 'reject'" class="reject-reason">
          <el-form :model="reviewForm" ref="reviewForm" :rules="reviewRules">
            <el-form-item label="拒绝原因" prop="rejectReason">
              <el-input
                type="textarea"
                :rows="3"
                placeholder="请输入拒绝原因"
                v-model="reviewForm.rejectReason">
              </el-input>
            </el-form-item>
          </el-form>
        </div>

        <div v-else class="approve-message">
          您确定要通过这个任务的分享申请吗？通过后将在灵感页面对所有用户展示。
        </div>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button @click="reviewDialogVisible = false">取消</el-button>
        <el-button 
          :type="reviewAction === 'approve' ? 'success' : 'danger'"
          @click="submitReview"
          :loading="reviewLoading">
          {{ reviewAction === 'approve' ? '通过' : '拒绝' }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getPendingReviewTasks, reviewTask } from '@/api/share'

export default {
  name: 'ReviewTasks',
  data() {
    return {
      loading: false,
      tasks: [],
      total: 0,
      currentPage: 1,
      pageSize: 12,
      reviewDialogVisible: false,
      reviewAction: 'approve', // 'approve' 或 'reject'
      currentTask: null,
      currentTaskId: null,
      reviewLoading: false,
      reviewForm: {
        rejectReason: ''
      },
      reviewRules: {
        rejectReason: [
          { required: true, message: '请输入拒绝原因', trigger: 'blur' },
          { min: 2, max: 200, message: '长度在 2 到 200 个字符', trigger: 'blur' }
        ]
      },
      videoLoadedMap: {},
      prevTasksLength: 0, // 用于跟踪任务列表变化
      preloadTimer: null // 用于防抖处理
    }
  },
  created() {
    this.fetchTasks()
  },
  mounted() {
    // 在组件挂载后添加监听
    window.addEventListener('load', this.preloadVideos)
  },
  updated() {
    // 优化更新后的处理，不再每次更新都重新预加载视频
    // 只有在显示的任务发生变化时才触发预加载
    if (this.prevTasksLength !== this.tasks.length) {
      this.prevTasksLength = this.tasks.length
      // 延迟执行预加载，避免阻塞主线程
      setTimeout(() => {
        this.preloadVideos()
      }, 500)
    }
  },
  beforeDestroy() {
    // 组件销毁前移除监听
    window.removeEventListener('load', this.preloadVideos)
  },
  methods: {
    // 预加载视频
    preloadVideos() {
      // 清除可能正在执行的预加载定时器
      if (this.preloadTimer) {
        clearTimeout(this.preloadTimer)
        this.preloadTimer = null
      }
      
      console.log('===== 开始预处理视频加载 =====')
      
      // 优先处理主要视频元素，然后再处理参数中的视频
      this.$nextTick(() => {
        // 首先简单标记所有任务的主要视频为未加载状态
        this.tasks.forEach(task => {
          if (task.type === 'digital_human' && task.result_url) {
            this.$set(this.videoLoadedMap, task.result_url, false)
          }
        })
        
        // 延迟处理实际视频加载，避免界面卡顿
        this.preloadTimer = setTimeout(() => {
          this.batchProcessVideos()
        }, 300)
      })
    },
    
    // 分批处理视频，避免一次处理太多导致界面卡顿
    batchProcessVideos() {
      // 获取所有视频元素
      const videos = Array.from(document.querySelectorAll('video'))
      
      // 如果没有视频，直接返回
      if (videos.length === 0) {
        console.log('没有找到视频元素')
        return
      }
      
      console.log(`找到 ${videos.length} 个视频元素，开始分批处理`)
      
      // 一次只处理少量视频，防止卡顿
      const batchSize = 2
      let processedCount = 0
      
      // 处理单个批次
      const processBatch = () => {
        // 处理当前批次
        const currentBatch = videos.slice(processedCount, processedCount + batchSize)
        processedCount += currentBatch.length
        
        // 如果没有视频需要处理，结束
        if (currentBatch.length === 0) {
          return
        }
        
        // 处理当前批次的视频
        currentBatch.forEach(video => {
          // 跳过没有src的视频
          if (!video.src) return
          
          // 为视频添加事件监听
          this.setupVideoEvents(video)
        })
        
        // 如果还有视频需要处理，安排下一批
        if (processedCount < videos.length) {
          setTimeout(processBatch, 100)
        } else {
          console.log('===== 所有视频预处理完成 =====')
        }
      }
      
      // 开始处理第一批
      processBatch()
    },
    
    // 为单个视频设置事件监听
    setupVideoEvents(video) {
      if (!video || !video.src) return
      
      const videoId = video.src
      
      // 移除之前可能存在的事件监听器
      video.onloadeddata = null
      video.oncanplay = null
      video.onerror = null
      
      // 添加事件监听器
      video.onloadeddata = () => this.handleVideoEvent(video, 'loadeddata')
      video.oncanplay = () => this.handleVideoEvent(video, 'canplay')
      video.onerror = (e) => {
        console.error(`视频加载失败: ${videoId}`, e)
        this.$set(this.videoLoadedMap, videoId, false)
      }
      
      // 检查视频是否已经加载
      if (video.readyState >= 2) {
        this.$set(this.videoLoadedMap, videoId, true)
      }
    },

    // 获取待审核任务列表
    async fetchTasks() {
      this.loading = true
      
      // 防止预加载视频导致页面卡顿
      if (this.preloadTimer) {
        clearTimeout(this.preloadTimer)
        this.preloadTimer = null
      }
      
      // 在开始新的加载前清除之前的视频加载状态
      this.videoLoadedMap = {}
      
      try {
        const result = await getPendingReviewTasks(this.currentPage, this.pageSize)
        if (result.success) {
          this.tasks = result.tasks || []
          this.total = result.total || 0
        } else {
          this.$message.error(result.message || '获取待审核任务失败')
        }
      } catch (error) {
        console.error('获取待审核任务失败:', error)
        this.$message.error('获取待审核任务失败: ' + (error.message || '未知错误'))
      } finally {
        this.loading = false
        
        // 任务列表加载完成后，延迟一段时间再预加载视频
        // 这样可以让页面先渲染出基本内容，再处理视频加载
        setTimeout(() => {
          this.prevTasksLength = this.tasks.length
          this.preloadVideos()
        }, 800)
      }
    },

    // 处理分页变化
    handlePageChange(page) {
      this.currentPage = page
      this.fetchTasks()
    },

    // 获取任务类型标签类型
    getTaskTypeTagType(type) {
      return type === 'comfyui' ? 'success' : 'primary'
    },

    // 获取任务类型文本
    getTaskTypeText(type) {
      return type === 'comfyui' ? '图像处理' : '数字人'
    },

    // 打开审核对话框
    openReviewDialog(task, action) {
      this.currentTask = task
      this.currentTaskId = task.id
      this.reviewAction = action
      this.reviewForm.rejectReason = ''
      this.reviewDialogVisible = true
    },

    // 提交审核
    async submitReview() {
      // 如果是拒绝操作，先验证表单
      if (this.reviewAction === 'reject') {
        try {
          await this.$refs.reviewForm.validate()
        } catch (error) {
          return
        }
      }

      this.reviewLoading = true
      try {
        const result = await reviewTask({
          shareId: this.currentTask.id,
          taskId: this.currentTask.task_id,
          mode: this.currentTask.mode,
          taskType: this.currentTask.task_type,
          status: this.reviewAction === 'approve' ? 'approved' : 'rejected',
          rejectReason: this.reviewForm.rejectReason
        })

        if (result.success) {
          this.$message.success(result.message || '审核操作成功')
          this.reviewDialogVisible = false
          // 刷新任务列表
          this.fetchTasks()
        } else {
          this.$message.error(result.message || '审核操作失败')
        }
      } catch (error) {
        console.error('审核任务失败:', error)
        this.$message.error('审核任务失败: ' + (error.message || '未知错误'))
      } finally {
        this.reviewLoading = false
        this.currentTaskId = null
      }
    },

    handleVideoError(event, url) {
      // 如果没有URL，不进行处理
      if (!url) return
      
      // 不输出太多日志，只在开发环境输出
      if (process.env.NODE_ENV === 'development') {
        console.error(`视频加载失败: ${url}`)
      }
      
      // 标记视频加载失败
      this.$set(this.videoLoadedMap, url, false)
    },

    videoLoaded(event, url) {
      if (event && event.target && url) {
        const video = event.target
        
        // 检查视频状态，多种事件都可能触发此方法
        const isLoaded = 
          // 视频已经加载足够数据可以开始播放
          video.readyState >= 3 || 
          // 特定事件类型表明视频已经可以播放
          event.type === 'canplay' || 
          event.type === 'play' || 
          event.type === 'timeupdate' ||
          // 视频已经开始播放
          !video.paused
        
        if (isLoaded) {
          console.log(`视频加载成功 [${event.type}]: ${url}, 状态: ${video.readyState}`)
          this.$set(this.videoLoadedMap, url, true)
          
          // 如果视频是自动播放的，设置为静音以绕过浏览器限制
          if (!video.paused && video.muted === false) {
            video.muted = true
          }
        } else {
          console.log(`视频加载中 [${event.type}]: ${url}, 状态: ${video.readyState}`)
        }
      }
    },

    isVideoLoaded(url) {
      // 如果没有URL，认为已加载
      if (!url) return true
      
      // 查找URL对应的加载状态
      const status = this.videoLoadedMap[url]
      
      // 如果状态为明确的true，则返回true
      // 否则返回false（包括状态不存在或为false的情况）
      return status === true
    },

    handleAudioError(event, value) {
      console.error(`音频加载失败: ${value}`, event)
    },

    getFileName(value) {
      const parts = value.split('/')
      return parts[parts.length - 1]
    },

    forceLoadVideo(url) {
      if (!url) {
        console.warn('视频URL为空，无法加载')
        return
      }
      
      console.log('手动加载视频:', url)
      
      // 立即标记为已加载，移除加载中状态
      this.$set(this.videoLoadedMap, url, true)
      
      // 尝试加载对应的视频但不阻塞界面
      setTimeout(() => {
        try {
          const videos = document.querySelectorAll(`video[src="${url}"]`)
          if (videos.length === 0) {
            return
          }
          
          // 对于找到的每个视频，尝试加载
          Array.from(videos).forEach(video => {
            try {
              // 简单调用load方法，不进行复杂的DOM操作
              video.load()
            } catch (err) {
              console.warn('加载视频时出错:', err)
            }
          })
        } catch (err) {
          console.error('强制加载视频出错:', err)
        }
      }, 50)
    },
    
    // 处理视频事件
    handleVideoEvent(video, eventType) {
      if (!video || !video.src) return
      
      // 简化日志，减少控制台输出
      if (eventType === 'canplay' || eventType === 'loadeddata') {
        // 视频可以播放，更新状态
        this.$set(this.videoLoadedMap, video.src, true)
      }
    }
  }
}
</script>

<style scoped>
.review-tasks-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.content-wrapper {
  background-color: #f5f7fa;
  border-radius: 4px;
  padding: 20px;
}

.task-list-card {
  margin-bottom: 20px;
}

.task-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.task-card {
  transition: all 0.3s;
}

.task-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.1);
}

.task-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-info {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.task-title {
  font-weight: bold;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}

.task-actions {
  display: flex;
  gap: 10px;
}

.task-content {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.task-creator {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  color: #909399;
}

.task-time {
  margin-left: auto;
}

.task-description {
  background-color: #f8f9fa;
  padding: 10px;
  border-radius: 4px;
  font-size: 14px;
}

.description-label {
  font-weight: bold;
  margin-bottom: 5px;
}

.result-preview {
  width: 100%;
  margin: 10px 0;
  text-align: center;
}

.preview-image {
  max-width: 100%;
  max-height: 200px;
  object-fit: contain;
  border-radius: 4px;
}

.preview-video {
  max-width: 100%;
  max-height: 300px;
  width: 100%;
  height: auto;
  border-radius: 4px;
  background-color: #000;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  object-fit: contain;
  position: relative;
  z-index: 1;
  display: block;
}

.params-section {
  margin-top: 10px;
}

.params-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.param-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  font-size: 13px;
}

.param-label {
  font-weight: bold;
  min-width: 80px;
}

.param-value {
  flex: 1;
  word-break: break-word;
}

.param-video {
  width: 100%;
  border-radius: 4px;
  background-color: #000;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  object-fit: contain;
  position: relative;
  z-index: 1;
  display: block;
}

audio {
  border-radius: 12px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.review-dialog-content {
  margin-bottom: 20px;
}

.task-info-summary {
  background-color: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 20px;
}

.review-task-title {
  font-weight: bold;
  margin-bottom: 5px;
}

.reject-reason {
  margin-top: 20px;
}

.approve-message {
  color: #67c23a;
  padding: 10px;
  background-color: #f0f9eb;
  border-radius: 4px;
  text-align: center;
}

@media (max-width: 768px) {
  .task-grid {
    grid-template-columns: 1fr;
  }
  
  .task-card-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .task-actions {
    margin-top: 10px;
    width: 100%;
    justify-content: space-between;
  }
}

.video-container {
  position: relative;
  width: 100%;
  margin: 10px 0;
  min-height: 250px;
  background-color: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
  transform: translateZ(0);  /* 创建新的堆叠上下文 */
  isolation: isolate;  /* 创建新的堆叠上下文，现代浏览器 */
  z-index: 1;
}

.video-loading {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.5);
  color: white;
  border-radius: 4px;
  z-index: 2;
}

.video-loading i {
  font-size: 24px;
  margin-bottom: 5px;
}

.force-load-btn {
  margin-top: 5px;
  color: #fff;
  text-decoration: underline;
  background: transparent;
  border: none;
  box-shadow: none;
}

.force-load-btn:hover {
  text-shadow: 0 0 5px rgba(255,255,255,0.5);
}

.param-video-container, .param-audio-container {
  position: relative;
  width: 100%;
  margin-bottom: 10px;
  min-height: 120px;
  background-color: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
  transform: translateZ(0);  /* 创建新的堆叠上下文 */
  isolation: isolate;  /* 创建新的堆叠上下文，现代浏览器 */
  z-index: 1;
  padding-bottom: 8px;
}

.param-video-loading {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.5);
  color: white;
  border-radius: 4px;
  z-index: 2;
}

.param-audio-filename {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>