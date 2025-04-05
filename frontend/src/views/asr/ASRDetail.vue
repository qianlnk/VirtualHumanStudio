<template>
  <div class="asr-detail-container">
    <div class="page-header">
      <h2>语音识别任务详情</h2>
      <div>
        <el-button type="primary" @click="goBack">返回列表</el-button>
      </div>
    </div>
    
    <div v-loading="loading" class="detail-content">
      <el-card v-if="task">
        <div slot="header" class="card-header">
          <span>{{ task.name }}</span>
          <el-tag :type="getStatusType(task.status)" class="status-tag">{{ getStatusText(task.status) }}</el-tag>
        </div>
        
        <div class="task-info">
          <div class="info-item">
            <span class="label">任务ID：</span>
            <span>{{ task.id }}</span>
          </div>
          <div class="info-item">
            <span class="label">创建时间：</span>
            <span>{{ formatDate(task.created_at) }}</span>
          </div>
          <div class="info-item" v-if="task.description">
            <span class="label">任务描述：</span>
            <span>{{ task.description }}</span>
          </div>
        </div>
        
        <div class="audio-section" v-if="task.input_file">
          <h3>原始音频</h3>
          <div v-if="audioUrl" class="audio-player">
            <audio controls ref="audioPlayer" style="width: 100%">
              <source :src="audioUrl" type="audio/mpeg">
              您的浏览器不支持音频播放
            </audio>
          </div>
          <div v-else class="loading-audio">
            <el-skeleton :loading="true" animated>
              <template slot="template">
                <el-skeleton-item variant="rect" style="width: 100%; height: 54px"/>
              </template>
            </el-skeleton>
          </div>
          <div class="action-buttons">
            <el-button type="primary" @click="downloadAudio">下载音频</el-button>
          </div>
        </div>
        
        <div class="content-section" v-if="task.output_text">
          <h3>识别结果</h3>
          <div class="text-content">{{ task.output_text }}</div>
        </div>
        
        <div class="error-section" v-if="task.status === 'failed'">
          <h3>错误信息</h3>
          <div class="error-message">{{ task.error_msg || '任务处理失败' }}</div>
          <div class="action-buttons">
            <el-button type="primary" @click="handleRetry">重试任务</el-button>
          </div>
        </div>
      </el-card>
      
      <el-empty v-else description="未找到任务信息"></el-empty>
    </div>
  </div>
</template>

<script>
import { getAudioUrl, downloadFile } from '@/utils/fileAccess'

export default {
  name: 'ASRDetail',
  data() {
    return {
      loading: false,
      task: null,
      audioUrl: null,
      refreshInterval: null
    }
  },
  computed: {
    taskId() {
      return this.$route.params.id
    }
  },
  created() {
    this.fetchTaskDetail()
  },
  mounted() {
    // mounted钩子中不再直接访问this.task
    // 定时刷新的逻辑已移至fetchTaskDetail方法中处理
  },
  beforeDestroy() {
    this.clearRefreshInterval()
    // 释放音频URL
    if (this.audioUrl) {
      URL.revokeObjectURL(this.audioUrl)
    }
  },
  methods: {
    // 获取任务详情
    async fetchTaskDetail() {
      this.loading = true
      try {
        const response = await this.$http.get(`/api/asr/${this.taskId}`)
        
        // 检查响应数据格式，后端API返回的数据包含在task字段中
        if (response.data && response.data.task) {
          this.task = response.data.task
          console.log('获取到的任务详情:', this.task)
          
          // 如果任务已完成，获取音频URL
          if (this.task.input_file) {
            this.getAudioUrl()
          }
          
          // 根据任务状态设置刷新
          if (this.task.status === 'pending' || this.task.status === 'processing') {
            // 如果任务正在处理中，设置定时刷新
            this.startRefreshInterval()
          } else {
            this.clearRefreshInterval()
          }
        } else {
          console.error('任务详情数据格式不正确:', response.data)
          this.$message.warning('获取任务详情数据格式不正确')
        }
      } catch (error) {
        console.error('获取任务详情失败:', error)
        this.$message.error('获取任务详情失败: ' + (error.message || '未知错误'))
      } finally {
        this.loading = false
      }
    },
    
    // 获取音频URL
    async getAudioUrl() {
      try {
        this.audioUrl = await getAudioUrl(this.task.input_file)
      } catch (error) {
        console.error('获取音频失败:', error)
        this.$message.error('获取音频失败')
      }
    },
    
    // 下载音频文件
    async downloadAudio() {
      try {
        const audioPath = await getAudioUrl(this.task.input_file)
        const fileName = this.task.name ? `${this.task.name}.mp3` : `asr_input_${this.taskId}.mp3`
        await downloadFile(audioPath, fileName)
      } catch (error) {
        console.error('下载文件失败:', error)
        this.$message.error('下载文件失败')
      }
    },
    
    // 返回列表页
    goBack() {
      this.$router.push('/speech2text')
    },
    
    // 重试任务
    async handleRetry() {
      try {
        await this.$http.post(`/api/asr/${this.taskId}/retry`)
        this.$message.success('重试任务已提交')
        this.fetchTaskDetail()
      } catch (error) {
        this.$message.error('重试任务失败')
      }
    },
    
    // 开始定时刷新
    startRefreshInterval() {
      this.clearRefreshInterval() // 先清除之前的定时器
      this.refreshInterval = setInterval(() => {
        this.fetchTaskDetail()
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
      return new Date(dateString).toLocaleString()
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
    }
  }
}
</script>

<style scoped>
.asr-detail-container {
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
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 20px;
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
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
}

.info-item {
  margin-bottom: 10px;
  display: flex;
}

.info-item .label {
  font-weight: bold;
  width: 100px;
}

.content-section,
.audio-section,
.error-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.content-section h3,
.audio-section h3,
.error-section h3 {
  margin-bottom: 15px;
  color: #409EFF;
}

.text-content {
  background: #f8f8f8;
  padding: 15px;
  border-radius: 4px;
  white-space: pre-wrap;
  color: #333;
  max-height: 300px;
  overflow-y: auto;
}

.error-message {
  color: #f56c6c;
  background: rgba(245, 108, 108, 0.1);
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.action-buttons {
  margin-top: 15px;
  display: flex;
  justify-content: flex-end;
}

.loading-audio {
  margin-bottom: 15px;
}

.audio-player {
  margin-bottom: 15px;
}
</style>