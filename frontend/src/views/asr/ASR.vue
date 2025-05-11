<template>
  <div class="asr-container">
    <div class="page-header" :class="{'mobile-header': isMobile}">
      <div class="header-left">
        <h2>语音识别</h2>
      </div>
      <div class="header-right">
        <el-button v-if="!isMobile" type="primary" @click="handleUpload" icon="el-icon-plus">创建语音识别任务</el-button>
        <el-button v-if="!isMobile" type="text" size="small" class="view-toggle" @click="toggleView">
          <i :class="isCardView ? 'el-icon-menu' : 'el-icon-s-grid'"></i>
          <span class="toggle-text">{{ isCardView ? '列表视图' : '卡片视图' }}</span>
        </el-button>
      </div>
    </div>

    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>

    <!-- 任务列表（表格视图） -->
    <div v-loading="loading" class="task-list" v-show="!isCardView">
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
    
    <!-- 卡片视图（瀑布流） -->
    <div v-show="isCardView" class="card-list" v-loading="loading">
      <el-empty v-if="tasks.length === 0" description="暂无语音识别任务"></el-empty>
      
      <div v-else class="card-view-content">
        <div class="waterfall-container" ref="cardContainer" :class="{'mobile-card-container': isMobile}">
          <div class="task-card" v-for="item in tasks" :key="item.id">
            <div class="task-card-header">
              <h3 class="task-card-title">{{ item.name }}</h3>
              <div class="status-icon">
                <i v-if="item.status === 'completed'" class="el-icon-check" style="color: #67c23a;"></i>
                <i v-else-if="item.status === 'failed'" class="el-icon-close" style="color: #f56c6c;"></i>
                <el-tag v-else :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
              </div>
            </div>
            <div class="task-card-content">
              <div class="task-card-info">
                <p v-if="item.output_text" class="text-ellipsis">
                  <span class="info-label">识别结果:</span> {{ item.output_text }}
                </p>
                <p><span class="info-label">创建时间:</span> {{ formatDate(item.created_at) }}</p>
              </div>
            </div>
            <div class="task-card-footer">
              <el-button 
                type="text" 
                size="small" 
                class="action-btn"
                @click="viewDetail(item.id)"
              >
                <i class="el-icon-view"></i> 查看
              </el-button>
              <el-button 
                type="text" 
                size="small" 
                class="action-btn"
                @click="playAudio(item)" 
                :disabled="!item.input_file || item.status !== 'completed'"
              >
                <i class="el-icon-video-play"></i> 播放
              </el-button>
              <el-button 
                v-if="item.status === 'failed'" 
                type="text" 
                size="small" 
                class="action-btn"
                @click="handleRetry(item)"
              >重试</el-button>
              <el-button 
                type="text" 
                size="small" 
                class="action-btn"
                @click="handleDelete(item)"
              >删除</el-button>
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

    <!-- 移动端悬浮添加按钮 -->
    <div v-if="isMobile" class="floating-add-btn" @click="handleUpload">
      <i class="el-icon-plus"></i>
    </div>

    <!-- 音频播放器 -->
    <audio ref="audioPlayer" style="display: none"></audio>
    
    <!-- 创建任务对话框 -->
    <el-dialog 
      title="创建语音识别任务" 
      :visible.sync="uploadDialogVisible" 
      :fullscreen="isMobile"
      :modal="true"
      :close-on-click-modal="false"
      :append-to-body="true"
      :show-close="!isMobile"
      custom-class="asr-dialog"
      width="50%">
      
      <!-- 移动端顶部导航 -->
      <div v-if="isMobile" class="mobile-header-bar">
        <div class="header-back" @click="uploadDialogVisible = false">
          <i class="el-icon-arrow-left"></i>
          <span>返回</span>
        </div>
      </div>
      
      <el-form :model="asrForm" :rules="asrRules" ref="asrForm" label-width="100px" class="asr-form" :label-position="isMobile ? 'top' : 'left'">
        <el-form-item label="任务名称" prop="name">
          <el-input 
            v-model="asrForm.name" 
            placeholder="请输入任务名称"
            @focus="handleInputFocus"
            @blur="handleInputBlur"></el-input>
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
          <!-- 录音预览 - 改进播放控件布局 -->
          <div v-if="recordedAudio" class="recorded-audio-preview">
            <div class="audio-player-wrapper">
              <audio :src="recordedAudioUrl" controls controlsList="nodownload" ref="previewAudio"></audio>
              <div class="audio-player-fallback" v-if="isMobile">
                <el-button size="small" type="primary" icon="el-icon-video-play" @click="playPreviewAudio">播放</el-button>
                <el-button size="small" type="info" icon="el-icon-video-pause" @click="pausePreviewAudio">暂停</el-button>
              </div>
            </div>
            <div class="preview-actions">
              <el-button size="small" type="primary" @click="useRecordedAudio">使用录制的音频</el-button>
              <el-button size="small" @click="discardRecordedAudio">放弃</el-button>
            </div>
          </div>
        </el-form-item>
        <el-form-item v-else label="音频URL" prop="audioUrl">
          <el-input 
            v-model="asrForm.audioUrl" 
            placeholder="请输入音频文件URL"
            @focus="handleInputFocus"
            @blur="handleInputBlur"></el-input>
        </el-form-item>
        
        <!-- 移动端底部按钮 -->
        <div v-if="isMobile" class="mobile-form-footer">
          <el-button type="primary" @click="submitASRTask" class="mobile-submit-btn">创建任务</el-button>
        </div>
      </el-form>
      
      <!-- 桌面端底部按钮 -->
      <div v-if="!isMobile" slot="footer" class="dialog-footer">
        <el-button @click="uploadDialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitASRTask">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getAudioUrl } from '@/utils/fileAccess'
import '@/assets/styles/card-view.css'

export default {
  name: 'ASR',
  data() {
    return {
      loading: false,
      tasks: [],
      currentPage: 1,
      pageSize: 10,
      cardPageSize: 10,
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
      },
      isCardView: false,
      loadingMore: false,
      hasMoreData: true,
      initialLoaded: false,
      observer: null,
      isMobile: false, // 是否为移动端
      lastScrollTop: 0, // 记录上次滚动位置
      // 视口控制相关变量
      originalViewportContent: null,
      isInputFocused: false
    }
  },
  
  created() {
    // 从本地存储中读取用户偏好的视图模式
    const savedViewMode = localStorage.getItem('asr_view_mode')
    if (savedViewMode) {
      this.isCardView = savedViewMode === 'card'
    }
    
    // 初始加载数据
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
    
    // 设置移动端视口
    if (this.isMobile) {
      this.setupMobileViewport();
    }
  },
  
  beforeDestroy() {
    // 移除滚动事件监听器
    window.removeEventListener('scroll', this.handleWindowScroll)
    
    // 清除IntersectionObserver
    if (this.observer) {
      this.observer.disconnect()
      this.observer = null
    }
    
    // 移除事件监听
    window.removeEventListener('resize', this.checkDeviceType);
    
    // 重置视口设置
    this.resetMobileViewport();
  },
  
  methods: {
    // 辅助方法: 调试日志
    debug(...args) {
      console.log('[ASR]', ...args)
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
        this.debug('已有请求进行中，跳过')
        return
      }
      
      if (!loadMore) {
        this.loading = true
      } else {
        this.loadingMore = true
      }
      
      this.debug('请求数据:', '页码=', this.currentPage, '每页数量=', this.isCardView ? this.cardPageSize : this.pageSize)
      
      try {
        const response = await this.$http.get('/api/asr', {
          params: {
            page: this.currentPage,
            size: this.isCardView ? this.cardPageSize : this.pageSize
          }
        })
        
        const newTasks = response.data.items || []
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
        console.error('获取ASR任务列表失败:', error)
        this.$message.error('获取任务列表失败: ' + (error.message || '未知错误'))
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
          this.loadMoreTasks();
        }
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
    
    // 切换视图
    toggleView() {
      this.isCardView = !this.isCardView;
      
      // 切换视图后滚动到顶部
      this.$nextTick(() => {
        this.scrollToTop();
      });
      
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

    // 设置移动端视口
    setupMobileViewport() {
      if (!this.isMobile) return;
      
      let viewportMeta = document.querySelector('meta[name="viewport"]');
      if (!viewportMeta) {
        viewportMeta = document.createElement('meta');
        viewportMeta.name = 'viewport';
        document.head.appendChild(viewportMeta);
      }
      
      // 保存原始视口设置
      if (!this.originalViewportContent) {
        this.originalViewportContent = viewportMeta.content;
      }
      
      // 设置禁止用户缩放的视口
      viewportMeta.content = 'width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no';
    },
    
    // 重置移动端视口
    resetMobileViewport() {
      if (!this.isMobile) return;
      
      let viewportMeta = document.querySelector('meta[name="viewport"]');
      if (viewportMeta && this.originalViewportContent) {
        viewportMeta.content = this.originalViewportContent;
      } else if (viewportMeta) {
        viewportMeta.content = 'width=device-width, initial-scale=1.0';
      }
    },
    
    // 处理输入框焦点
    handleInputFocus() {
      this.isInputFocused = true;
    },
    
    // 处理输入框失焦
    handleInputBlur() {
      this.isInputFocused = false;
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
      
      // 对话框打开时设置移动端视口
      if (this.isMobile) {
        this.$nextTick(() => {
          this.setupMobileViewport();
        });
      }
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
        // 确保在移动端上获取正确的音频格式
        let options = {
          audio: {
            echoCancellation: true,
            noiseSuppression: true,
            autoGainControl: true
          }
        };
        
        const stream = await navigator.mediaDevices.getUserMedia(options);
        const audioTracks = stream.getAudioTracks();
        
        if (audioTracks.length > 0) {
          console.log('使用音频设备:', audioTracks[0].label);
        }
        
        // 对于iOS，需要特殊处理媒体录制器
        let mimeType = 'audio/webm';
        if (!MediaRecorder.isTypeSupported('audio/webm')) {
          mimeType = 'audio/mp4';
          if (!MediaRecorder.isTypeSupported('audio/mp4')) {
            mimeType = '';  // 让浏览器选择支持的格式
          }
        }
        
        this.mediaRecorder = new MediaRecorder(stream, {
          mimeType: mimeType || ''
        });
        
        this.recordedChunks = [];
        
        this.mediaRecorder.ondataavailable = (event) => {
          if (event.data.size > 0) {
            this.recordedChunks.push(event.data);
            console.log('获取录音数据块，大小:', event.data.size);
          }
        };
        
        this.mediaRecorder.onstop = () => {
          // 确定正确的MIME类型
          let audioType = mimeType || 'audio/webm';
          // 对于iOS Safari，通常使用audio/mp4
          if (/iPhone|iPad|iPod/.test(navigator.userAgent)) {
            audioType = 'audio/mp4';
          }
          
          const blob = new Blob(this.recordedChunks, { type: audioType });
          console.log('录音完成，创建Blob，大小:', blob.size, '类型:', audioType);
          
          // 为iOS创建特殊文件名
          const fileName = /iPhone|iPad|iPod/.test(navigator.userAgent) 
            ? 'recording.m4a' 
            : 'recording.webm';
            
          this.recordedAudio = new File([blob], fileName, { type: audioType });
          this.recordedAudioUrl = URL.createObjectURL(blob);
          
          // 停止所有音轨
          stream.getTracks().forEach(track => track.stop());
          
          // 确保UI更新
          this.$nextTick(() => {
            if (this.$refs.previewAudio) {
              this.$refs.previewAudio.load();
            }
          });
        };
        
        this.mediaRecorder.start(1000); // 每秒触发一次ondataavailable
        this.isRecording = true;
        console.log('开始录音，使用MIME类型:', this.mediaRecorder.mimeType);
      } catch (error) {
        console.error('录音失败:', error);
        this.$message.error('无法访问麦克风，请确保已授予麦克风权限');
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
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchTasks()
    },

    // 当前页改变
    handleCurrentChange(page) {
      this.currentPage = page
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
          
          // 重置视口设置
          if (this.isMobile) {
            this.$nextTick(() => {
              this.resetMobileViewport();
            });
          }
        }
      } catch (error) {
        const errorMessage = (error.response && error.response.data && error.response.data.error) || error.message || '创建任务失败';
        this.$message.error(errorMessage);
      }
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
    // 滚动到页面顶部
    scrollToTop() {
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    },
    // 播放预览音频
    playPreviewAudio() {
      if (this.$refs.previewAudio) {
        this.$refs.previewAudio.play().catch(err => {
          console.error('播放预览音频失败:', err);
          this.$message.error('播放失败，请尝试使用此音频或重新录制');
        });
      }
    },
    
    // 暂停预览音频
    pausePreviewAudio() {
      if (this.$refs.previewAudio) {
        this.$refs.previewAudio.pause();
      }
    },
  }
}
</script>

<style scoped>
.asr-container {
  padding: 20px;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
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
  font-size: 1.4rem;
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
  border-color: rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  transform: none; /* 移除向上移动效果 */
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
  .asr-container {
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
  
  .task-list {
    margin-top: 10px;
    padding-bottom: 60px;
  }
  
  /* 悬浮按钮移动端样式 */
  .floating-add-btn {
    bottom: 80px;
    right: 16px;
    width: 56px;
    height: 56px;
    background: linear-gradient(135deg, #3f51b5, #2196f3);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
    z-index: 1001; /* 确保在底部菜单之上 */
  }
  
  .floating-add-btn i {
    font-size: 28px;
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
  .task-list,
  .card-view-content,
  .waterfall-container {
    -webkit-overflow-scrolling: touch;
  }
  
  /* 空状态优化 */
  .el-empty {
    margin-top: 60px !important;
  }
  
  /* 触碰反馈优化 */
  .task-card:active {
    transform: none !important; /* 确保激活时没有变形 */
    opacity: 0.95;
  }
}

/* 悬浮添加按钮 */
.floating-add-btn {
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
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.3);
  z-index: 100;
  cursor: pointer;
  transition: all 0.3s;
}

.floating-add-btn i {
  font-size: 24px;
}

.floating-add-btn:active {
  transform: scale(0.95);
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
}

/* 移动端对话框样式 */
@media screen and (max-width: 768px) {
  /* 确保对话框占满全屏 */
  .el-dialog.asr-dialog {
    width: 100% !important;
    height: 100% !important;
    margin: 0 !important;
    border-radius: 0 !important;
    overflow: hidden !important;
  }
  
  /* 确保对话框内容区域可滚动 */
  .el-dialog__body {
    padding: 0 !important;
    overflow-y: auto !important;
    -webkit-overflow-scrolling: touch !important;
    padding-top: 56px !important; /* 为顶部导航留出空间 */
    height: calc(100% - 56px) !important;
  }
  
  /* 确保表单可编辑 */
  .asr-form {
    padding: 10px 15px 70px !important; /* 为底部按钮留出空间 */
  }
  
  /* 固定底部按钮 */
  .mobile-form-footer {
    position: fixed !important;
    bottom: 0 !important;
    left: 0 !important;
    right: 0 !important;
    z-index: 2002 !important;
    background-color: #fff !important;
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1) !important;
  }
  
  /* 移动端底部按钮样式 */
  .mobile-submit-btn {
    width: 100%;
    height: 56px;
    font-size: 16px;
    font-weight: 500;
    border-radius: 0;
    margin: 0;
    background: linear-gradient(135deg, #1976d2, #64b5f6);
    border: none;
    color: #fff;
    letter-spacing: 1px;
    display: flex;
    align-items: center;
    justify-content: center;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
    transition: all 0.3s ease;
  }
  
  .mobile-submit-btn:active {
    background: linear-gradient(135deg, #1565c0, #42a5f5);
    transform: translateY(1px);
  }
  
  /* 移动端顶部导航栏 */
  .mobile-header-bar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 56px;
    background-color: #409EFF;
    display: flex;
    align-items: center;
    padding: 0 15px;
    z-index: 2003;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.15);
    color: white;
  }
  
  .header-back {
    display: flex;
    align-items: center;
    color: #fff;
    font-size: 16px;
    cursor: pointer;
    font-weight: 500;
  }
  
  .header-back i {
    margin-right: 5px;
    font-size: 18px;
  }
  
  /* 录音区域响应式调整 */
  .audio-upload-container {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .recorded-audio-preview {
    margin-top: 10px;
    width: 100%;
  }
  
  .recorded-audio-preview audio {
    width: 100%;
    max-width: 100%;
  }
  
  .preview-actions {
    display: flex;
    gap: 8px;
    margin-top: 8px;
    flex-wrap: wrap;
  }
  
  /* 修复iOS上的滚动问题 */
  .el-dialog__wrapper {
    -webkit-overflow-scrolling: touch;
  }
  
  /* 输入框样式优化 */
  .el-textarea__inner,
  .el-input__inner {
    font-size: 16px !important; /* 避免iOS自动缩放 */
    padding: 10px !important;
    line-height: 1.5 !important;
  }
}

/* 增强录音预览样式 */
.recorded-audio-preview {
  margin-top: 15px;
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.05);
}

.audio-player-wrapper {
  width: 100%;
  margin-bottom: 12px;
}

.audio-player-wrapper audio {
  width: 100%;
  max-width: 100%;
  height: 40px;
  margin-bottom: 10px;
}

.audio-player-fallback {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin: 10px 0;
}

.preview-actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-top: 12px;
}

/* 修复移动端音频播放问题 */
@media screen and (max-width: 768px) {
  .el-dialog.asr-dialog {
    width: 100% !important;
    height: 100% !important;
    margin: 0 !important;
    border-radius: 0 !important;
    overflow: hidden !important;
  }
  
  /* 确保对话框内容区域可滚动 */
  .el-dialog__body {
    padding: 0 !important;
    overflow-y: auto !important;
    -webkit-overflow-scrolling: touch !important;
    padding-top: 56px !important; /* 为顶部导航留出空间 */
    height: calc(100% - 56px) !important;
  }
  
  /* 移动端录音组件优化 */
  .audio-upload-container {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-bottom: 15px;
  }
  
  .recorded-audio-preview {
    margin-top: 15px;
    padding: 15px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.2);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
  
  .audio-player-wrapper {
    width: 100%;
  }
  
  .audio-player-wrapper audio {
    width: 100%;
    height: 40px;
    border-radius: 20px;
    background: rgba(0, 0, 0, 0.2);
  }
  
  .audio-player-fallback {
    margin: 12px 0;
  }
  
  .audio-player-fallback .el-button {
    flex: 1;
    height: 40px;
    border-radius: 20px;
  }
  
  .preview-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 15px;
  }
  
  .preview-actions .el-button {
    width: 100%;
    height: 44px;
    margin: 0;
  }
}

/* 修复移动端音频播放问题 */
@media screen and (max-width: 768px) {
  /* ... 其他样式 ... */
  
  /* 修复iOS上的输入框放大问题 */
  input[type="text"],
  input[type="url"],
  input[type="email"],
  input[type="number"],
  input[type="password"],
  textarea,
  select {
    font-size: 16px !important; /* 关键：16px或更大可以防止iOS缩放 */
    max-height: none !important;
  }
  
  .el-input__inner,
  .el-textarea__inner {
    font-size: 16px !important;
    line-height: 20px !important;
  }
  
  /* 对话框类容器禁止缩放 */
  .el-dialog__wrapper,
  .el-dialog,
  .el-dialog__body {
    touch-action: pan-y !important;
  }
  
  /* 输入框聚焦时的样式，提供用户反馈 */
  .el-input.is-focus .el-input__inner {
    border-color: #409EFF !important;
    box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2) !important;
  }
}
</style>