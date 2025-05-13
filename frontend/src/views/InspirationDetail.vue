<template>
  <div class="inspiration-detail-container">
    <!-- 标题 -->
    <div class="page-header" :class="{'mobile-header': isMobile}">
      <div class="header-left">
        <i class="el-icon-arrow-left back-icon" @click="goBack"></i>
        <h2>灵感详情</h2>
      </div>
      <div class="header-right">
        <p v-if="!isMobile">查看创作详情内容</p>
        <i class="el-icon-refresh refresh-icon" @click="refreshContent" :class="{'is-loading': loading}"></i>
      </div>
    </div>

    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>

    <!-- 加载动画 -->
    <div v-if="loading" class="loading-container">
      <div class="loading-animation">
        <i class="el-icon-loading"></i>
        <p>正在加载内容...</p>
      </div>
      <el-skeleton :rows="5" animated />
    </div>

    <!-- 没有内容时显示 -->
    <el-empty v-else-if="!task" description="未找到创作内容">
      <el-button type="primary" size="small" @click="goBack">返回列表</el-button>
    </el-empty>

    <!-- 详情内容 -->
    <div v-else class="task-detail">
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

      <!-- 回到顶部按钮 -->
      <el-backtop :visibility-height="200" :right="40" :bottom="80">
        <div class="back-to-top">
          <i class="el-icon-caret-top"></i>
        </div>
      </el-backtop>
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
  </div>
</template>

<script>
import { getDirectFileUrl } from '@/utils/fileAccess'
import axios from 'axios'

export default {
  name: 'InspirationDetail',
  data() {
    return {
      task: null,
      loading: true,
      detailVideoLoaded: false,
      paramVideoLoadedMap: {},
      previewVisible: false,
      previewUrl: '',
      isMobile: false,
      taskCache: {},
      retryCount: 0
    }
  },
  created() {
    this.fetchTaskDetail()
    this.checkMobileDevice()
  },
  mounted() {
    window.addEventListener('resize', this.checkMobileDevice)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkMobileDevice)
    this.cleanupVideoResources()
  },
  methods: {
    async fetchTaskDetail() {
      this.loading = true
      this.retryCount = 0; // 重置重试计数
      
      try {
        const taskId = this.$route.params.id
        if (!taskId) {
          this.loading = false
          return
        }
        
        // 首先检查缓存中是否已有该任务
        if (this.taskCache[taskId]) {
          console.log('使用缓存的任务数据:', taskId)
          this.task = this.taskCache[taskId]
          this.loading = false
          return
        }
        
        // 使用专门的详情接口
        console.log('正在请求任务详情，ID:', taskId)
        const response = await axios.get(`/api/inspiration/${taskId}`)
        this.task = response.data
        
        // 将任务添加到缓存
        this.taskCache[taskId] = response.data
        
        console.log('获取任务详情成功:', this.task)
      } catch (error) {
        console.error('获取任务详情失败:', error.response && error.response.status || '未知错误', error.message)
        this.$message.error('获取任务详情失败')
        
        // 如果是404错误，表示任务不存在
        if (error.response && error.response.status === 404) {
          this.$message.error('未找到该创作内容')
          setTimeout(() => this.goBack(), 1500) // 1.5秒后返回
        } 
        // 如果未达到最大重试次数，则重试
        else if (this.retryCount < 3) {
          this.retryCount++
          setTimeout(() => {
            console.log(`第${this.retryCount}次重试获取任务详情`)
            this.fetchTaskDetail()
          }, 1000 * this.retryCount) // 重试间隔递增
        } else {
          setTimeout(() => this.goBack(), 1500) // 重试失败后返回列表页
        }
      } finally {
        this.loading = false
      }
    },
    
    goBack() {
      try {
        window.history.length > 1 ? this.$router.go(-1) : this.$router.push('/inspiration')
      } catch (e) {
        console.error('返回上一页失败，跳转到灵感页', e)
        this.$router.push('/inspiration')
      }
    },
    
    checkMobileDevice() {
      const userAgent = navigator.userAgent || navigator.vendor || window.opera
      const isMobileByUA = /android|webos|iphone|ipad|ipod|blackberry|iemobile|opera mini|mobile/i.test(userAgent)
      const isMobileByWidth = window.innerWidth <= 576
      this.isMobile = isMobileByUA || isMobileByWidth
    },
    
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
    
    loadDetailVideo() {
      if (this.detailVideoLoaded) return
      
      this.detailVideoLoaded = true
      
      if (this.$refs.detailVideo) {
        try {
          this.$refs.detailVideo.style.zIndex = "1"
          this.$refs.detailVideo.style.transform = "translateZ(0)"
          this.$refs.detailVideo.style.backgroundColor = "#000"
          this.$refs.detailVideo.preload = "metadata"
          
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
    
    handleDetailVideoCanPlay() {
      this.detailVideoLoaded = true
    },
    
    handleDetailVideoError(event) {
      console.error('详情视频加载失败:', event)
      this.detailVideoLoaded = false
      
      if (this.task && this.task.result_url) {
        setTimeout(() => {
          this.retryLoadVideo(this.$refs.detailVideo, this.task.result_url)
        }, 1000)
      }
    },
    
    loadParamVideo(event, url) {
      if (this.paramVideoLoadedMap[url]) return
      
      this.$set(this.paramVideoLoadedMap, url, true)
      
      event.stopPropagation()
      
      const videoEl = event.target
      if (videoEl && videoEl.tagName === 'VIDEO') {
        try {
          videoEl.style.zIndex = "1"
          videoEl.style.transform = "translateZ(0)"
          videoEl.style.backgroundColor = "#000"
          videoEl.preload = "metadata"
          
          videoEl.innerHTML = this.getVideoSourceElements(url)
          
          videoEl.load()
        } catch (error) {
          console.error('加载参数视频失败:', error)
          this.$set(this.paramVideoLoadedMap, url, false)
        }
      }
    },
    
    handleParamVideoCanPlay(url) {
      this.$set(this.paramVideoLoadedMap, url, true)
    },
    
    handleParamVideoError(event, url) {
      console.error('参数视频加载失败:', url, event)
      this.$set(this.paramVideoLoadedMap, url, false)
      
      if (url) {
        setTimeout(() => {
          const videoEl = event.target.closest('video')
          if (videoEl) {
            this.retryLoadVideo(videoEl, url)
          }
        }, 1000)
      }
    },
    
    formatDate(timeStr) {
      const date = new Date(timeStr)
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    },
    
    previewImage(url) {
      this.previewUrl = url
      this.previewVisible = true
    },
    
    getVideoSourceElements(url) {
      if (!url) return ''
      
      const baseUrl = this.getDirectVideoUrl(url)
      const hasExtension = /\.(mp4|webm|mov|ogg)(\?|$)/i.test(url)
      
      if (hasExtension) {
        return `<source src="${baseUrl}" type="video/mp4">`
      }
      
      return `
        <source src="${baseUrl}" type="video/mp4">
        <source src="${baseUrl}" type="video/webm">
        <source src="${baseUrl}" type="video/ogg">
      `
    },
    
    retryLoadVideo(videoEl, url) {
      if (!videoEl || !url) return
      
      try {
        while (videoEl.firstChild) {
          videoEl.removeChild(videoEl.firstChild)
        }
        
        videoEl.innerHTML = this.getVideoSourceElements(url)
        
        videoEl.style.zIndex = "1"
        videoEl.style.transform = "translateZ(0)"
        videoEl.style.backgroundColor = "#000"
        videoEl.preload = "metadata"
        videoEl.load()
      } catch (error) {
        console.error('重新加载视频失败:', error)
      }
    },
    
    getDirectVideoUrl(url) {
      if (!url) return ''
      const timestamp = Date.now()
      return getDirectFileUrl(url) + `&_t=${timestamp}`
    },
    
    cleanupVideoResources() {
      if (this.$refs.detailVideo) {
        try {
          this.$refs.detailVideo.pause()
        } catch (e) {
          console.error('暂停视频时出错:', e)
        }
      }
      
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
      
      this.detailVideoLoaded = false
      this.paramVideoLoadedMap = {}
    },
    
    // 刷新内容
    refreshContent() {
      if (this.loading) return;
      
      const taskId = this.$route.params.id;
      if (taskId) {
        // 从缓存中删除当前任务
        if (this.taskCache[taskId]) {
          delete this.taskCache[taskId];
        }
        
        // 重置视频加载状态
        this.detailVideoLoaded = false;
        this.paramVideoLoadedMap = {};
        
        // 重置重试计数
        this.retryCount = 0;
        
        // 重新获取任务详情
        this.fetchTaskDetail();
        
        this.$message({
          message: '正在刷新内容...',
          type: 'info',
          duration: 1000
        });
      }
    }
  },
  watch: {
    '$route.params.id': {
      handler(newId, oldId) {
        if (newId && newId !== oldId) {
          console.log('路由参数变化，重新获取任务详情:', newId)
          this.fetchTaskDetail()
        }
      },
      immediate: false
    }
  }
}
</script>

<style scoped>
.inspiration-detail-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

/* 页面标题样式 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
}

.header-left {
  display: flex;
  align-items: center;
}

.back-icon {
  font-size: 20px;
  margin-right: 15px;
  cursor: pointer;
  color: #409EFF;
}

.back-icon:hover {
  color: #66b1ff;
}

.page-header h2 {
  margin: 0;
  font-size: 22px;
  color: #303133;
}

.header-right p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

/* 移动端标题样式 */
.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  padding: 10px 15px;
  border-bottom: 1px solid #ebeef5;
}

.mobile-header h2 {
  font-size: 18px;
}

.mobile-header-placeholder {
  height: 50px;
  margin-bottom: 10px;
}

/* 加载动画容器 */
.loading-container {
  padding: 20px;
}

/* 详情页样式 */
.task-detail {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  padding: 20px;
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

/* 预览区域 */
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

/* 参数展示 */
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

/* 回到顶部按钮样式 */
.back-to-top {
  height: 100%;
  width: 100%;
  background-color: #409EFF;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

/* 添加加载动画样式 */
.loading-animation {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin: 20px 0;
}

.loading-animation i {
  font-size: 32px;
  color: #409EFF;
  margin-bottom: 10px;
}

.loading-animation p {
  color: #606266;
  font-size: 14px;
}

/* 响应式适配 */
@media (max-width: 576px) {
  .inspiration-detail-container {
    padding: 10px 5px;
  }

  .task-detail {
    padding: 15px 10px;
    border-radius: 0;
    box-shadow: none;
  }
  
  .mobile-header {
    padding: 10px;
  }
  
  .back-icon {
    margin-right: 10px;
  }
  
  .mobile-header h2 {
    font-size: 16px;
  }
  
  .mobile-header-placeholder {
    height: 45px;
  }
  
  .params-grid {
    grid-template-columns: 1fr;
  }
  
  .detail-meta {
    flex-direction: column;
    gap: 10px;
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
  }
  
  .meta-label, .meta-value {
    font-size: 12px;
  }
}

.refresh-icon {
  font-size: 20px;
  color: #409EFF;
  cursor: pointer;
  margin-left: 15px;
  transition: transform 0.3s ease;
}

.refresh-icon:hover {
  color: #66b1ff;
  transform: rotate(90deg);
}

.refresh-icon.is-loading {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style> 