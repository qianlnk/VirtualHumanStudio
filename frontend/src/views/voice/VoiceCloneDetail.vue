<template>
  <div class="voice-clone-detail-container">
    <div class="page-header">
      <div class="header-left">
        <el-button icon="el-icon-back" @click="goBack">返回</el-button>
        <h2>{{ task.name }}</h2>
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
          <audio controls :src="baseUrl + task.prompt_file" style="width: 100%"></audio>
          <p class="audio-info-text">音频时长: {{ task.audio_duration || '未知' }} 秒</p>
        </div>
      </el-card>
      
      <!-- 进度信息 -->
      <el-card class="detail-card" v-if="task.status === 'processing'">
        <div slot="header" class="clearfix">
          <span>处理进度</span>
          <el-button style="float: right; padding: 3px 0" type="text" @click="refreshProgress">刷新</el-button>
        </div>
        <div class="progress-info">
          <el-progress :percentage="progress" :status="progressStatus"></el-progress>
          <p class="progress-text">{{ progressText }}</p>
        </div>
      </el-card>
      
      <!-- 结果信息 -->
      <el-card class="detail-card" v-if="task.status === 'completed'">
        <div slot="header" class="clearfix">
          <span>克隆结果</span>
        </div>
        <div class="result-info">
          <div class="sample-audio" v-if="task.sample_url">
            <h4>示例音频</h4>
            <audio controls :src="baseUrl + task.sample_url" style="width: 100%"></audio>
          </div>
          
          <div class="action-buttons">
            <el-button type="primary" @click="addToLibrary">添加到音色库</el-button>
            <el-button type="success" @click="useForTTS">用于TTS</el-button>
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
        progress: 0,
        status_message: ''
      },
      progress: 0,
      progressText: '',
      progressTimer: null,
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
      }
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
    baseUrl() {
      return (process.env.VUE_APP_API_URL || '') + '/'
    }
  },
  created() {
    this.fetchTaskDetail()
  },
  beforeDestroy() {
    this.clearProgressTimer()
  },
  methods: {
    // 获取任务详情
    fetchTaskDetail() {
      this.loading = true
      this.$http.get(`/api/voice/clone/${this.taskId}`)
        .then(response => {
          console.log('音色克隆任务详情数据:', response.data)
          this.task = {
            ...this.task,
            ...response.data.voice_clone
          }
          
          // 如果任务正在处理中，获取进度
          if (this.task.status === 'processing') {
            this.fetchProgress()
            this.startProgressTimer()
          }
          
          // 预填充音色库表单
          if (this.task.name) {
            this.libraryForm.name = this.task.name
            this.libraryForm.description = this.task.description || ''
          }
        })
        .catch(error => {
          console.error('获取音色克隆任务详情失败', error)
          this.$message.error('获取音色克隆任务详情失败')
        })
        .finally(() => {
          this.loading = false
        })
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