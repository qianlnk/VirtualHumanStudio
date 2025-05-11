<template>
  <div class="digital-human-container">
    <div class="page-header" :class="{'mobile-header': isMobile}">
      <div class="header-left">
        <h2>数字人合成</h2>
      </div>
      <div class="header-right">
        <el-button v-if="!isMobile" type="primary" @click="showCreateDialog" icon="el-icon-plus">创建任务</el-button>
        <el-button v-if="!isMobile" type="text" size="small" class="view-toggle" @click="toggleView">
          <i :class="isCardView ? 'el-icon-menu' : 'el-icon-s-grid'"></i>
          <span class="toggle-text">{{ isCardView ? '列表视图' : '卡片视图' }}</span>
        </el-button>
      </div>
    </div>
    
    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>
    
    <!-- 任务列表 -->
    <div v-loading="loading" class="task-list mobile-card-view" v-show="!isCardView">
      <el-empty v-if="digitalHumans.length === 0" description="暂无数字人合成任务"></el-empty>
      
      <el-table v-else :data="digitalHumans" style="width: 100%" class="responsive-table" ref="dataTable">
        <el-table-column prop="name" label="任务名称" min-width="300" show-overflow-tooltip>
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
                @click="viewDetail(scope.row.id)"
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
      <el-empty v-if="digitalHumans.length === 0" description="暂无数字人合成任务"></el-empty>
      
      <div v-else class="card-view-content" ref="cardViewContent">
        <div class="waterfall-container" ref="cardContainer" :class="{'mobile-card-container': isMobile}">
          <div class="task-card" v-for="(item, index) in digitalHumans" :key="item.id || index" :ref="`taskCard_${item.id || index}`">
            <div class="task-card-header">
              <h3 class="task-card-title">{{ item.name || '未命名任务' }}</h3>
              <el-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
            </div>
            <div class="task-card-content">
              <div class="task-card-info">
                <p><i class="el-icon-time"></i> {{ formatDate(item.created_at) }}</p>
              </div>
            </div>
            <div class="task-card-footer">
              <el-button type="text" size="small" class="action-btn" @click="viewDetail(item.id)">查看</el-button>
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
    <el-dialog 
      title="创建数字人合成任务" 
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
        class="digital-human-form">
        
        <el-form-item label="任务名称" prop="name">
          <el-input 
            v-model="form.name" 
            placeholder="请输入任务名称"
            @focus="handleInputFocus"
            @blur="handleInputBlur"></el-input>
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
        
        <!-- 移动端底部按钮 -->
        <div v-if="isMobile" class="mobile-form-footer">
          <el-button type="primary" @click="submitForm" :loading="submitting" class="mobile-submit-btn">创建任务</el-button>
        </div>
      </el-form>
      
      <!-- 桌面端底部按钮 -->
      <div v-if="!isMobile" slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">创建</el-button>
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
      videoFileList: [],
      isMobile: false, // 移动设备检测标志
      lastScrollTop: 0, // 记录上次滚动位置，用于判断滚动方向
      // 视口控制相关变量
      originalViewportContent: null,
      isInputFocused: false
    }
  },
  created() {
    // 检测设备类型
    this.checkDeviceType()
    // 监听窗口大小变化
    window.addEventListener('resize', this.checkDeviceType)
    
    // 从本地存储中读取用户偏好的视图模式
    const savedViewMode = localStorage.getItem('digital_human_view_mode')
    if (savedViewMode && !this.isMobile) {
      this.isCardView = savedViewMode === 'card'
    } else if (this.isMobile) {
      // 移动端强制使用卡片视图
      this.isCardView = true
      console.log('移动端：强制使用卡片视图')
    }
    
    // 初始加载数据
    this.fetchDigitalHumans()
  },
  methods: {
    // 检查设备类型
    checkDeviceType() {
      const oldValue = this.isMobile
      // 使用更可靠的检测方法
      this.isMobile = window.innerWidth <= 768 || /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
      
      console.log('设备检测:', '是否移动端=', this.isMobile, '屏幕宽度=', window.innerWidth)
      
      // 只有当状态改变时才强制切换视图
      if (oldValue !== this.isMobile && this.isMobile) {
        console.log('切换到移动端，强制使用卡片视图')
        this.isCardView = true
        // 重新加载数据以适应新视图
        this.currentPage = 1
        this.fetchDigitalHumans()
      }
    },
    
    // 获取数字人任务列表
    async fetchDigitalHumans(loadMore = false) {
      if (this.loading || (loadMore && this.loadingMore)) {
        console.log('已有请求进行中，跳过')
        return
      }
      
      if (!loadMore) {
        this.loading = true
      } else {
        this.loadingMore = true
      }
      
      console.log('===== 开始请求数字人数据 =====')
      console.log('请求数据:', '页码=', this.currentPage, '每页数量=', this.isCardView ? this.cardPageSize : this.pageSize)
      
      try {
        // 尝试不同的API端点，找到一个可用的
        const endpoints = [
          '/api/digital-human',
          '/api/digitalhuman',
          '/api/digital_human',
          '/api/digital-humans'
        ]
        
        let response = null
        let successEndpoint = null
        
        // 依次尝试每个端点
        for (const endpoint of endpoints) {
          try {
            console.log('尝试API请求URL:', endpoint)
            const startTime = Date.now()
            const result = await this.$http.get(endpoint, {
              params: {
                page: this.currentPage,
                size: this.isCardView ? this.cardPageSize : this.pageSize
              }
            })
            console.log('请求成功，耗时:', Date.now() - startTime, 'ms')
            
            // 如果请求成功，保存响应并记录成功的端点
            response = result
            successEndpoint = endpoint
            // 成功后跳出循环
            break
          } catch (err) {
            console.log(`端点 ${endpoint} 请求失败:`, err.message)
            // 继续尝试下一个端点
          }
        }
        
        // 如果所有端点都失败，抛出错误
        if (!response) {
          throw new Error('所有API端点都请求失败')
        }
        
        console.log('成功使用的端点:', successEndpoint)
        console.log('HTTP响应状态码:', response.status)
        console.log('完整原始响应:', response)
        console.log('原始响应数据:', response.data)
        
        // 检查不同的响应结构
        let newTasks = []
        let totalCount = 0
        
        // 测试在渲染前强制将数据转换为数组
        if (response.data && typeof response.data === 'object' && !Array.isArray(response.data)) {
          if (response.data.items && Array.isArray(response.data.items)) {
            // 标准结构: {items: [...], total: 100}
            console.log('检测到标准响应结构: {items: [...], total: 100}')
            newTasks = response.data.items
            totalCount = response.data.total || 0
          } else if (response.data.data && Array.isArray(response.data.data)) {
            // {data: [...], total: 100} 结构
            console.log('检测到data字段响应结构: {data: [...], total: 100}')
            newTasks = response.data.data
            totalCount = response.data.total || 0
          } else if (response.data.list && Array.isArray(response.data.list)) {
            // {list: [...], total: 100} 结构
            console.log('检测到list字段响应结构: {list: [...], total: 100}')
            newTasks = response.data.list
            totalCount = response.data.total || response.data.totalCount || response.data.count || 0
          } else if (response.data.records && Array.isArray(response.data.records)) {
            // {records: [...], total: 100} 结构 (MyBatis-Plus分页)
            console.log('检测到MyBatis-Plus分页结构: {records: [...], total: 100}')
            newTasks = response.data.records
            totalCount = response.data.total || 0
          } else if (response.data.content && Array.isArray(response.data.content)) {
            // Spring Data分页结构
            console.log('检测到Spring Data分页结构')
            newTasks = response.data.content
            totalCount = response.data.totalElements || 0
          } else {
            // 如果是对象但没有识别出结构，尝试强制转换
            console.log('尝试从对象转换:', response.data)
            // 将对象的所有值合并为数组
            const objectValues = Object.values(response.data).filter(val => val && typeof val === 'object')
            if (objectValues.length > 0) {
              // 取第一个有效对象
              const firstValidValue = objectValues[0]
              if (Array.isArray(firstValidValue)) {
                console.log('从对象中提取数组:', firstValidValue)
                newTasks = firstValidValue
                totalCount = firstValidValue.length
              }
            }
          }
        } else if (Array.isArray(response.data)) {
          // 数组结构: [{...}, {...}]
          console.log('检测到数组响应结构: [{...}, {...}]')
          newTasks = response.data
          totalCount = response.headers['x-total-count'] || newTasks.length
        } else {
          console.error('无法识别的响应格式:', response.data)
          console.log('响应数据类型:', typeof response.data)
          if (typeof response.data === 'object') {
            console.log('响应数据键:', Object.keys(response.data))
          }
          
          // 尝试添加一些测试数据以便进行可视化调试
          console.log('添加模拟数据用于测试视图渲染')
          newTasks = [
            { id: 1, name: '测试任务1', status: 'completed', created_at: new Date().toISOString() },
            { id: 2, name: '测试任务2', status: 'processing', created_at: new Date().toISOString() }
          ]
          totalCount = 2
        }
        
        console.log('解析后数据数量:', newTasks.length, '总数:', totalCount)
        if (newTasks.length > 0) {
          console.log('第一条数据样例:', newTasks[0])
        }
        
        if (loadMore) {
          // 追加新数据
          this.digitalHumans = [...this.digitalHumans, ...newTasks]
        } else {
          // 重置数据
          this.digitalHumans = newTasks
        }
        
        this.total = totalCount
        
        // 判断是否还有更多数据
        const hasMore = this.digitalHumans.length < totalCount
        this.hasMoreData = hasMore
        
        console.log('更新后数据量：', this.digitalHumans.length, '总数：', this.total, '是否还有更多：', this.hasMoreData)
        console.log('===== 数据请求完成 =====')
        
        // 强制DOM更新
        this.$forceUpdate()
        
      } catch (error) {
        console.error('获取数字人任务列表失败:', error)
        console.error('错误详情:', error.response || error.message || error)
        this.$message.error('获取任务列表失败: ' + (error.message || '未知错误'))
        
        // 添加一些测试数据以便进行可视化调试
        console.log('添加模拟数据用于测试视图渲染')
        const testData = [
          { id: 1, name: '测试任务1', status: 'completed', created_at: new Date().toISOString() },
          { id: 2, name: '测试任务2', status: 'processing', created_at: new Date().toISOString() }
        ]
        this.digitalHumans = testData
        this.total = 2
        this.hasMoreData = false
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
        console.log('跳过加载更多:', '加载中=', this.loadingMore, '没有更多数据=', !this.hasMoreData)
        return
      }
      
      console.log('开始加载更多数据，当前页码：', this.currentPage)
      this.currentPage++
      this.fetchDigitalHumans(true)
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
          console.log('没有找到加载更多触发元素')
          return
        }
        
        console.log('设置IntersectionObserver')
        
        // 创建新的IntersectionObserver
        this.observer = new IntersectionObserver((entries) => {
          const entry = entries[0]
          console.log('触发IntersectionObserver:', 
            '可见=', entry.isIntersecting, 
            '加载中=', this.loadingMore, 
            '有更多数据=', this.hasMoreData,
            '当前页=', this.currentPage,
            '已有数据=', this.digitalHumans.length)
            
          // 只有在元素可见、不在加载中、有更多数据且当前有数据时才触发加载
          if (entry.isIntersecting && !this.loadingMore && this.hasMoreData && this.digitalHumans.length > 0) {
            console.log('IntersectionObserver触发加载更多')
            this.loadMoreTasks()
          }
        }, {
          root: null,
          threshold: 0.5, // 更改阈值，至少看到50%时才触发
          rootMargin: '100px' // 增加根元素边距
        })
        
        // 开始观察
        this.observer.observe(triggerElement)
      })
    },
    
    // 处理窗口滚动事件
    handleWindowScroll() {
      // 记录滚动方向
      const currentScrollTop = window.pageYOffset || document.documentElement.scrollTop
      const scrollingDown = currentScrollTop > this.lastScrollTop
      this.lastScrollTop = currentScrollTop
      
      // 只有向下滚动时才检查是否需要加载更多
      if (!scrollingDown) {
        return
      }
      
      // 加载更多数据条件
      if (this.isCardView && !this.loadingMore && this.hasMoreData && this.digitalHumans.length > 0) {
        const scrollTop = window.pageYOffset || document.documentElement.scrollTop
        const windowHeight = window.innerHeight
        const documentHeight = Math.max(
          document.body.scrollHeight, document.documentElement.scrollHeight,
          document.body.offsetHeight, document.documentElement.offsetHeight,
          document.body.clientHeight, document.documentElement.clientHeight
        )
        
        // 当滚动到距离底部阈值距离时触发加载
        if (documentHeight - scrollTop - windowHeight < this.scrollThreshold) {
          console.log('滚动触发加载更多, 距离底部:', documentHeight - scrollTop - windowHeight)
          this.loadMoreTasks()
        }
      }
    },
    
    // 切换视图
    toggleView() {
      this.isCardView = !this.isCardView
      
      // 保存视图模式到本地存储
      localStorage.setItem('digital_human_view_mode', this.isCardView ? 'card' : 'list')
      
      // 切换视图后滚动到顶部
      this.$nextTick(() => {
        window.scrollTo({
          top: 0,
          behavior: 'smooth'
        })
      })
      
      // 重置状态
      this.currentPage = 1
      this.digitalHumans = []
      this.hasMoreData = true
      
      // 重新加载第一页数据
      this.fetchDigitalHumans()
      
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
    
    // 重置表单
    resetForm() {
      this.form = {
        name: '',
        description: '',
        task_code: '',
        chaofen: 0,
        watermark_switch: 0,
        pn: 1,
        audio_file: null,
        video_file: null
      }
      this.audioFileList = []
      this.videoFileList = []
      
      // 如果表单引用存在，重置验证
      if (this.$refs.form) {
        this.$refs.form.resetFields()
      }
    },
    
    // 处理音频文件改变
    handleAudioChange(file) {
      if (file && file.raw) {
        this.form.audio_file = file.raw
        this.audioFileList = [{ name: file.name, url: URL.createObjectURL(file.raw) }]
      }
    },
    
    // 处理视频文件改变
    handleVideoChange(file) {
      if (file && file.raw) {
        this.form.video_file = file.raw
        this.videoFileList = [{ name: file.name, url: URL.createObjectURL(file.raw) }]
      }
    },
    
    // 查看详情
    viewDetail(id) {
      this.$router.push(`/digital-human/${id}`)
    },
    
    // 确认删除
    confirmDelete(id) {
      this.$confirm('确定要删除此任务吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.deleteTask(id)
      }).catch(() => {
        // 取消删除
      })
    },
    
    // 删除任务
    async deleteTask(id) {
      try {
        await this.$http.delete(`/api/digital-human/${id}`)
        this.$message.success('删除成功')
        this.fetchDigitalHumans()
      } catch (error) {
        console.error('删除任务失败:', error)
        this.$message.error('删除任务失败: ' + (error.message || '未知错误'))
      }
    },
    
    // 分页大小改变
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchDigitalHumans()
    },
    
    // 当前页改变
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchDigitalHumans()
    },
    
    // 显示创建对话框
    showCreateDialog() {
      this.dialogVisible = true
      this.resetForm()
      
      // 对话框打开时设置移动端视口
      if (this.isMobile) {
        this.$nextTick(() => {
          this.setupMobileViewport();
        });
      }
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
      this.$http.post('/api/digital-human', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
        .then(() => {
          this.$message.success('创建数字人任务成功')
          this.dialogVisible = false
          this.fetchDigitalHumans()
          
          // 重置视口设置
          if (this.isMobile) {
            this.$nextTick(() => {
              this.resetMobileViewport();
            });
          }
        })
        .catch(error => {
          console.error('创建数字人任务失败:', error)
          this.$message.error('创建数字人任务失败: ' + ((error.response && error.response.data && error.response.data.error) || error.message))
        })
        .finally(() => {
          this.submitting = false
        })
    },
  },
  watch: {
    digitalHumans: {
      handler(newVal) {
        console.log('===== digitalHumans数组已更新 =====')
        console.log('数组长度:', newVal.length)
        console.log('数组内容:', newVal)
        console.log('DOM是否更新:', this.$el.querySelectorAll('.task-card').length)
      },
      deep: true
    }
  },
  mounted() {
    console.log('===== DigitalHuman组件已挂载 =====')
    console.log('初始数据状态:', this.digitalHumans)
    
    // 添加滚动事件监听器
    window.addEventListener('scroll', this.handleWindowScroll)
    
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
    
    // 移除窗口大小变化监听
    window.removeEventListener('resize', this.checkDeviceType);
    
    // 重置视口设置
    this.resetMobileViewport();
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
  .digital-human-form {
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