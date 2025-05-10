<template>
  <div class="tts-detail-container">
    <!-- PC端顶部导航栏 -->
    <div class="page-header">
      <h2>语音合成任务详情</h2>
      <div>
        <el-button type="primary" @click="goBack">返回列表</el-button>
      </div>
    </div>
    
    <!-- 移动端顶部导航栏 -->
    <div class="mobile-header-bar">
      <div class="header-back" @click="goBack">
        <i class="el-icon-arrow-left"></i>
        <span>返回</span>
      </div>
      <h2 class="header-title">语音合成任务</h2>
    </div>
    
    <!-- 移动端头部占位 -->
    <div class="mobile-header-placeholder"></div>
    
    <div class="detail-content-wrapper">
      <div v-loading="loading" class="detail-content">
        <!-- PC端展示 -->
        <div class="desktop-content-view">
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
            
            <div class="content-section" v-if="task.input_text">
              <h3>输入文本</h3>
              <div class="text-content-wrapper">
                <div class="text-content">{{ task.input_text }}</div>
                <el-button class="copy-btn" type="text" icon="el-icon-document-copy" @click="copyInputText"></el-button>
              </div>
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
        
        <!-- 移动端展示 -->
        <div class="mobile-content-view">
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
              <div class="info-item" v-if="task.type">
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
            
            <!-- 输入文本 - 始终显示 -->
            <div class="mobile-section">
              <div class="section-header">
                <h4 class="section-title">输入文本</h4>
                <el-button class="mobile-copy-btn" type="text" icon="el-icon-document-copy" @click="copyInputText"></el-button>
              </div>
              <div class="mobile-text-content">{{ task.input_text || '无输入文本' }}</div>
            </div>
            
            <!-- 音频结果 -->
            <div class="mobile-section" v-if="task.status === 'completed'">
              <h4 class="section-title">生成结果</h4>
              <div v-if="audioUrl" class="audio-player">
                <audio controls ref="mobileAudioPlayer" style="width: 100%">
                  <source :src="audioUrl" type="audio/mpeg">
                  您的浏览器不支持音频播放
                </audio>
              </div>
              <div v-else class="loading-audio">
                <el-skeleton :loading="true" animated>
                  <template slot="template">
                    <el-skeleton-item variant="rect" style="width: 100%; height: 36px"/>
                  </template>
                </el-skeleton>
              </div>
              <div class="action-buttons">
                <el-button type="primary" @click="downloadOutput" size="small">下载音频</el-button>
              </div>
            </div>
            
            <!-- 错误信息 -->
            <div class="mobile-section" v-if="task.status === 'failed'">
              <h4 class="section-title">错误信息</h4>
              <div class="error-message">{{ task.error_msg || '任务处理失败' }}</div>
            </div>
          </div>
          
          <el-empty v-else description="未找到任务信息"></el-empty>
        </div>
      </div>
    </div>

    <!-- 添加隐藏的输入框用于复制 -->
    <input
      ref="copyInput"
      type="text"
      class="copy-input"
      :value="task ? task.input_text : ''"
      readonly
    />
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
      baseURL: process.env.VUE_APP_API_URL || 'http://localhost:8080',
      screenWidth: window.innerWidth
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
    window.addEventListener('resize', this.handleResize)
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
    window.removeEventListener('resize', this.handleResize)
  },
  methods: {
    // 处理窗口大小变化
    handleResize() {
      this.screenWidth = window.innerWidth
    },
    
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
    },
    
    // 复制输入文本到剪贴板
    copyInputText() {
      if (!this.task || !this.task.input_text) return
      
      const copyInput = this.$refs.copyInput
      if (!copyInput) return
      
      try {
        // 选择文本
        copyInput.select()
        copyInput.setSelectionRange(0, copyInput.value.length)
        
        // 执行复制
        const successful = document.execCommand('copy')
        if (successful) {
          this.$message.success('文本已复制到剪贴板')
        } else {
          this.$message.error('复制失败，请手动复制')
        }
      } catch (err) {
        console.error('复制失败:', err)
        this.$message.error('复制失败，请手动复制')
      }
    }
  }
}
</script>

<style>
/* 全局样式覆盖，强制禁止水平滚动 */
html, body {
  overflow-x: hidden !important;
  width: 100% !important;
  max-width: 100% !important;
  position: relative;
}

@media screen and (max-width: 768px) {
  html, body, #app, .app-container {
    overflow-x: hidden !important;
    width: 100% !important;
    max-width: 100% !important;
  }
}
</style>

<style scoped>
.tts-detail-container {
  padding: 20px;
  width: 100%;
  box-sizing: border-box;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  width: 100%;
  box-sizing: border-box;
}

.detail-content {
  background-color: #fff;
  border-radius: 4px;
  width: 100%;
  box-sizing: border-box;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  box-sizing: border-box;
}

.status-tag {
  margin-left: 10px;
}

.task-info {
  margin-bottom: 20px;
  width: 100%;
  box-sizing: border-box;
}

.info-item {
  margin-bottom: 10px;
  line-height: 1.5;
  width: 100%;
  box-sizing: border-box;
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
  width: 100%;
  box-sizing: border-box;
}

.text-content-wrapper {
  position: relative;
  width: 100%;
  box-sizing: border-box;
}

.text-content {
  padding: 10px;
  background-color: #f8f8f8;
  border-radius: 4px;
  white-space: pre-wrap;
  line-height: 1.5;
  color: #333;
  width: 100%;
  box-sizing: border-box;
}

.copy-btn {
  position: absolute;
  top: 5px;
  right: 5px;
  font-size: 16px;
  padding: 5px;
  color: #606266;
  background-color: rgba(255, 255, 255, 0.7);
  border-radius: 4px;
}

.copy-btn:hover {
  color: #409EFF;
  background-color: rgba(255, 255, 255, 0.9);
}

.audio-player {
  margin: 15px 0;
  width: 100%;
  box-sizing: border-box;
}

.action-buttons {
  margin-top: 15px;
  width: 100%;
  box-sizing: border-box;
}

.error-message {
  color: #f56c6c;
  padding: 10px;
  background-color: #fef0f0;
  border-radius: 4px;
  width: 100%;
  box-sizing: border-box;
}

/* 移动端适配样式 */
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
  z-index: 1001;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  color: #fff;
  width: 100%;
  display: none; /* 默认隐藏，在移动端显示 */
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
  display: none; /* 默认隐藏，在移动端显示 */
}

/* 移动端内容区 */
.detail-content-wrapper {
  width: 100%;
  box-sizing: border-box;
}

.mobile-content-view {
  display: none; /* 默认隐藏，在移动端显示 */
  padding: 0;
  width: 100%;
  background-color: #fff;
  min-height: 100%;
  box-sizing: border-box;
}

.mobile-content-inner {
  background-color: #fff;
  min-height: calc(100vh - 56px);
  width: 100%;
  box-sizing: border-box;
}

/* 移动端基本信息 */
.basic-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
  width: 100%;
  box-sizing: border-box;
  background-color: #fff;
}

.create-time {
  font-size: 12px;
  color: #666;
}

.mobile-task-info {
  padding: 8px 12px;
  background-color: #fff;
  box-sizing: border-box;
  width: 100%;
}

.mobile-section {
  padding: 8px 12px;
  border-top: 1px solid #f0f0f0;
  margin-top: 8px;
  background-color: #fff;
  box-sizing: border-box;
  width: 100%;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin: 5px 0;
  color: #333;
  width: 100%;
  box-sizing: border-box;
}

/* 移动端文本内容 */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  width: 100%;
  box-sizing: border-box;
}

.mobile-copy-btn {
  padding: 2px 5px;
  margin: 0;
  font-size: 16px;
  color: #409EFF;
}

.mobile-text-content {
  padding: 10px;
  background-color: #f8f8f8;
  border-radius: 4px;
  white-space: pre-wrap;
  line-height: 1.5;
  font-size: 14px;
  height: 100px; /* 固定高度 */
  overflow-y: auto; /* 允许垂直滚动 */
  overflow-x: hidden; /* 禁止水平滚动 */
  word-break: break-word;
  color: #333; /* 确保文本颜色可见 */
  width: 100%;
  box-sizing: border-box;
}

/* PC端特有样式 */
.desktop-content-view {
  display: block; /* 默认显示 */
  width: 100%;
  box-sizing: border-box;
}

/* 媒体查询 - 移动端适配 */
@media screen and (max-width: 768px) {
  /* 隐藏PC端，显示移动端 */
  .page-header {
    display: none;
  }
  
  .desktop-content-view {
    display: none;
  }
  
  .mobile-header-bar {
    display: flex;
    width: 100%;
  }
  
  .mobile-header-placeholder {
    display: block;
  }
  
  .mobile-content-view {
    display: block;
    height: calc(100vh - 56px);
    overflow-y: auto; /* 允许垂直滚动 */
    overflow-x: hidden; /* 禁止水平滚动 */
    width: 100%;
    box-sizing: border-box;
  }
  
  .tts-detail-container {
    padding: 0;
    width: 100%;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 1000;
    background-color: #fff;
    height: 100vh;
    overflow-x: hidden; /* 禁止水平滚动 */
    box-sizing: border-box;
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
    background-color: #fff;
    height: calc(100vh - 56px);
    overflow-y: auto; /* 允许垂直滚动 */
    overflow-x: hidden; /* 禁止水平滚动 */
  }
  
  .detail-content {
    width: 100%;
    margin: 0;
    background-color: #fff;
    box-shadow: none;
    border-radius: 0;
    min-height: 100%;
    overflow-x: hidden; /* 禁止水平滚动 */
    box-sizing: border-box;
  }
  
  .mobile-content-inner {
    background-color: #fff;
    min-height: 100%;
    width: 100%;
    box-sizing: border-box;
    overflow-x: hidden; /* 禁止水平滚动 */
  }
  
  /* 移除之前的Flex布局相关样式 */
  .mobile-section {
    flex-shrink: unset;
    width: 100%;
    box-sizing: border-box;
    overflow-x: hidden; /* 禁止水平滚动 */
  }
  
  .basic-info {
    flex-shrink: unset;
    width: 100%;
    box-sizing: border-box;
  }
  
  .mobile-task-info {
    flex-shrink: unset;
    width: 100%;
    box-sizing: border-box;
  }
  
  /* 输入文本区域固定高度 */
  .mobile-section:nth-child(3) {
    flex: unset;
    width: 100%;
    box-sizing: border-box;
  }
  
  .mobile-section:nth-child(3) .mobile-text-content {
    flex: unset;
    height: 100px; /* 固定高度 */
    max-height: 100px;
    overflow-y: auto; /* 允许垂直滚动 */
    overflow-x: hidden; /* 禁止水平滚动 */
    width: 100%;
    box-sizing: border-box;
  }
  
  .audio-player {
    margin: 8px 0;
    height: 36px;
    width: 100%;
    box-sizing: border-box;
  }
  
  .action-buttons {
    margin-top: 10px;
    margin-bottom: 20px;
    width: 100%;
    box-sizing: border-box;
  }
  
  .action-buttons .el-button {
    width: 100%;
    box-sizing: border-box;
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

/* 添加隐藏输入框样式 */
.copy-input {
  position: absolute;
  left: -9999px;
  top: -9999px;
  opacity: 0;
  pointer-events: none;
}
</style>