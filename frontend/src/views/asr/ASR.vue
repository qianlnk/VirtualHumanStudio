<template>
  <div class="asr-container">
    <div class="page-header">
      <h2>语音识别</h2>
      <el-button type="primary" @click="handleUpload">创建语音识别任务</el-button>
    </div>

    <!-- 任务列表 -->
    <div v-loading="loading" class="task-list">
      <el-empty v-if="tasks.length === 0" description="暂无语音识别任务"></el-empty>
      
      <el-table v-else :data="tasks" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="name" label="任务名称" width="150"></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ getStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="output_text" label="识别结果" show-overflow-tooltip>
          <template slot-scope="scope">
            <span>{{ scope.row.output_text || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300">
          <template slot-scope="scope">
            <el-button 
              type="text" 
              size="small" 
              @click="viewDetail(scope.row.id)"
            >
              <i class="el-icon-view"></i> 查看
            </el-button>
            <el-button 
              type="text" 
              size="small" 
              @click="playAudio(scope.row)" 
              :disabled="!scope.row.input_file || scope.row.status !== 'completed'"
            >
              <i class="el-icon-video-play"></i> 播放
            </el-button>
            <el-button 
              v-if="scope.row.status === 'failed'" 
              type="text" 
              size="small" 
              @click="handleRetry(scope.row)"
            >重试</el-button>
            <el-button type="text" size="small" @click="handleDelete(scope.row)">删除</el-button>
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

    <!-- 音频播放器 -->
    <audio ref="audioPlayer" style="display: none"></audio>
    
    <!-- 创建任务对话框 -->
    <el-dialog title="创建语音识别任务" :visible.sync="uploadDialogVisible" width="50%" :close-on-click-modal="false">
      <el-form :model="asrForm" :rules="asrRules" ref="asrForm" label-width="100px" class="asr-form">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="asrForm.name" placeholder="请输入任务名称"></el-input>
        </el-form-item>
        <el-form-item label="音频来源" prop="audioSource">
          <el-radio-group v-model="asrForm.audioSource">
            <el-radio label="file">上传文件</el-radio>
            <el-radio label="url">音频URL</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="asrForm.audioSource === 'file'" label="音频文件" prop="audioFile">
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
            <div slot="tip" class="el-upload__tip">只能上传mp3/wav/m4a格式的音频文件，且不超过50MB</div>
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
        <el-form-item v-else label="音频URL" prop="audioUrl">
          <el-input v-model="asrForm.audioUrl" placeholder="请输入音频文件URL"></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="uploadDialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitASRTask">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getAudioUrl } from '@/utils/fileAccess'

export default {
  name: 'ASR',
  data() {
    return {
      loading: false,
      tasks: [],
      currentPage: 1,
      pageSize: 10,
      total: 0,
      uploadDialogVisible: false,
      uploadHeaders: {
        Authorization: `Bearer ${localStorage.getItem('token')}`
      },
      asrForm: {
        name: '',
        audioSource: 'file',
        audioUrl: '',
        audioFile: null
      },
      // 文件上传相关
      fileList: [],
      // 录音相关数据
      isRecording: false,
      mediaRecorder: null,
      recordedChunks: [],
      recordedAudio: null,
      recordedAudioUrl: null,
      asrRules: {
        name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
        audioSource: [{ required: true, message: '请选择音频来源', trigger: 'change' }],
        audioFile: [{ required: true, message: '请上传音频文件', trigger: 'change' }],
        audioUrl: [{ 
          required: true, 
          message: '请输入音频URL', 
          trigger: 'blur',
          validator: (rule, value, callback) => {
            if (this.asrForm.audioSource === 'url' && !value) {
              callback(new Error('请输入音频URL'));
            } else {
              callback();
            }
          }
        }]
      }
    }
  },
  created() {
    this.fetchTasks()
  },
  methods: {
    // 获取任务列表
    async fetchTasks() {
      this.loading = true
      try {
        const response = await this.$http.get('/api/asr', {
          params: {
            page: this.currentPage,
            size: this.pageSize
          }
        })
        console.log('ASR任务列表响应:', response.data)
        
        if (response.data && response.data.items) {
          this.tasks = response.data.items
          this.total = response.data.total
          console.log('ASR任务列表:', this.tasks)
        } else {
          console.error('ASR任务列表数据格式不正确:', response.data)
          this.$message.warning('获取任务列表数据格式不正确')
          this.tasks = []
          this.total = 0
        }
      } catch (error) {
        console.error('获取ASR任务列表失败:', error)
        this.$message.error('获取任务列表失败: ' + (error.message || '未知错误'))
      } finally {
        this.loading = false
      }
    },
    


    // 状态类型
    getStatusType(status) {
      const statusMap = {
        pending: 'info',
        processing: 'warning',
        completed: 'success',
        failed: 'danger'
      }
      return statusMap[status] || 'info'
    },

    // 状态文本
    getStatusText(status) {
      const statusMap = {
        pending: '等待中',
        processing: '处理中',
        completed: '已完成',
        failed: '失败'
      }
      return statusMap[status] || status
    },
    
    // 格式化日期
    formatDate(dateString) {
      if (!dateString) return ''
      return new Date(dateString).toLocaleString()
    },

    // 处理上传
    handleUpload() {
      this.uploadDialogVisible = true
      this.asrForm = {
        name: '',
        audioSource: 'file',
        audioUrl: '',
        audioFile: null
      }
      this.fileList = []
      this.discardRecordedAudio()
    },

    // 上传前验证
    beforeUpload(file) {
      // 检查文件类型和扩展名
      const validMimeTypes = ['audio/mpeg', 'audio/mp3', 'audio/wav', 'audio/x-wav', 'audio/m4a']
      const isAudioType = validMimeTypes.includes(file.type)
      const fileName = file.name.toLowerCase()
      const isValidExtension = fileName.endsWith('.mp3') || fileName.endsWith('.wav') || fileName.endsWith('.m4a')
      const isLt50M = file.size / 1024 / 1024 < 50

      if (!isAudioType || !isValidExtension) {
        let errorMsg = '只能上传MP3/WAV/M4A格式的音频文件！'
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
        this.asrForm.audioFile = file
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
        this.asrForm.audioFile = file
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

    // 上传成功
    handleUploadSuccess(response) {
      if (response.success) {
        this.$message.success('上传成功')
        this.uploadDialogVisible = false
        this.fetchTasks()
      } else {
        this.$message.error(response.message || '上传失败')
      }
    },

    // 上传失败
    handleUploadError() {
      this.$message.error('上传失败')
    },

    // 重试任务
    async handleRetry(task) {
      try {
        await this.$http.post(`/api/asr/${task.id}/retry`)
        this.$message.success('重试任务已提交')
        this.fetchTasks()
      } catch (error) {
        this.$message.error('重试任务失败')
      }
    },

    // 删除任务
    async handleDelete(task) {
      try {
        await this.$confirm('确定要删除这个任务吗？')
        await this.$http.delete(`/api/asr/${task.id}`)
        this.$message.success('删除成功')
        this.fetchTasks()
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('删除失败')
        }
      }
    },

    // 分页大小改变
    handleSizeChange(val) {
      this.pageSize = val
      this.fetchTasks()
    },

    // 当前页改变
    handleCurrentChange(val) {
      this.currentPage = val
      this.fetchTasks()
    },
    async handleFileChange(file) {
      if (!file) return;
      this.asrForm.audioFile = file.raw;
    },
    
    // 播放音频
    async playAudio(task) {
      if (!task.input_file) {
        this.$message.warning('没有可播放的音频文件');
        return;
      }
      
      try {
        const audioPlayer = this.$refs.audioPlayer;
        audioPlayer.src = await getAudioUrl(task.input_file);
        await audioPlayer.play();
      } catch (error) {
        console.error('音频播放失败:', error);
        this.$message.error('音频播放失败');
      }
    },
    // 查看详情
    viewDetail(id) {
      this.$router.push(`/speech2text/${id}`);
    },
    
    async submitASRTask() {
      try {
        await this.$refs.asrForm.validate();
        
        const formData = new FormData();
        formData.append('name', this.asrForm.name);
        
        if (this.asrForm.audioSource === 'file' && this.asrForm.audioFile) {
          formData.append('audio_file', this.asrForm.audioFile);
        } else if (this.asrForm.audioSource === 'url') {
          formData.append('audio_url', this.asrForm.audioUrl);
        } else {
          throw new Error('请提供音频文件或URL');
        }
        
        const response = await this.$http.post('/api/asr', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        });
        
        if (response.data) {
          this.$message.success('任务创建成功');
          this.uploadDialogVisible = false;
          this.fetchTasks();
          this.$refs.asrForm.resetFields();
          this.fileList = [];
          this.discardRecordedAudio();
        }
      } catch (error) {
        const errorMessage = (error.response && error.response.data && error.response.data.error) || error.message || '创建任务失败';
        this.$message.error(errorMessage);
      }
    }
  }
}
</script>

<style scoped>
.asr-container {
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
}

/* 移除自定义result-text样式，使用Element UI的show-overflow-tooltip属性处理长文本 */

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

/* 自定义上传组件样式 */
.upload-demo {
  display: flex;
  justify-content: center;
}

.el-upload__tip {
  text-align: center;
}

.asr-form {
  padding: 20px;
}

.asr-form .el-form-item {
  margin-bottom: 22px;
}

.asr-form .el-upload {
  width: 100%;
}

.asr-form .el-upload-dragger {
  width: 100%;
  height: 180px;
}

.asr-form .el-radio-group {
  width: 100%;
}

/* 音频上传容器样式 */
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

/* 音频播放器样式 */
.audio-player {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 5px 0;
}

.audio-player audio {
  width: 100%;
  margin-bottom: 5px;
}

.audio-player .el-button {
  margin-top: 5px;
}
</style>