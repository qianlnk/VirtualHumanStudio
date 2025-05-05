<template>
  <div class="voice-clone-container">
    <div class="page-header">
      <div class="header-left">
        <h2>音色克隆</h2>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="createVoiceClone" icon="el-icon-plus">创建音色克隆任务</el-button>
        <el-button type="text" size="small" class="view-toggle" @click="toggleView">
          <i :class="isCardView ? 'el-icon-menu' : 'el-icon-s-grid'"></i>
          <span class="toggle-text">{{ isCardView ? '列表视图' : '卡片视图' }}</span>
        </el-button>
      </div>
    </div>
    
    <!-- 任务列表（表格视图） -->
    <div v-loading="loading" class="task-list" v-show="!isCardView">
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
            <el-tag :type="getStatusType(scope.row.status)">{{ getStatusText(scope.row.status) }}</el-tag>
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
    
    <!-- 卡片视图（瀑布流） -->
    <div v-show="isCardView" class="card-list" v-loading="loading">
      <el-empty v-if="tasks.length === 0" description="暂无音色克隆任务"></el-empty>
      
      <div v-else class="card-view-content">
        <div class="waterfall-container" ref="cardContainer">
          <div class="task-card" v-for="item in tasks" :key="item.id">
            <div class="task-card-header">
              <h3 class="task-card-title">{{ item.speaker_name }}</h3>
              <el-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
            </div>
            <div class="task-card-content">
              <div class="task-card-info">
                <p class="text-ellipsis"><span class="info-label">提示词:</span> {{ item.prompt_text || '-' }}</p>
                <p><span class="info-label">创建时间:</span> {{ formatDate(item.created_at) }}</p>
              </div>
            </div>
            <div class="task-card-footer">
              <el-button type="text" size="small" class="action-btn" @click="viewDetail(item.id)">查看</el-button>
              <el-button type="text" size="small" class="action-btn" @click="confirmDelete(item.id)">删除</el-button>
            </div>
          </div>
        </div>
        
        <!-- 加载状态区域 -->
        <div class="load-more-container" ref="loadMoreTrigger">
          <template v-if="loadingMore">
            <div class="loading-indicator">
              <i class="el-icon-loading"></i>
              <p>加载中...</p>
            </div>
          </template>
          <template v-else-if="hasMoreData">
            <p>向下滚动加载更多</p>
          </template>
          <template v-else>
            <p>没有更多数据了</p>
          </template>
        </div>
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
import '@/assets/styles/card-view.css'

export default {
  name: 'VoiceClone',
  data() {
    return {
      loading: false,
      submitting: false,
      tasks: [],
      currentPage: 1,
      pageSize: 10,
      cardPageSize: 10,
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
      },
      isCardView: false,
      loadingMore: false,
      hasMoreData: true,
      initialLoaded: false,
      observer: null
    }
  },
  created() {
    // 从本地存储中读取用户偏好的视图模式
    const savedViewMode = localStorage.getItem('voice_clone_view_mode')
    if (savedViewMode) {
      this.isCardView = savedViewMode === 'card'
    }
    
    // 初始加载数据
    this.loadInitialData()
  },
  
  mounted() {
    // 添加滚动事件监听器用于卡片视图加载更多
    window.addEventListener('scroll', this.handleWindowScroll)
  },
  
  beforeDestroy() {
    // 移除滚动事件监听器
    window.removeEventListener('scroll', this.handleWindowScroll)
    
    // 清除IntersectionObserver
    if (this.observer) {
      this.observer.disconnect()
      this.observer = null
    }
  },
  
  methods: {
    // 辅助方法: 调试日志
    debug(...args) {
      console.log('[VoiceClone]', ...args)
    },
    
    // 初始加载数据
    loadInitialData() {
      if (this.initialLoaded) {
        this.debug('已加载初始数据，跳过')
        return
      }
      
      this.debug('加载初始数据')
      this.currentPage = 1
      this.tasks = []
      this.hasMoreData = true
      this.fetchTasks()
      this.initialLoaded = true
    },
    
    // 获取音色克隆任务列表
    fetchTasks(loadMore = false) {
      if (this.loading || (loadMore && this.loadingMore)) {
        this.debug('已有请求进行中，跳过')
        return
      }
      
      if (!loadMore) {
        this.loading = true
      } else {
        this.loadingMore = true
      }
      
      this.debug('请求数据:', '页码=', this.currentPage, '每页数量=', this.isCardView ? this.cardPageSize : this.pageSize)
      
      this.$http.get(`/api/voice/clones`, {
        params: {
          page: this.currentPage,
          size: this.isCardView ? this.cardPageSize : this.pageSize
        }
      })
        .then(response => {
          const newTasks = response.data.voice_clones || []
          this.total = response.data.total || 0
          
          this.debug('获取到新数据:', newTasks.length, '总数:', this.total)
          
          if (loadMore) {
            // 追加新数据
            this.tasks = [...this.tasks, ...newTasks]
          } else {
            // 重置数据
            this.tasks = newTasks
          }
          
          // 判断是否还有更多数据
          this.hasMoreData = this.tasks.length < this.total
          this.debug('当前数据量：', this.tasks.length, '总数：', this.total, '是否还有更多：', this.hasMoreData)
        })
        .catch(error => {
          console.error('获取音色克隆任务列表失败', error)
          this.$message.error('获取音色克隆任务列表失败')
        })
        .finally(() => {
          this.loading = false
          this.loadingMore = false
          
          // 在数据加载完成后重新设置观察者
          if (this.isCardView) {
            this.$nextTick(() => {
              this.setupIntersectionObserver()
            })
          }
        })
    },
    
    // 加载更多数据
    loadMoreTasks() {
      if (this.loadingMore || !this.hasMoreData) {
        this.debug('跳过加载更多:', '加载中=', this.loadingMore, '没有更多数据=', !this.hasMoreData)
        return
      }
      
      this.debug('开始加载更多数据，当前页码：', this.currentPage)
      this.currentPage++
      this.fetchTasks(true)
    },
    
    // 处理滚动事件
    handleWindowScroll() {
      if (!this.isCardView || this.loadingMore || !this.hasMoreData) return
      
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop
      const windowHeight = window.innerHeight
      const documentHeight = Math.max(
        document.body.scrollHeight, document.documentElement.scrollHeight,
        document.body.offsetHeight, document.documentElement.offsetHeight,
        document.body.clientHeight, document.documentElement.clientHeight
      )
      
      // 当滚动到距离底部200px时触发加载
      if (documentHeight - scrollTop - windowHeight < 200) {
        this.debug('窗口滚动触发加载更多')
        this.loadMoreTasks()
      }
    },
    
    // 设置IntersectionObserver
    setupIntersectionObserver() {
      // 如果已经有observer，先断开连接
      if (this.observer) {
        this.observer.disconnect()
        this.observer = null
      }
      
      this.$nextTick(() => {
        // 获取加载更多的触发元素
        const triggerElement = this.$refs.loadMoreTrigger
        if (!triggerElement) {
          this.debug('未找到加载更多触发元素')
          return
        }
        
        this.debug('设置观察者')
        
        // 创建新的IntersectionObserver
        this.observer = new IntersectionObserver((entries) => {
          const entry = entries[0]
          this.debug('intersection事件:', '可见=', entry.isIntersecting, '加载中=', this.loadingMore, '有更多数据=', this.hasMoreData)
          if (entry.isIntersecting && !this.loadingMore && this.hasMoreData) {
            this.debug('观察者触发加载更多')
            this.loadMoreTasks()
          }
        }, {
          root: null,
          threshold: 0,
          rootMargin: '50px'
        })
        
        // 开始观察
        this.observer.observe(triggerElement)
      })
    },
    
    // 切换视图模式（列表/卡片）
    toggleView() {
      this.isCardView = !this.isCardView
      localStorage.setItem('voice_clone_view_mode', this.isCardView ? 'card' : 'list')
      
      // 重置状态
      this.currentPage = 1
      this.tasks = []
      this.hasMoreData = true
      
      // 重新加载第一页数据
      this.fetchTasks()
      
      // 如果切换到卡片视图，设置IntersectionObserver用于无限滚动
      if (this.isCardView) {
        this.$nextTick(() => {
          this.setupIntersectionObserver()
        })
      }
    },
    
    // 分页处理
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchTasks()
    },
    
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchTasks()
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
    
    formatDate(dateString) {
      if (!dateString) return ''
      return new Date(dateString).toLocaleString()
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
  padding: 15px;
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #fff;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.header-left, .header-right {
  display: flex;
  align-items: center;
}

.header-right {
  gap: 10px;
}

.page-header h2 {
  font-size: 1.4em;
  margin: 0;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.view-toggle {
  margin-left: 10px;
}

.task-list {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 15px;
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

/* 卡片视图相关样式 */
.card-list {
  position: relative;
  min-height: 300px;
}

.card-view-content {
  display: flex;
  flex-direction: column;
  min-height: 300px;
}

.waterfall-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap:15px;
  margin-bottom: 30px;
  width: 100%;
}

@media screen and (min-width: 768px) {
  .waterfall-container {
    grid-template-columns: repeat(2, minmax(320px, 1fr));
    gap: 40px;
  }
}

@media screen and (min-width: 1200px) {
  .waterfall-container {
    grid-template-columns: repeat(3, minmax(320px, 1fr));
    gap: 40px;
  }
}

@media screen and (min-width: 1600px) {
  .waterfall-container {
    grid-template-columns: repeat(4, minmax(320px, 1fr));
    gap: 40px;
  }
}

.task-card {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  overflow: hidden;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  height: 100%;
  display: flex;
  flex-direction: column;
  width: 100%;
  min-width: auto;
  margin-bottom: 0;
}

.task-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  border-color: rgba(255, 255, 255, 0.2);
}

.task-card-header {
  padding: 10px 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 0, 0, 0.2);
}

.task-card-title {
  margin: 0;
  font-size: 14px;
  color: #fff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
  background: linear-gradient(120deg, #e6f7ff, #1890ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  max-width: 65%;
}

.task-card-content {
  padding: 10px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 90px;
}

.task-card-info {
  margin-bottom: 10px;
  overflow: hidden;
}

.task-card-info p {
  margin: 6px 0;
  font-size: 13px;
  color: #ddd;
}

.text-ellipsis {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-word;
  max-height: 60px;
}

.info-label {
  color: #aaa;
  margin-right: 5px;
}

.task-card-footer {
  padding: 10px;
  display: flex;
  justify-content: space-around;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 0, 0, 0.1);
  margin-top: auto;
}

.action-btn {
  padding: 4px 8px;
  margin: 0 2px;
  border-radius: 4px;
  transition: all 0.3s;
  font-size: 13px;
  color: #1890ff;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateY(-2px);
  color: #fff;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
}

.load-more-container {
  text-align: center;
  padding: 20px 0;
  margin: 20px 0;
  color: #909399;
  font-size: 14px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  clear: both;
  order: 999;
  border: 1px dashed rgba(255, 255, 255, 0.2);
}

.load-more-container p {
  margin: 0;
  padding: 15px 30px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 20px;
  backdrop-filter: blur(5px);
}

.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.loading-indicator i {
  font-size: 24px;
  color: #409EFF;
}

.loading-indicator p {
  margin: 0;
  background: transparent;
  padding: 5px 0;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

/* 响应式样式 */
@media screen and (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .header-right {
    margin-top: 10px;
    width: 100%;
    justify-content: space-between;
  }
  
  .toggle-text {
    display: none;
  }
}
</style>