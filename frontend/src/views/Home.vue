<template>
  <div class="home-container">
    <!-- 顶部欢迎区域 -->
    <div class="welcome-section">
      <div class="welcome-decorations">
        <div class="welcome-bg-orb"></div>
        <div class="welcome-bg-orb secondary"></div>
        <div class="welcome-grid"></div>
        <div class="tech-circles"></div>
      </div>
      <div class="welcome-content">
        <h1>欢迎使用 <span class="highlight">Virtual Human Studio</span></h1>
        <p>探索AI驱动的数字人创作平台，释放创意无限可能</p>
        <div class="welcome-actions">
          <el-button type="primary" class="action-button" @click="navigateTo('/digital-human')">
            <i class="el-icon-video-camera"></i> 开始创作
          </el-button>
          <el-button class="action-button secondary" @click="navigateTo('/voice-clone')">
            <i class="el-icon-microphone"></i> 克隆音色
          </el-button>
        </div>
      </div>
    </div>

    <!-- 功能卡片区域 -->
    <div class="section-title">
      <div class="section-accent"></div>
      <h2>核心功能</h2>
    </div>
    <el-row :gutter="30" class="feature-section">
      <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="(card, index) in featureCards" :key="index">
        <div class="glass-card feature-card"
             @click="navigateTo(card.route)"
             tabindex="0"
             @keyup.enter="navigateTo(card.route)">
          <div class="card-icon-wrapper">
            <div class="card-icon">
              <i :class="card.icon"></i>
            </div>
          </div>
          <div class="card-content">
            <h3>{{ card.title }}</h3>
            <p>{{ card.description }}</p>
            <div class="card-action">
              <span class="explore-link">
                <i class="el-icon-right"></i> 立即体验
              </span>
            </div>
          </div>
          <div class="hover-effect"></div>
          <div class="card-decoration"></div>
        </div>
      </el-col>
    </el-row>
    
    <!-- 统计数据区域 -->
    <div class="section-title stats-title">
      <div class="section-accent"></div>
      <h2>平台数据</h2>
    </div>
    <el-row :gutter="30" class="stats-section">
      <el-col :xs="12" :sm="8" :md="6" :lg="4" v-for="(stat, index) in stats" :key="index">
        <div class="glass-card stat-card" :class="getStatClass(index)">
          <div class="stat-icon">
            <i :class="getStatIcon(index)"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value" v-countup:onece="stat.value">{{ stat.value }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
          <div class="stat-bg-circle"></div>
        </div>
      </el-col>
    </el-row>
    
    <!-- 最近活动区域 -->
    <div class="glass-card recent-section">
      <div class="section-header">
        <div class="section-accent"></div>
        <h2>最近活动</h2>
        <div class="section-actions">
          <el-button type="text" size="small" @click="handleRefresh" :loading="loading">
            <i class="el-icon-refresh"></i> 刷新
          </el-button>
        </div>
      </div>
      
      <!-- 移动端下拉刷新容器 -->
      <el-pull-refresh 
        v-if="isMobile" 
        v-model="isRefreshing" 
        @refresh="handleRefresh">
        
        <!-- 移动端卡片式布局 -->
        <div class="activity-cards">
          <div
            v-for="(activity, index) in recentActivities"
            :key="index"
            class="glass-card activity-card"
            :class="'activity-' + getStatusType(activity.status)"
            @click="viewDetail(activity)">
            <div class="activity-card-header">
              <div class="activity-type-tag" :class="'type-' + getActivityTypeClass(activity.type)">
                <i :class="getActivityTypeIcon(activity.type)"></i>
                <span>{{ activity.type }}</span>
              </div>
              <el-tag size="mini" :type="getStatusType(activity.status)" class="status-tag">{{ activity.status }}</el-tag>
            </div>
            <div class="activity-card-content">
              <div class="activity-name">{{ activity.name }}</div>
              <div class="activity-time">
                <i class="el-icon-time"></i> {{ activity.created_at }}
              </div>
            </div>
            <div class="activity-card-footer">
              <el-button type="text" size="mini" class="detail-button" @click.stop="viewDetail(activity)">
                查看详情 <i class="el-icon-arrow-right"></i>
              </el-button>
            </div>
            <div class="card-highlight"></div>
          </div>
          
          <!-- 无限滚动加载更多 -->
          <el-infinite-scroll
            v-if="hasMoreActivities"
            :infinite-scroll-disabled="loading"
            :infinite-scroll-distance="10"
            @load="loadMoreActivities">
            <div class="loading-more" v-if="loading">加载中...</div>
          </el-infinite-scroll>
          
          <!-- 无数据提示 -->
          <div class="no-data" v-if="recentActivities.length === 0">
            <i class="el-icon-document"></i>
            <p>暂无活动数据</p>
          </div>
        </div>
      </el-pull-refresh>
      
      <!-- 桌面端表格布局 -->
      <div class="table-container" v-else>
        <el-table
          :data="recentActivities"
          style="width: 100%"
          :header-cell-style="{ background: 'transparent', color: 'rgba(44, 62, 80, 0.85)', fontWeight: '600' }"
          :cell-style="{ background: 'transparent', color: 'rgba(44, 62, 80, 0.75)' }"
          :row-class-name="tableRowClassName"
          size="small"
          highlight-current-row>
        <el-table-column prop="type" label="类型" width="160">
          <template slot-scope="scope">
            <div class="type-cell">
              <div class="type-icon-container" :class="'type-' + getActivityTypeClass(scope.row.type)">
                <i :class="getActivityTypeIcon(scope.row.type)"></i>
              </div>
              <span>{{ scope.row.type }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="180">
          <template slot-scope="scope">
            <span class="table-text-ellipsis table-name" :title="scope.row.name">{{ scope.row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120" align="center">
          <template slot-scope="scope">
            <el-tag size="small" :type="getStatusType(scope.row.status)" effect="light" class="status-tag-table">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            <span class="created-time">
              <i class="el-icon-time"></i> {{ scope.row.created_at }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template slot-scope="scope">
            <el-button type="text" size="small" @click="viewDetail(scope.row)" class="table-action-button">
              <i class="el-icon-view"></i> 查看
            </el-button>
          </template>
        </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'HomeView',
  data() {
    return {
      isMobile: false,
      isRefreshing: false,  // 下拉刷新状态
      loading: false,       // 加载状态
      page: 1,             // 当前页码
      pageSize: 10,         // 每页数量
      hasMoreActivities: true, // 是否有更多数据
      featureCards: [
        {
          title: '音色克隆',
          description: '克隆您喜欢的声音，创建个性化音色',
          icon: 'el-icon-microphone',
          route: '/voice-clone'
        },
        {
          title: '文本转语音',
          description: '将文本转换为自然流畅的语音',
          icon: 'el-icon-reading',
          route: '/tts'
        },
        {
          title: '数字人合成',
          description: '创建逼真的数字人视频',
          icon: 'el-icon-video-camera',
          route: '/digital-human'
        },
        {
          title: '音色库',
          description: '管理您的音色资源',
          icon: 'el-icon-collection',
          route: '/voice-library'
        },
        {
          title: '联系我们',
          description: '获取技术支持和商务合作信息',
          icon: 'el-icon-message',
          route: '/contact'
        }
      ],
      stats: [
        { label: '音色克隆任务', value: 0 },
        { label: 'TTS任务', value: 0 },
        { label: '数字人任务', value: 0 },
        { label: '语音识别任务', value: 0 },
        { label: '图像处理任务', value: 0 }
      ],
      recentActivities: []
    }
  },
  created() {
    this.checkDeviceType()
    window.addEventListener('resize', this.checkDeviceType)
    this.fetchStats()
    this.fetchRecentActivities()
  },
  
  beforeDestroy() {
    window.removeEventListener('resize', this.checkDeviceType)
  },
  methods: {
    checkDeviceType() {
      this.isMobile = window.innerWidth < 768
    },
    navigateTo(route) {
      this.$router.push(route)
    },
    getStatClass(index) {
      const classes = ['voice-stat', 'tts-stat', 'digital-human-stat', 'speech-stat', 'image-stat']
      return classes[index % classes.length]
    },
    getStatIcon(index) {
      const icons = [
        'el-icon-microphone',
        'el-icon-reading',
        'el-icon-video-camera',
        'el-icon-headset',
        'el-icon-picture'
      ]
      return icons[index % icons.length]
    },
    getActivityTypeClass(type) {
      const typeMap = {
        '音色克隆': 'voice',
        'TTS': 'tts',
        '数字人': 'human',
        '语音识别': 'speech',
        '图像处理': 'image'
      }
      return typeMap[type] || 'default'
    },
    getActivityTypeIcon(type) {
      const iconMap = {
        '音色克隆': 'el-icon-microphone',
        'TTS': 'el-icon-reading',
        '数字人': 'el-icon-video-camera',
        '语音识别': 'el-icon-headset',
        '图像处理': 'el-icon-picture'
      }
      return iconMap[type] || 'el-icon-document'
    },
    tableRowClassName({row}) {
      const status = row.status;
      if (status === 'completed') {
        return 'success-row';
      } else if (status === 'failed') {
        return 'danger-row';
      } else if (status === 'processing') {
        return 'warning-row';
      }
      return '';
    },
    getStatusType(status) {
      const statusMap = {
        'pending': 'info',
        'processing': 'warning',
        'completed': 'success',
        'failed': 'danger'
      }
      return statusMap[status] || 'info'
    },
    viewDetail(row) {
      if (row.type === '音色克隆') {
        this.$router.push(`/voice-clone/${row.id}`)
      } else if (row.type === 'TTS') {
        this.$router.push(`/tts/${row.id}`)
      } else if (row.type === '数字人') {
        this.$router.push(`/digital-human/${row.id}`)
      } else if (row.type === '语音识别') {
        this.$router.push(`/speech2text/${row.id}`)
      } else if (row.type === '图像处理') {
        this.$router.push(`/accessory/${row.id}`)
      }
    },
    fetchStats() {
      Promise.all([
        this.$http.get('/api/voice/clones?size=1'),
        this.$http.get('/api/tts?size=1'),
        this.$http.get('/api/digital-human?size=1'),
        this.$http.get('/api/asr?size=1'),
        this.$http.get('/api/accessory?size=1')
      ])
        .then(([voiceCloneRes, ttsRes, digitalHumanRes, asrRes, comfyuiRes]) => {
          this.stats[0].value = voiceCloneRes.data.total || 0
          this.stats[1].value = ttsRes.data.total || 0
          this.stats[2].value = digitalHumanRes.data.total || 0
          
          // 更新统计数据数组，添加语音识别和图像处理任务的统计
          if (this.stats.length === 3) {
            this.stats.push({ label: '语音识别任务', value: asrRes.data.total || 0 })
            this.stats.push({ label: '图像处理任务', value: comfyuiRes.data.total || 0 })
          } else {
            // 如果已经有这些统计项，则更新它们的值
            const asrIndex = this.stats.findIndex(stat => stat.label === '语音识别任务')
            const comfyuiIndex = this.stats.findIndex(stat => stat.label === '图像处理任务')
            
            if (asrIndex !== -1) {
              this.stats[asrIndex].value = asrRes.data.total || 0
            }
            
            if (comfyuiIndex !== -1) {
              this.stats[comfyuiIndex].value = comfyuiRes.data.total || 0
            }
          }
        })
        .catch(error => {
          console.error('获取统计数据失败', error)
        })
    },
    // 获取最近活动数据（支持分页）
    fetchRecentActivities(isRefresh = false) {
      // 如果是刷新，重置页码
      if (isRefresh) {
        this.page = 1
        this.hasMoreActivities = true
      }
      
      // 设置加载状态
      this.loading = true
      
      // 计算每个API请求的数量
      const size = Math.ceil(this.pageSize / 5)
      
      Promise.all([
        this.$http.get(`/api/voice/clones?size=${size}&page=${this.page}`),
        this.$http.get(`/api/tts?size=${size}&page=${this.page}`),
        this.$http.get(`/api/digital-human?size=${size}&page=${this.page}`),
        this.$http.get(`/api/asr?size=${size}&page=${this.page}`),
        this.$http.get(`/api/accessory?size=${size}&page=${this.page}`)
      ])
        .then(([voiceCloneRes, ttsRes, digitalHumanRes, asrRes, comfyuiRes]) => {
          // 处理音色克隆数据
          const voiceClones = (voiceCloneRes.data.voice_clones || []).map(item => ({
            id: item.id,
            type: '音色克隆',
            name: item.name || '未命名音色',
            status: item.status,
            created_at: new Date(item.created_at).toLocaleString()
          }))
          
          // 处理TTS任务数据
          const ttsTasks = (ttsRes.data.tts_tasks || []).map(item => ({
            id: item.id,
            type: 'TTS',
            name: item.name || '未命名TTS任务',
            status: item.status,
            created_at: new Date(item.created_at).toLocaleString()
          }))
          
          // 处理数字人数据
          const digitalHumans = (digitalHumanRes.data.digital_humans || []).map(item => ({
            id: item.id,
            type: '数字人',
            name: item.name || '未命名数字人',
            status: item.status,
            created_at: new Date(item.created_at).toLocaleString()
          }))
          
          // 处理语音识别数据
          const asrTasks = (asrRes.data.asr_tasks || []).map(item => ({
            id: item.id,
            type: '语音识别',
            name: item.name || '未命名识别任务',
            status: item.status,
            created_at: new Date(item.created_at).toLocaleString()
          }))
          
          // 处理图像处理数据
          const comfyuiTasks = (comfyuiRes.data.accessories || []).map(item => ({
            id: item.id,
            type: '图像处理',
            name: item.name || '未命名图像任务',
            status: item.status,
            created_at: new Date(item.created_at).toLocaleString()
          }))
          
          // 合并并排序所有活动数据
          const newActivities = [...voiceClones, ...ttsTasks, ...digitalHumans, ...asrTasks, ...comfyuiTasks]
            .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
          
          // 如果是刷新或首次加载，直接替换数据
          if (isRefresh || this.page === 1) {
            this.recentActivities = newActivities
          } else {
            // 否则追加数据
            this.recentActivities = [...this.recentActivities, ...newActivities]
          }
          
          // 判断是否还有更多数据
          this.hasMoreActivities = newActivities.length >= this.pageSize
          
          // 更新页码
          if (this.hasMoreActivities) {
            this.page++
          }
          
          // 重置加载和刷新状态
          this.loading = false
          if (this.isRefreshing) {
            this.isRefreshing = false
          }
        })
        .catch(error => {
          console.error('获取最近活动失败', error)
          this.loading = false
          this.isRefreshing = false
        })
    },
    
    // 处理下拉刷新
    handleRefresh() {
      this.fetchRecentActivities(true)
    },
    
    // 加载更多数据
    loadMoreActivities() {
      if (!this.loading && this.hasMoreActivities) {
        this.fetchRecentActivities()
      }
    }
  }
}
</script>

<style scoped>
.home-container {
  min-height: 100vh;
  padding: 20px;
  color: rgba(44, 62, 80, 0.9);
  position: relative;
  overflow-x: hidden;
  transition: all 0.3s ease;
  scroll-behavior: smooth;
}

@media (min-width: 768px) {
  .home-container {
    padding: 40px;
  }
}

@media (max-width: 767px) {
  .home-container {
    padding: 15px;
  }
}

.welcome-section {
  position: relative;
  text-align: center;
  margin-bottom: 40px;
  padding: 40px 20px;
  border-radius: 20px;
  overflow: hidden;
}

@media (min-width: 768px) {
  .welcome-section {
    margin-bottom: 70px;
    padding: 60px 20px 70px;
    border-radius: 24px;
  }
}

.welcome-decorations {
  position: absolute;
  inset: 0;
  overflow: hidden;
  z-index: 0;
}

.welcome-content {
  position: relative;
  z-index: 1;
  max-width: 800px;
  margin: 0 auto;
}

.welcome-section h1 {
  font-size: 2.2em;
  margin-bottom: 16px;
  background: linear-gradient(120deg, #333333, #000000);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  line-height: 1.2;
  font-weight: 700;
}

.welcome-section h1 .highlight {
  background: linear-gradient(120deg, #2c3e50, #4a6572);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  font-weight: 800;
  position: relative;
  white-space: nowrap;
}

.welcome-section h1 .highlight::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: -5px;
  width: 100%;
  height: 3px;
  background: linear-gradient(90deg, transparent, rgba(74, 101, 114, 0.7), transparent);
  border-radius: 3px;
}

@media (min-width: 768px) {
  .welcome-section h1 {
    font-size: 3em;
    margin-bottom: 20px;
  }
}

.welcome-section p {
  font-size: 1.1em;
  color: #555555;
  max-width: 600px;
  margin: 0 auto 30px;
  line-height: 1.6;
}

@media (min-width: 768px) {
  .welcome-section p {
    font-size: 1.3em;
    margin-bottom: 36px;
  }
}

.welcome-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
  margin-top: 25px;
}

.action-button {
  padding: 12px 24px;
  font-weight: 600;
  border-radius: 8px;
  transition: all 0.3s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.action-button.el-button--primary {
  background: linear-gradient(90deg, #2c3e50, #4a6572);
  border: none;
  box-shadow: 0 5px 15px rgba(44, 62, 80, 0.2);
}

.action-button.el-button--primary:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(44, 62, 80, 0.3);
}

.action-button.secondary {
  background-color: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(0, 0, 0, 0.1);
  color: #2c3e50;
}

.action-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px rgba(0, 0, 0, 0.1);
}

.action-button.secondary:hover {
  background-color: rgba(255, 255, 255, 0.95);
  border-color: rgba(0, 0, 0, 0.15);
}

.action-button i {
  font-size: 16px;
}

.glass-card {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(15px);
  -webkit-backdrop-filter: blur(15px);
  border-radius: 16px;
  border: 1px solid rgba(200, 200, 200, 0.4);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.06);
  transition: all 0.4s cubic-bezier(0.165, 0.84, 0.44, 1);
}

.feature-card {
  height: auto;
  min-height: 180px;
  padding: 20px;
  margin-bottom: 20px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

@media (min-width: 768px) {
  .feature-card {
    height: 200px;
    padding: 30px;
    margin-bottom: 30px;
  }
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.2);
}

.feature-card .hover-effect {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(45deg, transparent, rgba(0, 0, 0, 0.03));
  opacity: 0;
  transition: opacity 0.3s ease;
}

.feature-card:hover .hover-effect {
  opacity: 1;
}

.card-icon {
  font-size: 2em;
  margin-bottom: 15px;
  color: #333333;
}

@media (min-width: 768px) {
  .card-icon {
    font-size: 2.5em;
    margin-bottom: 20px;
  }
}

.card-content h3 {
  font-size: 1.2em;
  margin-bottom: 8px;
  color: #222222;
}

@media (min-width: 768px) {
  .card-content h3 {
    font-size: 1.4em;
    margin-bottom: 10px;
  }
}

.card-content p {
  color: #555555;
  line-height: 1.3;
  font-size: 0.9em;
}

@media (min-width: 768px) {
  .card-content p {
    line-height: 1.4;
    font-size: 1em;
  }
}

.stats-title {
  margin-top: 40px;
}

.stats-section {
  margin: 25px 0 40px;
}

@media (min-width: 768px) {
  .stats-section {
    margin: 30px 0 60px;
  }
}

.stat-card {
  padding: 20px;
  margin-bottom: 20px;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: all 0.4s ease;
  border-radius: 16px;
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  position: relative;
  z-index: 2;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.05);
  transition: transform 0.3s ease;
}

.stat-card:hover .stat-icon {
  transform: scale(1.1);
}

.stat-icon i {
  font-size: 20px;
  color: #2c3e50;
}

.stat-content {
  position: relative;
  z-index: 2;
  text-align: center;
}

.stat-value {
  font-size: 2.4em;
  font-weight: 700;
  margin-bottom: 8px;
  background: linear-gradient(120deg, #2c3e50, #4a6572);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  display: inline-block;
  line-height: 1;
}

.stat-label {
  color: rgba(85, 85, 85, 0.9);
  font-size: 0.9em;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-bg-circle {
  position: absolute;
  bottom: -20px;
  right: -20px;
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: radial-gradient(circle at center, rgba(0, 0, 0, 0.03), transparent 70%);
  z-index: 1;
}

/* Stat card color variants */
.voice-stat .stat-icon {
  background: linear-gradient(135deg, #f8f9fa, #e9ecef);
  border: 1px solid rgba(44, 62, 80, 0.05);
}

.tts-stat .stat-icon {
  background: linear-gradient(135deg, #f1f8ff, #e1f0ff);
  border: 1px solid rgba(44, 62, 80, 0.05);
}

.digital-human-stat .stat-icon {
  background: linear-gradient(135deg, #fff8f1, #ffedd8);
  border: 1px solid rgba(44, 62, 80, 0.05);
}

.speech-stat .stat-icon {
  background: linear-gradient(135deg, #f1fff8, #dff7eb);
  border: 1px solid rgba(44, 62, 80, 0.05);
}

.image-stat .stat-icon {
  background: linear-gradient(135deg, #f8f1ff, #eedff7);
  border: 1px solid rgba(44, 62, 80, 0.05);
}

@media (min-width: 768px) {
  .stat-card {
    padding: 25px;
    margin-bottom: 0;
  }
  
  .stat-value {
    font-size: 2.8em;
    margin-bottom: 10px;
  }
  
  .stat-label {
    font-size: 1em;
  }
}

@media (max-width: 767px) {
  .stat-card {
    flex-direction: row;
    align-items: center;
    padding: 15px;
  }
  
  .stat-icon {
    width: 40px;
    height: 40px;
    margin-right: 15px;
    margin-bottom: 0;
  }
  
  .stat-icon i {
    font-size: 18px;
  }
  
  .stat-content {
    text-align: left;
  }
  
  .stat-value {
    font-size: 1.8em;
    margin-bottom: 2px;
  }
  
  .stat-label {
    font-size: 0.75em;
  }
}

.recent-section {
  padding: 22px;
  margin-top: 30px;
  border-radius: 18px;
}

@media (min-width: 768px) {
  .recent-section {
    padding: 30px;
    margin-top: 40px;
  }
}

@media (max-width: 767px) {
  .recent-section {
    padding: 18px;
    margin-top: 25px;
  }
}

.table-container {
  overflow-x: auto;
  margin: 0 -15px;
  padding: 0 15px;
  width: calc(100% + 30px);
}

@media (min-width: 768px) {
  .table-container {
    margin: 0;
    padding: 0;
    width: 100%;
  }
}

.section-header {
  margin-bottom: 15px;
}

@media (min-width: 768px) {
  .section-header {
    margin-bottom: 20px;
  }
}

.section-header h2 {
  color: #222222;
  font-size: 1.5em;
  margin: 0;
}

@media (min-width: 768px) {
  .section-header h2 {
    font-size: 1.8em;
  }
}

.hidden-xs-only {
  display: none;
}

@media (min-width: 768px) {
  .hidden-xs-only {
    display: table-cell;
  }
}

.el-table {
  background-color: transparent !important;
  font-size: 13px;
  min-width: 500px;
}

@media (min-width: 768px) {
  .el-table {
    font-size: 14px;
    min-width: auto;
  }
}

.el-table::before {
  display: none;
}

.el-table tr {
  background-color: transparent !important;
}

.el-table td, .el-table th {
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  padding: 10px 5px;
}

@media (min-width: 768px) {
  .el-table td, .el-table th {
    padding: 12px 8px;
  }
}

.el-button--text {
  color: #333333;
  padding: 5px;
}

.el-button--text:hover {
  color: #000000;
}

/* 活动卡片样式 */
.activity-cards {
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 5px 0;
}

.activity-card {
  padding: 15px;
  margin-bottom: 0;
  cursor: pointer;
  transition: all 0.3s ease;
}

.activity-card:active {
  transform: scale(0.98);
}

.activity-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.activity-type {
  font-weight: bold;
  color: #222222;
  font-size: 14px;
}

.activity-card-content {
  margin-bottom: 10px;
}

.activity-name {
  font-size: 13px;
  color: #555555;
  margin-bottom: 5px;
  display: -webkit-box;
  -webkit-line-clamp: 2; /* show up to 2 lines on mobile cards */
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.activity-time {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.6);
}

.activity-card-footer {
  display: flex;
  justify-content: flex-end;
}

/* 加载更多和无数据样式 */
.loading-more {
  text-align: center;
  color: #555555;
  padding: 15px 0;
  font-size: 14px;
}

.no-data {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px 0;
  color: rgba(0, 0, 0, 0.5);
}

.no-data i {
  font-size: 40px;
  margin-bottom: 10px;
}

.no-data p {
  font-size: 14px;
}

/* Tech decorative elements for hero */
.welcome-bg-orb {
  position: absolute;
  top: -60px;
  right: -80px;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle at 30% 30%, rgba(74, 101, 114, 0.25), rgba(44, 62, 80, 0.1) 60%, transparent 70%);
  filter: blur(60px);
  opacity: 0.6;
  pointer-events: none;
  z-index: 0;
  animation: pulse 10s infinite alternate ease-in-out;
}

.welcome-bg-orb.secondary {
  top: 40%;
  left: -120px;
  width: 350px;
  height: 350px;
  background: radial-gradient(circle at 70% 70%, rgba(100, 100, 120, 0.2), rgba(70, 70, 90, 0.05) 60%, transparent 70%);
  opacity: 0.4;
  animation-delay: 2s;
  animation-duration: 14s;
}

@keyframes pulse {
  0% { transform: scale(1); opacity: 0.4; }
  50% { transform: scale(1.05); opacity: 0.5; }
  100% { transform: scale(1); opacity: 0.4; }
}

.welcome-grid {
  position: absolute;
  inset: -20%;
  background-image:
    linear-gradient(rgba(0,0,0,0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0,0,0,0.04) 1px, transparent 1px);
  background-size: 30px 30px;
  transform: rotate(2deg);
  -webkit-mask-image: radial-gradient(ellipse at 60% 50%, rgba(0,0,0,0.8), transparent 75%);
  mask-image: radial-gradient(ellipse at 60% 50%, rgba(0,0,0,0.8), transparent 75%);
  opacity: 0.15;
  pointer-events: none;
  z-index: 0;
}

.tech-circles {
  position: absolute;
  inset: 0;
  z-index: 0;
  opacity: 0.5;
  pointer-events: none;
  overflow: hidden;
}

.tech-circles::before,
.tech-circles::after {
  content: '';
  position: absolute;
  border: 1px dashed rgba(44, 62, 80, 0.1);
  border-radius: 50%;
}

.tech-circles::before {
  top: -150px;
  right: -150px;
  width: 300px;
  height: 300px;
}

.tech-circles::after {
  bottom: -100px;
  left: -100px;
  width: 200px;
  height: 200px;
}

@media (prefers-reduced-motion: reduce) {
  .welcome-bg-orb {
    animation: none;
  }
}

/* Section accent bar */
.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.section-accent {
  width: 36px;
  height: 4px;
  border-radius: 999px;
  background: linear-gradient(90deg, #2c3e50, #4a6572);
  box-shadow: 0 0 10px rgba(0,0,0,0.08);
}

.section-title {
  display: flex;
  align-items: center;
  margin-bottom: 24px;
  padding-left: 10px;
}

.section-title h2 {
  font-size: 1.7em;
  font-weight: 700;
  margin: 0;
  color: #2c3e50;
  position: relative;
}

@media (min-width: 768px) {
  .section-title {
    margin-bottom: 30px;
  }
  
  .section-title h2 {
    font-size: 2em;
  }
}

/* Table text ellipsis */
.table-text-ellipsis {
  display: block;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
/* Merged into main .home-container block above for consistency */

/* Feature card hover refinements */
.feature-card:hover {
  border-color: rgba(0, 0, 0, 0.15);
  box-shadow: 0 14px 46px rgba(0, 0, 0, 0.28);
}

/* Respect reduced motion preferences */
@media (prefers-reduced-motion: reduce) {
  .feature-card,
  .feature-card .hover-effect {
    transition: none;
  }
  .feature-card:hover {
    transform: none;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  }
}

/* Desktop table row hover highlight */
@media (min-width: 768px) {
  .el-table .el-table__body tr:hover > td {
    background-color: rgba(0, 0, 0, 0.03) !important;
  }
}
</style>
<style scoped>
/* ========== Modern tech theme refinements (append) ========== */
/* Typography & smoothing */
.home-container {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  overflow-x: hidden;
}

@media (prefers-color-scheme: dark) {
  .home-container {
    /* Dark mode styles would go here */
    /* Currently not implementing full dark mode */
    /* but setting up the hook for future implementation */
  }
}

/* Welcome headline polish */
.welcome-section h1 {
  letter-spacing: 0.5px;
  text-shadow: 0 4px 24px rgba(74, 101, 114, 0.22);
}
.welcome-section p {
  color: rgba(85, 85, 85, 0.92);
}

/* Elevate generic glass look */
.glass-card {
  position: relative;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

/* Feature section spacing */
.feature-section {
  margin-top: 6px;
}

/* Feature card: minimalist, techy */
.feature-card {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  border-radius: 16px;
}
@media (min-width: 768px) {
  .feature-card {
    gap: 16px;
  }
}

/* Icon container with neon glaze */
.feature-card .card-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: linear-gradient(180deg, rgba(240, 240, 240, 0.9), rgba(210, 210, 210, 0.7));
  border: 1px solid rgba(0, 0, 0, 0.05);
  box-shadow:
    inset 0 0 24px rgba(100, 100, 100, 0.1),
    0 6px 18px rgba(0, 0, 0, 0.05);
  display: grid;
  place-items: center;
  color: #333333;
  flex: 0 0 auto;
}
@media (min-width: 768px) {
  .feature-card .card-icon {
    width: 64px;
    height: 64px;
    border-radius: 16px;
  }
}
.feature-card .card-icon i {
  font-size: 1.5rem;
}

/* Card content tweaks */
.card-content h3 {
  letter-spacing: 0.3px;
}
.card-content p {
  color: rgba(85, 85, 85, 0.9);
}

.card-action {
  margin-top: 12px;
}

.explore-link {
  color: #409EFF;
  font-weight: 600;
  font-size: 0.95em;
  display: inline-flex;
  align-items: center;
  background: linear-gradient(90deg, #409EFF, #2c3e50);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  position: relative;
  padding: 3px 0;
  transition: all 0.3s ease;
}

.explore-link::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: 0;
  width: 100%;
  height: 1px;
  background: linear-gradient(90deg, #409EFF, transparent);
  opacity: 0.5;
  transition: opacity 0.3s ease, transform 0.3s ease;
  transform: scaleX(0.3);
  transform-origin: left;
}

.explore-link i {
  margin-right: 6px;
  font-size: 1em;
  transition: transform 0.3s ease;
  color: #409EFF;
  -webkit-text-fill-color: #409EFF;
}

.feature-card:hover .explore-link i {
  transform: translateX(4px);
}

.feature-card:hover .explore-link::after {
  opacity: 0.8;
  transform: scaleX(1);
}

/* Sheen sweep on hover */
.feature-card .hover-effect {
  background: linear-gradient(100deg, transparent 40%, rgba(0,0,0,0.05) 50%, transparent 60%);
  transform: translateX(-120%);
  transition: transform 0.6s ease, opacity 0.3s ease;
  opacity: 0;
}
.feature-card:hover .hover-effect {
  opacity: 1;
  transform: translateX(120%);
}

/* Click/keyboard feedback */
.feature-card:active {
  transform: translateY(-3px) scale(0.995);
}
.feature-card:focus-visible {
  outline: 2px solid rgba(0, 0, 0, 0.2);
  outline-offset: 2px;
}

/* Stats: add subtle radial energy */
.stat-card {
  display: grid;
  place-items: center;
  background: linear-gradient(180deg, rgba(255,255,255,0.9), rgba(245,245,245,0.7));
  position: relative;
  overflow: hidden;
}
.stat-card::before {
  content: '';
  position: absolute;
  top: -20%;
  left: -10%;
  width: 220px;
  height: 220px;
  background: radial-gradient(closest-side, rgba(100, 100, 100, 0.1), transparent 70%);
  filter: blur(10px);
  pointer-events: none;
}
.stat-value {
  letter-spacing: 0.5px;
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}
.stat-label {
  color: rgba(85, 85, 85, 0.88);
}

/* Table polish */
.el-table th {
  border-bottom: 1px solid rgba(0, 0, 0, 0.10);
}
.el-table td, .el-table th {
  padding: 12px 10px;
}
.el-table .cell {
  color: rgba(0, 0, 0, 0.75);
}
.el-table .el-table__header th .cell {
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  letter-spacing: 0.2px;
}

/* Text buttons accent */
.el-button--text {
  color: #555555;
}
.el-button--text:hover {
  color: #000000;
}

/* Activity name: already clamped; add slight weight */
.activity-name {
  font-weight: 500;
}

/* Mobile fine-tune */
@media (max-width: 767px) {
  .feature-card {
    padding: 16px;
    min-height: 160px;
  }
  
  .stat-card {
    padding: 20px;
    margin-bottom: 15px;
  }
  
  .recent-section {
    padding: 18px;
  }
  
  .welcome-section {
    padding: 30px 15px;
  }
  
  .welcome-section h1 {
    font-size: 2em;
  }
  
  .welcome-section p {
    font-size: 1em;
    margin-bottom: 20px;
  }
  
  .action-button {
    padding: 10px 20px;
    font-size: 0.95em;
  }
  
  .section-title h2 {
    font-size: 1.5em;
  }
  
  .section-accent {
    width: 28px;
    height: 3px;
  }
  
  .activity-type-tag {
    padding: 4px 8px;
    font-size: 12px;
  }
  
  .activity-card {
    padding: 16px;
  }
  
  .no-data i {
    font-size: 36px;
  }
}

/* Respect reduced motion (extends previous block) */
@media (prefers-reduced-motion: reduce) {
  .feature-card,
  .feature-card .hover-effect,
  .el-table tr {
    transition: none !important;
  }
  .feature-card:hover .hover-effect {
    transform: none !important;
  }
}

/* General animation effects */
.glass-card {
  will-change: transform, box-shadow;
}

.glass-card:hover {
  transform: translateY(-5px) translateZ(0);
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.welcome-content,
.feature-card,
.stat-card,
.activity-card {
  animation: fadeIn 0.6s ease forwards;
}

.feature-card:nth-child(2) {
  animation-delay: 0.1s;
}

.feature-card:nth-child(3) {
  animation-delay: 0.2s;
}

.feature-card:nth-child(4) {
  animation-delay: 0.3s;
}

.feature-card:nth-child(5) {
  animation-delay: 0.4s;
}

.stat-card:nth-child(2) {
  animation-delay: 0.1s;
}

.stat-card:nth-child(3) {
  animation-delay: 0.2s;
}

.stat-card:nth-child(4) {
  animation-delay: 0.3s;
}

.stat-card:nth-child(5) {
  animation-delay: 0.4s;
}

@media (prefers-reduced-motion: reduce) {
  .feature-card,
  .stat-card,
  .activity-card,
  .welcome-content {
    animation: none !important;
  }
}

/* Accessibility improvements */
.activity-card:focus,
.feature-card:focus {
  outline: 2px solid rgba(74, 101, 114, 0.4);
  outline-offset: 2px;
}

.action-button:focus {
  outline: 2px solid rgba(74, 101, 114, 0.4);
  outline-offset: 2px;
}
</style>