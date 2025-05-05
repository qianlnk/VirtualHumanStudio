<template>
  <div class="digital-human-container">
    <div class="page-header">
      <div class="header-left">
        <h2>数字人合成</h2>
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
      <el-empty v-if="digitalHumans.length === 0" description="暂无数字人合成任务"></el-empty>
      
      <el-table v-else :data="digitalHumans" style="width: 100%" class="responsive-table">
        <el-table-column prop="name" label="任务名称" min-width="300" show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="text-ellipsis">{{ scope.row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">{{ getStatusText(scope.row.status) }}</el-tag>
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
                @click="downloadResult(scope.row.id)" 
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
    
    <!-- 卡片视图 -->
    <div v-show="isCardView" class="card-list" v-loading="loading">
      <el-empty v-if="digitalHumans.length === 0" description="暂无数字人合成任务"></el-empty>
      
      <div v-else class="card-view-content">
        <div class="waterfall-container">
          <div class="task-card" v-for="item in digitalHumans" :key="item.id">
            <div class="task-card-header">
              <h3 class="task-card-title">{{ item.name }}</h3>
              <el-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
            </div>
            <div class="task-card-content">
              <div class="task-card-info">
                <p><i class="el-icon-time"></i> {{ formatDate(item.created_at) }}</p>
              </div>
            </div>
            <div class="task-card-footer">
              <el-button type="text" size="small" class="action-btn" @click="viewDetail(item.id)">查看</el-button>
              <el-button 
                type="text" 
                size="small" 
                class="action-btn" 
                @click="downloadResult(item.id)" 
                :disabled="item.status !== 'completed'"
              >下载</el-button>
              <el-button type="text" size="small" class="action-btn" @click="confirmDelete(item.id)">删除</el-button>
            </div>
          </div>
        </div>
        
        <!-- 加载更多提示 -->
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
    
    <!-- 创建数字人任务对话框 -->
    <el-dialog title="创建数字人合成任务" :visible.sync="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="form" label-width="100px">
        
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称"></el-input>
        </el-form-item>
        
        <el-form-item label="音频文件" prop="audio_file">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleAudioChange"
            :limit="1"
            :file-list="audioFileList">
            <el-button size="small" type="primary">选择音频文件</el-button>
            <div slot="tip" class="el-upload__tip">只能上传WAV/MP3文件</div>
          </el-upload>
        </el-form-item>
        
        <el-form-item label="视频文件" prop="video_file">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleVideoChange"
            :limit="1"
            :file-list="videoFileList">
            <el-button size="small" type="primary">选择视频文件</el-button>
            <div slot="tip" class="el-upload__tip">只能上传MP4文件</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">创建</el-button>
      </div>
    </el-dialog>
    
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
  </div>
</template>

<script>
import axios from 'axios'
import { v4 as uuidv4 } from 'uuid'
import '@/assets/styles/card-view.css'

export default {
  name: 'DigitalHuman',
  data() {
    return {
      loading: false,
      submitting: false,
      dialogVisible: false,
      digitalHumans: [],
      currentPage: 1,
      pageSize: 10,
      cardPageSize: 10,
      total: 0,
      isCardView: false, // 是否使用卡片视图
      loadingMore: false,
      hasMoreData: true,
      initialLoaded: false,
      scrollThreshold: 200, // 滚动阈值，距离底部多少像素时触发加载
      observer: null, // IntersectionObserver实例
      form: {
        name: '',
        description: '',
        task_code: '',
        chaofen: 0,
        watermark_switch: 0,
        pn: 1,
        audio_file: null,
        video_file: null
      },
      rules: {
        name: [
          { required: true, message: '请输入任务名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ],
        audio_file: [
          { required: true, message: '请上传音频文件', trigger: 'change' }
        ],
        video_file: [
          { required: true, message: '请上传视频文件', trigger: 'change' }
        ]
      },
      audioFileList: [],
      videoFileList: []
    }
  },
  created() {
    // 从本地存储中读取用户的视图偏好设置
    const viewMode = localStorage.getItem('digital_human_view_mode')
    if (viewMode) {
      this.isCardView = viewMode === 'card'
    } else {
      // 移动设备默认使用卡片视图
      this.isCardView = window.innerWidth <= 768
    }
    
    this.debug('组件创建，开始加载初始数据')
    this.loadInitialData()
  },
  
  mounted() {
    // 添加滚动事件监听器
    window.addEventListener('scroll', this.handleScroll)
  },
  
  beforeDestroy() {
    // 移除滚动事件监听器
    window.removeEventListener('scroll', this.handleScroll)
    // 清除IntersectionObserver
    if (this.observer) {
      this.observer.disconnect()
      this.observer = null
    }
  },
  methods: {
    // 辅助方法: 调试日志
    debug(...args) {
      console.log('[DigitalHuman]', ...args)
    },
    
    // 初始加载数据
    loadInitialData() {
      if (this.initialLoaded) {
        this.debug('已加载初始数据，跳过')
        return
      }
      
      this.debug('加载初始数据')
      this.currentPage = 1
      this.digitalHumans = []
      this.hasMoreData = true
      this.fetchDigitalHumans()
      this.initialLoaded = true
    },
    
    // 获取数字人列表
    async fetchDigitalHumans(loadMore = false) {
      if ((this.loading && !loadMore) || (loadMore && this.loadingMore)) {
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
        const response = await axios.get('/api/digital-human', {
          params: {
            page: this.currentPage,
            page_size: this.isCardView ? this.cardPageSize : this.pageSize
          }
        })
        
        const newTasks = response.data.digital_humans || []
        this.total = response.data.total || 0
        
        this.debug('获取到新数据:', newTasks.length, '总数:', this.total)
        
        if (loadMore) {
          // 追加新数据
          this.digitalHumans = [...this.digitalHumans, ...newTasks]
        } else {
          // 重置数据
          this.digitalHumans = newTasks
        }
        
        // 判断是否还有更多数据
        this.hasMoreData = this.digitalHumans.length < this.total
        this.debug('当前数据量：', this.digitalHumans.length, '总数：', this.total, '是否还有更多：', this.hasMoreData)
        
        // 如果加载的数据少于请求的数量，说明没有更多数据了
        if (newTasks.length < (this.isCardView ? this.cardPageSize : this.pageSize)) {
          this.hasMoreData = false
        }
      } catch (error) {
        console.error('获取数字人列表失败:', error)
        this.$message.error('获取数字人列表失败')
      } finally {
        this.loading = false
        this.loadingMore = false
        
        // 在数据加载完成后，重新设置IntersectionObserver
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
      this.fetchDigitalHumans(true)
    },
    
    // 处理滚动事件
    handleScroll() {
      if (!this.isCardView || this.loadingMore || !this.hasMoreData) return
      
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop
      const windowHeight = window.innerHeight
      const documentHeight = Math.max(
        document.body.scrollHeight, document.documentElement.scrollHeight,
        document.body.offsetHeight, document.documentElement.offsetHeight,
        document.body.clientHeight, document.documentElement.clientHeight
      )
      
      // 当滚动到距离底部阈值距离时触发加载
      if (documentHeight - scrollTop - windowHeight < this.scrollThreshold) {
        this.debug('窗口滚动触发加载更多:', '文档高度=', documentHeight, '滚动位置=', scrollTop, '窗口高度=', windowHeight)
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
    
    // 切换列表/卡片视图
    toggleView() {
      this.isCardView = !this.isCardView
      // 保存用户偏好到本地存储
      localStorage.setItem('digital_human_view_mode', this.isCardView ? 'card' : 'list')
      
      // 重置分页并重新加载数据
      this.currentPage = 1
      this.digitalHumans = []
      this.hasMoreData = true
      this.fetchDigitalHumans()
      
      // 在DOM更新后设置IntersectionObserver
      if (this.isCardView) {
        this.$nextTick(() => {
          this.setupIntersectionObserver()
        })
      }
    },
    
    // 显示创建对话框
    showCreateDialog() {
      this.dialogVisible = true
      this.resetForm()
    },
    
    // 重置表单
    resetForm() {
      if (this.$refs.form) {
        this.$refs.form.resetFields()
      }
      const uuid = uuidv4()
      this.form = {
        name: '',
        description: '',
        task_code: uuid,
        chaofen: 0,
        watermark_switch: 0,
        pn: 1,
        audio_file: null,
        video_file: null
      }
      this.audioFileList = []
      this.videoFileList = []
    },
    
    // 处理音频文件变化
    handleAudioChange(file) {
      this.form.audio_file = file.raw
    },
    
    // 处理视频文件变化
    handleVideoChange(file) {
      this.form.video_file = file.raw
    },
    
    // 提交表单
    submitForm() {
      this.$refs.form.validate(valid => {
        if (valid) {
          this.createDigitalHuman()
        } else {
          return false
        }
      })
    },
    
    // 创建数字人任务
    createDigitalHuman() {
      this.submitting = true
      
      // 创建FormData对象
      const formData = new FormData()
      formData.append('name', this.form.name)
      formData.append('description', this.form.description)
      formData.append('task_code', this.form.task_code)
      formData.append('chaofen', this.form.chaofen)
      formData.append('watermark_switch', this.form.watermark_switch)
      formData.append('pn', this.form.pn)
      formData.append('audio_file', this.form.audio_file)
      formData.append('video_file', this.form.video_file)
      
      // 发送请求
      axios.post('/api/digital-human', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
        .then(() => {
          this.$message.success('创建数字人任务成功')
          this.dialogVisible = false
          this.fetchDigitalHumans()
        })
        .catch(error => {
          console.error('创建数字人任务失败:', error)
          this.$message.error('创建数字人任务失败: ' + ((error.response && error.response.data && error.response.data.error) || error.message))
        })
        .finally(() => {
          this.submitting = false
        })
    },
    
    // 查看详情
    viewDetail(id) {
      this.$router.push(`/digital-human/${id}`)
    },
    
    // 下载结果
    downloadResult(id) {
      window.open(`/api/digital-human/${id}/download`, '_blank')
    },
    
    // 确认删除
    confirmDelete(id) {
      this.$confirm('确定要删除这个数字人任务吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.deleteDigitalHuman(id)
      }).catch(() => {
        // 取消删除
      })
    },
    
    // 删除数字人任务
    deleteDigitalHuman(id) {
      axios.delete(`/api/digital-human/${id}`)
        .then(() => {
          this.$message.success('删除成功')
          this.fetchDigitalHumans()
        })
        .catch(error => {
          console.error('删除失败:', error)
          this.$message.error('删除失败: ' + ((error.response && error.response.data && error.response.data.error) || error.message))
        })
    },
    
    // 处理分页大小变化
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchDigitalHumans()
    },
    
    // 处理页码变化
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchDigitalHumans()
    },
    
    // 格式化日期
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    },
    
    // 获取状态类型
    getStatusType(status) {
      switch (status) {
        case 'completed': return 'success'
        case 'processing': return 'warning'
        case 'pending': return 'info'
        case 'failed': return 'danger'
        default: return 'info'
      }
    },
    
    // 获取状态文本
    getStatusText(status) {
      switch (status) {
        case 'completed': return '已完成'
        case 'processing': return '处理中'
        case 'pending': return '等待中'
        case 'failed': return '失败'
        default: return '未知'
      }
    }
  }
}
</script>

<style scoped>
.digital-human-container {
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

/* 移动端底部菜单 */
.mobile-footer-menu {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: none;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(10px);
  z-index: 100;
  padding: 10px 0;
  justify-content: space-around;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.3);
}

.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 16px;
  color: #ddd;
  cursor: pointer;
  transition: all 0.3s;
}

.menu-item i {
  font-size: 18px;
  margin-bottom: 4px;
}

.menu-item span {
  font-size: 12px;
}

.menu-item:hover, .menu-item.active {
  color: #409EFF;
}

@media screen and (max-width: 768px) {
  .mobile-footer-menu {
    display: flex;
  }
  
  .task-list, .card-list {
    padding-bottom: 70px;
  }
}
</style>