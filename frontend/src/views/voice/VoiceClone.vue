<template>
  <div class="voice-clone-container">
    <div class="page-header">
      <h2>音色克隆</h2>
      <el-button type="primary" @click="createVoiceClone">创建音色克隆任务</el-button>
    </div>
    
    <!-- 任务列表 -->
    <div v-loading="loading" class="task-list">
      <el-empty v-if="tasks.length === 0" description="暂无音色克隆任务"></el-empty>
      
      <el-table v-else :data="tasks" style="width: 100%">
        <el-table-column prop="speaker_name" label="说话人名称" width="150"></el-table-column>
        <el-table-column prop="prompt_text" label="提示词" show-overflow-tooltip>
          <template slot-scope="scope">
            <span>{{ scope.row.prompt_text || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template slot-scope="scope">
            <el-button type="text" size="small" @click="viewDetail(scope.row.id)">查看</el-button>
            <el-button type="text" size="small" @click="confirmDelete(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination-container" v-if="total > pageSize">
        <el-pagination
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          :current-page="currentPage"
          :page-sizes="[10, 20, 30, 50]"
          :page-size="pageSize"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total">
        </el-pagination>
      </div>
    </div>
    
    <!-- 音色克隆创建表单 -->
    <el-dialog title="创建音色克隆任务" :visible.sync="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="form" label-width="100px">
        <el-form-item label="说话人名称" prop="speaker_name">
          <el-input v-model="form.speaker_name" placeholder="请输入说话人名称"></el-input>
        </el-form-item>
        
        <el-form-item label="音频文件" prop="audio_file">
          <div class="audio-upload-container">
            <el-upload
              class="upload-demo"
              action="#"
              :http-request="uploadAudio"
              :limit="1"
              :file-list="fileList"
              :before-upload="beforeUpload">
              <el-button size="small" type="primary">点击上传</el-button>
            </el-upload>
            <el-button size="small" type="success" @click="startRecording" v-if="!isRecording">录制</el-button>
            <el-button size="small" type="danger" @click="stopRecording" v-if="isRecording">停止录制</el-button>
            <div slot="tip" class="el-upload__tip">只能上传mp3/wav文件，且不超过50MB</div>
          </div>
          <!-- 录音预览 -->
          <div v-if="recordedAudio" class="recorded-audio-preview">
            <audio :src="recordedAudioUrl" controls></audio>
            <div class="preview-actions">
              <el-button size="small" type="primary" @click="useRecordedAudio">使用录制的音频</el-button>
              <el-button size="small" @click="discardRecordedAudio">放弃</el-button>
            </div>
          </div>
        </el-form-item>
        
        <el-form-item label="提示文本" prop="prompt_text">
          <el-input 
            type="textarea" 
            v-model="form.prompt_text" 
            placeholder="请输入提示文本"
            :rows="3">
          </el-input>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  name: 'VoiceClone',
  data() {
    return {
      loading: false,
      submitting: false,
      tasks: [],
      currentPage: 1,
      pageSize: 10,
      total: 0,
      dialogVisible: false,
      fileList: [],
      form: {
        speaker_name: '',
        audio_file: null,
        prompt_text: ''
      },
      // 录音相关数据
      isRecording: false,
      mediaRecorder: null,
      recordedChunks: [],
      recordedAudio: null,
      recordedAudioUrl: null,
      rules: {
        speaker_name: [
          { required: true, message: '请输入说话人名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ],
        audio_file: [
          { required: true, message: '请上传音频文件', trigger: 'change' }
        ],
        prompt_text: [
          { required: true, message: '请输入提示文本', trigger: 'blur' }
        ]
      }
    }
  },
  created() {
    this.fetchTasks()
  },
  methods: {
    // 获取音色克隆任务列表
    fetchTasks() {
      this.loading = true
      this.$http.get(`/api/voice/clones?page=${this.currentPage}&size=${this.pageSize}`)
        .then(response => {
          this.tasks = response.data.voice_clones || []
          this.total = response.data.total || 0
        })
        .catch(error => {
          console.error('获取音色克隆任务列表失败', error)
          this.$message.error('获取音色克隆任务列表失败')
        })
        .finally(() => {
          this.loading = false
        })
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
    
    // 生成随机字符串
    generateRandomString(length = 6) {
      const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
      let result = ''
      for (let i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * characters.length))
      }
      return result
    },
    
    // 创建音色克隆任务
    createVoiceClone() {
      this.dialogVisible = true
      this.form = {
        speaker_name: '',
        audio_file: null,
        prompt_text: ''
      }
      this.fileList = []
    },
    
    // 提交表单
    submitForm() {
      this.$refs.form.validate(valid => {
        if (valid) {
          if (!this.form.audio_file) {
            return this.$message.error('请上传音频文件')
          }
          
          this.submitting = true
          
          const formData = new FormData()
          // 生成任务名称：说话人名称 + 随机字符串
          const taskName = `${this.form.speaker_name}_${this.generateRandomString()}`
          formData.append('name', taskName)
          formData.append('prompt_file', this.form.audio_file)
          formData.append('prompt_text', this.form.prompt_text)
          formData.append('speaker_name', this.form.speaker_name)
          
          this.$http.post('/api/voice/clone', formData, {
            headers: {
              'Content-Type': 'multipart/form-data'
            }
          })
            .then(response => {
              this.$message.success('创建音色克隆任务成功')
              this.dialogVisible = false
              this.fetchTasks()
              
              // 跳转到详情页
              this.$router.push(`/voice-clone/${response.data.voice_clone.id}`)
            })
            .catch(error => {
              console.error('创建音色克隆任务失败', error)
              this.$message.error('创建音色克隆任务失败: ' + ((error.response && error.response.data && error.response.data.message) || '未知错误'))
            })
            .finally(() => {
              this.submitting = false
            })
        }
      })
    },
    
    // 查看详情
    viewDetail(id) {
      this.$router.push(`/voice-clone/${id}`)
    },
    
    // 确认删除
    confirmDelete(id) {
      this.$confirm('此操作将永久删除该音色克隆任务, 是否继续?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
        .then(() => {
          this.deleteTask(id)
        })
        .catch(() => {
          this.$message.info('已取消删除')
        })
    },
    
    // 删除任务
    deleteTask(id) {
      this.$http.delete(`/api/voice/clone/${id}`)
        .then(() => {
          this.$message.success('删除成功')
          this.fetchTasks()
        })
        .catch(error => {
          console.error('删除失败', error)
          this.$message.error('删除失败')
        })
    },
    
    // 处理页码变化
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchTasks()
    },
    
    // 处理每页显示数量变化
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchTasks()
    },

    // 上传前验证
    beforeUpload(file) {
      // 检查文件类型和扩展名
      const validMimeTypes = ['audio/mpeg', 'audio/mp3', 'audio/wav', 'audio/x-wav']
      const isAudioType = validMimeTypes.includes(file.type)
      const fileName = file.name.toLowerCase()
      const isValidExtension = fileName.endsWith('.mp3') || fileName.endsWith('.wav')
      const isLt50M = file.size / 1024 / 1024 < 50

      if (!isAudioType || !isValidExtension) {
        let errorMsg = '只能上传MP3/WAV格式的音频文件！'
        if (!isAudioType) {
          errorMsg += `\n检测到的文件类型: ${file.type || '未知'}`
        }
        this.$message.error(errorMsg)
        return false
      }
      if (!isLt50M) {
        this.$message.error('音频文件大小不能超过50MB！')
        return false
      }
      return true
    },

    // 自定义上传
    uploadAudio(params) {
      const file = params.file
      const isValid = this.beforeUpload(file)
      if (isValid) {
        this.form.audio_file = file
        this.fileList = [{ name: file.name, url: URL.createObjectURL(file) }]
      }
    },
    
    // 开始录音
    async startRecording() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
        this.mediaRecorder = new MediaRecorder(stream)
        this.recordedChunks = []
        
        this.mediaRecorder.ondataavailable = (event) => {
          if (event.data.size > 0) {
            this.recordedChunks.push(event.data)
          }
        }
        
        this.mediaRecorder.onstop = () => {
          const blob = new Blob(this.recordedChunks, { type: 'audio/wav' })
          this.recordedAudio = blob
          this.recordedAudioUrl = URL.createObjectURL(blob)
          // 停止所有音轨
          stream.getTracks().forEach(track => track.stop())
        }
        
        this.mediaRecorder.start()
        this.isRecording = true
      } catch (error) {
        console.error('录音失败:', error)
        this.$message.error('无法访问麦克风，请确保已授予麦克风权限')
      }
    },
    
    // 停止录音
    stopRecording() {
      if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
        this.mediaRecorder.stop()
        this.isRecording = false
      }
    },
    
    // 使用录制的音频
    useRecordedAudio() {
      if (this.recordedAudio) {
        const file = new File([this.recordedAudio], 'recorded_audio.wav', { type: 'audio/wav' })
        this.form.audio_file = file
        this.fileList = [{ name: file.name, url: this.recordedAudioUrl }]
        this.discardRecordedAudio()
      }
    },
    
    // 放弃录制的音频
    discardRecordedAudio() {
      if (this.recordedAudioUrl) {
        URL.revokeObjectURL(this.recordedAudioUrl)
      }
      this.recordedAudio = null
      this.recordedAudioUrl = null
      this.recordedChunks = []
    },
  }
}
</script>

<style scoped>
.voice-clone-container {
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

.task-list {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 20px;
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  margin-top: 20px;
}

.pagination-container {
  margin-top: 20px;
  text-align: center;
}

.audio-file-info {
  margin-top: 10px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.audio-file-info p {
  margin: 5px 0;
  color: #606266;
  font-size: 14px;
}

.audio-upload-container {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.recorded-audio-preview {
  margin-top: 10px;
  padding: 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #f5f7fa;
}

.recorded-audio-preview audio {
  width: 100%;
  margin-bottom: 10px;
}

.preview-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}
</style>