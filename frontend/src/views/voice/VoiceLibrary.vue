<template>
  <div class="voice-library-container">
    <div class="page-header" :class="{'mobile-header': isMobile}">
      <div class="header-left">
        <h2>音色库</h2>
      </div>
      <div class="header-right">
        <!-- <el-button type="primary" @click="showUploadDialog">上传音色</el-button> -->
        <el-button v-if="!isMobile" type="primary" @click="refreshVoices">刷新</el-button>
        <el-button v-if="!isMobile" type="text" size="small" class="view-toggle" @click="toggleView">
          <i :class="isCardView ? 'el-icon-menu' : 'el-icon-s-grid'"></i>
          <span class="toggle-text">{{ isCardView ? '列表视图' : '卡片视图' }}</span>
        </el-button>
      </div>
    </div>
    
    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>
    
    <!-- 移动端浮动刷新按钮 -->
    <div v-if="isMobile" class="floating-refresh-btn" @click="refreshVoices">
      <i class="el-icon-refresh"></i>
    </div>
    
    <!-- 音色列表（表格视图） -->
    <div v-loading="loading" class="voice-list" v-show="!isCardView">
      <el-empty v-if="voices.length === 0" description="暂无音色"></el-empty>
      
      <el-table v-else :data="voices" style="width: 100%">
        <el-table-column prop="name" label="音色名称" width="150"></el-table-column>
        <el-table-column prop="alias" label="别名" width="150">
          <template slot-scope="scope">
            {{ scope.row.alias || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="gender" label="性别" width="80">
          <template slot-scope="scope">
            {{ scope.row.gender || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
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
        <el-table-column label="操作" width="150">
          <template slot-scope="scope">
            <el-button type="text" size="small" @click="useForTTS(scope.row.id)">用于TTS</el-button>
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
    
    <!-- 音色卡片视图（瀑布流） -->
    <div v-show="isCardView" class="card-list" v-loading="loading">
      <el-empty v-if="voices.length === 0" description="暂无音色"></el-empty>
      
      <div v-else class="card-view-content">
        <div class="waterfall-container" ref="cardContainer" :class="{'mobile-card-container': isMobile}">
          <div class="voice-card" v-for="item in voices" :key="item.id">
            <div class="voice-card-header">
              <h3 class="voice-card-title">{{ item.name }}</h3>
              <el-tag size="small" :type="item.gender === '男' ? 'primary' : 'warning'">
                {{ item.gender || '未知' }}
              </el-tag>
            </div>
            <div class="voice-card-content">
              <div class="voice-card-info">
                <p v-if="item.alias"><span class="info-label">别名:</span> {{ item.alias }}</p>
                <p><span class="info-label">创建时间:</span> {{ formatDate(item.created_at) }}</p>
                <p v-if="item.description"><span class="info-label">描述:</span> {{ item.description }}</p>
              </div>
            </div>
            <div class="voice-card-footer">
              <el-button 
                type="text" 
                size="small" 
                class="action-btn"
                @click="playAudio(item)" 
                :disabled="!item.sample_file">
                <i class="el-icon-video-play"></i> 试听
              </el-button>
              <el-button 
                type="text" 
                size="small" 
                class="action-btn"
                @click="useForTTS(item.id)">
                用于TTS
              </el-button>
              <el-button 
                type="text" 
                size="small" 
                class="action-btn"
                @click="confirmDelete(item.id)">
                删除
              </el-button>
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
import '@/assets/styles/card-view.css'

export default {
  name: 'VoiceLibrary',
  data() {
    return {
      loading: false,
      uploading: false,
      voices: [],
      currentPage: 1,
      pageSize: 10,
      cardPageSize: 10,
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
      },
      isCardView: false,
      loadingMore: false,
      hasMoreData: true,
      initialLoaded: false,
      observer: null,
      isMobile: false,
      lastScrollTop: 0
    }
  },
  computed: {
    baseUrl() {
      return (process.env.VUE_APP_API_URL || '') + '/'
    }
  },

  created() {
    // 从本地存储中读取用户偏好的视图模式
    const savedViewMode = localStorage.getItem('voice_library_view_mode')
    if (savedViewMode) {
      this.isCardView = savedViewMode === 'card'
    }
    
    // 初始加载第一页数据
    this.loadInitialData()
    
    // 检测设备类型
    this.checkDeviceType();
    // 监听窗口大小变化
    window.addEventListener('resize', this.checkDeviceType);
  },
  
  mounted() {
    // 添加滚动事件监听器
    window.addEventListener('scroll', this.handleWindowScroll);
    
    // 如果是移动端，默认使用卡片视图
    if (this.isMobile) {
      this.isCardView = true;
    }
  },
  
  beforeDestroy() {
    // 移除事件监听
    window.removeEventListener('resize', this.checkDeviceType);
    window.removeEventListener('scroll', this.handleWindowScroll);
    
    // 清除IntersectionObserver
    if (this.observer) {
      this.observer.disconnect()
      this.observer = null
    }
  },
  
  methods: {
    // 辅助方法: 调试日志
    debug(...args) {
      console.log('[VoiceLibrary]', ...args)
    },
    
    // 显示上传对话框
    showUploadDialog() {
      // 重置表单
      this.uploadForm = {
        name: '',
        description: '',
        model_name: '',
        model_file: null,
        sample_file: null
      }
      this.modelFileList = []
      this.sampleFileList = []
      this.uploadDialogVisible = true
    },
    
    // 初始加载数据
    loadInitialData() {
      if (this.initialLoaded) {
        this.debug('已加载初始数据，跳过')
        return
      }
      
      this.debug('加载初始数据')
      this.currentPage = 1
      this.voices = []
      this.hasMoreData = true
      this.fetchVoices()
      this.initialLoaded = true
    },
    
    // 获取音色列表
    fetchVoices(loadMore = false) {
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
      
      this.$http.get(`/api/voices`, {
        params: {
          page: this.currentPage,
          size: this.isCardView ? this.cardPageSize : this.pageSize
        }
      })
        .then(response => {
          const newVoices = response.data.voices || []
          this.total = response.data.total || 0
          
          this.debug('获取到新数据:', newVoices.length, '总数:', this.total)
          
          if (loadMore) {
            // 追加新数据
            this.voices = [...this.voices, ...newVoices]
          } else {
            // 重置数据
            this.voices = newVoices
          }
          
          // 判断是否还有更多数据
          this.hasMoreData = this.voices.length < this.total
          this.debug('当前数据量：', this.voices.length, '总数：', this.total, '是否还有更多：', this.hasMoreData)
        })
        .catch(error => {
          console.error('获取音色列表失败', error)
          this.$message.error('获取音色列表失败')
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
    loadMoreVoices() {
      if (this.loadingMore || !this.hasMoreData) {
        this.debug('跳过加载更多:', '加载中=', this.loadingMore, '没有更多数据=', !this.hasMoreData)
        return
      }
      
      this.debug('开始加载更多数据，当前页码：', this.currentPage)
      this.currentPage++
      this.fetchVoices(true)
    },
    
    // 处理窗口滚动事件
    handleWindowScroll() {
      // 记录滚动方向
      const currentScrollTop = window.pageYOffset || document.documentElement.scrollTop;
      const scrollingDown = currentScrollTop > this.lastScrollTop;
      this.lastScrollTop = currentScrollTop;
      
      // 加载更多数据条件
      if (this.isCardView && scrollingDown && !this.loadingMore && this.hasMoreData) {
        const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
        const windowHeight = window.innerHeight;
        const documentHeight = Math.max(
          document.body.scrollHeight, document.documentElement.scrollHeight,
          document.body.offsetHeight, document.documentElement.offsetHeight,
          document.body.clientHeight, document.documentElement.clientHeight
        );
        
        // 当滚动到距离底部阈值距离时触发加载
        if (documentHeight - scrollTop - windowHeight < 200) {
          this.loadMoreVoices();
        }
      }
    },
    
    // 滚动到页面顶部
    scrollToTop() {
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    },
    
    // 切换视图
    toggleView() {
      this.isCardView = !this.isCardView;
      
      // 切换视图后滚动到顶部
      this.$nextTick(() => {
        this.scrollToTop();
      });
      
      // 重置状态
      this.currentPage = 1
      this.voices = []
      this.hasMoreData = true
      
      // 重新加载第一页数据
      this.fetchVoices()
      
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
      this.fetchVoices()
    },
    
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchVoices()
    },
    
    // 格式化日期
    formatDate(dateString) {
      if (!dateString) return ''
      return new Date(dateString).toLocaleString()
    },
    
    // 播放音频
    async playAudio(voice) {
      if (!voice.sample_file) {
        this.$message.warning('该音色没有示例音频')
        return
      }
      
      const audioPlayer = this.$refs.audioPlayer
      
      // 如果正在播放同一个音频，则暂停
      if (this.currentPlayingId === voice.id && !audioPlayer.paused) {
        audioPlayer.pause()
        this.currentPlayingId = null
        return
      }
      
      try {
        // 获取带认证的音频URL
        const audioUrl = await getAudioUrl(voice.sample_file)
        
        // 播放新的音频
        audioPlayer.src = audioUrl
        await audioPlayer.play()
        this.currentPlayingId = voice.id
        
        // 播放完成后重置状态
        audioPlayer.onended = () => {
          this.currentPlayingId = null
          // 释放URL对象
          if (audioUrl) {
            window.URL.revokeObjectURL(audioUrl)
          }
        }
      } catch (error) {
        console.error('音频播放失败:', error)
        this.$message.error('音频加载失败，请稍后重试')
        this.currentPlayingId = null
      }
    },
    
    // 用于TTS
    useForTTS(voiceId) {
      const voice = this.voices.find(v => v.id === voiceId)
      if (!voice) return
      
      this.loading = true
      this.$router.replace({
        path: '/tts',
        query: { 
          voice_id: voiceId,
          voice_name: voice.name
        }
      })
      
      // 使用nextTick确保DOM更新完成
      this.$nextTick(() => {
        this.loading = false
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
    },

    // 检查设备类型
    checkDeviceType() {
      this.isMobile = window.innerWidth <= 768;
      
      // 在移动端强制使用卡片视图
      if (this.isMobile) {
        this.isCardView = true;
        
        // 隐藏视图切换按钮，只在移动端上这样做
        const viewToggleBtn = document.querySelector('.view-toggle');
        if (viewToggleBtn) {
          viewToggleBtn.style.display = 'none';
        }
      }
    },

    // 刷新音色列表
    refreshVoices() {
      this.currentPage = 1;
      this.fetchVoices();
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
            this.loadMoreVoices()
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
  }
}
</script>

<style scoped>
.voice-library-container {
  padding: 20px;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 999;
  border-radius: 0;
  padding: 8px 10px;
  margin: 0;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.mobile-header-placeholder {
  height: 50px;
  width: 100%;
  margin: 0;
  padding: 0;
}

.header-left, .header-right {
  display: flex;
  align-items: center;
}

.header-left {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.voice-list {
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
  gap: 15px;
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

.voice-card {
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

.voice-card:hover {
  border-color: rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  transform: none;
}

.voice-card-header {
  padding: 12px 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 0, 0, 0.2);
}

.voice-card-title {
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

.voice-card-content {
  padding: 10px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 90px;
}

.voice-card-info {
  margin-bottom: 10px;
  overflow: hidden;
}

.voice-card-info p {
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

.voice-card-footer {
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
  .voice-library-container {
    padding: 0;
    width: 100%;
    overflow-x: hidden;
    overflow-y: auto;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    -webkit-overflow-scrolling: touch;
  }
  
  .page-header {
    flex-direction: row;
    align-items: center;
    padding: 10px 12px;
    margin: 0;
    height: 50px;
    box-sizing: border-box;
  }
  
  .header-right {
    margin-top: 0;
    width: auto;
    justify-content: flex-end;
  }
  
  .toggle-text {
    display: none;
  }
  
  .page-header h2 {
    margin: 0;
    font-size: 1.3em;
    max-width: 200px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: bold;
  }
  
  .page-header .el-button {
    margin: 0;
    padding: 5px 8px;
    font-size: 12px;
  }
  
  .view-toggle {
    padding: 3px 6px;
  }
  
  .mobile-header-placeholder {
    height: 52px;
    margin: 0;
    padding: 0;
  }
  
  .voice-list {
    margin-top: 10px;
    padding-bottom: 60px;
  }
  
  /* 移动端底部菜单激活状态 */
  .mobile-footer-menu .menu-item.active {
    color: #2196f3;
    font-weight: bold;
  }
  
  .mobile-footer-menu .menu-item.active i {
    transform: scale(1.1);
  }
  
  /* 修复iOS移动端滑动问题 */
  .card-list, 
  .voice-list,
  .card-view-content,
  .waterfall-container {
    -webkit-overflow-scrolling: touch;
  }
  
  /* 空状态优化 */
  .el-empty {
    margin-top: 60px !important;
  }
  
  /* 触碰反馈优化 */
  .voice-card:active {
    transform: none !important;
    opacity: 0.95;
  }
  
  /* 隐藏在移动端不重要的表格列 */
  .hide-on-mobile {
    display: none;
  }
  
  /* 浮动刷新按钮 */
  .floating-refresh-btn {
    bottom: 20px;
    right: 16px;
    width: 56px;
    height: 56px;
    background: linear-gradient(135deg, #3f51b5, #2196f3);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
    z-index: 1001;
  }
  
  .floating-refresh-btn i {
    font-size: 28px;
  }
}

/* 浮动刷新按钮 */
.floating-refresh-btn {
  position: fixed;
  bottom: 70px;
  right: 15px;
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: linear-gradient(135deg, #1976d2, #64b5f6);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  cursor: pointer;
  transition: all 0.3s;
  z-index: 99;
}

.floating-refresh-btn:hover, .floating-refresh-btn:active {
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}
</style>