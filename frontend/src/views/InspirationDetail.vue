<template>
  <div class="inspiration-detail-container" :class="{'mobile-container': isMobile}">
    <!-- 标题 -->
    <div class="page-header" v-show="!isMobile">
      <div class="header-left">
        <i class="el-icon-arrow-left back-icon" @click="goBack">返回</i>
        <h2>{{ task.name }}</h2>
      </div>
      <div class="header-right">
        <!-- <i class="el-icon-refresh refresh-icon" @click="refreshContent" :class="{'is-loading': loading}"></i> -->
      </div>
    </div>


    <!-- 移动端顶部导航栏 - 只在移动端显示 -->
    <div class="mobile-header-bar" v-show="isMobile">
      <div class="header-back" @click="goBack">
        <i class="el-icon-arrow-left"></i>
        <span>返回</span>
      </div>
    </div>

    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>

    <!-- 加载动画 -->
    <div v-if="loading" class="loading-container">
      <div class="loading-animation">
        <i class="el-icon-loading"></i>
        <p>正在加载内容...</p>
      </div>
      <el-skeleton :rows="5" animated />
    </div>

    <!-- 没有内容时显示 -->
    <el-empty v-else-if="!task" description="未找到创作内容">
      <el-button type="primary" size="small" @click="goBack">返回列表</el-button>
    </el-empty>

    <!-- 详情内容 -->
    <div v-else class="task-detail" :class="{'mobile-detail': isMobile}">
      <div class="detail-header">
        <h2 v-if="!isMobile">{{ task.description || task.name }}</h2>
        <p v-if="task.description && !isMobile">{{ task.description }}</p>
      </div>
      
      <!-- 预览区域 -->
      <div class="detail-preview" :class="{'mobile-preview': isMobile}">
        <!-- 数字人视频或者输出参数为视频的任务 -->
        <div v-if="hasVideoOutput || (task.type === 'digital_human')" class="detail-video" :class="{'mobile-video': isMobile}">
          <div class="video-placeholder" v-if="!detailVideoLoaded" @click="loadDetailVideo">
            <div class="loading-indicator">
              <i class="el-icon-video-play"></i>
              <span>点击加载视频</span>
            </div>
          </div>
          <video 
            v-show="detailVideoLoaded"
            ref="detailVideo"
            class="full-video"
            controls 
            preload="none"
            playsinline
            webkit-playsinline
            @click="loadDetailVideo"
            @canplay="handleDetailVideoCanPlay"
            @error="handleDetailVideoError">
            <source :src="detailVideoLoaded ? getDirectVideoUrl(getVideoSource()) : ''" type="video/mp4">
            您的浏览器不支持视频播放
          </video>
        </div>
        
        <!-- 图像处理结果 - 展示所有图片输出 -->
        <div v-else class="detail-image" :class="{'mobile-image': isMobile}" 
             @touchstart="handleTouchStart" 
             @touchmove="handleTouchMove" 
             @touchend="handleTouchEnd">
          <el-carousel 
            :height="carouselHeight" 
            :interval="3000" 
            :arrow="isMobile ? 'never' : 'always'" 
            indicator-position="none" 
            ref="imageCarousel"
            class="detail-carousel"
            @change="handleCarouselChange">
            <!-- 如果有输出参数中的图片，优先展示这些图片 -->
            <template v-if="hasOutputImages">
              <el-carousel-item v-for="(image, index) in outputImages" :key="index" class="carousel-item">
                <img :src="image.value" :alt="image.label || image.key" class="full-image" @click="previewImage(image.value)">
              </el-carousel-item>
            </template>
            <!-- 如果没有输出参数图片，但有result_url，则展示result_url -->
            <el-carousel-item v-else-if="task.result_url" class="carousel-item">
              <img :src="task.result_url" :alt="task.name" class="full-image" @click="previewImage(task.result_url)">
            </el-carousel-item>
            <!-- 都没有时显示占位符 -->
            <el-carousel-item v-else class="carousel-item">
              <div class="no-image-placeholder">
                <i class="el-icon-picture"></i>
                <p>暂无图片</p>
              </div>
            </el-carousel-item>
          </el-carousel>
          
          <!-- 自定义指示器 -->
          <div class="custom-indicators" v-if="hasOutputImages && outputImages.length > 1">
            <span 
              v-for="(item, index) in outputImages" 
              :key="index" 
              :class="['indicator-dot', { active: currentCarouselIndex === index }]"
              @click="setActiveItem(index)">
            </span>
          </div>
        </div>
      </div>
      
      <!-- 输入参数 -->
      <div class="detail-params" v-if="task.input_params && task.input_params.length > 0" :class="{'mobile-params-section': isMobile}">
        <h3>输入参数</h3>
        <div class="params-grid" :class="{'mobile-grid': isMobile}">
          <div v-for="(param, index) in task.input_params" :key="'input-'+index" class="param-item" :class="{'mobile-param': isMobile}">
            <div class="param-label">{{ param.label || param.key }}</div>
            
            <!-- 不同类型参数的展示 -->
            <div v-if="param.type === 'image' || param.type === 'mask'" class="param-image">
              <img :src="param.value" :alt="param.label" class="thumbnail" @click="previewImage(param.value)">
            </div>
            <div v-else-if="param.type === 'video'" class="param-video">
              <div class="video-placeholder" v-if="!paramVideoLoadedMap[param.value]" @click="loadParamVideo($event, param.value, param.key || param.label)">
                <div class="loading-indicator small">
                  <i class="el-icon-video-play"></i>
                  <span>点击加载</span>
                </div>
              </div>
              <video 
                v-show="paramVideoLoadedMap[param.value]"
                :data-url="param.value"
                :id="'video-' + (param.key || 'input') + '-' + index"
                class="thumbnail-video" 
                controls
                playsinline
                webkit-playsinline
                preload="none"
                @click="loadParamVideo($event, param.value, param.key || param.label)"
                @canplay="handleParamVideoCanPlay(param.value)"
                @error="handleParamVideoError($event, param.value)">
                <source :src="paramVideoLoadedMap[param.value] ? getDirectVideoUrl(param.value) : ''" type="video/mp4">
                您的浏览器不支持视频播放
              </video>
            </div>
            <div v-else-if="param.type === 'audio'" class="param-audio">
              <audio :src="param.value" controls class="audio-player" preload="none"></audio>
            </div>
            <div v-else class="param-text">
              {{ param.value }}
            </div>
          </div>
        </div>
      </div>
      
      <!-- 输出参数单独显示 -->
      <div class="detail-params" v-if="task.output_params && task.output_params.length > 0 && showOutputParams" :class="{'mobile-params-section': isMobile}">
        <h3>输出参数</h3>
        <div class="params-grid" :class="{'mobile-grid': isMobile}">
          <div v-for="(param, index) in task.output_params" :key="'output-'+index" class="param-item" :class="{'mobile-param': isMobile, 'output-param': true}">
            <!-- 跳过已经在顶部预览展示的图片和视频 -->
            <template v-if="!(param.type === 'image' || param.type === 'mask' || param.type === 'video')">
              <div class="param-label">{{ param.label || param.key }}</div>
              
              <!-- 不同类型参数的展示 -->
              <div v-if="param.type === 'audio'" class="param-audio">
                <audio :src="param.value" controls class="audio-player" preload="none"></audio>
              </div>
              <div v-else class="param-text">
                {{ param.value }}
              </div>
            </template>
          </div>
        </div>
      </div>
      
      <div class="detail-meta" :class="{'mobile-meta': isMobile}">
        <div class="meta-info">
          <div class="meta-user">
            <i class="el-icon-user"></i>
            <span>{{ task.username }}</span>
          </div>
          <div class="meta-likes">
            <i class="el-icon-star-off" :class="{'is-liked': task.is_liked}" @click="toggleLike"></i>
            <span>{{ task.likes || 0 }}</span>
          </div>
        </div>
        <div class="meta-time">
          <i class="el-icon-time"></i>
          <span>{{ formatDate(task.created_at) }}</span>
        </div>
      </div>

      <!-- 底部固定按钮 - 移动端 -->
      <div class="mobile-fixed-bottom" v-if="isMobile">
        <el-button type="primary" size="medium" @click="createSimilar" class="create-similar-btn">
          <i class="el-icon-magic-stick"></i>
          画同款
        </el-button>
      </div>

      <!-- PC端按钮 -->
      <div class="meta-actions" v-if="!isMobile">
        <el-button type="primary" size="small" @click="createSimilar">
          <i class="el-icon-magic-stick"></i>
          画同款
        </el-button>
      </div>

      <!-- 回到顶部按钮 -->
      <el-backtop :visibility-height="200" :right="40" :bottom="80">
        <div class="back-to-top">
          <i class="el-icon-caret-top"></i>
        </div>
      </el-backtop>
    </div>

    <!-- 图片预览 -->
    <div class="image-preview-overlay" v-if="previewVisible" @click="closePreview" 
         @wheel.prevent="handlePreviewZoom" 
         @mousedown="startPreviewDrag" 
         @mousemove="onPreviewDrag" 
         @mouseup="stopPreviewDrag" 
         @mouseleave="stopPreviewDrag">
      <img :src="previewUrl" :style="previewImageStyle" class="preview-image">
      <div class="preview-close" @click.stop="closePreview">
        <i class="el-icon-close"></i>
      </div>
    </div>
  </div>
</template>

<script>
import { getDirectFileUrl } from '@/utils/fileAccess'
import axios from 'axios'
import { mapState } from 'vuex'

export default {
  name: 'InspirationDetail',
  data() {
    return {
      task: null,
      loading: true,
      detailVideoLoaded: false,
      paramVideoLoadedMap: {},
      outputVideoLoadedMap: {},
      previewVisible: false,
      previewUrl: '',
      isMobile: false,
      taskCache: {},
      retryCount: 0,
      isLiked: false,
      outputImages: [],
      hasVideoOutput: false,
      previewImageStyle: {},
      previewDragStart: null,
      previewDragOffset: { x: 0, y: 0 },
      touchStartX: 0,
      touchEndX: 0,
      touchStartY: 0,
      currentCarouselIndex: 0,
      videoSource: '',
      showOutputParams: false
    }
  },
  created() {
    this.fetchTaskDetail()
    this.checkMobileDevice()
  },
  mounted() {
    window.addEventListener('resize', this.checkMobileDevice)
    
    // 等待DOM完全渲染后再初始化视频
    this.$nextTick(() => {
      // 初始化轮播图当前索引
      if (this.$refs.imageCarousel) {
        this.currentCarouselIndex = this.$refs.imageCarousel.activeIndex;
      }
    })
  },
  updated() {
    // 检查是否有视频需要加载但尚未加载
    if (this.task && this.hasVideoOutput && !this.detailVideoLoaded && this.$refs.detailVideo) {
      console.log('在updated钩子中检测到未加载的视频，准备加载');
      this.loadDetailVideo();
    }
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkMobileDevice)
    this.cleanupVideoResources()
  },
  computed: {
    ...mapState({
      currentUser: state => state.user.currentUser
    }),
    // 判断是否有输出图片
    hasOutputImages() {
      return this.outputImages && this.outputImages.length > 0;
    },
    // 判断是否为视频任务
    isVideoTask() {
      return this.task && (
        this.task.type === 'digital_human' || 
        this.hasVideoOutput || 
        (this.task.result_url && this.task.result_url.match(/\.(mp4|webm|mov|ogg)(\?|$)/i))
      );
    },
    // 计算理想的轮播图高度
    carouselHeight() {
      if (!this.isMobile) return '400px';
      // 移动端下使用屏幕宽度作为高度参考
      const width = window.innerWidth;
      const height = Math.round(width * 1.0); // 1:1 比例使图片更接近屏幕宽度
      return `${height}px`;
    },
    // 获取视频URL
    videoUrl() {
      if (!this.task) return '';
      
      // 如果是数字人任务，直接使用result_url
      if (this.task.type === 'digital_human' && this.task.result_url) {
        return this.task.result_url;
      } 
      // 检查输出参数中是否有视频
      else if (this.task.output_params && Array.isArray(this.task.output_params)) {
        const videoParam = this.task.output_params.find(param => 
          param.type === 'video' && param.value);
        if (videoParam) {
          return videoParam.value;
        }
      }
      // 兜底使用result_url
      return this.task.result_url || '';
    }
  },
  methods: {
    async fetchTaskDetail() {
      this.loading = true
      this.retryCount = 0; // 重置重试计数
      this.outputImages = []; // 重置输出图片列表
      this.hasVideoOutput = false; // 重置视频输出标志
      
      try {
        const taskId = this.$route.params.id
        if (!taskId) {
          this.loading = false
          return
        }
        
        // 首先检查缓存中是否已有该任务
        if (this.taskCache[taskId]) {
          console.log('使用缓存的任务数据:', taskId)
          this.task = this.taskCache[taskId]
          this.processTaskData(this.task);
          this.loading = false
          
          // 在数据加载后延迟检查视频
          setTimeout(() => {
            this.checkAndPreloadVideo();
          }, 300);
          
          return
        }
        
        // 使用专门的详情接口
        console.log('正在请求任务详情，ID:', taskId)
        const response = await axios.get(`/api/inspiration/${taskId}`)
        this.task = response.data
        
        // 将任务添加到缓存
        this.taskCache[taskId] = response.data
        
        console.log('获取任务详情成功:', this.task)
        
        // 处理任务数据
        this.processTaskData(this.task);
        
        // 在数据加载后延迟检查视频
        setTimeout(() => {
          this.checkAndPreloadVideo();
        }, 300);
        
      } catch (error) {
        console.error('获取任务详情失败:', error.response && error.response.status || '未知错误', error.message)
        this.$message.error('获取任务详情失败')
        
        // 如果是404错误，表示任务不存在
        if (error.response && error.response.status === 404) {
          this.$message.error('未找到该创作内容')
          setTimeout(() => this.goBack(), 1500) // 1.5秒后返回
        } 
        // 如果未达到最大重试次数，则重试
        else if (this.retryCount < 3) {
          this.retryCount++
          setTimeout(() => {
            console.log(`第${this.retryCount}次重试获取任务详情`)
            this.fetchTaskDetail()
          }, 1000 * this.retryCount) // 重试间隔递增
        } else {
          setTimeout(() => this.goBack(), 1500) // 重试失败后返回列表页
        }
      } finally {
        this.loading = false
      }
    },
    
    // 处理任务数据，提取输出图片和视频信息
    processTaskData(taskData) {
      if (!taskData) return;
      
      this.outputImages = [];
      this.hasVideoOutput = false;
      this.videoSource = '';
      
      // 处理输出参数中的图片和视频
      if (taskData.output_params && Array.isArray(taskData.output_params)) {
        // 提取图片类型参数
        this.outputImages = taskData.output_params.filter(param => 
          param.type === 'image' || param.type === 'mask');
        
        // 检查是否有视频类型参数
        const videoParams = taskData.output_params.filter(param => 
          param.type === 'video' && param.value);
          
        this.hasVideoOutput = videoParams.length > 0;
        
        if (this.hasVideoOutput && videoParams[0]) {
          this.videoSource = videoParams[0].value;
          console.log('找到视频源:', this.videoSource);
        }
        
        console.log('输出图片数量:', this.outputImages.length);
        console.log('是否有视频输出:', this.hasVideoOutput);
        console.log('视频源:', this.videoSource);
        
        // 如果有轮播图，确保轮播图可以正常滚动
        if (this.hasOutputImages && this.$refs.imageCarousel) {
          this.$nextTick(() => {
            this.$refs.imageCarousel.setActiveItem(0);
            this.currentCarouselIndex = 0;
          });
        }
      }
    },
    
    // 获取视频源
    getVideoSource() {
      // 首先检查是否有输出参数中的视频
      if (this.videoSource) {
        console.log('使用输出参数中的视频源:', this.videoSource);
        return this.videoSource;
      }
      
      // 然后检查是否是数字人任务
      if (this.task && this.task.type === 'digital_human' && this.task.result_url) {
        console.log('使用数字人任务的视频源:', this.task.result_url);
        return this.task.result_url;
      }
      
      // 最后检查result_url是否是视频
      if (this.task && this.task.result_url && this.task.result_url.match(/\.(mp4|webm|mov|ogg)(\?|$)/i)) {
        console.log('使用任务结果URL作为视频源:', this.task.result_url);
        return this.task.result_url;
      }
      
      console.log('未找到可用的视频源');
      return '';
    },
    
    goBack() {
      try {
        window.history.length > 1 ? this.$router.go(-1) : this.$router.push('/inspiration')
      } catch (e) {
        console.error('返回上一页失败，跳转到灵感页', e)
        this.$router.push('/inspiration')
      }
    },
    
    checkMobileDevice() {
      const userAgent = navigator.userAgent || navigator.vendor || window.opera
      const isMobileByUA = /android|webos|iphone|ipad|ipod|blackberry|iemobile|opera mini|mobile/i.test(userAgent)
      const isMobileByWidth = window.innerWidth <= 576
      this.isMobile = isMobileByUA || isMobileByWidth
    },
    
    getTaskTypeTag(task) {
      if (task.type === 'digital_human') {
        return 'primary'
      } else {
        switch (task.task_type) {
          case 'accessory':
            return 'success'
          case 'workflow':
            return 'warning'
          default:
            return 'info'
        }
      }
    },
    
    getTaskTypeText(task) {
      if (task.type === 'digital_human') {
        return '数字人'
      } else {
        switch (task.task_type) {
          case 'accessory':
            return '饰品替换'
          case 'workflow':
            return '图像处理'
          default:
            return task.task_type || '图像处理'
        }
      }
    },
    
    loadDetailVideo() {
      if (this.detailVideoLoaded) {
        console.log('视频已加载，不需要重复加载');
        return;
      }
      
      console.log('开始加载详情视频');
      this.detailVideoLoaded = true;
      
      if (this.$refs.detailVideo) {
        try {
          console.log('视频元素已找到，准备加载');
          this.$refs.detailVideo.style.zIndex = "1";
          this.$refs.detailVideo.style.transform = "translateZ(0)";
          this.$refs.detailVideo.style.backgroundColor = "#000";
          this.$refs.detailVideo.preload = "metadata";
          
          // 获取视频URL
          const videoUrl = this.getVideoSource();
          
          if (videoUrl) {
            console.log('加载视频源:', videoUrl);
            this.$refs.detailVideo.innerHTML = this.getVideoSourceElements(videoUrl);
            this.$refs.detailVideo.load();
          } else {
            console.error('没有找到可用的视频URL');
            this.detailVideoLoaded = false;
          }
        } catch (error) {
          console.error('加载详情视频失败:', error);
          this.detailVideoLoaded = false;
        }
      } else {
        console.error('视频元素引用不存在');
        this.detailVideoLoaded = false;
      }
    },
    
    handleDetailVideoCanPlay() {
      this.detailVideoLoaded = true
    },
    
    handleDetailVideoError(event) {
      console.error('详情视频加载失败:', event)
      this.detailVideoLoaded = false
      
      // 获取视频URL
      const videoUrl = this.getVideoSource();
      
      if (videoUrl) {
        setTimeout(() => {
          console.log('重试加载视频:', videoUrl);
          this.retryLoadVideo(this.$refs.detailVideo, videoUrl);
        }, 1000);
      }
    },
    
    loadParamVideo(event, url, paramName) {
      if (this.paramVideoLoadedMap[url]) return
      
      this.$set(this.paramVideoLoadedMap, url, true)
      
      event.stopPropagation()
      console.log(`加载输入参数视频: ${paramName || 'unknown'}, URL: ${url}`);
      
      const videoEl = event.target.closest('video') || this.$el.querySelector(`[data-url="${url}"]`);
      if (videoEl && videoEl.tagName === 'VIDEO') {
        try {
          videoEl.style.zIndex = "1"
          videoEl.style.transform = "translateZ(0)"
          videoEl.style.backgroundColor = "#000"
          videoEl.preload = "metadata"
          
          videoEl.innerHTML = this.getVideoSourceElements(url)
          
          videoEl.load()
        } catch (error) {
          console.error(`加载输入参数视频失败: ${paramName || 'unknown'}`, error)
          this.$set(this.paramVideoLoadedMap, url, false)
        }
      }
    },
    
    handleParamVideoCanPlay(url) {
      this.$set(this.paramVideoLoadedMap, url, true)
    },
    
    handleParamVideoError(event, url) {
      console.error('输入参数视频加载失败:', url, event)
      this.$set(this.paramVideoLoadedMap, url, false)
      
      if (url) {
        setTimeout(() => {
          const videoEl = event.target.closest('video')
          if (videoEl) {
            this.retryLoadVideo(videoEl, url)
          }
        }, 1000)
      }
    },
    
    formatDate(timeStr) {
      const date = new Date(timeStr)
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    },
    
    previewImage(url) {
      this.previewUrl = url
      this.previewVisible = true
      // 重置预览图片状态
      this.previewImageStyle = {}
      this.previewDragStart = null
      this.previewDragOffset = { x: 0, y: 0 }
    },
    
    getVideoSourceElements(url) {
      if (!url) return ''
      
      const baseUrl = this.getDirectVideoUrl(url)
      const hasExtension = /\.(mp4|webm|mov|ogg)(\?|$)/i.test(url)
      
      if (hasExtension) {
        return `<source src="${baseUrl}" type="video/mp4">`
      }
      
      return `
        <source src="${baseUrl}" type="video/mp4">
        <source src="${baseUrl}" type="video/webm">
        <source src="${baseUrl}" type="video/ogg">
      `
    },
    
    retryLoadVideo(videoEl, url) {
      if (!videoEl || !url) return
      
      try {
        while (videoEl.firstChild) {
          videoEl.removeChild(videoEl.firstChild)
        }
        
        videoEl.innerHTML = this.getVideoSourceElements(url)
        
        videoEl.style.zIndex = "1"
        videoEl.style.transform = "translateZ(0)"
        videoEl.style.backgroundColor = "#000"
        videoEl.preload = "metadata"
        videoEl.load()
      } catch (error) {
        console.error('重新加载视频失败:', error)
      }
    },
    
    getDirectVideoUrl(url) {
      if (!url) return ''
      const timestamp = Date.now()
      return getDirectFileUrl(url) + `&_t=${timestamp}`
    },
    
    async toggleLike() {
      if (!this.currentUser) {
        this.$message.warning('请先登录')
        return
      }
      try {
        const response = await axios.post(`/api/inspiration/${this.task.id}/like`)
        if (response.data.success) {
          this.task.is_liked = !this.task.is_liked
          this.task.likes = this.task.is_liked ? (this.task.likes || 0) + 1 : (this.task.likes || 1) - 1
        }
      } catch (error) {
        console.error('点赞失败:', error)
        this.$message.error('操作失败，请稍后重试')
      }
    },

    createSimilar() {
      const route = this.task.type === 'digital_human' ? '/digital-human/create' : '/image/create'
      let prompt = ''
      if (this.task.input_params && Array.isArray(this.task.input_params)) {
        const promptParam = this.task.input_params.find(param => param.key === 'prompt')
        if (promptParam) {
          prompt = promptParam.value || ''
        }
      }
      
      this.$router.push({
        path: route,
        query: {
          prompt: prompt
        }
      })
    },

    cleanupVideoResources() {
      if (this.$refs.detailVideo) {
        try {
          this.$refs.detailVideo.pause()
        } catch (e) {
          console.error('暂停视频时出错:', e)
        }
      }
      
      const videos = this.$el.querySelectorAll('video')
      videos.forEach(video => {
        try {
          if (!video.paused) {
            video.pause()
          }
        } catch (e) {
          console.error('暂停视频时出错:', e)
        }
      })
      
      this.detailVideoLoaded = false
      this.paramVideoLoadedMap = {}
    },
    
    // 刷新内容
    refreshContent() {
      if (this.loading) return;
      
      const taskId = this.$route.params.id;
      if (taskId) {
        // 从缓存中删除当前任务
        if (this.taskCache[taskId]) {
          delete this.taskCache[taskId];
        }
        
        // 重置视频加载状态
        this.detailVideoLoaded = false;
        this.paramVideoLoadedMap = {};
        
        // 重置重试计数
        this.retryCount = 0;
        
        // 重新获取任务详情
        this.fetchTaskDetail();
        
        this.$message({
          message: '正在刷新内容...',
          type: 'info',
          duration: 1000
        });
      }
    },

    closePreview() {
      // 添加渐隐效果
      const previewOverlay = document.querySelector('.image-preview-overlay');
      if (previewOverlay) {
        previewOverlay.style.opacity = '0';
        previewOverlay.style.transition = 'opacity 0.3s ease';
        
        setTimeout(() => {
          this.previewVisible = false;
          // 重置预览图片状态
          this.previewImageStyle = {};
          this.previewDragStart = null;
          this.previewDragOffset = { x: 0, y: 0 };
        }, 300);
      } else {
        this.previewVisible = false;
        // 重置预览图片状态
        this.previewImageStyle = {};
        this.previewDragStart = null;
        this.previewDragOffset = { x: 0, y: 0 };
      }
    },

    handlePreviewZoom(event) {
      const scale = event.deltaY > 0 ? 1.1 : 0.9;
      this.previewImageStyle = {
        transform: `scale(${scale})`,
        transition: 'transform 0.3s ease'
      };
    },

    startPreviewDrag(event) {
      this.previewDragStart = {
        x: event.clientX,
        y: event.clientY
      };
    },

    onPreviewDrag(event) {
      if (!this.previewDragStart) return;

      const dx = event.clientX - this.previewDragStart.x;
      const dy = event.clientY - this.previewDragStart.y;

      this.previewImageStyle = {
        transform: `translate(${dx}px, ${dy}px)`,
        transition: 'transform 0.3s ease'
      };
    },

    stopPreviewDrag() {
      this.previewDragStart = null;
      this.previewImageStyle = {};
    },

    // 在轮播图加载完成后调整高度
    adjustCarouselHeight() {
      if (!this.isMobile || !this.$refs.imageCarousel) return;
      
      this.$nextTick(() => {
        const carousel = this.$refs.imageCarousel;
        const container = carousel.$el.querySelector('.el-carousel__container');
        if (container) {
          // 设置容器高度
          const width = window.innerWidth;
          const height = Math.round(width * 0.8); // 可根据需要调整比例
          container.style.height = `${height}px`;
          
          // 更新滑块高度
          const items = carousel.$el.querySelectorAll('.el-carousel__item');
          items.forEach(item => {
            item.style.height = `${height}px`;
          });
          
          console.log('已调整轮播图高度:', height);
        }
      });
    },

    handleTouchStart(event) {
      this.touchStartX = event.touches[0].clientX;
      this.touchStartY = event.touches[0].clientY;
    },
    
    handleTouchMove(event) {
      // 不阻止默认行为，允许页面滚动
      const touchX = event.touches[0].clientX;
      // const touchY = event.touches[0].clientY;
      
      // 计算水平和垂直移动距离
      const diffX = this.touchStartX - touchX;
      const diffY = Math.abs(event.touches[0].clientY - this.touchStartY);
      
      // 只有当水平滑动距离大于垂直滑动距离且大于阈值时，才阻止默认行为并记录位置
      if (Math.abs(diffX) > diffY && Math.abs(diffX) > 10) {
        event.preventDefault();
        this.touchEndX = touchX;
      }
    },
    
    handleTouchEnd() {
      const diff = this.touchStartX - this.touchEndX;
      const threshold = 50; // 滑动阈值
      
      if (Math.abs(diff) > threshold) {
        if (diff > 0) {
          // 向左滑动，显示下一张
          this.nextImage();
        } else {
          // 向右滑动，显示上一张
          this.prevImage();
        }
      }
    },
    
    nextImage() {
      if (this.$refs.imageCarousel) {
        this.$refs.imageCarousel.next();
        this.currentCarouselIndex = this.$refs.imageCarousel.activeIndex;
      }
    },
    
    prevImage() {
      if (this.$refs.imageCarousel) {
        this.$refs.imageCarousel.prev();
        this.currentCarouselIndex = this.$refs.imageCarousel.activeIndex;
      }
    },
    
    setActiveItem(index) {
      if (this.$refs.imageCarousel) {
        this.$refs.imageCarousel.setActiveItem(index);
        this.currentCarouselIndex = index;
      }
    },
    
    // 轮播图状态变化监听
    handleCarouselChange(index) {
      this.currentCarouselIndex = index;
    },

    // 加载输出参数视频
    loadOutputVideo(event, url, paramName) {
      if (this.outputVideoLoadedMap[url]) return;
      
      this.$set(this.outputVideoLoadedMap, url, true);
      
      event.stopPropagation();
      console.log(`加载输出参数视频: ${paramName || 'unknown'}, URL: ${url}`);
      
      const videoEl = event.target.closest('video') || this.$el.querySelector(`[data-url="${url}"]`);
      if (videoEl && videoEl.tagName === 'VIDEO') {
        try {
          videoEl.style.zIndex = "1";
          videoEl.style.transform = "translateZ(0)";
          videoEl.style.backgroundColor = "#000";
          videoEl.preload = "metadata";
          
          videoEl.innerHTML = this.getVideoSourceElements(url);
          videoEl.load();
        } catch (error) {
          console.error(`加载输出参数视频失败: ${paramName || 'unknown'}`, error);
          this.$set(this.outputVideoLoadedMap, url, false);
        }
      }
    },
    
    handleOutputVideoCanPlay(url) {
      this.$set(this.outputVideoLoadedMap, url, true);
    },
    
    handleOutputVideoError(event, url) {
      console.error('输出参数视频加载失败:', url, event);
      this.$set(this.outputVideoLoadedMap, url, false);
      
      if (url) {
        setTimeout(() => {
          const videoEl = event.target.closest('video');
          if (videoEl) {
            this.retryLoadVideo(videoEl, url);
          }
        }, 1000);
      }
    },

    // 检查并预加载视频
    checkAndPreloadVideo() {
      console.log('检查是否需要预加载视频');
      
      if (!this.task) {
        console.log('任务数据不存在，无法预加载视频');
        return;
      }
      
      // 检查是否有输出视频
      if (this.task.output_params && Array.isArray(this.task.output_params)) {
        const videoParams = this.task.output_params.filter(param => 
          param.type === 'video' && param.value);
          
        if (videoParams.length > 0) {
          this.hasVideoOutput = true;
          this.videoSource = videoParams[0].value;
          console.log('找到输出视频，URL:', this.videoSource);
          
          // 强制DOM更新
          this.$forceUpdate();
          
          // 等待DOM更新后加载视频
          this.$nextTick(() => {
            console.log('DOM更新后检查video元素', this.$refs.detailVideo ? '存在' : '不存在');
            
            setTimeout(() => {
              if (this.$refs.detailVideo) {
                console.log('开始加载详情视频');
                this.loadDetailVideo();
              } else {
                console.log('视频元素引用不存在，等待下一次更新');
                // 设置标记，等待下一次DOM更新时自动加载
                this.$nextTick(() => {
                  if (this.$refs.detailVideo) {
                    console.log('在nextTick中找到video元素，开始加载');
                    this.loadDetailVideo();
                  }
                });
              }
            }, 500);
          });
          return;
        }
      }
      
      // 检查是否是数字人任务
      if (this.task && this.task.type === 'digital_human' && this.task.result_url) {
        console.log('数字人任务，加载视频:', this.task.result_url);
        this.hasVideoOutput = true; // 设置为视频任务
        
        // 强制DOM更新
        this.$forceUpdate();
        
        this.$nextTick(() => {
          setTimeout(() => {
            if (this.$refs.detailVideo) {
              this.loadDetailVideo();
            } else {
              console.log('数字人任务视频元素引用不存在');
            }
          }, 500);
        });
        return;
      }
      
      console.log('没有找到需要预加载的视频');
    },
  },
  watch: {
    '$route.params.id': {
      handler(newId, oldId) {
        if (newId && newId !== oldId) {
          console.log('路由参数变化，重新获取任务详情:', newId)
          this.fetchTaskDetail()
        }
      },
      immediate: false
    },
    // 监听轮播图变化
    '$refs.imageCarousel.activeIndex': {
      handler(newIndex) {
        if (newIndex !== undefined) {
          this.currentCarouselIndex = newIndex;
        }
      },
      deep: true
    }
  }
}
</script>

<style scoped>
.inspiration-detail-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  min-height: calc(100vh - 40px);
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
  padding: 0 12px;
  z-index: 2000 !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  color: #fff;
  width: 100%;
  box-sizing: border-box;
}

.header-back {
  display: flex;
  align-items: center;
  font-size: 16px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
  z-index: 2100 !important;
  position: relative;
}


.mobile-container {
  padding: 0;
  margin: 0;
  width: 100%;
  max-width: 100%;
  min-height: 100vh;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #fff;
  z-index: 2000;
  overflow-y: auto;
  overflow-x: hidden;
}

/* 页面标题样式 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 20px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
}

.header-left {
  display: flex;
  align-items: center;
  font-size: 16px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
  z-index: 2100 !important;
  position: relative;
}

.back-icon {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
  transition: color 0.3s ease;
  margin-right: 5px;
}

.back-icon:hover {
  color: #409EFF;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: #303133;
  font-weight: 600;
  max-width: 220px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-left: 10px;
}

.refresh-icon {
  font-size: 18px;
  color: #909399;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 8px;
  border-radius: 4px;
}

.refresh-icon:hover {
  color: #409EFF;
  background-color: #ecf5ff;
}

.refresh-icon.is-loading {
  animation: rotating 2s linear infinite;
}

@keyframes rotating {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 移动端标题样式 */
.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background-color: #409EFF;
  display: flex;
  align-items: center;
  padding: 0 12px;
  z-index: 2000 !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  color: #fff;
  width: 100%;
  box-sizing: border-box;
}

.mobile-header h2 {
  font-size: 16px;
  font-weight: 600;
  max-width: 220px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mobile-header .back-icon {
  font-size: 18px;
}

.mobile-header-placeholder {
  height: 56px;
  margin-bottom: 10px;
}

/* 底部固定按钮 */
.mobile-fixed-bottom {
    position: fixed !important;
    bottom: 0 !important;
    left: 0 !important;
    right: 0 !important;
    z-index: 2002 !important;
    background-color: #fff !important;
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1) !important;
}

.create-similar-btn {
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

/* 加载动画容器 */
.loading-container {
  padding: 24px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
}

/* 详情页样式 */
.task-detail {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.mobile-detail {
  border-radius: 0;
  box-shadow: none;
  padding-bottom: 80px; /* 为固定底部按钮留出空间 */
}

.detail-header {
  padding: 24px;
  border-bottom: 1px solid #ebeef5;
  background-color: #fafafa;
}

.mobile-detail .detail-header {
  padding: 15px;
  display: none;
}

.detail-header h2 {
  margin: 0 0 12px;
  font-size: 24px;
  color: #303133;
  font-weight: 600;
}

.detail-header p {
  color: #606266;
  font-size: 14px;
  margin: 0;
  line-height: 1.6;
}

/* 预览区域 */
.detail-preview {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #ebeef5;
  position: relative;
}

.mobile-preview {
  margin: 0;
  padding: 0;
  border-bottom: none;
}

.detail-image {
  width: 100%;
}

.mobile-image {
  width: 100%;
  margin: 0;
  overflow: hidden;
}

.mobile-image .el-carousel {
  border-radius: 0;
  box-shadow: none;
  width: 100%;
}

.mobile-image .el-carousel__container {
  height: auto !important;
}

.mobile-image .carousel-item {
  height: auto;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mobile-image .full-image {
  width: 100%;
  height: auto;
  max-height: none;
  object-fit: cover;
  object-position: center;
}

/* 自定义指示器样式 */
.custom-indicators {
  position: absolute;
  bottom: 40px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  justify-content: center;
  z-index: 10;
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 20px;
  padding: 6px 10px;
}

.indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.5);
  margin: 0 4px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.indicator-dot.active {
  background-color: #fff;
  transform: scale(1.2);
}

/* 调整移动端轮播图指示器 */
.mobile-image .el-carousel__indicators {
  display: none !important;
}

.mobile-image .el-carousel__indicators--outside {
  display: none !important;
}

.mobile-image .el-carousel__indicator {
  padding: 8px 4px;
}

.mobile-image .el-carousel__button {
  width: 8px;
  height: 8px;
  background-color: rgba(255, 255, 255, 0.4);
  border-radius: 50%;
}

.mobile-image .el-carousel__indicator.is-active .el-carousel__button {
  background-color: #fff;
}

.detail-carousel {
  width: 100%;
  overflow: hidden;
  touch-action: pan-y;
}

.full-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  cursor: zoom-in;
  background-color: #f8f8f8;
  transition: transform 0.3s ease;
}

.full-video {
  width: 100%;
  max-height: 400px;
  border-radius: 4px;
  background-color: #000;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 1;
  object-fit: contain;
}

.detail-video {
  position: relative;
  margin-bottom: 20px;
  width: 100%;
  background-color: #000;
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mobile-video {
  width: 100vw;
  margin: 0 -1px;
}

.mobile-video .full-video {
  width: 100%;
  min-height: 200px;
  max-height: none; 
  height: auto;
  object-fit: contain;
  margin: 0 auto;
  display: block;
}

.video-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 2;
  transition: background-color 0.3s ease;
}

.video-placeholder:hover {
  background-color: #111;
}

.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #f8f8f8;
  transition: transform 0.3s ease;
}

.video-placeholder:hover .loading-indicator {
  transform: scale(1.1);
}

.loading-indicator i {
  font-size: 36px;
  margin-bottom: 8px;
  color: #409EFF;
}

/* 元信息样式 */
.detail-meta {
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-top: 20px;
}

.mobile-meta {
  border-radius: 0;
  padding: 12px;
}

.meta-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.meta-user, .meta-likes {
  display: flex;
  align-items: center;
  color: #606266;
}

.meta-user i, .meta-likes i {
  margin-right: 5px;
  font-size: 16px;
}

.meta-likes i {
  cursor: pointer;
  transition: all 0.3s ease;
}

.meta-likes i:hover {
  color: #f56c6c;
  transform: scale(1.1);
}

.meta-likes i.is-liked {
  color: #f56c6c;
}

.meta-time {
  display: flex;
  align-items: center;
  color: #909399;
  font-size: 14px;
  margin-bottom: 10px;
}

.meta-time i {
  margin-right: 5px;
}

/* 参数展示 */
.detail-params {
  margin-bottom: 20px;
  padding: 20px;
}

.mobile-params-section {
  padding: 12px;
}

.detail-params h3 {
  margin: 0 0 10px;
  font-size: 16px;
  color: #303133;
  border-left: 4px solid #409EFF;
  padding-left: 10px;
}

.params-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 15px;
  padding: 5px;
}

.mobile-grid {
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 10px;
  padding: 0;
}

.param-item {
  background-color: #f9f9f9;
  border-radius: 6px;
  padding: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.mobile-param {
  padding: 8px;
  margin-bottom: 8px;
}

.param-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.param-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
  font-weight: 500;
}

.param-image {
  text-align: center;
}

.thumbnail {
  max-width: 100%;
  max-height: 120px;
  object-fit: contain;
  cursor: pointer;
  border-radius: 4px;
  transition: transform 0.3s ease;
}

.thumbnail:hover {
  transform: scale(1.05);
}

.param-video {
  position: relative;
  min-height: 120px;
  margin: 10px 0;
  background-color: #000;
  border-radius: 4px;
  overflow: hidden;
}

.thumbnail-video {
  width: 100%;
  height: auto;
  max-height: 150px;
  border-radius: 4px;
  background-color: #000;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  display: block;
  margin: 0 auto;
}

.audio-player {
  width: 100%;
}

.param-text {
  font-size: 14px;
  color: #303133;
  word-break: break-word;
}

.detail-meta {
  border-top: 1px solid #ebeef5;
  padding-top: 15px;
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  justify-content: space-between;
}

.meta-item {
  display: flex;
  align-items: center;
}

.meta-label {
  font-size: 13px;
  color: #909399;
  margin-right: 5px;
}

.meta-value {
  font-size: 13px;
  color: #606266;
}

/* 图片预览 */
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.9);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: grab;
  opacity: 1;
  transition: opacity 0.3s ease;
}

.image-preview-overlay:active {
  cursor: grabbing;
}

.preview-image {
  max-width: 90%;
  max-height: 90vh;
  object-fit: contain;
  user-select: none;
  -webkit-user-drag: none;
}

.preview-close {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  font-size: 20px;
  z-index: 10000;
}

.preview-close:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

/* 回到顶部按钮样式 */
.back-to-top {
  height: 100%;
  width: 100%;
  background-color: #409EFF;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

/* 添加加载动画样式 */
.loading-animation {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin: 20px 0;
}

.loading-animation i {
  font-size: 32px;
  color: #409EFF;
  margin-bottom: 10px;
}

.loading-animation p {
  color: #606266;
  font-size: 14px;
}

/* 占位图样式 */
.no-image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #f5f7fa;
  color: #909399;
}

.no-image-placeholder i {
  font-size: 48px;
  margin-bottom: 10px;
}

/* 轮播图样式增强 */
.detail-image .el-carousel {
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.detail-image .el-carousel__indicators {
  z-index: 10;
}

.detail-image .el-carousel__item {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f8f8f8;
  overflow: hidden;
}

.mobile-image .el-carousel__indicators--outside {
  margin-top: 8px;
}

/* 响应式适配 */
@media (max-width: 576px) {
  .inspiration-detail-container {
    padding: 0;
  }

  .task-detail {
    border-radius: 0;
    box-shadow: none;
  }
  
  .mobile-header {
    padding: 10px;
  }
  
  .back-icon {
    margin-right: 10px;
  }
  
  .mobile-header h2 {
    font-size: 16px;
  }
  
  .mobile-header-placeholder {
    height: 45px;
  }
  
  .params-grid {
    grid-template-columns: 1fr;
  }
  
  .detail-meta {
    flex-direction: column;
    gap: 10px;
  }
  
  .detail-header h2 {
    font-size: 18px;
  }
  
  .detail-header p {
    font-size: 12px;
  }
  
  .detail-preview h3,
  .detail-params h3 {
    font-size: 14px;
  }
  
  .param-item {
    padding: 10px;
  }
  
  .meta-label, .meta-value {
    font-size: 12px;
  }

  .el-carousel__container {
    height: 250px !important;
  }
}

/* 添加过渡动画 */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter, .fade-leave-to {
  opacity: 0;
}

.meta-actions {
  display: flex;
  justify-content: center;
}

.meta-actions .el-button {
  padding: 8px 20px;
}

.meta-actions .el-button i {
  margin-right: 5px;
}

/* 轮播图样式 */
.el-carousel {
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.mobile-image .el-carousel {
  border-radius: 0;
  box-shadow: none;
}

.el-carousel__item {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f8f8f8;
}

.output-param {
  background-color: #f0f9ff;
}

.param-video {
  position: relative;
  min-height: 120px;
  margin: 10px 0;
  background-color: #000;
  border-radius: 4px;
  overflow: hidden;
}

.thumbnail-video {
  width: 100%;
  height: auto;
  max-height: 150px;
  border-radius: 4px;
  background-color: #000;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  display: block;
  margin: 0 auto;
}

.video-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  min-height: 120px;
  background-color: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 2;
  transition: background-color 0.3s ease;
}

.video-placeholder:hover {
  background-color: #111;
}

.loading-indicator.small {
  transform: scale(0.8);
}

.loading-indicator.small i {
  font-size: 24px;
}
</style>