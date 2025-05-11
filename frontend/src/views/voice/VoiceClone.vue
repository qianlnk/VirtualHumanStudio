<template>
  <div class="voice-clone-container">
    <div class="page-header" :class="{'mobile-header': isMobile}">
      <div class="header-left">
        <h2>音色克隆</h2>
      </div>
      <div class="header-right">
        <el-button v-if="!isMobile" type="primary" @click="createVoiceClone" icon="el-icon-plus">创建音色克隆任务</el-button>
        <el-button v-if="!isMobile" type="text" size="small" class="view-toggle" @click="toggleView">
          <i :class="isCardView ? 'el-icon-menu' : 'el-icon-s-grid'"></i>
          <span class="toggle-text">{{ isCardView ? '列表视图' : '卡片视图' }}</span>
        </el-button>
      </div>
    </div>
    
    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>
    
    <!-- 任务列表（表格视图） -->
    <div v-loading="loading" class="task-list mobile-card-view" v-show="!isCardView">
      <el-empty v-if="tasks.length === 0" description="暂无音色克隆任务"></el-empty>
      
      <el-table v-else :data="tasks" style="width: 100%" class="responsive-table">
        <el-table-column prop="speaker_name" label="说话人名称" width="150"></el-table-column>
        <el-table-column prop="prompt_text" label="提示词" show-overflow-tooltip>
          <template slot-scope="scope">
            <span>{{ scope.row.prompt_text || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" class="hide-on-mobile">
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
        <div class="waterfall-container" ref="cardContainer" :class="{'mobile-card-container': isMobile}">
          <div class="task-card" v-for="item in tasks" :key="item.id">
            <div class="task-card-header">
              <h3 class="task-card-title">{{ item.speaker_name }}</h3>
              <el-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
            </div>
            <div class="task-card-content">
              <div class="task-card-info">
                <div class="text-with-copy card-text">
                  <p class="text-ellipsis">{{ item.prompt_text || '-' }}</p>
                </div>
                <p><i class="el-icon-time"></i> {{ formatDate(item.created_at) }}</p>
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
    
    <!-- 移动端悬浮添加按钮 -->
    <div v-if="isMobile" class="floating-add-btn" @click="createVoiceClone">
      <i class="el-icon-plus"></i>
    </div>
    
    <!-- 音色克隆创建表单 -->
    <el-dialog 
      :visible.sync="dialogVisible" 
      :fullscreen="isMobile"
      :modal="false"
      :close-on-click-modal="false"
      :append-to-body="true"
      :show-close="!isMobile"
      width="600px"
      custom-class="voice-clone-dialog">
      
      <!-- 移动端顶部导航 -->
      <div v-if="isMobile" class="mobile-header-bar">
        <div class="header-back" @click="dialogVisible = false">
          <i class="el-icon-arrow-left"></i>
          <span>返回</span>
        </div>
      </div>
      
      <el-form :model="form" :rules="rules" ref="form" label-width="100px" class="voice-clone-form" label-position="left">
        <el-form-item label="说话人名称" prop="speaker_name">
          <el-input v-model="form.speaker_name" placeholder="请输入说话人名称"></el-input>
        </el-form-item>
        
        <el-form-item label="录制/上传音频" prop="audio_file">
          <div class="audio-upload-container">
            <!-- 上传按钮 -->
            <el-button 
              type="primary" 
              @click="showFileUpload = true" 
              class="upload-button">
              <i class="el-icon-upload2"></i> 上传音频
            </el-button>
            
            <!-- 录制按钮 -->
            <el-button 
              :type="isRecording ? 'danger' : 'success'" 
              @click="isRecording ? stopRecording() : startRecording()" 
              class="record-button">
              <i :class="isRecording ? 'el-icon-video-pause' : 'el-icon-video-play'"></i> 
              {{ isRecording ? '停止录制' : '开始录制' }}
            </el-button>
            
            <div v-if="showFileUpload" class="upload-area">
              <el-upload
                class="upload-demo"
                action="#"
                :http-request="uploadAudio"
                :limit="1"
                :file-list="fileList"
                :before-upload="beforeUpload">
                <el-button size="small" type="primary">选择文件</el-button>
              </el-upload>
              <div class="upload-tip">支持mp3/wav格式，不超过50MB</div>
            </div>
            
            <div v-if="form.audio_file" class="audio-selected">
              <div class="audio-file-info">
                <i class="el-icon-document"></i>
                <span class="file-name">{{ fileList.length > 0 ? fileList[0].name : '已选择音频文件' }}</span>
                <el-button type="text" @click="clearAudio" class="clear-btn">
                  <i class="el-icon-delete"></i>
                </el-button>
              </div>
            </div>
          </div>
          
          <!-- 录音预览 -->
          <div v-if="recordedAudio" class="recorded-audio-preview">
            <div class="audio-info">
              <span>录音时长: {{ recordingDuration }} 秒</span>
              <span class="audio-size">大小: {{ formatFileSize(recordedAudio.size) }}</span>
            </div>
            <audio :src="recordedAudioUrl" controls ref="audioPlayer" @error="handleAudioError" @canplay="handleCanPlay"></audio>
            <div class="preview-actions">
              <el-button size="small" type="primary" @click="useRecordedAudio">使用录制的音频</el-button>
              <el-button size="small" @click="discardRecordedAudio">放弃</el-button>
              <el-button size="small" type="info" @click="playRecordedAudio">播放</el-button>
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
        
        <!-- 移动端底部按钮 -->
        <div v-if="isMobile" class="mobile-form-footer">
          <el-button type="primary" :loading="submitting" @click="submitForm" class="mobile-submit-btn">创建任务</el-button>
        </div>
      </el-form>
      
      <!-- 桌面端底部按钮 -->
      <span v-if="!isMobile" slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import '@/assets/styles/card-view.css'
import axios from 'axios'

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
      observer: null,
      isMobile: false,
      lastScrollTop: 0,
      showFileUpload: false,
      recordingDuration: 0,
      audioPlayer: null
    }
  },
  created() {
    // 从本地存储中读取用户偏好的视图模式
    const savedViewMode = localStorage.getItem('voice_clone_view_mode')
    if (savedViewMode) {
      this.isCardView = savedViewMode === 'card'
    }
    
    // 检测设备类型
    this.checkDeviceType();
    
    // 如果是移动端，默认使用卡片视图
    if (this.isMobile) {
      this.isCardView = true;
    }
    
    // 初始加载数据
    this.loadInitialData();
    
    // 监听窗口大小变化
    window.addEventListener('resize', this.checkDeviceType);
  },
  
  mounted() {
    // 添加滚动事件监听器用于卡片视图加载更多
    window.addEventListener('scroll', this.handleWindowScroll)
    
    // 如果是卡片视图，初始化交叉观察器
    if (this.isCardView) {
      this.$nextTick(() => {
        this.setupIntersectionObserver()
      })
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
  },
  
  methods: {
    // 辅助方法: 调试日志
    debug(...args) {
      console.log('[VoiceClone]', ...args)
    },
    
    // 切换视图模式
    toggleView() {
      this.isCardView = !this.isCardView
      // 保存用户偏好到本地存储
      localStorage.setItem('voice_clone_view_mode', this.isCardView ? 'card' : 'list')
      
      // 当切换到卡片视图时，重置并加载数据
      if (this.isCardView) {
        this.currentPage = 1
        this.tasks = []
        this.hasMoreData = true
        this.fetchTasks()
        
        // 设置滚动监听
        this.$nextTick(() => {
          this.setupIntersectionObserver()
        })
      }
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
      
      const baseURL = process.env.VUE_APP_API_URL || ''
      const token = localStorage.getItem('token') || ''
      
      this.debug('请求数据:', '页码=', this.currentPage, '每页数量=', this.isCardView ? this.cardPageSize : this.pageSize)
      
      // 发送真实API请求
      axios.get(`${baseURL}/api/voice/clones`, {
        params: {
          page: this.currentPage,
          size: this.isCardView ? this.cardPageSize : this.pageSize
        },
        headers: { Authorization: `Bearer ${token}` }
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
          console.error('获取音色克隆任务列表失败:', error)
          this.$message.error('获取任务列表失败')
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
    
    // 设置IntersectionObserver用于检测滚动到底部
    setupIntersectionObserver() {
      if (!this.$refs.loadMoreTrigger) return
      
      // 如果已存在观察者，先断开连接
      if (this.observer) {
        this.observer.disconnect()
      }
      
      // 创建新的观察者
      this.observer = new IntersectionObserver(entries => {
        if (entries[0].isIntersecting && !this.loadingMore && this.hasMoreData) {
          this.loadMoreTasks()
        }
      }, { rootMargin: '0px 0px 200px 0px' })
      
      // 观察加载更多触发器
      this.observer.observe(this.$refs.loadMoreTrigger)
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
    
    // 处理滚动事件 - 优化为与TTS组件相同的方法
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
    
    // 加载更多任务
    loadMoreTasks() {
      if (this.loadingMore || !this.hasMoreData) {
        this.debug('跳过加载更多:', '加载中=', this.loadingMore, '没有更多数据=', !this.hasMoreData)
        return
      }
      
      this.debug('开始加载更多数据，当前页码：', this.currentPage)
      this.currentPage++
      this.fetchTasks(true)
    },
    
    // 格式化日期
    formatDate(dateString) {
      if (!dateString) return '-';
      const date = new Date(dateString);
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      }).replace(/\//g, '-');
    },
    
    // 获取状态类型
    getStatusType(status) {
      const statusMap = {
        'pending': 'info',
        'processing': 'warning',
        'completed': 'success',
        'failed': 'danger'
      };
      return statusMap[status] || 'info';
    },
    
    // 获取状态文本
    getStatusText(status) {
      const statusTextMap = {
        'pending': '等待中',
        'processing': '处理中',
        'completed': '已完成',
        'failed': '失败'
      };
      return statusTextMap[status] || '未知';
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
      
      // 重置表单验证
      if (this.$refs.form) {
        this.$refs.form.clearValidate()
      }
    },
    
    // 提交表单
    submitForm() {
      console.log('提交表单前检查:', {
        hasAudioFile: !!this.form.audio_file,
        fileName: this.form.audio_file ? this.form.audio_file.name : null,
        fileSize: this.form.audio_file ? this.form.audio_file.size : 0,
        fileType: this.form.audio_file ? this.form.audio_file.type : null,
        speakerName: this.form.speaker_name,
        promptText: this.form.prompt_text
      });
      
      this.$refs.form.validate((valid) => {
        if (!valid) {
          console.log('表单验证失败');
          return;
        }
        
        if (!this.form.audio_file) {
          console.error('表单中没有音频文件');
          return this.$message.error('请上传音频文件');
        }
        
        // 验证文件有效性
        console.log('提交的音频文件:', {
          name: this.form.audio_file.name,
          type: this.form.audio_file.type,
          size: this.form.audio_file.size,
        });
        
        if (this.form.audio_file.size === 0) {
          return this.$message.error('音频文件无效，请重新录制或上传');
        }
        
        // 检查文件格式
        const fileExt = this.form.audio_file.name.split('.').pop().toLowerCase();
        const allowedFormats = ['mp3', 'wav', 'webm', 'ogg', 'mpeg'];
        
        if (!allowedFormats.includes(fileExt)) {
          console.warn('文件扩展名不是常规音频格式:', fileExt);
          this.$message.warning('文件格式可能不受支持，将尝试提交');
        }
        
        this.submitting = true;
        const baseURL = process.env.VUE_APP_API_URL || '';
        const token = localStorage.getItem('token') || '';
        
        // 创建表单数据
        const formData = new FormData();
        // 生成任务名称：说话人名称 + 随机字符串
        const taskName = `${this.form.speaker_name}_${this.generateRandomString()}`;
        formData.append('name', taskName);
        formData.append('prompt_file', this.form.audio_file);
        formData.append('prompt_text', this.form.prompt_text);
        formData.append('speaker_name', this.form.speaker_name);
        
        console.log('发送请求前检查:', {
          endpoint: `${baseURL}/api/voice/clone`,
          taskName,
          textLength: this.form.prompt_text.length,
          speakerName: this.form.speaker_name
        });
        
        // 发送创建请求
        axios.post(`${baseURL}/api/voice/clone`, formData, {
          headers: { 
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'multipart/form-data'
          }
        })
          .then(response => {
            console.log('创建成功，响应:', response.data);
            this.$message.success('创建音色克隆任务成功');
            this.dialogVisible = false;
            
            // 刷新任务列表
            this.currentPage = 1;
            this.fetchTasks();
            
            // 跳转到详情页 - 确保路径正确
            if (response.data && response.data.voice_clone && response.data.voice_clone.id) {
              const id = response.data.voice_clone.id;
              // 和viewDetail方法保持一致的路径格式
              this.$router.push({
                path: `/voice-clone/${id}`,
                query: { from: 'create' }
              }).catch(err => {
                if (err.name !== 'NavigationDuplicated') {
                  console.error('导航错误', err);
                }
              });
            }
          })
          .catch(error => {
            console.error('创建音色克隆任务失败', error);
            
            let errorMsg = '未知错误';
            if (error.response) {
              console.error('错误响应:', {
                status: error.response.status,
                headers: error.response.headers,
                data: error.response.data
              });
              
              if (error.response.data && error.response.data.message) {
                errorMsg = error.response.data.message;
              } else if (error.response.status === 400) {
                errorMsg = '请求参数错误，请检查音频文件是否有效';
              } else if (error.response.status === 413) {
                errorMsg = '音频文件太大，超出服务器接收限制';
              } else if (error.response.status === 415) {
                errorMsg = '不支持的音频文件格式';
              }
            }
            
            this.$message.error('创建音色克隆任务失败: ' + errorMsg);
          })
          .finally(() => {
            this.submitting = false;
          });
      });
    },
    
    // 查看任务详情
    viewDetail(id) {
      // 记录当前任务ID以便于返回时恢复
      localStorage.setItem('last_voice_clone_id', id);
      
      // 确保ID是有效的
      if (!id) {
        this.$message.error('无效的任务ID');
        return;
      }
      
      // 跳转到详情页 - 根据路由配置使用正确的路径
      this.$router.push({
        path: `/voice-clone/${id}`,
        query: { from: 'list' }
      }).catch(err => {
        // 如果是重复导航错误，忽略它
        if (err.name !== 'NavigationDuplicated') {
          console.error('导航错误', err);
          this.$message.error('无法跳转到详情页');
        }
      });
    },
    
    // 确认删除
    confirmDelete(id) {
      this.$confirm('确定要删除此音色克隆任务吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.deleteTask(id);
      }).catch(() => {
        // 取消删除
      });
    },
    
    // 删除任务
    deleteTask(id) {
      this.loading = true;
      const baseURL = process.env.VUE_APP_API_URL || '';
      const token = localStorage.getItem('token') || '';
      
      // 发送删除请求
      axios.delete(`${baseURL}/api/voice/clone/${id}`, {
        headers: { Authorization: `Bearer ${token}` }
      })
        .then(() => {
          this.$message.success('删除成功');
          // 从本地列表中移除
          this.tasks = this.tasks.filter(task => task.id !== id);
          
          // 如果当前页的任务都删完了且不是第一页，则加载上一页数据
          if (this.tasks.length === 0 && this.currentPage > 1) {
            this.currentPage -= 1;
            this.fetchTasks();
          }
        })
        .catch(error => {
          console.error('删除失败', error);
          this.$message.error('删除失败');
        })
        .finally(() => {
          this.loading = false;
        });
    },

    // 上传前检查
    beforeUpload(file) {
      const isAudio = file.type.includes('audio');
      const isLt50M = file.size / 1024 / 1024 < 50;
      
      if (!isAudio) {
        this.$message.error('只能上传音频文件!');
        return false;
      }
      if (!isLt50M) {
        this.$message.error('音频文件大小不能超过 50MB!');
        return false;
      }
      
      return isAudio && isLt50M;
    },
    
    // 自定义上传方法
    uploadAudio(params) {
      const file = params.file;
      const isValid = this.beforeUpload(file);
      if (!isValid) return;
      
      // 更新文件列表显示
      this.fileList = [{ name: file.name, url: URL.createObjectURL(file) }];
      // 保存文件到表单
      this.form.audio_file = file;
      // 隐藏上传区域
      this.showFileUpload = false;
    },
    
    // 开始录音
    async startRecording() {
      try {
        // 检查麦克风权限
        if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
          this.$message.error('您的浏览器不支持录音功能，请使用Chrome或Firefox等现代浏览器');
          return;
        }

        // 请求麦克风权限
        this.$message.info('正在请求麦克风权限...');
        
        const stream = await navigator.mediaDevices.getUserMedia({ 
          audio: {
            echoCancellation: true,
            noiseSuppression: true,
            autoGainControl: true
          } 
        });
        
        this.$message.success('麦克风权限获取成功');
        this.isRecording = true;
        this.recordedChunks = [];
        this.recordingStartTime = Date.now(); // 记录开始时间
        
        // 优先尝试使用wav格式（更兼容）
        let supportedMimeTypes = [
          'audio/wav',
          'audio/mp3',
          'audio/webm;codecs=opus',
          'audio/webm',
          'audio/ogg'
        ];
        
        let mimeType = '';
        for (let type of supportedMimeTypes) {
          if (MediaRecorder.isTypeSupported(type)) {
            mimeType = type;
            console.log('使用的录音MIME类型:', mimeType);
            break;
          }
        }
        
        // 设置媒体录制器选项
        const options = mimeType ? { mimeType, audioBitsPerSecond: 128000 } : {};
        
        try {
          this.mediaRecorder = new MediaRecorder(stream, options);
          console.log('MediaRecorder创建成功，使用MIME类型:', this.mediaRecorder.mimeType);
        } catch (err) {
          console.warn('创建MediaRecorder失败，尝试不指定MIME类型', err);
          this.mediaRecorder = new MediaRecorder(stream);
        }
        
        // 录音数据处理
        this.mediaRecorder.ondataavailable = (event) => {
          console.log('收到录音数据块，大小:', event.data.size);
          if (event.data.size > 0) {
            this.recordedChunks.push(event.data);
          }
        };
        
        // 录音结束处理
        this.mediaRecorder.onstop = () => {
          try {
            // 判断是否有录音数据
            if (this.recordedChunks.length === 0) {
              this.$message.warning('没有录到任何声音，请检查麦克风');
              return;
            }
            
            // 计算录音时长
            const recordingDuration = (Date.now() - this.recordingStartTime) / 1000;
            this.recordingDuration = Math.round(recordingDuration);
            console.log(`录音完成，时长约 ${this.recordingDuration} 秒`);
            
            // 确定MIME类型，优先使用mp3或wav
            let mimeType = this.mediaRecorder.mimeType;
            
            // 如果录制的不是mp3或wav，我们尝试将其标记为mp3
            // 注意：这并不会转换音频格式，只是改变了文件类型标记
            if (!mimeType.includes('mp3') && !mimeType.includes('wav')) {
              console.log('录制的格式不是mp3或wav，将尝试标记为mp3格式');
              mimeType = 'audio/mpeg';
            }
            
            console.log('创建音频Blob，使用MIME类型:', mimeType);
            const audioBlob = new Blob(this.recordedChunks, { type: mimeType });
            this.recordedAudio = audioBlob;
            this.recordedAudioUrl = URL.createObjectURL(audioBlob);
            
            // 显示文件信息
            console.log('录音文件创建成功:', {
              type: audioBlob.type,
              size: this.formatFileSize(audioBlob.size),
              url: this.recordedAudioUrl
            });
            
            // 停止所有音轨
            stream.getTracks().forEach(track => track.stop());
            
            this.$message.success(`录音完成! 时长: ${this.recordingDuration}秒`);
          } catch (error) {
            console.error('处理录音数据出错:', error);
            this.$message.error('处理录音失败: ' + error.message);
          }
        };
        
        // 录音错误处理
        this.mediaRecorder.onerror = (event) => {
          console.error('录音时发生错误:', event.error);
          this.$message.error(`录音错误: ${event.error.message || '未知错误'}`);
          this.stopRecording();
        };
        
        // 开始录音
        this.mediaRecorder.start(1000); // 每秒获取一次录音数据
        this.$message.info('开始录音...');
      } catch (error) {
        console.error('录音失败:', error);
        
        // 更具体的错误信息
        let errorMessage = '无法访问麦克风';
        if (error.name === 'NotAllowedError' || error.name === 'PermissionDeniedError') {
          errorMessage = '麦克风权限被拒绝，请在浏览器设置中允许访问麦克风';
        } else if (error.name === 'NotFoundError' || error.name === 'DevicesNotFoundError') {
          errorMessage = '未检测到麦克风设备，请确保麦克风已连接';
        } else if (error.name === 'NotReadableError' || error.name === 'TrackStartError') {
          errorMessage = '麦克风被其他应用占用，请关闭其他使用麦克风的应用';
        } else if (error.name === 'OverconstrainedError') {
          errorMessage = '无法满足音频录制要求，请检查麦克风设置';
        } else if (error.name === 'TypeError') {
          errorMessage = '浏览器不支持此录音格式';
        }
        
        this.$message.error(errorMessage);
        this.isRecording = false;
      }
    },
    
    // 停止录音
    stopRecording() {
      if (this.mediaRecorder && this.isRecording) {
        try {
          this.mediaRecorder.stop();
        } catch (error) {
          console.error('停止录音时出错:', error);
        }
        this.isRecording = false;
      }
    },
    
    // 使用录制的音频
    useRecordedAudio() {
      if (this.recordedAudio) {
        try {
          console.log('录音对象信息:', {
            type: this.recordedAudio.type,
            size: this.recordedAudio.size,
            chunks: this.recordedChunks.length
          });
          
          // 默认使用MP3格式，这是服务器最常接受的格式
          let fileExt = 'mp3';
          let mimeType = 'audio/mpeg';
          
          // 但如果浏览器使用了其他格式录制，我们需要尊重这个格式
          // 同时确保扩展名与MIME类型匹配
          if (this.recordedAudio.type) {
            if (this.recordedAudio.type.includes('wav')) {
              fileExt = 'wav';
              mimeType = 'audio/wav';
            } else if (this.recordedAudio.type.includes('mp3') || this.recordedAudio.type.includes('mpeg')) {
              fileExt = 'mp3';
              mimeType = 'audio/mpeg';
            } else if (this.recordedAudio.type.includes('webm')) {
              fileExt = 'mp3'; // 强制转换为mp3
              mimeType = 'audio/mpeg';
            } else if (this.recordedAudio.type.includes('ogg')) {
              fileExt = 'mp3'; // 强制转换为mp3
              mimeType = 'audio/mpeg';
            }
          }
          
          // 生成带时间戳的文件名，并包含说话人名称
          const timestamp = new Date().getTime();
          const speakerName = (this.form.speaker_name || '录音').replace(/\s+/g, '_');
          const fileName = `${speakerName}_${timestamp}.${fileExt}`;
          
          console.log('将使用以下参数创建文件:', {
            fileName: fileName,
            mimeType: mimeType,
            blobType: this.recordedAudio.type,
            blobSize: this.recordedAudio.size
          });
          
          // 创建文件对象
          const file = new File([this.recordedAudio], fileName, { 
            type: mimeType,
            lastModified: Date.now()
          });
          
          // 验证文件是否创建成功
          if (!file || file.size === 0) {
            throw new Error('创建文件失败或文件大小为0');
          }
          
          console.log('文件创建成功:', {
            name: file.name,
            type: file.type,
            size: file.size
          });
          
          // 直接设置表单数据
          this.form.audio_file = file;
          this.fileList = [{ name: fileName, url: this.recordedAudioUrl }];
          this.$message.success(`录音已设置为当前音频: ${fileName}`);
          
          // 显示调试信息
          console.log('表单数据已更新:', {
            hasAudioFile: !!this.form.audio_file,
            fileName: this.form.audio_file ? this.form.audio_file.name : null,
            fileSize: this.form.audio_file ? this.form.audio_file.size : 0,
            fileType: this.form.audio_file ? this.form.audio_file.type : null
          });
          
          // 这样写可以检查文件对象是否真正设置
          setTimeout(() => {
            console.log('500ms后检查表单状态:', {
              hasAudioFile: !!this.form.audio_file,
              audioFileName: this.form.audio_file ? this.form.audio_file.name : null
            });
          }, 500);
        } catch (error) {
          console.error('处理录音出错:', error);
          this.$message.error('处理录音失败: ' + error.message);
        }
      } else {
        this.$message.warning('没有可用的录音');
      }
    },
    
    // 放弃录制的音频
    discardRecordedAudio() {
      if (this.recordedAudioUrl) {
        URL.revokeObjectURL(this.recordedAudioUrl);
      }
      this.recordedAudio = null;
      this.recordedAudioUrl = null;
      this.recordedChunks = [];
    },
    
    // 分页相关方法
    handleSizeChange(size) {
      this.pageSize = size;
      this.currentPage = 1;
      this.tasks = []; // 清空现有数据
      this.fetchTasks();
    },
    
    handleCurrentChange(page) {
      this.currentPage = page;
      this.tasks = []; // 清空现有数据
      this.fetchTasks();
    },
    
    // 清除已选音频
    clearAudio() {
      this.form.audio_file = null;
      this.fileList = [];
      this.showFileUpload = false;
    },
    
    // 处理音频错误
    handleAudioError(event) {
      console.error('音频播放错误:', event.error);
      this.$message.error('音频播放错误');
    },
    
    // 处理音频可播放事件
    handleCanPlay() {
      const audio = this.$refs.audioPlayer;
      if (audio) {
        this.recordingDuration = Math.round(audio.duration);
      }
    },
    
    // 播放录制的音频
    playRecordedAudio() {
      const audio = this.$refs.audioPlayer;
      if (audio) {
        audio.play();
      }
    },
    
    // 格式化文件大小
    formatFileSize(size) {
      if (size < 1024) {
        return size + ' B';
      } else if (size < 1024 * 1024) {
        return (size / 1024).toFixed(2) + ' KB';
      } else if (size < 1024 * 1024 * 1024) {
        return (size / (1024 * 1024)).toFixed(2) + ' MB';
      } else {
        return (size / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
      }
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

/* 录音相关样式 */
.audio-upload-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
  margin-bottom: 15px;
}

.recorded-audio-preview {
  margin-top: 15px;
  background-color: #f4f9ff;
  border-radius: 8px;
  padding: 15px;
  border: 1px solid #e0efff;
}

.preview-actions {
  display: flex;
  gap: 10px;
}

.preview-actions .el-button {
  flex: 1;
  height: 36px;
}

/* 移动端样式优化 */
/* 移动端底部菜单 */
.mobile-footer-menu {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(20, 20, 40, 0.95);
  display: flex;
  height: 60px;
  justify-content: center;
  align-items: center;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(10px);
  z-index: 1000;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.mobile-footer-menu .menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 5px 20px;
  color: #2196f3;
  transition: all 0.3s;
  cursor: pointer;
  font-weight: bold;
}

.mobile-footer-menu .menu-item i {
  font-size: 24px;
  margin-bottom: 3px;
}

.mobile-footer-menu .menu-item span {
  font-size: 14px;
}

.mobile-footer-menu .menu-item.active {
  color: #2196f3;
  font-weight: bold;
}

.mobile-footer-menu .menu-item.active i {
  transform: scale(1.1);
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
  font-size: 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  cursor: pointer;
  transition: all 0.3s;
  z-index: 99;
}

.floating-add-btn:hover, .floating-add-btn:active {
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

/* 响应式样式 */
@media screen and (max-width: 768px) {
  .voice-clone-container {
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
  
  /* 隐藏在移动端不重要的表格列 */
  .hide-on-mobile {
    display: none;
  }
  
  /* 移动端卡片容器优化 */
  .mobile-card-container {
    grid-template-columns: repeat(2, 1fr) !important;
    gap: 8px !important;
    padding: 8px;
  }
  
  /* 移动端头部固定 */
  .mobile-header {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 1000;
    border-radius: 0;
    margin-bottom: 0;
  }
}

/* 移动端对话框样式优化 */
@media screen and (max-width: 768px) {
  .el-dialog {
    margin: 0 !important;
    border-radius: 0 !important;
    display: flex !important;
    flex-direction: column !important;
    height: 100vh !important;
    width: 100% !important;
    overflow: hidden !important;
    position: fixed !important;
    top: 0 !important;
    left: 0 !important;
    background-color: #fff !important;
  }
  
  .el-dialog__body {
    padding: 0 !important;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    flex: 1;
    padding-top: 50px !important; /* 为顶部导航留出空间 */
  }
  
  /* 确保移动端全屏时，不会出现双重标题 */
  .el-dialog__header {
    padding: 0 !important;
    height: 0 !important;
    display: none !important;
  }
}

/* 防止遮罩导致页面无法点击 */
.el-dialog__wrapper {
  overflow: auto !important;
}

/* 移除遮罩效果 */
:deep(.v-modal) {
  opacity: 0 !important;
  background-color: transparent !important;
}

/* 确保元素可点击 */
.voice-clone-dialog,
.voice-clone-form,
.audio-upload-container,
.el-button {
  opacity: 1 !important;
  pointer-events: auto !important;
}

/* 录音按钮样式优化 */
.audio-upload-container .el-button {
  transition: all 0.3s;
}

.audio-upload-container .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.audio-upload-container .el-button:active {
  transform: translateY(0);
}

/* 录音预览样式优化 */
.recorded-audio-preview audio {
  width: 100%;
  margin-bottom: 10px;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.2);
}

/* 表单输入框样式优化 */
.el-input__inner,
.el-textarea__inner {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
  transition: all 0.3s;
}

.el-input__inner:focus,
.el-textarea__inner:focus {
  background: rgba(255, 255, 255, 0.15);
  border-color: #409EFF;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.el-input__inner::placeholder,
.el-textarea__inner::placeholder {
  color: rgba(255, 255, 255, 0.5);
}

/* 对话框按钮样式优化 */
.el-dialog__footer .el-button {
  transition: all 0.3s;
}

.el-dialog__footer .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.el-dialog__footer .el-button:active {
  transform: translateY(0);
}

/* 移动端对话框样式 - 简化并修复 */
.voice-clone-dialog {
  background: #fff; /* 使用白色背景 */
}

.mobile-header-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 50px;
  background-color: #fff;
  display: flex;
  align-items: center;
  padding: 0 15px;
  z-index: 100;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.header-back {
  display: flex;
  align-items: center;
  color: #409EFF;
  font-size: 16px;
  cursor: pointer;
}

.header-back i {
  margin-right: 5px;
}

/* 表单样式 */
.voice-clone-form {
  padding: 20px 15px 100px;
}

/* 录音和上传按钮 */
.upload-button, 
.record-button {
  width: 100%;
  height: 42px;
  font-size: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.3s;
  margin-bottom: 10px;
  padding: 0 16px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.upload-button i, 
.record-button i {
  margin-right: 8px;
  font-size: 16px;
}

/* 音频上传容器 */
.audio-upload-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
  margin-bottom: 15px;
}

/* 上传区域样式 */
.upload-area {
  background-color: #f8f9fa;
  border-radius: 4px;
  padding: 15px;
  border: 1px dashed #dcdfe6;
  margin-top: 5px;
}

/* 已选择音频文件区域 */
.audio-selected {
  padding: 12px 15px;
  background-color: #f0f9ff;
  border-radius: 6px;
  border: 1px solid #cce7ff;
  margin-top: 5px;
}

.audio-file-info {
  display: flex;
  align-items: center;
}

.audio-file-info i {
  font-size: 18px;
  color: #409EFF;
}

.file-name {
  margin-left: 10px;
  flex: 1;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.clear-btn {
  padding: 4px 8px;
}

.clear-btn i {
  font-size: 16px;
}

.upload-tip {
  margin-top: 8px;
  color: #909399;
  font-size: 12px;
}

/* 录音预览 */
.recorded-audio-preview {
  margin-top: 15px;
  background-color: #f4f9ff;
  border-radius: 8px;
  padding: 15px;
  border: 1px solid #e0efff;
}

.recorded-audio-preview audio {
  width: 100%;
  margin-bottom: 15px;
  border-radius: 6px;
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.preview-actions {
  display: flex;
  gap: 10px;
}

.preview-actions .el-button {
  flex: 1;
  height: 36px;
}

/* 移动端特定优化 */
@media screen and (max-width: 768px) {
  /* 顶部导航栏优化 */
  .mobile-header-bar {
    background-color: #409EFF;
    color: #fff;
    height: 56px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  }
  
  .header-back {
    color: #fff;
    font-size: 16px;
  }
  
  /* 表单布局优化 */
  .el-form-item {
    display: block; /* 改为块级显示以防止错位 */
    margin-bottom: 20px;
    padding: 0 15px;
  }
  
  .el-form-item__label {
    display: block;
    width: 100% !important;
    text-align: left !important;
    line-height: 1.5;
    padding: 0;
    margin-bottom: 10px;
    font-size: 15px;
    color: #303133;
    font-weight: 500;
  }
  
  .el-form-item__content {
    margin-left: 0 !important;
    padding-left: 0;
    width: 100%;
  }
  
  /* 输入框样式 */
  .el-input__inner {
    height: 44px;
    font-size: 15px;
  }
  
  .el-textarea__inner {
    min-height: 100px;
    font-size: 15px;
    line-height: 1.6;
  }
  
  /* 按钮样式调整 */
  .upload-button, 
  .record-button {
    font-size: 14px;
  }
  
  /* 底部按钮 */
  .mobile-form-footer {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    padding: 0;
    background-color: #fff;
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
    z-index: 100;
  }
  
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
  
  /* 调整表单内容区，避免被底部按钮遮挡 */
  .voice-clone-form {
    padding-bottom: 80px;
  }
  
  /* 音频容器确保不溢出 */
  .audio-upload-container {
    max-width: 100%;
    overflow: hidden;
  }
  
  /* 改进音频播放器样式 */
  audio {
    width: 100%;
    height: 40px;
    border-radius: 20px;
    outline: none;
  }
  
  /* 确保所有元素可见 */
  .el-button, 
  .audio-upload-container, 
  .upload-button, 
  .record-button {
    opacity: 1 !important;
    visibility: visible !important;
    display: flex !important;
  }
}

/* 输入框美化 */
.voice-clone-dialog :deep(.el-input__inner) {
  border-radius: 4px;
  border: 1px solid #dcdfe6;
  background-color: #fff;
  transition: border-color 0.2s cubic-bezier(0.645, 0.045, 0.355, 1);
  padding: 0 15px;
}

.voice-clone-dialog :deep(.el-input__inner:focus),
.voice-clone-dialog :deep(.el-textarea__inner:focus) {
  border-color: #409EFF;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.voice-clone-dialog :deep(.el-textarea__inner) {
  padding: 10px 15px;
  line-height: 1.5;
  border-radius: 4px;
}

/* 确保遮罩不影响交互 */
:deep(.v-modal) {
  opacity: 0 !important;
}

/* 确保对话框能全屏显示 */
.el-dialog {
  display: flex;
  flex-direction: column;
}

.el-dialog__body {
  flex: 1;
  overflow-y: auto;
}

/* 表单布局 */
.voice-clone-form {
  padding: 20px 15px 100px;
}
</style>