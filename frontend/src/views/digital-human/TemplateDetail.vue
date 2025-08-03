<template>
  <div class="digital-human-detail-container">
    <!-- PC端顶部导航栏 -->
    <div class="page-header">
      <h2>数字人模版详情</h2>
      <div>
        <el-button type="primary" @click="goBack">返回列表</el-button>
      </div>
    </div>
    <!-- 移动端顶部导航栏 -->
    <div class="mobile-header-bar" v-show="isMobile">
      <div class="header-back" @click="goBack">
        <i class="el-icon-arrow-left"></i>
        <span>返回</span>
      </div>
      <h2 class="header-title">数字人模版</h2>
    </div>
    <div class="mobile-header-placeholder" v-show="isMobile"></div>
    <div class="detail-content-wrapper">
      <div v-loading="loading" class="detail-content">
        <!-- PC端展示 -->
        <div class="desktop-content-view">
          <el-card v-if="template">
            <div slot="header" class="card-header">
              <span>{{ template.name }}</span>
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
              <h3>原始视频</h3>
              <div class="media-preview">
                <div class="video-preview">
                  <div v-if="template.video_url" class="video-container">
                    <video controls style="width: 100%; max-height: 300px; background-color: #000;" :src="template.video_url"></video>
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
              <h3>处理结果</h3>
              <div class="result-preview">
                <!-- 换脸视频结果 -->
                <div class="result-item" v-if="template.replace_face_video_url">
                  <h4>换脸视频</h4>
                  <div class="video-container">
                    <video controls style="width: 100%; max-height: 300px; background-color: #000;" :src="template.replace_face_video_url"></video>
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
                    <video controls style="width: 100%; max-height: 300px; background-color: #000;" :src="template.replace_background_video_url"></video>
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
                    <video controls style="width: 100%; max-height: 300px; background-color: #000;" :src="template.result_video_url"></video>
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
              <h4 class="section-title">原始视频</h4>
              <div v-if="template.video_url" class="video-player">
                <video controls style="width: 100%; max-height: 200px; background-color: #000; display: block;" :src="template.video_url"></video>
              </div>
              <div v-else class="video-placeholder">
                <el-empty description="无原始视频"/>
              </div>
            </div>
            <div class="mobile-section">
              <h4 class="section-title">人脸图片</h4>
              <div v-if="template.face_image_url">
                <img :src="template.face_image_url" alt="人脸图片" style="max-width:120px;max-height:120px;display:block;margin:8px 0;"/>
                <a :href="template.face_image_url" target="_blank">下载/查看</a>
              </div>
              <div v-else><el-empty description="无"/></div>
            </div>
            <div class="mobile-section">
              <h4 class="section-title">背景</h4>
              <div v-if="template.background_url">
                <video v-if="isVideo(template.background_url)" :src="template.background_url" controls style="max-width:300px;max-height:180px;display:block;margin:8px 0;"/>
                <img v-else :src="template.background_url" alt="背景" style="max-width:120px;max-height:120px;display:block;margin:8px 0;"/>
                <a :href="template.background_url" target="_blank">下载/查看</a>
              </div>
              <div v-else><el-empty description="无"/></div>
            </div>
            
            <!-- 移动端处理结果展示 -->
            <div class="mobile-result-section" v-if="template.status === 'completed' || template.status === 'processing'">
              <h4 class="section-title">处理结果</h4>
              
              <!-- 换脸视频结果 -->
              <div v-if="template.replace_face_video_url" class="mobile-result-item">
                <h5>换脸视频</h5>
                <div class="video-player">
                  <video controls style="width: 100%; max-height: 200px; background-color: #000; display: block;" :src="template.replace_face_video_url"></video>
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
                  <video controls style="width: 100%; max-height: 200px; background-color: #000; display: block;" :src="template.replace_background_video_url"></video>
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
                  <video controls style="width: 100%; max-height: 200px; background-color: #000; display: block;" :src="template.result_video_url"></video>
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
<style scoped>
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

/* 基础样式 */
.digital-human-detail-container {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
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
  max-width: 1200px;
  margin: 0 auto;
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
.mobile-header-bar {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  background: white;
  padding: 12px 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  align-items: center;
}

.mobile-header-placeholder {
  display: none;
  height: 60px;
}

.header-back {
  display: flex;
  align-items: center;
  color: #409eff;
  cursor: pointer;
  font-size: 14px;
}

.header-back i {
  margin-right: 4px;
}

.header-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  flex: 1;
  text-align: center;
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
}

@media (max-width: 768px) {
  .digital-human-detail-container {
    padding: 0;
    background-color: white;
  }
  
  .page-header {
    display: none;
  }
  
  .mobile-header-bar {
    display: flex;
  }
  
  .mobile-header-placeholder {
    display: block;
  }
  
  .desktop-content-view {
    display: none;
  }
  
  .mobile-content-view {
    display: block;
  }
  
  .detail-content {
    box-shadow: none;
    border-radius: 0;
  }
  
  .media-preview {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}
</style> 