<template>
  <div class="home-container">
    <!-- 顶部欢迎区域 -->
    <div class="welcome-section">
      <h1>欢迎使用 Virtual Human Studio</h1>
      <p>探索AI驱动的数字人创作平台</p>
    </div>

    <!-- 功能卡片区域 -->
    <el-row :gutter="30" class="feature-section">
      <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="(card, index) in featureCards" :key="index">
        <div class="glass-card feature-card" @click="navigateTo(card.route)">
          <div class="card-icon">
            <i :class="card.icon"></i>
          </div>
          <div class="card-content">
            <h3>{{ card.title }}</h3>
            <p>{{ card.description }}</p>
          </div>
          <div class="hover-effect"></div>
        </div>
      </el-col>
    </el-row>
    
    <!-- 统计数据区域 -->
    <el-row :gutter="30" class="stats-section">
      <el-col :xs="12" :sm="8" :md="6" :lg="4" v-for="(stat, index) in stats" :key="index">
        <div class="glass-card stat-card">
          <div class="stat-value">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </el-col>
    </el-row>
    
    <!-- 最近活动区域 -->
    <div class="glass-card recent-section">
      <div class="section-header">
        <h2>最近活动</h2>
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
            @click="viewDetail(activity)">
            <div class="activity-card-header">
              <div class="activity-type">{{ activity.type }}</div>
              <el-tag size="mini" :type="getStatusType(activity.status)">{{ activity.status }}</el-tag>
            </div>
            <div class="activity-card-content">
              <div class="activity-name">{{ activity.name }}</div>
              <div class="activity-time">{{ activity.created_at }}</div>
            </div>
            <div class="activity-card-footer">
              <el-button type="text" size="mini" @click.stop="viewDetail(activity)">查看详情</el-button>
            </div>
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
          :header-cell-style="{background: 'transparent', color: '#333'}"
          :cell-style="{background: 'transparent', color: '#333'}"
          size="small">
        <el-table-column prop="type" label="类型" width="180">
          <template slot-scope="scope">
            <span>{{ scope.row.type }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="180">
          <template slot-scope="scope">
            <span>{{ scope.row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template slot-scope="scope">
            <el-tag size="mini" :type="getStatusType(scope.row.status)">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间"></el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template slot-scope="scope">
            <el-button type="text" size="mini" @click="viewDetail(scope.row)">查看</el-button>
          </template>
        </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Home',
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
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #333;
}

@media (min-width: 768px) {
  .home-container {
    padding: 40px;
  }
}

.welcome-section {
  text-align: center;
  margin-bottom: 30px;
}

@media (min-width: 768px) {
  .welcome-section {
    margin-bottom: 60px;
  }
}

.welcome-section h1 {
  font-size: 1.8em;
  margin-bottom: 10px;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

@media (min-width: 768px) {
  .welcome-section h1 {
    font-size: 2.5em;
  }
}

.welcome-section p {
  font-size: 1em;
  color: #b3e5fc;
}

@media (min-width: 768px) {
  .welcome-section p {
    font-size: 1.2em;
  }
}

.glass-card {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
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
  background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.1));
  opacity: 0;
  transition: opacity 0.3s ease;
}

.feature-card:hover .hover-effect {
  opacity: 1;
}

.card-icon {
  font-size: 2em;
  margin-bottom: 15px;
  color: #64b5f6;
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
  color: #fff;
}

@media (min-width: 768px) {
  .card-content h3 {
    font-size: 1.4em;
    margin-bottom: 10px;
  }
}

.card-content p {
  color: #b3e5fc;
  line-height: 1.3;
  font-size: 0.9em;
}

@media (min-width: 768px) {
  .card-content p {
    line-height: 1.4;
    font-size: 1em;
  }
}

.stats-section {
  margin: 30px 0;
}

@media (min-width: 768px) {
  .stats-section {
    margin: 40px 0;
  }
}

.stat-card {
  padding: 15px;
  text-align: center;
  margin-bottom: 15px;
}

@media (min-width: 768px) {
  .stat-card {
    padding: 30px;
    margin-bottom: 0;
  }
}

.stat-value {
  font-size: 1.8em;
  font-weight: bold;
  margin-bottom: 5px;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

@media (min-width: 768px) {
  .stat-value {
    font-size: 2.5em;
    margin-bottom: 10px;
  }
}

.stat-label {
  color: #b3e5fc;
  font-size: 0.9em;
}

@media (min-width: 768px) {
  .stat-label {
    font-size: 1.1em;
  }
}

.recent-section {
  padding: 15px;
  margin-top: 30px;
}

@media (min-width: 768px) {
  .recent-section {
    padding: 30px;
    margin-top: 40px;
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
  color: #fff;
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
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding: 10px 5px;
}

@media (min-width: 768px) {
  .el-table td, .el-table th {
    padding: 12px 8px;
  }
}

.el-button--text {
  color: #64b5f6;
  padding: 5px;
}

.el-button--text:hover {
  color: #1976d2;
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
  color: #fff;
  font-size: 14px;
}

.activity-card-content {
  margin-bottom: 10px;
}

.activity-name {
  font-size: 13px;
  color: #b3e5fc;
  margin-bottom: 5px;
  word-break: break-all;
}

.activity-time {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

.activity-card-footer {
  display: flex;
  justify-content: flex-end;
}

/* 加载更多和无数据样式 */
.loading-more {
  text-align: center;
  color: #b3e5fc;
  padding: 15px 0;
  font-size: 14px;
}

.no-data {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px 0;
  color: rgba(255, 255, 255, 0.5);
}

.no-data i {
  font-size: 40px;
  margin-bottom: 10px;
}

.no-data p {
  font-size: 14px;
}
</style>