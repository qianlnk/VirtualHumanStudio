<template>
  <div class="workflow-container">
    <div class="page-header">
      <h2>{{ currentModule ? currentModule.name : '图像处理' }}</h2>
      <el-button type="primary" @click="showCreateDialog" :disabled="!currentModule">创建{{currentModule ? currentModule.name : ''}}任务</el-button>
    </div>

    <!-- 任务列表 -->
    <el-card class="task-list">
      <el-table
        v-loading="loading"
        :data="tasks"
        style="width: 100%"
        :empty-text="'暂无数据'"
      >
        <el-table-column prop="name" label="任务名称" width="400" show-overflow-tooltip></el-table-column>
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
            <el-button
              type="text"
              size="small"
              @click="viewDetail(scope.row.id)">查看</el-button>
            <el-button
              type="text"
              size="small"
              @click="deleteTask(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

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
        <template v-for="(param, index) in currentModule.inputParams">
          <el-form-item :key="index" :label="param.alias" :prop="`params.${param.key}`">
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
        </template>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">创建</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getImageProcessingModules, createImageProcessingTask, getImageProcessingTasks, deleteImageProcessingTask, getImageProcessingTaskDetail } from '@/utils/imageProcessingApi'

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
      lastY: 0
    }
  },
  created() {
    console.log('组件创建，当前路由参数:', this.$route.params)
    const moduleId = this.moduleId || this.$route.params.moduleId
    if (!moduleId) {
      this.showModuleList = true
      this.fetchModules()
    } else {
      console.log('初始化模块ID:', moduleId)
    }
    // 模块的初始化和任务列表获取由watch处理
  },
  watch: {
    '$route.params.moduleId': {
      immediate: true,
      deep: true,
      async handler() {
          console.log('重新初始化模块数据')
          // 重置状态
          this.tasks = []
          this.loading = true
          this.currentModule = null
          this.form = {
            taskName: '',
            params: {}
          }
          this.previewUrls = {}
          
          try {
            const moduleInitialized = await this.initModule()
            if (moduleInitialized && this.currentModule) {
              await this.fetchTasks()
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
  methods: {
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
          await this.fetchModules()
        }
        // 根据路由参数找到当前模块
        const moduleId = this.moduleId || this.$route.params.moduleId
        if (moduleId) {
          // 如果有moduleId参数，查找对应模块
          this.currentModule = this.modules.find(m => m.id === moduleId)
          if (this.currentModule) {
            // 初始化表单验证规则
            this.initFormRules()
            this.showModuleList = false
            return true
          } else {
            this.$message.error('未找到指定的处理模块，即将返回模块列表页面')
            // 重定向到模块列表页面
            this.$router.push('/home')
            return false
          }
        }
        return false
      } catch (error) {
        this.$message.error('初始化模块失败：' + error.message)
        return false
      } finally {
        this.moduleLoading = false
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
    async fetchTasks() {
      if (!this.currentModule) return
      
      this.loading = true
      try {
        const response = await getImageProcessingTasks(this.currentModule.id)
        if (response.success) {
          this.tasks = response.tasks
        }
      } catch (error) {
        this.$message.error('获取任务列表失败：' + error.message)
      } finally {
        this.loading = false
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
    }
  }
}
</script>

<style scoped>
.workflow-container {
  padding: 40px;
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #fff;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.task-list {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 20px;
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.preview-container {
  text-align: center;
}

.preview-image {
  max-width: 100%;
  max-height: 500px;
}

.upload-item {
  display: inline-block;
}

.page-header h2 {
  font-size: 2em;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.image-preview {
  margin-top: 10px;
}

.preview-thumbnail {
  max-width: 200px;
  max-height: 200px;
  border-radius: 4px;
}

.module-list {
  margin-top: 20px;
}

.module-card {
  cursor: pointer;
  transition: all 0.3s;
  height: 100%;
}

.module-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 2px 12px 0 rgba(0,0,0,.1);
}

.module-card h3 {
  margin: 0 0 10px 0;
}

.module-card p {
  color: #666;
  margin: 0;
  font-size: 14px;
}

.mask-editor-container {
  padding: 20px;
}

.editor-tools {
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  align-items: center;
}

.tool-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tool-label {
  font-weight: bold;
  min-width: 70px;
}

.canvas-container {
  position: relative;
  margin: 20px 0;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  width: 100%;
  min-height: 400px;
  background-color: #f5f5f5;
  height: 600px;
}

.canvas-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.editor-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: transparent;
}

.base-canvas {
  z-index: 1;
}

.draw-canvas {
  z-index: 2;
}

.editor-actions {
  margin-top: 20px;
  text-align: right;
}
</style>

<style scoped>
.workflow-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.task-list {
  margin-bottom: 20px;
}

.preview-container {
  text-align: center;
}

.preview-image {
  max-width: 100%;
  max-height: 500px;
}

.upload-item {
  margin-bottom: 10px;
}

.image-preview {
  margin-top: 10px;
}

.preview-thumbnail {
  max-width: 200px;
  max-height: 200px;
  border-radius: 4px;
}

.module-list {
  margin-top: 20px;
}

.module-card {
  cursor: pointer;
  transition: all 0.3s;
  height: 100%;
}

.module-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 2px 12px 0 rgba(0,0,0,.1);
}

.module-card h3 {
  margin: 0 0 10px 0;
}

.module-card p {
  color: #666;
  margin: 0;
  font-size: 14px;
}

.mask-editor-container {
  padding: 20px;
}

.editor-tools {
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  align-items: center;
}

.tool-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tool-label {
  font-weight: bold;
  min-width: 70px;
}

.canvas-container {
  position: relative;
  margin: 20px 0;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  width: 100%;
  min-height: 400px;
  background-color: #f5f5f5;
  height: 600px;
}

.canvas-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.editor-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: transparent;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.base-canvas {
  z-index: 1;
}

.draw-canvas {
  z-index: 2;
}

.editor-actions {
  margin-top: 20px;
  text-align: right;
}

.brush-size-label {
  margin-left: 10px;
  color: #666;
  font-size: 14px;
}

/* 任务详情样式 */
.task-detail-dialog {
  max-width: 800px !important;
}

.task-detail {
  padding: 20px;
  font-size: 14px;
  line-height: 1.6;
}

.task-detail p {
  margin: 10px 0;
}

.params-list {
  margin: 10px 0;
  padding: 10px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.param-item {
  margin: 10px 0;
}

.param-image {
  max-width: 200px;
  max-height: 200px;
  margin-top: 5px;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
}

.result-preview {
  margin: 10px 0;
}

.output-item {
  margin: 10px 0;
}

.output-image {
  max-width: 100%;
  max-height: 400px;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
}

.error-msg {
  color: #f56c6c;
  margin: 10px 0;
  padding: 10px;
  background-color: #fef0f0;
  border-radius: 4px;
}
</style>