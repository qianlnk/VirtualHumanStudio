<template>
  <div class="membership-container">
    <div class="page-header">
      <h2>会员中心</h2>
    </div>
    
    <!-- 顶部卡片：我的会员和今日使用量 -->
    <el-row :gutter="20" class="top-cards">
      <!-- 当前会员状态 -->
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <el-card class="membership-status" shadow="hover">
          <div class="membership-header">
            <i class="el-icon-medal"></i>
            <h3>我的会员</h3>
          </div>
          <div v-loading="loading" class="membership-content">
            <template v-if="membership">
              <div class="membership-level" :class="getLevelClass(membership.level)">
                <span class="level-icon"><i class="el-icon-trophy"></i></span>
                <span class="level-name">{{ getLevelName(membership.level) }}</span>
              </div>
              
              <div class="membership-info">
                <div class="info-item">
                  <span class="label">会员状态：</span>
                  <span class="value">{{ membership.status === 'active' ? '生效中' : '已过期' }}</span>
                </div>
                <div class="info-item">
                  <span class="label">到期时间：</span>
                  <span class="value">{{ formatDate(membership.expire_date) }}</span>
                </div>
                <div class="info-item">
                  <span class="label">自动续费：</span>
                  <span class="value">{{ membership.auto_renew ? '已开启' : '已关闭' }}</span>
                  <el-button 
                    v-if="membership.auto_renew && membership.level !== 'free'" 
                    type="text" 
                    @click="handleCancelAutoRenew"
                    :loading="cancelling">取消自动续费</el-button>
                </div>
              </div>
            </template>
            <div v-else class="no-membership">
              <i class="el-icon-warning"></i>
              <p>暂无会员信息</p>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <!-- 每日使用统计 -->
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <el-card class="usage-card" shadow="hover">
          <div class="usage-header">
            <i class="el-icon-data-analysis"></i>
            <h3>今日使用量</h3>
            <el-button 
              type="text" 
              icon="el-icon-refresh" 
              class="refresh-btn"
              @click="fetchDailyUsage" 
              :loading="usageLoading">
              刷新
            </el-button>
          </div>
          <div v-loading="usageLoading" class="usage-content">
            <div class="usage-progress-container">
              <el-progress 
                :percentage="usagePercentage" 
                :color="usageColor"
                :format="formatUsage"
                :stroke-width="18"></el-progress>
                
              <div class="usage-info">
                <span>{{ usageCountText }}</span>
                <span>{{ remainingText }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 待审核订单区域 -->
    <el-row v-if="pendingOrders.length > 0" class="pending-orders-row">
      <el-col :span="24">
        <el-card class="pending-orders-card" shadow="hover">
          <div class="pending-orders-header">
            <i class="el-icon-time"></i>
            <h3>待审核订单</h3>
            <el-button 
              type="text" 
              icon="el-icon-refresh" 
              class="refresh-btn"
              @click="fetchPendingOrders" 
              :loading="orderLoading">
              刷新
            </el-button>
          </div>
          <div class="pending-orders-content">
            <el-table 
              :data="pendingOrders" 
              stripe 
              style="width: 100%"
              v-loading="orderLoading">
              <el-table-column prop="created_at" label="提交时间" width="180">
                <template slot-scope="scope">
                  {{ formatDateTime(scope.row.created_at) }}
                </template>
              </el-table-column>
              <el-table-column prop="plan_name" label="会员方案" width="120"></el-table-column>
              <el-table-column prop="payment_method" label="支付方式" width="100">
                <template slot-scope="scope">
                  {{ scope.row.payment_method === 'wechat' ? '微信支付' : '支付宝' }}
                </template>
              </el-table-column>
              <el-table-column prop="amount" label="金额" width="100">
                <template slot-scope="scope">
                  ¥{{ scope.row.amount }}
                </template>
              </el-table-column>
              <el-table-column prop="payment_remark" label="支付备注"></el-table-column>
              <el-table-column prop="status" label="状态" width="120">
                <template>
                  <el-tag type="warning">等待审核</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <div class="pending-tip">
              <i class="el-icon-info"></i> 
              订单审核通常在1-2个工作日内完成，如有疑问请联系客服
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 会员计划列表，放在底部 -->
    <el-row>
      <el-col :span="24">
        <el-card class="plans-card" shadow="hover">
          <div class="plans-header">
            <i class="el-icon-shopping-cart-2"></i>
            <h3>会员套餐</h3>
            <el-button 
              v-if="membershipPlans.length === 0 && !plansLoading" 
              type="text" 
              icon="el-icon-refresh" 
              class="refresh-btn"
              @click="fetchMembershipPlans">
              重新加载
            </el-button>
          </div>
          <div v-loading="plansLoading" class="plans-content">
            <el-empty v-if="membershipPlans.length === 0 && !plansLoading" description="暂无会员套餐数据">
              <el-button type="primary" size="small" @click="fetchMembershipPlans">刷新</el-button>
            </el-empty>
            <el-row :gutter="20" class="plan-list" v-else>
              <el-col :xs="24" :sm="24" :md="12" :lg="6" :xl="6" v-for="plan in membershipPlans" :key="plan.id">
                <div 
                  class="plan-card" 
                  :class="{ 
                    'active-plan': membership && membership.level === plan.level,
                    'free-plan': plan.level === 'free'
                   }"
                >
                  <div class="plan-header">
                    <h3>{{ plan.name }}</h3>
                    <div class="plan-price" v-if="plan.price > 0">
                      <span class="currency">¥</span>
                      <span class="amount">{{ plan.price }}</span>
                      <span class="period">/ {{ getDurationText(plan.duration) }}</span>
                    </div>
                    <div class="plan-price" v-else>免费</div>
                  </div>
                  
                  <div class="plan-features">
                    <p>{{ plan.description }}</p>
                    <ul>
                      <li v-for="(feature, index) in parseFeatures(plan.features)" :key="index">
                        <i class="el-icon-check"></i> {{ feature }}
                      </li>
                    </ul>
                  </div>
                  
                  <div class="plan-action">
                    <el-button 
                      :type="membership && membership.level === plan.level ? 'success' : 'primary'"
                      :disabled="membership && membership.level === plan.level"
                      @click="handlePurchasePlan(plan)"
                      v-if="plan.level !== 'free'"
                      >
                      {{ membership && membership.level === plan.level ? '当前方案' : '立即购买' }}
                    </el-button>
                    <span v-else-if="plan.level === 'free'">免费方案</span>
                  </div>
                </div>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 支付对话框 -->
    <el-dialog 
      title="确认购买" 
      :visible.sync="paymentDialogVisible" 
      width="420px"
      :close-on-click-modal="false"
      >
      <div class="payment-dialog-content">
        <div class="selected-plan-info" v-if="selectedPlan">
          <h4>{{ selectedPlan.name }}</h4>
          <div class="plan-price-large">
            <span class="currency">¥</span>
            <span class="amount">{{ selectedPlan.price }}</span>
          </div>
          <p class="plan-duration">有效期: {{ getDurationText(selectedPlan.duration) }}</p>
          
          <el-divider></el-divider>
          
          <div class="payment-options">
            <h4>支付方式</h4>
            <div class="payment-methods">
              <el-radio-group v-model="paymentMethod">
                <el-radio label="wechat">
                  <i class="payment-icon wechat"></i> 微信支付
                </el-radio>
                <el-radio label="alipay">
                  <i class="payment-icon alipay"></i> 支付宝
                </el-radio>
              </el-radio-group>
            </div>
            
            <!-- 显示收款码区域 -->
            <div class="payment-qrcode-container" v-if="paymentMethod">
              <div v-if="paymentMethod === 'wechat'" class="qrcode-box">
                <img :src="getQRCodeUrl('wechat')" alt="微信支付" class="payment-qrcode">
                <div class="qrcode-tip">请使用微信扫码支付</div>
              </div>
              <div v-else-if="paymentMethod === 'alipay'" class="qrcode-box">
                <img :src="getQRCodeUrl('alipay')" alt="支付宝支付" class="payment-qrcode">
                <div class="qrcode-tip">请使用支付宝扫码支付</div>
              </div>
            </div>
            
            <!-- 备注输入框 -->
            <div class="payment-remark">
              <p class="remark-tip">重要提示: 支付后请填写备注信息，以便管理员核实</p>
              <el-input 
                v-model="paymentRemark" 
                type="text" 
                placeholder="请填写支付账号后4位或其他备注信息"
                maxlength="50"
                show-word-limit
              ></el-input>
            </div>
            
            <div class="auto-renew-option">
              <el-checkbox v-model="autoRenew">开启自动续费</el-checkbox>
              <el-tooltip content="开启后，会员到期时将自动从您的账户扣款续费" placement="top">
                <i class="el-icon-question"></i>
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button @click="paymentDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitPaymentRequest" :loading="purchasing">
          已完成支付，提交审核
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { 
  getMembershipPlans, 
  getUserMembership, 
  purchaseMembership, 
  cancelAutoRenew,
  getDailyUsage,
  getUserPendingOrders
} from '@/api/membership'

export default {
  name: 'Membership',
  data() {
    return {
      // 会员数据
      membership: null,
      membershipPlans: [],
      dailyUsage: {
        usage_count: 0,
        daily_limit: 0
      },
      
      // 弹窗和表单
      selectedPlan: null,
      paymentDialogVisible: false,
      paymentMethod: 'wechat',
      paymentRemark: '',  // 支付备注
      autoRenew: false,
      
      // 加载状态
      loading: false,
      plansLoading: false,
      usageLoading: false,
      purchasing: false,
      cancelling: false,
      
      // 订单状态
      orderStatus: null,
      pendingOrders: [],
      orderLoading: false
    }
  },
  computed: {
    // 计算使用进度百分比
    usagePercentage() {
      if (!this.dailyUsage.daily_limit || this.dailyUsage.daily_limit < 0) return 0
      const percentage = (this.dailyUsage.usage_count / this.dailyUsage.daily_limit) * 100
      return Math.min(percentage, 100)
    },
    // 根据使用百分比设置颜色
    usageColor() {
      const percentage = this.usagePercentage
      if (percentage < 50) return '#67c23a'
      if (percentage < 80) return '#e6a23c'
      return '#f56c6c'
    },
    // 使用量显示文本
    usageCountText() {
      const count = this.dailyUsage.usage_count || 0
      const limit = this.dailyUsage.daily_limit < 0 ? '无限制' : this.dailyUsage.daily_limit
      return count + ' / ' + limit
    },
    // 剩余量显示文本
    remainingText() {
      if (this.dailyUsage.daily_limit < 0) {
        return '剩余: 无限制'
      }
      const remaining = this.dailyUsage.daily_limit - this.dailyUsage.usage_count
      return '剩余: ' + remaining
    }
  },
  created() {
    this.fetchMembershipData()
  },
  methods: {
    // 获取所有会员相关数据
    async fetchMembershipData() {
      this.fetchUserMembership()
      this.fetchMembershipPlans()
      this.fetchDailyUsage()
    },
    
    // 获取用户会员信息
    async fetchUserMembership() {
      this.loading = true
      try {
        const response = await getUserMembership()
        if (response.success) {
          this.membership = response.membership
        } else {
          this.$message.error(response.message)
        }
      } catch (error) {
        console.error('获取会员信息失败:', error)
        this.$message.error('获取会员信息失败')
      } finally {
        this.loading = false
      }
    },
    
    // 获取会员计划列表
    async fetchMembershipPlans() {
      this.plansLoading = true
      try {
        const response = await getMembershipPlans()
        console.log('会员计划数据:', response)
        if (response.success) {
          if (Array.isArray(response.plans) && response.plans.length > 0) {
            this.membershipPlans = response.plans
          } else {
            console.warn('会员计划数据为空或格式错误:', response.plans)
            this.$message.warning('会员计划数据为空，使用备用数据')
            this.useFallbackPlans()
          }
        } else {
          this.$message.error(response.message)
          this.useFallbackPlans()
        }
      } catch (error) {
        console.error('获取会员计划失败:', error)
        this.$message.error('获取会员计划失败，使用备用数据')
        this.useFallbackPlans()
      } finally {
        this.plansLoading = false
      }
    },
    
    // 使用备用会员计划数据
    useFallbackPlans() {
      this.membershipPlans = [
        {
          id: 1,
          name: '免费用户',
          level: 'free',
          price: 0,
          duration: -1,
          task_priority: 0,
          daily_limit: 5,
          description: '免费体验版，功能受限',
          features: JSON.stringify(["基础功能使用","每日限制5次使用","无优先级","有广告"]),
          is_active: true
        },
        {
          id: 2,
          name: '月度会员',
          level: 'monthly',
          price: 39.9,
          duration: 30,
          task_priority: 1,
          daily_limit: 30,
          description: '月度会员，限时优惠',
          features: JSON.stringify(["所有基础功能","每日限制30次使用","普通优先级","无广告"]),
          is_active: true
        },
        {
          id: 3,
          name: '季度会员',
          level: 'quarter',
          price: 99.9,
          duration: 90,
          task_priority: 2,
          daily_limit: 100,
          description: '季度会员，超值优惠',
          features: JSON.stringify(["所有高级功能","每日限制100次使用","较高优先级","无广告","专属客服"]),
          is_active: true
        },
        {
          id: 4,
          name: '年度会员',
          level: 'yearly',
          price: 299.9,
          duration: 365,
          task_priority: 3,
          daily_limit: -1,
          description: '年度会员，尊享特权',
          features: JSON.stringify(["所有高级功能","无使用次数限制","最高优先级","无广告","专属客服","优先体验新功能"]),
          is_active: true
        }
      ]
    },
    
    // 获取每日使用量
    async fetchDailyUsage() {
      this.usageLoading = true
      try {
        const response = await getDailyUsage()
        if (response.success) {
          this.dailyUsage = response.usage
        } else {
          this.$message.error(response.message)
        }
      } catch (error) {
        console.error('获取使用量失败:', error)
        this.$message.error('获取使用量失败')
      } finally {
        this.usageLoading = false
      }
    },
    
    // 处理购买会员
    handlePurchasePlan(plan) {
      this.selectedPlan = plan
      this.paymentDialogVisible = true
      // 默认不选自动续费
      this.autoRenew = false
    },
    
    // 确认购买会员 (修改为提交支付请求)
    async submitPaymentRequest() {
      if (!this.selectedPlan) return
      
      if (!this.paymentRemark.trim()) {
        this.$message.warning('请填写支付备注信息')
        return
      }
      
      this.purchasing = true
      try {
        const data = {
          plan_id: this.selectedPlan.id,
          auto_renew: this.autoRenew,
          payment_method: this.paymentMethod,
          payment_remark: this.paymentRemark,
          order_status: 'pending',  // 待审核状态
          created_at: new Date().toISOString()
        }
        
        const response = await purchaseMembership(data)
        if (response.success) {
          this.$message.success('支付申请已提交，请等待管理员审核')
          this.paymentDialogVisible = false
          // 刷新所有会员相关数据
          this.fetchMembershipData()
          // 获取待审核订单
          this.fetchPendingOrders()
        } else {
          this.$message.error(response.message)
        }
      } catch (error) {
        console.error('提交支付申请失败:', error)
        this.$message.error('提交支付申请失败')
      } finally {
        this.purchasing = false
      }
    },
    
    // 获取待审核订单
    async fetchPendingOrders() {
      this.orderLoading = true
      try {
        // 这里应该是从后端获取当前用户的待审核订单
        // 实际实现时连接后端API
        const response = await getUserPendingOrders()
        if (response.success) {
          this.pendingOrders = response.orders || []
        }
      } catch (error) {
        console.error('获取待审核订单失败:', error)
      } finally {
        this.orderLoading = false
      }
    },
    
    // 取消自动续费
    async handleCancelAutoRenew() {
      this.cancelling = true
      try {
        const response = await cancelAutoRenew()
        if (response.success) {
          this.$message.success(response.message || '已取消自动续费')
          // 刷新所有会员相关数据
          this.fetchMembershipData()
        } else {
          this.$message.error(response.message)
        }
      } catch (error) {
        console.error('取消自动续费失败:', error)
        this.$message.error('取消自动续费失败')
      } finally {
        this.cancelling = false
      }
    },
    
    // 格式化日期
    formatDate(date) {
      if (!date) return '永久'
      const d = new Date(date)
      return d.getFullYear() + '-' + (d.getMonth() + 1) + '-' + d.getDate()
    },
    
    // 根据level获取会员名称
    getLevelName(level) {
      const levelMap = {
        'free': '免费用户',
        'monthly': '月度会员',
        'quarter': '季度会员',
        'yearly': '年度会员'
      }
      return levelMap[level] || '未知'
    },
    
    // 根据level获取CSS类名
    getLevelClass(level) {
      return `level-${level}`
    },
    
    // 解析features JSON字符串
    parseFeatures(features) {
      if (!features) return [];
      try {
        // 如果是字符串，尝试解析JSON
        if (typeof features === 'string') {
          const parsed = JSON.parse(features);
          // 确保解析结果是数组
          return Array.isArray(parsed) ? parsed : [features];
        }
        // 如果已经是数组，直接返回
        if (Array.isArray(features)) {
          return features;
        }
        // 其他情况，转换为字符串并作为单个功能点显示
        return [features.toString()];
      } catch (e) {
        console.error('解析功能特性失败:', e);
        // 如果解析失败，将原始字符串作为单个功能点返回
        return features ? [features.toString()] : [];
      }
    },
    
    // 获取时长显示文本
    getDurationText(days) {
      if (days < 0) return '永久'
      if (days >= 365) return Math.floor(days / 365) + '年'
      if (days >= 30) return Math.floor(days / 30) + '个月'
      return days + '天'
    },
    
    // 格式化使用量显示
    formatUsage(percentage) {
      if (this.dailyUsage.daily_limit < 0) return '无限制'
      return `${percentage}%`
    },
    
    // 格式化日期时间
    formatDateTime(dateTime) {
      if (!dateTime) return '';
      const date = new Date(dateTime);
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
    },
    
    // 获取二维码图片URL
    getQRCodeUrl(method) {
      if (!this.selectedPlan) return '';

      if (method === 'wechat' && this.selectedPlan.wechat_qr_code) {
        return this.selectedPlan.wechat_qr_code;
      } else if (method === 'alipay' && this.selectedPlan.alipay_qr_code) {
        return this.selectedPlan.alipay_qr_code;
      }
      
      // 备用方案：如果套餐数据中没有二维码URL，则使用拼接的路径
      return `/uploads/qrcode/${this.selectedPlan.level}_${method}.jpg`;
    }
  }
}
</script>

<style scoped>
/* 顶部卡片样式 */
.top-cards {
  margin-bottom: 20px;
}

.membership-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  color: #409EFF;
  font-weight: 400;
}

/* 会员状态卡片和使用量卡片高度一致 */
.membership-status, .usage-card {
  height: 100%;
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
}

.membership-content, .usage-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 220px; /* 固定最小高度 */
  padding: 15px;
}

/* 使用量进度条容器 */
.usage-progress-container {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 20px 0;
}

/* 使用量信息 */
.usage-info {
  display: flex;
  justify-content: space-between;
  margin-top: 15px;
  color: #606266;
  font-size: 14px;
}

.membership-header, .usage-header, .plans-header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.membership-header i, .usage-header i, .plans-header i {
  font-size: 24px;
  margin-right: 10px;
  color: #409EFF;
}

.membership-header h3, .usage-header h3, .plans-header h3 {
  margin: 0;
  font-weight: 500;
}

/* 刷新按钮样式 */
.refresh-btn {
  margin-left: auto;
  padding: 0;
  font-size: 12px;
}

/* 会员等级样式 */
.membership-level {
  display: flex;
  align-items: center;
  padding: 10px 15px;
  border-radius: 5px;
  margin-bottom: 20px;
}

.level-icon {
  font-size: 24px;
  margin-right: 10px;
}

.level-name {
  font-size: 18px;
  font-weight: 500;
}

/* 不同等级的颜色 */
.level-free {
  background-color: #f2f6fc;
  color: #909399;
}

.level-monthly {
  background-color: #ecf5ff;
  color: #409EFF;
}

.level-quarter {
  background-color: #f0f9eb;
  color: #67c23a;
}

.level-yearly {
  background-color: #fdf6ec;
  color: #e6a23c;
}

/* 会员信息样式 */
.membership-info {
  padding: 0;
  margin-top: auto;
}

.info-item {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}

.info-item .label {
  color: #606266;
  width: 80px;
}

.info-item .value {
  color: #303133;
  flex: 1;
}

/* 无会员信息时的样式 */
.no-membership {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
}

.no-membership i {
  font-size: 40px;
  margin-bottom: 10px;
}

/* 会员计划卡片 */
.plan-list {
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
}

.plan-card {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  transition: all 0.3s;
  margin-bottom: 20px;
  overflow: hidden;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.plan-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0,0,0,0.1);
}

.active-plan {
  border-color: #67c23a;
  box-shadow: 0 0 10px rgba(103, 194, 58, 0.2);
}

.free-plan {
  opacity: 0.85;
}

.plan-header {
  background-color: #f5f7fa;
  padding: 15px;
  text-align: center;
  border-bottom: 1px solid #e4e7ed;
}

.plan-header h3 {
  margin: 0 0 10px;
  font-weight: 500;
}

.plan-price {
  font-size: 24px;
  font-weight: 700;
  color: #409EFF;
}

.plan-price .currency {
  font-size: 16px;
  vertical-align: top;
}

.plan-price .period {
  font-size: 14px;
  font-weight: normal;
  color: #909399;
}

.plan-features {
  padding: 15px;
  flex-grow: 1; /* 使内容区域占用剩余空间 */
}

.plan-features p {
  margin-top: 0;
  margin-bottom: 15px;
  font-size: 14px;
  color: #606266;
  word-break: break-word; /* 防止长文本溢出 */
}

.plan-features ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.plan-features li {
  margin-bottom: 10px;
  font-size: 14px;
  display: flex;
  align-items: flex-start;
}

.plan-features li i {
  margin-right: 5px;
  color: #67c23a;
}

.plan-action {
  padding: 15px;
  text-align: center;
  border-top: 1px solid #e4e7ed;
}

/* 支付对话框 */
.payment-dialog-content {
  text-align: center;
}

.selected-plan-info h4 {
  margin-top: 0;
  font-weight: 500;
}

.plan-price-large {
  font-size: 36px;
  font-weight: 700;
  color: #409EFF;
  margin: 20px 0;
}

.plan-price-large .currency {
  font-size: 20px;
  vertical-align: super;
}

.plan-duration {
  color: #909399;
  margin-bottom: 20px;
}

.payment-options {
  text-align: left;
}

.payment-methods {
  margin-bottom: 20px;
}

.payment-icon {
  display: inline-block;
  width: 24px;
  height: 24px;
  vertical-align: middle;
  margin-right: 5px;
}

.auto-renew-option {
  margin-top: 20px;
  display: flex;
  align-items: center;
}

.auto-renew-option i {
  margin-left: 5px;
  color: #909399;
  cursor: pointer;
}

/* 支付收款码样式 */
.payment-qrcode-container {
  margin: 20px 0;
  text-align: center;
}

.qrcode-box {
  display: inline-block;
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background-color: #fff;
}

.payment-qrcode {
  width: 200px;
  height: 200px;
  object-fit: contain;
}

.qrcode-tip {
  margin-top: 10px;
  font-size: 12px;
  color: #606266;
}

/* 备注输入框样式 */
.payment-remark {
  margin: 20px 0;
}

.remark-tip {
  margin-bottom: 10px;
  font-size: 13px;
  color: #f56c6c;
  font-weight: bold;
}

/* 对于小屏幕设备的适配 */
@media (max-width: 768px) {
  .membership-status, .usage-card {
    height: auto;
  }
  
  .membership-content, .usage-content {
    min-height: auto;
    padding: 10px;
  }
  
  .plan-list {
    display: block;
  }
  
  .plan-card {
    margin-right: 0;
    height: auto; /* 在移动端不强制相同高度 */
  }
  
  .plan-features {
    min-height: auto; /* 在移动端取消最小高度限制 */
  }
}

/* 待审核订单样式 */
.pending-orders-row {
  margin-bottom: 20px;
}

.pending-orders-header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.pending-orders-header i {
  font-size: 24px;
  margin-right: 10px;
  color: #E6A23C;
}

.pending-orders-header h3 {
  margin: 0;
  font-weight: 500;
}

.pending-orders-content {
  margin-bottom: 10px;
}

.pending-tip {
  margin-top: 15px;
  color: #909399;
  font-size: 13px;
  display: flex;
  align-items: center;
}

.pending-tip i {
  margin-right: 5px;
  color: #E6A23C;
}
</style> 