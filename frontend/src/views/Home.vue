<template>
  <div class="home-container">
    <!-- 顶部欢迎区域 -->
    <div class="welcome-section">
      <h1>欢迎使用 Virtual Human Studio</h1>
      <p>探索AI驱动的数字人创作平台</p>
    </div>

    <!-- 功能卡片区域 -->
    <el-row :gutter="30" class="feature-section">
      <el-col :span="6" v-for="(card, index) in featureCards" :key="index">
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
      <el-col :span="4" v-for="(stat, index) in stats" :key="index">
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
      <el-table 
        :data="recentActivities" 
        style="width: 100%"
        :header-cell-style="{background: 'transparent', color: '#333'}"
        :cell-style="{background: 'transparent', color: '#333'}">
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
        <el-table-column prop="status" label="状态">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间"></el-table-column>
        <el-table-column label="操作" width="120">
          <template slot-scope="scope">
            <el-button type="text" size="small" @click="viewDetail(scope.row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Home',
  data() {
    return {
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
    this.fetchStats()
    this.fetchRecentActivities()
  },
  methods: {
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
    fetchRecentActivities() {
      Promise.all([
        this.$http.get('/api/voice/clones?size=5'),
        this.$http.get('/api/tts?size=5'),
        this.$http.get('/api/digital-human?size=5'),
        this.$http.get('/api/asr?size=5'),
        this.$http.get('/api/accessory?size=5')
      ])
        .then(([voiceCloneRes, ttsRes, digitalHumanRes, asrRes, comfyuiRes]) => {
          // 处理音色克隆数据，确保type和name字段正确设置
          const voiceClones = (voiceCloneRes.data.voice_clones || []).map(item => {
            console.log('音色克隆项:', item) // 调试日志
            return {
              id: item.id,
              type: '音色克隆', // 显式设置类型
              name: item.name || '未命名音色', // 提供默认名称
              status: item.status,
              created_at: new Date(item.created_at).toLocaleString()
            }
          })
          
          // 处理TTS任务数据
          const ttsTasks = (ttsRes.data.tts_tasks || []).map(item => {
            console.log('TTS项:', item) // 调试日志
            return {
              id: item.id,
              type: 'TTS', // 显式设置类型
              name: item.name || '未命名TTS任务', // 提供默认名称
              status: item.status,
              created_at: new Date(item.created_at).toLocaleString()
            }
          })
          
          // 处理数字人数据
          const digitalHumans = (digitalHumanRes.data.digital_humans || []).map(item => {
            console.log('数字人项:', item) // 调试日志
            return {
              id: item.id,
              type: '数字人', // 显式设置类型
              name: item.name || '未命名数字人', // 提供默认名称
              status: item.status,
              created_at: new Date(item.created_at).toLocaleString()
            }
          })
          
          // 处理语音识别数据
          const asrTasks = (asrRes.data.asr_tasks || []).map(item => {
            console.log('语音识别项:', item) // 调试日志
            return {
              id: item.id,
              type: '语音识别', // 显式设置类型
              name: item.name || '未命名识别任务', // 提供默认名称
              status: item.status,
              created_at: new Date(item.created_at).toLocaleString()
            }
          })
          
          // 处理图像处理数据
          const comfyuiTasks = (comfyuiRes.data.accessories || []).map(item => {
            console.log('图像处理项:', item) // 调试日志
            return {
              id: item.id,
              type: '图像处理', // 显式设置类型
              name: item.name || '未命名图像任务', // 提供默认名称
              status: item.status,
              created_at: new Date(item.created_at).toLocaleString()
            }
          })
          
          this.recentActivities = [...voiceClones, ...ttsTasks, ...digitalHumans, ...asrTasks, ...comfyuiTasks]
            .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
            .slice(0, 10)
        })
        .catch(error => {
          console.error('获取最近活动失败', error)
        })
    }
  }
}
</script>

<style scoped>
.home-container {
  min-height: 100vh;
  padding: 40px;
  background: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
  color: #333;
}

.welcome-section {
  text-align: center;
  margin-bottom: 60px;
}

.welcome-section h1 {
  font-size: 2.5em;
  margin-bottom: 10px;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.welcome-section p {
  font-size: 1.2em;
  color: #b3e5fc;
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
  height: 200px;
  padding: 30px;
  margin-bottom: 30px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
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
  font-size: 2.5em;
  margin-bottom: 20px;
  color: #64b5f6;
}

.card-content h3 {
  font-size: 1.4em;
  margin-bottom: 10px;
  color: #fff;
}

.card-content p {
  color: #b3e5fc;
  line-height: 1.4;
}

.stats-section {
  margin: 40px 0;
}

.stat-card {
  padding: 30px;
  text-align: center;
}

.stat-value {
  font-size: 2.5em;
  font-weight: bold;
  margin-bottom: 10px;
  background: linear-gradient(120deg, #64b5f6, #1976d2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.stat-label {
  color: #b3e5fc;
  font-size: 1.1em;
}

.recent-section {
  padding: 30px;
  margin-top: 40px;
}

.section-header {
  margin-bottom: 20px;
}

.section-header h2 {
  color: #fff;
  font-size: 1.8em;
  margin: 0;
}

.el-table {
  background-color: transparent !important;
}

.el-table::before {
  display: none;
}

.el-table tr {
  background-color: transparent !important;
}

.el-table td, .el-table th {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.el-button--text {
  color: #64b5f6;
}

.el-button--text:hover {
  color: #1976d2;
}
</style>