<template>
  <div class="tts-container">
    <div class="page-header">
      <h2>语音合成</h2>
      <el-button type="primary" @click="showCreateDialog">创建语音合成任务</el-button>
    </div>
    
    <!-- 任务列表 -->
    <div v-loading="loading" class="task-list">
      <el-empty v-if="tasks.length === 0" description="暂无语音合成任务"></el-empty>
      
      <el-table v-else :data="tasks" style="width: 100%">
        <el-table-column prop="name" label="任务名称" width="180"></el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ getStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="speaker_name" label="使用音色" width="150"></el-table-column>
        <el-table-column label="操作" width="250">
          <template slot-scope="scope">
            <el-button 
              type="text" 
              size="small" 
              @click="viewDetail(scope.row.id)"
            >查看</el-button>
            <el-button 
              type="text" 
              size="small" 
              @click="playAudio(scope.row)" 
              :disabled="scope.row.status !== 'completed'"
            >
              <i class="el-icon-video-play"></i> 播放
            </el-button>
            <el-button 
              type="text" 
              size="small" 
              @click="downloadOutput(scope.row.id)" 
              :disabled="scope.row.status !== 'completed'"
            >下载</el-button>
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
    
    <!-- 音频播放器 -->
    <audio ref="audioPlayer" style="display: none"></audio>
    
    <!-- 创建TTS任务对话框 -->
    <el-dialog title="创建语音合成任务" :visible.sync="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="form" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称"></el-input>
        </el-form-item>
        
        <el-form-item label="任务类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio label="text2speech">文本转语音</el-radio>
            <el-radio label="speech2text">语音转文本</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <template v-if="form.type === 'text2speech'">
          <el-form-item label="输入文本" prop="input_text">
            <el-input 
              type="textarea" 
              v-model="form.input_text" 
              placeholder="请输入要转换的文本"
              :rows="5">
            </el-input>
          </el-form-item>
          
          <el-form-item label="选择音色" prop="speaker_name">
            <el-select v-model="form.speaker_name" placeholder="请选择音色">
              <el-option 
                v-for="voice in voices" 
                :key="voice.id" 
                :label="voice.name" 
                :value="voice.name">
              </el-option>
            </el-select>
          </el-form-item>
        </template>
        
        <template v-else>
          <el-form-item label="音频文件" prop="input_file">
            <el-upload
              class="upload-demo"
              action="#"
              :http-request="uploadAudio"
              :limit="1"
              :file-list="fileList"
              :before-upload="beforeUpload">
              <el-button size="small" type="primary">点击上传</el-button>
              <div slot="tip" class="el-upload__tip">只能上传mp3/wav文件，且不超过50MB</div>
            </el-upload>
          </el-form-item>
        </template>
        
        <el-form-item label="描述" prop="description">
          <el-input 
            type="textarea" 
            v-model="form.description" 
            placeholder="请输入任务描述"
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
import axios from 'axios'
import { downloadFile, getAudioUrl } from '@/utils/fileAccess'

export default {
  name: 'TTS',
  data() {
    return {
      loading: false,
      submitting: false,
      dialogVisible: false,
      tasks: [],
      voices: [],
      total: 0,
      currentPage: 1,
      pageSize: 10,
      baseURL: process.env.VUE_APP_API_URL || '',
      token: localStorage.getItem('token') || '',
      form: {
        name: '',
        description: '',
        type: 'text2speech',
        input_text: '',
        speaker_name: '',
      },
      rules: {
        name: [
          { required: true, message: '请输入任务名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ],
        type: [
          { required: true, message: '请选择任务类型', trigger: 'change' }
        ],
        input_text: [
          { required: true, message: '请输入要转换的文本', trigger: 'blur' }
        ],
        speaker_name: [
          { required: true, message: '请选择音色', trigger: 'change' }
        ]
      },
      fileList: [],
      audioFile: null
    }
  },
  created() {
    this.fetchTasks()
    this.fetchVoices()
    // 检查是否从音色库页面跳转过来
    if (this.$route.query.voice_id) {
      this.loadVoiceInfo(this.$route.query.voice_id)
    }
  },
  methods: {
    // 获取任务列表
    async fetchTasks() {
      this.loading = true
      try {
        const response = await axios.get(`${this.baseURL}/api/tts`, {
          params: {
            page: this.currentPage,
            size: this.pageSize
          },
          headers: { Authorization: `Bearer ${this.token}` }
        })
        
        this.tasks = response.data.tts_tasks
        this.total = response.data.total
      } catch (error) {
        console.error('获取TTS任务列表失败:', error)
        this.$message.error('获取任务列表失败')
      } finally {
        this.loading = false
      }
    },
    
    // 获取音色列表
    async fetchVoices() {
      try {
        const response = await axios.get(`${this.baseURL}/api/voices`, {
          headers: { Authorization: `Bearer ${this.token}` }
        })
        
        this.voices = response.data.voices
      } catch (error) {
        console.error('获取音色列表失败:', error)
        this.$message.error('获取音色列表失败')
      }
    },
    
    // 加载音色信息（从音色库页面跳转过来时）
    async loadVoiceInfo(voiceId) {
      try {
        // 查找音色信息
        const voice = this.voices.find(v => v.id === parseInt(voiceId))
        if (voice) {
          this.form.speaker_name = voice.name
          this.showCreateDialog()
        }
      } catch (error) {
        console.error('加载音色信息失败:', error)
      }
    },
    
    // 显示创建对话框
    showCreateDialog() {
      this.dialogVisible = true
      this.form = {
        name: '',
        description: '',
        type: 'text2speech',
        input_text: '',
        speaker_name: this.form.speaker_name || ''
      }
      this.fileList = []
      this.audioFile = null
      
      // 重置表单验证
      if (this.$refs.form) {
        this.$refs.form.clearValidate()
      }
    },
    
    // 提交表单
    submitForm() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        
        this.submitting = true
        
        try {
          const formData = new FormData()
          formData.append('name', this.form.name)
          formData.append('description', this.form.description)
          formData.append('type', this.form.type)
          
          if (this.form.type === 'text2speech') {
            formData.append('input_text', this.form.input_text)
            formData.append('speaker_name', this.form.speaker_name)
          } else {
            if (!this.audioFile) {
              this.$message.error('请上传音频文件')
              this.submitting = false
              return
            }
            formData.append('input_file', this.audioFile)
          }
          
          await axios.post(`${this.baseURL}/api/tts`, formData, {
            headers: { 
              Authorization: `Bearer ${this.token}`,
              'Content-Type': 'multipart/form-data'
            }
          })
          
          this.$message.success('创建TTS任务成功')
          this.dialogVisible = false
          this.fetchTasks()
        } catch (error) {
          console.error('创建TTS任务失败:', error)
          this.$message.error((error.response && error.response.data && error.response.data.error) || '创建任务失败')
        } finally {
          this.submitting = false
        }
      })
    },
    
    // 上传音频
    uploadAudio(options) {
      this.audioFile = options.file
    },
    
    // 上传前检查
    beforeUpload(file) {
      const isAudio = file.type === 'audio/mpeg' || file.type === 'audio/wav' || file.type === 'audio/mp3'
      const isLt50M = file.size / 1024 / 1024 < 50
      
      if (!isAudio) {
        this.$message.error('只能上传MP3或WAV格式的音频文件!')
      }
      if (!isLt50M) {
        this.$message.error('音频文件大小不能超过50MB!')
      }
      
      return isAudio && isLt50M
    },
    
    // 查看详情
    viewDetail(id) {
      this.$router.push(`/tts/${id}`)
    },
    
    // 播放音频
    async playAudio(task) {
      if (task.status !== 'completed') {
        this.$message.warning('任务尚未完成，无法播放')
        return
      }
      
      if (!task.output_file) {
        this.$message.warning('没有可播放的音频文件')
        return
      }
      
      const audioPlayer = this.$refs.audioPlayer
      audioPlayer.src = await getAudioUrl(task.output_file)
      await audioPlayer.play()
    },
    
    // 下载输出文件
    async downloadOutput(taskId) {
      try {
        const task = this.tasks.find(t => t.id === taskId)
        const audioPath = await getAudioUrl(task.output_file)
        const fileName = task && task.name ? `${task.name}.mp3` : `tts_output_${taskId}.mp3`
        await downloadFile(audioPath, fileName)
      } catch (error) {
        console.error('下载文件失败:', error)
        this.$message.error('下载文件失败')
      }
    },
    
    // 获取文件扩展名
    getFileExtension(filePath) {
      if (!filePath) return ''
      const match = filePath.match(/\.[^.]+$/)
      return match ? match[0] : ''
    },
    
    // 确认删除
    confirmDelete(id) {
      this.$confirm('此操作将永久删除该任务, 是否继续?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.deleteTask(id)
      }).catch(() => {
        this.$message.info('已取消删除')
      })
    },
    
    // 删除任务
    async deleteTask(id) {
      try {
        await axios.delete(`${this.baseURL}/api/tts/${id}`, {
          headers: { Authorization: `Bearer ${this.token}` }
        })
        
        this.$message.success('删除成功')
        this.fetchTasks()
      } catch (error) {
        console.error('删除任务失败:', error)
        this.$message.error('删除任务失败')
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
    
    // 处理分页大小变化
    handleSizeChange(size) {
      this.pageSize = size
      this.fetchTasks()
    },
    
    // 处理页码变化
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchTasks()
    }
  }
}
</script>

<style scoped>
.tts-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.task-list {
  background-color: #fff;
  padding: 20px;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.el-upload__tip {
  line-height: 1.2;
}
</style>