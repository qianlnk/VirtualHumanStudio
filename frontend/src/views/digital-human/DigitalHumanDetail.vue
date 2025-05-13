<template>
  <div class="digital-human-detail-container">
    <!-- PC端顶部导航栏 -->
    <div class="page-header">
      <h2>数字人合成任务详情</h2>
      <div>
        <el-button type="primary" @click="goBack">返回列表</el-button>
      </div>
    </div>
    
    <!-- 移动端顶部导航栏 - 只在移动端显示 -->
    <div class="mobile-header-bar" v-show="isMobile">
      <div class="header-back" @click="goBack">
        <i class="el-icon-arrow-left"></i>
        <span>返回</span>
      </div>
      <h2 class="header-title">数字人合成任务</h2>
    </div>
    
    <!-- 移动端头部占位 - 只在移动端显示 -->
    <div class="mobile-header-placeholder" v-show="isMobile"></div>
    
    <div class="detail-content-wrapper">
      <div v-loading="loading" class="detail-content">
        <!-- PC端展示 -->
        <div class="desktop-content-view">
          <el-card v-if="digitalHuman">
            <div slot="header" class="card-header">
              <span>{{ digitalHuman.name }}</span>
              <el-tag :type="getStatusType(digitalHuman.status)" class="status-tag">{{ getStatusText(digitalHuman.status) }}</el-tag>
            </div>
            
            <div class="task-info">
              <div class="info-item">
                <span class="label">创建时间：</span>
                <span>{{ formatDate(digitalHuman.created_at) }}</span>
              </div>
              <div class="info-item">
                <span class="label">任务代码：</span>
                <span>{{ digitalHuman.task_code }}</span>
              </div>
              <div class="info-item" v-if="digitalHuman.description">
                <span class="label">任务描述：</span>
                <span>{{ digitalHuman.description }}</span>
              </div>
              <div class="info-item">
                <span class="label">超分：</span>
                <span>{{ digitalHuman.chaofen ? '开启' : '关闭' }}</span>
              </div>
              <div class="info-item">
                <span class="label">水印：</span>
                <span>{{ digitalHuman.watermark_switch ? '开启' : '关闭' }}</span>
              </div>
              <div class="info-item">
                <span class="label">PN值：</span>
                <span>{{ digitalHuman.pn }}</span>
              </div>
            </div>
            
            <!-- 音频和视频预览 -->
            <div class="media-section">
              <h3>输入文件</h3>
              <div class="media-preview">
                <div class="audio-preview">
                  <h4>音频文件</h4>
                  <audio controls style="width: 100%" ref="audioPlayer">
                    <source :src="audioUrl" type="audio/wav">
                    您的浏览器不支持音频播放
                  </audio>
                </div>
                <div class="video-preview" :class="{'video-loaded': videoUrl && videoLoaded, 'video-loading': !videoUrl || !videoLoaded}">
                  <h4>原始视频文件</h4>
                  <div v-if="videoUrl" class="video-container">
                    <video 
                      controls 
                      style="width: 100%; max-height: 300px; background-color: #000; z-index: 0 !important; transform: translateZ(0);" 
                      ref="videoPlayer"
                      class="desktop-video"
                      data-video-type="original"
                      muted
                      playsinline
                      preload="metadata">
                      <source :src="videoUrl" type="video/mp4">
                      您的浏览器不支持视频播放
                    </video>
                  </div>
                  <div v-else class="video-placeholder">
                    <el-skeleton :loading="true" animated>
                      <template slot="template">
                        <el-skeleton-item variant="rect" style="width: 100%; height: 300px; background-color: #f4f4f4;"/>
                      </template>
                    </el-skeleton>
                    <div class="placeholder-text">加载原始视频中...</div>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- 合成结果 -->
            <div class="result-section">
              <h3>合成结果</h3>
              <!-- 有结果视频时显示视频播放器 -->
              <div v-if="digitalHuman && digitalHuman.status === 'completed' && digitalHuman.result_url && resultUrl" 
                  class="result-preview" :class="{'video-loaded': resultUrl && resultLoaded}">
                <video 
                  controls 
                  style="width: 100%; max-height: 400px; background-color: #000; z-index: 0 !important; transform: translateZ(0);" 
                  ref="resultPlayer"
                  class="desktop-video"
                  data-video-type="result"
                  muted
                  playsinline
                  preload="metadata">
                  <source :src="resultUrl" type="video/mp4">
                  您的浏览器不支持视频播放
                </video>
              </div>
              <!-- 根据不同状态显示不同内容 -->
              <div v-else class="result-preview empty-result" :class="{'processing': digitalHuman && digitalHuman.status === 'processing'}">
                <template v-if="digitalHuman && digitalHuman.status === 'processing'">
                  <div class="processing-info">
                    <el-progress :percentage="progress" :format="progressFormat"></el-progress>
                    <p>视频正在处理中，请等待完成后查看结果</p>
                  </div>
                </template>
                <template v-else-if="digitalHuman && digitalHuman.status === 'failed'">
                  <div class="error-message">
                    处理失败: {{ digitalHuman.error_msg || '任务处理失败' }}
                  </div>
                </template>
                <template v-else>
                  <div class="waiting-info">
                    <p>尚未生成结果视频</p>
                  </div>
                </template>
              </div>
            </div>
            
            <!-- 下载结果按钮区域 - PC端 -->
            <div v-if="digitalHuman && digitalHuman.status === 'completed' && digitalHuman.result_url" class="download-section">
              <div class="download-container">
                <h3>下载视频</h3>
                <p class="download-tips">合成任务已完成，您可以下载生成的视频文件</p>
                <div class="action-buttons-group">
                <el-button type="primary" size="large" @click="downloadResult" class="download-button">
                  <i class="el-icon-download"></i> 下载合成结果
                </el-button>
                  <el-button type="success" size="large" @click="shareTask" class="share-button" :disabled="isShared">
                    <i class="el-icon-share"></i> {{ getShareButtonText() }}
                  </el-button>
                </div>
              </div>
            </div>
          </el-card>
          
          <el-empty v-else description="未找到任务信息"></el-empty>
        </div>
        
        <!-- 移动端展示 -->
        <div class="mobile-content-view">
          <div v-if="digitalHuman" class="mobile-content-inner">
            <!-- 基本信息 - 简化版 -->
            <div class="basic-info">
              <div class="status-tag">
                <el-tag :type="getStatusType(digitalHuman.status)">{{ getStatusText(digitalHuman.status) }}</el-tag>
              </div>
              <div class="create-time">创建时间：{{ formatDate(digitalHuman.created_at) }}</div>
            </div>
            
            <!-- 任务信息 -->
            <div class="mobile-task-info">
              <div class="info-item">
                <span class="label">任务代码：</span>
                <span>{{ digitalHuman.task_code }}</span>
              </div>
              <div class="info-item" v-if="digitalHuman.description">
                <span class="label">任务描述：</span>
                <span>{{ digitalHuman.description }}</span>
              </div>
              <div class="info-item">
                <span class="label">超分：</span>
                <span>{{ digitalHuman.chaofen ? '开启' : '关闭' }}</span>
              </div>
              <div class="info-item">
                <span class="label">水印：</span>
                <span>{{ digitalHuman.watermark_switch ? '开启' : '关闭' }}</span>
              </div>
              <div class="info-item">
                <span class="label">PN值：</span>
                <span>{{ digitalHuman.pn }}</span>
              </div>
            </div>
            
            <!-- 音频文件 -->
            <div class="mobile-section">
              <h4 class="section-title">音频文件</h4>
              <div v-if="audioUrl" class="audio-player">
                <audio controls ref="mobileAudioPlayer" style="width: 100%">
                  <source :src="audioUrl" type="audio/wav">
                  您的浏览器不支持音频播放
                </audio>
              </div>
              <div v-else class="loading-audio">
                <el-skeleton :loading="true" animated>
                  <template slot="template">
                    <el-skeleton-item variant="rect" style="width: 100%; height: 36px"/>
                  </template>
                </el-skeleton>
              </div>
            </div>
            
            <!-- 原始视频文件 -->
            <div class="mobile-section" id="mobile-original-video">
              <h4 class="section-title">原始视频文件</h4>
              <div v-if="videoUrl" class="video-player">
                <video 
                  ref="mobileVideoPlayer" 
                  controls
                  style="width: 100%; max-height: 200px; background-color: #000; display: block; z-index: 0 !important; transform: translateZ(0);" 
                  data-video-type="original"
                  muted
                  playsinline
                  webkit-playsinline>
                  <source :src="videoUrl" type="video/mp4">
                  您的浏览器不支持视频播放
                </video>
              </div>
              <div v-else class="video-placeholder">
                <el-skeleton :loading="true" animated>
                  <template slot="template">
                    <el-skeleton-item variant="rect" style="width: 100%; height: 120px; background-color: #f4f4f4;"/>
                  </template>
                </el-skeleton>
                <div class="placeholder-text">加载原始视频中...</div>
              </div>
            </div>
            
            <!-- 合成结果视频 -->
            <div class="mobile-section" id="mobile-result-video">
              <h4 class="section-title">合成结果视频</h4>
              <div v-if="digitalHuman && digitalHuman.status === 'completed' && digitalHuman.result_url && resultUrl" 
                 class="video-player has-result">
                <video 
                  ref="mobileResultPlayer" 
                  controls
                  style="width: 100%; max-height: 200px; background-color: #000; display: block; z-index: 0 !important; transform: translateZ(0);" 
                  data-video-type="result"
                  muted
                  playsinline
                  webkit-playsinline>
                  <source :src="resultUrl" type="video/mp4">
                  您的浏览器不支持视频播放
                </video>
              </div>
              <div v-else class="result-placeholder" :class="{'processing': digitalHuman && digitalHuman.status === 'processing'}">
                <template v-if="digitalHuman && digitalHuman.status === 'processing'">
                  <div class="processing-info">
                    <el-progress :percentage="progress" :format="progressFormat"></el-progress>
                    <p>视频正在处理中，请等待完成后查看结果</p>
                  </div>
                </template>
                <template v-else-if="digitalHuman && digitalHuman.status === 'failed'">
                  <div class="error-message">
                    处理失败: {{ digitalHuman.error_msg || '任务处理失败' }}
                  </div>
                </template>
                <template v-else>
                  <div class="waiting-info">
                    <p>尚未生成结果视频</p>
                  </div>
                </template>
              </div>
            </div>
            
            <!-- 下载结果按钮区域 - 移动端 -->
            <div v-if="digitalHuman && digitalHuman.status === 'completed' && digitalHuman.result_url" class="mobile-download-section">
              <div class="mobile-download-container">
                <h3>下载视频</h3>
                <p>合成任务已完成，您可以下载生成的视频</p>
                <div class="action-buttons">
                  <el-button type="primary" @click="downloadResult" class="download-button" block>
                    <i class="el-icon-download"></i> 下载合成结果
                  </el-button>
                  <el-button type="success" @click="shareTask" class="share-button" :disabled="isShared" block style="margin-top: 10px;">
                    <i class="el-icon-share"></i> {{ getShareButtonText() }}
                </el-button>
                </div>
              </div>
            </div>
          </div>
          
          <el-empty v-else description="未找到任务信息"></el-empty>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import { getAudioUrl, downloadFile, getDirectFileUrl } from '@/utils/fileAccess'
import { shareTask } from '@/api/share'

export default {
  name: 'DigitalHumanDetail',
  data() {
    return {
      loading: false,
      digitalHuman: null,
      progress: 0,
      refreshInterval: null,
      baseURL: process.env.VUE_APP_API_URL || 'http://localhost:8080',
      audioUrl: '',
      videoUrl: null,
      resultUrl: null,
      screenWidth: window.innerWidth,
      videoLoaded: false,
      resultLoaded: false,
      globalStyleEl: null,
      isMobile: false,
      loadingProgress: false,
      progressTimer: null,
      error: null,
      isShared: false,
      shareStatus: 'private'
    }
  },
  computed: {
    token() {
      return localStorage.getItem('token') || ''
    },
    taskId() {
      return this.$route.params.id
    }
  },
  created() {
    this.fetchDigitalHumanDetail()
    window.addEventListener('resize', this.handleResize)
    // 初始化检测设备类型
    this.checkDeviceType()
  },
  mounted() {
    // 如果任务状态是pending或processing，设置定时刷新
    if (this.digitalHuman && (this.digitalHuman.status === 'pending' || this.digitalHuman.status === 'processing')) {
      this.startRefreshInterval()
    }
    
    // 添加视频错误处理
    this.$nextTick(() => {
      // 延迟一点时间注册事件，确保DOM已完全加载
      setTimeout(() => {
        // 为所有视频元素添加错误处理
        const videoElements = ['videoPlayer', 'mobileVideoPlayer', 'resultPlayer', 'mobileResultPlayer']
        videoElements.forEach(refName => {
          const videoEl = this.$refs[refName]
          if (videoEl) {
            // 清除可能存在的旧事件处理器
            videoEl.removeEventListener('error', this.handleVideoError, true)
            videoEl.removeEventListener('loadeddata', this.handleVideoLoaded, false)
            
            // 添加新的事件处理器
            videoEl.addEventListener('error', this.handleVideoError, true)
            
            // 添加加载成功处理
            videoEl.addEventListener('loadeddata', this.handleVideoLoaded, false)
            
            console.log(`已为视频元素 ${refName} 添加事件处理`)
          } else {
            console.warn(`未找到视频元素: ${refName}`)
          }
        })
      }, 500)
    })
    
    // 修复全局样式，确保菜单在合适的层级
    this.fixGlobalStyles()
  },
  beforeDestroy() {
    this.clearRefreshInterval()
    window.removeEventListener('resize', this.handleResize)
    
    // 移除所有视频事件监听器
    const videoElements = ['videoPlayer', 'mobileVideoPlayer', 'resultPlayer', 'mobileResultPlayer']
    videoElements.forEach(refName => {
      const videoEl = this.$refs[refName]
      if (videoEl) {
        videoEl.removeEventListener('error', this.handleVideoError, true)
        videoEl.removeEventListener('loadeddata', this.handleVideoLoaded, false)
        
        // 停止视频播放
        try {
          videoEl.pause()
          videoEl.src = ''
          videoEl.load()
        } catch (e) {
          console.error('清理视频元素时出错:', e)
        }
      }
    })
    
    // 释放媒体URL
    this.clearVideoUrls()
    
    // 恢复全局样式
    this.restoreGlobalStyles()
  },
  methods: {
    // 处理视频加载完成
    handleVideoLoaded(event) {
      const videoElement = event.target;
      const videoType = videoElement.dataset.videoType || 'unknown';
      console.log(`视频已加载成功: ${videoType}`);
      
      if (videoType === 'original') {
        this.videoLoaded = true;
      } else if (videoType === 'result') {
        this.resultLoaded = true;
      }
    },
    
    // 处理视频错误
    handleVideoError(event) {
      const videoElement = event.target;
      
      // 通过dataset标记区分是哪种类型的视频
      const videoType = videoElement.dataset.videoType || 'unknown';
      console.error(`视频加载失败: ${videoType}`, event);
      
      // 如果是视频标签本身的错误
      if (videoElement.tagName === 'VIDEO') {
        const sourceElements = videoElement.querySelectorAll('source');
        console.log(`${videoType}视频加载错误，源元素数量:`, sourceElements.length);
        
        // 获取正确的视频URL
        let correctUrl = '';
        if (videoType === 'original') {
          correctUrl = this.videoUrl;
        } else if (videoType === 'result') {
          correctUrl = this.resultUrl;
        } else {
          console.warn('未知视频类型:', videoType);
          return; // 无法确定视频源，不处理
        }
        
        if (!correctUrl) {
          console.warn(`${videoType}视频URL为空，无法重新加载`);
          return;
        }
        
        console.log(`为${videoType}视频重新设置URL:`, correctUrl);
        
        // 清理所有现有源
        while (videoElement.firstChild) {
          videoElement.removeChild(videoElement.firstChild);
        }
        
        // 添加新的源
        const newSource = document.createElement('source');
        newSource.setAttribute('src', correctUrl);
        newSource.setAttribute('type', 'video/mp4');
        videoElement.appendChild(newSource);
        
        console.log(`已尝试使用MP4格式重新加载${videoType}视频`);
        
        // 重新加载视频
        try {
          videoElement.load();
        } catch (e) {
          console.error(`加载${videoType}视频时出错:`, e);
        }
      }
    },
    
    // 处理窗口大小变化
    handleResize() {
      this.screenWidth = window.innerWidth
      // 更新设备类型
      this.checkDeviceType()
    },
    
    // 获取数字人详情
    fetchDigitalHumanDetail() {
      this.loading = true
      
      // 清除之前的视频URL
      this.clearVideoUrls()
      
      axios.get(`${this.baseURL}/api/digital-human/${this.taskId}`, {
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      })
        .then(response => {
          // 确保从response.data中获取正确的数据结构
          this.digitalHuman = response.data.digital_human || response.data
          
          // 加载媒体URL
          this.loadMediaUrls()
          
          // 如果任务正在处理中，开始轮询进度
          if (this.digitalHuman && (this.digitalHuman.status === 'pending' || this.digitalHuman.status === 'processing')) {
            this.startRefreshInterval()
          }
        })
        .catch(error => {
          console.error('获取数字人详情失败:', error)
          this.$message.error('获取数字人详情失败: ' + ((error.response && error.response.data && error.response.data.error) || error.message))
        })
        .finally(() => {
          this.loading = false
        })
    },
    
    // 查询进度
    queryProgress() {
      axios.get(`${this.baseURL}/api/digital-human/${this.taskId}/progress`, {
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      })
        .then(response => {
          this.progress = response.data.progress || 0
          
          // 如果任务已完成或失败，刷新详情并停止轮询
          if (response.data.status === 'completed' || response.data.status === 'failed') {
            this.fetchDigitalHumanDetail()
            this.clearRefreshInterval()
          }
        })
        .catch(error => {
          console.error('查询进度失败:', error)
          // 出错时不显示错误消息，避免频繁弹窗
        })
    },
    
    // 开始定时刷新
    startRefreshInterval() {
      this.clearRefreshInterval() // 先清除可能存在的定时器
      this.refreshInterval = setInterval(() => {
        this.queryProgress()
      }, 5000) // 每5秒查询一次进度
    },
    
    // 清除定时刷新
    clearRefreshInterval() {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval)
        this.refreshInterval = null
      }
    },
    
    // 下载结果
    downloadResult() {
      if (!this.digitalHuman || !this.digitalHuman.result_url) {
        this.$message.warning('没有可下载的结果文件')
        return
      }
      const fileName = `digital_human_${this.taskId}${this.getFileExtension(this.digitalHuman.result_url)}`
      downloadFile(this.digitalHuman.result_url, fileName)
    },

    // 获取文件扩展名
    getFileExtension(filePath) {
      if (!filePath) return ''
      const match = filePath.match(/\.[^.]+$/)
      return match ? match[0] : ''
    },

    // 加载媒体URL
    async loadMediaUrls() {
      console.log('====== 开始加载媒体URL ======');
      
      // 首先清除先前的URL和状态
      this.clearVideoUrls();
      
      try {
        if (!this.digitalHuman) {
          console.warn('未找到数字人数据，无法加载媒体');
          return;
        }
        
        console.log('数字人状态:', this.digitalHuman.status);
        console.log('数字人ID:', this.digitalHuman.id || this.taskId);
        
        // 1. 先加载音频 - 音频加载通常较快，不会影响视频
        await this.loadAudioFile();
        
        // 2. 使用Promise.all同时触发两个视频的加载，但有延迟差异
        const loadVideos = async () => {
          // 先触发原始视频加载
          const originalVideo = this.loadOriginalVideo().catch(err => {
            console.error('原始视频加载失败:', err);
          });
          
          // 延迟加载结果视频，避免同时竞争资源
          await new Promise(resolve => setTimeout(resolve, 500));
          
          const resultVideo = this.loadResultVideo().catch(err => {
            console.error('结果视频加载失败:', err);
          });
          
          // 等待两个视频都加载完成
          await Promise.all([originalVideo, resultVideo]);
        };
        
        await loadVideos();
        
        console.log('====== 媒体URL加载完成 ======');
      } catch (error) {
        console.error('加载媒体URL总错误:', error);
        this.$message.error('加载媒体文件失败: ' + error.message);
      }
    },
    
    // 单独处理音频加载
    async loadAudioFile() {
      try {
        if (!this.digitalHuman.audio_url) {
          console.warn('没有音频URL');
          return;
        }
        
        console.log('正在加载音频URL:', this.digitalHuman.audio_url);
        this.audioUrl = await getAudioUrl(this.digitalHuman.audio_url);
        console.log('音频URL加载完成:', this.audioUrl);
        
        // 重新设置音频源
        this.$nextTick(() => {
          const audioPlayers = ['audioPlayer', 'mobileAudioPlayer'];
          audioPlayers.forEach(player => {
            if (this.$refs[player]) {
              this.$refs[player].load();
              console.log(`已加载音频到${player}播放器`);
            }
          });
        });
      } catch (error) {
        console.error('音频加载失败:', error);
      }
    },
    
    // 单独处理原始视频加载
    async loadOriginalVideo() {
      try {
        // 重置加载状态
        this.videoLoaded = false;
        
        if (!this.digitalHuman.video_url) {
          console.warn('没有原始视频URL');
          this.videoUrl = null;
          return;
        }
        
        console.log('===== 原始视频处理开始 =====');
        console.log('原始视频URL:', this.digitalHuman.video_url);
        
        // 生成随机时间戳参数，防止缓存问题
        const timeStamp = Date.now();
        let videoUrlTemp = getDirectFileUrl(this.digitalHuman.video_url);
        
        // 确保URL中有时间戳，防止缓存
        if (videoUrlTemp) {
          // 添加时间戳参数
          videoUrlTemp = videoUrlTemp.includes('?') 
            ? `${videoUrlTemp}&_t=${timeStamp}` 
            : `${videoUrlTemp}?_t=${timeStamp}`;
        }
        
        console.log('处理后的原始视频URL:', videoUrlTemp);
        
        // 确保URL是有效的
        if (videoUrlTemp) {
          // 重要：设置到视频源
          this.videoUrl = videoUrlTemp;
          
          // 重新加载视频元素
          this.$nextTick(() => {
            setTimeout(() => {
              console.log('开始加载原始视频到播放器...');
              try {
                const videoPlayers = ['videoPlayer', 'mobileVideoPlayer'];
                videoPlayers.forEach(player => {
                  const videoEl = this.$refs[player];
                  if (videoEl) {
                    console.log(`重新加载${player}播放器的原始视频`);
                    
                    // 清除所有源并添加新源
                    while (videoEl.firstChild) {
                      if (videoEl.firstChild.tagName === 'SOURCE') {
                        videoEl.removeChild(videoEl.firstChild);
                      } else {
                        break; // 保留非source元素
                      }
                    }
                    
                    // 添加新的源
                    const sourceEl = document.createElement('source');
                    sourceEl.setAttribute('src', this.videoUrl);
                    sourceEl.setAttribute('type', 'video/mp4');
                    videoEl.insertBefore(sourceEl, videoEl.firstChild);
                    
                    // 确保视频元素标记了类型
                    videoEl.dataset.videoType = 'original';
                    
                    try {
                      videoEl.load();
                    } catch (loadError) {
                      console.error('加载原始视频时出错:', loadError);
                    }
                  } else {
                    console.warn(`未找到${player}播放器引用`);
                  }
                });
                console.log('原始视频播放器加载完成');
              } catch (innerError) {
                console.error('加载原始视频元素错误:', innerError);
              }
            }, 100);
          });
        }
        console.log('===== 原始视频处理完成 =====');
      } catch (error) {
        console.error('原始视频处理错误:', error);
      }
    },
    
    // 单独处理结果视频加载
    async loadResultVideo() {
      try {
        // 重置加载状态
        this.resultLoaded = false;
        
        if (!this.digitalHuman.result_url) {
          console.warn('没有结果视频URL');
          this.resultUrl = null;
          return;
        }
        
        console.log('===== 结果视频处理开始 =====');
        console.log('结果视频URL:', this.digitalHuman.result_url);
        
        // 生成随机时间戳参数，防止缓存问题
        const timeStamp = Date.now() + 100; // 与原始视频的时间戳稍有不同，进一步防止混淆
        let resultUrlTemp = getDirectFileUrl(this.digitalHuman.result_url);
        
        // 确保URL中有时间戳，防止缓存
        if (resultUrlTemp) {
          // 添加时间戳参数
          resultUrlTemp = resultUrlTemp.includes('?') 
            ? `${resultUrlTemp}&_t=${timeStamp}` 
            : `${resultUrlTemp}?_t=${timeStamp}`;
        }
        
        console.log('处理后的结果视频URL:', resultUrlTemp);
        
        // 确保URL是有效的
        if (resultUrlTemp) {
          // 重要：设置到视频源
          this.resultUrl = resultUrlTemp;
          
          // 重新加载视频元素 - 延时更长确保与原始视频加载分开
          this.$nextTick(() => {
            setTimeout(() => {
              console.log('开始加载结果视频到播放器...');
              try {
                const resultPlayers = ['resultPlayer', 'mobileResultPlayer'];
                resultPlayers.forEach(player => {
                  const videoEl = this.$refs[player];
                  if (videoEl) {
                    console.log(`重新加载${player}播放器的结果视频`);
                    
                    // 清除所有源并添加新源
                    while (videoEl.firstChild) {
                      if (videoEl.firstChild.tagName === 'SOURCE') {
                        videoEl.removeChild(videoEl.firstChild);
                      } else {
                        break; // 保留非source元素
                      }
                    }
                    
                    // 添加新的源
                    const sourceEl = document.createElement('source');
                    sourceEl.setAttribute('src', this.resultUrl);
                    sourceEl.setAttribute('type', 'video/mp4');
                    videoEl.insertBefore(sourceEl, videoEl.firstChild);
                    
                    // 确保视频元素标记了类型
                    videoEl.dataset.videoType = 'result';
                    
                    try {
                      videoEl.load();
                    } catch (loadError) {
                      console.error('加载结果视频时出错:', loadError);
                    }
                  } else {
                    console.warn(`未找到${player}播放器引用`);
                  }
                });
                console.log('结果视频播放器加载完成');
              } catch (innerError) {
                console.error('加载结果视频元素错误:', innerError);
              }
            }, 800); // 延迟更长时间，确保与原始视频加载分开
          });
        }
        console.log('===== 结果视频处理完成 =====');
      } catch (error) {
        console.error('结果视频处理错误:', error);
      }
    },
    
    // 返回列表
    goBack() {
      this.$router.push('/digital-human')
    },
    
    // 格式化进度
    progressFormat(percentage) {
      return percentage === 100 ? '完成' : `${percentage}%`
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
    },
    
    // 清除视频URL方法
    clearVideoUrls() {
      // 释放之前的URL
      if (this.videoUrl && this.videoUrl.startsWith('blob:')) {
        try {
          URL.revokeObjectURL(this.videoUrl)
          console.log('已释放原始视频URL')
        } catch (e) {
          console.error('释放原始视频URL时出错:', e)
        }
      }
      
      if (this.resultUrl && this.resultUrl.startsWith('blob:')) {
        try {
          URL.revokeObjectURL(this.resultUrl)
          console.log('已释放结果视频URL')
        } catch (e) {
          console.error('释放结果视频URL时出错:', e)
        }
      }
      
      if (this.audioUrl && this.audioUrl.startsWith('blob:')) {
        try {
          URL.revokeObjectURL(this.audioUrl)
          console.log('已释放音频URL')
        } catch (e) {
          console.error('释放音频URL时出错:', e)
        }
      }
      
      // 重置URL为null
      this.videoUrl = null
      this.resultUrl = null
      this.videoLoaded = false
      this.resultLoaded = false
    },
    
    // 修复全局样式，确保菜单在合适的层级
    fixGlobalStyles() {
      // 动态创建样式元素
      this.globalStyleEl = document.createElement('style')
      this.globalStyleEl.type = 'text/css'
      this.globalStyleEl.innerHTML = `
        /* 底部菜单和侧边栏保持高层级 */
        .app-footer, .app-sidebar {
          z-index: 3000 !important;
          position: relative !important;
        }
        
        /* 顶部菜单放在较低层级，让详情页内容可以覆盖它 */
        .el-menu, .el-submenu, .el-menu-item, .app-header, .app-navbar {
          z-index: 1 !important;
        }
        
        /* 视频元素保持低层级 */
        video, iframe, embed, object {
          z-index: 1 !important;
        }
      `
      document.head.appendChild(this.globalStyleEl)
    },
    
    // 恢复全局样式
    restoreGlobalStyles() {
      if (this.globalStyleEl && this.globalStyleEl.parentNode) {
        this.globalStyleEl.parentNode.removeChild(this.globalStyleEl)
        this.globalStyleEl = null
      }
    },

    // 初始化检测设备类型
    checkDeviceType() {
      this.isMobile = window.innerWidth <= 768; // 假设小于等于768px为移动端
    },
    
    // 分享任务
    async shareTask() {
      try {
        this.$confirm('确定要分享此任务到灵感页吗?', '分享确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'info'
        }).then(async () => {
          const loading = this.$loading({
            lock: true,
            text: '正在提交分享请求...',
            spinner: 'el-icon-loading'
          })
          
          try {
            const result = await shareTask({
              taskId: this.digitalHuman.id,
              mode: 'digital_human',
              taskType: 'digital_human'
            })
            
            loading.close()
            
            if (result.success) {
              this.$message.success(result.message)
              // 更新本地状态
              this.digitalHuman.is_shared = true
              this.digitalHuman.share_status = 'pending_review'
              // 刷新任务信息
              await this.fetchDigitalHuman()
            } else {
              this.$message.error(result.message)
            }
          } catch (error) {
            loading.close()
            this.$message.error('分享失败: ' + (error.message || '未知错误'))
          }
        }).catch(() => {
          // 用户取消分享，不执行任何操作
        })
      } catch (error) {
        this.$message.error('分享操作失败: ' + (error.message || '未知错误'))
      }
    },
    
    // 获取分享按钮文本
    getShareButtonText() {
      if (!this.digitalHuman) {
        return '分享到灵感页'
      }
      
      if (this.digitalHuman.is_shared) {
        switch (this.digitalHuman.share_status) {
          case 'pending_review':
            return '审核中'
          case 'approved':
            return '已分享'
          case 'rejected':
            return '分享被拒绝'
          default:
            return '已分享'
        }
      }
      return '分享到灵感页'
    },
    
    // 获取任务详情
    async fetchDigitalHuman() {
      try {
        this.loading = true
        const id = this.$route.params.id
        
        const response = await axios.get(`/api/digital-human/${id}`, {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        
        if (response.data && (response.data.digital_human || response.data)) {
          this.digitalHuman = response.data.digital_human || response.data
          
          // 更新分享状态
          this.isShared = this.digitalHuman.is_shared || false
          this.shareStatus = this.digitalHuman.share_status || 'private'
          
          // 加载媒体URL
          this.loadMediaUrls()
          
          // 如果任务正在处理中，启动进度查询
          if (this.digitalHuman && (this.digitalHuman.status === 'pending' || this.digitalHuman.status === 'processing')) {
            this.startRefreshInterval()
          }
        } else {
          this.$message.error('获取数字人合成任务信息失败')
        }
      } catch (error) {
        console.error('获取数字人合成任务信息失败', error)
        this.$message.error('获取数字人合成任务信息失败: ' + ((error.response && error.response.data && error.response.data.error) || error.message))
      } finally {
        this.loading = false
      }
    },
  }
}
</script>

<style>
/* 全局样式覆盖，强制禁止水平滚动 */
.digital-human-detail-container {
  transform: translateZ(0); /* 创建新的堆叠上下文 */
  isolation: isolate; /* 现代浏览器隔离堆叠上下文 */
  position: relative;
  z-index: 100 !important; /* 提高详情页整体层级 */
  overflow-x: hidden !important;
  width: 100% !important;
  max-width: 100% !important;
}

/* 强制所有视频元素不能超出其父容器的堆叠上下文 */
.digital-human-detail-container video, 
.digital-human-detail-container audio, 
.digital-human-detail-container .video-container, 
.digital-human-detail-container .video-preview, 
.digital-human-detail-container .result-preview,
.digital-human-detail-container .video-player {
  transform: translateZ(0);
  position: relative !important;
  z-index: 0 !important; /* 使用固定的低值而不是auto */
  isolation: isolate;
  max-width: 100%;
  max-height: 100%;
}

/* 移动端顶部导航栏 - 提高层级 */
.digital-human-detail-container .mobile-header-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background-color: #409EFF;
  display: flex;
  align-items: center;
  padding: 0 12px;
  z-index: 2000 !important; /* 确保导航栏在最上层 */
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  color: #fff;
  width: 100%;
  box-sizing: border-box;
  display: none; /* 默认隐藏，在移动端通过v-show显示 */
}

/* PC端页面标题改进 */
.digital-human-detail-container .page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background-color: transparent; /* 改为透明背景 */
  border-bottom: 1px solid rgba(235, 238, 245, 0.6); /* 稍微透明的边框 */
  margin-bottom: 20px;
  box-shadow: none; /* 移除阴影 */
}

.digital-human-detail-container .page-header h2 {
  color: #303133;
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

/* PC端下载区域样式 */
.digital-human-detail-container .download-section {
  margin-top: 40px;
  margin-bottom: 30px;
  border-top: 1px solid #ebeef5;
  padding-top: 30px;
}

.digital-human-detail-container .download-container {
  background-color: #f0f9ff;
  border: 1px solid #a0cfff;
  border-radius: 8px;
  padding: 25px 30px;
  text-align: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.digital-human-detail-container .download-container:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.1);
}

.digital-human-detail-container .download-container h3 {
  color: #409EFF;
  font-size: 20px;
  margin-bottom: 15px;
  font-weight: 600;
}

.digital-human-detail-container .download-tips {
  color: #606266;
  margin-bottom: 20px;
  font-size: 15px;
  line-height: 1.5;
}

.digital-human-detail-container .download-button {
  padding: 12px 25px;
  font-size: 16px;
  border-radius: 6px;
  transition: all 0.3s;
  min-width: 200px;
}

.digital-human-detail-container .download-button:hover {
  transform: scale(1.05);
  background-color: #66b1ff;
  border-color: #66b1ff;
}

.digital-human-detail-container .download-button i {
  margin-right: 8px;
  font-size: 18px;
}

/* 移动端下载区域样式 */
.digital-human-detail-container .mobile-download-section {
  margin-top: 30px;
  margin-bottom: 60px;
  padding: 0 12px;
}

.digital-human-detail-container .mobile-download-container {
  background: linear-gradient(135deg, #e6f3ff 0%, #f0f9ff 100%);
  border-radius: 12px;
  padding: 20px 15px;
  text-align: center;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.08);
  border: 1px solid #d9ecff;
}

.digital-human-detail-container .download-icon {
  font-size: 32px;
  color: #409EFF;
  margin-bottom: 10px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
}

.digital-human-detail-container .download-text {
  color: #606266;
  margin-bottom: 15px;
  font-size: 14px;
  line-height: 1.4;
  padding: 0 10px;
}

.digital-human-detail-container .mobile-download-button {
  width: 100%;
  padding: 12px 0;
  font-size: 16px;
  border-radius: 6px;
  margin-bottom: 5px;
  background-color: #409EFF;
  border-color: #409EFF;
  color: white;
  box-shadow: 0 4px 8px rgba(64, 158, 255, 0.3);
}

.digital-human-detail-container .mobile-download-button:active {
  transform: scale(0.98);
  box-shadow: 0 2px 4px rgba(64, 158, 255, 0.3);
}

/* 移动端返回按钮样式增强 */
.digital-human-detail-container .header-back {
  display: flex;
  align-items: center;
  font-size: 16px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
  z-index: 2100 !important; /* 确保返回按钮在最顶层 */
  position: relative;
}

.digital-human-detail-container .header-back:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.digital-human-detail-container .header-back i {
  margin-right: 4px;
  font-size: 16px;
}

.digital-human-detail-container .header-title {
  margin: 0 0 0 10px;
  font-size: 16px;
  font-weight: 500;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.digital-human-detail-container .mobile-header-placeholder {
  height: 56px;
  width: 100%;
  flex-shrink: 0;
  display: none; /* 默认隐藏，在移动端显示 */
}

/* PC端特有样式 */
.digital-human-detail-container .desktop-content-view {
  display: block; /* 默认显示 */
  width: 100%;
  box-sizing: border-box;
}

/* 移动端内容区 */
.digital-human-detail-container .mobile-content-view {
  display: none; /* 默认隐藏，在移动端显示 */
  padding: 0;
  width: 100%;
  background-color: #fff;
  min-height: 100%;
  box-sizing: border-box;
}

.digital-human-detail-container .mobile-content-inner {
  min-height: calc(100vh - 56px);
  width: 100%;
  box-sizing: border-box;
}

/* 移动端基本信息 */
.digital-human-detail-container .basic-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #ebeef5;
  width: 100%;
  box-sizing: border-box;
}

.digital-human-detail-container .create-time {
  font-size: 12px;
  color: #909399;
}

.digital-human-detail-container .mobile-task-info {
  padding: 8px 12px;
  box-sizing: border-box;
  width: 100%;
}

.digital-human-detail-container .mobile-section {
  padding: 8px 12px;
  border-top: 1px solid #ebeef5;
  margin-top: 8px;
  box-sizing: border-box;
  width: 100%;
}

.digital-human-detail-container .section-title {
  font-size: 16px;
  font-weight: 600;
  margin: 5px 0;
  color: #303133;
  width: 100%;
  box-sizing: border-box;
}

/* 媒体查询 - 移动端适配 */
@media screen and (max-width: 768px) {
  /* 数字人详情页容器移动端样式 */
  .digital-human-detail-container {
    padding: 0;
    width: 100%;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 5 !important; /* 低于全局菜单层级 */
    background-color: #fff;
    color: #333;
    height: 100vh;
    overflow-x: hidden;
    box-sizing: border-box;
    transform: translateZ(0);
    isolation: isolate;
  }
  
  /* 隐藏PC端，显示移动端 */
  .digital-human-detail-container .page-header {
    display: none;
  }
  
  .digital-human-detail-container .desktop-content-view {
    display: none;
  }
  
  .digital-human-detail-container .mobile-header-bar {
    z-index: 10 !important; /* 高于内容，低于应用菜单 */
    display: flex !important; /* 强制在移动端显示，覆盖v-show的影响 */
  }
  
  .digital-human-detail-container .mobile-header-placeholder {
    display: block;
  }
  
  .digital-human-detail-container .mobile-content-view {
    transform: translateZ(0);
    position: relative;
    isolation: isolate;
    z-index: 1 !important;
    display: block;
  }
  
  .digital-human-detail-container .detail-content-wrapper {
    transform: translateZ(0);
    position: absolute;
    top: 56px;
    left: 0;
    right: 0;
    bottom: 0;
    overflow-y: auto;
    overflow-x: hidden;
    box-sizing: border-box;
    padding-bottom: 120px !important; /* 增加更多底部空间 */
    z-index: 1 !important;
    isolation: isolate;
  }
  
  .digital-human-detail-container .detail-content {
    width: 100%;
    margin: 0;
    background: #fff;
    backdrop-filter: none;
    box-shadow: none;
    border: none;
    border-radius: 0;
    min-height: 100%;
    overflow-x: hidden;
    box-sizing: border-box;
    padding: 0;
    position: relative;
    z-index: auto !important; /* 应用auto，遵循父容器层级 */
  }
  
  .digital-human-detail-container .mobile-section {
    padding: 8px 12px;
    border-top: 1px solid #ebeef5;
    margin-top: 8px;
    box-sizing: border-box;
    width: 100%;
  }
  
  .digital-human-detail-container .section-title {
    color: #303133;
  }
  
  .digital-human-detail-container .mobile-text-content {
    background-color: #f8f8f8;
    color: #333;
  }
  
  .digital-human-detail-container .create-time {
    color: #909399;
  }
  
  .digital-human-detail-container .basic-info {
    border-bottom: 1px solid #ebeef5;
  }
  
  .digital-human-detail-container .audio-player,
  .digital-human-detail-container .video-player {
    margin: 8px 0;
    width: 100%;
    box-sizing: border-box;
  }
  
  .digital-human-detail-container .action-buttons {
    margin-top: 10px;
    margin-bottom: 20px;
    width: 100%;
    box-sizing: border-box;
  }
  
  .digital-human-detail-container .action-buttons .el-button {
    width: 100%;
    box-sizing: border-box;
  }
  
  /* 确保元素不溢出 */
  .digital-human-detail-container .el-tag {
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 12px;
    padding: 0 5px;
    height: 22px;
    line-height: 20px;
  }
  
  /* 视频播放相关容器 */
  .digital-human-detail-container .video-player, .digital-human-detail-container .video-container, .digital-human-detail-container .video-preview, .digital-human-detail-container .result-preview {
    transform: translateZ(0);
    position: relative;
    z-index: 0 !important;
    max-width: 100%;
    overflow: hidden;
  }
  
  /* 特别处理结果视频区域 */
  .digital-human-detail-container #mobile-result-video {
    margin-bottom: 20px !important; /* 减小底部边距，因为现在有专门的下载区域 */
    padding-bottom: 0;
  }
  
  /* 移动端下载区域特殊处理 */
  .digital-human-detail-container .mobile-download-section {
    margin-top: 30px;
    margin-bottom: 120px !important; /* 确保有足够空间在底部菜单上方 */
    padding: 0 15px;
  }
  
  .digital-human-detail-container .mobile-download-container {
    padding: 20px 15px;
  }
}

.digital-human-detail-container .mobile-video-player {
  position: relative;
  overflow: hidden;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  margin: 10px 0;
  background-color: #000;
}

.digital-human-detail-container .mobile-video-player video {
  display: block;
  width: 100%;
}

.digital-human-detail-container .native-video {
  max-width: 100% !important;
  margin: 0 auto;
  display: block;
  background: #000;
}

.digital-human-detail-container .mobile-video {
  display: block;
  width: 100%;
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.digital-human-detail-container .desktop-video {
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.digital-human-detail-container .video-preview, .digital-human-detail-container .result-preview {
  border: 2px solid #ebeef5;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
  transition: all 0.3s ease;
  position: relative;
}

.digital-human-detail-container .video-preview:before {
  content: "原始视频";
  position: absolute;
  top: -10px;
  left: 10px;
  background-color: rgba(64, 158, 255, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #409EFF;
  z-index: 1;
  font-weight: bold;
}

.digital-human-detail-container .result-preview:before {
  content: "结果视频";
  position: absolute;
  top: -10px;
  left: 10px;
  background-color: rgba(103, 194, 58, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #67c23a;
  z-index: 1;
  font-weight: bold;
}

.digital-human-detail-container .video-loaded {
  border-color: #67c23a;
  box-shadow: 0 0 10px rgba(103, 194, 58, 0.2);
}

.digital-human-detail-container .video-loading {
  border-color: #e6a23c;
  box-shadow: 0 0 10px rgba(230, 162, 60, 0.2);
}

.digital-human-detail-container #mobile-original-video {
  border-left: 4px solid #409EFF;
  position: relative;
}

.digital-human-detail-container #mobile-original-video:before {
  content: "";
  position: absolute;
  left: -4px;
  top: 0;
  bottom: 0;
  width: 4px;
  background-color: #409EFF;
}

.digital-human-detail-container #mobile-result-video {
  border-left: 4px solid #67c23a;
  position: relative;
}

.digital-human-detail-container #mobile-result-video:before {
  content: "";
  position: absolute;
  left: -4px;
  top: 0;
  bottom: 0;
  width: 4px;
  background-color: #67c23a;
}

/* 移动端视频容器增强 */
.digital-human-detail-container .mobile-section .video-player {
  position: relative;
  overflow: hidden;
  border-radius: 8px;
  margin-top: 10px;
  background-color: #000;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

/* 视频控制增强 */
video::-webkit-media-controls {
  display: flex !important;
  visibility: visible !important;
  opacity: 1 !important;
}

/* 视频播放按钮样式增强 */
video::-webkit-media-controls-play-button {
  display: flex !important;
  visibility: visible !important;
  opacity: 1 !important;
  background-color: rgba(255, 255, 255, 0.7);
  border-radius: 50%;
  width: 40px;
  height: 40px;
}

/* 添加移动端视频错误状态 */
.digital-human-detail-container .video-error {
  border: 2px solid #f56c6c;
  background-color: rgba(245, 108, 108, 0.1);
  padding: 10px;
  text-align: center;
  color: #f56c6c;
  margin: 10px 0;
  border-radius: 4px;
}

.digital-human-detail-container .processing-info, .digital-human-detail-container .waiting-info {
  padding: 20px;
  text-align: center;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.digital-human-detail-container .processing-info {
  border-left: 4px solid #e6a23c;
}

.digital-human-detail-container .waiting-info {
  border-left: 4px solid #909399;
  color: #909399;
}

.digital-human-detail-container .video-container {
  width: 100%;
  background-color: #000;
  border-radius: 4px;
  overflow: hidden;
}

.digital-human-detail-container .video-placeholder {
  width: 100%;
  position: relative;
  min-height: 120px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.digital-human-detail-container .placeholder-text {
  text-align: center;
  color: #909399;
  margin-top: 10px;
  font-size: 14px;
}

.digital-human-detail-container .result-placeholder {
  width: 100%;
  min-height: 160px;
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 6px;
  margin-bottom: 10px;
}

.digital-human-detail-container .empty-result {
  min-height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f9f9f9;
  padding: 20px;
  border-radius: 6px;
}

.digital-human-detail-container .has-result {
  border: 2px solid #67c23a;
}

.digital-human-detail-container .processing {
  border-left: 4px solid #e6a23c;
}

.action-buttons-group {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.share-button {
  margin-left: 10px;
}
</style>