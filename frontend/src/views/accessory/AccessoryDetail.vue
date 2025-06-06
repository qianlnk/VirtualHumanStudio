<template>
  <div class="accessory-detail-container">
    <div class="page-header">
      <h2>饰品替换任务详情</h2>
      <div>
        <el-button type="primary" @click="goBack">返回列表</el-button>
      </div>
    </div>
    <div class="image-preview-overlay" v-if="isPreviewActive" @click="closePreview" @wheel.prevent="handleZoom" @mousedown="startDrag" @mousemove="onDrag" @mouseup="stopDrag" @mouseleave="stopDrag">
      <img :src="previewUrl" :style="imageStyle" class="preview-image" ref="previewImage">
    </div>
    
    <div v-loading="loading" class="detail-content">
      <el-card v-if="accessory">
        <div slot="header" class="card-header">
          <span>{{ accessory.name }}</span>
          <el-tag :type="getStatusType(accessory.status)" class="status-tag">{{ getStatusText(accessory.status) }}</el-tag>
        </div>
        
        <div class="task-info">
          <div class="info-item">
            <span class="label">创建时间：</span>
            <span>{{ formatDate(accessory.created_at) }}</span>
          </div>
          <div class="info-item" v-if="accessory.description">
            <span class="label">任务描述：</span>
            <span>{{ accessory.description }}</span>
          </div>
        </div>
        
        <!-- 输入图片展示 -->
        <div class="images-section">
          <h3>输入图片</h3>
          <div class="image-grid">
            <div class="image-item">
              <h4>白底图片</h4>
              <div class="image-container">
                <img :src="item_image" alt="白底图片" class="detail-image" crossorigin="anonymous" @click="previewImageDirect(item_image)">
              </div>
            </div>
            <div class="image-item">
              <h4>模特物品图片</h4>
              <div class="image-container">
                <img :src="model_image" alt="模特物品图片" class="detail-image" crossorigin="anonymous" @click="previewImageDirect(model_image)">
              </div>
            </div>
            <div class="image-item">
              <h4>蒙版图片</h4>
              <div class="image-container">
                <img :src="mask_image" alt="蒙版图片" class="detail-image" crossorigin="anonymous" @click="previewImageDirect(mask_image)">
              </div>
            </div>
          </div>
        </div>
        
        <!-- 生成结果 -->
        <div class="result-section" v-if="accessory.status === 'completed' && accessory.result_image">
          <h3>替换结果</h3>
          <div class="result-container">
            <img :src="result_image" alt="替换结果" class="result-image" crossorigin="anonymous" @click="previewImageDirect(result_image)">
          </div>
          <div class="action-buttons">
            <el-button type="primary" @click="downloadResult">下载结果</el-button>
            <el-button type="success" @click="shareTask" class="share-button" :disabled="isShared">
              {{ getShareButtonText() }}
            </el-button>
          </div>
        </div>
        
        <!-- 错误信息 -->
        <div class="error-section" v-if="accessory.status === 'failed'">
          <h3>错误信息</h3>
          <div class="error-message">{{ accessory.error_msg || '任务处理失败' }}</div>
        </div>
      </el-card>
      
      <el-empty v-else description="未找到任务信息"></el-empty>
    </div>
    
    
  </div>
</template>

<script>
import axios from 'axios'
import { downloadFile, getImageUrl } from '@/utils/fileAccess'
import { shareTask } from '@/api/share'

export default {
  name: 'AccessoryDetail',
  data() {
    return {
      loading: false,
      accessory: null,
      refreshInterval: null,
      baseURL: process.env.VUE_APP_API_URL || 'http://localhost:8080',
      // 保存原始URL，用于下载
      item_image: '',
      model_image: '',
      mask_image: '',
      result_image: '',
      // 图片预览相关
      isPreviewActive: false,
      previewUrl: '',
      zoomLevel: 1,
      isDragging: false,
      dragStartX: 0,
      dragStartY: 0,
      translateX: 0,
      translateY: 0,
      // 分享相关状态
      sharing: false
    }
  },
  computed: {
    token() {
      return localStorage.getItem('token') || ''
    },
    taskId() {
      return this.$route.params.id
    },
    imageStyle() {
      return {
        transform: `scale(${this.zoomLevel}) translate(${this.translateX}px, ${this.translateY}px)`,
        transition: this.isDragging ? 'none' : 'transform 0.3s'
      }
    },
    // 是否已分享
    isShared() {
      return this.accessory && (this.accessory.is_shared || this.sharing)
    }
  },
  created() {
    this.fetchAccessoryDetail()
  },
  mounted() {
    // 如果任务状态是pending或processing，设置定时刷新
    if (this.accessory && (this.accessory.status === 'pending' || this.accessory.status === 'processing')) {
      this.startRefreshInterval()
    }
  },
  beforeDestroy() {
    this.clearRefreshInterval()
  },
  methods: {
    // 获取任务详情
    async fetchAccessoryDetail() {
      if (!this.taskId) {
        this.$message.error('无效的任务ID')
        this.$router.push('/accessory')
        return
      }
      this.loading = true
      axios.get(`${this.baseURL}/api/accessory/${this.taskId}`, {
          headers: { 
            'Authorization': `Bearer ${this.token}` 
          }
        })
        .then(response => {
          this.accessory = response.data.accessory || response.data

          // 加载图片URL
          this.loadImageUrls()

          // 如果任务正在处理中，开始轮询进度
          if (this.accessory && ( this.accessory.status === 'pending' || this.accessory.status === 'processing')) {
            this.startRefreshInterval()
          }
        })
        .catch(error => {
          console.error('获取饰品替换任务详情失败:', error)
          this.$message.error('获取饰品替换任务详情失败' + ((error.response && error.response.data && error.response.data.error) || error.message))
        })
        .finally(() => {
          this.loading = false
        })
    },
    
    // 加载图片URL
    async loadImageUrls() {
      try {
        if (this.accessory) {  
          if (this.accessory.item_image) {
            this.item_image = await getImageUrl(this.accessory.item_image)
            // 重新设置图片源
            this.$nextTick(() => {
              if (this.$refs.item_image && this.$refs.item_image.src !== this.item_image) {
                this.$refs.item_image.src = this.item_image
              }
            })
          }
          if (this.accessory.model_image) {
            this.model_image = await getImageUrl(this.accessory.model_image)
            // 重新设置图片源
            this.$nextTick(() => {
              if (this.$refs.model_image && this.$refs.modelImage.src!== this.model_image) {
                this.$refs.model_image.src = this.model_image
              }
            })
          }
          if (this.accessory.mask_image) {
            this.mask_image = await getImageUrl(this.accessory.mask_image)
            // 重新设置图片源
            this.$nextTick(() => {
              if (this.$refs.mask_image && this.$refs.mask_image.src!== this.maskImageUrl) {
                this.$refs.mask_image.src = this.mask_image
              }
            })
          }
          if (this.accessory.result_image) {
            this.result_image = await getImageUrl(this.accessory.result_image)
            // 重新设置图片源
            this.$nextTick(() => {
              if (this.$refs.result_image && this.$refs.result_image.src!== this.result_image) {
                this.$refs.result_image.src = this.result_image
              }
            })
          }
        }
      } catch (error) {
        console.error('加载图片URL失败:', error)
        this.$message.error('加载图片文件失败: ' + error.message)
      }
    },
    
    // 直接预览图片
    previewImageDirect(url) {
      if (!url) return
      this.previewUrl = url
      this.isPreviewActive = true
      this.resetZoom()
    },

    // 关闭预览
    closePreview() {
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
    
    // 下载结果图片
    async downloadResult() {
      try {
        if (!this.accessory || !this.accessory.result_image) {
          this.$message.warning('没有可下载的结果')
          return
        }
        
        let resultImageUrl =  this.accessory.result_image
        
        // 确保URL是完整的
        if (!resultImageUrl.startsWith('http')) {
          resultImageUrl = `${this.baseURL}${resultImageUrl}`
        }
        
        const fileName = this.accessory.name ? `${this.accessory.name}.png` : `accessory_result_${this.taskId}.png`
        await downloadFile(resultImageUrl, fileName)
        this.$message.success('下载成功')
      } catch (error) {
        console.error('下载文件失败:', error)
        this.$message.error('下载文件失败')
      }
    },
    
    // 返回列表页
    goBack() {
      this.$router.push('/accessory')
    },
    
    // 开始定时刷新
    startRefreshInterval() {
      this.clearRefreshInterval() // 先清除之前的定时器
      this.refreshInterval = setInterval(() => {
        this.fetchAccessoryDetail()
      }, 5000) // 每5秒刷新一次
    },
    
    // 清除定时刷新
    clearRefreshInterval() {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval)
        this.refreshInterval = null
      }
    },
    
    // 格式化日期
    formatDate(dateString) {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleString()
    },
    
    // 获取状态类型
    getStatusType(status) {
      const statusMap = {
        'pending': 'info',
        'processing': 'warning',
        'completed': 'success',
        'failed': 'danger'
      }
      return statusMap[status] || 'info'
    },
    
    // 获取状态文本
    getStatusText(status) {
      const statusMap = {
        'pending': '等待中',
        'processing': '处理中',
        'completed': '已完成',
        'failed': '失败'
      }
      return statusMap[status] || status
    },
    
    // 分享任务
    async shareTask() {
      try {
        if (!this.accessory || this.accessory.status !== 'completed') {
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
            text: '正在提交分享请求...',
            duration: 2000
          })
          
          const result = await shareTask({
            taskId: this.accessory.id,
            mode: 'comfyui',
            taskType: 'accessory'
          })
          
          if (result.success) {
            this.$message.success(result.message)
            // 更新本地任务状态
            this.accessory.is_shared = true
            this.accessory.share_status = 'pending_review'
            this.accessory.share_time = new Date().toISOString()
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
      if (!this.accessory) return '分享到灵感页'
      
      // 根据分享状态返回不同文本
      if (this.accessory.is_shared) {
        switch (this.accessory.share_status) {
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
.accessory-detail-container {
  padding: 40px;
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #fff;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.page-header h2 {
  font-size: 2em;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.detail-content {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-tag {
  margin-left: 10px;
}

.task-info {
  margin-bottom: 20px;
  padding: 15px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
}

.info-item {
  margin-bottom: 10px;
  display: flex;
}

.info-item .label {
  font-weight: bold;
  width: 100px;
  color: #64b5f6;
}

.images-section, .result-section, .error-section {
  margin-top: 20px;
  padding: 15px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
}

.images-section h3, .result-section h3, .error-section h3 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #64b5f6;
  font-size: 1.2em;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-top: 15px;
}

.image-item h4 {
  margin-top: 0;
  margin-bottom: 10px;
  color: #fff;
  font-size: 1em;
}

.image-container {
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.2);
  padding: 10px;
  text-align: center;
}

.detail-image {
  max-width: 100%;
  max-height: 200px;
  object-fit: contain;
  cursor: pointer;
  transition: transform 0.2s;
}

.detail-image:hover {
  transform: scale(1.05);
}

.result-container {
  text-align: center;
  margin-bottom: 15px;
}

.result-image {
  max-width: 100%;
  max-height: 400px;
  object-fit: contain;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.2);
  padding: 10px;
  cursor: pointer;
  transition: transform 0.2s;
}

.result-image:hover {
  transform: scale(1.02);
}

.action-buttons {
  margin-top: 15px;
  text-align: center;
}

.error-message {
  color: #f56c6c;
  padding: 10px;
  background: rgba(245, 108, 108, 0.1);
  border-radius: 4px;
}

/* 图片预览对话框样式 */
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
</style>