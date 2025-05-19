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
          <div class="meta-likes" v-if="!isMobile">
            <i class="el-icon-star-off" :class="{'is-liked': isLiked}" @click="toggleLike"></i>
            <span>{{ likeCount || 0 }}</span>
          </div>
        </div>
        <div class="meta-time">
          <i class="el-icon-time"></i>
          <span>{{ formatDate(task.created_at) }}</span>
        </div>
      </div>

      <!-- 社交互动区域 - PC端 -->
      <div class="social-interaction" :class="{'mobile-social': isMobile}" v-if="!isMobile">
        <div class="interaction-buttons">
          <el-button type="text" class="interaction-btn" @click="toggleLike" :class="{'active': isLiked}">
            <i class="el-icon-star-off"></i>
            {{ isLiked ? '已点赞' : '点赞' }} ({{ likeCount || 0 }})
          </el-button>
          <el-button type="text" class="interaction-btn" @click="toggleFavorite" :class="{'active': isFavorited}">
            <i class="el-icon-collection"></i>
            {{ isFavorited ? '已收藏' : '收藏' }} ({{ favoriteCount || 0 }})
          </el-button>
          <el-button type="text" class="interaction-btn" @click="handleCommentsClick">
            <i class="el-icon-chat-dot-round"></i>
            评论 ({{ commentCount || 0 }})
          </el-button>
        </div>

        <!-- 评论区域 - PC端 -->
        <div class="comments-section" v-show="showComments">
          <div class="comment-header">
            <h3>评论 ({{ commentCount || 0 }})</h3>
          </div>

          <!-- 评论输入框 -->
          <div class="comment-input" v-if="isUserLoggedIn">
            <el-input 
              type="textarea" 
              :rows="2" 
              placeholder="说点什么..." 
              v-model="newComment"
              maxlength="200"
              show-word-limit
              resize="none"
            ></el-input>
            <div class="comment-submit">
              <el-button 
                type="primary" 
                size="small" 
                @click="submitComment" 
                :loading="submittingComment"
                :disabled="!newComment.trim()"
              >
                发表评论
              </el-button>
            </div>
          </div>
          <div class="login-to-comment" v-else>
            <p>请先<a href="javascript:void(0)" @click="goToLogin">登录</a>后再发表评论</p>
          </div>

          <!-- 评论列表 -->
          <div class="comments-list">
            <div v-if="loadingComments" class="loading-comments">
              <i class="el-icon-loading"></i>
              <span>加载评论中...</span>
            </div>
            <el-empty description="暂无评论" v-else-if="comments.length === 0"></el-empty>
            <div v-else class="comment-item" v-for="comment in comments" :key="comment.id || comment.ID">
              <div class="comment-author">
                <span class="username">{{ getCommentUsername(comment.user_id) }}</span>
                <div class="comment-actions" v-if="isCommentAuthor(comment)">
                  <el-button type="text" size="mini" @click="deleteComment(comment.id || comment.ID)">
                    <i class="el-icon-delete"></i>
                  </el-button>
                </div>
              </div>
              <div class="comment-content">{{ comment.content || comment.Content }}</div>
              <div class="comment-time">{{ formatCommentTime(comment.created_at || comment.CreatedAt) }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 移动端垂直悬浮按钮 -->
      <div class="mobile-float-buttons" v-if="isMobile">
        <div class="float-button" @click="toggleLike" :class="{'active': isLiked}">
          <div class="icon-container">
            <i class="el-icon-star-off"></i>
          </div>
          <span class="count">{{ likeCount || 0 }}</span>
        </div>
        
        <div class="float-button" @click="toggleFavorite" :class="{'active': isFavorited}">
          <div class="icon-container">
            <i class="el-icon-collection"></i>
          </div>
          <span class="count">{{ favoriteCount || 0 }}</span>
        </div>
        
        <div class="float-button" @click="handleMobileCommentsClick">
          <div class="icon-container">
            <i class="el-icon-chat-dot-round"></i>
          </div>
          <span class="count">{{ commentCount || 0 }}</span>
        </div>
      </div>

      <!-- 移动端画同款按钮 -->
      <div class="mobile-create-similar" v-if="isMobile" @click="createSimilar">
        <div class="create-similar-btn">
          <i class="el-icon-magic-stick"></i>
          <span>画同款</span>
        </div>
      </div>

      <!-- 移动端评论弹出层 -->
      <div class="mobile-comments-drawer" :class="{'active': showMobileComments}" v-if="isMobile">
        <div class="drawer-overlay" @click="closeMobileComments"></div>
        <div class="drawer-content">
          <div class="drawer-header">
            <span>评论 ({{ commentCount || 0 }})</span>
            <i class="el-icon-close" @click="closeMobileComments"></i>
          </div>
          
          <div class="mobile-comments-container">
            <!-- 评论列表 -->
            <div class="mobile-comments-list">
              <div v-if="loadingComments" class="loading-comments">
                <i class="el-icon-loading"></i>
                <span>加载评论中...</span>
              </div>
              <el-empty description="暂无评论" v-else-if="comments.length === 0"></el-empty>
              <div v-else class="mobile-comment-item" v-for="comment in comments" :key="comment.id || comment.ID">
                <div class="comment-author">
                  <span class="username">{{ getCommentUsername(comment.user_id) }}</span>
                  <div class="comment-actions" v-if="isCommentAuthor(comment)">
                    <el-button type="text" size="mini" @click="deleteComment(comment.id || comment.ID)">
                      <i class="el-icon-delete"></i>
                    </el-button>
                  </div>
                </div>
                <div class="comment-content">{{ comment.content || comment.Content }}</div>
                <div class="comment-time">{{ formatCommentTime(comment.created_at || comment.CreatedAt) }}</div>
              </div>
            </div>
          </div>
          
          <!-- 移动端评论输入区 -->
          <div class="mobile-comment-input-container">
            <div v-if="isUserLoggedIn">
              <div class="mobile-comment-input">
                <el-input 
                  type="textarea" 
                  :rows="2" 
                  placeholder="说点什么..." 
                  v-model="newComment"
                  maxlength="200"
                  resize="none"
                  @keyup.enter.native="submitComment"
                ></el-input>
              </div>
            </div>
            <div class="login-to-comment" v-else>
              <p>请先<a href="javascript:void(0)" @click="goToLogin">登录</a>后再发表评论</p>
            </div>
          </div>
        </div>
      </div>

      <!-- PC端按钮 -->
      <div class="meta-actions" v-if="!isMobile">
        <el-button type="primary" size="small" @click="createSimilar">
          <i class="el-icon-magic-stick"></i>
          画同款
        </el-button>
        
        <el-button type="info" size="small" @click="toggleLike" :class="{'is-liked': isLiked}">
          <i class="el-icon-star-off"></i>
          {{ isLiked ? '已点赞' : '点赞' }}
        </el-button>
        
        <el-button type="info" size="small" @click="toggleFavorite" :class="{'is-favorited': isFavorited}">
          <i class="el-icon-collection"></i>
          {{ isFavorited ? '已收藏' : '收藏' }}
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
import { likeShareTask, unlikeShareTask, favoriteShareTask, unfavoriteShareTask, addComment, deleteComment, getLikes, getFavorites, getComments } from '@/api/share'

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
      isFavorited: false,
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
      showOutputParams: false,
      // 社交互动相关数据
      likeCount: 0,
      favoriteCount: 0,
      commentCount: 0,
      comments: [],
      userInfoMap: {},
      showComments: false,
      showMobileComments: false, // 移动端评论显示状态
      newComment: '',
      submittingComment: false,
      // 加载状态
      loadingComments: false,
      loadingLikes: false,
      loadingFavorites: false
    }
  },
  created() {
    this.fetchTaskDetail()
    this.checkMobileDevice()
  },
  mounted() {
    window.addEventListener('resize', this.checkMobileDevice)
    
    // 刷新用户信息以确保用户状态最新
    if (this.$store.getters.isAuthenticated) {
      this.$store.dispatch('refreshUserInfo')
        .then(() => {
          console.log('用户信息刷新成功');
          this.logUserStatus();
          console.log('登录状态判断(mounted):', this.isUserLoggedIn);
          this.debugState();
        })
        .catch(error => {
          console.error('刷新用户信息失败:', error);
        });
    } else {
      console.log('用户未登录，不需要刷新用户信息');
      console.log('登录状态判断(mounted):', this.isUserLoggedIn);
      this.debugState();
    }
    
    // 等待DOM完全渲染后再初始化视频
    this.$nextTick(() => {
      // 初始化轮播图当前索引
      if (this.$refs.imageCarousel) {
        this.currentCarouselIndex = this.$refs.imageCarousel.activeIndex;
      }
      
      // 调试当前用户登录状态
      this.logUserStatus();
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
    
    // 确保组件销毁时移除body类，防止页面滚动被锁定
    if (this.showMobileComments) {
      document.body.classList.remove('comment-drawer-open');
    }
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
    },
    // 判断用户是否已登录的可靠方法
    isUserLoggedIn() {
      const isAuthenticated = this.$store.getters.isAuthenticated;
      const token = localStorage.getItem('token');
      const userObj = this.currentUser || JSON.parse(localStorage.getItem('user') || '{}');
      const hasId = userObj && userObj.id;
      
      // 记录日志用于调试
      console.log('用户登录状态判断:', { isAuthenticated, hasToken: !!token, hasId });
      
      return isAuthenticated && !!token && !!hasId;
    }
  },
  methods: {
    async fetchTaskDetail() {
      this.loading = true
      this.retryCount = 0; // 重置重试计数
      this.outputImages = []; // 重置输出图片列表
      this.hasVideoOutput = false; // 重置视频输出标志
      
      // 重置社交互动状态，确保每次加载详情都有正确状态
      this.isLiked = false;
      this.isFavorited = false;
      this.comments = [];
      this.likeCount = 0;
      this.favoriteCount = 0;
      this.commentCount = 0;
      
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
          
          // 加载社交数据
          await this.fetchSocialData(taskId);
          
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
        
        // 加载社交数据
        await this.fetchSocialData(taskId);
        
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
    
    // 添加一个获取分享任务ID的方法
    getShareTaskId() {
      // 优先使用share_id，这是正确的分享任务ID
      if (this.task && this.task.share_id) {
        console.log('使用task.share_id作为分享任务ID:', this.task.share_id);
        return this.task.share_id;
      }
      
      // 如果没有share_id，退回到使用id
      if (this.task && this.task.id) {
        console.log('警告：使用task.id作为分享任务ID:', this.task.id);
        return this.task.id;
      }
      
      // 都没有，使用路由参数
      const routeId = this.$route.params.id;
      console.log('警告：使用路由参数作为分享任务ID:', routeId);
      return routeId;
    },
    
    // 修改toggleLike方法
    async toggleLike() {
      this.logUserStatus(); // 打印调试信息
      
      const isAuthenticated = this.$store.getters.isAuthenticated;
      const token = localStorage.getItem('token');
      
      // 如果用户未登录，引导用户登录
      if (!isAuthenticated || !token) {
        this.$confirm('登录后才能点赞，是否前往登录?', '提示', {
          confirmButtonText: '去登录',
          cancelButtonText: '取消',
          type: 'info'
        }).then(() => {
          sessionStorage.setItem('redirect_after_login', this.$route.fullPath);
          this.$router.push('/login');
        }).catch(() => {});
        return;
      }
      
      try {
        // 获取正确的分享任务ID
        const shareTaskId = this.getShareTaskId();
        console.log('执行点赞操作，使用ID:', shareTaskId);
        
        // 用户已登录，进行点赞/取消点赞操作
        if (this.isLiked) {
          // 取消点赞
          const result = await unlikeShareTask(shareTaskId);
          if (result.success) {
            this.isLiked = false;
            this.likeCount = Math.max(0, this.likeCount - 1);
            this.$message.success('已取消点赞');
          } else {
            // 处理API错误
            this.handleApiError(result, '取消点赞失败');
          }
        } else {
          // 添加点赞
          const result = await likeShareTask(shareTaskId);
          if (result.success) {
            this.isLiked = true;
            this.likeCount += 1;
            this.$message.success('点赞成功');
          } else {
            // 处理API错误
            this.handleApiError(result, '点赞失败');
          }
        }
      } catch (error) {
        console.error('点赞操作失败:', error);
        // 检查是否是401错误
        if (error.response && error.response.status === 401) {
          this.handleUnauthorized();
        } else {
          this.$message.error('操作失败，请稍后重试');
        }
      }
    },

    createSimilar() {
      // 使用getUserData获取用户数据
      const userData = this.getUserData();
      if (!userData || !userData.id) {
        // 替换警告消息为更友好的交互
        this.$confirm('登录后才能使用此功能，是否前往登录?', '提示', {
          confirmButtonText: '去登录',
          cancelButtonText: '取消',
          type: 'info'
        }).then(() => {
          // 保存当前页面URL以便登录后返回
          sessionStorage.setItem('redirect_after_login', this.$route.fullPath);
          this.$router.push('/login');
        }).catch(() => {});
        return;
      }

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
        
        // 重置社交互动状态
        this.isLiked = false;
        this.isFavorited = false;
        this.comments = [];
        this.likeCount = 0;
        this.favoriteCount = 0;
        this.commentCount = 0;
        
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

    // 切换收藏状态
    async toggleFavorite() {
      this.logUserStatus(); // 打印调试信息
      
      const isAuthenticated = this.$store.getters.isAuthenticated;
      const token = localStorage.getItem('token');
      
      // 如果用户未登录，引导用户登录
      if (!isAuthenticated || !token) {
        this.$confirm('登录后才能收藏，是否前往登录?', '提示', {
          confirmButtonText: '去登录',
          cancelButtonText: '取消',
          type: 'info'
        }).then(() => {
          sessionStorage.setItem('redirect_after_login', this.$route.fullPath);
          this.$router.push('/login');
        }).catch(() => {});
        return;
      }
      
      try {
        // 获取正确的分享任务ID
        const shareTaskId = this.getShareTaskId();
        console.log('执行收藏操作，使用ID:', shareTaskId);
        
        // 用户已登录，进行收藏/取消收藏操作
        if (this.isFavorited) {
          // 取消收藏
          const result = await unfavoriteShareTask(shareTaskId);
          if (result.success) {
            this.isFavorited = false;
            this.favoriteCount = Math.max(0, this.favoriteCount - 1);
            this.$message.success('已取消收藏');
          } else {
            // 处理API错误
            this.handleApiError(result, '取消收藏失败');
          }
        } else {
          // 添加收藏
          const result = await favoriteShareTask(shareTaskId);
          if (result.success) {
            this.isFavorited = true;
            this.favoriteCount += 1;
            this.$message.success('收藏成功');
          } else {
            // 处理API错误
            this.handleApiError(result, '收藏失败');
          }
        }
      } catch (error) {
        console.error('收藏操作失败:', error);
        // 检查是否是401错误
        if (error.response && error.response.status === 401) {
          this.handleUnauthorized();
        } else {
          this.$message.error('操作失败，请稍后重试');
        }
      }
    },

    // 新增处理API错误的方法
    handleApiError(result, defaultMessage) {
      // 如果错误消息包含登录相关内容，处理为未授权错误
      if (result.message && (result.message.includes('登录') || result.message.includes('认证') || result.message.includes('授权'))) {
        this.handleUnauthorized();
      } else {
        this.$message.error(result.message || defaultMessage);
      }
    },

    // 新增处理未授权错误的方法
    handleUnauthorized() {
      // 清除本地认证信息
      this.$store.commit('clearAuth');
      this.$confirm('登录状态已失效，请重新登录', '提示', {
        confirmButtonText: '去登录',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        sessionStorage.setItem('redirect_after_login', this.$route.fullPath);
        this.$router.push('/login');
      }).catch(() => {});
    },

    // 提交评论
    async submitComment() {
      const isAuthenticated = this.$store.getters.isAuthenticated;
      const token = localStorage.getItem('token');
      
      if (!isAuthenticated || !token) {
        this.$confirm('登录后才能发表评论，是否前往登录?', '提示', {
          confirmButtonText: '去登录',
          cancelButtonText: '取消',
          type: 'info'
        }).then(() => {
          sessionStorage.setItem('redirect_after_login', this.$route.fullPath);
          this.$router.push('/login');
        }).catch(() => {});
        return;
      }
      
      if (!this.newComment.trim()) {
        this.$message.warning('评论内容不能为空');
        return;
      }
      
      this.submittingComment = true;
      
      try {
        // 获取正确的分享任务ID
        const shareTaskId = this.getShareTaskId();
        console.log('提交评论，使用ID:', shareTaskId);
        
        const result = await addComment(shareTaskId, this.newComment.trim());
        if (result.success) {
          this.$message.success('评论成功');
          // 重新获取评论列表
          await this.fetchComments();
          // 清空评论框
          this.newComment = '';
        } else {
          // 处理API错误
          this.handleApiError(result, '评论提交失败');
        }
      } catch (error) {
        console.error('评论提交失败:', error);
        // 检查是否是401错误
        if (error.response && error.response.status === 401) {
          this.handleUnauthorized();
        } else {
          this.$message.error('评论提交失败，请稍后重试');
        }
      } finally {
        this.submittingComment = false;
      }
    },

    // 删除评论
    async deleteComment(commentId) {
      const isAuthenticated = this.$store.getters.isAuthenticated;
      const token = localStorage.getItem('token');
      
      if (!isAuthenticated || !token) {
        this.$confirm('登录后才能删除评论，是否前往登录?', '提示', {
          confirmButtonText: '去登录',
          cancelButtonText: '取消',
          type: 'info'
        }).then(() => {
          sessionStorage.setItem('redirect_after_login', this.$route.fullPath);
          this.$router.push('/login');
        }).catch(() => {});
        return;
      }
      
      try {
        const result = await this.$confirm('确定要删除这条评论吗?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        });
        
        if (result === 'confirm') {
          const deleteResult = await deleteComment(commentId);
          if (deleteResult.success) {
            this.$message.success('评论已删除');
            // 更新评论列表 - 直接从服务器重新获取最新评论列表
            await this.fetchComments();
            // 更新评论计数
            this.commentCount = this.comments.length;
          } else {
            // 处理API错误
            this.handleApiError(deleteResult, '删除评论失败');
          }
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除评论失败:', error);
          // 检查是否是401错误
          if (error.response && error.response.status === 401) {
            this.handleUnauthorized();
          } else {
            this.$message.error('删除评论失败，请稍后重试');
          }
        }
      }
    },

    // 新增单独获取评论方法
    async fetchComments() {
      if (!this.task) return;
      
      const shareTaskId = this.getShareTaskId();
      console.log('获取评论列表，使用ID:', shareTaskId);
      
      this.loadingComments = true;
      try {
        const commentsResult = await getComments(shareTaskId);
        if (commentsResult.success) {
          this.comments = commentsResult.comments || [];
          this.commentCount = this.comments.length;
          
          // 更新用户信息映射
          commentsResult.userInfos.forEach(user => {
            this.$set(this.userInfoMap, user.user_id, user);
          });
          
          // 按时间排序评论，最新的在前面
          this.comments.sort((a, b) => new Date(b.created_at || b.CreatedAt) - new Date(a.created_at || a.CreatedAt));
        }
      } catch (error) {
        console.error('获取评论数据失败:', error);
      } finally {
        this.loadingComments = false;
      }
    },

    // 格式化评论时间
    formatCommentTime(timeStr) {
      const now = new Date();
      const commentTime = new Date(timeStr);
      const diffMs = now - commentTime;
      
      // 转换为秒
      const diffSec = Math.floor(diffMs / 1000);
      
      if (diffSec < 60) {
        return '刚刚';
      } else if (diffSec < 3600) {
        return `${Math.floor(diffSec / 60)}分钟前`;
      } else if (diffSec < 86400) {
        return `${Math.floor(diffSec / 3600)}小时前`;
      } else if (diffSec < 86400 * 7) {
        return `${Math.floor(diffSec / 86400)}天前`;
      } else {
        return this.formatDate(timeStr);
      }
    },

    // 获取评论用户名
    getCommentUsername(userId) {
      console.log('获取评论用户名，使用ID:', userId);
      console.log('userInfoMap:', this.userInfoMap);
      if (this.userInfoMap[userId]) {
        console.log('userInfoMap[userId]:', this.userInfoMap[userId]);
        return this.userInfoMap[userId].username || '用户';
      }
      console.log('userInfoMap[userId]不存在');
      return '用户';
    },

    // 判断当前用户是否是评论作者
    isCommentAuthor(comment) {
      if (!this.isUserLoggedIn) return false;
      
      try {
        // 使用getUserData确保能获取到用户信息
        const userData = this.getUserData();
        if (!userData || !userData.id) return false;
        
        // 确保用相同的数据类型比较ID
        const commentUserId = comment.user_id ? parseInt(comment.user_id) : null;
        const currentUserId = userData.id ? parseInt(userData.id) : null;
        
        const match = commentUserId !== null && currentUserId !== null && currentUserId === commentUserId;
        console.log(`评论作者判断: 评论UserID=${commentUserId}, 当前用户ID=${currentUserId}, 匹配结果=${match}`);
        
        return match;
      } catch (e) {
        console.error('判断评论作者错误:', e);
        return false;
      }
    },

    // 修改fetchSocialData方法
    async fetchSocialData(taskId) {
      if (!taskId) return;
      
      // 获取正确的分享任务ID
      // 注意：这里可能需要先请求详情来获取share_id
      const shareTaskId = this.task ? this.getShareTaskId() : taskId;
      console.log('获取社交数据，使用ID:', shareTaskId);
      
      // 确保用户信息是最新的
      if (this.$store.getters.isAuthenticated) {
        try {
          await this.$store.dispatch('refreshUserInfo');
          console.log('社交数据获取前刷新用户信息成功');
        } catch (error) {
          console.error('刷新用户信息失败:', error);
        }
      }
      
      // 调试当前用户登录状态
      this.logUserStatus();
      
      // 获取用户数据 - 即使this.currentUser为undefined也能获取
      const userData = this.getUserData();
      
      try {
        // 并行获取所有社交数据，提高加载速度
        const [likesResult, favoritesResult, commentsResult] = await Promise.all([
          getLikes(shareTaskId),
          getFavorites(shareTaskId),
          getComments(shareTaskId)
        ]);
        
        // 处理点赞数据
        if (likesResult.success) {
          this.likeCount = likesResult.userInfos.length;
          // 更新用户信息映射
          likesResult.userInfos.forEach(user => {
            this.$set(this.userInfoMap, user.user_id, user);
          });
          
          // 检查当前用户是否已点赞
          if (this.isUserLoggedIn && userData && userData.id) {
            // 修复点赞检查逻辑，使用user_id
            this.isLiked = likesResult.userInfos.some(user => {
              // 确保用相同的数据类型比较ID
              const userId = parseInt(user.user_id);
              const currentUserId = parseInt(userData.id);
              const match = userId === currentUserId;
              console.log(`比较点赞用户ID: ${userId} vs ${currentUserId}, 匹配结果: ${match}`);
              return match;
            });
            console.log('当前用户ID:', userData.id, '是否已点赞:', this.isLiked);
          }
        }
        
        // 处理收藏数据
        if (favoritesResult.success) {
          this.favoriteCount = favoritesResult.userInfos.length;
          // 更新用户信息映射
          favoritesResult.userInfos.forEach(user => {
            this.$set(this.userInfoMap, user.user_id, user);
          });
          
          // 检查当前用户是否已收藏
          if (this.isUserLoggedIn && userData && userData.id) {
            // 修复收藏检查逻辑，使用user_id
            this.isFavorited = favoritesResult.userInfos.some(user => {
              // 确保用相同的数据类型比较ID
              const userId = parseInt(user.user_id);
              const currentUserId = parseInt(userData.id);
              const match = userId === currentUserId;
              console.log(`比较收藏用户ID: ${userId} vs ${currentUserId}, 匹配结果: ${match}`);
              return match;
            });
            console.log('当前用户ID:', userData.id, '是否已收藏:', this.isFavorited);
          }
        }
        
        // 处理评论数据
        if (commentsResult.success) {
          this.comments = commentsResult.comments || [];
          this.commentCount = this.comments.length;
          
          // 更新用户信息映射
          commentsResult.userInfos.forEach(user => {
            this.$set(this.userInfoMap, user.user_id, user);
          });
          
          // 按时间排序评论，最新的在前面
          this.comments.sort((a, b) => new Date(b.created_at || b.CreatedAt) - new Date(a.created_at || a.CreatedAt));
        }
      } catch (error) {
        console.error('获取社交数据失败:', error);
        this.$message.error('获取社交数据失败，请刷新页面重试');
      }
    },

    // 修改showComments的切换处理
    handleCommentsClick() {
      // 使用计算属性判断用户是否登录
      if (!this.isUserLoggedIn) {
        // 更友好的交互，提供用户选择
        this.$confirm('登录后才能发表评论，是否前往登录?', '提示', {
          confirmButtonText: '去登录',
          cancelButtonText: '只看评论',
          type: 'info'
        }).then(() => {
          sessionStorage.setItem('redirect_after_login', this.$route.fullPath);
          this.$router.push('/login');
        }).catch(() => {
          // 用户选择"只看评论"，也触发评论加载
          this.showComments = !this.showComments;
          if (this.showComments && this.comments.length === 0) {
            this.fetchComments();
          }
        });
      } else {
        // 已登录用户直接切换评论显示状态
        this.showComments = !this.showComments;
        if (this.showComments && this.comments.length === 0) {
          this.fetchComments();
        }
      }
    },

    // 移动端评论点击处理
    handleMobileCommentsClick() {
      // 显示移动端评论抽屉
      this.showMobileComments = true;
      document.body.classList.add('comment-drawer-open'); // 添加阻止滚动的类
      
      // 确保页面滚动到顶部
      setTimeout(() => {
        // 如果有评论列表容器，重置其滚动位置
        const commentsContainer = document.querySelector('.mobile-comments-container');
        if (commentsContainer) {
          commentsContainer.scrollTop = 0;
        }
      }, 100);
      
      // 如果评论还未加载，则加载评论
      if (this.comments.length === 0) {
        this.fetchComments();
      }
    },
    
    // 关闭移动端评论
    closeMobileComments() {
      this.showMobileComments = false;
      document.body.classList.remove('comment-drawer-open'); // 移除阻止滚动的类
      
      // 允许评论区内容自然渐隐
      const drawerContent = document.querySelector('.drawer-content');
      if (drawerContent) {
        drawerContent.style.transform = 'translateY(100%)';
      }
    },

    // 添加方法跳转到登录页面
    goToLogin() {
      // 保存当前页面URL以便登录后返回
      sessionStorage.setItem('redirect_after_login', this.$route.fullPath);
      this.$router.push('/login');
    },

    // 调试用户登录状态
    logUserStatus() {
      const userData = this.getUserData();
      console.log('当前用户信息:', userData || this.currentUser);
      
      // 检查Vuex存储的用户信息
      console.log('Vuex存储用户状态:', this.$store.getters.isAuthenticated);
      console.log('localStorage中的token:', localStorage.getItem('token'));
      
      const userStr = localStorage.getItem('user');
      console.log('localStorage中的user字符串:', userStr);
      
      try {
        const userObj = JSON.parse(userStr || '{}');
        console.log('解析后的user对象:', userObj);
        console.log('是否有ID:', !!userObj.id);
        
        if (userData) {
          console.log('用户ID:', userData.id, '类型:', typeof userData.id);
          console.log('userData是否有ID属性:', 'id' in userData);
        } else if (this.currentUser) {
          console.log('用户ID:', this.currentUser.id, '类型:', typeof this.currentUser.id);
          console.log('this.currentUser是否有ID属性:', 'id' in this.currentUser);
        } else {
          console.log('userData和this.currentUser都为空或未定义');
        }
      } catch(e) {
        console.error('解析user对象失败:', e);
      }
    },

    // 添加新的调试方法
    debugState() {
      console.log('----------- 调试状态信息 -----------');
      console.log('isUserLoggedIn:', this.isUserLoggedIn);
      console.log('当前用户信息:', this.currentUser);
      console.log('localStorage中用户信息:', localStorage.getItem('user'));
      console.log('localStorage中token:', localStorage.getItem('token'));
      console.log('Vuex中的认证状态:', this.$store.getters.isAuthenticated);
      console.log('--------------------------------------');
    },

    // 添加获取用户数据的方法，兼容this.currentUser为undefined的情况
    getUserData() {
      // 尝试从Vuex获取
      if (this.currentUser && this.currentUser.id) {
        console.log('从Vuex获取用户数据:', this.currentUser);
        return this.currentUser;
      }
      
      // 尝试从localStorage获取
      try {
        const userStr = localStorage.getItem('user');
        if (userStr) {
          const userData = JSON.parse(userStr);
          console.log('从localStorage获取用户数据:', userData);
          return userData;
        }
      } catch (error) {
        console.error('解析localStorage用户数据失败:', error);
      }
      
      // 都失败则返回null
      console.log('无法获取用户数据');
      return null;
    }
  },
  watch: {
    '$route.params.id': {
      handler(newId, oldId) {
        if (newId && newId !== oldId) {
          console.log('路由参数变化，重新获取任务详情:', newId)
          this.fetchTaskDetail()
          
          // 重置社交互动状态，确保当返回列表页再进入时状态正确
          this.isLiked = false
          this.isFavorited = false
          this.comments = []
          this.likeCount = 0
          this.favoriteCount = 0
          this.commentCount = 0
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
  },
  // 在activated钩子中刷新社交数据，确保在从缓存中恢复时状态最新
  activated() {
    if (this.$route.params.id) {
      // 只刷新社交数据，不重新加载整个任务详情
      this.fetchSocialData(this.$route.params.id)
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
  padding: 0 !important;
  margin: 0 !important;
  width: 100vw !important; /* 使用视口宽度单位确保占满屏幕 */
  max-width: 100vw !important;
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
  -webkit-overflow-scrolling: touch; /* 增强iOS滚动体验 */
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
    padding: 0;
}

.mobile-interaction-buttons {
    display: flex;
    width: 100%;
    height: 56px;
    align-items: center;
}

.mobile-interaction-buttons .interaction-btn {
    flex: 1;
    height: 100%;
    margin: 0;
    border-radius: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0;
    font-size: 12px;
}

.mobile-interaction-buttons .interaction-btn i {
    font-size: 18px;
    margin-right: 0;
    margin-bottom: 3px;
}

.mobile-interaction-buttons .interaction-btn.active {
    color: #409EFF;
}

.mobile-interaction-buttons .create-similar-btn {
    flex: 2;
    height: 56px;
    font-size: 16px;
    font-weight: 500;
    border-radius: 0;
    margin: 0;
    background: linear-gradient(135deg, #1976d2, #64b5f6);
    border: none;
    color: #fff;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
}

.mobile-interaction-buttons .create-similar-btn i {
    margin-right: 5px;
    margin-bottom: 0;
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
  border-radius: 0 !important;
  box-shadow: none !important;
  margin: 0 !important;
  padding: 0 !important;
  width: 100vw !important;
  padding-bottom: 120px !important; /* 为底部画同款按钮留出更多空间 */
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
  width: 100vw !important;
  margin: 0 !important;
  padding: 0 !important;
}

.detail-image {
  width: 100%;
}

.mobile-image {
  width: 100vw !important;
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
  gap: 15px;
  margin-top: 20px;
}

.meta-actions .el-button {
  padding: 10px 20px;
  border-radius: 20px;
  transition: all 0.3s ease;
}

.meta-actions .el-button i {
  margin-right: 5px;
}

.meta-actions .el-button.is-liked,
.meta-actions .el-button.is-favorited {
  background-color: #409EFF;
  border-color: #409EFF;
  color: #fff;
}

.meta-actions .el-button:hover {
  background-color: #64b5f6;
}

.meta-actions .el-button:active {
  background-color: #409EFF;
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

/* 社交互动区域 */
.social-interaction {
  margin-top: 20px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.mobile-social {
  border-radius: 0;
  box-shadow: none;
}

.interaction-buttons {
  display: flex;
  justify-content: space-around;
  padding: 15px 0;
  border-bottom: 1px solid #ebeef5;
}

.interaction-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  transition: all 0.3s ease;
}

.interaction-btn:hover {
  color: #409EFF;
  background: rgba(64, 158, 255, 0.1);
  border-radius: 20px;
}

.interaction-btn.active {
  color: #409EFF;
  font-weight: 500;
}

.interaction-btn i {
  margin-right: 5px;
  font-size: 18px;
}

/* 评论区域 */
.comments-section {
  padding: 20px;
  background-color: #fff;
}

.comment-header {
  margin-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
  padding-bottom: 10px;
}

.comment-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.comment-input {
  margin-bottom: 20px;
}

.comment-submit {
  text-align: right;
  margin-top: 10px;
}

.login-to-comment {
  text-align: center;
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 4px;
  margin-bottom: 20px;
}

.login-to-comment a {
  color: #409EFF;
  text-decoration: none;
  font-weight: 500;
}

.comments-list {
  max-height: 500px;
  overflow-y: auto;
}

.loading-comments {
  text-align: center;
  padding: 20px;
  color: #909399;
}

.loading-comments i {
  margin-right: 8px;
  font-size: 20px;
}

.comment-item {
  padding: 15px;
  border-bottom: 1px solid #f2f2f2;
  transition: background-color 0.3s ease;
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-item:hover {
  background-color: #f8f9fa;
}

.comment-author {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.comment-author .username {
  font-weight: 500;
  color: #303133;
}

.comment-content {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 8px;
  word-break: break-word;
}

.comment-time {
  color: #909399;
  font-size: 12px;
}

.comment-actions {
  opacity: 0;
  transition: opacity 0.3s ease;
}

.comment-item:hover .comment-actions {
  opacity: 1;
}

@media (max-width: 576px) {
  .social-interaction {
    margin-top: 10px;
  }
  
  .interaction-buttons {
    padding: 10px 0;
  }
  
  .interaction-btn {
    padding: 5px 10px;
    font-size: 13px;
  }
  
  .comments-section {
    padding: 15px;
  }
  
  .comment-item {
    padding: 10px;
  }
  
  .comment-actions {
    opacity: 1;
  }
}

/* 移动端底部固定按钮 - 不再需要 */
.mobile-fixed-bottom {
  display: none;
}

/* 移动端垂直悬浮按钮 */
.mobile-float-buttons {
  position: fixed;
  right: 16px;
  bottom: 130px;
  display: flex;
  flex-direction: column;
  align-items: center;
  z-index: 1000;
  pointer-events: none; /* 防止影响下方内容的点击 */
}

.float-button {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 15px;
  cursor: pointer;
  pointer-events: auto; /* 恢复按钮自身的点击 */
}

.icon-container {
  width: 45px;
  height: 45px;
  border-radius: 50%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  color: white;
  font-size: 20px;
  margin-bottom: 4px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.float-button:active .icon-container {
  transform: scale(0.9);
}

.float-button.active .icon-container {
  background-color: #409EFF;
}

.float-button .count {
  color: white;
  font-size: 12px;
  background-color: rgba(0, 0, 0, 0.3);
  padding: 1px 5px;
  border-radius: 10px;
  min-width: 20px;
  text-align: center;
}

/* 移动端画同款按钮 */
.mobile-create-similar {
  position: fixed;
  bottom: 20px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  padding: 0 20px;
  pointer-events: none; /* 防止影响下方内容的点击 */
}

.create-similar-btn {
  width: 100%;
  max-width: 300px;
  height: 50px;
  background: linear-gradient(135deg, #1976d2, #64b5f6);
  color: white;
  border-radius: 25px;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 16px;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  cursor: pointer;
  transition: all 0.3s ease;
  pointer-events: auto; /* 恢复按钮自身的点击 */
}

.create-similar-btn:active {
  transform: scale(0.97);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.create-similar-btn i {
  margin-right: 8px;
  font-size: 20px;
}

/* 移动端评论抽屉 */
.mobile-comments-drawer {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 2001;
  display: flex;
  flex-direction: column;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.mobile-comments-drawer.active {
  opacity: 1;
  pointer-events: auto;
}

.drawer-overlay {
  position: fixed; /* 使用fixed定位替代absolute */
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 2001;
  touch-action: none; /* 禁止在遮罩上的所有触摸操作 */
}

.drawer-content {
  position: fixed; /* 使用fixed定位替代absolute */
  bottom: 0;
  left: 0;
  width: 100%;
  height: 70%;
  background-color: white;
  border-radius: 16px 16px 0 0;
  z-index: 2002;
  display: flex;
  flex-direction: column;
  transform: translateY(100%);
  transition: transform 0.3s ease;
  overflow: hidden; /* 防止内容溢出 */
}

.mobile-comments-drawer.active .drawer-content {
  transform: translateY(0);
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #ebeef5;
  flex-shrink: 0; /* 防止header被压缩 */
}

.drawer-header span {
  font-weight: 600;
  font-size: 16px;
}

.drawer-header i {
  font-size: 20px;
  color: #909399;
  cursor: pointer;
  padding: 5px;
}

.mobile-comments-container {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  -webkit-overflow-scrolling: touch;
  padding: 0;
  height: calc(100% - 120px); /* 留出评论输入区域的高度 */
  position: relative;
  touch-action: pan-y; /* 只允许垂直滑动 */
  -webkit-overflow-scrolling: touch; /* 增强iOS滚动体验 */
}

.mobile-comments-drawer .drawer-content {
  display: flex;
  flex-direction: column;
  height: 70%; /* 提高高度，占据更多屏幕空间 */
  max-height: 70%; /* 限制最大高度 */
  overflow: hidden;
}

.mobile-comments-list {
  padding: 0 16px 120px 16px; /* 增加底部空间，防止内容被遮挡 */
  transform: translate3d(0, 0, 0); /* 强制使用硬件加速，提高滚动性能 */
}

.mobile-comment-input-container {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: #fff;
  border-top: 1px solid #ebeef5;
  padding: 12px 0;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  z-index: 2003; /* 确保在最上层 */
}

.mobile-comment-input {
  display: flex;
  align-items: center;
  justify-content: center; /* 居中输入框 */
}

.mobile-comment-input .el-input {
  width: 100%;
  max-width: 450px; /* 限制最大宽度，在大屏幕上也能居中 */
}

.mobile-comment-input .el-textarea__inner {
  min-height: 60px !important;
  font-size: 16px !important; /* 使用!important确保字体大小不变 */
  padding: 12px;
  border-radius: 18px;
  resize: none;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

/* 适配调整 */
.mobile-detail {
  padding-bottom: 120px; /* 增加底部空间，确保内容不被底部按钮遮挡 */
}

@media (max-width: 576px) {
  .task-detail {
    padding-right: 15px;
  }
  
  /* 确保登录提示文字居中 */
.login-to-comment {
  text-align: center;
    padding: 12px 0;
  }
}

/* 防止移动端输入框放大 */
@media screen and (max-width: 768px) {
  input[type="text"],
  input[type="number"],
  input[type="email"],
  input[type="tel"],
  input[type="password"],
  input[type="search"],
  textarea,
  select {
    font-size: 16px !important; /* 大于16px的字体不会触发iOS自动缩放 */
  }

  .el-textarea__inner {
    font-size: 16px !important;
  }
  
  /* 确保评论抽屉内的文本区域不缩放 */
  :deep(.mobile-comments-drawer .el-textarea__inner) {
    font-size: 16px !important;
    -webkit-text-size-adjust: 100%;
    text-size-adjust: 100%;
  }
}
</style>

<!-- 添加全局样式，确保评论抽屉内的文本区域不缩放 -->
<style>
.mobile-comments-drawer .el-textarea__inner {
  font-size: 16px !important;
  -webkit-text-size-adjust: 100%;
  text-size-adjust: 100%;
}

/* 修复页面整体偏左和滚动条问题 */
body.comment-drawer-open {
  overflow: hidden !important;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  height: 100%; /* 确保占满整个视口高度 */
  touch-action: none !important; /* 完全禁止背景滚动 */
}

/* 在移动端隐藏浏览器滚动条，使用更现代的方式 */
html, body {
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE and Edge */
  overscroll-behavior: none; /* 防止iOS橡皮筋效果 */
}

html::-webkit-scrollbar, 
body::-webkit-scrollbar,
.mobile-container::-webkit-scrollbar,
.mobile-comments-container::-webkit-scrollbar {
  display: none; /* Chrome, Safari and Opera */
}

.inspiration-detail-container {
  overflow-x: hidden;
  width: 100%;
}

/* 修复评论区域滚动问题 */
.mobile-comments-container {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  -webkit-overflow-scrolling: touch;
  padding: 0;
  height: calc(100% - 120px); /* 留出评论输入区域的高度 */
  position: relative;
}

.mobile-comments-drawer .drawer-content {
  display: flex;
  flex-direction: column;
  max-height: 60%; /* 限制最大高度 */
}

.mobile-comment-input-container {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: #fff;
  border-top: 1px solid #ebeef5;
  padding: 12px 0;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  z-index: 10;
}

/* 恢复移动端评论输入框样式 */
.mobile-comment-input {
  display: flex;
  align-items: center;
  width: 92%;
  max-width: 600px;
  padding: 0;
  margin: 0 auto;
}

.mobile-comment-input .el-input {
  width: 100%;
}

.mobile-comment-input .el-textarea__inner {
  min-height: 45px !important;
  font-size: 16px !important;
  padding: 10px 15px;
  border-radius: 18px;
  resize: none;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  width: 100% !important;
  box-sizing: border-box !important;
}
</style>