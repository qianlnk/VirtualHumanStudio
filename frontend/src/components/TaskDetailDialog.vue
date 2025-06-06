<template>
  <el-dialog
    title="创作详情"
    :visible.sync="dialogVisible"
    :width="isMobile ? '95%' : '60%'"
    append-to-body
    :fullscreen="isMobile"
    :custom-class="isMobile ? 'mobile-detail-dialog' : ''"
    @closed="handleClosed">
    <div v-if="task" class="task-detail">
      <div class="detail-header">
        <h2>{{ task.name }}</h2>
        <p v-if="task.description">{{ task.description }}</p>
      </div>
      
      <!-- 预览区域 -->
      <div class="detail-preview">
        <!-- 数字人视频 -->
        <div v-if="task.type === 'digital_human'" class="detail-video">
          <h3>合成视频</h3>
          <div class="video-placeholder" v-if="!detailVideoLoaded" @click="loadDetailVideo">
            <div class="loading-indicator">
              <i class="el-icon-video-play"></i>
              <span>点击加载视频</span>
            </div>
          </div>
          <video 
            v-show="detailVideoLoaded"
            ref="detailVideo"
            class="full-video"
            controls 
            preload="none"
            playsinline
            webkit-playsinline
            @click="loadDetailVideo"
            @canplay="handleDetailVideoCanPlay"
            @error="handleDetailVideoError">
            <source :src="detailVideoLoaded ? getDirectVideoUrl(task.result_url) : ''" type="video/mp4">
            您的浏览器不支持视频播放
          </video>
        </div>
        
        <!-- 图像处理结果 -->
        <div v-else class="detail-image">
          <h3>处理结果</h3>
          <img :src="task.result_url" :alt="task.name" class="full-image" @click="previewImage(task.result_url)">
        </div>
      </div>
      
      <!-- 输入参数 -->
      <div class="detail-params" v-if="task.input_params && task.input_params.length > 0">
        <h3>输入参数</h3>
        <div class="params-grid">
          <div v-for="(param, index) in task.input_params" :key="'input-'+index" class="param-item">
            <div class="param-label">{{ param.label || param.key }}</div>
            
            <!-- 不同类型参数的展示 -->
            <div v-if="param.type === 'image' || param.type === 'mask'" class="param-image">
              <img :src="param.value" :alt="param.label" class="thumbnail" @click="previewImage(param.value)">
            </div>
            <div v-else-if="param.type === 'video'" class="param-video">
              <div class="video-placeholder" v-if="!paramVideoLoadedMap[param.value]" @click="loadParamVideo($event, param.value)">
                <div class="loading-indicator small">
                  <i class="el-icon-video-play"></i>
                  <span>点击加载</span>
                </div>
              </div>
              <video 
                v-show="paramVideoLoadedMap[param.value]"
                :data-url="param.value"
                class="thumbnail-video" 
                controls
                playsinline
                webkit-playsinline
                preload="none"
                @click="loadParamVideo($event, param.value)"
                @canplay="handleParamVideoCanPlay(param.value)"
                @error="handleParamVideoError($event, param.value)">
                <source :src="paramVideoLoadedMap[param.value] ? getDirectVideoUrl(param.value) : ''" type="video/mp4">
                您的浏览器不支持视频播放
              </video>
            </div>
            <div v-else-if="param.type === 'audio'" class="param-audio">
              <audio :src="param.value" controls class="audio-player" preload="none"></audio>
            </div>
            <div v-else class="param-text">
              {{ param.value }}
            </div>
          </div>
        </div>
      </div>
      
      <!-- 输出参数 -->
      <div class="detail-params" v-if="task.output_params && task.output_params.length > 0">
        <h3>输出参数</h3>
        <div class="params-grid">
          <div v-for="(param, index) in task.output_params" :key="'output-'+index" class="param-item">
            <div class="param-label">{{ param.label || param.key }}</div>
            
            <!-- 不同类型参数的展示 -->
            <div v-if="param.type === 'image' || param.type === 'mask'" class="param-image">
              <img :src="param.value" :alt="param.label" class="thumbnail" @click="previewImage(param.value)">
            </div>
            <div v-else-if="param.type === 'video'" class="param-video">
              <div class="video-placeholder" v-if="!paramVideoLoadedMap[param.value]" @click="loadParamVideo($event, param.value)">
                <div class="loading-indicator small">
                  <i class="el-icon-video-play"></i>
                  <span>点击加载</span>
                </div>
              </div>
              <video 
                v-show="paramVideoLoadedMap[param.value]"
                :data-url="param.value"
                class="thumbnail-video" 
                controls
                playsinline
                webkit-playsinline
                preload="none"
                @click="loadParamVideo($event, param.value)"
                @canplay="handleParamVideoCanPlay(param.value)"
                @error="handleParamVideoError($event, param.value)">
                <source :src="paramVideoLoadedMap[param.value] ? getDirectVideoUrl(param.value) : ''" type="video/mp4">
                您的浏览器不支持视频播放
              </video>
            </div>
            <div v-else-if="param.type === 'audio'" class="param-audio">
              <audio :src="param.value" controls class="audio-player" preload="none"></audio>
            </div>
            <div v-else class="param-text">
              {{ param.value }}
            </div>
          </div>
        </div>
      </div>
      
      <div class="detail-meta">
        <div class="meta-item">
          <span class="meta-label">创建者：</span>
          <span class="meta-value">{{ task.username }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">创建时间：</span>
          <span class="meta-value">{{ formatDate(task.created_at) }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">任务类型：</span>
          <span class="meta-value">
            <el-tag size="mini" :type="getTaskTypeTag(task)">
              {{ getTaskTypeText(task) }}
            </el-tag>
          </span>
        </div>
      </div>
    </div>
    
    <!-- 图片预览 -->
    <el-dialog
      :visible.sync="previewVisible"
      :width="isMobile ? '100%' : '80%'"
      :fullscreen="isMobile"
      center
      custom-class="preview-dialog"
      append-to-body>
      <img :src="previewUrl" alt="预览图" class="preview-fullsize">
    </el-dialog>
  </el-dialog>
</template>

<script>
import { getDirectFileUrl } from '@/utils/fileAccess'

export default {
  name: 'TaskDetailDialog',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    task: {
      type: Object,
      default: null
    }
  },
  data() {
    return {
      dialogVisible: false,
      detailVideoLoaded: false, // 详情视频是否已加载
      paramVideoLoadedMap: {}, // 跟踪参数视频是否已加载
      previewVisible: false,
      previewUrl: '',
      isMobile: false
    }
  },
  watch: {
    visible(val) {
      this.dialogVisible = val
      if (val) {
        // 打开对话框时重置状态
        this.detailVideoLoaded = false
        this.paramVideoLoadedMap = {}
        this.checkMobileDevice()
      }
    },
    dialogVisible(val) {
      // 同步回父组件
      this.$emit('update:visible', val)
      if (!val) {
        this.$emit('close')
      }
    }
  },
  methods: {
    // 检测移动设备
    checkMobileDevice() {
      // 移动设备检测
      const userAgent = navigator.userAgent || navigator.vendor || window.opera
      const isMobileByUA = /android|webos|iphone|ipad|ipod|blackberry|iemobile|opera mini|mobile/i.test(userAgent)
      
      // 屏幕宽度检测（作为备选）
      const isMobileByWidth = window.innerWidth <= 576
      
      // 更新移动设备状态
      this.isMobile = isMobileByUA || isMobileByWidth
    },
    
    // 获取任务类型标签类型
    getTaskTypeTag(task) {
      if (task.type === 'digital_human') {
        return 'primary'
      } else {
        switch (task.task_type) {
          case 'accessory':
            return 'success'
          case 'workflow':
            return 'warning'
          default:
            return 'info'
        }
      }
    },
    
    // 获取任务类型文本
    getTaskTypeText(task) {
      if (task.type === 'digital_human') {
        return '数字人'
      } else {
        switch (task.task_type) {
          case 'accessory':
            return '饰品替换'
          case 'workflow':
            return '图像处理'
          default:
            return task.task_type || '图像处理'
        }
      }
    },
    
    // 加载详情视频
    loadDetailVideo() {
      if (this.detailVideoLoaded) return // 如果已加载，不做处理
      
      // 标记为已加载状态
      this.detailVideoLoaded = true
      
      // 确保视频元素重新加载
      if (this.$refs.detailVideo) {
        try {
          // 设置视频元素样式确保正常显示
          this.$refs.detailVideo.style.zIndex = "1"
          this.$refs.detailVideo.style.transform = "translateZ(0)"
          this.$refs.detailVideo.style.backgroundColor = "#000"
          // 设置预加载策略
          this.$refs.detailVideo.preload = "metadata"
          
          // 清除现有源并添加多格式支持
          if (this.task && this.task.result_url) {
            this.$refs.detailVideo.innerHTML = this.getVideoSourceElements(this.task.result_url)
          }
          
          this.$refs.detailVideo.load()
        } catch (error) {
          console.error('加载详情视频失败:', error)
          this.detailVideoLoaded = false
        }
      }
    },
    
    // 详情视频可以播放时
    handleDetailVideoCanPlay() {
      // 标记视频已完成加载
      this.detailVideoLoaded = true
    },
    
    // 详情视频加载错误
    handleDetailVideoError(event) {
      console.error('详情视频加载失败:', event)
      // 重置视频加载状态，允许用户重试
      this.detailVideoLoaded = false
      
      if (this.task && this.task.result_url) {
        // 设置延迟，不要立即重试
        setTimeout(() => {
          // 尝试重新加载视频
          this.retryLoadVideo(this.$refs.detailVideo, this.task.result_url)
        }, 1000)
      }
    },
    
    // 加载参数视频
    loadParamVideo(event, url) {
      if (this.paramVideoLoadedMap[url]) return // 如果已加载，不做处理
      
      // 标记为已加载状态
      this.$set(this.paramVideoLoadedMap, url, true)
      
      // 停止事件冒泡
      event.stopPropagation()
      
      const videoEl = event.target
      // 确保视频元素重新加载
      if (videoEl && videoEl.tagName === 'VIDEO') {
        try {
          // 设置视频元素样式确保正常显示
          videoEl.style.zIndex = "1"
          videoEl.style.transform = "translateZ(0)"
          videoEl.style.backgroundColor = "#000"
          // 设置预加载策略
          videoEl.preload = "metadata"
          
          // 清除现有源并添加多格式支持
          videoEl.innerHTML = this.getVideoSourceElements(url)
          
          videoEl.load()
        } catch (error) {
          console.error('加载参数视频失败:', error)
          this.$set(this.paramVideoLoadedMap, url, false)
        }
      }
    },
    
    // 参数视频可以播放时
    handleParamVideoCanPlay(url) {
      // 标记视频已完成加载
      this.$set(this.paramVideoLoadedMap, url, true)
    },
    
    // 参数视频加载错误
    handleParamVideoError(event, url) {
      console.error('参数视频加载失败:', url, event)
      // 重置视频加载状态，允许用户重试
      this.$set(this.paramVideoLoadedMap, url, false)
      
      if (url) {
        // 设置延迟，不要立即重试
        setTimeout(() => {
          // 找到对应的视频元素
          const videoEl = event.target.closest('video')
          if (videoEl) {
            // 尝试重新加载视频
            this.retryLoadVideo(videoEl, url)
          }
        }, 1000)
      }
    },
    
    // 时间格式化（相对时间）
    formatTime(timeStr) {
      const date = new Date(timeStr)
      const now = new Date()
      const diff = Math.floor((now - date) / 1000) // 秒数差
      
      if (diff < 60) {
        return '刚刚'
      } else if (diff < 3600) {
        return Math.floor(diff / 60) + '分钟前'
      } else if (diff < 86400) {
        return Math.floor(diff / 3600) + '小时前'
      } else if (diff < 2592000) {
        return Math.floor(diff / 86400) + '天前'
      } else {
        return this.formatDate(timeStr)
      }
    },
    
    // 完整日期格式化
    formatDate(timeStr) {
      const date = new Date(timeStr)
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    },
    
    // 预览图片
    previewImage(url) {
      this.previewUrl = url
      this.previewVisible = true
    },
    
    // 检查视频格式添加多种格式支持
    getVideoSourceElements(url) {
      if (!url) return ''
      
      const baseUrl = this.getDirectVideoUrl(url)
      // 检查URL是否已经包含文件扩展名
      const hasExtension = /\.(mp4|webm|mov|ogg)(\?|$)/i.test(url)
      
      // 如果已经包含扩展名，就不添加其他格式
      if (hasExtension) {
        return `<source src="${baseUrl}" type="video/mp4">`
      }
      
      // 支持多种格式
      return `
        <source src="${baseUrl}" type="video/mp4">
        <source src="${baseUrl}" type="video/webm">
        <source src="${baseUrl}" type="video/ogg">
      `
    },
    
    // 尝试重新加载失败的视频
    retryLoadVideo(videoEl, url) {
      if (!videoEl || !url) return
      
      try {
        // 清除现有源
        while (videoEl.firstChild) {
          videoEl.removeChild(videoEl.firstChild)
        }
        
        // 添加多格式支持
        videoEl.innerHTML = this.getVideoSourceElements(url)
        
        // 确保正确的样式和属性
        videoEl.style.zIndex = "1"
        videoEl.style.transform = "translateZ(0)"
        videoEl.style.backgroundColor = "#000"
        videoEl.preload = "metadata"
        videoEl.load()
      } catch (error) {
        console.error('重新加载视频失败:', error)
      }
    },
    
    // 获取直接视频URL
    getDirectVideoUrl(url) {
      if (!url) return ''
      const timestamp = Date.now()
      return getDirectFileUrl(url) + `&_t=${timestamp}`
    },
    
    // 对话框关闭处理
    handleClosed() {
      // 清理视频资源
      this.cleanupVideoResources()
    },
    
    // 清理视频资源
    cleanupVideoResources() {
      // 暂停主视频
      if (this.$refs.detailVideo) {
        try {
          this.$refs.detailVideo.pause()
        } catch (e) {
          console.error('暂停视频时出错:', e)
        }
      }
      
      // 暂停所有参数视频
      const videos = this.$el.querySelectorAll('video')
      videos.forEach(video => {
        try {
          if (!video.paused) {
            video.pause()
          }
        } catch (e) {
          console.error('暂停视频时出错:', e)
        }
      })
      
      // 重置加载状态
      this.detailVideoLoaded = false
      this.paramVideoLoadedMap = {}
    }
  }
}
</script>

<style scoped>
/* 详情弹窗样式 */
.task-detail {
  padding: 0 10px;
}

.detail-header {
  text-align: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
}

.detail-header h2 {
  margin: 0 0 10px;
  font-size: 22px;
  color: #303133;
}

.detail-header p {
  color: #606266;
  font-size: 14px;
  margin: 0;
}

.detail-preview {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #ebeef5;
  position: relative;
}

.detail-preview h3 {
  margin: 0 0 10px;
  font-size: 16px;
  color: #303133;
}

.full-image {
  width: 100%;
  max-height: 400px;
  object-fit: contain;
  cursor: pointer;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  background-color: #f8f8f8;
  transition: transform 0.3s ease;
}

.full-image:hover {
  transform: scale(1.02);
}

.full-video {
  width: 100%;
  max-height: 400px;
  border-radius: 4px;
  background-color: #000;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 1;
  object-fit: contain;
  filter: brightness(1.2);
}

.detail-video {
  position: relative;
  margin-bottom: 20px;
  width: 100%;
  background-color: #000;
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 2;
  transition: background-color 0.3s ease;
}

.video-placeholder:hover {
  background-color: #111;
}

.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #f8f8f8;
  transition: transform 0.3s ease;
}

.video-placeholder:hover .loading-indicator {
  transform: scale(1.1);
}

.loading-indicator i {
  font-size: 36px;
  margin-bottom: 8px;
  color: #409EFF;
}

.loading-indicator.small i {
  font-size: 24px;
  margin-bottom: 4px;
}

.detail-params {
  margin-bottom: 20px;
}

.detail-params h3 {
  margin: 0 0 10px;
  font-size: 16px;
  color: #303133;
  border-left: 4px solid #409EFF;
  padding-left: 10px;
}

.params-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 15px;
  padding: 5px;
}

.param-item {
  background-color: #f9f9f9;
  border-radius: 6px;
  padding: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.param-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.param-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
  font-weight: 500;
}

.param-image {
  text-align: center;
}

.thumbnail {
  max-width: 100%;
  max-height: 120px;
  object-fit: contain;
  cursor: pointer;
  border-radius: 4px;
  transition: transform 0.3s ease;
}

.thumbnail:hover {
  transform: scale(1.05);
}

.param-video {
  position: relative;
  min-height: 100px;
}

.thumbnail-video {
  width: 100%;
  max-height: 120px;
  border-radius: 4px;
  background-color: #000;
  z-index: 1;
  object-fit: contain;
  filter: brightness(1.2);
}

.audio-player {
  width: 100%;
}

.param-text {
  font-size: 14px;
  color: #303133;
  word-break: break-word;
}

.detail-meta {
  border-top: 1px solid #ebeef5;
  padding-top: 15px;
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  justify-content: space-between;
}

.meta-item {
  display: flex;
  align-items: center;
}

.meta-label {
  font-size: 13px;
  color: #909399;
  margin-right: 5px;
}

.meta-value {
  font-size: 13px;
  color: #606266;
}

.preview-dialog {
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-fullsize {
  max-width: 100%;
  max-height: 80vh;
  object-fit: contain;
}

/* 移动端全屏对话框样式 */
.mobile-detail-dialog .el-dialog__header {
  background-color: #1976d2;
  padding: 15px 10px;
}

.mobile-detail-dialog .el-dialog__body {
  padding: 10px 10px 15px;
}

.mobile-detail-dialog .el-dialog__headerbtn {
  top: 15px;
}

/* 响应式适配 */
@media (max-width: 576px) {
  /* 详情页移动端适配 */
  .params-grid {
    grid-template-columns: 1fr;
  }
  
  .detail-meta {
    flex-direction: column;
    gap: 10px;
  }
  
  .task-detail {
    padding: 0 5px;
  }
  
  .detail-header h2 {
    font-size: 18px;
  }
  
  .detail-header p {
    font-size: 12px;
  }
  
  .detail-preview h3,
  .detail-params h3 {
    font-size: 14px;
  }
  
  .full-video, .full-image {
    max-height: 250px;
  }
  
  .param-item {
    padding: 10px;
    margin-bottom: 5px;
  }
  
  .param-item:hover {
    transform: translateY(-1px);
  }
  
  .thumbnail {
    max-height: 100px;
  }
  
  .thumbnail-video {
    max-height: 100px;
  }
  
  .loading-indicator i {
    font-size: 28px;
    margin-bottom: 5px;
  }
  
  .loading-indicator.small i {
    font-size: 20px;
    margin-bottom: 3px;
  }
  
  .loading-indicator span {
    font-size: 12px;
  }
  
  .loading-indicator.small span {
    font-size: 10px;
  }
  
  .meta-label, .meta-value {
    font-size: 12px;
  }
  
  /* 确保视频播放控件在移动端足够大 */
  video::-webkit-media-controls-panel {
    opacity: 1 !important;
  }
}
</style> 