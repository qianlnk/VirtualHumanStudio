<template>
  <div class="tts-detail-container">
    <div class="page-header">
      <h2>语音合成任务详情</h2>
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
            <span class="label">创建时间：</span>
            <span>{{ formatDate(task.created_at) }}</span>
          </div>
          <div class="info-item">
            <span class="label">任务类型：</span>
            <span>{{ getTypeText(task.type) }}</span>
          </div>
          <div class="info-item" v-if="task.speaker_name">
            <span class="label">使用音色：</span>
            <span>{{ task.speaker_name }}</span>
          </div>
          <div class="info-item" v-if="task.description">
            <span class="label">任务描述：</span>
            <span>{{ task.description }}</span>
          </div>
        </div>
        
        <div class="content-section" v-if="task.type === 'text2speech' && task.input_text">
          <h3>输入文本</h3>
          <div class="text-content">{{ task.input_text }}</div>
        </div>
        
        <div class="audio-section" v-if="task.status === 'completed'">
          <h3>生成结果</h3>
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
            <el-button type="primary" @click="downloadOutput">下载音频</el-button>
          </div>
        </div>
        
        <div class="error-section" v-if="task.status === 'failed'">
          <h3>错误信息</h3>
          <div class="error-message">{{ task.error_msg || '任务处理失败' }}</div>
        </div>
      </el-card>
      
      <el-empty v-else description="未找到任务信息"></el-empty>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import { getAudioUrl, downloadFile } from '@/utils/fileAccess'

export default {
  name: 'TTSDetail',
  data() {
    return {
      loading: false,
      task: null,
      audioUrl: null,
      refreshInterval: null,
      baseURL: process.env.VUE_APP_API_URL || 'http://192.168.218.233:8080'
    }
  },
  computed: {
    token() {
      return localStorage.getItem('token') || ''
    },
    taskId() {
      return this.$route.params.id
    }
  },
  created() {
    this.fetchTaskDetail()
  },
  mounted() {
    // 如果任务状态是pending或processing，设置定时刷新
    if (this.task && (this.task.status === 'pending' || this.task.status === 'processing')) {
      this.startRefreshInterval()
    }
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
        const response = await axios.get(`${this.baseURL}/api/tts/${this.taskId}`, {
          headers: { Authorization: `Bearer ${this.token}` }
        })
        
        this.task = response.data.tts_task
        
        // 如果任务已完成，获取音频URL
        if (this.task.status === 'completed') {
          this.getAudioUrl()
          this.clearRefreshInterval()
        } else if (this.task.status === 'pending' || this.task.status === 'processing') {
          // 如果任务正在处理中，设置定时刷新
          this.startRefreshInterval()
        } else {
          this.clearRefreshInterval()
        }
      } catch (error) {
        console.error('获取任务详情失败:', error)
        this.$message.error('获取任务详情失败')
      } finally {
        this.loading = false
      }
    },
    
    // 获取音频URL
    async getAudioUrl() {
      try {
        this.audioUrl = await getAudioUrl(this.task.output_file)
      } catch (error) {
        console.error('获取音频失败:', error)
        this.$message.error('获取音频失败')
      }
    },
    
    // 下载输出文件
    async downloadOutput() {
      try {
        const audioPath = await getAudioUrl(this.task.output_file)
        const fileName = this.task.name ? `${this.task.name}.mp3` : `tts_output_${this.taskId}.mp3`
        await downloadFile(audioPath, fileName)
      } catch (error) {
        console.error('下载文件失败:', error)
        this.$message.error('下载文件失败')
      }
    },
    
    // 返回列表页
    goBack() {
      this.$router.push('/tts')
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
      const statusTextMap = {
        'pending': '等待处理',
        'processing': '处理中',
        'completed': '已完成',
        'failed': '失败'
      }
      return statusTextMap[status] || status
    },
    
    // 获取任务类型文本
    getTypeText(type) {
      const typeTextMap = {
        'text2speech': '文本转语音',
        'speech2text': '语音转文本'
      }
      return typeTextMap[type] || type
    }
  }
}
</script>

<style scoped>
.tts-detail-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.detail-content {
  background-color: #fff;
  border-radius: 4px;
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
  line-height: 1.5;
}

.label {
  font-weight: bold;
  margin-right: 10px;
}

.content-section,
.audio-section,
.error-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.text-content {
  padding: 10px;
  background-color: #f8f8f8;
  border-radius: 4px;
  white-space: pre-wrap;
  line-height: 1.5;
}

.audio-player {
  margin: 15px 0;
}

.action-buttons {
  margin-top: 15px;
}

.error-message {
  color: #f56c6c;
  padding: 10px;
  background-color: #fef0f0;
  border-radius: 4px;
}
</style>