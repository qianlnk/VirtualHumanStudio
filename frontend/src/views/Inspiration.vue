<template>
  <div class="inspiration-container">
    <!-- 更改标题样式，参考TTS模块 -->
    <div class="page-header" :class="{'mobile-header': isMobile}">
      <div class="header-left">
        <h2>灵感空间</h2>
      </div>
      <div class="header-right">
        <p v-if="!isMobile">发现用户分享的精彩创作</p>
        <!-- <el-button type="primary" icon="el-icon-refresh" size="small" @click="refreshData" :loading="loading">刷新</el-button> -->
      </div>
    </div>

    <!-- 移动端头部占位 -->
    <div v-if="isMobile" class="mobile-header-placeholder"></div>

    <!-- 加载动画 -->
    <div v-if="loading && !loadingMore" class="loading-container">
      <el-skeleton :rows="5" animated />
    </div>

    <!-- 没有内容时显示 -->
    <el-empty v-else-if="tasks.length === 0" description="暂无灵感内容">
      <el-button type="primary" @click="refreshData" :loading="loading">刷新数据</el-button>
    </el-empty>

    <!-- 灵感内容 -->
    <div v-else class="inspiration-grid">
      <div v-for="task in tasks" :key="task.id" class="inspiration-card" 
        :data-task-id="task.id" 
        :data-video-url="task.type === 'digital_human' ? task.result_url : ''"
        @click="showTaskDetail(task)">
        <div class="card-media">
          <!-- 数字人视频内容 -->
          <div v-if="task.type === 'digital_human'" class="video-container">
            <div class="video-placeholder" v-if="!videoLoadedMap[task.id]">
              <div class="loading-indicator">
                <i class="el-icon-video-play"></i>
                <span>视频加载中...</span>
              </div>
            </div>
            <video 
              v-show="videoLoadedMap[task.id]"
              :data-id="task.id" 
              :data-url="task.result_url"
              class="preview-video" 
              :controls="!isMobile"
              preload="none"
              muted
              playsinline
              webkit-playsinline
              @click.stop="showTaskDetail(task)"
              @play="handleVideoPlay(task.id, $event)"
              @pause.stop
              @timeupdate.stop
              @canplay="handleVideoCanPlay(task.id)"
              @error="handleVideoError($event, task.id)">
              <source :src="videoLoadedMap[task.id] ? getDirectVideoUrl(task.result_url) : ''" type="video/mp4">
              您的浏览器不支持视频播放
            </video>
            <div 
              v-if="videoLoadedMap[task.id]" 
              class="video-control-overlay" 
              @click.stop="showTaskDetail(task)">
              <!-- 视频控制覆盖层，点击时跳转到详情页 -->
            </div>
          </div>
          <!-- 图片内容 -->
          <div v-else class="image-container">
            <div class="image-placeholder" v-if="!imageLoadedMap[task.id]">
              <div class="loading-indicator">
                <i class="el-icon-picture"></i>
                <span>图片加载中...</span>
              </div>
            </div>
            <img 
              :src="task.result_url" 
              :alt="task.name" 
              class="preview-image" 
              @click.stop="showTaskDetail(task)"
              @load="handleImageLoaded(task.id)"
              @error="handleImageError(task.id)"
              :style="{width: '100%', height: 'auto'}">
          </div>
        </div>
        <!-- <div class="task-type-tag">
          <el-tag size="mini">
            {{ getTaskTypeText(task) }}
          </el-tag>
        </div> -->
      </div>
      
      <!-- 固定的加载更多触发器 -->
      <div class="load-more-trigger" ref="loadMoreTrigger"></div>
    </div>
    
    <!-- 加载更多提示 -->
    <div class="load-more-container" v-if="tasks.length > 0">
      <template v-if="loadingMore">
        <div class="loading-indicator">
          <i class="el-icon-loading"></i>
          <p>加载更多内容...</p>
        </div>
      </template>
      <template v-else-if="!hasMoreData && initialLoaded">
        <p>已经到底啦~</p>
      </template>
      <template v-else-if="hasMoreData && initialLoaded">
        <p>向下滚动加载更多</p>
        <el-button v-if="tasks.length > 0" type="text" @click="loadMoreTasks" :disabled="loadingMore">点击加载更多</el-button>
      </template>
    </div>

    <!-- 图片预览 -->
    <el-dialog
      :visible.sync="previewVisible"
      width="80%"
      center
      custom-class="preview-dialog"
      append-to-body>
      <img :src="previewUrl" alt="预览图" class="preview-fullsize">
    </el-dialog>
    

  </div>
</template>

<script>
import axios from 'axios'
import { getDirectFileUrl } from '@/utils/fileAccess'
export default {
  name: 'Inspiration',
  data() {
    return {
      tasks: [],
      loading: true,
      total: 0,
      currentPage: 1,
      pageSize: 12,
      previewVisible: false,
      previewUrl: '',

      videoLoadedMap: {}, // 跟踪每个视频是否已加载
      imageLoadedMap: {}, // 跟踪每个图片是否已加载
      maxConcurrentVideos: 3, // 最大同时加载的视频数量
      activeVideos: [], // 当前活动的视频元素
      cleanupTimer: null, // 清理定时器
      autoLoadVideos: true, // 始终启用自动加载视频
      intersectionObserver: null, // 用于监测元素是否进入视口
      videoIntersectionObserver: null, // 用于监测视频是否露出一半
      videoClickHandlers: {}, // 存储视频点击处理函数
      isMobile: false, // 是否为移动设备
      resizeTimer: null, // 用于延迟更新布局
      // 无限滚动相关属性
      loadingMore: false, // 是否正在加载更多
      hasMoreData: true, // 是否还有更多数据
      initialLoaded: false, // 是否已初始化加载
      observer: null, // 无限滚动的交叉观察器
      scrollThreshold: 200, // 滚动阈值
      lastScrollTop: 0, // 上次滚动位置
      pauseScrollListening: false, // 是否暂停滚动监听
    }
  },
  created() {
    // 重置初始状态
    this.currentPage = 1
    this.tasks = []
    this.hasMoreData = true
    this.initialLoaded = false
    this.loading = false  // 确保初始loading状态为false
    
    // 立即获取数据 - 使用setTimeout确保在下一个事件循环中执行
    setTimeout(() => {
      this.fetchInspiration()
    }, 0)
    
    // 检测设备类型
    this.checkDeviceType()
    
    console.log('组件创建，开始获取数据')
  },
  mounted() {
    console.log('组件挂载完成，初始loading状态=', this.loading)
    
    // 添加全局视频管理
    window.addEventListener('visibilitychange', this.handleVisibilityChange)
    
    // 检测设备类型
    this.checkDeviceType()
    
    // 初始化交叉观察器，用于视频懒加载
    this.setupIntersectionObserver()
    
    // 设置视频播放交叉观察器
    this.setupVideoPlayObserver()
    
    // 监听窗口大小变化
    window.addEventListener('resize', this.handleWindowResize)
    
    // 设置定期清理不活跃视频的定时器
    this.cleanupTimer = setInterval(this.cleanupInactiveVideos, 60000); // 每分钟检查一次
    
    // 添加滚动事件监听器，用于无限滚动
    window.addEventListener('scroll', this.handleWindowScroll)
    
    // 在DOM更新后确保CSS正确应用
    this.$nextTick(() => {
      // 强制使用纯CSS方法实现卡片高度自适应
      this.applyCardLayout()
      
      // 添加全局图片加载事件监听
      window.addEventListener('load', this.updateMasonryLayout)
      
      // 设置无限滚动观察器
      this.setupScrollObserver()
      
      // 检查是否需要初始化加载数据（如果created钩子中的加载失败）
      if (this.tasks.length === 0 && !this.loading) {
        console.log('挂载后检测到无数据，尝试重新加载')
        this.fetchInspiration()
      }
      
      // 添加一个延迟的布局更新，确保所有内容加载后布局正确
      setTimeout(() => {
        this.updateMasonryLayout()
      }, 2000)
    })
  },
  beforeDestroy() {
    // 移除事件监听
    window.removeEventListener('resize', this.handleWindowResize)
    window.removeEventListener('visibilitychange', this.handleVisibilityChange)
    window.removeEventListener('load', this.updateMasonryLayout)
    window.removeEventListener('scroll', this.handleWindowScroll)
    
    // 清理所有视频资源
    this.cleanup()
    
    // 清除定时器
    if (this.cleanupTimer) {
      clearInterval(this.cleanupTimer)
      this.cleanupTimer = null
    }
    
    // 断开交叉观察器
    this.cleanupObservers()
  },
  methods: {
    // 设置视频播放交叉观察器
    setupVideoPlayObserver() {
      // 检查浏览器支持
      if ('IntersectionObserver' in window) {
        // 创建交叉观察器，当视频元素露出一半时触发自动播放
        this.videoIntersectionObserver = new IntersectionObserver((entries) => {
          entries.forEach(entry => {
            if (entry.isIntersecting && entry.intersectionRatio >= 0.5) {
              const videoEl = entry.target;
              const taskId = videoEl.dataset.id;
              
              // 如果视频已加载且处于暂停状态，尝试播放
              if (this.videoLoadedMap[taskId] && videoEl.paused && !document.hidden) {
                // 先暂停其他视频，再播放当前视频
                this.pauseOtherVideos(taskId);
                
                videoEl.muted = true; // 确保静音，解决自动播放限制
              videoEl.play().catch(err => {
                console.warn('自动播放失败:', err);
                  // 如果失败，尝试静音播放
                  if (err.name === 'NotAllowedError') {
                    videoEl.muted = true;
                    videoEl.play().catch(e => {
                      console.error('静音播放也失败:', e);
                    });
                  }
                });
              }
            } else {
              // 如果视频不可见或露出不足一半，则暂停
              const videoEl = entry.target;
        if (videoEl && !videoEl.paused) {
          videoEl.pause();
        }
            }
          });
        }, {
          threshold: [0.5] // 当50%的元素可见时触发
        });
      }
    },
    
    // 观察视频元素用于自动播放
    observeVideosForAutoPlay() {
      if (this.videoIntersectionObserver) {
        // 先断开所有观察
        this.videoIntersectionObserver.disconnect();
        
        // 获取所有视频元素
        const videos = this.$el.querySelectorAll('video.preview-video');
        
        // 观察每个视频元素
        videos.forEach(video => {
          this.videoIntersectionObserver.observe(video);
        });
      }
    },
    
    // 处理详情关闭
    handleDetailClose() {
      this.detailVisible = false;
      this.currentTask = null;
      
      // 重新恢复视频观察
      this.$nextTick(() => {
        this.observeVideosForAutoPlay();
      });
    },
    
    // 当视频可以播放时，加入自动播放观察
    handleVideoCanPlay(taskId) {
      // 更新视频加载状态
      this.$set(this.videoLoadedMap, taskId, true)
      
      // 找到对应的视频元素
      const videoEl = this.$el.querySelector(`video[data-id="${taskId}"]`)
      if (videoEl) {
        // 确保视频控件显示（非移动设备）
        videoEl.controls = !this.isMobile
        
        // 为移动设备设置额外属性
        if (this.isMobile) {
          videoEl.setAttribute('playsinline', '')
          videoEl.setAttribute('webkit-playsinline', '')
        }
        
        // 添加到活动视频列表
        if (!this.activeVideos.includes(videoEl)) {
          this.activeVideos.push(videoEl)
        }
        
        // 记录最后活动时间
        videoEl.dataset.lastPlayed = Date.now()
        
        // 防止事件冒泡
        const stopPropagation = (e) => e.stopPropagation()
        videoEl.addEventListener('play', stopPropagation)
        videoEl.addEventListener('pause', stopPropagation)
        videoEl.addEventListener('timeupdate', stopPropagation)
        
        // 将视频添加到自动播放观察
        if (this.videoIntersectionObserver) {
          this.videoIntersectionObserver.observe(videoEl);
        }
      }
      
      // 在下一个Vue更新周期执行，确保DOM已更新
      this.$nextTick(() => {
        // 查找相关元素
        const card = this.$el.querySelector(`.inspiration-card[data-task-id="${taskId}"]`)
        const videoPlaceholder = card ? card.querySelector('.video-placeholder') : null
        
        // 隐藏占位符
        if (videoPlaceholder) {
          videoPlaceholder.style.display = 'none'
        }
        
        // 显示卡片
        if (card) {
          card.style.opacity = '1'
          card.style.transition = 'opacity 0.3s ease, transform 0.3s ease'
          
          // 触发布局更新
          this.updateMasonryLayout()
        }
      })
    },
    
    // 获取灵感内容
    fetchInspiration(loadMore = false) {
      // 检查加载状态前先输出当前状态
      console.log('fetchInspiration调用:', '当前loading状态=', this.loading, '加载更多状态=', this.loadingMore)
      
      // 简化判断条件，避免复杂逻辑
      if (loadMore && this.loadingMore) {
        console.log('已有加载更多请求进行中，跳过')
        return
      }
      
      if (!loadMore && this.loading) {
        console.log('已有初始加载请求进行中，跳过')
        return
      }
      
      console.log('开始请求数据:', '页码=', this.currentPage, '每页数量=', this.pageSize, '是否加载更多=', loadMore)
      
      // 设置适当的加载状态
      if (loadMore) {
        this.loadingMore = true
      } else {
        this.loading = true
      }
      
      // 执行请求，添加超时处理
      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => reject(new Error('请求超时')), 15000)
      })
      
      const fetchPromise = axios.get(`/api/inspiration?page=${this.currentPage}&size=${this.pageSize}`)
      
      // 使用Promise.race确保请求不会无限等待
      Promise.race([fetchPromise, timeoutPromise])
        .then(response => {
          // 如果是超时错误，则由catch块处理
          if (!response || !response.data) throw new Error('无效响应')
          
          const newTasks = response.data.tasks || []
          this.total = response.data.total || 0
          
          console.log('获取到数据:', '新数据条数=', newTasks.length, '总数=', this.total)
          
          if (loadMore) {
            // 追加新数据
            this.tasks = [...this.tasks, ...newTasks]
          } else {
            // 重置数据
            this.tasks = newTasks
          }
          
          // 判断是否还有更多数据 - 修复：比较总数与当前任务数量
          this.hasMoreData = this.tasks.length < this.total
          
          console.log('数据处理完成:', '当前数据总数=', this.tasks.length, '是否还有更多=', this.hasMoreData, '总记录数=', this.total)
          
          // 如果没有更多数据但是首次加载，确保设置为没有更多数据
          if (this.total === 0 || this.tasks.length >= this.total) {
            this.hasMoreData = false
            console.log('设置没有更多数据标志')
          }
          
          // 重置视频加载状态
          this.resetVideoLoadStates(loadMore)
          
          // 添加一个短暂延迟，让DOM更新后再执行下一步操作
          setTimeout(() => {
            // 在DOM更新后重新观察视频元素
            this.observeVideoElements();
            
            // 加载当前可见的视频
            this.loadVisibleVideos();
            
            // 数据加载完成后应用瀑布流布局
            this.updateMasonryLayout();
            
            // 重新设置无限滚动观察器
            this.setupScrollObserver();
            
          }, 100);
        })
        .catch(error => {
          console.error('获取灵感内容失败:', error)
          this.$message.error('获取灵感内容失败，请刷新页面重试')
        })
        .finally(() => {
          // 无论成功失败，都确保状态被重置
          console.log('请求完成，重置加载状态', '加载更多=', loadMore)
          if (loadMore) {
            this.loadingMore = false
          } else {
            this.loading = false
          }
          this.initialLoaded = true
        })
    },
    
    // 重置视频加载状态
    resetVideoLoadStates(loadMore = false) {
      if (!loadMore) {
        // 完全重置所有视频的加载状态
        this.videoLoadedMap = {}
        this.imageLoadedMap = {}
        
        // 清除活动视频列表
        this.activeVideos = []
        
        // 尝试暂停所有视频元素
        this.pauseAllVideos()
      } else {
        // 仅处理新加载的任务
        const startIdx = this.tasks.length - this.pageSize
        if (startIdx >= 0) {
          const newTasks = this.tasks.slice(startIdx)
          newTasks.forEach(task => {
            // 为新任务初始化加载状态
            if (!this.videoLoadedMap[task.id]) {
              this.$set(this.videoLoadedMap, task.id, false)
            }
            if (!this.imageLoadedMap[task.id]) {
              this.$set(this.imageLoadedMap, task.id, false)
            }
          })
        }
      }
    },
    
    // 获取直接视频URL
    getDirectVideoUrl(url) {
      if (!url) return ''
      const timestamp = Date.now()
      return getDirectFileUrl(url) + `&_t=${timestamp}`
    },
    
    // 检查视频格式添加多种格式支持
    getVideoSourceElements(url) {
      if (!url) return ''
      
      const baseUrl = this.getDirectVideoUrl(url)
      // 检查URL是否已经包含文件扩展名
      const hasExtension = /\.(mp4|webm|mov|ogg)(\?|$)/i.test(url)
      
      // 如果已经包含扩展名，就不添加其他格式
      if (hasExtension) {
        return `<source src="${baseUrl}" type="video/mp4">`
      }
      
      // 支持多种格式
      return `
        <source src="${baseUrl}" type="video/mp4">
        <source src="${baseUrl}" type="video/webm">
        <source src="${baseUrl}" type="video/ogg">
      `
    },
    
    // 尝试重新加载失败的视频
    retryLoadVideo(videoEl, url) {
      if (!videoEl || !url) return
      
      try {
        // 清除现有源
        while (videoEl.firstChild) {
          videoEl.removeChild(videoEl.firstChild)
        }
        
        // 添加多格式支持
        videoEl.innerHTML = this.getVideoSourceElements(url)
        
        // 确保正确的样式和属性
        videoEl.style.zIndex = "1"
        videoEl.style.transform = "translateZ(0)"
        videoEl.style.backgroundColor = "#000"
        videoEl.preload = "metadata"
        videoEl.load()
      } catch (error) {
        console.error('重新加载视频失败:', error)
      }
    },
    
    // 加载视频
    loadVideo(event, taskId, url) {
      // 防止事件冒泡
      event.stopPropagation && event.stopPropagation();
      
      // 如果该视频已经在加载中或已加载，不要重复加载
      if (this.videoLoadedMap[taskId]) return;
      
      // 模拟加载中状态
      this.$set(this.videoLoadedMap, taskId, 'loading');
      
      // 获取视频元素
      const videoContainer = this.$el.querySelector(`.inspiration-card[data-task-id="${taskId}"] .video-container`);
      const videoEl = videoContainer ? videoContainer.querySelector('video') : null;
      
      if (!videoEl) {
        console.error('找不到视频元素:', taskId);
        this.$set(this.videoLoadedMap, taskId, false);
        return;
      }
      
      // 检查并限制同时加载的视频数量
      if (this.activeVideos.length >= this.maxConcurrentVideos) {
        // 找出最早加载的视频并暂停它
        const oldestVideo = this.activeVideos.shift();
        try {
          if (oldestVideo && !oldestVideo.paused) {
            oldestVideo.pause();
          }
        } catch (e) {
          console.warn('暂停旧视频时出错:', e);
        }
      }
      
      // 更新视频源
      const sourceEl = videoEl.querySelector('source');
      if (sourceEl) {
        sourceEl.src = this.getDirectVideoUrl(url);
        videoEl.load(); // 重新加载视频
        
        // 添加到活动视频列表
        this.activeVideos.push(videoEl);
      } else {
        console.error('找不到视频源元素:', taskId);
        this.$set(this.videoLoadedMap, taskId, false);
      }
    },
    
    // 处理视频加载错误
    handleVideoError(event, taskId) {
      console.error('视频加载失败:', taskId, event);
      
      // 更新状态为加载失败
      this.$set(this.videoLoadedMap, taskId, false);
      
      // 从活动视频列表中移除
      const videoEl = event.target;
      const index = this.activeVideos.indexOf(videoEl);
      if (index !== -1) {
        this.activeVideos.splice(index, 1);
      }
      
      // 显示错误消息
      this.$message.error('视频加载失败，请稍后重试');
    },
    
    // 显示任务详情
    showTaskDetail(task) {
      if (!task || !task.id) return;
      // 跳转到详情页面
      this.$router.push({
        name: 'InspirationDetail',
        params: { id: task.id }
      });
    },
    
    // 获取任务类型标签类型
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
    
    // 获取任务类型文本
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
    
    // 预览图片
    previewImage(url) {
      this.previewUrl = url
      this.previewVisible = true
    },
    
    // 分页大小改变
    handleSizeChange(size) {
      this.pageSize = size
      this.fetchInspiration()
    },
    
    // 当前页改变
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchInspiration()
    },
    
    // 设置交叉观察器
    setupIntersectionObserver() {
      // 检查浏览器支持
      if ('IntersectionObserver' in window) {
        // 创建一个新的交叉观察器
        this.intersectionObserver = new IntersectionObserver((entries) => {
          entries.forEach(entry => {
            // 当视频卡片进入视口
            if (entry.isIntersecting && entry.intersectionRatio >= 0.5) {
              const cardElement = entry.target;
              const taskId = cardElement.dataset.taskId;
              const videoUrl = cardElement.dataset.videoUrl;
              
              // 如果视频尚未加载
              if (taskId && videoUrl && !this.videoLoadedMap[taskId]) {
                // 延迟一点加载，避免太多视频同时加载
                setTimeout(() => {
                  // 再次检查元素是否仍在视口内
                  if (this.isElementInViewport(cardElement)) {
                    this.loadVideo({ target: cardElement }, taskId, videoUrl);
                  }
                }, 300);
              }
            }
          });
        }, {
          root: null, // 使用视口作为参考
          rootMargin: '0px', // 边界范围
          threshold: 0.5 // 当50%的元素可见时触发
        });
      }
    },
    
    // 检查元素是否在视口内
    isElementInViewport(el) {
      if (!el) return false;
      const rect = el.getBoundingClientRect();
      return (
        rect.top >= 0 &&
        rect.left >= 0 &&
        rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
        rect.right <= (window.innerWidth || document.documentElement.clientWidth)
      );
    },
    
    // 观察视频元素
    observeVideoElements() {
      // 如果观察器存在
      if (this.intersectionObserver) {
        // 首先取消所有之前的观察
        this.intersectionObserver.disconnect();
        
        // 获取所有视频卡片
        this.tasks.forEach(task => {
          if (task.type === 'digital_human') {
            const cardElement = this.$el.querySelector(`.inspiration-card[data-task-id="${task.id}"]`);
            if (cardElement) {
              // 加入数据属性
              cardElement.dataset.taskId = task.id;
              cardElement.dataset.videoUrl = task.result_url;
              // 观察这个元素
              this.intersectionObserver.observe(cardElement);
            }
          }
        });
      }
    },
    
    // 加载当前可见的视频
    loadVisibleVideos() {
      // 延迟执行，确保DOM已更新
      setTimeout(() => {
        // 获取所有未加载的数字人卡片
        const visibleCards = [];
        
      this.tasks.forEach(task => {
        if (task.type === 'digital_human' && !this.videoLoadedMap[task.id]) {
          const cardElement = this.$el.querySelector(`.inspiration-card[data-task-id="${task.id}"]`);
          if (cardElement && this.isElementInViewport(cardElement)) {
              visibleCards.push({
                taskId: task.id,
                url: task.result_url,
                element: cardElement,
                rect: cardElement.getBoundingClientRect()
              });
          }
        }
      });
        
        // 按照在视口中的位置从上到下排序
        visibleCards.sort((a, b) => a.rect.top - b.rect.top);
        
        // 限制同时加载的视频数量
        const maxInitialLoad = this.isMobile ? 2 : 3;
        visibleCards.slice(0, maxInitialLoad).forEach((card, index) => {
          // 添加延迟，避免同时加载太多视频
          setTimeout(() => {
            this.loadVideo({ target: card.element }, card.taskId, card.url);
          }, index * 300); // 每300ms加载一个
        });
      }, 500);
    },
    
    // 处理图片加载完成事件
    handleImageLoaded(taskId) {
      // 更新图片加载状态
      this.$set(this.imageLoadedMap, taskId, true);
      
      // 在nextTick中执行，确保DOM已更新
      this.$nextTick(() => {
        // 查找相关元素
        const card = this.$el.querySelector(`.inspiration-card[data-task-id="${taskId}"]`);
        const placeholder = card ? card.querySelector('.image-placeholder') : null;
        
        // 隐藏占位符
        if (placeholder) {
          placeholder.style.display = 'none';
        }
        
        // 显示卡片
        if (card) {
          card.style.opacity = '1';
        }
        
        // 触发布局更新
        this.updateMasonryLayout()
      });
    },
    
    // 图片加载失败处理
    handleImageError(taskId) {
      console.error('图片加载失败:', taskId)
      this.$set(this.imageLoadedMap, taskId, false)
    },
    
    // 清理不活跃的视频以节省资源
    cleanupInactiveVideos() {
      const now = Date.now();
      
      // 获取所有已加载的视频
      const videoElements = this.$el.querySelectorAll('video');
      
      videoElements.forEach(video => {
        // 如果视频暂停且最后播放时间超过1分钟，或者完全不在视口中
        if ((video.paused && video.dataset.lastPlayed && now - video.dataset.lastPlayed > 60000) || 
            !this.isElementInViewport(video)) {
          const taskId = video.dataset.id;
          if (taskId) {
            // 标记视频未加载
            this.$set(this.videoLoadedMap, taskId, false);
            
            // 移除视频源，释放资源
            while (video.firstChild) {
              video.removeChild(video.firstChild);
            }
            
            // 创建一个空的source元素
            const emptySource = document.createElement('source');
            emptySource.src = '';
            video.appendChild(emptySource);
            
            console.log('清理不活跃视频:', taskId);
            
            // 显示占位符
            const card = video.closest('.inspiration-card');
            if (card) {
              const placeholder = card.querySelector('.video-placeholder');
              if (placeholder) {
                placeholder.style.display = 'flex';
              }
            }
          }
        }
      });
    },
    
    // 在视频播放时更新最后播放时间
    handleVideoPlay(taskId, event) {
      const video = event.target;
      if (video) {
        // 更新最后播放时间
        video.dataset.lastPlayed = Date.now();
        
        // 如果正在播放此视频，暂停其他所有视频
        if (!video.paused) {
          this.pauseOtherVideos(taskId);
        }
      }
    },
    
    // 暂停除指定ID外的所有视频
    pauseOtherVideos(excludeTaskId) {
      const videos = document.querySelectorAll('video.preview-video');
      videos.forEach(video => {
        const videoTaskId = video.dataset.id;
        if (videoTaskId !== excludeTaskId && !video.paused) {
          video.pause();
        }
      });
    },
    
    // 暂停所有视频
    pauseAllVideos() {
      const videos = document.querySelectorAll('video.preview-video');
      videos.forEach(video => {
        if (!video.paused) {
          video.pause();
        }
      });
    },
    
    // 初始化瀑布流布局
    initializeMasonryLayout() {
      // 在nextTick中执行，确保DOM已更新
      this.$nextTick(() => {
        // 获取所有卡片
        const cards = this.$el.querySelectorAll('.inspiration-card')
        
        // 设置卡片初始状态
        cards.forEach(card => {
          card.style.opacity = '0.6'
          card.style.transition = 'opacity 0.3s ease, transform 0.3s ease'
        })
        
        // 调整网格边距
        this.adjustGridMargins()
        
        // 加载可见图片
        this.loadVisibleImages()
        
        // 更新布局
        this.updateMasonryLayout()
      })
    },
    
    // 加载可见区域内的图片
    loadVisibleImages() {
      // 获取可视区域高度
      const viewportHeight = window.innerHeight;
      
      // 获取所有图片
      const images = this.$el.querySelectorAll('.preview-image');
      
      // 遍历图片
      images.forEach(img => {
        // 获取图片位置
        const rect = img.getBoundingClientRect();
        
        // 如果图片在可视区域内或附近
        if (rect.top < viewportHeight + 300) {
          // 加载图片
          const card = img.closest('.inspiration-card');
          if (card) {
            const taskId = card.dataset.taskId;
            if (taskId) {
              this.$set(this.imageLoadedMap, taskId, true);
            }
          }
          
          // 设置图片完全不透明
          if (card) {
            setTimeout(() => {
              card.style.opacity = '1';
            }, 100);
          }
        }
      });
    },
    
    // 应用卡片布局
    applyCardLayout() {
      // 添加样式到head
      const styleEl = document.createElement('style')
      styleEl.textContent = `
        .inspiration-grid {
          position: relative;
          width: 100%;
          min-height: 200px;
        }
        .inspiration-card {
          position: absolute;
          width: calc(25% - 20px);
          margin: 0;
          transition: transform 0.3s ease, box-shadow 0.3s ease, opacity 0.5s ease;
          opacity: 0;
          background-color: #f8f8f8;
          border-radius: 8px;
          overflow: hidden;
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
          cursor: pointer;
          will-change: transform, opacity, left, top;
        }
        .inspiration-card:hover {
          transform: translateY(-5px);
          box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
          z-index: 10;
        }
        .card-media {
          width: 100%;
          height: auto !important;
          position: relative;
        }
        .image-container {
          width: 100%;
          height: auto !important;
          background-color: #fff;
          position: relative;
          overflow: hidden;
        }
        .preview-image {
          width: 100%;
          height: auto !important;
          display: block;
          max-width: 100%;
          transition: transform 0.4s ease;
        }
        .inspiration-card:hover .preview-image {
          transform: scale(1.03);
        }
        @media (min-width: 993px) and (max-width: 1400px) {
          .inspiration-card {
            width: calc(25% - 20px);
          }
        }
        @media (min-width: 577px) and (max-width: 992px) {
          .inspiration-card {
            width: calc(33.333% - 15px);
          }
        }
        @media (max-width: 576px) {
          .inspiration-container {
            padding: 8px 4px; /* 减小移动端的容器边距 */
          }
          
          .page-header {
            margin-bottom: 10px;
          }
          
          .global-controls {
            margin-bottom: 8px; /* 减小控制按钮下边距 */
            flex-direction: column;
            align-items: flex-end;
            gap: 5px;
          }
          
          .inspiration-card {
            width: calc(50% - 2px) !important; /* 移动端卡片占据更多空间 */
            margin-bottom: 4px !important; /* 减小卡片间垂直间距 */
            border-radius: 6px; /* 稍微减小圆角 */
            box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
          }
          
          /* 移动端视频控制 */
          .video-control-overlay {
            height: 40px;
            opacity: 0.8;
          }
          
          .preview-video::-webkit-media-controls {
            opacity: 1 !important;
          }
          
          /* 移动端减小任务类型标签尺寸 */
          .task-type-tag {
            top: 3px;
            right: 3px;
            padding: 0;
          }
          
          .task-type-tag .el-tag--mini {
            height: 18px;
            line-height: 18px;
            padding: 0 4px;
            font-size: 10px;
          }
          
          /* 移动端预览对话框 */
          .preview-dialog {
            width: 95% !important;
            margin: 0 !important;
          }
          
          /* 详情对话框移动端调整 */
          .task-detail {
            padding: 0 5px;
          }
          
          /* 详情页移动端适配 */
          .params-grid {
            grid-template-columns: 1fr;
          }
          
          .detail-meta {
            flex-direction: column;
            gap: 10px;
          }
          
          /* 减小移动端的分页控件空间 */
          .pagination-container {
            margin-top: 10px;
          }
          
          /* 对瀑布流布局进行移动端优化 */
          .inspiration-grid {
            margin: 0 auto !important;
          }
          
          /* 移动端视频占位符优化 */
          .video-placeholder .loading-indicator {
            transform: scale(0.8);
          }
          
          .video-placeholder .loading-indicator i {
            font-size: 28px;
            margin-bottom: 4px;
          }
          
          .video-placeholder .loading-indicator span {
            font-size: 12px;
          }
          
          /* 移动端卡片悬停效果优化 */
          .inspiration-card:hover {
            transform: translateY(-2px) !important;
          }
        }
      `
      document.head.appendChild(styleEl)
      
      // 初始化瀑布流布局
      this.updateMasonryLayout()
      
      // 在滚动时更新布局
      window.addEventListener('scroll', this.debounce(this.updateMasonryLayout, 300))
      
      // 在加载图片时更新布局
      document.querySelectorAll('.preview-image').forEach(img => {
        img.onload = () => {
          const card = img.closest('.inspiration-card')
          if (card) {
            card.style.opacity = '1'
            // 图片加载完成后重新计算布局
            this.updateMasonryLayout()
          }
        }
      })
    },
    
    // 防抖函数
    debounce(fn, delay) {
      let timer = null
      return function() {
        const context = this
        const args = arguments
        clearTimeout(timer)
        timer = setTimeout(() => {
          fn.apply(context, args)
        }, delay)
      }
    },
    
    // 更新真正的瀑布流布局
    updateMasonryLayout() {
      // 使用requestAnimationFrame确保在下一帧执行
      window.requestAnimationFrame(() => {
        const grid = this.$el.querySelector('.inspiration-grid')
        if (!grid) return
        
        // 获取卡片和网格容器尺寸
        const cards = Array.from(grid.querySelectorAll('.inspiration-card'))
        if (cards.length === 0) return // 如果没有卡片则退出
        
        const containerWidth = grid.clientWidth
        
        // 确定列数
        let columns = 4
        if (window.innerWidth <= 576) {
          columns = 2
        } else if (window.innerWidth <= 992) {
          columns = 3
        }
        
        // 计算卡片宽度和间距
        let margin = 10
        if (window.innerWidth >= 1401) {
          margin = 15
        } else if (window.innerWidth <= 576) {
          margin = 2 // 移动端更小的间距
        }
        
        const cardWidth = (containerWidth - (margin * 2 * columns)) / columns
        
        // 初始化列高度数组和位置缓存
        const columnsHeights = Array(columns).fill(0)
        const cardPositions = {} // 用于记录卡片之前的位置
        
        // 首先收集所有卡片当前位置
        cards.forEach(card => {
          const taskId = card.dataset.taskId
          if (taskId) {
            cardPositions[taskId] = {
              left: card.style.left || '0px',
              top: card.style.top || '0px'
            }
          }
        })
        
        // 为每张卡片定位
        cards.forEach(card => {
          // 确保卡片已渲染且已加载
          if (card.offsetHeight === 0) return
          
          const taskId = card.dataset.taskId
          
          // 找出最短的列
          const minColumnIndex = columnsHeights.indexOf(Math.min(...columnsHeights))
          
          // 计算卡片位置
          const x = minColumnIndex * (cardWidth + margin * 2)
          const y = columnsHeights[minColumnIndex]
          
          // 设置卡片宽度
          card.style.width = `${cardWidth}px`
          
          // 防止不必要的动画，仅当位置确实改变时才更新
          const newLeft = `${x}px`
          const newTop = `${y}px`
          
          const hasPositionChanged = !taskId || 
                                  !cardPositions[taskId] || 
                                  cardPositions[taskId].left !== newLeft || 
                                  cardPositions[taskId].top !== newTop
          
          if (hasPositionChanged) {
            // 应用动画过渡效果
            card.style.transition = 'transform 0.3s ease, box-shadow 0.3s ease, opacity 0.5s ease, left 0.3s ease, top 0.3s ease'
            
            // 设置卡片位置
            card.style.left = newLeft
            card.style.top = newTop
          } else {
            // 如果位置没有变化，移除过渡效果以避免闪烁
            card.style.transition = 'transform 0.3s ease, box-shadow 0.3s ease, opacity 0.5s ease'
          }
          
          // 确保卡片显示
          card.style.opacity = '1'
          
          // 添加随机微调效果（仅在非移动端），并限制最大值，防止过大的位移
          if (window.innerWidth > 576) {
            const randomDelay = Math.random() * 0.15
            card.style.transitionDelay = `${randomDelay}s`
          } else {
            card.style.transitionDelay = '0s' // 移动端不添加延迟
          }
          
          // 更新列高度
          columnsHeights[minColumnIndex] += card.offsetHeight + margin
        })
        
        // 计算最大高度
        const maxColumnHeight = Math.max(...columnsHeights)
        
        // 设置网格容器高度，确保有额外空间
        const maxHeight = maxColumnHeight + 100 // 添加更多额外空间
        grid.style.height = `${maxHeight}px`
        
        // 更新加载更多触发器位置
        this.updateLoadMoreTriggerPosition()
        
        // 在所有卡片布局完成后，设置视频自动播放观察
        this.$nextTick(() => {
          this.observeVideosForAutoPlay()
        })
      })
    },
    
    // 调整网格边距以适应容器宽度
    adjustGridMargins() {
      const grid = this.$el.querySelector('.inspiration-grid');
      const container = this.$el.querySelector('.inspiration-container');
      
      if (!grid || !container) return;
      
      // 根据屏幕宽度决定卡片边距
      let cardMargin = 10; // 默认边距
      
      if (window.innerWidth >= 1401) {
        cardMargin = 15;
      } else if (window.innerWidth >= 993) {
        cardMargin = 10;
      } else if (window.innerWidth >= 577) {
        cardMargin = 7.5;
      } else {
        cardMargin = 5;
      }
      
      // 设置网格负边距
      grid.style.margin = `0 -${cardMargin}px`;
      grid.style.width = `calc(100% + ${cardMargin * 2}px)`;
    },
    
    // 清理所有观察器
    cleanupObservers() {
      // 如果有交叉观察器，断开连接
      if (this.intersectionObserver) {
        this.intersectionObserver.disconnect()
        this.intersectionObserver = null
      }
      
      // 如果有视频观察器，断开连接
      if (this.videoIntersectionObserver) {
        this.videoIntersectionObserver.disconnect()
        this.videoIntersectionObserver = null
      }
      
      // 清理无限滚动观察器
      if (this.observer) {
        this.observer.disconnect()
        this.observer = null
      }
    },
    
    // 检测设备类型
    checkDeviceType() {
      const userAgent = navigator.userAgent || navigator.vendor || window.opera;
      const isMobileByUA = /android|webos|iphone|ipad|ipod|blackberry|iemobile|opera mini|mobile/i.test(userAgent);
      
      // 屏幕宽度检测（作为备选）
      const isMobileByWidth = window.innerWidth <= 576;
      
      // 更新移动设备状态
      this.isMobile = isMobileByUA || isMobileByWidth;
    },
    
    // 返回上一页
    goBack() {
      this.$router.go(-1);
    },
    
    // 处理窗口大小变化
    handleWindowResize() {
      // 更新设备类型
      this.checkDeviceType();
      
      // 防抖处理布局更新
      if (this.resizeTimer) {
        clearTimeout(this.resizeTimer);
      }
      
      this.resizeTimer = setTimeout(() => {
        this.updateMasonryLayout();
      }, 300);
    },
    
    // 加载更多数据
    loadMoreTasks() {
      // 添加更详细的日志以排查问题
      console.log('尝试加载更多数据，状态:', 
        'loadingMore=', this.loadingMore, 
        'hasMoreData=', this.hasMoreData, 
        'detailVisible=', this.detailVisible,
        'pauseScrollListening=', this.pauseScrollListening,
        'currentPage=', this.currentPage,
        'tasks.length=', this.tasks.length,
        'total=', this.total)
      
      if (this.loadingMore || !this.hasMoreData || this.detailVisible || this.pauseScrollListening) {
        console.log('跳过加载更多:', 
          '加载中=', this.loadingMore, 
          '没有更多数据=', !this.hasMoreData, 
          '详情页打开=', this.detailVisible,
          '暂停滚动监听=', this.pauseScrollListening)
        return
      }
      
      console.log('加载更多数据，当前页码:', this.currentPage, '即将请求页码:', this.currentPage + 1)
      this.currentPage++
      this.fetchInspiration(true)
    },
    
    // 处理滚动事件 - 添加更多调试信息
    handleWindowScroll() {
      // 如果滚动监听被暂停或详情页打开，则不处理
      if (this.pauseScrollListening || this.detailVisible) {
        return
      }
      
      // 记录滚动方向
      const currentScrollTop = window.pageYOffset || document.documentElement.scrollTop
      const scrollingDown = currentScrollTop > this.lastScrollTop
      this.lastScrollTop = currentScrollTop
      
      // 加载更多数据条件
      if (scrollingDown && !this.loadingMore && this.hasMoreData) {
        const scrollTop = window.pageYOffset || document.documentElement.scrollTop
        const windowHeight = window.innerHeight
        const documentHeight = Math.max(
          document.body.scrollHeight, document.documentElement.scrollHeight,
          document.body.offsetHeight, document.documentElement.offsetHeight,
          document.body.clientHeight, document.documentElement.clientHeight
        )
        
        // 计算距离底部的距离
        const distanceToBottom = documentHeight - scrollTop - windowHeight
        
        // 当滚动到距离底部阈值距离时触发加载
        if (distanceToBottom < this.scrollThreshold) {
          console.log('触发滚动加载更多:', '滚动位置=', scrollTop, '窗口高度=', windowHeight, 
                    '文档高度=', documentHeight, '距离底部=', distanceToBottom,
                    '阈值=', this.scrollThreshold, '有更多数据=', this.hasMoreData)
          this.loadMoreTasks()
        }
      }
    },
    
    // 设置无限滚动观察器
    setupScrollObserver() {
      // 如果已经有observer，先断开连接
      if (this.observer) {
        this.observer.disconnect()
        this.observer = null
      }
      
      this.$nextTick(() => {
        // 获取网格容器
        const grid = this.$el.querySelector('.inspiration-grid')
        if (!grid) {
          console.error('无法找到网格容器元素')
          return
        }
        
        // 获取或创建一个用于检测底部的元素
        let loadMoreTrigger = this.$refs.loadMoreTrigger
        if (!loadMoreTrigger) {
          console.error('无法找到加载更多触发器元素')
          return
        }
        
        // 确保已有的触发器有正确的样式
        loadMoreTrigger.style.height = '50px'
        loadMoreTrigger.style.zIndex = '2'
        
        // 检查是否还有更多数据
        if (!this.hasMoreData) {
          loadMoreTrigger.style.display = 'none'
          console.log('没有更多数据，隐藏加载触发器')
          return
        } else {
          loadMoreTrigger.style.display = 'block'
        }
        
        console.log('设置无限滚动观察器', '是否有更多数据=', this.hasMoreData)
        
        // 创建新的IntersectionObserver
        this.observer = new IntersectionObserver((entries) => {
          const entry = entries[0]
          const isVisible = entry.isIntersecting
          
          console.log('触发器元素交叉状态:', 
                      '可见=', isVisible, 
                      '加载中=', this.loadingMore, 
                      '有更多数据=', this.hasMoreData, 
                      '交叉比例=', entry.intersectionRatio)
          
          // 当触发元素进入视口且满足条件时加载更多
          if (isVisible && !this.loadingMore && this.hasMoreData && !this.detailVisible && !this.pauseScrollListening) {
            console.log('观察器触发加载更多')
            this.loadMoreTasks()
          }
        }, {
          root: null,
          threshold: 0.1, // 降低阈值，使其更容易触发
          rootMargin: '300px' // 增大触发距离
        })
        
        // 开始观察
        this.observer.observe(loadMoreTrigger)
        console.log('开始观察加载更多触发器元素')
        
        // 确保触发器位于正确位置
        this.$nextTick(() => {
          this.updateLoadMoreTriggerPosition()
        })
      })
    },
    
    // 清理所有资源
    cleanup() {
      // 暂停所有视频
      this.pauseAllVideos()
      
      // 清理活动视频列表
      this.activeVideos = []
      
      // 清理加载状态
      this.videoLoadedMap = {}
      this.imageLoadedMap = {}
      
      // 暂停滚动监听
      this.pauseScrollListening = true
    },
    
    // 处理页面可见性变化
    handleVisibilityChange() {
      if (document.hidden) {
        // 页面隐藏时，暂停所有视频
        this.pauseAllVideos();
        
        // 暂停滚动监听
        this.pauseScrollListening = true;
        
        console.log('页面不可见，暂停视频和滚动监听');
      } else {
        // 页面可见时，恢复滚动监听
        this.pauseScrollListening = false;
        
        // 重新设置观察器
        this.setupIntersectionObserver();
        this.setupScrollObserver();
        this.observeVideosForAutoPlay();
        
        console.log('页面可见，恢复滚动监听和观察');
      }
    },
    
    // 刷新数据
    refreshData() {
      console.log('手动刷新数据')
      // 重置所有状态
      this.currentPage = 1
      this.tasks = []
      this.hasMoreData = true
      this.initialLoaded = false
      this.loading = false
      this.loadingMore = false
      this.total = 0
      
      // 重置视频状态
      this.videoLoadedMap = {}
      this.imageLoadedMap = {}
      this.activeVideos = []
      
      // 暂停所有视频
      this.pauseAllVideos()
      
      // 清理观察器并重新设置
      this.cleanupObservers()
      
      // 立即获取数据
      this.$nextTick(() => {
        console.log('刷新数据：开始获取新数据')
        this.fetchInspiration()
        
        // 重新设置观察器
        setTimeout(() => {
          this.setupIntersectionObserver()
          this.setupScrollObserver()
          this.setupVideoPlayObserver()
        }, 500) // 延迟一点时间确保DOM更新
      })
    },
    
    // 更新加载更多触发器的位置
    updateLoadMoreTriggerPosition() {
      const grid = this.$el.querySelector('.inspiration-grid')
      const loadMoreTrigger = this.$refs.loadMoreTrigger
      
      if (!grid || !loadMoreTrigger) return
      
      // 计算网格容器高度
      const cardsHeight = Array.from(this.$el.querySelectorAll('.inspiration-card'))
        .reduce((maxHeight, card) => {
          const bottom = card.offsetTop + card.offsetHeight
          return Math.max(maxHeight, bottom)
        }, 0)
      
      // 确保网格高度足够
      const gridHeight = Math.max(cardsHeight + 100, 200) // 至少200px高
      grid.style.height = `${gridHeight}px`
      
      // 设置触发器位置在内容底部
      loadMoreTrigger.style.top = `${cardsHeight}px`
      loadMoreTrigger.style.bottom = 'auto'
      
      console.log('更新加载触发器位置:', 
                  '卡片总高度=', cardsHeight, 
                  '网格高度=', gridHeight, 
                  '触发器位置=', `${cardsHeight}px`)
    },
  }
}
</script>

<style scoped>
/* 视频控制覆盖层样式 */
.video-control-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 80px; /* 增大控制区高度 */
  z-index: 3;
  background: linear-gradient(transparent, rgba(0,0,0,0.7)); /* 增强渐变效果 */
  opacity: 0;
  transition: opacity 0.3s;
  pointer-events: none;
}

.video-container:hover .video-control-overlay {
  opacity: 1;
}

/* 视频控制样式 */
.preview-video::-webkit-media-controls {
  opacity: 0;
  transition: opacity 0.3s ease;
}

.inspiration-card:hover .preview-video::-webkit-media-controls {
  opacity: 1;
}

/* 容器样式 */
.inspiration-container {
  padding: 20px;
  max-width: 1600px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
  overflow: hidden; /* 防止内容溢出 */
}

/* 标题样式 - 参考TTS模块 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 12px 15px;
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.header-left {
  display: flex;
  align-items: center;
}

.back-icon {
  font-size: 20px;
  margin-right: 10px;
  cursor: pointer;
  color: #409EFF;
}

.header-left h2 {
  font-size: 1.4rem;
  margin: 0;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-right p {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

/* 移动端标题栏样式 */
.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 999;
  border-radius: 0;
  padding: 8px 10px;
  margin: 0;
  background: rgba(156, 16, 16, 0.1);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.mobile-header-placeholder {
  height: 50px;
  width: 100%;
  margin: 0;
  padding: 0;
}

/* 瀑布流布局 */
.inspiration-grid {
  position: relative;
  width: 100%;
  min-height: 200px;
  margin: 0 auto;
}

/* 卡片样式 */
.inspiration-card {
  position: absolute;
  width: calc(25% - 20px);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease, box-shadow 0.3s ease, opacity 0.5s ease;
  cursor: pointer;
  background-color: #f8f8f8;
  box-sizing: border-box;
  opacity: 0; /* 初始不可见，等待加载 */
  will-change: transform, opacity; /* 优化动画性能 */
}

.inspiration-card:hover {
  transform: translateY(-5px) !important;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  z-index: 10 !important; /* 确保悬停卡片在其他卡片上方 */
}

/* 卡片内容容器 */
.card-media {
  width: 100%;
  position: relative;
  height: auto !important;
}

/* 图片预览容器 */
.image-container {
  width: 100%;
  background-color: #fff;
  position: relative;
  height: auto !important;
  overflow: hidden;
}

/* 图片预览 */
.preview-image {
  width: 100%;
  height: auto !important;
  display: block;
  object-fit: contain;
  max-width: 100%;
  transition: transform 0.4s ease;
}

/* 图片悬停效果 */
.inspiration-card:hover .preview-image {
  transform: scale(1.03);
}

/* 视频容器 */
.video-container {
  width: 100%;
  position: relative;
  background-color: #000;
  padding-top: 133.33%; /* 固定3:4比例 */
  height: 0;
  overflow: hidden;
}

/* 视频占位符 */
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
  transition: background-color 0.3s;
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
  transition: transform 0.3s;
}

.video-placeholder:hover .loading-indicator {
  transform: scale(1.1);
}

.loading-indicator i {
  font-size: 36px;
  margin-bottom: 8px;
  color: #409EFF;
}

/* 加载更多触发器 - 修改样式使其更明显 */
.load-more-trigger {
  width: 100%;
  height: 50px; /* 增加高度 */
  position: absolute;
  bottom: 0;
  left: 0;
  z-index: 2;
  background: rgba(64, 158, 255, 0.1); /* 轻微背景色以便调试 */
  border-top: 1px dashed rgba(64, 158, 255, 0.3); /* 添加虚线边框 */
  text-align: center;
  line-height: 50px;
  font-size: 0; /* 正常情况下隐藏文本 */
}

/* 加载更多容器 */
.load-more-container {
  text-align: center;
  padding: 20px 0;
  color: #909399;
  font-size: 14px;
}

.load-more-container .loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #606266;
}

.load-more-container .loading-indicator i {
  font-size: 24px;
  margin-bottom: 10px;
  color: #409EFF;
}

.load-more-container p {
  margin: 5px 0;
}

/* 响应式布局调整 */
@media (min-width: 1401px) {
  .inspiration-card {
    width: calc(25% - 30px);
  }
  
  .inspiration-container {
    max-width: 1600px;
    padding: 30px;
  }
}

@media (min-width: 993px) and (max-width: 1400px) {
  .inspiration-card {
    width: calc(25% - 20px);
  }
  
  .inspiration-container {
    max-width: 1200px;
    padding: 25px;
  }
}

@media (min-width: 577px) and (max-width: 992px) {
  .inspiration-card {
    width: calc(33.333% - 15px);
  }
  
  .inspiration-container {
    max-width: 900px;
    padding: 20px;
  }
  
  .preview-video::-webkit-media-controls {
    opacity: 1 !important;
  }
  
  .video-control-overlay {
    height: 70px;
    opacity: 0.7; /* 平板端默认显示一定的控制层 */
  }
}

@media (max-width: 576px) {
  .inspiration-container {
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
  
  .mobile-header {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 999;
    border-radius: 0;
    padding: 8px 10px;
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
  
  .mobile-header .header-left h2 {
    font-size: 18px;
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: bold;
  }
  
  .mobile-header-placeholder {
    height: 52px;
    margin-bottom: 8px;
  }
  
  .inspiration-card {
    width: calc(50% - 6px) !important; /* 移动端卡片占据更多空间 */
    margin-bottom: 4px !important; /* 减小卡片间垂直间距 */
    border-radius: 6px; /* 稍微减小圆角 */
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  }
  
  /* 移动端视频控制 */
  .video-control-overlay {
    height: 40px;
    opacity: 0.8;
  }
  
  .preview-video::-webkit-media-controls {
    opacity: 1 !important;
  }
  
  /* 移动端减小任务类型标签尺寸 */
  .task-type-tag {
    top: 3px;
    right: 3px;
    padding: 0;
  }
  
  .task-type-tag .el-tag--mini {
    height: 18px;
    line-height: 18px;
    padding: 0 4px;
    font-size: 10px;
  }
  
  /* 移动端预览对话框 */
  .preview-dialog {
    width: 95% !important;
    margin: 0 !important;
  }
  
  /* 详情对话框移动端调整 */
  .task-detail {
    padding: 0 5px;
  }
  
  /* 详情页移动端适配 */
  .params-grid {
    grid-template-columns: 1fr;
  }
  
  .detail-meta {
    flex-direction: column;
    gap: 10px;
  }
  
  /* 加载更多移动端样式 */
  .load-more-container {
    margin-bottom: 70px; /* 防止底部菜单遮挡 */
    padding: 15px 0;
  }
  
  /* 对瀑布流布局进行移动端优化 */
  .inspiration-grid {
    margin: 8px !important;
    padding-bottom: 20px; /* 减小底部空间，由load-more-container提供 */
  }
  
  /* 移动端视频占位符优化 */
  .video-placeholder .loading-indicator {
    transform: scale(0.8);
  }
  
  .video-placeholder .loading-indicator i {
    font-size: 28px;
    margin-bottom: 4px;
  }
  
  .video-placeholder .loading-indicator span {
    font-size: 12px;
  }
  
  /* 移动端卡片悬停效果优化 */
  .inspiration-card:hover {
    transform: translateY(-2px) !important;
  }
  
  /* 修复iOS移动端滑动问题 */
  .inspiration-grid {
    -webkit-overflow-scrolling: touch;
  }
}

.image-placeholder {
  position: relative;
  width: 100%;
  padding-bottom: 75%; /* 默认4:3比例 */
  background-color: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.loading-indicator.small i {
  font-size: 24px;
  margin-bottom: 4px;
}

.image-placeholder .loading-indicator i {
  color: #909399;
}

.task-type-tag {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 3;
  background-color: rgba(255, 255, 255, 0.8);
  border-radius: 4px;
  padding: 2px;
}

.preview-dialog {
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-fullsize {
  max-width: 100%;
  max-height: 80vh;
  object-fit: contain;
}

/* 加载容器 */
.loading-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

/* 视频预览样式 */
.preview-video {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover; /* 充满整个容器 */
  background-color: #000;
  z-index: 2;
}

/* 添加自定义样式处理视频任务卡片 */
.inspiration-card[data-task-id] {
  overflow: hidden;
  background-color: #000;
}

/* 详情弹窗样式 */
.task-detail {
  padding: 0 10px;
}

.detail-header {
  text-align: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
}

.detail-header h2 {
  margin: 0 0 10px;
  font-size: 22px;
  color: #303133;
}

.detail-header p {
  color: #606266;
  font-size: 14px;
  margin: 0;
}

.detail-preview {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #ebeef5;
  position: relative;
}

.detail-preview h3 {
  margin: 0 0 10px;
  font-size: 16px;
  color: #303133;
}

.full-image {
  width: 100%;
  max-height: 400px;
  object-fit: contain;
  cursor: pointer;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  background-color: #f8f8f8;
}

.full-video {
  width: 100%;
  max-height: 400px;
  border-radius: 4px;
  background-color: #000;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 1;
  object-fit: contain;
  filter: brightness(1.2);
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

.detail-params {
  margin-bottom: 20px;
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

.param-item {
  background-color: #f9f9f9;
  border-radius: 6px;
  padding: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.param-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
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
}

.param-video {
  position: relative;
  min-height: 100px;
}

.thumbnail-video {
  width: 100%;
  max-height: 120px;
  border-radius: 4px;
  background-color: #000;
  z-index: 1;
  object-fit: contain;
  filter: brightness(1.2);
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
</style>

