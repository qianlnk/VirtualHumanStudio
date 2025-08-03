<template>
  <div class="digital-human-template-container">
    <div class="page-header" :class="{'mobile-header': isMobile}">
      <div class="header-left">
        <h2>数字人模版</h2>
      </div>
      <div class="header-right">
        <el-button v-if="!isMobile" type="primary" @click="showCreateDialog" icon="el-icon-plus">创建模版</el-button>
        <el-button v-if="!isMobile" type="text" size="small" class="view-toggle" @click="toggleView">
          <i :class="isCardView ? 'el-icon-menu' : 'el-icon-s-grid'"></i>
          <span class="toggle-text">{{ isCardView ? '列表视图' : '卡片视图' }}</span>
        </el-button>
      </div>
    </div>

    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>
    
    <!-- 任务列表 -->
    <div v-loading="loading" class="template-list" v-show="!isCardView">
      <el-empty v-if="templates.length === 0" description="暂无数字人模版"></el-empty>

      <el-table v-else :data="templates" style="width: 100%" class="responsive-table" ref="dataTable">
        <el-table-column prop="name" label="模版名称" min-width="300" show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="text-ellipsis">{{ scope.row.name || '未命名任务' }}</span>
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
        <el-table-column label="操作" width="200">
          <template slot-scope="scope">
            <div class="action-buttons">
              <el-button 
                type="text" 
                size="mini" 
                class="action-btn"
                @click="viewDetail(scope.row)"
              >查看</el-button>
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
          :current-page.sync="currentPage"
          :page-sizes="[10, 20, 30, 50]"
          :page-size="pageSize"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total">
        </el-pagination>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-show="isCardView" class="card-list" v-loading="loading">
      <el-empty v-if="templates.length === 0" description="暂无数字人模版"></el-empty>

      <div v-else class="card-view-content" ref="cardViewContent">
        <div class="waterfall-container" ref="cardContainer" :class="{'mobile-card-container': isMobile}">
          <div class="task-card" v-for="(item, index) in templates" :key="item.id || index" :ref="`templateCard_${item.id || index}`">
            <div class="task-card-header">
              <h3 class="task-card-title">{{ item.name || '未命名模版' }}</h3>
              <div class="status-icon">
                <i v-if="item.status === 'completed'" class="el-icon-check" style="color: #67c23a;"></i>
                <i v-else-if="item.status === 'failed'" class="el-icon-close" style="color: #f56c6c;"></i>
                <i v-else-if="item.status === 'processing'" class="el-icon-loading" style="color: #e6a23c;"></i>
                <i v-else-if="item.status === 'deleted'" class="el-icon-close" style="color: #f56c6c;"></i>
                <el-tag v-else :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
              </div>
            </div>
            <div class="task-card-content">
              <div class="task-card-info">
                <p><i class="el-icon-time"></i> {{ formatDate(item.created_at) }}</p>
                <p v-if="item.description"><i class="el-icon-document"></i> {{ item.description }}</p>
              </div>
            </div>
            <div class="task-card-footer">
              <el-button type="text" size="small" class="action-btn" @click="viewDetail(item)">查看</el-button>
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

    <!-- 创建模版对话框 -->
    <el-dialog 
      title="创建数字人模版" 
      :visible.sync="dialogVisible" 
      :fullscreen="isMobile"
      :modal="true"
      :close-on-click-modal="false"
      :append-to-body="true"
      :show-close="!isMobile"
      custom-class="digital-human-dialog"
      width="600px">

      <!-- 移动端顶部导航 -->
      <div v-if="isMobile" class="mobile-header-bar">
        <div class="header-back" @click="dialogVisible = false">
          <i class="el-icon-arrow-left"></i>
          <span>返回</span>
        </div>
      </div>

      <el-form 
        :model="form" 
        :rules="rules" 
        ref="form" 
        :label-width="isMobile ? '90px' : '100px'"
        :label-position="isMobile ? 'top' : 'left'"
        class="digital-human-template-form">

        <el-form-item label="模版名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入模版名称"/>
        </el-form-item>
        <el-form-item label="原始视频" prop="video_file">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleVideoChange"
            :limit="1"
            :file-list="videoFileList">
            <el-button size="small" type="primary">选择视频</el-button>
            <div slot="tip" class="el-upload__tip">只能上传MP4文件</div>
          </el-upload>
        </el-form-item>
        <el-form-item label="人脸图片">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleFaceChange"
            :limit="1"
            :file-list="faceFileList">
            <el-button size="small">选择图片</el-button>
            <div slot="tip" class="el-upload__tip">可选，仅支持JPG/PNG</div>
          </el-upload>
        </el-form-item>
        <el-form-item label="背景">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleBgChange"
            :limit="1"
            :file-list="bgFileList">
            <el-button size="small">选择文件</el-button>
            <div slot="tip" class="el-upload__tip">可选，支持JPG/PNG/MP4</div>
          </el-upload>
        </el-form-item>

        <!-- 移动端底部按钮 -->
        <div v-if="isMobile" class="mobile-form-footer">
          <el-button type="primary" @click="submitForm" :loading="submitting" class="mobile-submit-btn">创建任务</el-button>
        </div>
      </el-form>
      
      <!-- 桌面端底部按钮 -->
      <div v-if="!isMobile" slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">提交</el-button>
      </div>
    </el-dialog>
    
    <!-- 移动端悬浮添加按钮 -->
    <div v-if="isMobile" class="floating-add-btn" @click="showCreateDialog">
      <i class="el-icon-plus"></i>
    </div>
  </div>
</template>

<script>
import '@/assets/styles/card-view.css'

export default {
  name: 'DigitalHumanTemplate',
  data() {
    return {
      loading: false,
      submitting: false,
      dialogVisible: false,
      templates: [],
      currentPage: 1,
      pageSize: 10,
      cardPageSize: 10,
      total: 0,
      isCardView: false,
      loadingMore: false,
      hasMoreData: true,
      initialLoaded: false,
      scrollThreshold: 200, // 滚动阈值，距离底部多少像素时触发加载
      observer: null, // IntersectionObserver实例
      form: { 
        name: '',
        video_file: null,
        face_image_file: null,
        background_file: null
      },
      rules: {
        name: [ 
            { required: true, message: '请输入模版名称', trigger: 'blur' },
            { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ], 
        video_file: [ 
            { required: true, message: '请上传原始视频', trigger: 'change' } 
        ] 
      },
      videoFileList: [], 
      faceFileList: [],
      bgFileList: [],
      isMobile: false,
      lastScrollTop: 0, // 记录上次滚动位置，用于判断滚动方向
      // 视口控制相关变量
      originalViewportContent: null,
      isInputFocused: false
    }
  },
  created() {
    this.checkDeviceType()
    window.addEventListener('resize', this.checkDeviceType)
    const savedViewMode = localStorage.getItem('template_view_mode')
    if (savedViewMode && !this.isMobile) {
      this.isCardView = savedViewMode === 'card'
    } else if (this.isMobile) {
      this.isCardView = true
    }
    this.fetchTemplates()
  },
  methods: {
    checkDeviceType() {
      const oldValue = this.isMobile
      this.isMobile = window.innerWidth <= 768 || /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
      if (oldValue !== this.isMobile && this.isMobile) {
        this.isCardView = true
        this.currentPage = 1
        this.fetchTemplates()
      }
    },
    fetchTemplates(loadMore = false) {
      if (this.loading || (loadMore && this.loadingMore)) return
      if (!loadMore) this.loading = true
      else this.loadingMore = true
      
      const pageSize = this.isCardView ? this.cardPageSize : this.pageSize
      
      this.$http.get('/api/digital-human-templates', { params: { page: this.currentPage, size: pageSize } })
        .then(res => {
          let newTemplates = res.data.templates || []
          let totalCount = res.data.total || 0
          
          if (loadMore) {
            this.templates = [...this.templates, ...newTemplates]
          } else {
            this.templates = newTemplates
          }
          
          this.total = totalCount
          
          // 更准确地计算是否还有更多数据
          // 如果返回的数据少于请求的页大小，或者已经加载的总数等于总计数，就没有更多数据了
          this.hasMoreData = newTemplates.length === pageSize && this.templates.length < totalCount
          
          // 如果当前页没有数据但总数不为0，尝试加载前一页
          if (newTemplates.length === 0 && totalCount > 0 && this.currentPage > 1 && !loadMore) {
            this.currentPage = this.currentPage - 1
            this.fetchTemplates()
            return
          }
        })
        .catch(() => {
          this.hasMoreData = false
        })
        .finally(() => {
          this.loading = false
          this.loadingMore = false
          
          if (this.isCardView) {
            this.$nextTick(() => {
              this.initWaterfallLayout()
              // 只有当确实有更多数据时才设置观察器
              if (this.hasMoreData) {
                this.setupIntersectionObserver()
              } else {
                // 如果没有更多数据，断开观察器连接
                if (this.observer) {
                  this.observer.disconnect()
                  this.observer = null
                }
              }
            })
          }
        })
    },
    loadMoreTemplates() {
      // 增加额外的安全检查
      if (this.loadingMore || !this.hasMoreData || this.dialogVisible) return
      
      // 检查是否已加载所有数据
      if (this.templates.length >= this.total) {
        this.hasMoreData = false
        return
      }
      
      this.currentPage++
      this.fetchTemplates(true)
    },
    initWaterfallLayout() {
      const cardContainer = this.$refs.cardContainer
      if (!cardContainer) return
      this.$forceUpdate()
      setTimeout(() => { window.dispatchEvent(new Event('resize')) }, 100)
    },
    setupIntersectionObserver() {
      if (this.observer) { this.observer.disconnect(); this.observer = null }
      this.$nextTick(() => {
        const triggerElement = this.$refs.loadMoreTrigger
        if (!triggerElement) return
        this.observer = new IntersectionObserver((entries) => {
          const entry = entries[0]
          if (entry.isIntersecting && !this.loadingMore && this.hasMoreData && !this.dialogVisible) {
            this.loadMoreTemplates()
          }
        }, { root: null, threshold: 0, rootMargin: '200px' })
        this.observer.observe(triggerElement)
      })
    },
    toggleView() {
      this.isCardView = !this.isCardView
    //   localStorage.setItem('template_view_mode', this.isCardView ? 'card' : 'list')
      this.$nextTick(() => { window.scrollTo({ top: 0, behavior: 'smooth' }) })
      this.currentPage = 1
      this.templates = []
      this.hasMoreData = true
      this.fetchTemplates()
      if (this.isCardView) { this.$nextTick(() => { this.initWaterfallLayout(); this.setupIntersectionObserver() }) }
    },
    handleVideoChange(file) {
      this.form.video_file = file.raw
      this.videoFileList = [file]
    },
    handleFaceChange(file) {
      this.form.face_image_file = file.raw
      this.faceFileList = [file]
    },
    handleBgChange(file) {
      this.form.background_file = file.raw
      this.bgFileList = [file]
    },
    submitForm() {
      this.$refs.form.validate(valid => {
        if (!valid) return
        this.submitting = true
        const fd = new FormData()
        fd.append('name', this.form.name)
        if (this.form.video_file) fd.append('video_file', this.form.video_file)
        if (this.form.face_image_file) fd.append('face_image_file', this.form.face_image_file)
        if (this.form.background_file) fd.append('background_file', this.form.background_file)
        this.$http.post('/api/digital-human-template', fd, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
          .then(() => {
            this.$message.success('创建成功')
            this.dialogVisible = false
            this.fetchTemplates()
            this.resetForm()
          })
          .finally(() => { this.submitting = false })
      })
    },
    resetForm() {
      this.form = { name: '', video_file: null, face_image_file: null, background_file: null }
      this.videoFileList = []
      this.faceFileList = []
      this.bgFileList = []
      if (this.$refs.form) this.$refs.form.resetFields()
    },
    confirmDelete(id) {
      this.$confirm('确定要删除此模版吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.deleteTemplate(id)
      })
    },
    deleteTemplate(id) {
      this.$http.delete(`/api/digital-human-template/${id}`)
        .then(() => {
          this.$message.success('删除成功')
          // 重置分页状态和数据
          this.currentPage = 1
          this.templates = []
          this.hasMoreData = true
          // 重新获取第一页数据
          this.fetchTemplates()
        })
    },
    viewDetail(row) { this.$router.push(`/digital-human-template/${row.id}`) },
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchTemplates()
    },
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchTemplates()
    },
    showCreateDialog() {
      this.dialogVisible = true
      this.resetForm()
    },
    formatDate(dateString) { if (!dateString) return ''; return new Date(dateString).toLocaleString() },
    getStatusType(status) {
      const statusMap = { 
        completed: 'success', 
        failed: 'danger', 
        processing: 'warning',
        active: 'success', 
        disabled: 'info', 
        deleted: 'danger' 
      }
      return statusMap[status] || 'info'
    },
    getStatusText(status) {
      const statusMap = { 
        completed: '已完成', 
        failed: '失败', 
        processing: '处理中',
        active: '正常', 
        disabled: '禁用', 
        deleted: '已删除' 
      }
      return statusMap[status] || status
    },
  },
  mounted() {
    this.isCardView = true
    if (this.isCardView) { this.$nextTick(() => { this.initWaterfallLayout(); this.setupIntersectionObserver() }) }
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkDeviceType)
    if (this.observer) { this.observer.disconnect(); this.observer = null }
  }
}
</script>

<style scoped>
.digital-human-container {
  padding: 20px;
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #fff;
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

/* 移动端头部 */
.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 999;
  border-radius: 0;
  padding: 10px 12px;
  margin: 0;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  flex-direction: row;
  align-items: center;
  height: 50px;
  box-sizing: border-box;
}

.mobile-header-placeholder {
  height: 52px;
  margin: 0;
  padding: 0;
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

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

/* 强制移动端使用卡片视图 */
.mobile-card-view {
  display: none;
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

/* 移动端底部按钮样式 */
.mobile-form-footer {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 2002;
  background-color: #fff;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
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

/* 响应式样式 */
@media screen and (max-width: 768px) {
  .digital-human-container {
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
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 999;
    border-radius: 0;
    padding: 10px 12px;
    margin: 0;
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    flex-direction: row;
    align-items: center;
    height: 50px;
    box-sizing: border-box;
  }
  
  .header-left {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
  
  .card-list {
    margin-top: 0;
    padding-top: 0;
  }
  
  .card-view-content {
    padding: 0;
    margin: 0;
  }
  
  .waterfall-container {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    padding: 8px;
    margin: 0;
    padding-top: 0;
  }
  
  .mobile-card-container {
    margin-top: 0 !important;
    padding-top: 0 !important;
  }
  
  /* 悬浮按钮移动端样式 */
  .floating-add-btn {
    bottom: 80px;
    right: 16px;
    width: 56px;
    height: 56px;
    background: linear-gradient(135deg, #3f51b5, #2196f3);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
    z-index: 1001;
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
  
  /* 移动端对话框样式优化 */
  .el-dialog.digital-human-dialog {
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
  .digital-human-template-form {
    padding: 10px 15px 70px !important; /* 为底部按钮留出空间 */
  }
  
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
  
  .card-list .el-empty {
    margin: 0 !important;
    padding: 10px 0 !important;
  }
  
  .card-view-content {
    margin: 0 !important;
    padding: 0 !important;
  }
  
  .waterfall-container {
    margin: 0 !important;
  }
  
  /* 加载中动画优化 */
  .el-loading-spinner {
    top: 35% !important;
  }
  
  /* 消除列表显示时的底部空白 */
  .el-table {
    margin-bottom: 60px !important;
  }
}
</style>