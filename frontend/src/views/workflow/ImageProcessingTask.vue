<template>
  <div class="workflow-container">
    <div class="page-header" :class="{'mobile-header': isMobile}">
      <div class="header-left">
        <h2>{{ currentModule ? currentModule.name : '图像处理' }}</h2>
      </div>
      <div class="header-right">
        <el-button v-if="!isMobile" type="primary" @click="showCreateDialog" :disabled="!currentModule" icon="el-icon-plus">创建{{currentModule ? currentModule.name : ''}}任务</el-button>
        <el-button v-if="!isMobile" type="text" size="small" class="view-toggle" @click="toggleView">
          <i :class="isCardView ? 'el-icon-menu' : 'el-icon-s-grid'"></i>
          <span class="toggle-text">{{ isCardView ? '列表视图' : '卡片视图' }}</span>
        </el-button>
      </div>
    </div>

    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>

    <!-- 任务列表 -->
    <el-card class="task-list" v-show="!isCardView">
      <!-- 桌面端表格视图 -->
      <el-table
        v-loading="loading"
        :data="tasks"
        style="width: 100%"
        :empty-text="'暂无数据'"
      >
        <el-table-column prop="name" label="任务名称" min-width="250" show-overflow-tooltip></el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ getStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template slot-scope="scope">
            <div class="action-buttons">
              <el-button
                type="text"
                size="small"
                class="action-btn"
                @click="viewDetail(scope.row.id)">查看</el-button>
              <el-button
                type="text"
                size="small"
                class="action-btn"
                @click="deleteTask(scope.row)">删除</el-button>
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
    </el-card>
    
    <!-- 卡片视图（瀑布流） -->
    <div v-show="isCardView" class="card-list" v-loading="loading">
      <el-empty v-if="tasks.length === 0" description="暂无处理任务"></el-empty>
      
      <div v-else class="card-view-content">
        <div class="waterfall-container" ref="cardContainer" :class="{'mobile-card-container': isMobile}">
          <div class="task-card" v-for="item in tasks" :key="item.id">
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
              <el-button type="text" size="small" class="action-btn" @click="deleteTask(item)">删除</el-button>
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

    <!-- 结果预览对话框 -->
    <el-dialog title="处理结果" :visible.sync="previewDialogVisible" width="50%">
      <div class="preview-container" v-if="currentResult">
        <img :src="currentResult" class="preview-image" alt="处理结果">
      </div>
    </el-dialog>

    <!-- 蒙版编辑对话框 -->
    <el-dialog title="蒙版编辑" :visible.sync="maskEditorVisible" width="80%" :before-close="closeMaskEditor">
      <div class="mask-editor-container">
        <div class="editor-tools">
          <div class="tool-group">
            <span class="tool-label">模式:</span>
            <el-radio-group v-model="brushMode" size="small">
              <el-radio-button label="brush">画笔</el-radio-button>
              <el-radio-button label="eraser">橡皮擦</el-radio-button>
            </el-radio-group>
          </div>
          
          <div class="tool-group">
            <span class="tool-label">笔刷形状:</span>
            <el-radio-group v-model="brushShape" size="small">
              <el-radio-button label="round">圆形</el-radio-button>
              <el-radio-button label="square">方形</el-radio-button>
            </el-radio-group>
          </div>
          
          <div class="tool-group" v-if="brushMode === 'brush'">
            <span class="tool-label">颜色:</span>
            <el-radio-group v-model="brushColor" size="small">
              <el-radio-button label="black">黑色</el-radio-button>
              <el-radio-button label="white">白色</el-radio-button>
              <el-radio-button label="gray">灰色</el-radio-button>
            </el-radio-group>
          </div>
          
          <div class="tool-group">
            <span class="tool-label">大小:</span>
            <el-slider
              v-model="brushSize"
              :min="1"
              :max="50"
              :step="1"
              show-stops
              :show-tooltip="true"
              style="width: 200px;"
            ></el-slider>
            <span class="brush-size-label">{{brushSize}}px</span>
          </div>
          
          <el-button size="small" type="danger" @click="clearCanvas">清空</el-button>
        </div>
        
        <div class="canvas-container">
          <div class="canvas-wrapper">
            <canvas ref="baseCanvas" class="editor-canvas base-canvas"></canvas>
            <canvas ref="drawCanvas" class="editor-canvas draw-canvas" :style="{ cursor: brushMode === 'brush' ? 'crosshair' : 'cell' }"></canvas>
          </div>
        </div>
        
        <div class="editor-actions">
          <el-button type="primary" @click="saveMask">保存蒙版</el-button>
          <el-button @click="closeMaskEditor">取消</el-button>
        </div>
      </div>
    </el-dialog>
    
    <!-- 创建任务对话框 -->
    <el-dialog :title="`创建${currentModule ? currentModule.name : '图像处理'}任务`" :visible.sync="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="form" label-width="120px" v-if="currentModule && currentModule.inputParams">
        <!-- 任务名称输入框 -->
        <el-form-item label="任务名称" prop="taskName">
          <el-input
            type="textarea"
            v-model="form.taskName"
            placeholder="请输入任务名称"
            resize="both"
            :rows="2"
            style="width: 100%;"
          ></el-input>
        </el-form-item>
        <!-- 动态表单，根据模块配置生成不同的表单项 -->
        <div v-for="(param, index) in currentModule.inputParams" :key="index">
          <el-form-item :label="param.alias" :prop="`params.${param.key}`">
            <!-- 文本类型参数 -->
            <el-input 
              v-if="param.type === 'text'" 
              type="textarea"
              v-model="form.params[param.key]" 
              :placeholder="`请输入${param.alias}`"
              resize="both"
              :rows="3"
              style="width: 100%;"
            ></el-input>
            
            <!-- 选择类型参数 -->
            <el-select
              v-else-if="param.type === 'select'"
              v-model="form.params[param.key]"
              :placeholder="`请选择${param.alias}`">
              <el-option
                v-for="option in param.options"
                :key="option.value"
                :label="option.label"
                :value="option.value">
              </el-option>
            </el-select>
            
            <!-- 图片类型参数 -->
            <template v-else-if="param.type === 'image' || param.type === 'mask'">
              <div class="image-param-container">
                <el-upload
                  class="upload-item"
                  action="#"
                  :auto-upload="false"
                  :on-change="(file) => handleImageChange(file, param.key)"
                  :show-file-list="true"
                  accept="image/*"
                  :limit="1">
                  <el-button size="small" type="primary">选择图片</el-button>
                  <div slot="tip" class="el-upload__tip">{{ param.description }}</div>
                </el-upload>
                <div class="image-preview" v-if="previewUrls[param.key]">
                  <img :src="previewUrls[param.key]" class="preview-thumbnail" :alt="`${param.alias}预览`">
                </div>
                <el-button 
                  v-if="param.type === 'mask'" 
                  size="small" 
                  type="success" 
                  @click="startMaskEditing(param)"
                  :disabled="!form.params[param.key.replace('Mask', '')]">编辑蒙版</el-button>
              </div>
            </template>
          </el-form-item>
        </div>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">创建</el-button>
      </div>
    </el-dialog>

    <!-- 移动端悬浮添加按钮 -->
    <div v-if="isMobile" class="floating-add-btn" @click="showCreateDialog" :disabled="!currentModule">
      <i class="el-icon-plus"></i>
    </div>
  </div>
</template>

<script>
import { getImageProcessingModules, createImageProcessingTask, deleteImageProcessingTask, getImageProcessingTaskDetail, getImageProcessingTasks } from '@/utils/imageProcessingApi'
import '@/assets/styles/card-view.css'

export default {
  name: 'ImageProcessingTask',
  props: {
    moduleId: {
      type: String,
      default: ''
    }
  },
  data() {
    return {
      loading: false,
      submitting: false,
      isMobile: false,
      tasks: [],
      dialogVisible: false,
      previewDialogVisible: false,
      currentResult: null,
      currentModule: null,
      form: {
        taskName: '',
        params: {}
      },
      rules: {
        taskName: [{ required: true, message: '请输入任务名称', trigger: 'blur' }]
      },
      previewUrls: {},
      moduleLoading: true,
      modules: [],
      showModuleList: false,
      // 蒙版编辑器相关数据
      maskEditorVisible: false,
      currentMaskParam: null,
      brushMode: 'brush',
      brushShape: 'round',
      brushColor: 'black',
      brushSize: 10,
      baseImage: null,
      maskCanvas: null,
      isDrawing: false,
      lastX: 0,
      lastY: 0,
      // 视图和分页相关
      isCardView: false,
      total: 0,
      currentPage: 1,
      pageSize: 10,
      cardPageSize: 10,
      loadingMore: false,
      hasMoreData: true,
      initialLoaded: false,
      scrollThreshold: 200,
      lastScrollTop: 0 // 记录上次滚动位置
    }
  },
  created() {
    console.log('组件创建，当前路由参数:', this.$route.params)
    // 从本地存储中读取用户偏好的视图模式
    const savedViewMode = localStorage.getItem('image_processing_view_mode')
    if (savedViewMode) {
      this.isCardView = savedViewMode === 'card'
    }
    
    // 检测设备类型
    this.checkDeviceType()
  },
  watch: {
    '$route.params.moduleId': {
      immediate: true,
      deep: true,
      async handler(newVal) {
          console.log('监听到模块ID变化:', newVal)
          // 重置状态
          this.tasks = []
          this.loading = false // 先重置loading状态，避免影响fetchTasks的执行
          this.currentModule = null
          this.initialLoaded = false // 重置初始加载标记
          this.form = {
            taskName: '',
            params: {}
          }
          this.previewUrls = {}
          
          // 重置分页参数
          this.currentPage = 1
          this.total = 0
          this.hasMoreData = true
          
          try {
            const moduleInitialized = await this.initModule()
            if (moduleInitialized && this.currentModule) {
              console.log('模块初始化成功，开始加载任务列表')
              // 确保这里直接调用fetchTasks，不依赖其他地方的loading状态
              this.loading = false // 再次确保loading为false
              await this.fetchTasks()
              this.initialLoaded = true // 只有在成功获取数据后才标记为已加载
            }
          } catch (error) {
            console.error('模块初始化失败:', error)
            this.$message.error('加载模块失败：' + error.message)
          } finally {
            this.loading = false
          }
      }
    }
  },
  mounted() {
    console.log('组件挂载')
    
    // 监听窗口大小变化，更新设备类型
    window.addEventListener('resize', this.checkDeviceType)
    
    // 添加滚动事件监听器
    window.addEventListener('scroll', this.handleWindowScroll)
    
    // 组件挂载后，如果模块已初始化但数据未加载，主动加载数据
    if (this.currentModule && !this.initialLoaded) {
      console.log('模块已初始化，自动加载数据')
      this.loadInitialData()
    }
    
    // 如果是移动端，默认使用卡片视图
    if (this.isMobile) {
      this.isCardView = true
    }
  },
  
  updated() {
    // 避免频繁检查，使用防抖处理
    clearTimeout(this._dataLoadTimer)
    this._dataLoadTimer = setTimeout(() => {
      this.ensureDataLoaded()
    }, 100)
  },
  
  destroyed() {
    // 移除事件监听
    window.removeEventListener('resize', this.checkDeviceType)
    window.removeEventListener('scroll', this.handleWindowScroll)
    window.onscroll = null
    
    // 清除定时器
    if (this.loadingTimer) {
      clearTimeout(this.loadingTimer)
    }
  },
  methods: {
    // 检测设备类型
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
    // 获取所有模块
    async fetchModules() {
      this.moduleLoading = true
      try {
        const response = await getImageProcessingModules()
        if (response.success && response.modules) {
          this.modules = response.modules
        }
      } catch (error) {
        this.$message.error('获取模块列表失败：' + error.message)
      } finally {
        this.moduleLoading = false
      }
    },

    // 初始化模块
    async initModule() {
      this.moduleLoading = true
      try {
        if (!this.modules.length) {
          console.log('模块列表为空，开始获取模块列表')
          await this.fetchModules()
          console.log('获取到模块列表:', this.modules.length, '个模块')
        }
        
        // 尝试从路由参数或组件属性获取模块ID
        const moduleId = this.moduleId || this.$route.params.moduleId
        console.log('尝试获取模块ID:', moduleId)
        
        if (moduleId) {
          // 如果有指定模块ID，查找对应模块
          this.currentModule = this.modules.find(m => m.id === moduleId)
          if (this.currentModule) {
            console.log('找到指定模块:', this.currentModule.id, this.currentModule.name)
            // 初始化表单验证规则
            this.initFormRules()
            this.showModuleList = false
            return true
          } else {
            console.error('未找到指定的处理模块:', moduleId)
            this.$message.error('未找到指定的处理模块，即将返回模块列表页面')
            // 重定向到模块列表页面
            this.$router.push('/home')
            return false
          }
        } else if (this.modules.length > 0) {
          // 如果没有指定模块ID但有可用模块，使用第一个模块
          console.log('没有指定模块ID，使用第一个可用模块:', this.modules[0].id)
          this.currentModule = this.modules[0]
          this.initFormRules()
          this.showModuleList = false
          // 更新路由（可选）
          // this.$router.replace(`/image-processing/${this.currentModule.id}`)
          return true
        }
        
        if (!this.currentModule) {
          console.error('初始化失败：无法确定当前模块')
        }
        return false
      } catch (error) {
        console.error('初始化模块失败:', error)
        this.$message.error('初始化模块失败：' + error.message)
        return false
      } finally {
        this.moduleLoading = false
        console.log('模块初始化完成, 当前模块:', this.currentModule ? this.currentModule.id : '无')
      }
    },

    // 选择模块
    selectModule(module) {
      this.$router.push(`/image-processing/${module.id}`)
    },

    // 初始化表单验证规则
    initFormRules() {
      const rules = {}
      if (this.currentModule && this.currentModule.inputParams) {
        this.currentModule.inputParams.forEach(param => {
          if (param.required) {
            rules[`params.${param.key}`] = [{
              required: true,
              message: `请${param.type === 'select' ? '选择' : '输入'}${param.alias}`,
              trigger: param.type === 'select' ? 'change' : 'blur'
            }]
          }
        })
      }
      this.rules = rules
    },

    // 获取任务列表
    async fetchTasks(loadMore = false) {
      if (!this.currentModule) {
        console.log('没有当前模块，无法获取任务列表')
        return
      }
      
      // 记录当前的调用状态
      console.log('准备获取任务列表，当前状态:', 
        '加载中=', this.loading, 
        '加载更多中=', this.loadingMore, 
        '模块ID=', this.currentModule.id, 
        '页码=', this.currentPage,
        '总数=', this.total)
      
      // 防止无限请求：如果已知总数为0且不是首次加载，无需再请求
      if (this.initialLoaded && this.total === 0 && !loadMore) {
        console.log('已知总数为0，跳过请求')
        return
      }
      
      // 如果当前页已超过总页数，重置为第一页
      const totalPages = Math.ceil(this.total / (this.isCardView ? this.cardPageSize : this.pageSize)) || 1
      if (this.currentPage > totalPages && this.initialLoaded) {
        console.log('当前页码超出范围，重置为第1页', '当前页=', this.currentPage, '总页数=', totalPages)
        this.currentPage = 1
      }
      
      if (loadMore && this.loadingMore) {
        console.log('已经在加载更多任务，跳过此次请求')
        return
      } else if (!loadMore && this.loading) {
        console.log('已经在加载初始任务列表，跳过此次请求')
        return
      }
      
      // 设置加载状态
      if (loadMore) {
        this.loadingMore = true 
      } else {
        this.loading = true
      }
      
      try {
        console.log('发起API请求:', '页码=', this.currentPage, '每页数量=', this.isCardView ? this.cardPageSize : this.pageSize)
        const response = await getImageProcessingTasks(this.currentModule.id, {
          page: this.currentPage,
          size: this.isCardView ? this.cardPageSize : this.pageSize
        })
        
        if (response.success) {
          const newTasks = response.tasks || []
          this.total = response.total || 0
          
          console.log('获取到任务数据成功:', '新任务数量=', newTasks.length, '总数=', this.total)
          
          if (loadMore) {
            // 追加新数据
            this.tasks = [...this.tasks, ...newTasks]
          } else {
            // 重置数据
            this.tasks = newTasks
          }
          
          // 判断是否还有更多数据
          this.hasMoreData = this.tasks.length < this.total
          console.log('更新后的任务列表:', '当前列表长度=', this.tasks.length, '是否还有更多=', this.hasMoreData)
        } else {
          console.error('获取任务列表失败, 服务器返回错误:', response)
          this.$message.error('获取任务列表失败: ' + (response.message || '未知错误'))
        }
      } catch (error) {
        console.error('获取任务列表失败:', error)
        this.$message.error('获取任务列表失败：' + error.message)
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

    // 显示创建对话框
    showCreateDialog() {
      this.dialogVisible = true
      this.form.params = {}
      this.previewUrls = {}
      
      // 初始化默认值
      if (this.currentModule && this.currentModule.inputParams) {
        this.currentModule.inputParams.forEach(param => {
          if (param.default !== undefined) {
            this.$set(this.form.params, param.key, param.default)
          }
        })
      }
      
      if (this.$refs.form) {
        this.$refs.form.resetFields()
      }
    },

    // 处理图片上传
    handleImageChange(file, key) {
      if (!file) return
      
      // 创建预览URL并立即更新视图
      this.$set(this.previewUrls, key, URL.createObjectURL(file.raw))
      // 保存文件对象
      this.$set(this.form.params, key, file.raw)

      // 如果是原始图片，清除对应的蒙版
      const maskKey = key + 'Mask'
      if (this.previewUrls[maskKey]) {
        this.$delete(this.previewUrls, maskKey)
        this.$delete(this.form.params, maskKey)
      }
    },
    
    // 开始蒙版编辑
    async startMaskEditing(param) {
      // 验证参数
      if (!param || !param.key || !param.key.endsWith('Mask')) {
        this.$message.error('无效的蒙版参数')
        return
      }

      // 获取原始图片key（去掉Mask后缀）
      const baseImageKey = param.key.slice(0, -4) // 移除'Mask'后缀
      const baseImageFile = this.form.params[baseImageKey]
      
      // 验证原始图片是否存在
      if (!baseImageFile) {
        this.$message.error('请先上传原始图片')
        return
      }

      // 设置编辑器状态
      this.currentMaskParam = param
      this.maskEditorVisible = true

      // 初始化画布
      await this.$nextTick()
      const baseCanvas = this.$refs.baseCanvas
      const drawCanvas = this.$refs.drawCanvas
      if (!baseCanvas || !drawCanvas) {
        this.$message.error('画布初始化失败')
        this.maskEditorVisible = false
        return
      }

      // 设置画布样式
      baseCanvas.style.position = 'absolute'
      drawCanvas.style.position = 'absolute'
      baseCanvas.style.top = '0'
      drawCanvas.style.top = '0'
      baseCanvas.style.left = '0'
      drawCanvas.style.left = '0'
      baseCanvas.style.width = '100%'
      drawCanvas.style.width = '100%'
      baseCanvas.style.height = '100%'
      drawCanvas.style.height = '100%'
      baseCanvas.style.objectFit = 'contain'
      drawCanvas.style.objectFit = 'contain'

      // 加载原始图片
      const img = new Image()
      img.onload = () => {
        // 设置画布大小
        const maxWidth = 800
        const scale = Math.min(1, maxWidth / img.width)
        baseCanvas.width = img.width * scale
        baseCanvas.height = img.height * scale
        drawCanvas.width = baseCanvas.width
        drawCanvas.height = baseCanvas.height

        // 绘制原始图片
        const baseCtx = baseCanvas.getContext('2d')
        baseCtx.clearRect(0, 0, baseCanvas.width, baseCanvas.height)
        baseCtx.drawImage(img, 0, 0, baseCanvas.width, baseCanvas.height)

        // 初始化蒙版画布
        const drawCtx = drawCanvas.getContext('2d')
        drawCtx.clearRect(0, 0, drawCanvas.width, drawCanvas.height)

        // 如果已有蒙版图片，加载现有蒙版
        if (this.form.params[param.key]) {
          const maskImg = new Image()
          maskImg.onload = () => {
            // 清空绘图画布
            drawCtx.clearRect(0, 0, drawCanvas.width, drawCanvas.height)
            
            // 创建临时画布来解析蒙版图片
            const tempCanvas = document.createElement('canvas')
            tempCanvas.width = drawCanvas.width
            tempCanvas.height = drawCanvas.height
            const tempCtx = tempCanvas.getContext('2d')
            
            // 在临时画布上绘制蒙版图片
            tempCtx.drawImage(maskImg, 0, 0, tempCanvas.width, tempCanvas.height)
            
            // 从临时画布获取图像数据
            const imageData = tempCtx.getImageData(0, 0, tempCanvas.width, tempCanvas.height)
            const data = imageData.data
            
            // 创建新的画布数据，提取蒙版信息
            const newImageData = drawCtx.createImageData(drawCanvas.width, drawCanvas.height)
            const newData = newImageData.data
            
            // 处理图像数据，提取蒙版信息
            for (let i = 0; i < data.length; i += 4) {
              // 检查像素是否透明（蒙版区域）
              if (data[i + 3] < 255) { // 如果有透明度
                // 设置为黑色（或其他颜色）表示蒙版区域
                newData[i] = 0     // R
                newData[i + 1] = 0  // G
                newData[i + 2] = 0  // B
                newData[i + 3] = 255  // A (完全不透明)
              } else {
                // 设置为透明
                newData[i + 3] = 0  // A (完全透明)
              }
            }
            
            // 将处理后的图像数据绘制到绘图画布上
            drawCtx.putImageData(newImageData, 0, 0)
            
            // 设置绘图模式为半透明，以便用户可以看到底层图像
            drawCtx.globalAlpha = 0.5
            
            // 初始化绘制事件
            this.initDrawEvents()
          }
          maskImg.src = URL.createObjectURL(this.form.params[param.key])
        } else {
          // 如果没有蒙版，直接初始化绘制事件
          this.initDrawEvents()
        }
      }
      img.onerror = () => {
        this.$message.error('图片加载失败')
        this.maskEditorVisible = false
      }
      img.src = URL.createObjectURL(baseImageFile)
    },

    // 初始化绘制事件
    initDrawEvents() {
      const canvas = this.$refs.drawCanvas
      const ctx = canvas.getContext('2d')
      
      // 确保绘制前设置正确的透明度
      ctx.globalAlpha = 1.0

      // 获取缩放后的坐标
      const getScaledCoordinates = (e, canvas) => {
        const rect = canvas.getBoundingClientRect()
        const displayWidth = rect.width
        const displayHeight = rect.height
        const canvasWidth = canvas.width
        const canvasHeight = canvas.height
        
        // 计算画布的实际显示区域（考虑object-fit: contain的影响）
        const displayRatio = displayWidth / displayHeight
        const canvasRatio = canvasWidth / canvasHeight
        
        let effectiveWidth, effectiveHeight, offsetX, offsetY
        
        if (displayRatio > canvasRatio) {
          // 高度填满，宽度居中
          effectiveHeight = displayHeight
          effectiveWidth = effectiveHeight * canvasRatio
          offsetX = (displayWidth - effectiveWidth) / 2
          offsetY = 0
        } else {
          // 宽度填满，高度居中
          effectiveWidth = displayWidth
          effectiveHeight = effectiveWidth / canvasRatio
          offsetX = 0
          offsetY = (displayHeight - effectiveHeight) / 2
        }
        
        // 计算缩放比例
        const scaleX = canvasWidth / effectiveWidth
        const scaleY = canvasHeight / effectiveHeight
        
        // 获取鼠标相对于实际显示区域的位置
        const x = (e.clientX - rect.left - offsetX) * scaleX
        const y = (e.clientY - rect.top - offsetY) * scaleY
        
        // 确保坐标在画布范围内
        return {
          x: Math.max(0, Math.min(canvasWidth, x)),
          y: Math.max(0, Math.min(canvasHeight, y))
        }
      }

      // 鼠标按下事件
      canvas.addEventListener('mousedown', (e) => {
        this.isDrawing = true
        const coords = getScaledCoordinates(e, canvas)
        this.lastX = coords.x
        this.lastY = coords.y
      })

      // 鼠标移动事件
      canvas.addEventListener('mousemove', (e) => {
        if (!this.isDrawing) return

        const coords = getScaledCoordinates(e, canvas)
        const x = coords.x
        const y = coords.y

        // 确保每次绘制前重置透明度为完全不透明
        ctx.globalAlpha = 1.0
        
        if (this.brushMode === 'brush') {
          ctx.globalCompositeOperation = 'source-over'
          ctx.strokeStyle = this.brushColor
        } else {
          ctx.globalCompositeOperation = 'destination-out'
          ctx.strokeStyle = 'rgba(0,0,0,1)'
        }

        ctx.beginPath()
        ctx.moveTo(this.lastX, this.lastY)
        ctx.lineTo(x, y)
        ctx.lineWidth = this.brushSize
        ctx.lineCap = this.brushShape === 'round' ? 'round' : 'square'
        ctx.stroke()

        this.lastX = x
        this.lastY = y
      })

      // 鼠标松开事件
      canvas.addEventListener('mouseup', () => {
        this.isDrawing = false
      })

      // 鼠标离开事件
      canvas.addEventListener('mouseleave', () => {
        this.isDrawing = false
      })
    },

    // 清空画布
    clearCanvas() {
      const canvas = this.$refs.drawCanvas
      const ctx = canvas.getContext('2d')
      ctx.clearRect(0, 0, canvas.width, canvas.height)
    },

    // 保存蒙版
    saveMask() {
      const baseCanvas = this.$refs.baseCanvas
      const drawCanvas = this.$refs.drawCanvas
      
      // 确保绘图画布使用完全不透明的设置
      const drawCtx = drawCanvas.getContext('2d')
      drawCtx.globalAlpha = 1.0
      
      // 创建一个新的画布来合成结果
      const resultCanvas = document.createElement('canvas')
      resultCanvas.width = baseCanvas.width
      resultCanvas.height = baseCanvas.height
      const resultCtx = resultCanvas.getContext('2d')
      
      // 首先绘制原图
      resultCtx.drawImage(baseCanvas, 0, 0)
      
      // 将蒙版画布作为遮罩
      resultCtx.globalCompositeOperation = 'destination-out'
      resultCtx.drawImage(drawCanvas, 0, 0)
      
      // 获取结果图片数据
      const resultDataUrl = resultCanvas.toDataURL('image/png')
      const base64Data = resultDataUrl.split(',')[1]
      const byteCharacters = atob(base64Data)
      const byteArrays = []
      
      for (let i = 0; i < byteCharacters.length; i++) {
        byteArrays.push(byteCharacters.charCodeAt(i))
      }
      
      const blob = new Blob([new Uint8Array(byteArrays)], { type: 'image/png' })
      const file = new File([blob], 'mask.png', { type: 'image/png' })
      
      // 更新表单数据和预览
      this.handleImageChange({ raw: file }, this.currentMaskParam.key)
      
      // 重置画布状态
      drawCtx.globalAlpha = 1.0
      drawCtx.globalCompositeOperation = 'source-over'
      
      // 关闭编辑器
      this.maskEditorVisible = false
    },

    // 关闭蒙版编辑器
    closeMaskEditor() {
      if (this.isDrawing) {
        this.isDrawing = false
      }
      this.maskEditorVisible = false
    },

    // 提交表单
    async submitForm() {
      if (!this.$refs.form || !this.currentModule) return
      
      try {
        await this.$refs.form.validate()
        this.submitting = true

        // 创建FormData对象
        const formData = new FormData()
        
        // 添加基本信息
        formData.append('name', this.form.taskName)
        formData.append('workflow_name', this.currentModule.id)
        formData.append('workflow_config', JSON.stringify(this.currentModule.workflow_config || {}))
        
        // 添加输入参数
        const inputParams = []
        if (this.currentModule.inputParams) {
          for (const param of this.currentModule.inputParams) {
            const value = this.form.params[param.key]
            if (value) {
              // 构建参数对象
              const paramObj = {
                key: param.key,
                alias: param.alias,
                type: param.type
              }
              
              // 如果是文件类型参数，直接添加到FormData
              if (param.type === 'image' || param.type === 'mask') {
                formData.append(param.key, value)
                // 设置参数值为文件路径（将在后端处理）
                paramObj.value = ''
              } else {
                // 对于非文件类型参数，直接设置值
                paramObj.value = value
              }
              
              // 将参数对象添加到数组
              inputParams.push(paramObj)
            }
          }
        }
        // 将inputParams添加到FormData
        formData.append('input_params', JSON.stringify(inputParams))

        // 发送请求
        const response = await createImageProcessingTask(this.currentModule.id, formData)
        if (response.success) {
          this.$message.success('任务创建成功')
          this.dialogVisible = false
          this.fetchTasks()
        }
      } catch (error) {
        this.$message.error('创建任务失败：' + error.message)
      } finally {
        this.submitting = false
      }
    },

    // 查看任务详情
    async viewDetail(taskId) {
      if (!this.currentModule || !this.currentModule.id) return
      try {
        const response = await getImageProcessingTaskDetail(this.currentModule.id, taskId)
        if (response.success) {
          // 跳转到任务详情页面
          this.$router.push(`/image-processing/${this.currentModule.id}/task/${taskId}`)
        }
      } catch (error) {
        this.$message.error('获取任务详情失败：' + error.message)
      }
    },

    // 格式化输入参数
    formatInputParams(inputParamsStr) {
      try {
        const params = JSON.parse(inputParamsStr)
        return params.map(param => `
          <div class="param-item">
            <strong>${param.alias || param.key}：</strong>
            ${param.type === 'image' || param.type === 'mask' ? 
              `<img src="${param.value}" class="param-image" alt="${param.alias || param.key}">` : 
              param.value}
          </div>
        `).join('')
      } catch (error) {
        return '无法解析输入参数'
      }
    },

    // 格式化输出路径
    formatOutputPaths(outputPathsStr) {
      try {
        const paths = JSON.parse(outputPathsStr)
        return Object.entries(paths).map(([key, path]) => `
          <div class="output-item">
            <img src="${path}" class="output-image" alt="${key}">
          </div>
        `).join('')
      } catch (error) {
        return '无法解析输出结果'
      }
    },

    // 删除任务
    async deleteTask(task) {
      try {
        await this.$confirm('确认删除该任务？', '提示', {
          type: 'warning'
        })
        
        const response = await deleteImageProcessingTask(this.currentModule.id, task.id)
        if (response.success) {
          this.$message.success('删除任务成功')
          this.fetchTasks()
        }
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('删除任务失败：' + error.message)
        }
      }
    },

    // 格式化日期
    formatDate(date) {
      return new Date(date).toLocaleString()
    },

    // 获取状态类型
    getStatusType(status) {
      const statusMap = {
        pending: 'info',
        processing: 'warning',
        completed: 'success',
        failed: 'danger'
      }
      return statusMap[status] || 'info'
    },

    // 获取状态文本
    getStatusText(status) {
      const statusMap = {
        pending: '待处理',
        processing: '处理中',
        completed: '已完成',
        failed: '失败'
      }
      return statusMap[status] || status
    },

    // 切换视图
    toggleView() {
      this.isCardView = !this.isCardView
      localStorage.setItem('image_processing_view_mode', this.isCardView ? 'card' : 'list')
      
      // 重置状态
      this.currentPage = 1
      this.tasks = []
      this.hasMoreData = true
      this.loadingMore = false
      
      // 重新加载第一页数据
      this.fetchTasks()
      
      // 如果切换到卡片视图，设置IntersectionObserver用于无限滚动
      if (this.isCardView) {
        this.$nextTick(() => {
          this.setupIntersectionObserver()
        })
      }
    },

    // 处理分页
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1 // 改变每页条数后重置为第一页
      this.tasks = [] // 清空任务列表
      this.fetchTasks()
    },

    handleCurrentChange(page) {
      if (page === this.currentPage) {
        console.log('页码未变化，跳过请求')
        return
      }
      this.currentPage = page
      this.tasks = [] // 切换页码时清空任务列表
      this.fetchTasks()
    },

    // 确保数据加载
    ensureDataLoaded() {
      console.log('检查是否需要加载数据:', 
        'initialLoaded=', this.initialLoaded, 
        'currentModule=', !!this.currentModule, 
        'tasks.length=', this.tasks.length,
        'loading=', this.loading)
      
      // 如果没有当前模块，跳过
      if (!this.currentModule) {
        console.log('没有当前模块，跳过数据加载')
        return
      }
      
      // 如果任务列表为空且未在加载中且未标记为已加载，则加载数据
      if (this.tasks.length === 0 && !this.loading && !this.initialLoaded) {
        console.log('满足加载条件，开始强制加载任务列表')
        this.currentPage = 1
        this.loading = false // 确保loading状态重置
        // 使用setTimeout避免可能的状态冲突
        setTimeout(() => {
          this.fetchTasks().then(() => {
            console.log('强制加载任务列表完成')
            this.initialLoaded = true
          }).catch(err => {
            console.error('强制加载任务列表失败:', err)
          })
        }, 0)
      }
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
        if (documentHeight - scrollTop - windowHeight < this.scrollThreshold) {
          console.log('窗口滚动触发加载更多');
          this.loadMoreTasks();
        }
      }
    },
    
    // 加载更多数据
    loadMoreTasks() {
      if (this.loadingMore || !this.hasMoreData) {
        console.log('跳过加载更多:', '加载中=', this.loadingMore, '没有更多数据=', !this.hasMoreData)
        return
      }
      
      // 计算总页数
      const totalPages = Math.ceil(this.total / this.cardPageSize) || 1
      
      // 防止页码超出范围
      if (this.currentPage >= totalPages) {
        console.log('已达到最后一页，不再加载更多', '当前页=', this.currentPage, '总页数=', totalPages)
        this.hasMoreData = false
        return
      }
      
      console.log('开始加载更多数据，当前页码：', this.currentPage, '总页数：', totalPages)
      this.currentPage++
      this.fetchTasks(true)
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
          console.log('未找到加载更多触发元素')
          return
        }
        
        console.log('设置观察者')
        
        // 创建新的IntersectionObserver
        this.observer = new IntersectionObserver((entries) => {
          const entry = entries[0]
          console.log('intersection事件:', '可见=', entry.isIntersecting, '加载中=', this.loadingMore, '有更多数据=', this.hasMoreData)
          if (entry.isIntersecting && !this.loadingMore && this.hasMoreData) {
            console.log('观察者触发加载更多')
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

    // 初始加载数据
    loadInitialData() {
      if (this.initialLoaded) {
        console.log('已加载初始数据，跳过')
        return
      }
      
      console.log('加载初始数据')
      this.currentPage = 1
      this.tasks = []
      this.hasMoreData = true
      this.fetchTasks()
      this.initialLoaded = true
    },
  }
}
</script>

<style scoped>
.workflow-container {
  padding: 20px;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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
  padding: 12px 12px;
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
  .workflow-container {
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
  
  /* 悬浮按钮样式 */
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
    transform: scale(0.98);
    opacity: 0.9;
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
</style>

<style>
/* 全局覆盖Element UI组件的样式 */
.card-list .el-loading-mask {
  margin-top: 0 !important;
  padding-top: 0 !important;
}

@media screen and (max-width: 768px) {
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