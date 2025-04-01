<template>
  <div class="voice-clone-detail-container">
    <div class="page-header">
      <div class="header-left">
        <el-button icon="el-icon-back" @click="goBack">返回</el-button>
        <h2>{{ task.speaker_name }}</h2>
      </div>
      <el-tag :type="getStatusType(task.status)">{{ task.status }}</el-tag>
    </div>
    
    <div v-loading="loading" class="detail-content">
      <!-- 基本信息 -->
      <el-card class="detail-card">
        <div slot="header" class="clearfix">
          <span>基本信息</span>
        </div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="任务ID">{{ task.id }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(task.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ task.status }}</el-descriptions-item>
          <el-descriptions-item label="描述">{{ task.description || '无' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>
      
      <!-- 音频信息 -->
      <el-card class="detail-card">
        <div slot="header" class="clearfix">
          <span>原始音频</span>
        </div>
        <div class="audio-info">
          <audio controls :src="audioUrlWithToken" style="width: 100%"></audio>
          <div v-if="task.prompt_text" class="prompt-text">
            <h4>提示文本</h4>
            <p>{{ task.prompt_text }}</p>
          </div>
        </div>
      </el-card>
      
      <!-- 结果信息 -->
      <el-card class="detail-card" v-if="task.status === 'completed'">
        <div slot="header" class="clearfix">
          <span>克隆结果</span>
        </div>
        <div class="result-info">
          <div class="sample-audio" v-if="task.sample_file">
            <h4>示例音频</h4>
            <audio controls :src="sampleUrlWithToken" style="width: 100%"></audio>
          </div>
        </div>
      </el-card>
      
      <!-- 错误信息 -->
      <el-card class="detail-card" v-if="task.status === 'failed'">
        <div slot="header" class="clearfix">
          <span>错误信息</span>
        </div>
        <div class="error-info">
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
      </el-card>
    </div>
    
    <!-- 添加到音色库对话框 -->
    <el-dialog title="添加到音色库" :visible.sync="addToLibraryDialogVisible" width="500px">
      <el-form :model="libraryForm" :rules="libraryRules" ref="libraryForm" label-width="100px">
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
      sampleBlob: null
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
      return this.audioBlob ? URL.createObjectURL(this.audioBlob) : ''
    },
    sampleUrlWithToken() {
      return this.sampleBlob ? URL.createObjectURL(this.sampleBlob) : ''
    }
  },
  created() {
    this.fetchTaskDetail()
  },
  beforeDestroy() {
    this.clearProgressTimer()
    // 清理Blob URL
    if (this.audioBlob) {
      URL.revokeObjectURL(this.audioUrlWithToken)
    }
    if (this.sampleBlob) {
      URL.revokeObjectURL(this.sampleUrlWithToken)
    }
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
        if (this.task.prompt_file) {
          await this.fetchAudioFile(this.task.prompt_file, 'audio')
        }
        if (this.task.sample_file) {
          await this.fetchAudioFile(this.task.sample_file, 'sample')
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
        const token = localStorage.getItem('token')
        const response = await this.$http.get(url, {
          responseType: 'blob',
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })
        if (type === 'audio') {
          this.audioBlob = response.data
        } else {
          this.sampleBlob = response.data
        }
      } catch (error) {
        console.error(`获取${type}音频文件失败`, error)
        this.$message.error(`获取${type}音频文件失败`)
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
}

.header-left {
  display: flex;
  align-items: center;
}

.header-left h2 {
  margin: 0 0 0 15px;
  font-size: 22px;
  color: #303133;
}

.detail-content {
  margin-top: 20px;
}

.detail-card {
  margin-bottom: 20px;
}

.audio-info {
  padding: 10px;
}

.audio-info-text {
  margin-top: 10px;
  color: #606266;
  font-size: 14px;
}

.prompt-text {
  margin-top: 15px;
  border-top: 1px solid #EBEEF5;
  padding-top: 15px;
}

.prompt-text h4 {
  margin: 0 0 10px 0;
  font-size: 16px;
  color: #303133;
}

.prompt-text p {
  margin: 0;
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
}

.progress-info {
  padding: 10px;
}

.progress-text {
  margin-top: 10px;
  color: #606266;
  font-size: 14px;
}

.result-info {
  padding: 10px;
}

.sample-audio {
  margin-bottom: 20px;
}

.sample-audio h4 {
  margin: 0 0 10px 0;
  font-size: 16px;
  color: #303133;
}

.action-buttons {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.error-info {
  padding: 10px;
}

.retry-button {
  margin-top: 20px;
  text-align: center;
}
</style>