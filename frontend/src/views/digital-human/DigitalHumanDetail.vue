<template>
  <div class="digital-human-detail-container">
    <div class="page-header">
      <h2>数字人合成任务详情</h2>
      <div>
        <el-button type="primary" @click="goBack">返回列表</el-button>
      </div>
    </div>
    
    <div v-loading="loading" class="detail-content">
      <el-card v-if="digitalHuman">
        <div slot="header" class="card-header">
          <span>{{ digitalHuman.name }}</span>
          <el-tag :type="getStatusType(digitalHuman.status)" class="status-tag">{{ getStatusText(digitalHuman.status) }}</el-tag>
        </div>
        
        <div class="task-info">
          <div class="info-item">
            <span class="label">创建时间：</span>
            <span>{{ formatDate(digitalHuman.created_at) }}</span>
          </div>
          <div class="info-item">
            <span class="label">任务代码：</span>
            <span>{{ digitalHuman.task_code }}</span>
          </div>
          <div class="info-item" v-if="digitalHuman.description">
            <span class="label">任务描述：</span>
            <span>{{ digitalHuman.description }}</span>
          </div>
          <div class="info-item">
            <span class="label">超分：</span>
            <span>{{ digitalHuman.chaofen ? '开启' : '关闭' }}</span>
          </div>
          <div class="info-item">
            <span class="label">水印：</span>
            <span>{{ digitalHuman.watermark_switch ? '开启' : '关闭' }}</span>
          </div>
          <div class="info-item">
            <span class="label">PN值：</span>
            <span>{{ digitalHuman.pn }}</span>
          </div>
        </div>
        
        <!-- 进度条 -->
        <div class="progress-section" v-if="digitalHuman.status === 'processing'">
          <h3>处理进度</h3>
          <el-progress :percentage="progress" :format="progressFormat"></el-progress>
        </div>
        
        <!-- 音频和视频预览 -->
        <div class="media-section">
          <h3>输入文件</h3>
          <div class="media-preview">
            <div class="audio-preview">
              <h4>音频文件</h4>
              <audio controls style="width: 100%" ref="audioPlayer">
                <source :src="audioUrl" type="audio/wav">
                您的浏览器不支持音频播放
              </audio>
            </div>
            <div class="video-preview">
              <h4>视频文件</h4>
              <video controls style="width: 100%; max-height: 300px" ref="videoPlayer">
                <source :src="videoUrl" type="video/mp4">
                您的浏览器不支持视频播放
              </video>
            </div>
          </div>
        </div>
        
        <!-- 合成结果 -->
        <div class="result-section" v-if="digitalHuman.status === 'completed' && digitalHuman.result_url">
          <h3>合成结果</h3>
          <div class="result-preview">
            <video controls style="width: 100%; max-height: 400px" ref="resultPlayer">
              <source :src="resultUrl" type="video/mp4">
              您的浏览器不支持视频播放
            </video>
          </div>
          <div class="action-buttons">
            <el-button type="primary" @click="downloadResult">下载结果</el-button>
          </div>
        </div>
        
        <!-- 错误信息 -->
        <div class="error-section" v-if="digitalHuman.status === 'failed'">
          <h3>错误信息</h3>
          <div class="error-message">{{ digitalHuman.error_msg || '任务处理失败' }}</div>
        </div>
      </el-card>
      
      <el-empty v-else description="未找到任务信息"></el-empty>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import { getAudioUrl, getVideoUrl, downloadFile } from '@/utils/fileAccess'

export default {
  name: 'DigitalHumanDetail',
  data() {
    return {
      loading: false,
      digitalHuman: null,
      progress: 0,
      refreshInterval: null,
      baseURL: process.env.VUE_APP_API_URL || 'http://localhost:8080',
      audioUrl: '',
      videoUrl: '',
      resultUrl: ''
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
    this.fetchDigitalHumanDetail()
  },
  mounted() {
    // 如果任务状态是pending或processing，设置定时刷新
    if (this.digitalHuman && (this.digitalHuman.status === 'pending' || this.digitalHuman.status === 'processing')) {
      this.startRefreshInterval()
    }
  },
  beforeDestroy() {
    this.clearRefreshInterval()
  },
  methods: {
    // 获取数字人详情
    fetchDigitalHumanDetail() {
      this.loading = true
      axios.get(`${this.baseURL}/api/digital-human/${this.taskId}`, {
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      })
        .then(response => {
          // 确保从response.data中获取正确的数据结构
          this.digitalHuman = response.data.digital_human || response.data
          
          // 加载媒体URL
          this.loadMediaUrls()
          
          // 如果任务正在处理中，开始轮询进度
          if (this.digitalHuman && (this.digitalHuman.status === 'pending' || this.digitalHuman.status === 'processing')) {
            this.startRefreshInterval()
          }
        })
        .catch(error => {
          console.error('获取数字人详情失败:', error)
          this.$message.error('获取数字人详情失败: ' + ((error.response && error.response.data && error.response.data.error) || error.message))
        })
        .finally(() => {
          this.loading = false
        })
    },
    
    // 查询进度
    queryProgress() {
      axios.get(`${this.baseURL}/api/digital-human/${this.taskId}/progress`, {
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      })
        .then(response => {
          this.progress = response.data.progress || 0
          
          // 如果任务已完成或失败，刷新详情并停止轮询
          if (response.data.status === 'completed' || response.data.status === 'failed') {
            this.fetchDigitalHumanDetail()
            this.clearRefreshInterval()
          }
        })
        .catch(error => {
          console.error('查询进度失败:', error)
          // 出错时不显示错误消息，避免频繁弹窗
        })
    },
    
    // 开始定时刷新
    startRefreshInterval() {
      this.clearRefreshInterval() // 先清除可能存在的定时器
      this.refreshInterval = setInterval(() => {
        this.queryProgress()
      }, 5000) // 每5秒查询一次进度
    },
    
    // 清除定时刷新
    clearRefreshInterval() {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval)
        this.refreshInterval = null
      }
    },
    
    // 下载结果
    downloadResult() {
      if (!this.digitalHuman || !this.digitalHuman.result_url) {
        this.$message.warning('没有可下载的结果文件')
        return
      }
      const fileName = `digital_human_${this.taskId}${this.getFileExtension(this.digitalHuman.result_url)}`
      downloadFile(this.digitalHuman.result_url, fileName)
    },

    // 获取文件扩展名
    getFileExtension(filePath) {
      if (!filePath) return ''
      const match = filePath.match(/\.[^.]+$/)
      return match ? match[0] : ''
    },

    // 加载媒体URL
    async loadMediaUrls() {
      try {
        if (this.digitalHuman) {
          if (this.digitalHuman.audio_url) {
            this.audioUrl = await getAudioUrl(this.digitalHuman.audio_url)
            // 重新设置音频源
            this.$nextTick(() => {
              if (this.$refs.audioPlayer) {
                this.$refs.audioPlayer.load()
              }
            })
          }
          if (this.digitalHuman.video_url) {
            this.videoUrl = await getVideoUrl(this.digitalHuman.video_url)
            // 重新设置视频源
            this.$nextTick(() => {
              if (this.$refs.videoPlayer) {
                this.$refs.videoPlayer.load()
              }
            })
          }
          if (this.digitalHuman.result_url) {
            this.resultUrl = await getVideoUrl(this.digitalHuman.result_url)
            // 重新设置结果视频源
            this.$nextTick(() => {
              if (this.$refs.resultPlayer) {
                this.$refs.resultPlayer.load()
              }
            })
          }
        }
      } catch (error) {
        console.error('加载媒体URL失败:', error)
        this.$message.error('加载媒体文件失败: ' + error.message)
      }
    },
    
    // 返回列表
    goBack() {
      this.$router.push('/digital-human')
    },
    
    // 格式化进度
    progressFormat(percentage) {
      return percentage === 100 ? '完成' : `${percentage}%`
    },
    
    // 格式化日期
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    },
    
    // 获取状态类型
    getStatusType(status) {
      switch (status) {
        case 'completed': return 'success'
        case 'processing': return 'warning'
        case 'pending': return 'info'
        case 'failed': return 'danger'
        default: return 'info'
      }
    },
    
    // 获取状态文本
    getStatusText(status) {
      switch (status) {
        case 'completed': return '已完成'
        case 'processing': return '处理中'
        case 'pending': return '等待中'
        case 'failed': return '失败'
        default: return '未知'
      }
    }
  }
}
</script>

<style scoped>
.digital-human-detail-container {
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
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  color: #606266;
}

.progress-section,
.media-section,
.result-section,
.error-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.media-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.audio-preview,
.video-preview {
  flex: 1;
  min-width: 300px;
}

.action-buttons {
  margin-top: 15px;
  text-align: center;
}

.error-message {
  color: #f56c6c;
  padding: 10px;
  background-color: #fef0f0;
  border-radius: 4px;
}

.status-tag {
  margin-left: 10px;
}
</style>