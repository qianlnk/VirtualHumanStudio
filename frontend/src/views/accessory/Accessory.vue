<template>
  <div class="accessory-container">
    <div class="page-header">
      <h2>饰品替换</h2>
      <el-button type="primary" @click="showCreateDialog">创建饰品替换任务</el-button>
    </div>

    <!-- 任务列表 -->
    <el-card class="task-list">
      <el-table
        v-loading="loading"
        :data="tasks"
        style="width: 100%"
        :empty-text="'暂无数据'"
      >
        <el-table-column prop="name" label="任务名称" min-width="400" show-overflow-tooltip></el-table-column>
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
              @click="viewResult(scope.row)"
              :disabled="scope.row.status !== 'completed'">查看结果</el-button>
            <el-button
              type="text"
              size="small"
              @click="deleteTask(scope.row)">删除</el-button>
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

    <!-- 结果预览对话框 -->
    <el-dialog title="替换结果" :visible.sync="previewDialogVisible" width="50%">
      <div class="preview-container" v-if="currentResult">
        <img :src="currentResult" class="preview-image" alt="替换结果">
      </div>
    </el-dialog>
    
    <!-- 创建饰品替换任务对话框 -->
    <el-dialog title="创建饰品替换任务" :visible.sync="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="form" label-width="120px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称"></el-input>
        </el-form-item>
        <el-form-item label="白底图片" prop="whiteImage">
          <el-upload
            class="upload-item"
            action="#"
            :auto-upload="false"
            :on-change="handleWhiteImageChange"
            :show-file-list="true"
            accept="image/*"
            :limit="1">
            <el-button size="small" type="primary">选择图片</el-button>
            <div slot="tip" class="el-upload__tip">请上传白底图片</div>
          </el-upload>
          <div class="image-preview" v-if="previewUrls.whiteImage">
            <img :src="previewUrls.whiteImage" class="preview-thumbnail" alt="白底图片预览">
          </div>
        </el-form-item>

        <el-form-item label="模特物品图片" prop="modelImage">
          <el-upload
            class="upload-item"
            action="#"
            :auto-upload="false"
            :on-change="handleModelImageChange"
            :show-file-list="true"
            accept="image/*"
            :limit="1">
            <el-button size="small" type="primary">选择图片</el-button>
            <div slot="tip" class="el-upload__tip">请上传模特物品图片</div>
          </el-upload>
          <div class="image-preview" v-if="previewUrls.modelImage">
            <img :src="previewUrls.modelImage" class="preview-thumbnail" alt="模特物品图片预览">
          </div>
        </el-form-item>

        <el-form-item label="蒙版图片" prop="maskImage">
          <div class="mask-options">
            <el-upload
              class="upload-item"
              action="#"
              :auto-upload="false"
              :on-change="handleMaskImageChange"
              :show-file-list="true"
              accept="image/*"
              :limit="1">
              <el-button size="small" type="primary">选择图片</el-button>
              <div slot="tip" class="el-upload__tip">请上传蒙版图片或使用下方编辑器创建</div>
            </el-upload>
            <el-button size="small" type="success" @click="startMaskEditing" :disabled="!form.modelImage">编辑蒙版</el-button>
          </div>
          <div class="image-preview" v-if="previewUrls.maskImage">
            <img :src="previewUrls.maskImage" class="preview-thumbnail" alt="蒙版图片预览">
          </div>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">创建</el-button>
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
            <canvas ref="drawCanvas" class="editor-canvas draw-canvas"></canvas>
          </div>
        </div>
        
        <div class="editor-actions">
          <el-button type="primary" @click="saveMask">保存蒙版</el-button>
          <el-button @click="closeMaskEditor">取消</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import axios from 'axios'
import { getImageUrl } from '@/utils/fileAccess'

export default {
  name: 'Accessory',
  watch: {
    // 监听笔刷属性变化，更新鼠标样式
    brushMode() {
      this.updateCursor(this.$refs.drawCanvas)
    },
    brushShape() {
      this.updateCursor(this.$refs.drawCanvas)
    },
    brushSize() {
      this.updateCursor(this.$refs.drawCanvas)
    }
  },
  data() {
    return {
      loading: false,
      submitting: false,
      dialogVisible: false,
      maskEditorVisible: false,
      form: {
        name: '',
        whiteImage: null,
        modelImage: null,
        maskImage: null
      },
      previewUrls: {
        whiteImage: '',
        modelImage: '',
        maskImage: ''
      },
      rules: {
        name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
        whiteImage: [{ required: true, message: '请上传白底图片', trigger: 'change' }],
        modelImage: [{ required: true, message: '请上传模特物品图片', trigger: 'change' }],
        maskImage: [{ required: true, message: '请上传蒙版图片', trigger: 'change' }]
      },
      // 蒙版编辑器相关数据
      brushMode: 'brush',
      brushSize: 10,
      brushShape: 'round', // 笔刷形状：round(圆形)、square(方形)
      brushColor: 'black', // 遮照层颜色：black(黑色)、white(白色)、gray(灰色)
      isDrawing: false,
      lastX: 0,
      lastY: 0,
      canvasWidth: 0, // 将根据模特图片尺寸动态设置
      canvasHeight: 0, // 将根据模特图片尺寸动态设置
      modelImageSize: { width: 0, height: 0 }, // 存储模特图片的原始尺寸
      tasks: [],
      previewDialogVisible: false,
      currentResult: null,
      currentPage: 1,
      pageSize: 10,
      total: 0
    }
  },
  created() {
    this.fetchTasks()
  },
  methods: {
    // 显示创建对话框
    showCreateDialog() {
      this.dialogVisible = true
      this.resetForm()
    },
    
    handleWhiteImageChange(file) {
      this.form.whiteImage = file.raw
      this.createImagePreview(file.raw, 'whiteImage')
    },
    handleModelImageChange(file) {
      this.form.modelImage = file.raw
      this.createImagePreview(file.raw, 'modelImage')
      
      // 获取并存储模特图片的原始尺寸
      const img = new Image();
      img.onload = () => {
        this.modelImageSize = {
          width: img.width,
          height: img.height
        };
      };
      img.src = URL.createObjectURL(file.raw);
    },
    handleMaskImageChange(file) {
      // 检查蒙版图片尺寸是否与模特图片一致
      if (this.modelImageSize.width > 0 && this.modelImageSize.height > 0) {
        const img = new Image();
        img.onload = () => {
          if (img.width !== this.modelImageSize.width || img.height !== this.modelImageSize.height) {
            this.$message.warning(`蒙版图片尺寸(${img.width}x${img.height})与模特图片尺寸(${this.modelImageSize.width}x${this.modelImageSize.height})不一致，建议使用相同尺寸的图片`);
          }
          this.form.maskImage = file.raw
          this.createImagePreview(file.raw, 'maskImage')
        };
        img.src = URL.createObjectURL(file.raw);
      } else {
        this.form.maskImage = file.raw
        this.createImagePreview(file.raw, 'maskImage')
      }
    },
    
    // 创建图片预览
    createImagePreview(file, type) {
      if (!file) return
      
      const reader = new FileReader()
      reader.onload = (e) => {
        this.previewUrls[type] = e.target.result
      }
      reader.readAsDataURL(file)
    },
    
    // 开始蒙版编辑
    startMaskEditing() {
      if (!this.form.modelImage) {
        this.$message.warning('请先上传模特物品图片')
        return
      }
      
      this.maskEditorVisible = true
      this.$nextTick(() => {
        this.initCanvas()
      })
    },
    
    // 初始化画布
    async initCanvas() {
      // 获取canvas元素
      const baseCanvas = this.$refs.baseCanvas
      const drawCanvas = this.$refs.drawCanvas
      
      if (!baseCanvas || !drawCanvas) {
        console.error('Canvas元素未找到')
        return
      }
      
      // 使用模特图片的实际尺寸设置画布大小
      if (this.modelImageSize.width > 0 && this.modelImageSize.height > 0) {
        // 使用模特图片的实际尺寸
        this.canvasWidth = this.modelImageSize.width
        this.canvasHeight = this.modelImageSize.height
      } else {
        // 如果没有模特图片尺寸信息，使用默认尺寸
        this.canvasWidth = 800
        this.canvasHeight = 600
      }
      
      console.log('设置画布尺寸:', this.canvasWidth, 'x', this.canvasHeight)
      
      // 获取canvas的父容器
      const canvasWrapper = baseCanvas.parentElement
      if (canvasWrapper) {
        // 设置父容器的尺寸，确保能够容纳canvas
        canvasWrapper.style.width = `${this.canvasWidth}px`
        canvasWrapper.style.height = `${this.canvasHeight}px`
        canvasWrapper.style.position = 'relative' // 确保父容器使用相对定位
      }
      
      // 设置画布大小
      baseCanvas.width = this.canvasWidth
      baseCanvas.height = this.canvasHeight
      drawCanvas.width = this.canvasWidth
      drawCanvas.height = this.canvasHeight
      
      // 设置canvas的样式，确保它们能够正确叠加
      baseCanvas.style.position = 'absolute'
      baseCanvas.style.top = '0'
      baseCanvas.style.left = '0'
      baseCanvas.style.zIndex = '1'
      
      drawCanvas.style.position = 'absolute'
      drawCanvas.style.top = '0'
      drawCanvas.style.left = '0'
      drawCanvas.style.zIndex = '2'
      
      const baseCtx = baseCanvas.getContext('2d')
      const drawCtx = drawCanvas.getContext('2d')
      
      // 添加绘制事件监听
      this.setupDrawingEvents(drawCanvas)
      
      // 加载模特图片到基础画布
      if (this.previewUrls.modelImage) {
        try {
          console.log('开始加载模特图片:', this.previewUrls.modelImage)
          
          // 创建图片对象并加载
          const loadImage = (src) => {
            return new Promise((resolve, reject) => {
              const img = new Image()
              img.crossOrigin = 'Anonymous' // 允许跨域加载图片
              img.onload = () => {
                console.log('图片加载成功:', img.width, 'x', img.height)
                resolve(img)
              }
              img.onerror = (e) => {
                console.error('图片加载失败:', e)
                reject(new Error('图片加载失败'))
              }
              img.src = src
            })
          }
          
          // 加载模特图片
          const img = await loadImage(this.previewUrls.modelImage)
          
          // 清空基础画布
          baseCtx.clearRect(0, 0, this.canvasWidth, this.canvasHeight)
          
          // 填充白色背景，确保透明图片可见
          baseCtx.fillStyle = 'white'
          baseCtx.fillRect(0, 0, this.canvasWidth, this.canvasHeight)
          
          // 保持图片原始比例绘制到画布上
          baseCtx.drawImage(img, 0, 0, img.width, img.height, 0, 0, this.canvasWidth, this.canvasHeight)
          console.log('模特图片已绘制到画布')
          
          // 如果已有蒙版图片，加载到绘制画布
          if (this.previewUrls.maskImage) {
            try {
              const maskImg = await loadImage(this.previewUrls.maskImage)
              // 清空绘制画布
              drawCtx.clearRect(0, 0, this.canvasWidth, this.canvasHeight)
              // 保持蒙版图片原始比例绘制到画布上
              drawCtx.drawImage(maskImg, 0, 0, maskImg.width, maskImg.height, 0, 0, this.canvasWidth, this.canvasHeight)
            } catch (error) {
              console.error('蒙版图片加载失败:', error)
              this.$message.warning('蒙版图片加载失败')
              // 清空绘制画布
              drawCtx.clearRect(0, 0, this.canvasWidth, this.canvasHeight)
            }
          } else {
            // 清空绘制画布
            drawCtx.clearRect(0, 0, this.canvasWidth, this.canvasHeight)
          }
        } catch (error) {
          console.error('模特图片加载失败:', error)
          this.$message.error('模特图片加载失败，请重试')
        }
      } else {
        console.warn('没有模特图片可加载')
        // 显示提示信息
        baseCtx.fillStyle = 'white'
        baseCtx.fillRect(0, 0, this.canvasWidth, this.canvasHeight)
        baseCtx.font = '16px Arial'
        baseCtx.fillStyle = 'black'
        baseCtx.textAlign = 'center'
        baseCtx.fillText('请先上传模特图片', this.canvasWidth / 2, this.canvasHeight / 2)
      }
    },
    
    // 设置绘制事件
    setupDrawingEvents(canvas) {
      // 初始化鼠标样式
      this.updateCursor(canvas)
      
      // 鼠标按下事件
      canvas.addEventListener('mousedown', (e) => {
        this.isDrawing = true
        const rect = canvas.getBoundingClientRect()
        this.lastX = e.clientX - rect.left
        this.lastY = e.clientY - rect.top
      })
      
      // 鼠标移动事件
      canvas.addEventListener('mousemove', (e) => {
        if (!this.isDrawing) return
        
        const rect = canvas.getBoundingClientRect()
        const currentX = e.clientX - rect.left
        const currentY = e.clientY - rect.top
        
        this.draw(canvas, this.lastX, this.lastY, currentX, currentY)
        
        this.lastX = currentX
        this.lastY = currentY
      })
      
      // 鼠标释放事件
      canvas.addEventListener('mouseup', () => {
        this.isDrawing = false
      })
      
      // 鼠标离开事件
      canvas.addEventListener('mouseleave', () => {
        this.isDrawing = false
      })
    },
    
    // 更新鼠标样式
    updateCursor(canvas) {
      if (!canvas) return
      
      let cursorStyle = ''
      
      // 根据笔刷形状和模式设置鼠标样式
      if (this.brushMode === 'eraser') {
        // 橡皮擦模式
        cursorStyle = `url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="${this.brushSize}" height="${this.brushSize}" viewBox="0 0 ${this.brushSize} ${this.brushSize}"><circle cx="${this.brushSize/2}" cy="${this.brushSize/2}" r="${this.brushSize/2}" fill="rgba(255,0,0,0.5)" stroke="black" stroke-width="1"/></svg>') ${this.brushSize/2} ${this.brushSize/2}, auto`
      } else if (this.brushShape === 'round') {
        // 圆形笔刷
        const fillColor = this.brushColor === 'black' ? 'rgba(0,0,0,0.5)' : this.brushColor === 'white' ? 'rgba(255,255,255,0.5)' : 'rgba(128,128,128,0.5)';
        cursorStyle = `url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="${this.brushSize}" height="${this.brushSize}" viewBox="0 0 ${this.brushSize} ${this.brushSize}"><circle cx="${this.brushSize/2}" cy="${this.brushSize/2}" r="${this.brushSize/2}" fill="${fillColor}" stroke="black" stroke-width="1"/></svg>') ${this.brushSize/2} ${this.brushSize/2}, auto`
      } else {
        // 方形笔刷
        const fillColor = this.brushColor === 'black' ? 'rgba(0,0,0,0.5)' : this.brushColor === 'white' ? 'rgba(255,255,255,0.5)' : 'rgba(128,128,128,0.5)';
        cursorStyle = `url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="${this.brushSize}" height="${this.brushSize}" viewBox="0 0 ${this.brushSize} ${this.brushSize}"><rect x="0" y="0" width="${this.brushSize}" height="${this.brushSize}" fill="${fillColor}" stroke="black" stroke-width="1"/></svg>') ${this.brushSize/2} ${this.brushSize/2}, auto`
      }
      
      canvas.style.cursor = cursorStyle
    },
    
    // 绘制函数
    draw(canvas, startX, startY, endX, endY) {
      const ctx = canvas.getContext('2d')
      
      // 设置笔刷形状
      ctx.lineWidth = this.brushSize
      ctx.lineCap = this.brushShape === 'round' ? 'round' : 'square'
      ctx.lineJoin = this.brushShape === 'round' ? 'round' : 'miter'
      
      if (this.brushMode === 'brush') {
        // 根据选择的颜色设置笔刷颜色
        switch(this.brushColor) {
          case 'white':
            ctx.strokeStyle = '#ffffff'
            break
          case 'black':
            ctx.strokeStyle = '#000000'
            break
          case 'gray':
            ctx.strokeStyle = '#808080'
            break
          default:
            ctx.strokeStyle = '#000000'
        }
        ctx.globalCompositeOperation = 'source-over'
      } else {
        // 橡皮擦模式
        ctx.strokeStyle = 'black'
        ctx.globalCompositeOperation = 'destination-out'
      }
      
      // 如果是方形笔刷且不是橡皮擦模式，使用矩形绘制
      if (this.brushShape === 'square' && this.brushMode === 'brush') {
        // 计算矩形的宽度和高度
        const width = Math.abs(endX - startX)
        const height = Math.abs(endY - startY)
        
        // 如果是点击而不是拖动，绘制一个正方形
        if (width < 2 && height < 2) {
          const halfSize = this.brushSize / 2
          ctx.fillStyle = ctx.strokeStyle
          ctx.fillRect(startX - halfSize, startY - halfSize, this.brushSize, this.brushSize)
        } else {
          // 否则绘制线条
          ctx.beginPath()
          ctx.moveTo(startX, startY)
          ctx.lineTo(endX, endY)
          ctx.stroke()
        }
      } else {
        // 圆形笔刷或橡皮擦使用线条绘制
        ctx.beginPath()
        ctx.moveTo(startX, startY)
        ctx.lineTo(endX, endY)
        ctx.stroke()
      }
    },
    
    // 清空画布
    clearCanvas() {
      const drawCanvas = this.$refs.drawCanvas
      if (!drawCanvas) return
      
      const ctx = drawCanvas.getContext('2d')
      ctx.clearRect(0, 0, this.canvasWidth, this.canvasHeight)
    },
    
    // 保存蒙版
    saveMask() {
      const baseCanvas = this.$refs.baseCanvas
      const drawCanvas = this.$refs.drawCanvas
      if (!drawCanvas || !baseCanvas) {
        console.error('画布元素未找到')
        return
      }
      
      console.log('开始合并图层 - 画布尺寸:', this.canvasWidth, 'x', this.canvasHeight)
      console.log('模特图片尺寸:', this.modelImageSize.width, 'x', this.modelImageSize.height)
      
      try {
        // 创建一个临时画布用于合并模特图层和蒙版图层
        const mergeCanvas = document.createElement('canvas')
        mergeCanvas.width = this.canvasWidth
        mergeCanvas.height = this.canvasHeight
        const mergeCtx = mergeCanvas.getContext('2d')
        
        // 清空合并画布并填充白色背景
        mergeCtx.fillStyle = 'white'
        mergeCtx.fillRect(0, 0, mergeCanvas.width, mergeCanvas.height)
        
        // 首先绘制模特图层（baseCanvas）
        console.log('绘制模特图层')
        mergeCtx.drawImage(baseCanvas, 0, 0)
        
        // 然后绘制蒙版图层（drawCanvas），使用destination-out模式使蒙版部分变为透明
        console.log('绘制蒙版图层，使用destination-out模式')
        mergeCtx.globalCompositeOperation = 'destination-out' // 设置合成模式，使绘制的地方变为透明
        mergeCtx.drawImage(drawCanvas, 0, 0)
        mergeCtx.globalCompositeOperation = 'source-over' // 恢复默认合成模式
        
        // 检查合并后的画布是否有内容
        const imageData = mergeCtx.getImageData(0, 0, mergeCanvas.width, mergeCanvas.height)
        console.log('合并后画布数据大小:', imageData.data.length, '字节')
        
        let maskDataUrl = ''
        
        // 确保画布尺寸与模特图片尺寸一致
        if (this.modelImageSize.width > 0 && this.modelImageSize.height > 0 && 
            (mergeCanvas.width !== this.modelImageSize.width || mergeCanvas.height !== this.modelImageSize.height)) {
          console.log('画布尺寸与模特图片尺寸不一致，进行转换')
          // 如果画布尺寸与模特图片尺寸不一致，创建一个临时画布进行转换
          const tempCanvas = document.createElement('canvas')
          tempCanvas.width = this.modelImageSize.width
          tempCanvas.height = this.modelImageSize.height
          const tempCtx = tempCanvas.getContext('2d')
          
          // 清空临时画布并填充白色背景
          tempCtx.fillStyle = 'white'
          tempCtx.fillRect(0, 0, tempCanvas.width, tempCanvas.height)
          
          // 绘制合并后的画布内容到临时画布
          tempCtx.drawImage(mergeCanvas, 0, 0, mergeCanvas.width, mergeCanvas.height, 0, 0, this.modelImageSize.width, this.modelImageSize.height)
          
          // 将临时画布内容转换为图片，使用高质量输出
          maskDataUrl = tempCanvas.toDataURL('image/png', 1.0)
        } else {
          console.log('画布尺寸与模特图片尺寸一致，直接使用合并画布')
          // 将合并后的画布内容转换为图片，使用高质量输出
          maskDataUrl = mergeCanvas.toDataURL('image/png', 1.0)
        }
        
        console.log('已成功合并模特图层和蒙版图层，DataURL长度:', maskDataUrl.length)
        this.previewUrls.maskImage = maskDataUrl
        
        // 检查生成的DataURL是否包含有效数据
        if (!maskDataUrl || maskDataUrl === 'data:,' || maskDataUrl.length < 100) {
          console.error('生成的蒙版图片数据为空或异常')
          this.$message.error('生成蒙版图片失败，请重试')
          return
        }
        
        // 将DataURL转换为Blob
        fetch(maskDataUrl)
          .then(res => res.blob())
          .then(blob => {
            // 检查Blob大小
            console.log('蒙版图片Blob大小:', blob.size, '字节')
            if (blob.size <= 100) { // 如果Blob太小，可能是空图片
              console.error('生成的蒙版图片数据异常，大小过小:', blob.size, '字节')
              this.$message.error('生成蒙版图片异常，请重试')
              return
            }
            
            // 创建File对象
            const file = new File([blob], 'mask.png', { type: 'image/png' })
            this.form.maskImage = file
            console.log('已将蒙版图片设置到表单中，文件大小:', file.size, '字节')
            
            // 确保预览URL已设置
            this.$nextTick(() => {
              // 强制更新预览
              this.$forceUpdate()
              console.log('已强制更新预览')
            })
            this.$message.success('蒙版已保存')
            this.maskEditorVisible = false
          })
          .catch(err => {
            console.error('保存蒙版失败:', err)
            this.$message.error('保存蒙版失败')
          })
      } catch (error) {
        console.error('蒙版处理过程中发生错误:', error)
        this.$message.error('蒙版处理失败: ' + error.message)
      }
    },
    
    // 关闭蒙版编辑器
    closeMaskEditor() {
      this.$confirm('关闭编辑器将丢失未保存的更改，是否继续?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.maskEditorVisible = false
      }).catch(() => {})
    },
    async submitForm() {
      try {
        await this.$refs.form.validate()
        this.submitting = true
        
        const formData = new FormData()
        formData.append('name', this.form.name)
        formData.append('white_image', this.form.whiteImage)
        formData.append('model_image', this.form.modelImage)
        formData.append('mask_image', this.form.maskImage)

        const response = await axios.post('/api/accessory', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
        
        if (response.data) {
          this.$message.success('任务创建成功')
          this.dialogVisible = false
          this.fetchTasks()
          this.resetForm()
        }
      } catch (error) {
        const errorMessage = (error.response && error.response.data && error.response.data.error) || error.message || '创建任务失败'
        this.$message.error(errorMessage)
      } finally {
        this.submitting = false
      }
    },
    resetForm() {
      if (this.$refs.form) {
        this.$refs.form.resetFields()
      }
      this.form = {
        name: '',
        whiteImage: null,
        modelImage: null,
        maskImage: null
      }
      this.previewUrls = {
        whiteImage: '',
        modelImage: '',
        maskImage: ''
      }
    },
    async fetchTasks() {
      this.loading = true
      try {
        const response = await axios.get('/api/accessory', {
          params: {
            page: this.currentPage,
            page_size: this.pageSize
          }
        })
        this.tasks = response.data.accessories || []
        this.total = response.data.total || 0
      } catch (error) {
        this.$message.error('获取任务列表失败')
      } finally {
        this.loading = false
      }
    },
    getStatusType(status) {
      const statusMap = {
        pending: 'info',
        processing: 'warning',
        completed: 'success',
        failed: 'danger'
      }
      return statusMap[status] || 'info'
    },
    getStatusText(status) {
      const statusMap = {
        pending: '等待中',
        processing: '处理中',
        completed: '已完成',
        failed: '失败'
      }
      return statusMap[status] || status
    },
    // 查看详情页
    viewDetail(id) {
      this.$router.push(`/accessory/${id}`)
    },
    
    async viewResult(task) {
      try {
        const response = await axios.get(`/api/accessory/${task.id}`)
        // 使用getImageUrl函数获取图片URL，确保添加认证token
        this.currentResult = await getImageUrl(response.data.output_image)
        this.previewDialogVisible = true
      } catch (error) {
        this.$message.error('获取结果失败')
      }
    },
    async deleteTask(task) {
      try {
        await this.$confirm('确认删除该任务？')
        await axios.delete(`/api/accessory/${task.id}`)
        this.$message.success('删除成功')
        this.fetchTasks()
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('删除失败')
        }
      }
    },

    // 处理每页显示数量变化
    handleSizeChange(val) {
      this.pageSize = val
      this.fetchTasks()
    },

    // 处理页码变化
    handleCurrentChange(val) {
      this.currentPage = val
      this.fetchTasks()
    },

    // 格式化日期
    formatDate(date) {
      if (!date) return ''
      const d = new Date(date)
      return d.toLocaleString()
    }
  }
}
</script>

<style scoped>
.accessory-container {
  padding: 40px;
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #fff;
}

.upload-item {
  display: inline-block;
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

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.page-header h2 {
  font-size: 2em;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
  color: #fff;
}

.clearfix::after {
  content: '';
  display: table;
  clear: both;
}

.dialog-footer {
  text-align: right;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

/* 图片预览样式 */
.image-preview {
  margin-top: 10px;
  border: 1px dashed #ccc;
  padding: 5px;
  border-radius: 4px;
  background-color: rgba(255, 255, 255, 0.1);
  text-align: center;
}

.preview-thumbnail {
  max-width: 100%;
  max-height: 150px;
  object-fit: contain;
}

/* 蒙版选项样式 */
.mask-options {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 蒙版编辑器样式 */
.mask-editor-container {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.editor-tools {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.brush-size-label {
  margin-left: 5px;
  color: #606266;
}

.tool-group {
  display: flex;
  align-items: center;
  margin-right: 15px;
  margin-bottom: 10px;
}

.tool-label {
  margin-right: 8px;
  color: #606266;
  font-size: 14px;
}

.canvas-container {
  position: relative;
  display: flex;
  justify-content: center;
  margin: 10px 0;
  background-color: #eee;
  border-radius: 4px;
  overflow: auto; /* 允许滚动查看大图 */
  max-height: 80vh; /* 限制最大高度，避免超出屏幕 */
  padding: 10px; /* 添加内边距，使画布更容易看到 */
}

.canvas-wrapper {
  position: relative;
  display: inline-block;
  background-color: #f0f0f0;
  border: 1px solid #ddd;
  /* 确保canvas-wrapper有足够的空间显示canvas */
  min-width: 400px;
  min-height: 300px;
}

.editor-canvas {
  position: absolute;
  top: 0;
  left: 0;
  display: block;
}

.base-canvas {
  z-index: 1;
  background-color: white; /* 添加背景色以便于区分 */
}

.draw-canvas {
  z-index: 2;
}

.editor-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 10px;
}
</style>