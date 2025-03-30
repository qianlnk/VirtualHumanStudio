<template>
  <div class="home-container">
    <el-row :gutter="20">
      <el-col :span="6" v-for="(card, index) in featureCards" :key="index">
        <el-card class="feature-card" shadow="hover" @click.native="navigateTo(card.route)">
          <div class="card-icon">
            <i :class="card.icon"></i>
          </div>
          <div class="card-content">
            <h3>{{ card.title }}</h3>
            <p>{{ card.description }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" class="stats-row">
      <el-col :span="8" v-for="(stat, index) in stats" :key="index">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-value">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" class="recent-row">
      <el-col :span="24">
        <el-card class="recent-card" shadow="hover">
          <div slot="header" class="clearfix">
            <span>最近活动</span>
          </div>
          <el-table :data="recentActivities" style="width: 100%">
            <el-table-column prop="type" label="类型" width="180"></el-table-column>
            <el-table-column prop="name" label="名称" width="180"></el-table-column>
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
        </el-card>
      </el-col>
    </el-row>
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
        }
      ],
      stats: [
        { label: '音色克隆任务', value: 0 },
        { label: 'TTS任务', value: 0 },
        { label: '数字人任务', value: 0 }
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
      // 根据不同类型跳转到不同的详情页
      if (row.type === '音色克隆') {
        this.$router.push(`/voice-clone/${row.id}`)
      } else if (row.type === 'TTS') {
        this.$router.push(`/tts/${row.id}`)
      } else if (row.type === '数字人') {
        this.$router.push(`/digital-human/${row.id}`)
      }
    },
    fetchStats() {
      // 获取统计数据
      Promise.all([
        this.$http.get('/api/voice/clones?size=1'),
        this.$http.get('/api/tts?size=1'),
        this.$http.get('/api/digital-human?size=1')
      ])
        .then(([voiceCloneRes, ttsRes, digitalHumanRes]) => {
          this.stats[0].value = voiceCloneRes.data.total || 0
          this.stats[1].value = ttsRes.data.total || 0
          this.stats[2].value = digitalHumanRes.data.total || 0
        })
        .catch(error => {
          console.error('获取统计数据失败', error)
        })
    },
    fetchRecentActivities() {
      // 获取最近活动
      Promise.all([
        this.$http.get('/api/voice/clones?size=5'),
        this.$http.get('/api/tts?size=5'),
        this.$http.get('/api/digital-human?size=5')
      ])
        .then(([voiceCloneRes, ttsRes, digitalHumanRes]) => {
          const voiceClones = (voiceCloneRes.data.voice_clones || []).map(item => ({
            id: item.id,
            type: '音色克隆',
            name: item.name,
            status: item.status,
            created_at: new Date(item.created_at).toLocaleString()
          }))
          
          const ttsTasks = (ttsRes.data.tts_tasks || []).map(item => ({
            id: item.id,
            type: 'TTS',
            name: item.name,
            status: item.status,
            created_at: new Date(item.created_at).toLocaleString()
          }))
          
          const digitalHumans = (digitalHumanRes.data.digital_humans || []).map(item => ({
            id: item.id,
            type: '数字人',
            name: item.name,
            status: item.status,
            created_at: new Date(item.created_at).toLocaleString()
          }))
          
          // 合并并按时间排序
          this.recentActivities = [...voiceClones, ...ttsTasks, ...digitalHumans]
            .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
            .slice(0, 10) // 只显示最近10条
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
  padding: 20px;
}

.feature-card {
  height: 180px;
  margin-bottom: 20px;
  cursor: pointer;
  transition: transform 0.3s;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.feature-card:hover {
  transform: translateY(-5px);
}

.card-icon {
  font-size: 48px;
  color: #409EFF;
  margin-bottom: 15px;
}

.card-content {
  text-align: center;
}

.card-content h3 {
  margin: 0 0 10px 0;
  font-size: 18px;
  color: #303133;
}

.card-content p {
  margin: 0;
  font-size: 14px;
  color: #909399;
}

.stats-row {
  margin-top: 20px;
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
  padding: 20px 0;
}

.stat-value {
  font-size: 36px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
}

.stat-label {
  font-size: 16px;
  color: #606266;
}

.recent-row {
  margin-top: 20px;
}

.recent-card .el-table {
  margin-top: 10px;
}
</style>