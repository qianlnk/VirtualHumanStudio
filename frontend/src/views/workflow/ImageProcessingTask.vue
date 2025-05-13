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
          <div class="task-card" v-for="item in tasks" :key="item.id" :data-task-id="item.id">
            <div class="task-card-header">
              <h3 class="task-card-title">{{ item.name }}</h3>
              <div class="status-icon">
                <i v-if="item.status === 'completed'" class="el-icon-check" style="color: #67c23a;"></i>
                <i v-else-if="item.status === 'failed'" class="el-icon-close" style="color: #f56c6c;"></i>
                <el-tag v-else :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
              </div>
            </div>
            <div class="task-card-content">
              <!-- 添加图片轮播 -->
              <div class="task-card-carousel" v-if="hasImages(item)">
                <el-carousel 
                  indicator-position="outside" 
                  :height="isMobile ? '160px' : '180px'"
                  :interval="4000" 
                  arrow="hover" 
                  :autoplay="false" 
                  indicator-color="#ddd" 
                  type=""
                  :key="'carousel-' + item.id"
                  :ref="`carousel-${item.id}`"
                  loop
                  @change="(index) => updateCarouselIndex(item.id, index)">
                  <el-carousel-item v-for="(image, index) in getTaskImages(item)" :key="`${image.type}-${index}`">
                    <div class="carousel-item" :data-is-output="image.type === 'output'"
                      @touchstart="onCarouselTouchStart($event, item.id)" 
                      @touchmove="onCarouselTouchMove($event, item.id)" 
                      @touchend="onCarouselTouchEnd($event, item.id)">
                      <div class="carousel-image-wrapper">
                        <img :src="image.src" class="carousel-image" :alt="image.alt" draggable="false">
                      </div>
                      <div class="carousel-label">
                        <span :class="image.type === 'output' ? 'carousel-output-label' : 'carousel-input-label'">
                          {{ image.type === 'output' ? '输出: ' : '输入: ' }}{{ image.label }}
                        </span>
                      </div>
                    </div>
                  </el-carousel-item>
                </el-carousel>
              </div>
              
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
    <el-dialog 
      title="蒙版编辑" 
      :visible.sync="maskEditorVisible" 
      :width="isMobile ? '100%' : '80%'" 
      :fullscreen="isMobile"
      :before-close="closeMaskEditor"
      :append-to-body="true"
      :close-on-click-modal="false"
      :show-close="!isMobile"
      custom-class="mask-editor-dialog">
      
      <!-- 移动端顶部导航 -->
      <div v-if="isMobile" class="mobile-header-bar mask-editor-header">
        <div class="header-back" @click="closeMaskEditor">
          <i class="el-icon-arrow-left"></i>
          <span>返回</span>
        </div>
        <div class="header-title">蒙版编辑</div>
      </div>
      
      <div class="mask-editor-container" :class="{'mobile-mask-editor': isMobile}">
        <div class="editor-tools" :class="{'mobile-editor-tools': isMobile}">
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
              :style="{ width: isMobile ? '140px' : '200px' }"
            ></el-slider>
            <span class="brush-size-label">{{brushSize}}px</span>
          </div>
          
          <el-button size="small" type="danger" @click="clearCanvas">清空</el-button>
        </div>
        
        <div class="canvas-container" :class="{'mobile-canvas-container': isMobile}">
          <div class="canvas-wrapper">
            <canvas ref="baseCanvas" class="editor-canvas base-canvas"></canvas>
            <canvas ref="drawCanvas" class="editor-canvas draw-canvas" :style="{ cursor: customCursor || 'crosshair' }"></canvas>
          </div>
        </div>
        
        <div class="editor-actions" :class="{'mobile-editor-actions': isMobile}">
          <template v-if="!isMobile">
            <el-button type="primary" @click="saveMask">保存蒙版</el-button>
            <el-button @click="closeMaskEditor">取消</el-button>
          </template>
        </div>
      </div>
      
      <!-- 移动端底部按钮 -->
      <div v-if="isMobile" class="mobile-form-footer">
        <el-button type="primary" @click="saveMask" class="mobile-submit-btn">保存蒙版</el-button>
      </div>
    </el-dialog>
    
    <!-- 创建任务对话框 -->
    <el-dialog 
      :title="`创建${currentModule ? currentModule.name : '图像处理'}任务`" 
      :visible.sync="dialogVisible" 
      :width="isMobile ? '100%' : '600px'"
      :fullscreen="isMobile"
      :modal="true"
      :close-on-click-modal="false"
      :append-to-body="true"
      :show-close="!isMobile"
      custom-class="image-processing-dialog">
      
      <!-- 移动端顶部导航 -->
      <div v-if="isMobile" class="mobile-header-bar">
        <div class="header-back" @click="dialogVisible = false">
          <i class="el-icon-arrow-left"></i>
          <span>返回</span>
        </div>
        <div class="header-title">
          创建{{currentModule ? currentModule.name : '图像处理'}}任务
        </div>
      </div>
      
      <el-form 
        :model="form" 
        :rules="rules" 
        ref="form" 
        :label-width="isMobile ? '90px' : '120px'"
        :label-position="isMobile ? 'top' : 'left'"
        class="image-processing-form">
        
        <!-- 任务名称输入框 -->
        <el-form-item label="任务名称" prop="taskName">
          <el-input
            type="textarea"
            v-model="form.taskName"
            placeholder="请输入任务名称"
            resize="both"
            :rows="2"
            style="width: 100%;"
            @focus="handleInputFocus"
            @blur="handleInputBlur"
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
              @focus="handleInputFocus"
              @blur="handleInputBlur"
            ></el-input>
            
            <!-- 选择类型参数 -->
            <el-select
              v-else-if="param.type === 'select'"
              v-model="form.params[param.key]"
              :placeholder="`请选择${param.alias}`"
              style="width: 100%;">
              <el-option
                v-for="option in param.options"
                :key="option.value"
                :label="option.label"
                :value="option.value">
              </el-option>
            </el-select>
            
            <!-- 图片类型参数 -->
            <template v-else-if="param.type === 'image' || param.type === 'mask'">
              <div class="image-param-container" :class="{'mobile-image-container': isMobile}">
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
      cursorSize: 10, // 添加光标大小变量
      customCursor: '', // 添加自定义光标变量
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
      lastScrollTop: 0, // 记录上次滚动位置
      // 视口控制相关变量
      originalViewportContent: null,
      isInputFocused: false,
      // 轮播图触摸相关状态
      carouselTouchStart: null,
      carouselTouchCurrent: null,
      activeCarouselIndex: {},
      isDraggingCarousel: false
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
    },
    
    // 监听笔刷大小变化
    brushSize() {
      this.updateCursor();
    },
    
    // 监听笔刷形状变化
    brushShape() {
      this.updateCursor();
    },
    
    // 监听笔刷模式变化
    brushMode() {
      this.updateCursor();
    },
    
    // 监听笔刷颜色变化
    brushColor() {
      this.updateCursor();
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
      
      // 设置移动端视口
      this.setupMobileViewport()
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
    
    // 重置视口设置
    this.resetMobileViewport()
  },
  methods: {
    // 返回上一页
    goBack() {
      this.$router.back();
    },
    
    // 处理触摸事件，阻止图片拖拽但允许轮播滑动
    handleTouchEvent(e) {
      // 允许事件冒泡，以便轮播图可以响应滑动
      // 但阻止默认行为，防止图片拖拽
      if (e.target.tagName.toLowerCase() === 'img') {
        e.preventDefault();
      }
    },
    
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
      
      // 对话框打开时设置移动端视口
      if (this.isMobile) {
        this.$nextTick(() => {
          this.setupMobileViewport();
        });
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
      
      // 对话框打开时设置移动端视口
      if (this.isMobile) {
        this.$nextTick(() => {
          this.setupMobileViewport();
        });
      }

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
            
            // 更新光标
            this.updateCursor()
          }
          maskImg.src = URL.createObjectURL(this.form.params[param.key])
        } else {
          // 如果没有蒙版，直接初始化绘制事件
          this.initDrawEvents()
          
          // 更新光标
          this.updateCursor()
        }
      }
      img.onerror = () => {
        this.$message.error('图片加载失败')
        this.maskEditorVisible = false
      }
      img.src = URL.createObjectURL(baseImageFile)
    },
    
    // 更新光标样式
    updateCursor() {
      // 确保画布已经初始化
      if (!this.$refs.drawCanvas) return
      
      // 应用缩放因子使光标大小与实际绘制效果匹配
      // 进一步缩小缩放因子，使光标更贴近实际绘制线条粗细
      const scaleFactor = 0.35
      const displaySize = Math.round(this.brushSize * scaleFactor)
      // 确保光标最小尺寸
      const finalSize = Math.max(displaySize, 3)
      this.cursorSize = finalSize
      
      // 创建SVG光标
      let svgCursor
      if (this.brushShape === 'round') {
        // 圆形光标
        if (this.brushMode === 'brush') {
          // 画笔模式 - 实心圆
          svgCursor = `url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='${finalSize * 2}' height='${finalSize * 2}' viewBox='0 0 ${finalSize * 2} ${finalSize * 2}'><circle cx='${finalSize}' cy='${finalSize}' r='${finalSize}' fill='${this.brushColor === 'black' ? '%23000' : (this.brushColor === 'white' ? '%23fff' : '%23999')}' fill-opacity='0.3' stroke='%23000' stroke-width='1'/></svg>") ${finalSize} ${finalSize}, crosshair`
        } else {
          // 橡皮擦模式 - 空心圆
          svgCursor = `url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='${finalSize * 2}' height='${finalSize * 2}' viewBox='0 0 ${finalSize * 2} ${finalSize * 2}'><circle cx='${finalSize}' cy='${finalSize}' r='${finalSize}' fill='%23fff' fill-opacity='0.2' stroke='%23000' stroke-width='1.5'/></svg>") ${finalSize} ${finalSize}, cell`
        }
      } else {
        // 方形光标
        if (this.brushMode === 'brush') {
          // 画笔模式 - 实心方形
          svgCursor = `url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='${finalSize * 2}' height='${finalSize * 2}' viewBox='0 0 ${finalSize * 2} ${finalSize * 2}'><rect x='${finalSize - finalSize}' y='${finalSize - finalSize}' width='${finalSize * 2}' height='${finalSize * 2}' fill='${this.brushColor === 'black' ? '%23000' : (this.brushColor === 'white' ? '%23fff' : '%23999')}' fill-opacity='0.3' stroke='%23000' stroke-width='1'/></svg>") ${finalSize} ${finalSize}, crosshair`
        } else {
          // 橡皮擦模式 - 空心方形
          svgCursor = `url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='${finalSize * 2}' height='${finalSize * 2}' viewBox='0 0 ${finalSize * 2} ${finalSize * 2}'><rect x='${finalSize - finalSize}' y='${finalSize - finalSize}' width='${finalSize * 2}' height='${finalSize * 2}' fill='%23fff' fill-opacity='0.2' stroke='%23000' stroke-width='1.5'/></svg>") ${finalSize} ${finalSize}, cell`
        }
      }
      
      // 设置自定义光标
      this.customCursor = svgCursor
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
        
        // 获取坐标位置（支持触摸事件和鼠标事件）
        let clientX, clientY;
        
        if (e.touches && e.touches.length > 0) {
          // 触摸事件
          clientX = e.touches[0].clientX;
          clientY = e.touches[0].clientY;
        } else {
          // 鼠标事件
          clientX = e.clientX;
          clientY = e.clientY;
        }
        
        // 获取相对于实际显示区域的位置
        const x = (clientX - rect.left - offsetX) * scaleX
        const y = (clientY - rect.top - offsetY) * scaleY
        
        // 确保坐标在画布范围内
        return {
          x: Math.max(0, Math.min(canvasWidth, x)),
          y: Math.max(0, Math.min(canvasHeight, y))
        }
      }

      // 绘制函数
      const draw = (e) => {
        if (!this.isDrawing) return;

        const coords = getScaledCoordinates(e, canvas);
        const x = coords.x;
        const y = coords.y;

        // 确保每次绘制前重置透明度为完全不透明
        ctx.globalAlpha = 1.0;
        
        if (this.brushMode === 'brush') {
          ctx.globalCompositeOperation = 'source-over';
          ctx.strokeStyle = this.brushColor;
        } else {
          ctx.globalCompositeOperation = 'destination-out';
          ctx.strokeStyle = 'rgba(0,0,0,1)';
        }

        ctx.beginPath();
        ctx.moveTo(this.lastX, this.lastY);
        ctx.lineTo(x, y);
        ctx.lineWidth = this.brushSize;
        ctx.lineCap = this.brushShape === 'round' ? 'round' : 'square';
        ctx.stroke();

        this.lastX = x;
        this.lastY = y;
      };

      // 开始绘制
      const startDrawing = (e) => {
        this.isDrawing = true;
        const coords = getScaledCoordinates(e, canvas);
        this.lastX = coords.x;
        this.lastY = coords.y;
      };

      // 停止绘制
      const stopDrawing = () => {
        this.isDrawing = false;
      };

      // 鼠标事件
      canvas.addEventListener('mousedown', startDrawing);
      canvas.addEventListener('mousemove', draw);
      canvas.addEventListener('mouseup', stopDrawing);
      canvas.addEventListener('mouseleave', stopDrawing);

      // 触摸事件（移动端）
      canvas.addEventListener('touchstart', (e) => {
        e.preventDefault(); // 防止页面滚动
        startDrawing(e);
      }, { passive: false });
      
      canvas.addEventListener('touchmove', (e) => {
        e.preventDefault(); // 防止页面滚动
        draw(e);
      }, { passive: false });
      
      canvas.addEventListener('touchend', (e) => {
        e.preventDefault(); // 防止页面滚动
        stopDrawing();
      }, { passive: false });
      
      canvas.addEventListener('touchcancel', (e) => {
        e.preventDefault(); // 防止页面滚动
        stopDrawing();
      }, { passive: false });
    },

    // 清空画布
    clearCanvas() {
      const canvas = this.$refs.drawCanvas
      const ctx = canvas.getContext('2d')
      ctx.clearRect(0, 0, canvas.width, canvas.height)
      
      // 重置光标样式
      this.updateCursor()
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
      
      // 重置视口设置
      if (this.isMobile) {
        this.$nextTick(() => {
          this.resetMobileViewport();
        });
      }
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
          
          // 重置视口设置
          if (this.isMobile) {
            this.$nextTick(() => {
              this.resetMobileViewport();
            });
          }
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
        if (!outputPathsStr) return {};
        
        // 尝试将字符串解析为JSON对象
        let paths;
        if (typeof outputPathsStr === 'string') {
          try {
            paths = JSON.parse(outputPathsStr);
          } catch (e) {
            console.error('输出路径JSON解析失败:', e);
            return {};
          }
        } else if (typeof outputPathsStr === 'object') {
          // 已经是对象，直接使用
          paths = outputPathsStr;
        } else {
          console.error('无法识别的输出路径格式:', typeof outputPathsStr);
          return {};
        }
        
        // 过滤掉非字符串值，确保路径有效
        const result = {};
        if (paths && typeof paths === 'object') {
          Object.entries(paths).forEach(([key, path]) => {
            if (typeof path === 'string' && path.trim()) {
              result[key] = path;
            }
          });
        }
        
        return result;
      } catch (error) {
        console.error('解析输出图片路径失败:', error);
        return {};
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

    // 解析输入参数中的图片类型参数
    getImageParams(inputParamsStr) {
      try {
        if (!inputParamsStr) return [];
        
        // 尝试解析字符串
        let params;
        if (typeof inputParamsStr === 'string') {
          try {
            params = JSON.parse(inputParamsStr);
          } catch (e) {
            console.error('输入参数JSON解析失败:', e);
            return [];
          }
        } else if (Array.isArray(inputParamsStr)) {
          // 已经是数组，直接使用
          params = inputParamsStr;
        } else {
          console.error('无法识别的输入参数格式:', typeof inputParamsStr);
          return [];
        }
        
        // 过滤图片类型参数
        return params.filter(param => 
          (param.type === 'image' || param.type === 'mask') && 
          param.value && 
          typeof param.value === 'string'
        );
      } catch (error) {
        console.error('解析输入图片参数失败:', error);
        return [];
      }
    },

    // 解析输出路径中的图片
    getOutputImages(outputPathsStr) {
      try {
        if (!outputPathsStr) return {};
        
        // 尝试将字符串解析为JSON对象
        let paths;
        if (typeof outputPathsStr === 'string') {
          try {
            paths = JSON.parse(outputPathsStr);
          } catch (e) {
            console.error('输出路径JSON解析失败:', e);
            return {};
          }
        } else if (typeof outputPathsStr === 'object') {
          // 已经是对象，直接使用
          paths = outputPathsStr;
        } else {
          console.error('无法识别的输出路径格式:', typeof outputPathsStr);
          return {};
        }
        
        // 过滤掉非字符串值，确保路径有效
        const result = {};
        if (paths && typeof paths === 'object') {
          Object.entries(paths).forEach(([key, path]) => {
            if (typeof path === 'string' && path.trim()) {
              result[key] = path;
            }
          });
        }
        
        console.log('处理后的输出图片路径:', result);
        return result;
      } catch (error) {
        console.error('解析输出图片路径失败:', error);
        return {};
      }
    },

    // 检查任务是否有图片可显示
    hasImages(item) {
      // 对于已完成的任务检查是否有输出
      if (item.status === 'completed') {
        // 检查output_params（包含图片参数的情况）
        if (item.output_params) {
          try {
            let params = item.output_params;
            if (typeof params === 'string') {
              params = JSON.parse(params);
            }
            
            if (Array.isArray(params)) {
              for (const param of params) {
                if (param && (param.type === 'image' || param.type === 'mask') && 
                    param.value && typeof param.value === 'string' && param.value.trim()) {
                  return true;
                }
              }
            }
          } catch (e) {
            console.error('检查输出参数图片失败:', e);
          }
        }
        
        // 检查output_paths（直接包含图片路径的情况）
        if (item.output_paths) {
          try {
            let outputPaths = item.output_paths;
            if (typeof outputPaths === 'string') {
              outputPaths = JSON.parse(outputPaths);
            }
            
            if (typeof outputPaths === 'object' && !Array.isArray(outputPaths)) {
              // 检查是否有有效的图片路径
              for (const key in outputPaths) {
                if (typeof outputPaths[key] === 'string' && outputPaths[key].trim()) {
                  return true;
                }
              }
            }
          } catch (e) {
            console.error('检查输出图片路径失败:', e);
          }
        }
      }
      
      // 检查是否有输入图片
      if (item.input_params) {
        try {
          let params = item.input_params;
          if (typeof params === 'string') {
            params = JSON.parse(params);
          }
          
          if (Array.isArray(params)) {
            // 检查是否有图片类型的参数
            for (const param of params) {
              if (param && (param.type === 'image' || param.type === 'mask') && 
                  param.value && typeof param.value === 'string' && param.value.trim()) {
                return true;
              }
            }
          }
        } catch (e) {
          console.error('检查输入图片失败:', e);
        }
      }
      
      return false;
    },

    // 获取任务的图片（先输出后输入）
    getTaskImages(item) {
      const images = [];
      
      // 只有完成状态的任务才显示输出图片
      if (item.status === 'completed') {
        // 检查output_params中的图片
        if (item.output_params) {
          try {
            let params = item.output_params;
            if (typeof params === 'string') {
              params = JSON.parse(params);
            }
            
            if (Array.isArray(params)) {
              for (const param of params) {
                if (param && (param.type === 'image' || param.type === 'mask') && 
                    param.value && typeof param.value === 'string' && param.value.trim()) {
                  images.push({
                    src: param.value,
                    alt: param.alias || param.key,
                    label: param.alias || param.key,
                    type: 'output'
                  });
                }
              }
            }
          } catch (e) {
            console.error('处理输出参数图片失败:', e);
          }
        }
        
        // 检查output_paths中的图片路径
        if (item.output_paths && images.length === 0) {
          try {
            let paths = item.output_paths;
            if (typeof paths === 'string') {
              paths = JSON.parse(paths);
            }
            
            if (typeof paths === 'object' && !Array.isArray(paths)) {
              for (const [key, path] of Object.entries(paths)) {
                if (typeof path === 'string' && path.trim()) {
                  images.push({
                    src: path,
                    alt: `输出: ${key}`,
                    label: key,
                    type: 'output'
                  });
                }
              }
            }
          } catch (e) {
            console.error('处理输出路径图片失败:', e);
          }
        }
      }
      
      // 添加输入图片
      if (item.input_params) {
        try {
          let params = item.input_params;
          if (typeof params === 'string') {
            params = JSON.parse(params);
          }
          
          if (Array.isArray(params)) {
            for (const param of params) {
              if (param && (param.type === 'image' || param.type === 'mask') && 
                  param.value && typeof param.value === 'string' && param.value.trim()) {
                images.push({
                  src: param.value,
                  alt: param.alias || param.key,
                  label: param.alias || param.key,
                  type: 'input'
                });
              }
            }
          }
        } catch (e) {
          console.error('处理输入图片失败:', e);
        }
      }
      
      // 确保至少有一张图片可显示
      if (images.length === 0) {
        console.log('任务没有找到任何可显示的图片:', item.id);
      }
      
      return images;
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

    // 轮播图触摸事件处理
    onCarouselTouchStart(event, taskId) {
      this.isDraggingCarousel = true;
      this.carouselTouchStart = event.touches[0].clientX;
      this.carouselTouchCurrent = event.touches[0].clientX;
      
      // 确保该任务ID有索引记录
      if (this.activeCarouselIndex[taskId] === undefined) {
        this.activeCarouselIndex[taskId] = 0;
      }
    },

    onCarouselTouchMove(event) {
      if (!this.isDraggingCarousel) return;
      
      // 阻止默认行为，防止页面滚动
      event.preventDefault();
      this.carouselTouchCurrent = event.touches[0].clientX;
    },

    onCarouselTouchEnd(event, taskId) {
      if (!this.isDraggingCarousel) return;
      
      const delta = this.carouselTouchCurrent - this.carouselTouchStart;
      const images = this.getTaskImages(this.tasks.find(t => t.id === taskId));
      const maxIndex = images.length - 1;
      
      if (Math.abs(delta) > 50) { // 降低触发阈值，提高灵敏度
        if (delta > 0) {
          // 向右滑动，显示上一张
          if (this.activeCarouselIndex[taskId] === 0) {
            // 如果当前是第一张，循环到最后一张
            this.activeCarouselIndex[taskId] = maxIndex;
          } else {
            // 否则显示上一张
            this.activeCarouselIndex[taskId]--;
          }
        } else {
          // 向左滑动，显示下一张
          if (this.activeCarouselIndex[taskId] === maxIndex) {
            // 如果当前是最后一张，循环到第一张
            this.activeCarouselIndex[taskId] = 0;
          } else {
            // 否则显示下一张
            this.activeCarouselIndex[taskId]++;
          }
        }
        
        // 手动控制对应轮播图
        this.setCarouselToIndex(taskId, this.activeCarouselIndex[taskId]);
      }
      
      this.isDraggingCarousel = false;
    },

    // 更新轮播图索引记录
    updateCarouselIndex(taskId, index) {
      this.activeCarouselIndex[taskId] = index;
    },

    // 手动设置轮播图到指定索引
    setCarouselToIndex(taskId, index) {
      // 使用ref获取轮播图实例
      const refName = `carousel-${taskId}`;
      const carousel = this.$refs[refName];
      
      if (carousel) {
        if (Array.isArray(carousel) && carousel.length > 0) {
          carousel[0].setActiveItem(index);
        } else {
          carousel.setActiveItem(index);
        }
      }
    }
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

.header-left {
  display: flex;
  align-items: center;
}

.back-icon {
  font-size: 20px;
  margin-right: 10px;
  cursor: pointer;
  color: #fff;
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
  transform: none !important; /* 确保没有任何变形 */
}

.task-card:hover {
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  border-color: rgba(255, 255, 255, 0.2);
  transform: none !important; /* 确保悬停时没有变形 */
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
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  overflow: hidden; /* 改回hidden，完全禁止滚动 */
  min-height: 0; /* 允许内容区域收缩，避免溢出 */
  max-height: 100%; /* 限制最大高度，避免溢出 */
}

.task-card-info {
  margin: 8px 0 5px 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.task-card-info p {
  margin: 0;
  font-size: 13px;
  color: #ddd;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-card-footer {
  padding: 10px;
  display: flex;
  justify-content: space-around;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 0, 0, 0.1);
  margin-top: auto;
  flex-shrink: 0; /* 防止底部被压缩 */
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
    padding-bottom: 100px; /* 增加底部内边距 */
  }
  
  .task-list {
    margin-top: 10px;
    padding-bottom: 100px; /* 增加底部内边距 */
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
    padding-bottom: 120px; /* 增加更多底部空间 */
  }
  
  .card-view-content {
    padding: 0;
    margin: 0;
    padding-bottom: 120px; /* 增加更多底部空间 */
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
    opacity: 0.95;
    transform: none !important; /* 确保激活时没有变形 */
  }
  
  /* 加载更多容器样式 */
  .load-more-container {
    text-align: center;
    padding: 20px 0;
    margin: 20px 0 120px 0; /* 增加底部边距 */
    height: auto;
    min-height: 80px;
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

/* 添加蒙版编辑器样式 */
.mask-editor-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  padding: 10px;
  box-sizing: border-box;
}

.editor-tools {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 10px;
  background: rgba(0, 0, 0, 0.05);
  padding: 10px;
  border-radius: 5px;
  align-items: center;
}

.tool-group {
  display: flex;
  align-items: center;
  margin-right: 15px;
}

.tool-label {
  margin-right: 8px;
  font-size: 14px;
  white-space: nowrap;
}

.brush-size-label {
  margin-left: 8px;
  font-size: 12px;
  white-space: nowrap;
  min-width: 40px;
}

.canvas-container {
  position: relative;
  width: 100%;
  height: 500px;
  overflow: hidden;
  background: #f5f5f5;
  border-radius: 5px;
  margin-bottom: 10px;
  flex-grow: 1;
}

.canvas-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
}

.editor-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.base-canvas {
  z-index: 1;
}

.draw-canvas {
  z-index: 2;
  cursor: crosshair;
}

.editor-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 10px 0;
}

/* 图片上传和预览样式 */
.image-param-container {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  align-items: center;
}

.image-preview {
  width: 150px;
  height: 150px;
  border: 1px dashed #ccc;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: #f5f5f5;
}

.preview-thumbnail {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

/* 轮播图样式 */
.task-card-carousel {
  margin: 0 -10px 5px; /* 减少底部margin，避免内容溢出 */
  width: calc(100% + 20px);
  border-radius: 4px;
  overflow: hidden;
  padding: 3px;
  box-sizing: border-box;
  flex-shrink: 0; /* 防止轮播图被压缩 */
}

.carousel-item {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  position: relative;
}

.carousel-image-wrapper {
  height: 140px;
  max-height: 140px;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
}

.carousel-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.carousel-label {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 4px 8px;
  font-size: 12px;
  z-index: 1;
  display: flex;
  align-items: center;
  border-radius: 0 0 4px 4px;
}

.carousel-input-label {
  background: linear-gradient(90deg, #1976d2, #64b5f6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  font-weight: bold;
  text-shadow: 0 0 1px rgba(255, 255, 255, 0.5);
}

.carousel-output-label {
  background: linear-gradient(90deg, #43a047, #81c784);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  font-weight: bold;
  text-shadow: 0 0 1px rgba(255, 255, 255, 0.5);
  position: relative;
  padding-left: 18px;
}

.carousel-output-label::before {
  content: "★";
  position: absolute;
  left: 0;
  top: -1px;
  color: #4caf50;
  font-size: 14px;
  text-shadow: none;
  -webkit-text-fill-color: #4caf50;
}

/* 响应式样式调整 */
@media screen and (max-width: 768px) {
  .task-card-carousel {
    margin: 0 -8px 3px; /* 移动端底部margin更小 */
    width: calc(100% + 16px);
  }
  
  .task-card-footer {
    padding: 8px 10px; /* 减小移动端padding */
  }
  
  /* 移动端卡片样式调整 */
  .task-card {
    height: auto; /* 自适应高度 */
    min-height: 0;
  }
  
  .carousel-image-wrapper {
    height: 120px;
    max-height: 120px;
  }
  
  .carousel-label {
    padding: 3px 6px;
    font-size: 11px;
  }
}

/* 结果图片的特殊效果 */
.carousel-item[data-is-output="true"] .carousel-image-wrapper {
  border: 2px solid rgba(76, 175, 80, 0.5);
  box-shadow: 0 0 8px rgba(76, 175, 80, 0.3);
  box-sizing: border-box;
  margin: 2px;
  width: calc(100% - 4px);
}

@media screen and (max-width: 768px) {
  .carousel-output-label::before {
    font-size: 12px;
    top: 0;
  }
  
  .carousel-item[data-is-output="true"] .carousel-image-wrapper {
    margin: 1px;
    width: calc(100% - 2px);  /* 移动端边框更窄 */
  }
  
  /* 确保移动端也不显示滚动条 */
  .task-card-content {
    overflow: hidden;
    min-height: 0;
  }
  
  .task-card-info {
    margin: 5px 0 0 0;
  }
  
  .carousel-image-wrapper {
    height: 120px;
    max-height: 120px;
  }
}

/* 修复轮播图项在卡片模式下的样式 */
.el-carousel__item--card {
  border: none !important;
}

/* 调整轮播图容器样式避免边框显示不全 */
.task-card .el-carousel__container {
  padding: 1px;
  box-sizing: border-box;
}

@media screen and (max-width: 768px) {
  .task-card .el-carousel__indicators {
    bottom: 2px !important;
  }
  
  .task-card .el-carousel__arrow {
    width: 20px !important;
    height: 20px !important;
    font-size: 10px !important;
  }
  
  /* 移动端轮播图高度调整 */
  .task-card .el-carousel {
    height: 160px !important;
  }
  
  /* 确保移动端可以触摸滑动 */
  .task-card .el-carousel__container {
    touch-action: pan-y !important;
  }
  
  /* 隐藏卡片效果的多余内边距 */
  .task-card .el-carousel__item {
    padding: 0 !important;
  }
}

.el-carousel__container {
  position: relative;
  height: 100%;
  touch-action: pan-x !important; /* 启用横向触摸滑动 */
}

@media screen and (max-width: 768px) {
  .task-card .el-carousel__container {
    touch-action: pan-x !important;
    height: 160px !important;
  }
  
  /* 确保移动端的轮播图可以滑动 */
  .card-list .el-carousel {
    touch-action: pan-x !important;
  }
  
  .el-carousel__item {
    touch-action: pan-x !important;
  }
  
  /* 确保图片不会阻止滑动事件 */
  .carousel-image {
    pointer-events: none;
    user-select: none;
    -webkit-user-drag: none;
  }
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
  
  /* 移动端对话框样式 */
  .el-dialog.image-processing-dialog,
  .el-dialog.mask-editor-dialog {
    width: 100% !important;
    height: 100% !important;
    margin: 0 !important;
    border-radius: 0 !important;
    overflow: hidden !important;
    padding: 0 !important;
    position: fixed !important;
    top: 0 !important;
    left: 0 !important;
    transform: none !important;
  }
  
  .el-dialog.image-processing-dialog .el-dialog__body,
  .el-dialog.mask-editor-dialog .el-dialog__body {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    padding: 0 !important;
    overflow-y: auto !important;
    -webkit-overflow-scrolling: touch !important;
    padding-top: 56px !important; /* 为顶部导航留出空间 */
    padding-bottom: 60px !important; /* 为底部按钮留出空间 */
    height: 100% !important;
  }
  
  /* 隐藏原始对话框标题 */
  .image-processing-dialog .el-dialog__header,
  .mask-editor-dialog .el-dialog__header {
    display: none !important;
  }
  
  /* 移动端头部导航样式 */
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
    z-index: 2100; /* 提高z-index确保显示在最上层 */
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.15);
    color: white;
    font-size: 16px;
  }
  
  .header-back {
    display: flex;
    align-items: center;
    color: #fff !important;
    font-size: 16px;
    cursor: pointer;
    font-weight: 500;
    padding: 10px;
    min-width: 60px;
    min-height: 40px;
    margin-left: -10px;
  }
  
  .header-back i {
    margin-right: 5px;
    font-size: 18px;
    color: #fff !important;
  }
  
  .header-title {
    flex: 1;
    text-align: center;
    margin-right: 20px;
    font-size: 16px;
    font-weight: 500;
  }
  
  /* 确保表单可编辑 */
  .image-processing-form {
    padding: 10px 15px 90px !important; /* 增加底部空间 */
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
    height: 50px; /* 减小高度 */
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
  
  /* 图片容器移动端样式 */
  .mobile-image-container {
    display: flex;
    flex-direction: column;
  }
  
  /* 蒙版编辑器移动端样式 */
  .mobile-mask-editor {
    padding: 0 0 90px 0 !important; /* 增加底部空间 */
  }
  
  .mobile-editor-tools {
    padding: 5px !important;
    flex-wrap: wrap !important;
    justify-content: flex-start !important;
  }
  
  .mobile-editor-tools .tool-group {
    margin: 5px 0 !important;
    width: 100% !important;
  }
  
  .mobile-canvas-container {
    height: calc(100vh - 180px) !important;
  }
}

/* 蒙版编辑器头部样式 */
.mask-editor-header {
  background-color: #67C23A !important; /* 使用绿色以示区分 */
}

/* 增强返回按钮的触摸区域 */
.header-back {
  display: flex;
  align-items: center;
  color: #fff !important;
  font-size: 16px;
  cursor: pointer;
  font-weight: 500;
  padding: 10px;
  min-width: 60px;
  min-height: 40px;
  margin-left: -10px;
}

/* 覆盖Element UI轮播图样式 */
.task-card .el-carousel__indicators {
  bottom: 4px !important;
}

.task-card .el-carousel__button {
  width: 6px !important;
  height: 6px !important;
  background-color: rgba(255, 255, 255, 0.7) !important;
}

.task-card .el-carousel__indicator.is-active .el-carousel__button {
  background-color: #409EFF !important;
}

.task-card .el-carousel__arrow {
  width: 24px !important;
  height: 24px !important;
  font-size: 12px !important;
  background-color: rgba(0, 0, 0, 0.3) !important;
}

.task-card .el-carousel__arrow:hover {
  background-color: rgba(0, 0, 0, 0.5) !important;
}

.task-card .el-carousel__arrow--left {
  left: 5px !important;
}

.task-card .el-carousel__arrow--right {
  right: 5px !important;
}

/* 移动端样式优化 */
@media screen and (max-width: 768px) {
  .task-card .el-carousel__indicators {
    bottom: 2px !important;
  }
  
  .task-card .el-carousel__arrow {
    width: 20px !important;
    height: 20px !important;
    font-size: 10px !important;
  }
  
  /* 移动端轮播图高度调整 */
  .task-card .el-carousel {
    height: 160px !important;
  }
  
  /* 确保移动端可以触摸滑动 */
  .task-card .el-carousel__container {
    touch-action: pan-y !important;
  }
  
  /* 隐藏卡片效果的多余内边距 */
  .task-card .el-carousel__item {
    padding: 0 !important;
  }
}

/* 强化轮播图触摸滑动 */
.el-carousel {
  --webkit-user-select: none;
  user-select: none;
}

.el-carousel__mask {
  display: none !important; /* 移除遮罩，优化触摸体验 */
}

/* 确保轮播图容器样式避免边框显示不全 */
.task-card .el-carousel__container {
  padding: 1px;
  box-sizing: border-box;
  touch-action: pan-x !important;
}

/* 移动端触摸优化 */
@media screen and (max-width: 768px) {
  /* 优化轮播图点击区域 */
  .el-carousel__arrow {
    width: 28px !important; /* 增大点击区域 */
    height: 28px !important;
  }
  
  /* 确保轮播图不会阻止页面滚动 */
  .card-list {
    touch-action: pan-y !important;
  }
  
  /* 隐藏桌面端样式按钮，使用原生按钮样式 */
  .task-card .el-carousel__arrow i {
    font-size: 14px !important;
  }
}

/* 优化轮播图移动端触摸体验 */
.task-card-carousel {
  width: 100%;
  box-sizing: border-box;
  touch-action: pan-x !important;
  -webkit-user-select: none;
  user-select: none;
  overflow: visible !important;
}

.carousel-item {
  touch-action: pan-x !important;
  -webkit-user-select: none;
  user-select: none;
  cursor: grab;
}

.carousel-item:active {
  cursor: grabbing;
}

/* 确保图片不会阻止滑动事件 */
.carousel-image {
  pointer-events: none;
  user-select: none;
  -webkit-user-drag: none;
}

/* 增强移动端触摸体验 */
@media (max-width: 768px) {
  .el-carousel__container {
    touch-action: pan-x !important;
  }
  
  .el-carousel__item {
    touch-action: pan-x !important;
  }
  
  /* 增大点击区域 */
  .el-carousel__arrow {
    width: 32px !important;
    height: 32px !important;
    background-color: rgba(0, 0, 0, 0.4) !important;
  }
  
  .el-carousel__arrow i {
    font-size: 18px !important;
  }
}
</style>