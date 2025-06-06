<template>
  <div class="voice-clone-detail-container">
    <!-- PC端顶部导航栏，参考TTSDetail样式 -->
    <div class="page-header">
      <h2>音色克隆详情</h2>
      <div>
        <el-button type="primary" @click="goBack">返回列表</el-button>
      </div>
    </div>
    
    <!-- 顶部导航栏 -->
    <div class="mobile-header-bar">
      <div class="header-back" @click="goBack">
        <i class="el-icon-arrow-left"></i>
        <span>返回</span>
      </div>
      <h2 class="header-title">{{ task.speaker_name }}</h2>
    </div>
    
    <!-- 移动端头部占位 -->
    <div class="mobile-header-placeholder"></div>
    
    <div class="detail-content-wrapper">
      <div v-loading="loading" class="detail-content">
        <!-- 基本信息 - 简化版 -->
        <div class="basic-info">
          <div class="status-tag">
            <el-tag :type="getStatusType(task.status)">{{ task.status }}</el-tag>
          </div>
          <div class="create-time">创建时间：{{ formatDate(task.created_at) }}</div>
        </div>
        
        <!-- 移动端：合并显示所有内容 -->
        <div class="mobile-content-view">
          <!-- 原始音频和克隆结果整合 -->
          <div class="mobile-audio-section">
            <h4 class="section-title">原始音频</h4>
            <audio controls :src="audioUrlWithToken" class="audio-player"></audio>
            
            <div v-if="task.status === 'completed' && task.sample_file" class="sample-section">
              <h4 class="section-title">克隆结果</h4>
              <audio controls :src="sampleUrlWithToken" class="audio-player"></audio>
            </div>
            
            <div v-if="task.prompt_text" class="prompt-text">
              <h4 class="section-title">提示文本</h4>
              <p>{{ task.prompt_text }}</p>
            </div>
          </div>
        </div>
        
        <!-- 桌面端：分卡片显示 -->
        <div class="desktop-content-view">
          <!-- 基本信息卡片 -->
          <div class="info-card">
            <div class="info-header">
              <div class="info-header-left">
                <div class="info-title">音色克隆详情</div>
                <div class="create-time">创建时间：{{ formatDate(task.created_at) }}</div>
              </div>
              <el-tag :type="getStatusType(task.status)">{{ task.status }}</el-tag>
            </div>
          </div>
          
          <!-- 音频信息 -->
          <el-card class="detail-card">
            <div slot="header" class="card-header">
              <span>原始音频</span>
            </div>
            <div class="audio-info">
              <audio controls :src="audioUrlWithToken" class="audio-player"></audio>
              <div v-if="task.prompt_text" class="prompt-text">
                <h4>提示文本</h4>
                <p>{{ task.prompt_text }}</p>
              </div>
            </div>
          </el-card>
          
          <!-- 结果信息 -->
          <el-card class="detail-card" v-if="task.status === 'completed'">
            <div slot="header" class="card-header">
              <span>克隆结果</span>
            </div>
            <div class="result-info">
              <div class="sample-audio" v-if="task.sample_file">
                <h4>示例音频</h4>
                <audio controls :src="sampleUrlWithToken" class="audio-player"></audio>
              </div>
            </div>
          </el-card>
        </div>
        
        <!-- 错误信息 - 两种视图都显示 -->
        <div v-if="task.status === 'failed'" class="error-card">
          <h4 class="section-title">错误信息</h4>
          <el-alert
            title="处理失败"
            type="error"
            :description="task.error_message || '未知错误'"
            show-icon>
          </el-alert>
          <div class="retry-button">
            <el-button type="primary" @click="retryTask">重试</el-button>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 添加到音色库对话框 -->
    <el-dialog title="添加到音色库" :visible.sync="addToLibraryDialogVisible" width="90%" custom-class="mobile-dialog">
      <el-form :model="libraryForm" :rules="libraryRules" ref="libraryForm" label-width="80px">
        <el-form-item label="音色名称" prop="name">
          <el-input v-model="libraryForm.name" placeholder="请输入音色名称"></el-input>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input type="textarea" v-model="libraryForm.description" placeholder="请输入音色描述"></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="addToLibraryDialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitAddToLibrary">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  name: 'VoiceCloneDetail',
  data() {
    return {
      loading: false,
      submitting: false,
      screenWidth: window.innerWidth,
      task: {
        id: '',
        name: '',
        status: '',
        created_at: '',
        description: '',
        audio_url: '',
        audio_duration: 0,
        error_message: '',
        sample_url: '',
        status_message: ''
      },
      refreshTimer: null,
      addToLibraryDialogVisible: false,
      libraryForm: {
        name: '',
        description: ''
      },
      libraryRules: {
        name: [
          { required: true, message: '请输入音色名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ]
      },
      audioBlob: null,
      sampleBlob: null,
      audioDataUrl: '',
      sampleDataUrl: ''
    }
  },
  computed: {
    taskId() {
      return this.$route.params.id
    },
    progressStatus() {
      if (this.progress < 100) return ''
      return 'success'
    },
    audioUrlWithToken() {
      return this.audioDataUrl || (this.audioBlob ? URL.createObjectURL(this.audioBlob) : '')
    },
    sampleUrlWithToken() {
      return this.sampleDataUrl || (this.sampleBlob ? URL.createObjectURL(this.sampleBlob) : '')
    }
  },
  created() {
    this.fetchTaskDetail()
    window.addEventListener('resize', this.handleResize)
  },
  beforeDestroy() {
    this.clearProgressTimer()
    this.clearRefreshTimer()
    window.removeEventListener('resize', this.handleResize)
    
    // 只有在使用了Blob URL时才需要释放
    if (this.audioBlob && !this.audioDataUrl) {
      URL.revokeObjectURL(this.audioUrlWithToken)
    }
    if (this.sampleBlob && !this.sampleDataUrl) {
      URL.revokeObjectURL(this.sampleUrlWithToken)
    }
    
    // 清空所有数据
    this.audioBlob = null
    this.sampleBlob = null
    this.audioDataUrl = ''
    this.sampleDataUrl = ''
  },
  methods: {
    // 获取任务详情
    async fetchTaskDetail() {
      this.loading = true
      try {
        const response = await this.$http.get(`/api/voice/clone/${this.taskId}`)
        console.log('音色克隆任务详情数据:', response.data)
        this.task = {
          ...this.task,
          ...response.data.voice_clone
        }
        
        // 获取音频文件
        let audioSuccess = true
        let sampleSuccess = true
        
        if (this.task.prompt_file) {
          try {
            await this.fetchAudioFile(this.task.prompt_file, 'audio')
          } catch (err) {
            console.error('获取原始音频失败:', err)
            audioSuccess = false
          }
        }
        
        if (this.task.sample_file) {
          try {
            await this.fetchAudioFile(this.task.sample_file, 'sample')
          } catch (err) {
            console.error('获取示例音频失败:', err)
            sampleSuccess = false
          }
        }
        
        // 只有在两个音频都失败的情况下才显示错误
        if (!audioSuccess && !sampleSuccess) {
          console.warn('所有音频获取都失败了')
        }
        
        // 如果任务正在处理中，开始定时刷新
        if (this.task.status === 'processing') {
          this.startRefreshTimer()
        } else {
          this.clearRefreshTimer()
        }
      } catch (error) {
        console.error('获取音色克隆任务详情失败', error)
        this.$message.error('获取音色克隆任务详情失败')
      } finally {
        this.loading = false
      }
    },
    
    // 开始定时刷新
    startRefreshTimer() {
      this.clearRefreshTimer()
      this.refreshTimer = setInterval(() => {
        this.fetchTaskDetail()
      }, 5000) // 每5秒刷新一次
    },
    
    // 清除定时刷新
    clearRefreshTimer() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
        this.refreshTimer = null
      }
    },
    
    // 获取音频文件
    async fetchAudioFile(url, type) {
      try {
        if (!url) {
          console.warn(`${type}音频URL为空`)
          return
        }
        
        console.log(`开始获取${type}音频文件:`, url)
        const token = localStorage.getItem('token')
        
        // 检查是否是MP3格式
        const isMp3 = url.toLowerCase().includes('.mp3')
        console.log(`${type}音频格式: ${isMp3 ? 'MP3' : 'WAV/其他'}`)
        
        // 转换URL格式
        let fullUrl = url
        if (!url.startsWith('http') && !url.startsWith('/api/')) {
          fullUrl = `/api/file/view?path=${encodeURIComponent(url)}`
        }
        
        console.log(`处理后的${type}音频URL:`, fullUrl)
        
        const response = await this.$http({
          method: 'get',
          url: fullUrl,
          responseType: 'blob',
          headers: {
            'Authorization': `Bearer ${token}`
          },
          timeout: 30000
        })
        
        if (!response.data || response.data.size === 0) {
          console.error(`${type}音频响应数据为空或大小为0`)
          this.useDirectUrl(url, type)
          return
        }
        
        console.log(`${type}音频获取成功:`, response.data.type, response.data.size, 'bytes')
        
        if (isMp3) {
          // MP3格式文件专用处理
          this.handleMp3Audio(response.data, type)
        } else {
          // WAV和其他格式
          if (type === 'audio') {
            this.audioBlob = response.data
            this.audioDataUrl = ''
          } else {
            this.sampleBlob = response.data
            this.sampleDataUrl = ''
          }
        }
      } catch (error) {
        console.error(`获取${type}音频文件失败:`, error)
        // 尝试直接使用URL
        this.useDirectUrl(url, type)
      }
    },
    
    // 处理MP3格式音频
    handleMp3Audio(blob, type) {
      // 通过FileReader转换为Data URL
      const reader = new FileReader()
      reader.onload = (e) => {
        const dataUrl = e.target.result
        console.log(`${type} MP3音频转换为Data URL成功`)
        
        if (type === 'audio') {
          this.audioDataUrl = dataUrl
        } else {
          this.sampleDataUrl = dataUrl
        }
      }
      reader.onerror = (e) => {
        console.error(`${type} MP3音频转换Data URL失败:`, e)
      }
      reader.readAsDataURL(blob)
    },
    
    // 直接使用URL作为后备方案
    useDirectUrl(url, type) {
      try {
        if (!url) return
        
        // 构造加上token的URL
        let fullUrl = url
        if (!url.startsWith('http') && !url.startsWith('/api/')) {
          fullUrl = `/api/file/view?path=${encodeURIComponent(url)}`
        }
        
        const token = localStorage.getItem('token')
        const urlWithToken = `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}token=${token}`
        
        console.log(`使用直接URL作为${type}音频源:`, urlWithToken)
        
        if (type === 'audio') {
          this.audioDataUrl = urlWithToken
        } else {
          this.sampleDataUrl = urlWithToken
        }
      } catch (error) {
        console.error(`设置${type}直接URL失败:`, error)
      }
    },
    
    // 获取进度信息
    fetchProgress() {
      this.$http.get(`/api/voice/clone/${this.taskId}/progress`)
        .then(response => {
          this.progress = response.data.progress || 0
          this.progressText = response.data.status_message || '正在处理中...'
          
          // 如果进度已完成，刷新任务详情
          if (this.progress >= 100) {
            this.clearProgressTimer()
            setTimeout(() => {
              this.fetchTaskDetail()
            }, 1000)
          }
        })
        .catch(error => {
          console.error('获取进度失败', error)
        })
    },
    
    // 开始进度定时器
    startProgressTimer() {
      this.clearProgressTimer()
      this.progressTimer = setInterval(() => {
        this.fetchProgress()
      }, 5000) // 每5秒更新一次
    },
    
    // 清除进度定时器
    clearProgressTimer() {
      if (this.progressTimer) {
        clearInterval(this.progressTimer)
        this.progressTimer = null
      }
    },
    
    // 手动刷新进度
    refreshProgress() {
      this.fetchProgress()
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
    
    // 返回上一页
    goBack() {
      this.$router.push('/voice-clone')
    },
    
    // 显示添加到音色库对话框
    addToLibrary() {
      this.addToLibraryDialogVisible = true
    },
    
    // 提交添加到音色库
    submitAddToLibrary() {
      this.$refs.libraryForm.validate(valid => {
        if (valid) {
          this.submitting = true
          
          this.$http.post('/api/voice/library/add', {
            voice_clone_id: this.taskId,
            name: this.libraryForm.name,
            description: this.libraryForm.description
          })
            .then(() => {
              this.$message.success('已成功添加到音色库')
              this.addToLibraryDialogVisible = false
            })
            .catch(error => {
              console.error('添加到音色库失败', error)
              this.$message.error('添加到音色库失败')
            })
            .finally(() => {
              this.submitting = false
            })
        }
      })
    },
    
    // 用于TTS
    useForTTS() {
      // 先添加到音色库，然后跳转到TTS页面
      this.$http.post('/api/voice/library/add', {
        voice_clone_id: this.taskId,
        name: this.task.name || '克隆音色',
        description: this.task.description || ''
      })
        .then(response => {
          this.$message.success('已成功添加到音色库')
          // 跳转到TTS页面并传递音色ID
          this.$router.push({
            path: '/tts',
            query: { voice_id: response.data.id }
          })
        })
        .catch(error => {
          console.error('添加到音色库失败', error)
          this.$message.error('添加到音色库失败')
        })
    },
    
    // 重试任务
    retryTask() {
      this.$confirm('确定要重新提交此任务吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
        .then(() => {
          this.loading = true
          this.$http.post(`/api/voice/clone/${this.taskId}/retry`)
            .then(() => {
              this.$message.success('任务已重新提交')
              this.fetchTaskDetail()
            })
            .catch(error => {
              console.error('重试失败', error)
              this.$message.error('重试失败')
            })
            .finally(() => {
              this.loading = false
            })
        })
        .catch(() => {
          this.$message.info('已取消重试')
        })
    },
    handleResize() {
      this.screenWidth = window.innerWidth
    }
  }
}
</script>

<style scoped>
.voice-clone-detail-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  max-width: 95%;
  margin-left: auto;
  margin-right: auto;
  width: 100%;
  box-sizing: border-box;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 500;
  color: #ffffff;
}

/* PC端顶部导航栏 */
.page-header {
  display: none; /* 默认隐藏，只在PC端显示 */
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.voice-clone-detail-container {
  min-height: 100vh;
  padding: 0;
  background: #fff;
  width: 100%;
  max-width: 100%;
  overflow-x: hidden;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000; /* 提高层级，确保不被遮挡 */
}

/* 顶部导航栏 */
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
  z-index: 1001; /* 提高层级，确保在最上层 */
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  color: #fff;
  width: 100%;
  max-width: 100vw;
}

.header-back {
  display: flex;
  align-items: center;
  font-size: 16px;
  cursor: pointer;
  padding: 5px 10px; /* 增加点击区域 */
  border-radius: 4px;
  transition: background-color 0.2s;
}

.header-back:hover {
  background-color: rgba(255, 255, 255, 0.1); /* 鼠标悬停效果 */
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
}

.detail-content-wrapper {
  padding: 0;
  width: 100%;
  max-width: 100%;
  overflow-x: hidden;
  overflow-y: auto;
  position: absolute;
  top: 56px;
  left: 0;
  right: 0;
  bottom: 0;
  box-sizing: border-box;
  -webkit-overflow-scrolling: touch; /* 增强iOS滚动体验 */
}

.detail-content {
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

/* 基本信息简化版 */
.basic-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

.create-time {
  font-size: 12px;
  color: #666;
}

/* 合并后的移动端内容区 */
.mobile-content-view {
  display: none;
  padding: 0;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

.mobile-audio-section {
  padding: 8px 12px;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  margin: 5px 0;
  color: #333;
}

.sample-section {
  margin-top: 12px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

/* 桌面端卡片视图 */
.desktop-content-view {
  display: block;
  padding: 15px;
  max-width: 100%;
  box-sizing: border-box;
}

/* 基本信息卡片 */
.info-card {
  background: #fff;
  border-radius: 4px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  border: 1px solid #ebeef5;
  max-width: 100%;
}

.info-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  max-width: 100%;
}

.info-header-left {
  display: flex;
  flex-direction: column;
}

.info-title {
  font-size: 18px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}

/* 详情卡片 */
.detail-card {
  margin-bottom: 20px;
  border-radius: 4px;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  max-width: 100%;
}

.card-header {
  background-color: #fff;
  padding: 12px 15px;
  border-bottom: 1px solid #ebeef5;
  max-width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.audio-info {
  padding: 20px;
  max-width: 100%;
}

.audio-player {
  width: 100%;
  margin: 3px 0 8px 0;
  height: 36px;
  max-width: 100%;
}

.prompt-text {
  margin-top: 12px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
  max-width: 100%;
  box-sizing: border-box;
}

.prompt-text h4 {
  margin: 0 0 5px 0;
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.prompt-text p {
  margin: 0;
  color: #666;
  font-size: 13px;
  line-height: 1.4;
  padding: 0;
  background: none;
  border: none;
  max-width: 100%;
  word-break: break-word;
}

.result-info {
  padding: 20px;
  max-width: 100%;
}

.sample-audio {
  margin-bottom: 15px;
  max-width: 100%;
}

.sample-audio h4 {
  margin: 0 0 5px 0;
  font-size: 15px;
  color: #333;
  font-weight: 500;
}

/* 错误信息卡 */
.error-card {
  margin: 0;
  padding: 8px 12px;
  background: none;
  border-top: 1px solid #f0f0f0;
  max-width: 100%;
  box-sizing: border-box;
}

.retry-button {
  margin-top: 10px;
  text-align: center;
  max-width: 100%;
}

/* 移动端适配 */
@media screen and (max-width: 768px) {
  /* 去除所有内边距和边距，实现全屏 */
  html, body {
    margin: 0;
    padding: 0;
    overflow-x: hidden;
    width: 100%;
    max-width: 100vw;
    position: fixed;
  }
  
  .voice-clone-detail-container {
    width: 100%;
    max-width: 100vw;
    overflow-x: hidden;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
  }
  
  .mobile-header-bar {
    left: 0; /* 重置移动端左边距 */
    width: 100%;
  }
  
  .detail-content-wrapper {
    padding: 0;
    width: 100%;
    max-width: 100vw;
    overflow-x: hidden;
    overflow-y: auto;
    position: absolute;
    top: 56px;
    left: 0;
    right: 0;
    bottom: 0;
  }
  
  .detail-content {
    max-width: 100vw;
    width: 100%;
    margin: 0;
    overflow-x: hidden;
  }
  
  /* 显示移动端视图，隐藏桌面端视图 */
  .mobile-content-view {
    display: block;
    box-shadow: none;
    margin: 0;
    width: 100%;
    max-width: 100vw;
    overflow-x: hidden;
  }
  
  .desktop-content-view {
    display: none;
  }
  
  /* 简化的基本信息 */
  .basic-info {
    padding: 8px 12px;
    margin: 0;
    width: 100%;
    max-width: 100vw;
    box-sizing: border-box;
    overflow-x: hidden;
  }
  
  .create-time {
    font-size: 12px;
  }
  
  .mobile-audio-section {
    width: 100%;
    max-width: 100vw;
    box-sizing: border-box;
    overflow-x: hidden;
  }
  
  /* 错误卡片适配 */
  .error-card {
    padding: 8px 12px;
    margin: 0;
    box-shadow: none;
    width: 100%;
    max-width: 100vw;
    box-sizing: border-box;
    overflow-x: hidden;
  }
  
  .section-title {
    font-size: 14px;
    margin: 5px 0;
  }
  
  .prompt-text {
    margin-top: 12px;
    padding-top: 8px;
    width: 100%;
    max-width: 100vw;
    box-sizing: border-box;
    overflow-x: hidden;
  }
  
  .prompt-text p {
    font-size: 13px;
    line-height: 1.4;
    width: 100%;
    max-width: 100vw;
    word-break: break-word;
    white-space: normal;
    overflow-x: hidden;
  }
  
  .sample-section {
    margin-top: 8px;
    padding-top: 8px;
    width: 100%;
    max-width: 100vw;
    box-sizing: border-box;
    overflow-x: hidden;
  }
  
  .audio-player {
    margin: 3px 0 5px 0;
    height: 32px;
    width: 100%;
    max-width: 100vw;
  }
  
  .retry-button {
    margin-top: 8px;
    width: 100%;
    max-width: 100vw;
  }
  
  .retry-button .el-button {
    width: 100%;
    height: 36px;
    font-size: 14px;
    border-radius: 2px;
    max-width: 100vw;
  }
  
  /* 去除元素扰动 */
  * {
    transform: none !important;
    transition: none !important;
    animation: none !important;
    max-width: 100vw !important;
    overflow-x: hidden !important;
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
  
  /* 设置全局样式阻止水平滚动 */
  :global(body) {
    overflow-x: hidden !important;
    width: 100% !important;
    position: fixed !important;
    height: 100% !important;
  }
}

.mobile-dialog {
  max-width: 500px;
}

.mobile-dialog :deep(.el-dialog__body) {
  padding: 15px;
}

/* PC端特有样式 */
@media screen and (min-width: 769px) {
  .mobile-header-bar {
    display: none; /* 隐藏移动端头部 */
  }

  .mobile-header-placeholder {
    display: none; /* 隐藏移动端头部占位 */
  }
  
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    max-width: 95%;
    margin-left: auto;
    margin-right: auto;
    width: 100%;
    box-sizing: border-box;
  }
    
  .voice-clone-detail-container {
    padding: 20px;
    position: relative;
    left: 0;
    width: 100%;
    background: transparent;
    overflow-x: hidden;
    max-width: 100vw;
    box-sizing: border-box;
  }
  
  .detail-content-wrapper {
    position: relative;
    top: 0;
    width: 100%;
    padding: 0;
    overflow-x: hidden;
    max-width: 100vw;
    box-sizing: border-box;
  }

  .detail-content {
    background-color: #fff;
    border-radius: 4px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    max-width: 95%;
    margin-left: auto;
    margin-right: auto;
    width: 100%;
    box-sizing: border-box;
  }
  
  /* 调整PC端卡片内边距 */
  .desktop-content-view {
    padding: 20px;
    background-color: #fff;
    box-sizing: border-box;
    width: 100%;
  }
  
  .info-card {
    margin-bottom: 20px;
    background-color: #fff;
    border: none;
    box-shadow: none;
    padding: 0;
    width: 100%;
    box-sizing: border-box;
  }
  
  .info-header {
    border-bottom: 1px solid #ebeef5;
    padding-bottom: 15px;
    margin-bottom: 15px;
    width: 100%;
    box-sizing: border-box;
  }
  
  .detail-card {
    margin-bottom: 20px;
    box-shadow: none;
    border: 1px solid #ebeef5;
    width: 100%;
    box-sizing: border-box;
  }

  .basic-info {
    display: none; /* 隐藏简化版基本信息 */
  }
}
</style>