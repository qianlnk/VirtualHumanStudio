<template>
  <div class="voice-library-container">
    <div class="page-header">
      <h2>音色库</h2>
      <div>
        <el-button type="primary" @click="showUploadDialog">上传音色</el-button>
        <el-button type="primary" @click="refreshVoices">刷新</el-button>
      </div>
    </div>
    
    <!-- 音色列表 -->
    <div v-loading="loading" class="voice-list">
      <el-empty v-if="voices.length === 0" description="暂无音色"></el-empty>
      
      <el-table v-else :data="voices" style="width: 100%">
        <el-table-column prop="name" label="音色名称" width="150"></el-table-column>
        <el-table-column prop="model_name" label="模型名称" width="150"></el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip></el-table-column>
        <el-table-column label="文件" width="180">
          <template slot-scope="scope">
            <el-button 
              type="text" 
              size="small" 
              @click="downloadVoice(scope.row.id, 'model')">
              <i class="el-icon-download"></i> 模型文件
            </el-button>
            <el-button 
              v-if="scope.row.sample_file" 
              type="text" 
              size="small" 
              @click="downloadVoice(scope.row.id, 'sample')">
              <i class="el-icon-download"></i> 试听文件
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="试听" width="100">
          <template slot-scope="scope">
            <el-button 
              type="text" 
              size="small" 
              @click="playAudio(scope.row)" 
              :disabled="!scope.row.sample_file">
              <i class="el-icon-video-play"></i> 试听
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250">
          <template slot-scope="scope">
            <el-button type="text" size="small" @click="useForTTS(scope.row.id)">用于TTS</el-button>
            <el-button type="text" size="small" @click="downloadVoice(scope.row.id)">下载</el-button>
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
    
    <!-- 上传音色对话框 -->
    <el-dialog title="上传音色" :visible.sync="uploadDialogVisible" width="500px">
      <el-form :model="uploadForm" :rules="uploadRules" ref="uploadForm" label-width="100px">
        <el-form-item label="音色名称" prop="name">
          <el-input v-model="uploadForm.name" placeholder="请输入音色名称"></el-input>
        </el-form-item>

        <el-form-item label="模型名称" prop="model_name">
          <el-input v-model="uploadForm.model_name" placeholder="请输入模型名称"></el-input>
        </el-form-item>
        
        <el-form-item label="模型文件" prop="model_file">
          <el-upload
            class="upload-demo"
            action="#"
            :http-request="uploadModelFile"
            :limit="1"
            :file-list="modelFileList"
            :before-upload="beforeUploadModel">
            <el-button size="small" type="primary">点击上传</el-button>
            <div slot="tip" class="el-upload__tip">请上传模型文件</div>
          </el-upload>
        </el-form-item>

        <el-form-item label="试听音频" prop="sample_file">
          <el-upload
            class="upload-demo"
            action="#"
            :http-request="uploadSampleFile"
            :limit="1"
            :file-list="sampleFileList"
            :before-upload="beforeUploadSample">
            <el-button size="small" type="primary">点击上传</el-button>
            <div slot="tip" class="el-upload__tip">可选，支持mp3/wav文件，不超过50MB</div>
          </el-upload>
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input 
            type="textarea" 
            v-model="uploadForm.description" 
            placeholder="请输入音色描述"
            :rows="3">
          </el-input>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="uploadDialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="uploading" @click="submitUpload">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { downloadFile, getAudioUrl } from '@/utils/fileAccess'

export default {
  name: 'VoiceLibrary',
  data() {
    return {
      loading: false,
      uploading: false,
      voices: [],
      currentPage: 1,
      pageSize: 10,
      total: 0,
      currentPlayingId: null,
      uploadDialogVisible: false,
      modelFileList: [],
      sampleFileList: [],
      uploadForm: {
        name: '',
        description: '',
        model_name: '',
        model_file: null,
        sample_file: null
      },
      uploadRules: {
        name: [
          { required: true, message: '请输入音色名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ],
        model_name: [
          { required: true, message: '请输入模型名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ],
        model_file: [
          { required: true, message: '请上传模型文件', trigger: 'change' }
        ]
      }
    }
  },
  computed: {
    baseUrl() {
      return (process.env.VUE_APP_API_URL || '') + '/'
    }
  },

  created() {
    this.fetchVoices()
  },
  methods: {
    // 获取音色列表
    fetchVoices() {
      this.loading = true
      this.$http.get(`/api/voices?page=${this.currentPage}&size=${this.pageSize}`)
        .then(response => {
          this.voices = response.data.voices || []
          this.total = response.data.total || 0
        })
        .catch(error => {
          console.error('获取音色列表失败', error)
          this.$message.error('获取音色列表失败')
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
    
    // 播放音频
    playAudio(voice) {
      if (!voice.sample_file) {
        this.$message.warning('该音色没有示例音频')
        return
      }
      
      const audioPlayer = this.$refs.audioPlayer
      const audioUrl = getAudioUrl(voice.sample_file)
      
      // 如果正在播放同一个音频，则暂停
      if (this.currentPlayingId === voice.id && !audioPlayer.paused) {
        audioPlayer.pause()
        this.currentPlayingId = null
        return
      }
      
      // 播放新的音频
      audioPlayer.src = audioUrl
      audioPlayer.play()
      this.currentPlayingId = voice.id
      
      // 播放完成后重置状态
      audioPlayer.onended = () => {
        this.currentPlayingId = null
      }
    },
    
    // 用于TTS
    useForTTS(voiceId) {
      this.$router.push({
        path: '/tts',
        query: { voice_id: voiceId }
      })
    },
    
    // 确认删除
    confirmDelete(id) {
      this.$confirm('确定要删除这个音色吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
        .then(() => {
          this.deleteVoice(id)
        })
        .catch(() => {
          this.$message.info('已取消删除')
        })
    },
    
    // 删除音色
    deleteVoice(id) {
      this.loading = true
      this.$http.delete(`/api/voice/${id}`)
        .then(() => {
          this.$message.success('删除成功')
          this.fetchVoices()
        })
        .catch(error => {
          console.error('删除音色失败', error)
          this.$message.error('删除音色失败')
        })
        .finally(() => {
          this.loading = false
        })
    },
    
    // 刷新音色列表
    refreshVoices() {
      this.fetchVoices()
    },
    
    // 处理每页显示数量变化
    handleSizeChange(size) {
      this.pageSize = size
      this.fetchVoices()
    },
    
    // 处理页码变化
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchVoices()
    },
    
    // 显示上传对话框
    showUploadDialog() {
      this.uploadDialogVisible = true
      this.uploadForm = {
        name: '',
        description: '',
        model_name: '',
        model_file: null,
        sample_file: null
      }
      this.modelFileList = []
      this.sampleFileList = []
    },
    
    // 上传前检查模型文件
    beforeUploadModel(file) {
      this.uploadForm.model_file = file
      return false
    },
    
    // 上传前检查试听音频文件
    beforeUploadSample(file) {
      const isAudio = file.type.startsWith('audio/') || file.name.endsWith('.mp3') || file.name.endsWith('.wav')
      const isLt50M = file.size / 1024 / 1024 < 50
      
      if (!isAudio) {
        this.$message.error('只能上传mp3/wav格式的音频文件!')
        return false
      }
      
      if (!isLt50M) {
        this.$message.error('音频文件大小不能超过 50MB!')
        return false
      }
      
      this.uploadForm.sample_file = file
      return false
    },
    
    // 上传模型文件
    uploadModelFile(options) {
      this.uploadForm.model_file = options.file
      this.modelFileList = [{ name: options.file.name, url: '' }]
    },
    
    // 上传试听音频文件
    uploadSampleFile(options) {
      this.uploadForm.sample_file = options.file
      this.sampleFileList = [{ name: options.file.name, url: '' }]
    },
    
    // 提交上传表单
    submitUpload() {
      this.$refs.uploadForm.validate(valid => {
        if (!valid) {
          return false
        }
        
        if (!this.uploadForm.model_file) {
          this.$message.error('请上传模型文件')
          return false
        }
        
        // 创建FormData对象
        const formData = new FormData()
        formData.append('name', this.uploadForm.name)
        formData.append('model_name', this.uploadForm.model_name)
        formData.append('description', this.uploadForm.description || '')
        formData.append('model_file', this.uploadForm.model_file)
        if (this.uploadForm.sample_file) {
          formData.append('sample_file', this.uploadForm.sample_file)
        }
        
        this.uploading = true
        
        // 发送上传请求
        this.$http.post('/api/voice/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
          .then(() => {
            this.$message.success('上传成功')
            this.uploadDialogVisible = false
            this.fetchVoices()
          })
          .catch(error => {
            console.error('上传失败', error)
            this.$message.error('上传失败: ' + ((error.response && error.response.data && error.response.data.error) || '未知错误'))
          })
          .finally(() => {
            this.uploading = false
          })
      })
    },
    
    // 获取文件扩展名
    getFileExtension(filePath) {
      if (!filePath) return ''
      const match = filePath.match(/\.[^.]+$/)
      return match ? match[0] : ''
    },

    // 下载音色文件
    downloadVoice(id, type) {
      const voice = this.voices.find(v => v.id === id)
      if (!voice) return
      
      const filePath = type === 'model' ? voice.model_file : voice.sample_file
      const fileName = type === 'model' ? `${voice.name}_model${this.getFileExtension(filePath)}` : `${voice.name}_sample${this.getFileExtension(filePath)}`
      downloadFile(filePath, fileName)
    }
  }
}
</script>

<style scoped>
.voice-library-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 22px;
  color: #303133;
}

.voice-list {
  margin-top: 20px;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}
</style>