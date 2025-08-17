<template>
  <div class="digital-human-detail-container">
    <!-- PC端顶部导航栏 -->
    <div class="page-header">
      <h2>数字人模版详情</h2>
      <div>
        <el-button type="primary" class="action-button" @click="goBack">返回列表</el-button>
      </div>
    </div>
    <!-- 移动端顶部导航栏 -->
    <!-- 移动端顶部导航栏 - 只在移动端显示 -->
    <div class="mobile-header-bar" v-show="isMobile">
      <div class="header-back" @click="goBack">
        <i class="el-icon-arrow-left"></i>
        <span>返回</span>
      </div>
      <h2 class="header-title">数字人模版详情</h2>
    </div>
    
    <!-- 移动端头部占位 - 只在移动端显示 -->
    <div class="mobile-header-placeholder" v-show="isMobile"></div>
    <div class="detail-content-wrapper">
      <div v-loading="loading" class="detail-content">
        <!-- PC端展示 -->
        <div class="desktop-content-view">
          <el-card v-if="template">
            <div slot="header" class="card-header">
              <span style="color: #303133; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; font-weight: 500;">{{ template.name }}</span>
              <el-tag :type="getStatusType(template.status)" class="status-tag">{{ getStatusText(template.status) }}</el-tag>
            </div>
            <div class="task-info">
              <div class="info-item">
                <span class="label">创建时间：</span>
                <span>{{ formatDate(template.created_at) }}</span>
              </div>
              <div class="info-item" v-if="template.description">
                <span class="label">模版描述：</span>
                <span>{{ template.description }}</span>
              </div>
            </div>
            <div class="media-section">
              <h3 style="color: #303133; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">输入文件</h3>
              <div class="media-preview">
                <div class="video-preview">
                  <div v-if="template.video_url" class="video-container">
                    <video
                      controls
                      style="width: 100%; max-height: 300px; background-color: #000; z-index: 0 !important; transform: translateZ(0);"
                      class="desktop-video"
                      muted
                      playsinline
                      preload="metadata"
                      :src="template.video_url"></video>
                  </div>
                  <div v-else class="video-placeholder">
                    <el-empty description="无原始视频"/>
                  </div>
                </div>
                <div class="image-preview">
                  <h4>人脸图片</h4>
                  <div v-if="template.face_image_url">
                    <img :src="template.face_image_url" alt="人脸图片" style="max-width:120px;max-height:120px;display:block;margin:8px 0;"/>
                    <a :href="template.face_image_url" target="_blank">下载/查看</a>
                  </div>
                  <div v-else><el-empty description="无"/></div>
                </div>
                <div class="image-preview">
                  <h4>背景</h4>
                  <div v-if="template.background_url">
                    <video v-if="isVideo(template.background_url)" :src="template.background_url" controls style="max-width:300px;max-height:180px;display:block;margin:8px 0;"/>
                    <img v-else :src="template.background_url" alt="背景" style="max-width:120px;max-height:120px;display:block;margin:8px 0;"/>
                    <a :href="template.background_url" target="_blank">下载/查看</a>
                  </div>
                  <div v-else><el-empty description="无"/></div>
                </div>
              </div>
            </div>
            
            <!-- 处理结果展示 -->
            <div class="result-section" v-if="template.status === 'completed' || template.status === 'processing'">
              <h3 style="color: #303133; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">处理结果</h3>
              <div class="result-preview">
                <!-- 换脸视频结果 -->
                <div class="result-item" v-if="template.replace_face_video_url">
                  <h4>换脸视频</h4>
                  <div class="video-container">
                    <video
                      controls
                      style="width: 100%; max-height: 300px; background-color: #000; z-index: 0 !important; transform: translateZ(0);"
                      class="desktop-video"
                      muted
                      playsinline
                      preload="metadata"
                      :src="template.replace_face_video_url"></video>
                  </div>
                  <div class="result-actions">
                    <a :href="template.replace_face_video_url" target="_blank" class="download-link">
                      <el-button type="primary" size="small">下载视频</el-button>
                    </a>
                  </div>
                </div>
                
                <!-- 换背景视频结果 -->
                <div class="result-item" v-if="template.replace_background_video_url">
                  <h4>换背景视频</h4>
                  <div class="video-container">
                    <video
                      controls
                      style="width: 100%; max-height: 300px; background-color: #000; z-index: 0 !important; transform: translateZ(0);"
                      class="desktop-video"
                      muted
                      playsinline
                      preload="metadata"
                      :src="template.replace_background_video_url"></video>
                  </div>
                  <div class="result-actions">
                    <a :href="template.replace_background_video_url" target="_blank" class="download-link">
                      <el-button type="primary" size="small">下载视频</el-button>
                    </a>
                  </div>
                </div>
                
                <!-- 最终合成视频结果 -->
                <div class="result-item" v-if="template.result_video_url">
                  <h4>最终合成视频</h4>
                  <div class="video-container">
                    <video
                      controls
                      style="width: 100%; max-height: 300px; background-color: #000; z-index: 0 !important; transform: translateZ(0);"
                      class="desktop-video"
                      muted
                      playsinline
                      preload="metadata"
                      :src="template.result_video_url"></video>
                  </div>
                  <div class="result-actions">
                    <a :href="template.result_video_url" target="_blank" class="download-link">
                      <el-button type="success" size="small">下载最终视频</el-button>
                    </a>
                  </div>
                </div>
                
                <!-- 处理中状态 -->
                <div v-if="template.status === 'processing'" class="processing-status">
                  <el-alert
                    title="处理中"
                    type="info"
                    :closable="false"
                    show-icon>
                    <div slot="description">
                      模版正在处理中，请稍后查看结果...
                    </div>
                  </el-alert>
                </div>
                
                <!-- 处理失败状态 -->
                <div v-if="template.status === 'failed'" class="error-status">
                  <el-alert
                    title="处理失败"
                    type="error"
                    :closable="false"
                    show-icon>
                    <div slot="description">
                      <p>处理过程中出现错误：</p>
                      <p class="error-message">{{ template.error_msg || '未知错误' }}</p>
                    </div>
                  </el-alert>
                </div>
              </div>
            </div>
          </el-card>
          <el-empty v-else description="未找到模版信息"></el-empty>
        </div>
        <!-- 移动端展示 -->
        <div class="mobile-content-view">
          <div v-if="template" class="mobile-content-inner">
            <div class="basic-info">
              <div class="status-tag">
                <el-tag :type="getStatusType(template.status)">{{ getStatusText(template.status) }}</el-tag>
              </div>
              <div class="create-time">创建时间：{{ formatDate(template.created_at) }}</div>
            </div>
            <div class="mobile-task-info">
              <div class="info-item" v-if="template.description">
                <span class="label">模版描述：</span>
                <span>{{ template.description }}</span>
              </div>
            </div>
            <div class="mobile-section">
              <h4 class="section-title" style="color: #303133; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">原始视频</h4>
              <div v-if="template.video_url" class="video-player">
                <video
                  controls
                  style="width: 100%; max-height: 200px; background-color: #000; display: block; z-index: 0 !important; transform: translateZ(0);"
                  class="mobile-video"
                  muted
                  playsinline
                  webkit-playsinline
                  :src="template.video_url"></video>
              </div>
              <div v-else class="video-placeholder">
                <el-empty description="无原始视频"/>
              </div>
            </div>
            <div class="mobile-section">
              <h4 class="section-title" style="color: #303133; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">人脸图片</h4>
              <div v-if="template.face_image_url">
                <img :src="template.face_image_url" alt="人脸图片" style="max-width:120px;max-height:120px;display:block;margin:8px 0;"/>
                <a :href="template.face_image_url" target="_blank">下载/查看</a>
              </div>
              <div v-else><el-empty description="无"/></div>
            </div>
            <div class="mobile-section">
              <h4 class="section-title" style="color: #303133; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">背景</h4>
              <div v-if="template.background_url">
                <video v-if="isVideo(template.background_url)" :src="template.background_url" controls style="max-width:300px;max-height:180px;display:block;margin:8px 0;"/>
                <img v-else :src="template.background_url" alt="背景" style="max-width:120px;max-height:120px;display:block;margin:8px 0;"/>
                <a :href="template.background_url" target="_blank">下载/查看</a>
              </div>
              <div v-else><el-empty description="无"/></div>
            </div>
            
            <!-- 移动端处理结果展示 -->
            <div class="mobile-result-section" v-if="template.status === 'completed' || template.status === 'processing'">
              <h4 class="section-title" style="color: #303133; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">处理结果</h4>
              
              <!-- 换脸视频结果 -->
              <div v-if="template.replace_face_video_url" class="mobile-result-item">
                <h5>换脸视频</h5>
                <div class="video-player">
                  <video
                    controls
                    style="width: 100%; max-height: 200px; background-color: #000; display: block; z-index: 0 !important; transform: translateZ(0);"
                    class="mobile-video"
                    muted
                    playsinline
                    webkit-playsinline
                    :src="template.replace_face_video_url"></video>
                </div>
                <div class="mobile-result-actions">
                  <a :href="template.replace_face_video_url" target="_blank">
                    <el-button type="primary" size="small">下载视频</el-button>
                  </a>
                </div>
              </div>
              
              <!-- 换背景视频结果 -->
              <div v-if="template.replace_background_video_url" class="mobile-result-item">
                <h5>换背景视频</h5>
                <div class="video-player">
                  <video
                    controls
                    style="width: 100%; max-height: 200px; background-color: #000; display: block; z-index: 0 !important; transform: translateZ(0);"
                    class="mobile-video"
                    muted
                    playsinline
                    webkit-playsinline
                    :src="template.replace_background_video_url"></video>
                </div>
                <div class="mobile-result-actions">
                  <a :href="template.replace_background_video_url" target="_blank">
                    <el-button type="primary" size="small">下载视频</el-button>
                  </a>
                </div>
              </div>
              
              <!-- 最终合成视频结果 -->
              <div v-if="template.result_video_url" class="mobile-result-item">
                <h5>最终合成视频</h5>
                <div class="video-player">
                  <video
                    controls
                    style="width: 100%; max-height: 200px; background-color: #000; display: block; z-index: 0 !important; transform: translateZ(0);"
                    class="mobile-video"
                    muted
                    playsinline
                    webkit-playsinline
                    :src="template.result_video_url"></video>
                </div>
                <div class="mobile-result-actions">
                  <a :href="template.result_video_url" target="_blank">
                    <el-button type="success" size="small">下载最终视频</el-button>
                  </a>
                </div>
              </div>
              
              <!-- 处理中状态 -->
              <div v-if="template.status === 'processing'" class="mobile-processing-status">
                <el-alert
                  title="处理中"
                  type="info"
                  :closable="false"
                  show-icon>
                  <div slot="description">
                    模版正在处理中，请稍后查看结果...
                  </div>
                </el-alert>
              </div>
              
              <!-- 处理失败状态 -->
              <div v-if="template.status === 'failed'" class="mobile-error-status">
                <el-alert
                  title="处理失败"
                  type="error"
                  :closable="false"
                  show-icon>
                  <div slot="description">
                    <p>处理过程中出现错误：</p>
                    <p class="error-message">{{ template.error_msg || '未知错误' }}</p>
                  </div>
                </el-alert>
              </div>
            </div>
          </div>
          <el-empty v-else description="未找到模版信息"></el-empty>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  name: 'DigitalHumanTemplateDetail',
  data() {
    return {
      loading: false,
      template: null,
      isMobile: false
    }
  },
  methods: {
    fetchDetail() {
      this.loading = true
      const id = this.$route.params.id
      this.$http.get(`/api/digital-human-template/${id}`)
        .then(res => {
          this.template = res.data.template || null
        })
        .finally(() => { this.loading = false })
    },
    formatDate(dateString) {
      if (!dateString) return ''
      return new Date(dateString).toLocaleString()
    },
    isVideo(url) {
      return url && url.toLowerCase().endsWith('.mp4')
    },
    goBack() {
      if (window.history.length > 1) this.$router.back()
      else this.$router.push('/digital-human-template')
    },
    getStatusType(status) {
      switch (status) {
        case 'pending': return 'info'
        case 'processing': return 'warning'
        case 'completed': return 'success'
        case 'failed': return 'danger'
        default: return 'info'
      }
    },
    getStatusText(status) {
      switch (status) {
        case 'pending': return '等待处理'
        case 'processing': return '处理中'
        case 'completed': return '已完成'
        case 'failed': return '处理失败'
        default: return status
      }
    },
    checkDeviceType() {
      this.isMobile = window.innerWidth <= 768 || /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
    }
  },
  mounted() {
    this.checkDeviceType()
    window.addEventListener('resize', this.checkDeviceType)
    this.fetchDetail()
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkDeviceType)
  }
}
</script>
<style>
/* 按钮样式与DigitalHumanDetail.vue保持一致 */
.action-button {
  padding: 12px 24px;
  font-weight: 600;
  border-radius: 8px;
  transition: all 0.3s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.action-button.el-button--primary {
  background: linear-gradient(90deg, #2c3e50, #4a6572);
  border: none;
  box-shadow: 0 5px 15px rgba(44, 62, 80, 0.2);
}

.action-button.el-button--primary:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(44, 62, 80, 0.3);
}

.action-button.secondary {
  background-color: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(0, 0, 0, 0.1);
  color: #2c3e50;
}

.action-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px rgba(0, 0, 0, 0.1);
}

.action-button.secondary:hover {
  background-color: rgba(255, 255, 255, 0.95);
  border-color: rgba(0, 0, 0, 0.15);
}

.action-button i {
  font-size: 16px;
}

.action-button:focus {
  outline: 2px solid rgba(44, 62, 80, 0.4);
  outline-offset: 2px;
}

/* 桌面端卡片悬浮效果 */
@media screen and (min-width: 769px) {
  .el-card {
    transition: all 0.4s cubic-bezier(0.165, 0.84, 0.44, 1);
  }
  
  .el-card:hover {
    transform: translateY(-5px);
    box-shadow: 0 14px 46px rgba(0, 0, 0, 0.28);
  }
}

/* 处理结果样式 */
.result-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.result-section h3 {
  margin-bottom: 16px;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.result-preview {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.result-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  background-color: #fafafa;
}

.result-item h4 {
  margin: 0 0 12px 0;
  color: #606266;
  font-size: 16px;
  font-weight: 500;
}

.result-actions {
  margin-top: 12px;
  text-align: center;
}

.download-link {
  text-decoration: none;
}

.processing-status,
.error-status {
  margin-top: 16px;
}

.error-message {
  color: #f56c6c;
  font-size: 14px;
  margin-top: 8px;
  word-break: break-all;
}

/* 移动端处理结果样式 */
.mobile-result-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.mobile-result-item {
  margin-bottom: 20px;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background-color: #fafafa;
}

.mobile-result-item h5 {
  margin: 0 0 8px 0;
  color: #606266;
  font-size: 14px;
  font-weight: 500;
}

.mobile-result-actions {
  margin-top: 8px;
  text-align: center;
}

.mobile-processing-status,
.mobile-error-status {
  margin-top: 12px;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .result-item {
    padding: 12px;
  }
  
  .result-item h4 {
    font-size: 14px;
  }
  
  .mobile-result-item {
    padding: 10px;
  }
  
  .mobile-result-item h5 {
    font-size: 13px;
  }
}

/* 全局样式覆盖，强制禁止水平滚动 */
.digital-human-detail-container {
  padding: 20px;
  min-height: 100vh;
  color: rgba(44, 62, 80, 0.9);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  transform: translateZ(0); /* 创建新的堆叠上下文 */
  isolation: isolate; /* 现代浏览器隔离堆叠上下文 */
  position: relative;
  z-index: 100 !important; /* 提高详情页整体层级 */
  overflow-x: hidden !important;
  width: 100% !important;
  max-width: 100% !important;
  box-sizing: border-box;
}

/* 强制所有视频元素不能超出其父容器的堆叠上下文 */
.digital-human-detail-container video,
.digital-human-detail-container audio,
.digital-human-detail-container .video-container,
.digital-human-detail-container .video-preview,
.digital-human-detail-container .result-preview,
.digital-human-detail-container .video-player {
  transform: translateZ(0);
  position: relative !important;
  z-index: 0 !important; /* 使用固定的低值而不是auto */
  isolation: isolate;
  max-width: 100%;
  max-height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.page-header h2 {
  margin: 0;
  color: #303133;
  font-size: 20px;
  font-weight: 600;
}

.detail-content-wrapper {
  width: 100%;
  margin: 0;
}

.detail-content {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.desktop-content-view {
  display: block;
}

.mobile-content-view {
  display: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-tag {
  margin-left: 8px;
}

.task-info {
  margin-bottom: 20px;
}

.info-item {
  margin-bottom: 8px;
  display: flex;
  align-items: flex-start;
}

.info-item .label {
  font-weight: 500;
  color: #606266;
  min-width: 100px;
  margin-right: 8px;
}

.media-section {
  margin-top: 20px;
}

.media-section h3 {
  margin-bottom: 16px;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.media-preview {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 20px;
}

.video-preview,
.image-preview {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  background-color: #fafafa;
}

.video-preview h4,
.image-preview h4 {
  margin: 0 0 12px 0;
  color: #606266;
  font-size: 16px;
  font-weight: 500;
}

.video-container {
  position: relative;
  border-radius: 4px;
  overflow: hidden;
}

.video-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  background-color: #f5f7fa;
  border: 2px dashed #dcdfe6;
  border-radius: 4px;
}

/* 移动端样式 */
/* 移动端顶部导航栏 - 提高层级 */
.mobile-header-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background: linear-gradient(135deg, #2c3e50, #4a6572);
  display: flex;
  align-items: center;
  padding: 0 12px;
  z-index: 2000 !important; /* 确保导航栏在最上层 */
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  color: #fff;
  width: 100%;
  box-sizing: border-box;
  display: none; /* 默认隐藏，在移动端通过v-show显示 */
}

.mobile-header-placeholder {
  display: none;
  height: 60px;
}

/* 移动端返回按钮样式增强 */
.header-back {
  display: flex;
  align-items: center;
  font-size: 16px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
  z-index: 2100 !important; /* 确保返回按钮在最顶层 */
  position: relative;
  color: white;
}

.header-back:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.header-back i {
  margin-right: 4px;
}

.header-title {
  margin: 0 0 0 10px;
  font-size: 16px;
  font-weight: 500;
  color: white;
  flex: 1;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mobile-content-inner {
  padding: 16px;
}

.basic-info {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;
}

.create-time {
  margin-top: 8px;
  color: #909399;
  font-size: 14px;
}

.mobile-task-info {
  margin-bottom: 20px;
}

.mobile-section {
  margin-bottom: 24px;
}

.section-title {
  margin: 0 0 12px 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.video-player {
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 8px;
  position: relative;
  background-color: #000;
}

/* 视频预览和结果视频特殊样式 */
.video-preview, .result-preview {
  border: 2px solid #ebeef5;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
  transition: all 0.3s ease;
  position: relative;
}

.video-preview:before {
  content: "原始视频";
  position: absolute;
  top: -10px;
  left: 10px;
  background-color: rgba(64, 158, 255, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #409EFF;
  z-index: 1;
  font-weight: bold;
}

.result-preview:before {
  content: "结果视频";
  position: absolute;
  top: -10px;
  left: 10px;
  background-color: rgba(103, 194, 58, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #67c23a;
  z-index: 1;
  font-weight: bold;
}

.video-loaded {
  border-color: #67c23a;
  box-shadow: 0 0 10px rgba(103, 194, 58, 0.2);
}

.video-loading {
  border-color: #e6a23c;
  box-shadow: 0 0 10px rgba(230, 162, 60, 0.2);
}

/* 移动端视频容器增强 */
.mobile-video-player {
  position: relative;
  overflow: hidden;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  margin: 10px 0;
  background-color: #000;
}

.mobile-video-player video {
  display: block;
  width: 100%;
}

/* 特殊边框样式 */
#mobile-original-video {
  border-left: 4px solid #409EFF;
  position: relative;
}

#mobile-original-video:before {
  content: "";
  position: absolute;
  left: -4px;
  top: 0;
  bottom: 0;
  width: 4px;
  background-color: #409EFF;
}

#mobile-result-video {
  border-left: 4px solid #67c23a;
  position: relative;
}

#mobile-result-video:before {
  content: "";
  position: absolute;
  left: -4px;
  top: 0;
  bottom: 0;
  width: 4px;
  background-color: #67c23a;
}

/* 处理中和等待状态样式 */
.processing-info, .waiting-info {
  padding: 20px;
  text-align: center;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.processing-info {
  border-left: 4px solid #e6a23c;
}

.waiting-info {
  border-left: 4px solid #909399;
  color: #909399;
}

/* 视频容器和占位符样式 */
.video-container {
  width: 100%;
  background-color: #000;
  border-radius: 4px;
  overflow: hidden;
}

.video-placeholder {
  width: 100%;
  position: relative;
  min-height: 120px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.placeholder-text {
  text-align: center;
  color: #909399;
  margin-top: 10px;
  font-size: 14px;
}

/* 结果占位样式 */
.result-placeholder {
  width: 100%;
  min-height: 160px;
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 6px;
  margin-bottom: 10px;
}

.empty-result {
  min-height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f9f9f9;
  padding: 20px;
  border-radius: 6px;
}

.has-result {
  border: 2px solid #67c23a;
}

.processing {
  border-left: 4px solid #e6a23c;
}

/* PC端下载区域样式 */
.download-section {
  margin-top: 40px;
  margin-bottom: 30px;
  border-top: 1px solid #ebeef5;
  padding-top: 30px;
}

.download-container {
  background-color: #f0f9ff;
  border: 1px solid #a0cfff;
  border-radius: 8px;
  padding: 25px 30px;
  text-align: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.download-container:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.1);
}

.download-container h3 {
  color: #2c3e50;
  font-size: 20px;
  margin-bottom: 15px;
  font-weight: 600;
}

.download-tips {
  color: #606266;
  margin-bottom: 20px;
  font-size: 15px;
  line-height: 1.5;
}

.download-button {
  padding: 12px 25px;
  font-size: 16px;
  border-radius: 6px;
  transition: all 0.3s;
  min-width: 200px;
}

.download-button:hover {
  transform: scale(1.05);
  background-color: #66b1ff;
  border-color: #66b1ff;
}

.download-button i {
  margin-right: 8px;
  font-size: 18px;
}

/* 移动端下载区域样式 */
.mobile-download-section {
  margin-top: 30px;
  margin-bottom: 60px;
  padding: 0 12px;
}

.mobile-download-container {
  background: linear-gradient(135deg, #f1f3f6 0%, #ffffff 100%);
  border-radius: 12px;
  padding: 20px 15px;
  text-align: center;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.08);
  border: 1px solid #d9ecff;
}

.download-icon {
  font-size: 32px;
  color: #409EFF;
  margin-bottom: 10px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
}

/* 动作按钮组样式 */
.action-buttons-group {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

/* 添加视频错误状态 */
.video-error {
  border: 2px solid #f56c6c;
  background-color: rgba(245, 108, 108, 0.1);
  padding: 10px;
  text-align: center;
  color: #f56c6c;
  margin: 10px 0;
  border-radius: 4px;
}

/* 视频控制增强 */
video::-webkit-media-controls {
  display: flex !important;
  visibility: visible !important;
  opacity: 1 !important;
}

/* 视频播放按钮样式增强 */
video::-webkit-media-controls-play-button {
  display: flex !important;
  visibility: visible !important;
  opacity: 1 !important;
  background-color: rgba(255, 255, 255, 0.7);
  border-radius: 50%;
  width: 40px;
  height: 40px;
}

.desktop-video {
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.mobile-video {
  display: block;
  width: 100%;
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

/* 强化移动端样式 */
@media (max-width: 768px) {
  /* 数字人详情页容器移动端样式 */
  .digital-human-detail-container {
    padding: 0;
    width: 100%;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 5 !important; /* 低于全局菜单层级 */
    background-color: #fff;
    color: #333;
    height: 100vh;
    overflow-x: hidden;
    box-sizing: border-box;
    transform: translateZ(0);
    isolation: isolate;
  }
  
  /* 隐藏PC端，显示移动端 */
  .page-header {
    display: none;
  }
  
  .mobile-header-bar {
    z-index: 10 !important; /* 高于内容，低于应用菜单 */
    display: flex !important; /* 强制在移动端显示，覆盖v-show的影响 */
  }
  
  .mobile-header-placeholder {
    display: block;
  }
  
  .desktop-content-view {
    display: none;
  }
  
  .mobile-content-view {
    transform: translateZ(0);
    position: relative;
    isolation: isolate;
    z-index: 1 !important;
    display: block;
  }
  
  .detail-content-wrapper {
    padding: 0;
    width: 100%;
    position: absolute;
    top: 56px;
    left: 0;
    right: 0;
    bottom: 0;
    box-sizing: border-box;
    background-color: #ffffff;
    height: calc(100vh - 56px);
    overflow-y: auto; /* 允许垂直滚动 */
    overflow-x: hidden; /* 禁止水平滚动 */
    transform: translateZ(0);
    isolation: isolate;
    z-index: 1 !important;
  }
  
  .detail-content {
    box-shadow: none;
    border-radius: 0;
    width: 100%;
    margin: 0;
    background: rgba(255, 255, 255, 0.85);
    backdrop-filter: blur(15px);
    border-radius: 16px;
    border: 1px solid rgba(200, 200, 200, 0.4);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.06);
    min-height: 100%;
    overflow-x: hidden;
    box-sizing: border-box;
    padding: 0;
    position: relative;
    z-index: auto !important;
  }
  
  .media-preview {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  /* 移动端视频容器增强 */
  .mobile-section .video-player {
    position: relative;
    overflow: hidden;
    border-radius: 8px;
    margin-top: 10px;
    background-color: #000;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  }
  
  /* 确保元素不溢出 */
  .el-tag {
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 12px;
    padding: 0 5px;
    height: 22px;
    line-height: 20px;
  }
}
</style> 