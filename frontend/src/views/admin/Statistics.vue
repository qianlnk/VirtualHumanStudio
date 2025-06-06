<template>
  <div class="admin-statistics-container">
    <div class="page-header">
      <h2>数据统计</h2>
      
      <!-- 日期选择器 -->
      <div class="date-filter">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          align="right"
          unlink-panels
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          :picker-options="pickerOptions"
          @change="handleDateChange"
        >
        </el-date-picker>
        
        <el-button type="primary" size="small" @click="fetchAllData">刷新数据</el-button>
      </div>
    </div>
    
    <!-- 统计卡片 -->
    <div v-loading="loading.cards" class="stat-cards">
      <el-row :gutter="20">
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon">
              <i class="el-icon-user"></i>
            </div>
            <div class="stat-content">
              <div class="stat-title">活跃用户 (DAU)</div>
              <div class="stat-value">{{ summaryData.activeUsers }}</div>
              <div class="stat-desc">{{ dateRange ? '所选时间段平均' : '今日' }}</div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon blue">
              <i class="el-icon-user-solid"></i>
            </div>
            <div class="stat-content">
              <div class="stat-title">新增用户</div>
              <div class="stat-value">{{ summaryData.newUsers }}</div>
              <div class="stat-desc">{{ dateRange ? '所选时间段总计' : '今日' }}</div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon green">
              <i class="el-icon-s-data"></i>
            </div>
            <div class="stat-content">
              <div class="stat-title">登录次数</div>
              <div class="stat-value">{{ summaryData.totalLogins }}</div>
              <div class="stat-desc">{{ dateRange ? '所选时间段总计' : '今日' }}</div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon purple">
              <i class="el-icon-s-operation"></i>
            </div>
            <div class="stat-content">
              <div class="stat-title">请求总数</div>
              <div class="stat-value">{{ summaryData.totalRequests }}</div>
              <div class="stat-desc">{{ dateRange ? '所选时间段总计' : '今日' }}</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
    
    <!-- 图表区域 -->
    <div class="chart-container">
      <el-row :gutter="20">
        <!-- 活跃用户趋势图 -->
        <el-col :xs="24" :lg="12">
          <el-card shadow="hover" class="chart-card" v-loading="loading.userTrends">
            <div slot="header" class="chart-header">
              <span>用户数据趋势</span>
            </div>
            <div class="chart" id="userTrendsChart"></div>
          </el-card>
        </el-col>
        
        <!-- 模块使用占比图 -->
        <el-col :xs="24" :lg="12">
          <el-card shadow="hover" class="chart-card" v-loading="loading.moduleUsage">
            <div slot="header" class="chart-header">
              <span>模块使用占比</span>
            </div>
            <div class="chart" id="moduleUsageChart"></div>
          </el-card>
        </el-col>
      </el-row>
      
      <el-row :gutter="20" style="margin-top: 20px;">
        <!-- 用户活跃度排行 -->
        <el-col :xs="24" :lg="12">
          <el-card shadow="hover" class="chart-card" v-loading="loading.userActivity">
            <div slot="header" class="chart-header">
              <span>用户活跃度排行</span>
            </div>
            <el-table :data="userActivity" stripe style="width: 100%">
              <el-table-column prop="username" label="用户名" width="180"></el-table-column>
              <el-table-column prop="count" label="操作次数" width="120"></el-table-column>
              <el-table-column label="活跃度">
                <template slot-scope="scope">
                  <el-progress :percentage="calculatePercentage(scope.row.count)" :color="getProgressColor(scope.row.count)"></el-progress>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
        
        <!-- 最近登录记录 -->
        <el-col :xs="24" :lg="12">
          <el-card shadow="hover" class="chart-card" v-loading="loading.loginLogs">
            <div slot="header" class="chart-header">
              <span>最近登录记录</span>
              <el-button type="text" @click="viewMoreLogs">查看更多</el-button>
            </div>
            <el-table :data="loginLogs" stripe style="width: 100%">
              <el-table-column prop="username" label="用户名" width="120"></el-table-column>
              <el-table-column prop="login_time" label="登录时间" width="180">
                <template slot-scope="scope">
                  {{ formatDate(scope.row.login_time) }}
                </template>
              </el-table-column>
              <el-table-column prop="login_ip" label="IP地址"></el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </div>
    
    <!-- 登录日志弹窗 -->
    <el-dialog title="用户登录日志" :visible.sync="loginLogsDialogVisible" width="80%">
      <div v-loading="loading.allLoginLogs">
        <el-table :data="allLoginLogs" stripe style="width: 100%">
          <el-table-column prop="username" label="用户名" width="120"></el-table-column>
          <el-table-column prop="login_time" label="登录时间" width="180">
            <template slot-scope="scope">
              {{ formatDate(scope.row.login_time) }}
            </template>
          </el-table-column>
          <el-table-column prop="login_ip" label="IP地址" width="150"></el-table-column>
          <el-table-column prop="user_id" label="用户ID" width="80"></el-table-column>
        </el-table>
        
        <!-- 分页 -->
        <div class="pagination-container" v-if="loginLogsTotal > loginLogsPageSize">
          <el-pagination
            @size-change="handleLoginLogsSizeChange"
            @current-change="handleLoginLogsPageChange"
            :current-page="loginLogsPage"
            :page-sizes="[10, 20, 50, 100]"
            :page-size="loginLogsPageSize"
            layout="total, sizes, prev, pager, next, jumper"
            :total="loginLogsTotal">
          </el-pagination>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
// 引入图表库
import * as echarts from 'echarts'

export default {
  name: 'AdminStatistics',
  data() {
    return {
      loading: {
        cards: false,
        userTrends: false,
        moduleUsage: false,
        userActivity: false,
        loginLogs: false,
        allLoginLogs: false
      },
      dateRange: null,
      pickerOptions: {
        shortcuts: [
          {
            text: '最近一周',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
              picker.$emit('pick', [start, end]);
            }
          },
          {
            text: '最近一个月',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
              picker.$emit('pick', [start, end]);
            }
          },
          {
            text: '最近三个月',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 90);
              picker.$emit('pick', [start, end]);
            }
          }
        ]
      },
      summaryData: {
        activeUsers: 0,
        newUsers: 0,
        totalLogins: 0,
        totalRequests: 0
      },
      userActivity: [],
      loginLogs: [],
      statisticsData: [],
      moduleUsageData: [],
      
      // 登录日志弹窗数据
      loginLogsDialogVisible: false,
      allLoginLogs: [],
      loginLogsPage: 1,
      loginLogsPageSize: 20,
      loginLogsTotal: 0,
      
      // 图表实例
      userTrendsChart: null,
      moduleUsageChart: null
    }
  },
  created() {
    this.fetchAllData()
  },
  mounted() {
    // 初始化图表
    this.initCharts()
    
    // 监听窗口大小变化，重绘图表
    window.addEventListener('resize', this.resizeCharts)
  },
  beforeDestroy() {
    // 移除事件监听
    window.removeEventListener('resize', this.resizeCharts)
    
    // 销毁图表实例
    if (this.userTrendsChart) {
      this.userTrendsChart.dispose()
    }
    if (this.moduleUsageChart) {
      this.moduleUsageChart.dispose()
    }
  },
  methods: {
    // 获取所有数据
    fetchAllData() {
      this.fetchStatistics()
      this.fetchModuleUsage()
      this.fetchLoginLogs()
    },
    
    // 处理日期变化
    handleDateChange() {
      this.fetchAllData()
    },
    
    // 获取统计数据
    fetchStatistics() {
      this.loading.cards = true
      this.loading.userTrends = true
      
      let url = '/api/admin/statistics/users'
      
      if (this.dateRange) {
        const startDate = this.formatDateForAPI(this.dateRange[0])
        const endDate = this.formatDateForAPI(this.dateRange[1])
        url += `?start_date=${startDate}&end_date=${endDate}`
      }
      
      this.$http.get(url)
        .then(response => {
          this.statisticsData = response.data.statistics || []
          
          // 计算汇总数据
          this.calculateSummaryData()
          
          // 更新用户趋势图表
          this.updateUserTrendsChart()
        })
        .catch(error => {
          console.error('获取统计数据失败', error)
          this.$message.error('获取统计数据失败')
        })
        .finally(() => {
          this.loading.cards = false
          this.loading.userTrends = false
        })
    },
    
    // 获取模块使用数据
    fetchModuleUsage() {
      this.loading.moduleUsage = true
      this.loading.userActivity = true
      
      let url = '/api/admin/statistics/modules'
      
      if (this.dateRange) {
        const startDate = this.formatDateForAPI(this.dateRange[0])
        const endDate = this.formatDateForAPI(this.dateRange[1])
        url += `?start_date=${startDate}&end_date=${endDate}`
      }
      
      this.$http.get(url)
        .then(response => {
          this.moduleUsageData = response.data.module_usage || []
          this.userActivity = response.data.user_activity || []
          
          // 更新模块使用图表
          this.updateModuleUsageChart()
        })
        .catch(error => {
          console.error('获取模块使用数据失败', error)
          this.$message.error('获取模块使用数据失败')
        })
        .finally(() => {
          this.loading.moduleUsage = false
          this.loading.userActivity = false
        })
    },
    
    // 获取登录日志
    fetchLoginLogs(isDialog = false) {
      if (isDialog) {
        this.loading.allLoginLogs = true
      } else {
        this.loading.loginLogs = true
      }
      
      let url = '/api/admin/statistics/login-logs'
      let params = {}
      
      if (this.dateRange) {
        params.start_date = this.formatDateForAPI(this.dateRange[0])
        params.end_date = this.formatDateForAPI(this.dateRange[1])
      }
      
      if (isDialog) {
        params.page = this.loginLogsPage
        params.size = this.loginLogsPageSize
      } else {
        params.page = 1
        params.size = 5 // 首页只显示5条记录
      }
      
      this.$http.get(url, { params })
        .then(response => {
          if (isDialog) {
            this.allLoginLogs = response.data.logs || []
            this.loginLogsTotal = response.data.total || 0
          } else {
            this.loginLogs = response.data.logs || []
          }
        })
        .catch(error => {
          console.error('获取登录日志失败', error)
          this.$message.error('获取登录日志失败')
        })
        .finally(() => {
          if (isDialog) {
            this.loading.allLoginLogs = false
          } else {
            this.loading.loginLogs = false
          }
        })
    },
    
    // 计算汇总数据
    calculateSummaryData() {
      if (this.statisticsData.length === 0) {
        this.summaryData = {
          activeUsers: 0,
          newUsers: 0,
          totalLogins: 0,
          totalRequests: 0
        }
        return
      }
      
      // 计算活跃用户平均值
      const activeUsersSum = this.statisticsData.reduce((sum, item) => sum + item.active_users, 0)
      const activeUsersAvg = Math.round(activeUsersSum / this.statisticsData.length)
      
      // 计算其他累计值
      const newUsersSum = this.statisticsData.reduce((sum, item) => sum + item.new_users, 0)
      const totalLoginsSum = this.statisticsData.reduce((sum, item) => sum + item.total_logins, 0)
      const totalRequestsSum = this.statisticsData.reduce((sum, item) => sum + item.total_requests, 0)
      
      this.summaryData = {
        activeUsers: activeUsersAvg,
        newUsers: newUsersSum,
        totalLogins: totalLoginsSum,
        totalRequests: totalRequestsSum
      }
    },
    
    // 初始化图表
    initCharts() {
      // 初始化用户趋势图表
      this.userTrendsChart = echarts.init(document.getElementById('userTrendsChart'))
      
      // 初始化模块使用图表
      this.moduleUsageChart = echarts.init(document.getElementById('moduleUsageChart'))
    },
    
    // 更新用户趋势图表
    updateUserTrendsChart() {
      if (!this.userTrendsChart) return
      
      const dates = this.statisticsData.map(item => this.formatDate(item.date, 'short'))
      const activeUsers = this.statisticsData.map(item => item.active_users)
      const newUsers = this.statisticsData.map(item => item.new_users)
      const totalLogins = this.statisticsData.map(item => item.total_logins)
      
      const option = {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'shadow'
          }
        },
        legend: {
          data: ['活跃用户', '新增用户', '登录次数'],
          bottom: 0
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '10%',
          top: '3%',
          containLabel: true
        },
        xAxis: {
          type: 'category',
          data: dates,
          axisLabel: {
            rotate: 45,
            margin: 10
          }
        },
        yAxis: [
          {
            type: 'value',
            name: '用户数',
            position: 'left'
          },
          {
            type: 'value',
            name: '登录次数',
            position: 'right'
          }
        ],
        series: [
          {
            name: '活跃用户',
            type: 'line',
            smooth: true,
            data: activeUsers,
            itemStyle: {
              color: '#409EFF'
            }
          },
          {
            name: '新增用户',
            type: 'bar',
            data: newUsers,
            itemStyle: {
              color: '#67C23A'
            }
          },
          {
            name: '登录次数',
            type: 'line',
            smooth: true,
            yAxisIndex: 1,
            data: totalLogins,
            itemStyle: {
              color: '#E6A23C'
            }
          }
        ]
      }
      
      this.userTrendsChart.setOption(option)
    },
    
    // 更新模块使用图表
    updateModuleUsageChart() {
      if (!this.moduleUsageChart) return
      
      const data = this.moduleUsageData.map(item => ({
        name: item.module_name,
        value: item.count
      }))
      
      const option = {
        tooltip: {
          trigger: 'item',
          formatter: '{a} <br/>{b}: {c} ({d}%)'
        },
        legend: {
          type: 'scroll',
          orient: 'vertical',
          right: 10,
          top: 20,
          bottom: 20,
          data: data.map(item => item.name)
        },
        series: [
          {
            name: '模块使用',
            type: 'pie',
            radius: ['40%', '70%'],
            avoidLabelOverlap: false,
            itemStyle: {
              borderRadius: 10,
              borderColor: '#fff',
              borderWidth: 2
            },
            label: {
              show: false,
              position: 'center'
            },
            emphasis: {
              label: {
                show: true,
                fontSize: '18',
                fontWeight: 'bold'
              }
            },
            labelLine: {
              show: false
            },
            data: data
          }
        ]
      }
      
      this.moduleUsageChart.setOption(option)
    },
    
    // 重绘图表
    resizeCharts() {
      if (this.userTrendsChart) {
        this.userTrendsChart.resize()
      }
      if (this.moduleUsageChart) {
        this.moduleUsageChart.resize()
      }
    },
    
    // 格式化日期
    formatDate(dateString, type = 'full') {
      if (!dateString) return ''
      const date = new Date(dateString)
      
      if (type === 'short') {
        return `${date.getMonth() + 1}/${date.getDate()}`
      }
      
      return date.toLocaleString()
    },
    
    // 格式化日期为API格式
    formatDateForAPI(date) {
      if (!date) return ''
      
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      
      return `${year}-${month}-${day}`
    },
    
    // 计算活跃度百分比
    calculatePercentage(value) {
      if (this.userActivity.length === 0) return 0
      
      const maxValue = Math.max(...this.userActivity.map(item => item.count))
      return Math.round((value / maxValue) * 100)
    },
    
    // 获取进度条颜色
    getProgressColor(value) {
      const percentage = this.calculatePercentage(value)
      
      if (percentage >= 80) {
        return '#F56C6C' // 红色
      } else if (percentage >= 60) {
        return '#E6A23C' // 橙色
      } else if (percentage >= 40) {
        return '#67C23A' // 绿色
      } else {
        return '#409EFF' // 蓝色
      }
    },
    
    // 查看更多登录日志
    viewMoreLogs() {
      this.loginLogsDialogVisible = true
      this.fetchLoginLogs(true)
    },
    
    // 处理登录日志页码变化
    handleLoginLogsPageChange(page) {
      this.loginLogsPage = page
      this.fetchLoginLogs(true)
    },
    
    // 处理登录日志每页显示数量变化
    handleLoginLogsSizeChange(size) {
      this.loginLogsPageSize = size
      this.loginLogsPage = 1
      this.fetchLoginLogs(true)
    }
  }
}
</script>

<style scoped>
.admin-statistics-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 22px;
  color: #303133;
}

.date-filter {
  display: flex;
  align-items: center;
}

.date-filter .el-button {
  margin-left: 10px;
}

.stat-cards {
  margin-bottom: 30px;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 15px;
  margin-bottom: 20px;
  height: 120px;
}

.stat-icon {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background-color: #f0f2f5;
  margin-right: 15px;
  color: #409EFF;
  font-size: 28px;
}

.stat-icon.blue {
  color: #1890ff;
}

.stat-icon.green {
  color: #52c41a;
}

.stat-icon.purple {
  color: #722ed1;
}

.stat-content {
  flex: 1;
}

.stat-title {
  font-size: 16px;
  color: #909399;
  margin-bottom: 10px;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 5px;
}

.stat-desc {
  font-size: 12px;
  color: #909399;
}

.chart-container {
  margin-top: 20px;
}

.chart-card {
  margin-bottom: 20px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart {
  height: 350px;
}

.pagination-container {
  margin-top: 20px;
  text-align: center;
}
</style> 