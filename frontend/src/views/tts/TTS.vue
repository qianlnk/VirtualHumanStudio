<template>
  <div class="tts-container">
    <div class="page-header">
      <div class="header-left">
        <h2>语音合成</h2>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showCreateDialog" icon="el-icon-plus">创建任务</el-button>
        <el-button type="text" size="small" class="view-toggle" @click="toggleView">
          <i :class="isCardView ? 'el-icon-menu' : 'el-icon-s-grid'"></i>
          <span class="toggle-text">{{ isCardView ? '列表视图' : '卡片视图' }}</span>
        </el-button>
      </div>
    </div>
    
    <!-- 任务列表 -->
    <div v-loading="loading" class="task-list mobile-card-view" v-show="!isCardView">
      <el-empty v-if="tasks.length === 0" description="暂无语音合成任务"></el-empty>
      
      <el-table v-else :data="tasks" style="width: 100%" class="responsive-table">
        <el-table-column prop="input_text" label="输入文本" min-width="250" show-overflow-tooltip>
          <template slot-scope="scope">
            <div class="text-with-copy">
              <span class="text-content">{{ scope.row.input_text || '-' }}</span>
              <el-button 
                v-if="scope.row.input_text" 
                type="primary" 
                size="mini" 
                icon="el-icon-document-copy" 
                class="copy-btn" 
                @click.stop="copyText(scope.row.input_text)">
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" class="hide-on-mobile">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">{{ getStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="使用音色" width="180" class="hide-on-mobile">
          <template slot-scope="scope">
            <div>
              <div>{{ scope.row.speaker_name }}</div>
              <div v-if="scope.row.alias" style="color: #909399; font-size: 12px;">{{ scope.row.alias }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template slot-scope="scope">
            <div class="action-buttons">
              <el-button 
                type="text" 
                size="mini" 
                class="action-btn"
                @click="viewDetail(scope.row.id)"
              >查看</el-button>
              <el-button 
                type="text" 
                size="mini" 
                class="action-btn"
                @click="playAudio(scope.row)" 
                :disabled="scope.row.status !== 'completed'"
              >
                <i class="el-icon-video-play"></i>
              </el-button>
              <el-button 
                type="text" 
                size="mini" 
                class="action-btn"
                @click="downloadOutput(scope.row.id)" 
                :disabled="scope.row.status !== 'completed'"
              >下载</el-button>
              <el-button type="text" size="mini" class="action-btn" @click="confirmDelete(scope.row.id)">删除</el-button>
            </div>
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
      <el-empty v-if="tasks.length === 0" description="暂无语音合成任务"></el-empty>
      
      <div v-else class="card-view-content">
        <div class="waterfall-container" ref="cardContainer">
          <div class="task-card" v-for="item in tasks" :key="item.id">
            <div class="task-card-header">
              <h3 class="task-card-title">{{ item.speaker_name }}</h3>
              <el-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
            </div>
            <div class="task-card-content">
              <div class="task-card-info">
                <div class="text-with-copy card-text">
                  <p class="text-ellipsis">{{ item.input_text || '-' }}</p>
                  <el-button 
                    v-if="item.input_text" 
                    type="primary" 
                    size="mini" 
                    icon="el-icon-document-copy" 
                    class="copy-btn" 
                    @click.stop="copyText(item.input_text)">
                  </el-button>
                </div>
                <p><i class="el-icon-time"></i> {{ formatDate(item.created_at) }}</p>
              </div>
            </div>
            <div class="task-card-footer">
              <el-button type="text" size="small" class="action-btn" @click="viewDetail(item.id)">查看</el-button>
              <el-button 
                type="text" 
                size="small" 
                class="action-btn" 
                @click="playAudio(item)" 
                :disabled="item.status !== 'completed'"
              >
                <i class="el-icon-video-play"></i>
              </el-button>
              <el-button 
                type="text" 
                size="small" 
                class="action-btn" 
                @click="downloadOutput(item.id)" 
                :disabled="item.status !== 'completed'"
              >下载</el-button>
              <el-button type="text" size="small" class="action-btn" @click="confirmDelete(item.id)">删除</el-button>
            </div>
          </div>
        </div>

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
    
    <!-- 移动端底部菜单 -->
    <div class="mobile-footer-menu">
      <div class="menu-item" @click="isCardView = false">
        <i class="el-icon-menu"></i>
        <span>列表</span>
      </div>
      <div class="menu-item" @click="isCardView = true">
        <i class="el-icon-s-grid"></i>
        <span>卡片</span>
      </div>
      <div class="menu-item" @click="showCreateDialog">
        <i class="el-icon-plus"></i>
        <span>创建</span>
      </div>
    </div>
    
    <!-- 音频播放器 -->
    <audio ref="audioPlayer" style="display: none"></audio>
    
    <!-- 创建TTS任务对话框 -->
    <el-dialog title="创建语音合成任务" :visible.sync="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="form" label-width="100px">
        <el-form-item label="选择音色" prop="speaker_names">
          <el-select v-model="form.speaker_names" multiple placeholder="请选择音色（可多选）" style="width: 100%;">
            <el-option 
              v-for="voice in voices" 
              :key="voice.id" 
              :label="voice.name + (voice.alias ? ` (${voice.alias})` : '') + (voice.gender ? ` [${voice.gender}]` : '')" 
              :value="voice.name">
              <span>{{ voice.name }}</span>
              <span v-if="voice.alias" style="color: #909399; font-size: 12px;"> ({{ voice.alias }})</span>
              <span v-if="voice.gender" :style="{ color: voice.gender === '男' ? '#409EFF' : '#E6A23C', marginLeft: '8px' }">[{{ voice.gender }}]</span>
            </el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="输入文本" prop="input_text">
          <el-input 
            type="textarea" 
            v-model="form.input_text" 
            placeholder="请输入要转换的文本"
            :rows="5">
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
import { v4 as uuidv4 } from 'uuid'
import '@/assets/styles/card-view.css'

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
      cardPageSize: 10,
      isCardView: false,
      baseURL: process.env.VUE_APP_API_URL || '',
      token: localStorage.getItem('token') || '',
      form: {
        input_text: '',
        speaker_names: [],
      },
      loadingMore: false,
      hasMoreData: true,
      initialLoaded: false,
      scrollThreshold: 200,
      rules: {
        input_text: [
          { required: true, message: '请输入要转换的文本', trigger: 'blur' }
        ],
        speaker_names: [
          { required: true, message: '请选择音色', trigger: 'change' },
          { type: 'array', min: 1, message: '请至少选择一个音色', trigger: 'change' }
        ]
      },
      fileList: [],
      audioFile: null
    }
  },
  created() {
    // 从本地存储中读取用户偏好的视图模式
    const savedViewMode = localStorage.getItem('tts_view_mode')
    if (savedViewMode) {
      this.isCardView = savedViewMode === 'card'
    }
    
    // 加载音色列表
    this.fetchVoices()
    
    // 检查是否从音色库页面跳转过来
    if (this.$route.query.voice_id && this.$route.query.voice_name) {
      this.loading = true
      this.form.speaker_name = this.$route.query.voice_name
      this.$nextTick(() => {
        this.dialogVisible = true
        this.loading = false
      })
    }
  },
  
  mounted() {
    this.debug('组件挂载，开始加载初始数据')
    // 初始加载第一页数据
    this.loadInitialData()
    
    // 添加滚动事件监听器
    window.addEventListener('scroll', this.handleWindowScroll)
  },
  
  beforeDestroy() {
    // 移除滚动事件监听器
    window.removeEventListener('scroll', this.handleWindowScroll)
    window.onscroll = null
    
    // 清除IntersectionObserver
    if (this.observer) {
      this.observer.disconnect()
      this.observer = null
    }
  },
  methods: {
    // 辅助方法: 调试日志
    debug(...args) {
      console.log('[TTS]', ...args)
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
    
    // 获取任务列表
    async fetchTasks(loadMore = false) {
      if (this.loading || (loadMore && this.loadingMore)) {
        this.debug('已有请求进行中，跳过', '加载中=', this.loading, '加载更多中=', this.loadingMore)
        return
      }
      
      if (!loadMore) {
        this.loading = true
      } else {
        this.loadingMore = true
      }
      
      try {
        this.debug('请求数据:', '页码=', this.currentPage, '每页数量=', this.isCardView ? this.cardPageSize : this.pageSize)
        const response = await axios.get(`${this.baseURL}/api/tts`, {
          params: {
            page: this.currentPage,
            size: this.isCardView ? this.cardPageSize : this.pageSize
          },
          headers: { Authorization: `Bearer ${this.token}` }
        })
        
        const newTasks = response.data.tts_tasks || []
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
        
      } catch (error) {
        console.error('获取TTS任务列表失败:', error)
        this.$message.error('获取任务列表失败')
      } finally {
        this.loading = false
        this.loadingMore = false
        
        // 在数据加载完成后重新设置观察者
        if (this.isCardView) {
          this.$nextTick(() => {
            this.setupIntersectionObserver()
          })
        }
      }
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
        this.debug('窗口滚动触发加载更多:', '文档高度=', documentHeight, '滚动位置=', scrollTop, '窗口高度=', windowHeight)
        this.loadMoreTasks()
      }
    },
    
    // 获取音色列表
    async fetchVoices() {
      try {
        const response = await axios.get(`${this.baseURL}/api/voices`, {
          params: {
            size: 999 // 设置一个足够大的数值以获取所有音色
          },
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
        input_text: '',
        speaker_name: this.form.speaker_name || ''
      }
      
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
        const totalTasks = this.form.speaker_names.length
        let completedTasks = 0
        let failedTasks = 0
        
        try {
          const createTaskPromises = this.form.speaker_names.map(async (speakerName) => {
            try {
              const formData = new FormData()
              formData.append('name', uuidv4())
              formData.append('type', 'text2speech')
              formData.append('input_text', this.form.input_text)
              formData.append('speaker_name', speakerName)
              
              await axios.post(`${this.baseURL}/api/tts`, formData, {
                headers: { 
                  Authorization: `Bearer ${this.token}`,
                  'Content-Type': 'multipart/form-data'
                }
              })
              completedTasks++
              this.$message.success(`音色 ${speakerName} 的任务创建成功 (${completedTasks}/${totalTasks})`)
            } catch (error) {
              failedTasks++
              console.error(`音色 ${speakerName} 的任务创建失败:`, error)
              this.$message.error(`音色 ${speakerName} 的任务创建失败: ${(error.response && error.response.data && error.response.data.error) || '创建任务失败'}`)
            }
          })
          
          await Promise.all(createTaskPromises)
          
          if (completedTasks === totalTasks) {
            this.$message.success('所有TTS任务创建完成')
            this.dialogVisible = false
            this.fetchTasks()
          } else if (completedTasks > 0) {
            this.$message.warning(`部分任务创建成功 (${completedTasks}/${totalTasks}), 失败 ${failedTasks} 个`)
            this.dialogVisible = false
            this.fetchTasks()
          } else {
            this.$message.error('所有任务创建失败')
          }
        } catch (error) {
          console.error('批量创建TTS任务失败:', error)
          this.$message.error('批量创建任务失败')
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
    
    // 复制文本到剪贴板
    copyText(text) {
      if (!text) return
      
      navigator.clipboard.writeText(text)
        .then(() => {
          this.$message.success('文本已复制到剪贴板')
        })
        .catch(err => {
          console.error('复制失败:', err)
          this.$message.error('复制失败')
        })
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
    
    // 处理分页大小变化（列表视图使用）
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchTasks()
    },
    
    // 处理页码变化（列表视图使用）
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchTasks()
    },
    
    // 切换视图模式（列表/卡片）
    toggleView() {
      this.isCardView = !this.isCardView
      localStorage.setItem('tts_view_mode', this.isCardView ? 'card' : 'list')
      
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
          this.debug('intersection事件:', '可见=', entry.isIntersecting, '加载中=', this.loadingMore, '有更多数据=', this.hasMoreData, '可见比例=', entry.intersectionRatio)
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
    }
  }
}
</script>

<style scoped>
.tts-container {
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

.task-list {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 15px;
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.responsive-table {
  background: transparent !important;
}

.responsive-table .el-table__header-wrapper,
.responsive-table .el-table__body-wrapper {
  background-color: transparent;
}

.responsive-table th {
  background-color: rgba(0, 0, 0, 0.2) !important;
  color: #fff !important;
  font-weight: 600;
  padding: 8px 0;
}

.responsive-table td {
  background-color: transparent !important;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1) !important;
  color: #fff;
}

.action-buttons {
  display: flex;
  justify-content: space-around;
  flex-wrap: nowrap;
}

.action-buttons .action-btn {
  margin: 0 3px;
  transition: all 0.2s;
}

.action-buttons .action-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

.action-btn {
  padding: 4px 8px;
  margin: 0 2px;
  border-radius: 4px;
  transition: all 0.3s;
  font-size: 13px;
  color: #f5f5f5;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateY(-2px);
  color: #fff;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.el-upload__tip {
  line-height: 1.2;
}

.text-with-copy {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  padding: 4px 0;
  gap: 8px;
  overflow: hidden;
}

.text-content {
  flex: 1 1 auto;
  min-width: 0;
  word-break: break-all;
  margin-bottom: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.copy-btn {
  flex: 0 0 auto;
  padding: 2px;
  border-radius: 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  transition: all 0.3s;
  white-space: nowrap;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin: 0;
  min-width: auto;
  width: 24px;
  height: 24px;
}

.copy-btn i {
  margin-right: 0;
}

.copy-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
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